package report

import (
	"encoding/json"
	"flag"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/analyze"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/testutil/refscan"
)

var update = flag.Bool("update", false, "rewrite golden files")

// subset keeps the reference scan small enough for a readable golden file:
// every model of a handful of providers that together exercise successes,
// flaky models and most failure classes.
func subset(t *testing.T) ([]core.RunModel, []core.Attempt) {
	t.Helper()
	models, attempts := refscan.Load(t)
	keep := map[string]bool{"groq": true, "cf": true, "thai": true, "cu": true, "nvidia": true, "stkj": true, "amd": true, "oc": true, "ps": true, "cohere": true}
	var ms []core.RunModel
	ids := map[string]bool{}
	for _, m := range models {
		if keep[m.OwnedBy] {
			ms = append(ms, m)
			ids[m.ID] = true
		}
	}
	var as []core.Attempt
	for _, a := range attempts {
		if ids[a.ModelID] {
			as = append(as, a)
		}
	}
	return ms, as
}

func fixedMeta(lang string) Meta {
	ts := time.Date(2026, 11, 1, 22, 12, 58, 0, time.FixedZone("CST", 8*3600))
	return Meta{Title: "", Endpoint: "http://localhost:20128/v1", Dialect: "9router", Version: "test", Generated: ts.Add(45 * time.Minute),
		Started: ts, Finished: ts.Add(29 * time.Minute), Preset: "standard", Lang: lang, ExcludeNote: "excluded 14 combo models",
		Rounds: []RoundLine{{1, "first", 100, 20, 80}, {2, "retry", 10, 2, 8}}}
}

func TestMarkdownGolden(t *testing.T) {
	ms, as := subset(t)
	if len(ms) < 50 || len(ms) > 200 {
		t.Fatalf("subset size %d", len(ms))
	}
	pres := analyze.Enrich(ms, as, analyze.Options{EndpointName: "9router"})
	res := analyze.Build(ms, as, analyze.DefaultScoring(), &pres)
	for _, lang := range []string{"zh", "en"} {
		got := Markdown(res, fixedMeta(lang))
		golden := filepath.Join("..", "..", "testdata", "report."+lang+".golden.md")
		if *update {
			if err := os.WriteFile(golden, []byte(got), 0o644); err != nil {
				t.Fatal(err)
			}
			continue
		}
		want, err := os.ReadFile(golden)
		if err != nil {
			t.Fatalf("missing golden file (run with -update): %v", err)
		}
		if string(want) != got {
			line := firstDiffLine(string(want), got)
			t.Fatalf("%s report differs from golden at line %d; run `go test ./internal/report -update` after reviewing", lang, line)
		}
	}
	md := Markdown(res, fixedMeta("zh"))
	for _, must := range []string{"## 1. 结论摘要", "### 1.1", "### 1.3", "## 3. 提供商总览", "### 4.1 综合排名", "### 4.2", "### 4.3", "### 4.4", "## 5.", "### 5.1", "## 6.", "局限性"} {
		if !strings.Contains(md, must) {
			t.Errorf("missing section %q", must)
		}
	}
	if strings.Contains(md, "%!") {
		t.Error("formatting directive leaked into report")
	}
}

func firstDiffLine(a, b string) int {
	al, bl := strings.Split(a, "\n"), strings.Split(b, "\n")
	for i := 0; i < len(al) && i < len(bl); i++ {
		if al[i] != bl[i] {
			return i + 1
		}
	}
	return min(len(al), len(bl)) + 1
}

func TestJSONAndCSV(t *testing.T) {
	ms, as := subset(t)
	pres := analyze.Enrich(ms, as, analyze.Options{})
	res := analyze.Build(ms, as, analyze.DefaultScoring(), &pres)
	raw, err := JSON(res, fixedMeta("en"))
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatal(err)
	}
	if doc["schema_version"] != float64(SchemaVersion) || doc["summary"] == nil || doc["families"] == nil {
		t.Fatalf("json doc keys: %v", doc)
	}
	csvb, err := CSV(res)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(csvb)), "\n")
	if len(lines) != len(ms)+1 || !strings.HasPrefix(lines[0], "model_id,provider") {
		t.Fatalf("csv lines = %d", len(lines))
	}
}

func TestHelpers(t *testing.T) {
	if esc("a|b\nc") != "a\\|b c" || trunc("abcdef", 4) != "abc…" || fs(0) != "-" || pct(1, 0) != "-" || fctx(262144) != "262k" || fctx(0) != "-" || jitter(0) != "-" {
		t.Fatal("helpers")
	}
}
