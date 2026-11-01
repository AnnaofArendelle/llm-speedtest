package load

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/apiclient"
	"github.com/AnnaofArendelle/llm-speedtest/internal/testutil/mockllm"
)

func TestClosedLoop(t *testing.T) {
	srv := mockllm.New([]mockllm.ModelSpec{{ID: "p/m", OwnedBy: "p", Behavior: mockllm.Behavior{Latency: 30 * time.Millisecond, Content: strings.Repeat("w ", 8), ChunkDelay: 5 * time.Millisecond}}})
	defer srv.Close()
	cl, _ := apiclient.New(srv.URL+"/v1", "", apiclient.Options{})
	spec := Spec{Model: "p/m", Levels: []int{1, 3}, Duration: 400 * time.Millisecond, InputTokens: 30, OutputTokens: 16, Stream: true, Timeout: 5 * time.Second, SLO: SLO{TTFT: 2 * time.Second, TPOT: time.Second}}
	r := &Runner{Client: cl, Spec: spec, Name: "m"}
	results, reqs, err := r.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 2 || len(reqs) == 0 {
		t.Fatalf("results = %+v", results)
	}
	l1, l3 := results[0], results[1]
	if l1.Errors != 0 || l3.Errors != 0 || l1.OK < 3 || l3.OK <= l1.OK || l3.RPS <= l1.RPS {
		t.Fatalf("levels = %+v / %+v", l1, l3)
	}
	if l1.Goodput != 1 || l1.TTFT.Count != l1.OK || l1.ITL.Count == 0 || l1.OutputTokens != l1.OK*8 || l1.OutTPS <= 0 {
		t.Fatalf("l1 = %+v", l1)
	}
	if mi := srv.MaxInflight(); mi["p"] > 3 {
		t.Fatalf("inflight = %+v", mi)
	}
	// request cap
	spec.Requests = 4
	spec.Levels = []int{2}
	r = &Runner{Client: cl, Spec: spec}
	results, _, _ = r.Run(context.Background())
	if results[0].Requests != 4 {
		t.Fatalf("cap = %+v", results[0])
	}
	// non-stream mode
	spec.Stream = false
	spec.Requests = 3
	r = &Runner{Client: cl, Spec: spec}
	results, _, _ = r.Run(context.Background())
	if results[0].OK != 3 || results[0].TTFT.Count != 0 || results[0].OutputTokens != 24 {
		t.Fatalf("non-stream = %+v", results[0])
	}
}

func TestOpenLoopAndErrors(t *testing.T) {
	srv := mockllm.New([]mockllm.ModelSpec{
		{ID: "p/m", OwnedBy: "p", Behavior: mockllm.Behavior{Latency: 20 * time.Millisecond}},
		{ID: "p/bad", OwnedBy: "p", Behavior: mockllm.Behavior{Status: 429, Error: "Rate limit exceeded"}},
	})
	defer srv.Close()
	cl, _ := apiclient.New(srv.URL+"/v1", "", apiclient.Options{})
	spec := Spec{Model: "p/m", Mode: ModeRate, Rate: 40, Duration: 500 * time.Millisecond, MaxInflight: 4, Timeout: 5 * time.Second}
	r := &Runner{Client: cl, Spec: spec}
	results, reqs, err := r.Run(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(results) != 1 || results[0].Requests < 8 || results[0].Requests > 40 || results[0].Errors != 0 || results[0].Level != 4000 {
		t.Fatalf("open loop = %+v (%d reqs)", results[0], len(reqs))
	}
	spec.Model = "p/bad"
	spec.Mode = ModeConcurrency
	spec.Levels = []int{2}
	spec.Requests = 4
	r = &Runner{Client: cl, Spec: spec, Name: "bad"}
	results, _, _ = r.Run(context.Background())
	if results[0].Errors != 4 || results[0].Reasons["rate_limit"] != 4 || results[0].Goodput != 0 {
		t.Fatalf("errors = %+v", results[0])
	}
	// SLO violation -> goodput 0
	spec.Model = "p/m"
	spec.SLO = SLO{E2E: time.Nanosecond}
	r = &Runner{Client: cl, Spec: spec}
	results, _, _ = r.Run(context.Background())
	if results[0].OK != 4 || results[0].Goodput != 0 {
		t.Fatalf("slo = %+v", results[0])
	}
	// cancellation
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()
	spec.Requests = 0
	spec.Duration = 10 * time.Second
	r = &Runner{Client: cl, Spec: spec}
	start := time.Now()
	if _, _, err := r.Run(ctx); err == nil || time.Since(start) > 3*time.Second {
		t.Fatalf("cancel err=%v", err)
	}
	var z Spec
	z.Normalize()
	if z.Mode != ModeConcurrency || len(z.Levels) == 0 || z.MaxInflight == 0 {
		t.Fatal("normalize")
	}
}
