package family

import (
	"testing"

	"github.com/AnnaofArendelle/llm-speedtest/internal/refdata"
	"github.com/AnnaofArendelle/llm-speedtest/internal/testutil/refscan"
)

type listItem = refdata.ListItem

func loadList(t *testing.T) []listItem {
	t.Helper()
	return refscan.List(t)
}

func TestKeyExamples(t *testing.T) {
	cases := map[string]string{
		"deepseek-v4-flash":                            "deepseek|flash|4",
		"deepseek-v4.1-flash":                          "deepseek|flash|4",
		"deepseek-v4-flash-0731":                       "deepseek|flash|4",
		"deepseek-v4-flash-exp":                        "deepseek|flash|4",
		"deepseek-v4-flash-preview":                    "deepseek|flash|4",
		"deepseek-v4-flash:free":                       "deepseek|flash|4",
		"RouteWay/deepseek-v4.1-flash-1017-fast":       "deepseek|flash|4",
		"xjhc/DeepSeek-V4-Flash":                       "deepseek|flash|4",
		"deepseek-v4-pro-0813":                         "deepseek|pro|4",
		"deepseek-chat":                                "deepseek||",
		"gpt-oss-120b":                                 "gpt-oss||",
		"gpt-oss:20b":                                  "gpt-oss||",
		"openai/gpt-oss-20b":                           "gpt-oss||",
		"xkiro/cohere/command-r7b-12-2024":             "command-r||",
		"command-r-08-2024":                            "command-r||",
		"gpt-5-search-api-2025-10-14":                  "gpt|search|5",
		"gpt-4o-mini-2024-07-18":                       "gpt|mini|4",
		"claude-haiku-4-5-20251001":                    "claude|haiku|4",
		"claude-4.6-opus-max":                          "claude|opus|4",
		"claude-sonnet-5.5":                            "claude|sonnet|5",
		"flux.2-pro":                                   "flux|pro|2",
		"laguna-m.1:free":                              "laguna||1",
		"bge-base-en-v1.5:free":                        "bge||1",
		"glm-4.5v":                                     "glm|vision|4",
		"o4-mini":                                      "o|mini|4",
		"qwen3.8-max":                                  "qwen|max|3",
		"Qwen3-14B":                                    "qwen||3",
		"MiniCPM-V-4.7-35B-A3B":                        "minicpm|vision|4",
		"NVIDIA-Nemotron-3.5-Super-VL-120B-A12B-NVFP4": "nemotron|super+vl|3",
		"kimi-k3.1-pro-highspeed":                      "kimi|pro|3",
		"minimax-m2.7-highspeed:free":                  "minimax||2",
		"hy3":                                          "hy||3",
		"kilo-auto/free":                               "kilo-auto-free||",
		"mixtral-8x22b-instruct":                       "mixtral||",
		"dsv41flash":                                   "dsv|flash|41",
		"l3-lunaris-8b":                                "l||3",
	}
	for id, want := range cases {
		if got := Key(id, "").Key; got != want {
			t.Errorf("Key(%q) = %q, want %q (canon %q)", id, got, want, Key(id, "").Canon)
		}
	}
	if c := Key("deepseek-v4.1-flash", "").Canon; c != "deepseek-v4-flash" {
		t.Errorf("canon = %q", c)
	}
	if c := Key("gpt-4o-mini", "").Canon; c != "gpt-4o-mini" {
		t.Errorf("canon 4o = %q", c)
	}
	if c := Key("qwen3.8-max", "").Canon; c != "qwen3-max" {
		t.Errorf("canon qwen = %q", c)
	}
	if Display([]string{"a", "b", "b", "ccc"}) != "b" || Display([]string{"zz", "a"}) != "a" || Display(nil) != "" {
		t.Error("display")
	}
}

func TestGoldenList(t *testing.T) {
	items := loadList(t)
	if len(items) != 2306 {
		t.Fatalf("expected 2306 ids, got %d", len(items))
	}
	fams := map[string][]string{}
	for _, it := range items {
		f := Key(it.ID, it.OwnedBy)
		if f.Key == "" || f.Canon == "" {
			t.Fatalf("empty key for %q", it.ID)
		}
		fams[f.Key] = append(fams[f.Key], it.ID)
	}
	if len(fams) < 480 || len(fams) > 540 {
		t.Fatalf("family count drifted: %d", len(fams))
	}
	ds := fams["deepseek|flash|4"]
	if len(ds) < 75 {
		t.Fatalf("deepseek-v4-flash family too small: %d", len(ds))
	}
	for _, want := range []string{"RouteWay/deepseek-v4.1-flash-1017-fast", "xjhc/DeepSeek-V4-Flash", "kimchi/deepseek-v4-flash-0731", "thai/deepseek-v4.1-flash:free"} {
		found := false
		for _, id := range ds {
			if id == want {
				found = true
			}
		}
		if !found {
			t.Errorf("%s not in deepseek-v4-flash family", want)
		}
	}
	if n := len(fams["qwen||3"]); n < 80 {
		t.Errorf("qwen3 family = %d", n)
	}
	if n := len(fams["gpt-oss||"]); n < 10 {
		t.Errorf("gpt-oss family = %d", n)
	}
}

func TestLearn(t *testing.T) {
	items := loadList(t)
	models := make([]Model, 0, len(items))
	for _, it := range items {
		models = append(models, Model{ID: it.ID, OwnedBy: it.OwnedBy})
	}
	served := map[string]string{
		"gptaifree/deepseek-chat":    "deepseek-v4-flash-0731", // alias resolved by the upstream
		"amd/DeepSeek-V4-Flash":      "self-dploy/DeepSeek-V4-Flash",
		"orcarouter/auto":            "gpt-4o-mini", // generic names never follow served names
		"orcarouter/orcarouter/free": "deepseek-v4-flash-ga-260731",
	}
	as := Learn(nil, models, served, map[string]Override{"vyceai/deepseek-v4-flash-lr": {Key: "custom|x|1", Canon: "custom-x"}})
	if a := as["gptaifree/deepseek-chat"]; a.Key != "deepseek|flash|4" || a.Source != SourceServed || a.Canon != "deepseek-v4-flash" {
		t.Errorf("served learning: %+v", a)
	}
	if a := as["amd/DeepSeek-V4-Flash"]; a.Source != SourceRule || a.Key != "deepseek|flash|4" {
		t.Errorf("same family stays rule-based: %+v", a)
	}
	if a := as["orcarouter/auto"]; a.Source == SourceServed {
		t.Errorf("generic name followed served: %+v", a)
	}
	if a := as["orcarouter/orcarouter/free"]; a.Source != SourceRule || !a.Rule.Generic {
		t.Errorf("generic pseudo-model must stay rule-based: %+v", a)
	}
	// abbreviation expansion learned from the endpoint's own vocabulary
	for _, id := range []string{"RouteWay/dsv41flash", "unorouter/dsv41flash", "kios/dsv41flash"} {
		if a, ok := as[id]; ok {
			if a.Key != "deepseek|flash|4" || a.Source != SourceAbbrev {
				t.Errorf("abbrev %s: %+v", id, a)
			}
		}
	}
	abbrevSeen := false
	for id, a := range as {
		if a.Source == SourceAbbrev {
			abbrevSeen = true
			if a.Rule.Base == "l" && a.Key != "llama||3" {
				t.Errorf("l3 -> %s (%s)", a.Key, id)
			}
		}
	}
	if !abbrevSeen {
		t.Error("no abbreviation learned at all")
	}
	if a := as["vyceai/deepseek-v4-flash-lr"]; a.Key != "custom|x|1" || a.Source != SourceUser || a.Canon != "custom-x" {
		t.Errorf("override: %+v", a)
	}
	// determinism
	as2 := Learn(nil, models, served, nil)
	for id, a := range as2 {
		if b := as2[id]; a.Key != b.Key {
			t.Fatal("non-deterministic")
		}
	}
}

func FuzzKey(f *testing.F) {
	for _, s := range []string{"deepseek-v4.1-flash:free", "gpt-4o", "", "///", "a.b.c-1.2.3", "ＡＢＣ-１", "x:y:z", "[1m]", "-", "v", "4o", "r7b"} {
		f.Add(s, "owner")
	}
	f.Fuzz(func(t *testing.T, id, owner string) {
		a := Key(id, owner)
		b := Key(id, owner)
		if a.Key != b.Key || a.Canon != b.Canon {
			t.Fatal("non-deterministic")
		}
		if a.Key == "" {
			t.Fatal("empty key")
		}
	})
}
