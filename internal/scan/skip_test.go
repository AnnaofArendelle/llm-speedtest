package scan

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/apiclient"
	"github.com/AnnaofArendelle/llm-speedtest/internal/classify"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/testutil/mockllm"
)

// TestSkipModel skips a hanging model while the scan is in flight: the run
// must finish promptly, the model must be recorded as skipped (not as a
// timeout), later rounds must not touch it, and the other model is unaffected.
func TestSkipModel(t *testing.T) {
	llm := mockllm.New([]mockllm.ModelSpec{
		{ID: "a/fast", OwnedBy: "a", Behavior: mockllm.Behavior{Content: "ok"}},
		{ID: "a/hang", OwnedBy: "a", Behavior: mockllm.Behavior{Hang: 30 * time.Second, Content: "ok"}},
	})
	defer llm.Close()
	spec := Preset("standard")
	spec.First.Timeout, spec.Retry.Timeout, spec.Final.Timeout = 20*time.Second, 20*time.Second, 20*time.Second
	spec.Normalize()
	sk := core.NewSkipper()
	var mu sync.Mutex
	var starts []string
	cl, err := apiclient.New(llm.URL+"/v1", "", apiclient.Options{})
	if err != nil {
		t.Fatal(err)
	}
	r := &Runner{
		Client:  cl,
		Spec:    spec,
		Skipper: sk,
		Report: func(ev core.Event) {
			if ev.Kind == core.EventStart && ev.Attempt != nil {
				mu.Lock()
				starts = append(starts, ev.Attempt.ModelID)
				mu.Unlock()
			}
		},
	}
	targets := []Target{{ID: "a/fast", Provider: "a", Name: "fast"}, {ID: "a/hang", Provider: "a", Name: "hang"}}
	go func() {
		deadline := time.Now().Add(5 * time.Second)
		for time.Now().Before(deadline) {
			// Skip once the request has actually reached the mock server, so the
			// test also proves that an in-flight HTTP request gets cancelled.
			for _, it := range sk.Inflight() {
				if it.ID == "a/hang" && llm.Counts()["a/hang"] >= 1 {
					sk.Skip("a/hang")
					return
				}
			}
			time.Sleep(10 * time.Millisecond)
		}
	}()
	begin := time.Now()
	sum, err := r.Run(context.Background(), targets)
	if err != nil {
		t.Fatalf("run: %v", err)
	}
	if time.Since(begin) > 10*time.Second {
		t.Fatalf("run took %s; the skipped request should have been cancelled", time.Since(begin))
	}
	latest := r.Latest()
	if a := latest["a/hang"]; a.Reason != classify.ReasonSkipped || a.OK || a.Class != classify.ReasonSkipped {
		t.Fatalf("hang attempt = %+v", a)
	}
	if a := latest["a/fast"]; !a.OK {
		t.Fatalf("fast attempt = %+v", a)
	}
	if n := llm.Counts()["a/hang"]; n != 1 {
		t.Fatalf("skipped model was requested %d times; want exactly 1", n)
	}
	if llm.Counts()["a/fast"] < 2 {
		t.Fatalf("fast model should have been re-sampled: %d requests", llm.Counts()["a/fast"])
	}
	// A skipped model is a non-transient failure: it must not be a target of the
	// retry/final rounds, and only the available model is re-sampled.
	for _, rs := range sum.Rounds[1:] {
		if rs.Targets > 1 {
			t.Fatalf("round %d (%s) targeted %d models; the skipped model must be excluded: %+v", rs.Round, rs.Kind, rs.Targets, sum.Rounds)
		}
	}
	mu.Lock()
	defer mu.Unlock()
	seen := false
	for _, id := range starts {
		if id == "a/hang" {
			seen = true
		}
	}
	if !seen {
		t.Fatal("no start event for the skipped model")
	}
}
