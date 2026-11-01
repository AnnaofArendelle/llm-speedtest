// Package load generates closed-loop (fixed concurrency) or open-loop
// (Poisson arrivals) request load against one model and reports throughput,
// error rate, latency percentiles and goodput under an SLO.
package load

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/AnnaofArendelle/llm-speedtest/internal/apiclient"
	"github.com/AnnaofArendelle/llm-speedtest/internal/classify"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/latency"
	"github.com/AnnaofArendelle/llm-speedtest/internal/metrics"
)

// Modes.
const (
	ModeConcurrency = "concurrency"
	ModeRate        = "rate"
)

// SLO defines the goodput thresholds (zero disables a threshold).
type SLO struct {
	TTFT time.Duration `json:"ttft" yaml:"ttft"`
	TPOT time.Duration `json:"tpot" yaml:"tpot"`
	E2E  time.Duration `json:"e2e" yaml:"e2e"`
}

// Spec configures a load test.
type Spec struct {
	Model        string        `json:"model" yaml:"model"`
	Mode         string        `json:"mode" yaml:"mode"`
	Levels       []int         `json:"levels" yaml:"levels"`     // concurrency levels (closed loop)
	Rate         float64       `json:"rate" yaml:"rate"`         // requests per second (open loop)
	Duration     time.Duration `json:"duration" yaml:"duration"` // per level
	Requests     int           `json:"requests" yaml:"requests"` // optional cap per level
	RampUp       time.Duration `json:"ramp_up" yaml:"ramp_up"`   // closed loop: stagger worker start
	InputTokens  int           `json:"input_tokens" yaml:"input_tokens"`
	OutputTokens int           `json:"output_tokens" yaml:"output_tokens"`
	Stream       bool          `json:"stream" yaml:"stream"`
	Timeout      time.Duration `json:"timeout" yaml:"timeout"`
	StallTimeout time.Duration `json:"stall_timeout" yaml:"stall_timeout"`
	MaxInflight  int           `json:"max_inflight" yaml:"max_inflight"` // open loop hard cap
	SLO          SLO           `json:"slo" yaml:"slo"`
	Seed         int64         `json:"seed" yaml:"seed"`
}

// Default returns a conservative load spec.
func Default() Spec {
	return Spec{Mode: ModeConcurrency, Levels: []int{1, 2, 4}, Duration: 30 * time.Second, InputTokens: 100, OutputTokens: 64, Stream: true,
		Timeout: 120 * time.Second, StallTimeout: 30 * time.Second, MaxInflight: 64, SLO: SLO{TTFT: time.Second, TPOT: 50 * time.Millisecond}, Seed: 42}
}

// Normalize fills zero values.
func (s *Spec) Normalize() {
	d := Default()
	if s.Mode != ModeRate {
		s.Mode = ModeConcurrency
	}
	if len(s.Levels) == 0 {
		s.Levels = d.Levels
	}
	if s.Rate <= 0 {
		s.Rate = 2
	}
	if s.Duration <= 0 {
		s.Duration = d.Duration
	}
	if s.InputTokens <= 0 {
		s.InputTokens = d.InputTokens
	}
	if s.OutputTokens <= 0 {
		s.OutputTokens = d.OutputTokens
	}
	if s.Timeout <= 0 {
		s.Timeout = d.Timeout
	}
	if s.StallTimeout <= 0 {
		s.StallTimeout = d.StallTimeout
	}
	if s.MaxInflight <= 0 {
		s.MaxInflight = d.MaxInflight
	}
}

// LevelResult summarises one level.
type LevelResult struct {
	Level        int            `json:"level"`
	Mode         string         `json:"mode"`
	Requests     int            `json:"requests"`
	OK           int            `json:"ok"`
	Errors       int            `json:"errors"`
	Duration     time.Duration  `json:"duration"`
	RPS          float64        `json:"rps"`
	OutTPS       float64        `json:"out_tok_per_s"`
	Goodput      float64        `json:"goodput"` // share of ok requests meeting the SLO
	TTFT         metrics.Stats  `json:"ttft_s"`
	E2E          metrics.Stats  `json:"e2e_s"`
	ITL          metrics.Stats  `json:"itl_s"`
	QueueWait    metrics.Stats  `json:"queue_wait_s"`
	OutputTokens int            `json:"output_tokens"`
	Reasons      map[string]int `json:"reasons,omitempty"`
}

// Sink receives requests.
type Sink interface {
	Request(ctx context.Context, r core.LoadRequest) error
}

// SinkFunc adapts a function to Sink.
type SinkFunc func(ctx context.Context, r core.LoadRequest) error

// Request implements Sink.
func (f SinkFunc) Request(ctx context.Context, r core.LoadRequest) error { return f(ctx, r) }

// Runner executes a load test.
type Runner struct {
	Client *apiclient.Client
	Spec   Spec
	Sink   Sink
	Report func(core.Event)
	Stage  string
	Name   string // cleaned model name for classification
}

func (r *Runner) emit(ev core.Event) {
	if r.Report == nil {
		return
	}
	ev.At = time.Now()
	ev.Stage = r.Stage
	r.Report(ev)
}

// Run executes all levels and returns their results and every request.
func (r *Runner) Run(ctx context.Context) ([]LevelResult, []core.LoadRequest, error) {
	r.Spec.Normalize()
	if r.Stage == "" {
		r.Stage = "load"
	}
	var results []LevelResult
	var all []core.LoadRequest
	levels := r.Spec.Levels
	if r.Spec.Mode == ModeRate {
		levels = []int{int(r.Spec.Rate*100 + 0.5)} // level encodes rate×100 for storage
	}
	for i, lvl := range levels {
		r.emit(core.Event{Kind: core.EventLevel, Round: i + 1, Rounds: len(levels), Message: fmt.Sprintf("%s level %d", r.Spec.Mode, lvl)})
		var reqs []core.LoadRequest
		var err error
		if r.Spec.Mode == ModeRate {
			reqs, err = r.openLoop(ctx, lvl)
		} else {
			reqs, err = r.closedLoop(ctx, lvl)
		}
		res := Summarize(reqs, r.Spec, lvl)
		results = append(results, res)
		all = append(all, reqs...)
		r.emit(core.Event{Kind: core.EventProgress, Round: i + 1, Rounds: len(levels), Done: res.Requests, Total: res.Requests, OK: res.OK, Fail: res.Errors, Data: res})
		if err != nil {
			return results, all, err
		}
	}
	r.emit(core.Event{Kind: core.EventDone, Message: "load complete"})
	return results, all, nil
}

func (r *Runner) one(ctx context.Context, level, idx int, rnd *rand.Rand, queueWait time.Duration) core.LoadRequest {
	prompt := latency.Synthesize(r.Spec.InputTokens, fmt.Sprintf("%08x", rnd.Uint32()))
	req := apiclient.ChatRequest{Model: r.Spec.Model, Messages: []apiclient.Message{{Role: "user", Content: prompt}}, MaxTokens: r.Spec.OutputTokens}
	lr := core.LoadRequest{Level: level, Index: idx, At: time.Now(), QueueWait: queueWait}
	pctx, cancel := context.WithTimeout(ctx, r.Spec.Timeout)
	defer cancel()
	if r.Spec.Stream {
		res := r.Client.ChatStream(pctx, req, apiclient.StreamOptions{StallTimeout: r.Spec.StallTimeout})
		if res.OK() {
			lr.OK = true
			lr.TTFT, lr.E2E, lr.OutputTokens = res.TTFT(), res.E2E(), res.OutputTokens
			if res.OutputTokens > 1 {
				lr.ITL = res.TPOT()
			}
			return lr
		}
		status := res.Status
		if res.Err != nil && status == 200 || (status == 200 && res.FirstTokenAt.IsZero()) {
			status = 0
		}
		lr.Reason = classify.Evaluate(status, res.Body, r.Name).Reason
		if lr.Reason == "ok" {
			lr.Reason = "empty_response"
		}
		lr.E2E = time.Since(lr.At)
		return lr
	}
	res := r.Client.Chat(pctx, req)
	lr.E2E = res.Latency
	if res.OK() {
		lr.OK = true
		if res.Usage != nil {
			lr.OutputTokens = res.Usage.CompletionTokens
		} else {
			lr.OutputTokens = apiclient.EstimateTokens(res.Content)
		}
		return lr
	}
	lr.Reason = classify.Evaluate(res.Status, res.Body, r.Name).Reason
	return lr
}

func (r *Runner) closedLoop(ctx context.Context, level int) ([]core.LoadRequest, error) {
	var mu sync.Mutex
	var reqs []core.LoadRequest
	deadline := time.Now().Add(r.Spec.Duration)
	var counter int
	next := func() (int, bool) {
		mu.Lock()
		defer mu.Unlock()
		if r.Spec.Requests > 0 && counter >= r.Spec.Requests {
			return 0, false
		}
		if time.Now().After(deadline) {
			return 0, false
		}
		counter++
		return counter - 1, true
	}
	g, gctx := errgroup.WithContext(ctx)
	for w := 0; w < level; w++ {
		w := w
		g.Go(func() error {
			if r.Spec.RampUp > 0 && level > 1 {
				delay := time.Duration(int64(r.Spec.RampUp) * int64(w) / int64(level))
				select {
				case <-time.After(delay):
				case <-gctx.Done():
					return gctx.Err()
				}
			}
			rnd := rand.New(rand.NewSource(r.Spec.Seed + int64(level)*1000 + int64(w)))
			for {
				if gctx.Err() != nil {
					return gctx.Err()
				}
				idx, ok := next()
				if !ok {
					return nil
				}
				lr := r.one(gctx, level, idx, rnd, 0)
				if gctx.Err() != nil {
					return gctx.Err()
				}
				if r.Sink != nil {
					if err := r.Sink.Request(gctx, lr); err != nil {
						return err
					}
				}
				mu.Lock()
				reqs = append(reqs, lr)
				n := len(reqs)
				mu.Unlock()
				if n%10 == 0 {
					r.emit(core.Event{Kind: core.EventProgress, Done: n, Total: 0, Message: fmt.Sprintf("level %d", level)})
				}
			}
		})
	}
	err := g.Wait()
	sortReqs(reqs)
	if err != nil && ctx.Err() != nil {
		return reqs, ctx.Err()
	}
	return reqs, err
}

func (r *Runner) openLoop(ctx context.Context, level int) ([]core.LoadRequest, error) {
	var mu sync.Mutex
	var reqs []core.LoadRequest
	rate := r.Spec.Rate
	rnd := rand.New(rand.NewSource(r.Spec.Seed))
	sem := semaphore.NewWeighted(int64(r.Spec.MaxInflight))
	g, gctx := errgroup.WithContext(ctx)
	deadline := time.Now().Add(r.Spec.Duration)
	idx := 0
	for time.Now().Before(deadline) && (r.Spec.Requests == 0 || idx < r.Spec.Requests) {
		// exponential inter-arrival
		wait := time.Duration(rnd.ExpFloat64() / rate * float64(time.Second))
		select {
		case <-time.After(wait):
		case <-gctx.Done():
			return reqs, gctx.Err()
		}
		i := idx
		idx++
		seed := rnd.Int63()
		g.Go(func() error {
			qstart := time.Now()
			if err := sem.Acquire(gctx, 1); err != nil {
				return err
			}
			defer sem.Release(1)
			lr := r.one(gctx, level, i, rand.New(rand.NewSource(seed)), time.Since(qstart))
			if gctx.Err() != nil {
				return gctx.Err()
			}
			if r.Sink != nil {
				if err := r.Sink.Request(gctx, lr); err != nil {
					return err
				}
			}
			mu.Lock()
			reqs = append(reqs, lr)
			mu.Unlock()
			return nil
		})
	}
	err := g.Wait()
	sortReqs(reqs)
	if err != nil && ctx.Err() != nil {
		return reqs, ctx.Err()
	}
	return reqs, err
}

func sortReqs(reqs []core.LoadRequest) {
	sort.Slice(reqs, func(i, j int) bool { return reqs[i].Index < reqs[j].Index })
}

// Summarize computes the level result.
func Summarize(reqs []core.LoadRequest, spec Spec, level int) LevelResult {
	res := LevelResult{Level: level, Mode: spec.Mode, Requests: len(reqs), Reasons: map[string]int{}}
	if len(reqs) == 0 {
		return res
	}
	var ttft, e2e, itl, qw []float64
	first, last := reqs[0].At, reqs[0].At
	good := 0
	for _, q := range reqs {
		if q.At.Before(first) {
			first = q.At
		}
		if end := q.At.Add(q.E2E); end.After(last) {
			last = end
		}
		if q.QueueWait > 0 {
			qw = append(qw, q.QueueWait.Seconds())
		}
		if !q.OK {
			res.Errors++
			res.Reasons[q.Reason]++
			continue
		}
		res.OK++
		res.OutputTokens += q.OutputTokens
		e2e = append(e2e, q.E2E.Seconds())
		if q.TTFT > 0 {
			ttft = append(ttft, q.TTFT.Seconds())
		}
		if q.ITL > 0 {
			itl = append(itl, q.ITL.Seconds())
		}
		if meetsSLO(q, spec.SLO) {
			good++
		}
	}
	res.Duration = last.Sub(first)
	if res.Duration > 0 {
		res.RPS = float64(res.OK) / res.Duration.Seconds()
		res.OutTPS = float64(res.OutputTokens) / res.Duration.Seconds()
	}
	if res.OK > 0 {
		res.Goodput = float64(good) / float64(res.OK)
	}
	res.TTFT, res.E2E, res.ITL, res.QueueWait = metrics.Summarise(ttft), metrics.Summarise(e2e), metrics.Summarise(itl), metrics.Summarise(qw)
	return res
}

func meetsSLO(q core.LoadRequest, slo SLO) bool {
	if slo.TTFT > 0 && q.TTFT > 0 && q.TTFT > slo.TTFT {
		return false
	}
	if slo.TPOT > 0 && q.ITL > 0 && q.ITL > slo.TPOT {
		return false
	}
	if slo.E2E > 0 && q.E2E > slo.E2E {
		return false
	}
	return true
}
