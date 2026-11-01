package capability

import (
	"context"
	"testing"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/apiclient"
	"github.com/AnnaofArendelle/llm-speedtest/internal/testutil/mockllm"
)

func TestProbes(t *testing.T) {
	models := []mockllm.ModelSpec{
		{ID: "a/full", OwnedBy: "a", Behavior: mockllm.Behavior{Content: "PINEAPPLE", ToolCall: true, JSONMode: true, Vision: true, Embedding: true, MaxContext: 20000}},
		{ID: "a/plain", OwnedBy: "a", Behavior: mockllm.Behavior{Content: "1 2 3", NoUsage: true, IgnoreStream: true, MaxContext: 4000}},
		{ID: "b/missing", OwnedBy: "b", Behavior: mockllm.Behavior{Status: 404, Error: "The model 'missing' does not exist"}},
	}
	srv := mockllm.New(models)
	defer srv.Close()
	cl, _ := apiclient.New(srv.URL+"/v1", "", apiclient.Options{})
	spec := Default()
	spec.Timeout = 5 * time.Second
	spec.LongContextTiers = []int{8000, 32000}
	r := &Runner{Client: cl, Spec: spec}
	var targets []Target
	for _, m := range models {
		targets = append(targets, Target{ID: m.ID, Provider: m.OwnedBy, Name: m.ID})
	}
	results, err := r.Run(context.Background(), targets)
	if err != nil {
		t.Fatal(err)
	}
	mx := Summarize(results, spec.Probes)
	full := mx.Cells["a/full"]
	for _, p := range []string{"streaming", "usage_in_stream", "tools", "json_mode", "json_schema", "vision", "system_prompt", "max_tokens", "determinism", "reasoning_param", "long_context", "embeddings"} {
		if full[p].Status != Supported {
			t.Errorf("full %s = %+v", p, full[p])
		}
	}
	if full["long_context"].Evidence == "" || full["long_context"].Status != Supported {
		t.Errorf("long context = %+v", full["long_context"])
	}
	plain := mx.Cells["a/plain"]
	if plain["streaming"].Status != Unsupported || plain["usage_in_stream"].Status != Unsupported || plain["tools"].Status != Unsupported || plain["vision"].Status != Unsupported || plain["embeddings"].Status != Unsupported || plain["system_prompt"].Status != Unsupported {
		t.Errorf("plain = %+v", plain)
	}
	if plain["long_context"].Status != Unsupported {
		t.Errorf("plain long context = %+v", plain["long_context"])
	}
	if mx.Cells["b/missing"]["tools"].Status != Error || mx.Support["b/missing"] != 0 {
		t.Errorf("missing = %+v", mx.Cells["b/missing"])
	}
	if mx.Models[0] != "a/full" || mx.Support["a/full"] < 11 {
		t.Errorf("ordering = %v support=%v", mx.Models, mx.Support)
	}
	if Describe("tools", "zh") == "" || Describe("tools", "en") == "" {
		t.Error("describe")
	}
	if parseJSONObject("prefix {\"a\":1} suffix")["a"] != float64(1) || parseJSONObject("nope") != nil {
		t.Error("parse json")
	}
}
