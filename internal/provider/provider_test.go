package provider

import (
	"strings"
	"testing"
)

func TestResolveOwnedByGateway(t *testing.T) {
	models := []Model{
		{"kilo/a", "kilo"}, {"kilo/b", "kilo"}, {"RouteWay/x", "RouteWay"}, {"RouteWay/y:free", "RouteWay"}, {"groq/z", "groq"},
	}
	conns := map[string]string{"kilo/a": "occ-1/a", "kilo/b": "occ-1/b", "RouteWay/x": "occ-2/x", "groq/z": "groq/z"}
	res := Resolve(models, conns, nil, "my-gateway")
	if res.Field != "owned_by" {
		t.Fatalf("field = %s (%s)", res.Field, res.Rationale)
	}
	if a := res.Models["RouteWay/y:free"]; a.Key != "RouteWay" || a.Name != "RouteWay" {
		t.Fatalf("assignment = %+v", a)
	}
	if len(res.Providers) != 3 || res.Providers[0].Key != "RouteWay" && res.Providers[0].Key != "kilo" {
		t.Fatalf("providers = %+v", res.Providers)
	}
	if res.Providers[0].Models != 2 {
		t.Fatalf("ordering = %+v", res.Providers)
	}
	if len(res.Suggestions) != 0 {
		t.Fatalf("unexpected suggestions %+v", res.Suggestions)
	}
	if !strings.Contains(res.Rationale, "consistent") {
		t.Fatalf("rationale = %s", res.Rationale)
	}
}

func TestResolvePrefixWhenOwnedByUniform(t *testing.T) {
	models := []Model{{"openai/gpt-4o", "openrouter"}, {"anthropic/claude-3", "openrouter"}, {"meta/llama", "openrouter"}}
	res := Resolve(models, nil, nil, "")
	if res.Field != "prefix" {
		t.Fatalf("field = %s", res.Field)
	}
	if res.Models["anthropic/claude-3"].Key != "anthropic" {
		t.Fatal("prefix key")
	}
}

func TestResolveSingleProvider(t *testing.T) {
	models := []Model{{"gpt-4o", "openai"}, {"gpt-4o-mini", "openai"}, {"text-embedding-3", "openai"}}
	res := Resolve(models, nil, nil, "OpenAI")
	if res.Field != "endpoint" || res.Models["gpt-4o"].Key != "OpenAI" {
		t.Fatalf("single provider: %+v", res)
	}
	res = Resolve([]Model{{"llama3", "library"}, {"qwen", "library"}}, nil, nil, "")
	if res.Field != "endpoint" || res.Models["llama3"].Name != "endpoint" {
		t.Fatalf("generic owner: %+v", res)
	}
}

func TestResolveOverridesAndSuggestions(t *testing.T) {
	models := []Model{{"a/1", "a"}, {"b/2", "b"}, {"c/3", "c"}, {"d/4", "d"}}
	conns := map[string]string{"a/1": "occ-9/1", "b/2": "occ-9/2", "c/3": "occ-7/3"}
	res := Resolve(models, conns, map[string]Override{"a": {MergedInto: "b"}, "b": {DisplayName: "B Gateway"}}, "")
	if a := res.Models["a/1"]; a.Key != "b" || a.Name != "B Gateway" {
		t.Fatalf("merged = %+v", a)
	}
	if a := res.Models["b/2"]; a.Name != "B Gateway" {
		t.Fatalf("renamed = %+v", a)
	}
	found := false
	for _, p := range res.Providers {
		if p.Key == "b" && p.Models == 2 && len(p.Raw) == 2 {
			found = true
		}
	}
	if !found {
		t.Fatalf("provider merge: %+v", res.Providers)
	}
	// without the override, a and b share a connection -> suggestion
	res = Resolve(models, conns, nil, "")
	if len(res.Suggestions) != 1 || strings.Join(res.Suggestions[0].Keys, ",") != "a,b" {
		t.Fatalf("suggestions = %+v", res.Suggestions)
	}
	if res.Field != "owned_by" {
		t.Fatalf("field = %s %s", res.Field, res.Rationale)
	}
}

func TestInconsistentOwnedByFallsBackToPrefix(t *testing.T) {
	// owned_by claims two providers but every model is served by the same connection,
	// while the id prefix is consistent with the connections.
	models := []Model{{"p1/a", "x"}, {"p1/b", "y"}, {"p2/c", "x"}, {"p2/d", "y"}}
	conns := map[string]string{"p1/a": "c1/a", "p1/b": "c1/b", "p2/c": "c2/c", "p2/d": "c2/d"}
	res := Resolve(models, conns, nil, "")
	if res.Field != "prefix" {
		t.Fatalf("field = %s (%s)", res.Field, res.Rationale)
	}
}
