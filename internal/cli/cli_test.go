package cli

import (
	"bytes"
	"encoding/json"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/testutil/mockllm"
)

func newMock(t *testing.T) string {
	t.Helper()
	llm := mockllm.New([]mockllm.ModelSpec{
		{ID: "a/fast", OwnedBy: "a", Behavior: mockllm.Behavior{Latency: 5 * time.Millisecond, Content: "one two three four five", ChunkDelay: time.Millisecond, ToolCall: true, JSONMode: true, Vision: true, Embedding: true}},
		{ID: "a/slow", OwnedBy: "a", Behavior: mockllm.Behavior{Latency: 20 * time.Millisecond, Content: "ok"}},
		{ID: "b/missing", OwnedBy: "b", Behavior: mockllm.Behavior{Status: 404, Error: "The model 'missing' does not exist"}},
		{ID: "combo/x", OwnedBy: "combo", Behavior: mockllm.Behavior{Content: "ok"}},
	})
	t.Cleanup(llm.Close)
	return llm.URL + "/v1"
}

func run(t *testing.T, args ...string) (int, string, string) {
	t.Helper()
	var out, errb bytes.Buffer
	code := Main(args, &out, &errb)
	return code, out.String(), errb.String()
}

func mustContain(t *testing.T, s, want string) {
	t.Helper()
	if !strings.Contains(s, want) {
		t.Fatalf("output missing %q:\n%s", want, s)
	}
}

func TestVersionAndUsage(t *testing.T) {
	code, out, _ := run(t, "version")
	if code != 0 {
		t.Fatalf("version exit %d", code)
	}
	mustContain(t, out, "llm-speedtest")

	code, out, _ = run(t, "version", "--json")
	if code != 0 {
		t.Fatalf("version --json exit %d", code)
	}
	var v map[string]any
	if err := json.Unmarshal([]byte(out), &v); err != nil || v["version"] == nil {
		t.Fatalf("version --json: %v %s", err, out)
	}

	if code, out, _ := run(t); code != 0 || !strings.Contains(out, "Usage") {
		t.Fatalf("no args: exit %d, stdout %q", code, out)
	}
	if code, _, errb := run(t, "bogus"); code != 2 || !strings.Contains(errb, "Usage") {
		t.Fatalf("unknown command: exit %d, stderr %q", code, errb)
	}
	if code, _, _ := run(t, "bogus"); code != 2 {
		t.Fatalf("unknown command exit %d", code)
	}
	if code, _, _ := run(t, "scan"); code != 2 {
		t.Fatalf("scan without --url exit %d", code)
	}
	if code, _, _ := run(t, "load", "--url", "http://127.0.0.1:1/v1", "--model", "x"); code == 0 {
		t.Fatal("load without --confirm must fail")
	}
}

func TestConfigAndSkill(t *testing.T) {
	code, out, _ := run(t, "config", "example")
	if code != 0 {
		t.Fatalf("config example exit %d", code)
	}
	mustContain(t, out, "suite:")
	mustContain(t, out, "scoring:")

	code, out, _ = run(t, "skill", "show")
	if code != 0 {
		t.Fatalf("skill show exit %d", code)
	}
	mustContain(t, out, "llm-speedtest")

	dir := t.TempDir()
	zip := filepath.Join(dir, "skill.zip")
	if code, _, errb := run(t, "skill", "export", zip); code != 0 {
		t.Fatalf("skill export exit %d: %s", code, errb)
	}
	if st, err := os.Stat(zip); err != nil || st.Size() == 0 {
		t.Fatalf("skill zip: %v", err)
	}
	target := filepath.Join(dir, "skills")
	if code, _, errb := run(t, "skill", "install", "--target", target); code != 0 {
		t.Fatalf("skill install exit %d: %s", code, errb)
	}
	if _, err := os.Stat(filepath.Join(target, "llm-speedtest", "SKILL.md")); err != nil {
		t.Fatalf("installed SKILL.md: %v", err)
	}
	if code, _, _ := run(t, "skill", "install", "--target", target); code == 0 {
		t.Fatal("second install without --force should fail")
	}
	if code, _, errb := run(t, "skill", "install", "--target", target, "--force"); code != 0 {
		t.Fatalf("forced install exit %d: %s", code, errb)
	}
}

func TestCheckAndDiscover(t *testing.T) {
	base := newMock(t)
	code, out, errb := run(t, "check", "--url", base, "--json")
	if code != 0 {
		t.Fatalf("check exit %d: %s", code, errb)
	}
	var chk struct {
		Reachable bool `json:"reachable"`
		AuthOK    bool `json:"auth_ok"`
		Models    int  `json:"models"`
	}
	if err := json.Unmarshal([]byte(out), &chk); err != nil || !chk.Reachable || !chk.AuthOK || chk.Models != 4 {
		t.Fatalf("check json: %v %+v", err, chk)
	}
	code, out, _ = run(t, "check", "--url", base)
	if code != 0 || !strings.Contains(out, "4") {
		t.Fatalf("check text: %d %s", code, out)
	}

	code, out, errb = run(t, "discover", "--url", base, "--json")
	if code != 0 {
		t.Fatalf("discover exit %d: %s", code, errb)
	}
	var disc struct {
		Total    int `json:"total"`
		Filtered int `json:"filtered"`
	}
	if err := json.Unmarshal([]byte(out), &disc); err != nil || disc.Total != 4 || disc.Filtered != 3 {
		t.Fatalf("discover json: %v %+v", err, disc)
	}
	code, out, _ = run(t, "discover", "--url", base, "--include-owner", "b")
	if code != 0 || !strings.Contains(out, "1 after filter") {
		t.Fatalf("discover filtered: %d %s", code, out)
	}
	if code, _, _ := run(t, "check", "--url", "http://127.0.0.1:1/v1"); code == 0 {
		t.Fatal("check against a closed port must fail")
	}
}

func TestScanReportCompare(t *testing.T) {
	base := newMock(t)
	data := t.TempDir()
	md, js, csv := filepath.Join(data, "r.md"), filepath.Join(data, "r.json"), filepath.Join(data, "r.csv")
	code, out, errb := run(t, "scan", "--url", base, "--profile", "quick", "--data-dir", data, "--lang", "en", "--out", md, "--json-out", js, "--csv-out", csv, "--name", "mock")
	if code != 0 {
		t.Fatalf("scan exit %d: %s %s", code, out, errb)
	}
	mustContain(t, out+errb, "available")
	for _, f := range []string{md, js, csv} {
		raw, err := os.ReadFile(f)
		if err != nil || len(raw) == 0 {
			t.Fatalf("export %s: %v", f, err)
		}
	}
	report, _ := os.ReadFile(md)
	mustContain(t, string(report), "a/fast")
	mustContain(t, string(report), "b/missing")

	code, out, _ = run(t, "runs", "--data-dir", data, "--json")
	if code != 0 {
		t.Fatalf("runs exit %d", code)
	}
	ids := regexp.MustCompile(`"id":\s*"([^"]+)"`).FindAllStringSubmatch(out, -1)
	if len(ids) == 0 {
		t.Fatalf("runs --json listed no run: %s", out)
	}
	id := ids[0][1]
	code, out, _ = run(t, "runs", "--data-dir", data)
	if code != 0 || !strings.Contains(out, id) {
		t.Fatalf("runs text: %d %s", code, out)
	}

	for _, format := range []string{"md", "json", "csv"} {
		code, out, errb = run(t, "report", "--data-dir", data, "--run", id, "--format", format, "--lang", "en")
		if code != 0 || out == "" {
			t.Fatalf("report %s exit %d: %s", format, code, errb)
		}
	}
	code, out, _ = run(t, "report", "--data-dir", data, "--out", filepath.Join(data, "latest.md"))
	if code != 0 {
		t.Fatalf("report latest exit %d: %s", code, out)
	}
	if _, err := os.Stat(filepath.Join(data, "latest.md")); err != nil {
		t.Fatal(err)
	}

	code, out, errb = run(t, "compare", "--data-dir", data, "--a", id, "--b", id, "--json")
	if code != 0 {
		t.Fatalf("compare exit %d: %s", code, errb)
	}
	mustContain(t, out, "schema_version")
	mustContain(t, out, "a/fast")
	code, out, _ = run(t, "compare", "--data-dir", data, "--a", id, "--b", id)
	if code != 0 {
		t.Fatalf("compare text exit %d: %s", code, out)
	}

	// 1 of 3 models fails, so a 99% availability gate must trip with exit code 5.
	code, _, _ = run(t, "scan", "--url", base, "--profile", "quick", "--no-store", "--quiet", "--fail-under", "0.99")
	if code != 5 {
		t.Fatalf("fail-under exit %d", code)
	}
}

func TestStagesAndSuite(t *testing.T) {
	base := newMock(t)
	common := []string{"--url", base, "--no-store", "--quiet"}
	if code, _, errb := run(t, append([]string{"latency", "--model", "a/fast", "--samples", "1", "--warmup", "0", "--output-tokens", "8"}, common...)...); code != 0 {
		t.Fatalf("latency exit %d: %s", code, errb)
	}
	if code, _, errb := run(t, append([]string{"cap", "--model", "a/fast", "--long-context", "1000"}, common...)...); code != 0 {
		t.Fatalf("cap exit %d: %s", code, errb)
	}
	if code, _, errb := run(t, append([]string{"quality", "--model", "a/fast", "--trials", "1"}, common...)...); code != 0 {
		t.Fatalf("quality exit %d: %s", code, errb)
	}
	if code, _, errb := run(t, append([]string{"load", "--model", "a/fast", "--confirm", "--levels", "1", "--duration", "300ms", "--requests", "2", "--output-tokens", "4"}, common...)...); code != 0 {
		t.Fatalf("load exit %d: %s", code, errb)
	}

	js := filepath.Join(t.TempDir(), "suite.json")
	md := filepath.Join(t.TempDir(), "suite.md")
	args := append([]string{"suite", "--profile", "quick", "--top", "1", "--samples", "1", "--warmup", "0", "--trials", "1", "--long-context", "1000", "--output-tokens", "8", "--json-out", js, "--out", md, "--lang", "en"}, common...)
	code, out, errb := run(t, args...)
	if code != 0 {
		t.Fatalf("suite exit %d: %s %s", code, out, errb)
	}
	raw, err := os.ReadFile(js)
	if err != nil {
		t.Fatal(err)
	}
	var doc map[string]any
	if err := json.Unmarshal(raw, &doc); err != nil || doc["summary"] == nil {
		t.Fatalf("suite json: %v", err)
	}
	report, _ := os.ReadFile(md)
	for _, want := range []string{"Latency", "Capabilit", "Quality"} {
		mustContain(t, string(report), want)
	}
}
