// Package scan implements the multi-round availability scan: every model is
// probed with a minimal chat completion, transient failures are retried,
// successes are re-sampled and remaining transient failures get a final
// confirmation. Concurrency is bounded globally and per provider, and
// gateway cooldown hints are honoured before retrying.
package scan

import (
	"context"
	"errors"
	"math/rand"
	"sort"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/AnnaofArendelle/llm-speedtest/internal/apiclient"
	"github.com/AnnaofArendelle/llm-speedtest/internal/classify"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
)

// RoundSpec configures one round. A zero Concurrency disables the round.
type RoundSpec struct {
	Concurrency int           `json:"concurrency" yaml:"concurrency"`
	PerProvider int           `json:"per_provider" yaml:"per_provider"`
	Timeout     time.Duration `json:"timeout" yaml:"timeout"`
}

// Spec configures a scan.
type Spec struct {
	First           RoundSpec     `json:"first" yaml:"first"`
	Retry           RoundSpec     `json:"retry" yaml:"retry"`
	Final           RoundSpec     `json:"final" yaml:"final"`
	Resample        int           `json:"resample" yaml:"resample"`
	MaxTokens       int           `json:"max_tokens" yaml:"max_tokens"`
	Prompt          string        `json:"prompt" yaml:"prompt"`
	Seed            int64         `json:"seed" yaml:"seed"`
	MaxCooldownWait time.Duration `json:"max_cooldown_wait" yaml:"max_cooldown_wait"`
	MaxBackoff      time.Duration `json:"max_backoff" yaml:"max_backoff"`
}

// Presets.
const (
	PresetQuick    = "quick"
	PresetStandard = "standard"
	PresetDeep     = "deep"
)

// Preset returns the spec for a named preset (standard when unknown).
func Preset(name string) Spec {
	s := Spec{
		First:           RoundSpec{Concurrency: 10, PerProvider: 2, Timeout: 60 * time.Second},
		Retry:           RoundSpec{Concurrency: 6, PerProvider: 1, Timeout: 90 * time.Second},
		Final:           RoundSpec{Concurrency: 4, PerProvider: 1, Timeout: 90 * time.Second},
		Resample:        2,
		MaxTokens:       1,
		Prompt:          "hi",
		Seed:            42,
		MaxCooldownWait: 120 * time.Second,
		MaxBackoff:      30 * time.Second,
	}
	switch name {
	case PresetQuick:
		s.Retry = RoundSpec{}
		s.Final = RoundSpec{}
		s.Resample = 0
		s.First.Timeout = 45 * time.Second
	case PresetDeep:
		s.Resample = 4
	}
	return s
}

// Normalize fills zero values with safe defaults.
func (s *Spec) Normalize() {
	if s.First.Concurrency <= 0 {
		s.First.Concurrency = 1
	}
	if s.First.PerProvider <= 0 {
		s.First.PerProvider = s.First.Concurrency
	}
	if s.First.Timeout <= 0 {
		s.First.Timeout = 60 * time.Second
	}
	for _, r := range []*RoundSpec{&s.Retry, &s.Final} {
		if r.Concurrency > 0 {
			if r.PerProvider <= 0 {
				r.PerProvider = 1
			}
			if r.Timeout <= 0 {
				r.Timeout = s.First.Timeout
			}
		}
	}
	if s.MaxTokens <= 0 {
		s.MaxTokens = 1
	}
	if s.Prompt == "" {
		s.Prompt = "hi"
	}
	if s.MaxCooldownWait <= 0 {
		s.MaxCooldownWait = 120 * time.Second
	}
	if s.MaxBackoff <= 0 {
		s.MaxBackoff = 30 * time.Second
	}
	if s.Resample < 0 {
		s.Resample = 0
	}
}

// Rounds returns the total number of rounds the spec will run.
func (s Spec) Rounds() int {
	n := 1
	if s.Retry.Concurrency > 0 {
		n++
	}
	n += s.Resample
	if s.Final.Concurrency > 0 {
		n++
	}
	return n
}

// Target is a model to probe.
type Target struct {
	ID       string // model id sent to the endpoint
	Provider string // provider key used for the per-provider limit
	Name     string // cleaned name used for modality inference
}

// Sink receives attempts as they complete.
type Sink interface {
	Attempt(ctx context.Context, a core.Attempt) error
}

// SinkFunc adapts a function to Sink.
type SinkFunc func(ctx context.Context, a core.Attempt) error

// Attempt implements Sink.
func (f SinkFunc) Attempt(ctx context.Context, a core.Attempt) error { return f(ctx, a) }

// Reporter receives progress events; it must not block.
type Reporter func(core.Event)

// Summary describes a finished scan.
type Summary struct {
	Rounds    []RoundSummary `json:"rounds"`
	Attempts  int            `json:"attempts"`
	Models    int            `json:"models"`
	Available int            `json:"available"`
	Elapsed   time.Duration  `json:"elapsed"`
}

// RoundSummary describes one round.
type RoundSummary struct {
	Round    int           `json:"round"`
	Kind     string        `json:"kind"` // first | retry | resample | final
	Targets  int           `json:"targets"`
	OK       int           `json:"ok"`
	Fail     int           `json:"fail"`
	Elapsed  time.Duration `json:"elapsed"`
	Skipped  int           `json:"skipped"` // already present (resume)
	Cooldown int           `json:"cooldown_waits"`
}

// Runner executes a scan.
type Runner struct {
	Client *apiclient.Client
	Spec   Spec
	Sink   Sink
	Report Reporter
	Stage  string // event stage label, default "scan"
	// Prior attempts (for resuming an interrupted run).
	Prior []core.Attempt
	// Now is overridable for tests.
	Now func() time.Time

	mu        sync.Mutex
	latest    map[string]core.Attempt
	okCount   map[string]int
	cooldowns map[string]time.Time // model id -> earliest time to retry
	backoff   map[string]time.Duration
	doneSet   map[int]map[string]bool
	attempts  int
}

// Run executes all rounds for the targets.
func (r *Runner) Run(ctx context.Context, targets []Target) (Summary, error) {
	r.Spec.Normalize()
	if r.Now == nil {
		r.Now = time.Now
	}
	if r.Stage == "" {
		r.Stage = "scan"
	}
	r.latest = map[string]core.Attempt{}
	r.okCount = map[string]int{}
	r.cooldowns = map[string]time.Time{}
	r.backoff = map[string]time.Duration{}
	r.doneSet = map[int]map[string]bool{}
	for _, a := range r.Prior {
		r.record(a, false)
	}
	start := r.Now()
	sum := Summary{Models: len(targets)}
	byID := make(map[string]Target, len(targets))
	for _, t := range targets {
		byID[t.ID] = t
	}
	round := 1
	plan := []struct {
		kind string
		spec RoundSpec
	}{{"first", r.Spec.First}}
	if r.Spec.Retry.Concurrency > 0 {
		plan = append(plan, struct {
			kind string
			spec RoundSpec
		}{"retry", r.Spec.Retry})
	}
	for i := 0; i < r.Spec.Resample; i++ {
		plan = append(plan, struct {
			kind string
			spec RoundSpec
		}{"resample", r.Spec.First})
	}
	if r.Spec.Final.Concurrency > 0 {
		plan = append(plan, struct {
			kind string
			spec RoundSpec
		}{"final", r.Spec.Final})
	}
	total := len(plan)
	for _, p := range plan {
		var list []Target
		switch p.kind {
		case "first":
			list = targets
		case "retry", "final":
			for _, t := range targets {
				if r.okCount[t.ID] == 0 {
					if la, ok := r.latest[t.ID]; ok && !la.OK && classify.IsTransient(la.Reason) {
						list = append(list, t)
					}
				}
			}
		case "resample":
			for _, t := range targets {
				if r.okCount[t.ID] > 0 {
					list = append(list, t)
				}
			}
		}
		rs, err := r.runRound(ctx, round, total, p.kind, p.spec, list, byID)
		sum.Rounds = append(sum.Rounds, rs)
		if err != nil {
			sum.Attempts = r.attempts
			sum.Elapsed = r.Now().Sub(start)
			return sum, err
		}
		round++
	}
	sum.Attempts = r.attempts
	for _, t := range targets {
		if r.okCount[t.ID] > 0 {
			sum.Available++
		}
	}
	sum.Elapsed = r.Now().Sub(start)
	r.emit(core.Event{Kind: core.EventDone, Stage: r.Stage, Done: sum.Attempts, Total: sum.Attempts, OK: sum.Available, Fail: sum.Models - sum.Available, Message: "scan complete"})
	return sum, nil
}

func (r *Runner) record(a core.Attempt, countNew bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.doneSet[a.Round] == nil {
		r.doneSet[a.Round] = map[string]bool{}
	}
	r.doneSet[a.Round][a.ModelID] = true
	if la, ok := r.latest[a.ModelID]; !ok || a.Round >= la.Round {
		r.latest[a.ModelID] = a
	}
	if a.OK {
		r.okCount[a.ModelID]++
	}
	if countNew {
		r.attempts++
	}
}

func (r *Runner) emit(ev core.Event) {
	if r.Report == nil {
		return
	}
	if ev.At.IsZero() {
		ev.At = r.Now()
	}
	r.Report(ev)
}

func (r *Runner) runRound(ctx context.Context, round, totalRounds int, kind string, spec RoundSpec, list []Target, byID map[string]Target) (RoundSummary, error) {
	rs := RoundSummary{Round: round, Kind: kind}
	start := r.Now()
	// Skip targets already recorded for this round (resume).
	var todo []Target
	for _, t := range list {
		if r.doneSet[round][t.ID] {
			rs.Skipped++
			continue
		}
		todo = append(todo, t)
	}
	rnd := rand.New(rand.NewSource(r.Spec.Seed + int64(round)*7919))
	rnd.Shuffle(len(todo), func(i, j int) { todo[i], todo[j] = todo[j], todo[i] })
	rs.Targets = len(todo)
	r.emit(core.Event{Kind: core.EventRound, Stage: r.Stage, Round: round, Rounds: totalRounds, Total: len(todo), Message: kind})
	if len(todo) == 0 {
		rs.Elapsed = r.Now().Sub(start)
		return rs, nil
	}
	global := semaphore.NewWeighted(int64(spec.Concurrency))
	var pmu sync.Mutex
	perProv := map[string]*semaphore.Weighted{}
	provSem := func(p string) *semaphore.Weighted {
		pmu.Lock()
		defer pmu.Unlock()
		s, ok := perProv[p]
		if !ok {
			s = semaphore.NewWeighted(int64(spec.PerProvider))
			perProv[p] = s
		}
		return s
	}
	var done, ok, fail, cooldownWaits int
	var cmu sync.Mutex
	g, gctx := errgroup.WithContext(ctx)
	for _, t := range todo {
		t := t
		g.Go(func() error {
			ps := provSem(t.Provider)
			if err := ps.Acquire(gctx, 1); err != nil {
				return err
			}
			defer ps.Release(1)
			if err := global.Acquire(gctx, 1); err != nil {
				return err
			}
			defer global.Release(1)
			waited, err := r.waitCooldown(gctx, t)
			if err != nil {
				return err
			}
			if waited {
				cmu.Lock()
				cooldownWaits++
				cmu.Unlock()
			}
			a := r.probe(gctx, t, round, spec.Timeout)
			if gctx.Err() != nil {
				return gctx.Err() // cancelled while in flight: do not record a bogus failure
			}
			r.record(a, true)
			r.noteOutcome(t, a)
			if r.Sink != nil {
				if err := r.Sink.Attempt(gctx, a); err != nil {
					return err
				}
			}
			cmu.Lock()
			done++
			if a.OK {
				ok++
			} else {
				fail++
			}
			d, o, f := done, ok, fail
			cmu.Unlock()
			el := r.Now().Sub(start)
			var eta float64
			if d > 0 {
				eta = el.Seconds() / float64(d) * float64(len(todo)-d)
			}
			r.emit(core.Event{Kind: core.EventAttempt, Stage: r.Stage, Round: round, Rounds: totalRounds, Done: d, Total: len(todo), OK: o, Fail: f, ETA: eta, Attempt: &a})
			return nil
		})
	}
	err := g.Wait()
	rs.OK, rs.Fail, rs.Cooldown = ok, fail, cooldownWaits
	rs.Elapsed = r.Now().Sub(start)
	if err != nil {
		if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
			return rs, err
		}
		return rs, err
	}
	return rs, nil
}

// waitCooldown honours a pending cooldown for the model and the provider
// backoff. It returns whether it waited.
func (r *Runner) waitCooldown(ctx context.Context, t Target) (bool, error) {
	r.mu.Lock()
	until := r.cooldowns[t.ID]
	if pb := r.backoff[t.Provider]; pb > 0 {
		if pu := r.Now().Add(pb); pu.After(until) {
			until = pu
		}
	}
	r.mu.Unlock()
	now := r.Now()
	if !until.After(now) {
		return false, nil
	}
	wait := until.Sub(now)
	if wait > r.Spec.MaxCooldownWait {
		wait = r.Spec.MaxCooldownWait
	}
	r.emit(core.Event{Kind: core.EventCooldown, Stage: r.Stage, Message: t.ID, ETA: wait.Seconds()})
	timer := time.NewTimer(wait)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return true, ctx.Err()
	case <-timer.C:
		return true, nil
	}
}

func (r *Runner) noteOutcome(t Target, a core.Attempt) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if a.OK {
		r.backoff[t.Provider] = 0
		delete(r.cooldowns, t.ID)
		return
	}
	if a.Reason == "rate_limit" || a.Reason == "quota_exhausted" || a.HTTP == 429 || a.Inner == 429 {
		if d, ok := classify.ResetHint(a.Message); ok && d > 0 {
			if d > r.Spec.MaxCooldownWait {
				d = r.Spec.MaxCooldownWait
			}
			r.cooldowns[t.ID] = r.Now().Add(d)
		}
		if a.Reason == "rate_limit" {
			b := r.backoff[t.Provider]
			if b == 0 {
				b = time.Second
			} else {
				b *= 2
			}
			if b > r.Spec.MaxBackoff {
				b = r.Spec.MaxBackoff
			}
			r.backoff[t.Provider] = b
		}
	} else if d, ok := classify.ResetHint(a.Message); ok && d > 0 && classify.IsTransient(a.Reason) {
		if d > r.Spec.MaxCooldownWait {
			d = r.Spec.MaxCooldownWait
		}
		r.cooldowns[t.ID] = r.Now().Add(d)
	}
}

// probe sends one minimal chat completion and classifies the result.
func (r *Runner) probe(ctx context.Context, t Target, round int, timeout time.Duration) core.Attempt {
	pctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	res := r.Client.Chat(pctx, apiclient.ChatRequest{
		Model:     t.ID,
		Messages:  []apiclient.Message{{Role: "user", Content: r.Spec.Prompt}},
		MaxTokens: r.Spec.MaxTokens,
	})
	return AttemptFrom(res, t, round, r.Now())
}

// AttemptFrom converts a chat result into an attempt record.
func AttemptFrom(res apiclient.ChatResult, t Target, round int, at time.Time) core.Attempt {
	out := classify.Evaluate(res.Status, res.Body, t.Name)
	a := core.Attempt{
		ModelID: t.ID, Round: round, At: at, HTTP: res.Status, Inner: out.Inner, Conn: out.Conn,
		Latency: res.Latency, OK: out.OK, Class: out.Class, Reason: out.Reason,
		Message: classify.CleanMessage(out.Message), ServedModel: res.Model, FinishReason: res.FinishReason,
	}
	if res.Usage != nil {
		a.PromptTokens, a.CompletionTokens = res.Usage.PromptTokens, res.Usage.CompletionTokens
	}
	return a
}

// Latest returns the most recent attempt per model after Run.
func (r *Runner) Latest() map[string]core.Attempt {
	r.mu.Lock()
	defer r.mu.Unlock()
	out := make(map[string]core.Attempt, len(r.latest))
	for k, v := range r.latest {
		out[k] = v
	}
	return out
}

// SortTargets orders targets deterministically by provider then id.
func SortTargets(ts []Target) {
	sort.Slice(ts, func(i, j int) bool {
		if ts[i].Provider != ts[j].Provider {
			return ts[i].Provider < ts[j].Provider
		}
		return ts[i].ID < ts[j].ID
	})
}
