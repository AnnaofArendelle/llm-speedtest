// Package latency measures streaming latency characteristics (TTFT, time to
// second token, inter-token latency, time per output token, end-to-end
// latency and decode throughput) for selected models.
package latency

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/AnnaofArendelle/llm-speedtest/internal/apiclient"
	"github.com/AnnaofArendelle/llm-speedtest/internal/classify"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/metrics"
)

// Spec configures a latency run.
type Spec struct {
	Samples      int           `json:"samples" yaml:"samples"`             // measured samples per model
	Warmup       int           `json:"warmup" yaml:"warmup"`               // cold samples recorded but excluded from stats
	InputTokens  int           `json:"input_tokens" yaml:"input_tokens"`   // target prompt size
	OutputTokens int           `json:"output_tokens" yaml:"output_tokens"` // max_tokens
	Concurrency  int           `json:"concurrency" yaml:"concurrency"`     // models measured in parallel
	PerProvider  int           `json:"per_provider" yaml:"per_provider"`
	Timeout      time.Duration `json:"timeout" yaml:"timeout"`
	StallTimeout time.Duration `json:"stall_timeout" yaml:"stall_timeout"`
	Temperature  *float64      `json:"temperature,omitempty" yaml:"temperature,omitempty"`
	Seed         int64         `json:"seed" yaml:"seed"`
}

// Default returns the standard latency spec.
func Default() Spec {
	return Spec{Samples: 5, Warmup: 1, InputTokens: 200, OutputTokens: 128, Concurrency: 2, PerProvider: 1, Timeout: 120 * time.Second, StallTimeout: 30 * time.Second, Seed: 42}
}

// Normalize fills zero values.
func (s *Spec) Normalize() {
	d := Default()
	if s.Samples <= 0 {
		s.Samples = d.Samples
	}
	if s.Warmup < 0 {
		s.Warmup = 0
	}
	if s.InputTokens <= 0 {
		s.InputTokens = d.InputTokens
	}
	if s.OutputTokens <= 0 {
		s.OutputTokens = d.OutputTokens
	}
	if s.Concurrency <= 0 {
		s.Concurrency = d.Concurrency
	}
	if s.PerProvider <= 0 {
		s.PerProvider = 1
	}
	if s.Timeout <= 0 {
		s.Timeout = d.Timeout
	}
	if s.StallTimeout <= 0 {
		s.StallTimeout = d.StallTimeout
	}
}

// Target is a model to measure.
type Target struct {
	ID       string
	Provider string
	Name     string
}

// Sink receives samples.
type Sink interface {
	Sample(ctx context.Context, s core.LatencySample) error
}

// SinkFunc adapts a function to Sink.
type SinkFunc func(ctx context.Context, s core.LatencySample) error

// Sample implements Sink.
func (f SinkFunc) Sample(ctx context.Context, s core.LatencySample) error { return f(ctx, s) }

// Runner executes latency measurements.
type Runner struct {
	Client *apiclient.Client
	Spec   Spec
	Sink   Sink
	Report func(core.Event)
	Stage  string
	// Skipper (optional) lets users skip models while the run is in progress.
	Skipper *core.Skipper
}

// Run measures every target. Samples for one model are sequential; models
// run concurrently within the global and per-provider limits.
func (r *Runner) Run(ctx context.Context, targets []Target) ([]core.LatencySample, error) {
	r.Spec.Normalize()
	if r.Stage == "" {
		r.Stage = "latency"
	}
	total := len(targets) * (r.Spec.Samples + r.Spec.Warmup)
	var mu sync.Mutex
	var all []core.LatencySample
	done, okN, failN := 0, 0, 0
	start := time.Now()
	emit := func(ev core.Event) {
		if r.Report != nil {
			ev.At = time.Now()
			ev.Stage = r.Stage
			r.Report(ev)
		}
	}
	emit(core.Event{Kind: core.EventRound, Round: 1, Rounds: 1, Total: total, Message: "latency"})
	global := semaphore.NewWeighted(int64(r.Spec.Concurrency))
	var pmu sync.Mutex
	perProv := map[string]*semaphore.Weighted{}
	g, gctx := errgroup.WithContext(ctx)
	for _, t := range targets {
		t := t
		g.Go(func() error {
			pmu.Lock()
			ps, ok := perProv[t.Provider]
			if !ok {
				ps = semaphore.NewWeighted(int64(r.Spec.PerProvider))
				perProv[t.Provider] = ps
			}
			pmu.Unlock()
			if err := ps.Acquire(gctx, 1); err != nil {
				return err
			}
			defer ps.Release(1)
			if err := global.Acquire(gctx, 1); err != nil {
				return err
			}
			defer global.Release(1)
			rnd := rand.New(rand.NewSource(r.Spec.Seed + int64(len(t.ID))))
			for i := 0; i < r.Spec.Warmup+r.Spec.Samples; i++ {
				if gctx.Err() != nil {
					return gctx.Err()
				}
				if r.Skipper.IsSkipped(t.ID) {
					break
				}
				smp := r.one(gctx, t, i, i >= r.Spec.Warmup, rnd)
				if gctx.Err() != nil {
					return gctx.Err()
				}
				if r.Sink != nil {
					if err := r.Sink.Sample(gctx, smp); err != nil {
						return err
					}
				}
				mu.Lock()
				all = append(all, smp)
				done++
				if smp.OK {
					okN++
				} else {
					failN++
				}
				d, o, f := done, okN, failN
				mu.Unlock()
				el := time.Since(start)
				var eta float64
				if d > 0 {
					eta = el.Seconds() / float64(d) * float64(total-d)
				}
				emit(core.Event{Kind: core.EventSample, Done: d, Total: total, OK: o, Fail: f, ETA: eta, Sample: &smp})
				if !smp.OK && !classify.IsTransient(smp.Reason) {
					// Deterministic failure: do not burn the remaining samples.
					break
				}
			}
			return nil
		})
	}
	err := g.Wait()
	sort.Slice(all, func(i, j int) bool {
		if all[i].ModelID != all[j].ModelID {
			return all[i].ModelID < all[j].ModelID
		}
		return all[i].Index < all[j].Index
	})
	if err == nil {
		emit(core.Event{Kind: core.EventDone, Done: done, Total: total, OK: okN, Fail: failN, Message: "latency complete"})
	}
	return all, err
}

func (r *Runner) one(ctx context.Context, t Target, idx int, warm bool, rnd *rand.Rand) core.LatencySample {
	prompt := Synthesize(r.Spec.InputTokens, fmt.Sprintf("%08x", rnd.Uint32()))
	smp := core.LatencySample{ModelID: t.ID, Index: idx, Warm: warm, At: time.Now(), InputTokens: apiclient.EstimateTokens(prompt)}
	wctx, release := r.Skipper.Watch(ctx, r.Stage, t.ID, idx)
	defer release()
	if r.Report != nil {
		r.Report(core.Event{Kind: core.EventStart, Stage: r.Stage, Round: idx, Attempt: &core.Attempt{ModelID: t.ID, Round: idx, At: smp.At}})
	}
	pctx, cancel := context.WithTimeout(wctx, r.Spec.Timeout)
	defer cancel()
	req := apiclient.ChatRequest{Model: t.ID, Messages: []apiclient.Message{{Role: "user", Content: prompt}}, MaxTokens: r.Spec.OutputTokens, Temperature: r.Spec.Temperature}
	res := r.Client.ChatStream(pctx, req, apiclient.StreamOptions{StallTimeout: r.Spec.StallTimeout})
	if r.Skipper.IsSkipped(t.ID) {
		smp.Reason, smp.Message = classify.ReasonSkipped, "skipped by user"
		return smp
	}
	smp.ServedModel = res.Model
	if res.Usage != nil && res.Usage.PromptTokens > 0 {
		smp.InputTokens = res.Usage.PromptTokens
	}
	if !res.OK() {
		status := res.Status
		body := res.Body
		if res.Err != nil && status == 200 {
			status = 0
		}
		if status == 200 && res.FirstTokenAt.IsZero() {
			body = []byte("stream ended without any generated token")
			status = 0
		}
		out := classify.Evaluate(status, body, t.Name)
		if out.OK {
			out.Reason, out.Class = "empty_response", "empty_response"
		}
		smp.Reason, smp.Message = out.Reason, classify.CleanMessage(out.Message)
		if smp.Reason == "ok" {
			smp.Reason = "empty_response"
		}
		return smp
	}
	smp.OK = true
	smp.TTFT = res.TTFT()
	smp.TTST = res.TTST()
	smp.TPOT = res.TPOT()
	smp.E2E = res.E2E()
	smp.TokPerSec = res.TokensPerSecond()
	smp.OutputTokens = res.OutputTokens
	smp.TokensExact = res.TokensExact
	return smp
}

// ---------------------------------------------------------------- prompts

var corpus = []string{
	"The quick brown fox jumps over the lazy dog while the river keeps flowing towards the distant sea.",
	"Benchmarks are only useful when they are repeatable, documented and compared under the same conditions.",
	"A lighthouse keeper counts the ships at dusk and writes their names into a leather-bound ledger.",
	"城市的清晨从一杯热茶开始，街角的面包店刚刚打开卷帘门，空气里有麦香。",
	"测量延迟时应当同时记录首个字节与首个生成 token 的时间，因为代理可能缓冲响应。",
	"Les montagnes gardent la mémoire des glaciers, et les vallées racontent l'histoire des rivières.",
	"データは正直だが、解釈は慎重でなければならない。",
	"Every well-designed tool explains its own limitations before it presents a number.",
}

// Synthesize builds a prompt of roughly the requested token count. The nonce
// defeats prompt caches; the instruction asks for a long continuation so
// that generation runs until max_tokens.
func Synthesize(tokens int, nonce string) string {
	var sb strings.Builder
	fmt.Fprintf(&sb, "[session %s] Rewrite the following text in your own words, then continue it with as much detail as you can. Keep writing until you are cut off; do not summarise.\n\n", nonce)
	i := 0
	for apiclient.EstimateTokens(sb.String()) < tokens {
		sb.WriteString(corpus[i%len(corpus)])
		sb.WriteString(" ")
		i++
		if i > 100000 {
			break
		}
	}
	return sb.String()
}

// ---------------------------------------------------------------- aggregation

// ModelStats summarises the samples of one model.
type ModelStats struct {
	ModelID      string         `json:"model_id"`
	Samples      int            `json:"samples"`
	OK           int            `json:"ok"`
	Cold         *Point         `json:"cold,omitempty"` // first (warm-up) sample
	TTFT         metrics.Stats  `json:"ttft_s"`
	TTST         metrics.Stats  `json:"ttst_s"`
	TPOT         metrics.Stats  `json:"tpot_s"`
	E2E          metrics.Stats  `json:"e2e_s"`
	TokPerSec    metrics.Stats  `json:"tok_per_s"`
	OutputTokens metrics.Stats  `json:"output_tokens"`
	InputTPS     float64        `json:"input_tok_per_s"` // mean input tokens / TTFT
	TokensExact  bool           `json:"tokens_exact"`
	Served       string         `json:"served_model,omitempty"`
	Reasons      map[string]int `json:"reasons,omitempty"`
	Rank         int            `json:"rank,omitempty"`
}

// Point is a single-sample snapshot.
type Point struct {
	TTFT      float64 `json:"ttft_s"`
	E2E       float64 `json:"e2e_s"`
	TokPerSec float64 `json:"tok_per_s"`
}

// Summarize aggregates samples per model and ranks models by median TTFT
// (ties by tokens/s descending).
func Summarize(samples []core.LatencySample) []ModelStats {
	by := map[string][]core.LatencySample{}
	var order []string
	for _, s := range samples {
		if _, ok := by[s.ModelID]; !ok {
			order = append(order, s.ModelID)
		}
		by[s.ModelID] = append(by[s.ModelID], s)
	}
	out := []ModelStats{}
	for _, id := range order {
		ss := by[id]
		ms := ModelStats{ModelID: id, Samples: len(ss), Reasons: map[string]int{}, TokensExact: true}
		var ttft, ttst, tpot, e2e, tps, outTok []float64
		var inTPS []float64
		hasWarm := false
		for _, s := range ss {
			if s.Warm {
				hasWarm = true
			}
		}
		for _, s := range ss {
			if !s.OK {
				ms.Reasons[s.Reason]++
				continue
			}
			ms.OK++
			if ms.Served == "" {
				ms.Served = s.ServedModel
			}
			if !s.Warm && ms.Cold == nil {
				ms.Cold = &Point{TTFT: s.TTFT.Seconds(), E2E: s.E2E.Seconds(), TokPerSec: s.TokPerSec}
			}
			if hasWarm && !s.Warm {
				continue
			}
			if !s.TokensExact {
				ms.TokensExact = false
			}
			ttft = append(ttft, s.TTFT.Seconds())
			if s.TTST > 0 {
				ttst = append(ttst, s.TTST.Seconds())
			}
			if s.TPOT > 0 {
				tpot = append(tpot, s.TPOT.Seconds())
			}
			e2e = append(e2e, s.E2E.Seconds())
			if s.TokPerSec > 0 {
				tps = append(tps, s.TokPerSec)
			}
			outTok = append(outTok, float64(s.OutputTokens))
			if s.TTFT > 0 && s.InputTokens > 0 {
				inTPS = append(inTPS, float64(s.InputTokens)/s.TTFT.Seconds())
			}
		}
		ms.TTFT, ms.TTST, ms.TPOT, ms.E2E, ms.TokPerSec, ms.OutputTokens = metrics.Summarise(ttft), metrics.Summarise(ttst), metrics.Summarise(tpot), metrics.Summarise(e2e), metrics.Summarise(tps), metrics.Summarise(outTok)
		if len(inTPS) > 0 {
			ms.InputTPS = metrics.Summarise(inTPS).Mean
		}
		out = append(out, ms)
	}
	sort.SliceStable(out, func(i, j int) bool {
		a, b := out[i], out[j]
		if (a.TTFT.Count > 0) != (b.TTFT.Count > 0) {
			return a.TTFT.Count > 0
		}
		if r1, r2 := metrics.Round(a.TTFT.Median, 3), metrics.Round(b.TTFT.Median, 3); r1 != r2 {
			return r1 < r2
		}
		if a.TokPerSec.Median != b.TokPerSec.Median {
			return a.TokPerSec.Median > b.TokPerSec.Median
		}
		return a.ModelID < b.ModelID
	})
	for i := range out {
		if out[i].TTFT.Count > 0 {
			out[i].Rank = i + 1
		}
	}
	return out
}
