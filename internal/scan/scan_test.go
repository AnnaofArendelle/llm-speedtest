package scan

import (
	"context"
	"errors"
	"runtime"
	"sync"
	"testing"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/apiclient"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/modelid"
	"github.com/AnnaofArendelle/llm-speedtest/internal/testutil/mockllm"
)

func targetsFor(models []mockllm.ModelSpec) []Target {
	var ts []Target
	for _, m := range models {
		ts = append(ts, Target{ID: m.ID, Provider: m.OwnedBy, Name: modelid.Name(m.ID, m.OwnedBy)})
	}
	return ts
}

type memSink struct {
	mu   sync.Mutex
	list []core.Attempt
}

func (s *memSink) Attempt(_ context.Context, a core.Attempt) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.list = append(s.list, a)
	return nil
}

func (s *memSink) byModel(id string) []core.Attempt {
	s.mu.Lock()
	defer s.mu.Unlock()
	var out []core.Attempt
	for _, a := range s.list {
		if a.ModelID == id {
			out = append(out, a)
		}
	}
	return out
}

func testSpec() Spec {
	return Spec{
		First:           RoundSpec{Concurrency: 8, PerProvider: 2, Timeout: 400 * time.Millisecond},
		Retry:           RoundSpec{Concurrency: 4, PerProvider: 1, Timeout: 800 * time.Millisecond},
		Final:           RoundSpec{Concurrency: 4, PerProvider: 1, Timeout: 800 * time.Millisecond},
		Resample:        2,
		MaxTokens:       1,
		Seed:            7,
		MaxCooldownWait: 5 * time.Second,
		MaxBackoff:      2 * time.Second,
	}
}

func TestRunRounds(t *testing.T) {
	models := []mockllm.ModelSpec{
		{ID: "a/fast", OwnedBy: "a", Behavior: mockllm.Behavior{Latency: 20 * time.Millisecond, ServedModel: "fast-v2"}},
		{ID: "a/flaky", OwnedBy: "a", Behavior: mockllm.Behavior{Status: 503, Inner: 503, Error: "A server error occurred. Please try again.", FailFirst: 2}},
		{ID: "c/limited", OwnedBy: "c", Behavior: mockllm.Behavior{Status: 429, Inner: 429, Error: "Rate limit exceeded", ResetAfter: 1500 * time.Millisecond, FailFirst: 1}},
		{ID: "d/auth", OwnedBy: "d", Behavior: mockllm.Behavior{Status: 503, Inner: 401, Error: "Your authentication token is invalid."}},
		{ID: "e/missing", OwnedBy: "e", Behavior: mockllm.Behavior{Status: 404, Error: "The model 'missing' does not exist"}},
		{ID: "f/hang", OwnedBy: "f", Behavior: mockllm.Behavior{Hang: 3 * time.Second}},
		{ID: "g/quota", OwnedBy: "g", Behavior: mockllm.Behavior{Status: 503, Inner: 402, Error: "Insufficient funds. Please add credits to continue."}},
	}
	srv := mockllm.New(models)
	defer srv.Close()
	cl, _ := apiclient.New(srv.URL+"/v1", "", apiclient.Options{})
	sink := &memSink{}
	var events []core.Event
	var emu sync.Mutex
	r := &Runner{Client: cl, Spec: testSpec(), Sink: sink, Report: func(e core.Event) {
		emu.Lock()
		events = append(events, e)
		emu.Unlock()
	}}
	sum, err := r.Run(context.Background(), targetsFor(models))
	if err != nil {
		t.Fatal(err)
	}
	if sum.Models != 7 || len(sum.Rounds) != 5 || sum.Available != 3 {
		t.Fatalf("summary = %+v", sum)
	}
	// fast: first + 2 resamples
	if as := sink.byModel("a/fast"); len(as) != 3 || !as[0].OK || as[0].ServedModel != "fast-v2" || as[2].Round != 4 {
		t.Fatalf("fast attempts = %+v", as)
	}
	// flaky: R1 fail, R2 fail, final success
	as := sink.byModel("a/flaky")
	if len(as) != 3 || as[0].OK || as[0].Reason != "upstream_5xx" || as[1].OK || !as[2].OK || as[2].Round != 5 {
		t.Fatalf("flaky attempts = %+v", as)
	}
	// limited: R1 429 with lock, R2 waits for the reset hint then succeeds, then 2 resamples
	as = sink.byModel("c/limited")
	if len(as) < 2 || as[0].Reason != "rate_limit" || as[0].Inner != 429 || !as[1].OK {
		t.Fatalf("limited attempts = %+v", as)
	}
	if gap := as[1].At.Sub(as[0].At); gap < 1200*time.Millisecond {
		t.Fatalf("retry did not wait for the cooldown hint: gap=%v", gap)
	}
	if sum.Rounds[1].Cooldown == 0 {
		t.Fatalf("expected a cooldown wait in the retry round: %+v", sum.Rounds[1])
	}
	// auth / missing / quota: not transient -> a single attempt
	for _, id := range []string{"d/auth", "e/missing", "g/quota"} {
		if as := sink.byModel(id); len(as) != 1 || as[0].OK {
			t.Fatalf("%s attempts = %+v", id, as)
		}
	}
	if sink.byModel("d/auth")[0].Reason != "auth" || sink.byModel("e/missing")[0].Reason != "model_unavailable" || sink.byModel("g/quota")[0].Reason != "quota_billing" {
		t.Fatalf("reasons: %s %s %s", sink.byModel("d/auth")[0].Reason, sink.byModel("e/missing")[0].Reason, sink.byModel("g/quota")[0].Reason)
	}
	// hang: timeout in R1, R2 and final
	as = sink.byModel("f/hang")
	if len(as) != 3 || as[0].Reason != "timeout" || as[2].Round != 5 || as[0].HTTP != 0 {
		t.Fatalf("hang attempts = %+v", as)
	}
	// events: rounds, attempts, done
	emu.Lock()
	defer emu.Unlock()
	var rounds, attempts, dones int
	for _, e := range events {
		switch e.Kind {
		case core.EventRound:
			rounds++
		case core.EventAttempt:
			attempts++
		case core.EventDone:
			dones++
		}
	}
	if rounds != 5 || attempts != sum.Attempts || dones != 1 {
		t.Fatalf("events rounds=%d attempts=%d dones=%d sum=%d", rounds, attempts, dones, sum.Attempts)
	}
}

func TestPerProviderLimitAndCancel(t *testing.T) {
	var models []mockllm.ModelSpec
	for i := 0; i < 8; i++ {
		models = append(models, mockllm.ModelSpec{ID: "p/m" + string(rune('a'+i)), OwnedBy: "p", Behavior: mockllm.Behavior{Latency: 120 * time.Millisecond}})
	}
	for i := 0; i < 4; i++ {
		models = append(models, mockllm.ModelSpec{ID: "q/m" + string(rune('a'+i)), OwnedBy: "q", Behavior: mockllm.Behavior{Latency: 120 * time.Millisecond}})
	}
	srv := mockllm.New(models)
	defer srv.Close()
	cl, _ := apiclient.New(srv.URL+"/v1", "", apiclient.Options{})
	spec := Preset(PresetQuick)
	spec.First = RoundSpec{Concurrency: 8, PerProvider: 2, Timeout: 2 * time.Second}
	r := &Runner{Client: cl, Spec: spec, Sink: &memSink{}}
	if _, err := r.Run(context.Background(), targetsFor(models)); err != nil {
		t.Fatal(err)
	}
	mi := srv.MaxInflight()
	if mi["p"] > 2 || mi["q"] > 2 || mi["p"] == 0 {
		t.Fatalf("per-provider limit violated: %+v", mi)
	}

	// cancellation: hang forever, cancel after a moment, no goroutine leak
	hang := []mockllm.ModelSpec{}
	for i := 0; i < 6; i++ {
		hang = append(hang, mockllm.ModelSpec{ID: "h/m" + string(rune('a'+i)), OwnedBy: "h", Behavior: mockllm.Behavior{Hang: 10 * time.Second}})
	}
	srv2 := mockllm.New(hang)
	defer srv2.Close()
	cl2, _ := apiclient.New(srv2.URL+"/v1", "", apiclient.Options{})
	before := runtime.NumGoroutine()
	ctx, cancel := context.WithCancel(context.Background())
	spec.First.Timeout = 30 * time.Second
	r2 := &Runner{Client: cl2, Spec: spec, Sink: &memSink{}}
	go func() {
		time.Sleep(150 * time.Millisecond)
		cancel()
	}()
	start := time.Now()
	_, err := r2.Run(ctx, targetsFor(hang))
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("expected cancellation, got %v", err)
	}
	if time.Since(start) > 3*time.Second {
		t.Fatalf("cancel took too long: %v", time.Since(start))
	}
	time.Sleep(200 * time.Millisecond)
	if after := runtime.NumGoroutine(); after > before+4 {
		t.Fatalf("goroutine leak: before=%d after=%d", before, after)
	}
}

func TestResume(t *testing.T) {
	models := []mockllm.ModelSpec{
		{ID: "a/x", OwnedBy: "a"}, {ID: "a/y", OwnedBy: "a"}, {ID: "b/z", OwnedBy: "b"},
	}
	srv := mockllm.New(models)
	defer srv.Close()
	cl, _ := apiclient.New(srv.URL+"/v1", "", apiclient.Options{})
	prior := []core.Attempt{{ModelID: "a/x", Round: 1, OK: true, Reason: "ok"}, {ModelID: "b/z", Round: 1, OK: false, Reason: "auth", Class: "auth"}}
	spec := Preset(PresetQuick)
	spec.First.Timeout = time.Second
	sink := &memSink{}
	r := &Runner{Client: cl, Spec: spec, Sink: sink, Prior: prior}
	sum, err := r.Run(context.Background(), targetsFor(models))
	if err != nil {
		t.Fatal(err)
	}
	if sum.Rounds[0].Skipped != 2 || sum.Rounds[0].Targets != 1 || len(sink.list) != 1 || sink.list[0].ModelID != "a/y" {
		t.Fatalf("resume: %+v %+v", sum, sink.list)
	}
	if sum.Available != 2 { // a/x (prior) + a/y
		t.Fatalf("available = %d", sum.Available)
	}
	if c := srv.Counts(); c["a/x"] != 0 || c["b/z"] != 0 || c["a/y"] != 1 {
		t.Fatalf("counts = %+v", c)
	}
}

func TestPresetsAndNormalize(t *testing.T) {
	q := Preset(PresetQuick)
	if q.Rounds() != 1 || q.Retry.Concurrency != 0 {
		t.Fatalf("quick = %+v", q)
	}
	s := Preset(PresetStandard)
	if s.Rounds() != 5 {
		t.Fatalf("standard rounds = %d", s.Rounds())
	}
	d := Preset(PresetDeep)
	if d.Rounds() != 7 {
		t.Fatalf("deep rounds = %d", d.Rounds())
	}
	var z Spec
	z.Normalize()
	if z.First.Concurrency != 1 || z.MaxTokens != 1 || z.Prompt == "" || z.MaxCooldownWait == 0 {
		t.Fatalf("normalize = %+v", z)
	}
	ts := []Target{{ID: "b", Provider: "y"}, {ID: "a", Provider: "y"}, {ID: "c", Provider: "x"}}
	SortTargets(ts)
	if ts[0].ID != "c" || ts[1].ID != "a" {
		t.Fatal("sort")
	}
}
