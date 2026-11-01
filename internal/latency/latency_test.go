package latency

import (
	"context"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/apiclient"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/testutil/mockllm"
)

func TestRunAndSummarize(t *testing.T) {
	words := strings.Repeat("alpha beta gamma delta ", 5) // 20 words -> 20 chunks
	models := []mockllm.ModelSpec{
		{ID: "a/fast", OwnedBy: "a", Behavior: mockllm.Behavior{Latency: 50 * time.Millisecond, Content: words, ChunkDelay: 10 * time.Millisecond}},
		{ID: "a/slow", OwnedBy: "a", Behavior: mockllm.Behavior{Latency: 150 * time.Millisecond, Content: words, ChunkDelay: 20 * time.Millisecond, NoUsage: true}},
		{ID: "b/missing", OwnedBy: "b", Behavior: mockllm.Behavior{Status: 404, Error: "The model 'missing' does not exist"}},
	}
	srv := mockllm.New(models)
	defer srv.Close()
	cl, _ := apiclient.New(srv.URL+"/v1", "", apiclient.Options{})
	var sunk []core.LatencySample
	var smu sync.Mutex
	r := &Runner{Client: cl, Spec: Spec{Samples: 3, Warmup: 1, InputTokens: 50, OutputTokens: 64, Concurrency: 4, PerProvider: 2, Timeout: 5 * time.Second, StallTimeout: 2 * time.Second},
		Sink: SinkFunc(func(_ context.Context, s core.LatencySample) error {
			smu.Lock()
			defer smu.Unlock()
			sunk = append(sunk, s)
			return nil
		})}
	var targets []Target
	for _, m := range models {
		targets = append(targets, Target{ID: m.ID, Provider: m.OwnedBy, Name: m.ID})
	}
	samples, err := r.Run(context.Background(), targets)
	if err != nil {
		t.Fatal(err)
	}
	// fast: 4 samples (1 warm-up + 3); missing: 1 sample then stop (deterministic failure)
	count := map[string]int{}
	for _, s := range samples {
		count[s.ModelID]++
	}
	if count["a/fast"] != 4 || count["a/slow"] != 4 || count["b/missing"] != 1 || len(sunk) != 9 {
		t.Fatalf("counts = %+v sunk=%d", count, len(sunk))
	}
	stats := Summarize(samples)
	if len(stats) != 3 || stats[0].ModelID != "a/fast" || stats[1].ModelID != "a/slow" || stats[2].ModelID != "b/missing" {
		t.Fatalf("order = %v %v %v", stats[0].ModelID, stats[1].ModelID, stats[2].ModelID)
	}
	fast := stats[0]
	if fast.OK != 4 || fast.TTFT.Count != 3 || fast.Cold == nil || fast.Rank != 1 || !fast.TokensExact {
		t.Fatalf("fast = %+v", fast)
	}
	if fast.TTFT.Median < 40e-3 || fast.TTFT.Median > 400e-3 {
		t.Fatalf("fast ttft median = %v", fast.TTFT.Median)
	}
	// 20 chunks at 10ms -> ~100 tok/s; allow wide tolerance for CI machines
	if fast.TokPerSec.Median < 30 || fast.TokPerSec.Median > 130 {
		t.Fatalf("fast tok/s = %v", fast.TokPerSec.Median)
	}
	if fast.OutputTokens.Median != 20 || fast.TPOT.Median <= 0 || fast.TTST.Median <= 0 || fast.E2E.Median <= fast.TTFT.Median {
		t.Fatalf("fast stats = %+v", fast)
	}
	slow := stats[1]
	if slow.TokensExact || slow.TTFT.Median < fast.TTFT.Median || slow.E2E.Median < fast.E2E.Median {
		t.Fatalf("slow = %+v", slow)
	}
	if m := stats[2]; m.OK != 0 || m.Reasons["model_unavailable"] != 1 || m.Rank != 0 {
		t.Fatalf("missing = %+v", m)
	}
	if p := Synthesize(200, "abc"); apiclient.EstimateTokens(p) < 200 || !strings.Contains(p, "abc") {
		t.Fatal("synthesize")
	}
	var z Spec
	z.Normalize()
	if z.Samples != 5 || z.Timeout == 0 {
		t.Fatal("normalize")
	}
}

func TestCancel(t *testing.T) {
	srv := mockllm.New([]mockllm.ModelSpec{{ID: "h/x", OwnedBy: "h", Behavior: mockllm.Behavior{Hang: 5 * time.Second}}})
	defer srv.Close()
	cl, _ := apiclient.New(srv.URL+"/v1", "", apiclient.Options{})
	ctx, cancel := context.WithTimeout(context.Background(), 200*time.Millisecond)
	defer cancel()
	r := &Runner{Client: cl, Spec: Spec{Samples: 2, Timeout: 10 * time.Second}}
	start := time.Now()
	_, err := r.Run(ctx, []Target{{ID: "h/x", Provider: "h"}})
	if err == nil || time.Since(start) > 3*time.Second {
		t.Fatalf("cancel: err=%v elapsed=%v", err, time.Since(start))
	}
}
