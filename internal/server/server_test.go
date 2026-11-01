package server

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/app"
	"github.com/AnnaofArendelle/llm-speedtest/internal/config"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/scan"
	"github.com/AnnaofArendelle/llm-speedtest/internal/secrets"
	"github.com/AnnaofArendelle/llm-speedtest/internal/store"
	"github.com/AnnaofArendelle/llm-speedtest/internal/testutil/mockllm"
)

type env struct {
	t    *testing.T
	api  *httptest.Server
	llm  *mockllm.Server
	svc  *app.Service
	tok  string
	base string
}

func newEnv(t *testing.T, token string) *env {
	t.Helper()
	dir := t.TempDir()
	codec, _ := secrets.Load(filepath.Join(dir, ".secret"))
	st, err := store.Open(filepath.Join(dir, "db.sqlite"), codec)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	svc := app.New(st, "test")
	llm := mockllm.New([]mockllm.ModelSpec{
		{ID: "a/fast", OwnedBy: "a", Behavior: mockllm.Behavior{Latency: 10 * time.Millisecond, Content: "one two three", ChunkDelay: 2 * time.Millisecond, ToolCall: true, JSONMode: true, Vision: true, Embedding: true}},
		{ID: "a/slow", OwnedBy: "a", Behavior: mockllm.Behavior{Latency: 40 * time.Millisecond, Content: "ok"}},
		{ID: "b/missing", OwnedBy: "b", Behavior: mockllm.Behavior{Status: 404, Error: "The model 'missing' does not exist"}},
	})
	t.Cleanup(llm.Close)
	srv := New(svc, Options{AccessToken: token})
	api := httptest.NewServer(srv.Handler())
	t.Cleanup(api.Close)
	return &env{t: t, api: api, llm: llm, svc: svc, tok: token, base: llm.URL + "/v1"}
}

func (e *env) do(method, path string, body any) (int, []byte) {
	e.t.Helper()
	var rd io.Reader
	if body != nil {
		raw, _ := json.Marshal(body)
		rd = strings.NewReader(string(raw))
	}
	req, _ := http.NewRequest(method, e.api.URL+path, rd)
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if e.tok != "" {
		req.Header.Set("Authorization", "Bearer "+e.tok)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		e.t.Fatal(err)
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	return resp.StatusCode, raw
}

func (e *env) json(method, path string, body any, out any) int {
	e.t.Helper()
	code, raw := e.do(method, path, body)
	if out != nil {
		if err := json.Unmarshal(raw, out); err != nil {
			e.t.Fatalf("%s %s: bad json (%d): %s", method, path, code, raw)
		}
	}
	return code
}

func quickSuite() config.Suite {
	s := config.DefaultSuite(config.ProfileQuick)
	s.Specs.Scan = scan.Spec{First: scan.RoundSpec{Concurrency: 4, PerProvider: 2, Timeout: 2 * time.Second}}
	s.Specs.Latency.Samples, s.Specs.Latency.Warmup, s.Specs.Latency.InputTokens, s.Specs.Latency.OutputTokens, s.Specs.Latency.Timeout = 1, 0, 20, 8, 5*time.Second
	s.Specs.Capability.Timeout = 5 * time.Second
	s.Specs.Capability.LongContextTiers = []int{2000}
	s.Specs.Quality.Trials, s.Specs.Quality.Timeout = 1, 5*time.Second
	s.LatencyTop, s.CapabilityTop, s.QualityTop = 1, 1, 1
	return s
}

func TestEndpointsAndRunsAPI(t *testing.T) {
	e := newEnv(t, "")
	var ep store.Endpoint
	if code := e.json("POST", "/api/endpoints", map[string]any{"name": "mock", "base_url": e.base, "api_key": "sk-abcdefghijk"}, &ep); code != 200 || ep.ID == "" || !ep.HasKey {
		t.Fatalf("create endpoint = %d %+v", code, ep)
	}
	var list []store.Endpoint
	if code := e.json("GET", "/api/endpoints", nil, &list); code != 200 || len(list) != 1 {
		t.Fatalf("list = %d %+v", code, list)
	}
	var check app.CheckResult
	if code := e.json("POST", "/api/endpoints/"+ep.ID+"/check", nil, &check); code != 200 || !check.Reachable || check.Models != 3 {
		t.Fatalf("check = %d %+v", code, check)
	}
	var disc app.DiscoverResult
	if code := e.json("POST", "/api/endpoints/"+ep.ID+"/discover", map[string]any{"filter": map[string]any{"include_owners": []string{"a"}}}, &disc); code != 200 || disc.Filtered != 2 {
		t.Fatalf("discover = %d %+v", code, disc)
	}
	// overrides
	if code, _ := e.do("PUT", "/api/endpoints/"+ep.ID+"/overrides/provider", map[string]any{"raw_key": "a", "display_name": "Provider A"}); code != 200 {
		t.Fatal("provider override")
	}
	var ov map[string]any
	e.json("GET", "/api/endpoints/"+ep.ID+"/overrides", nil, &ov)
	if len(ov["providers"].([]any)) != 1 {
		t.Fatalf("overrides = %v", ov)
	}
	// start suite
	var started map[string]string
	code := e.json("POST", "/api/runs", map[string]any{"kind": "suite", "endpoint_id": ep.ID, "suite": quickSuite()}, &started)
	if code != 202 || started["id"] == "" {
		t.Fatalf("start = %d %v", code, started)
	}
	id := started["id"]
	// SSE: read until done
	req, _ := http.NewRequest("GET", e.api.URL+"/api/runs/"+id+"/events", nil)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req = req.WithContext(ctx)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatal(err)
	}
	defer resp.Body.Close()
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "text/event-stream") {
		t.Fatalf("content-type = %s", ct)
	}
	rd := bufio.NewReader(resp.Body)
	kinds := map[string]int{}
	gotSnapshot := false
	for {
		line, err := rd.ReadString('\n')
		if err != nil {
			break
		}
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "event: ") {
			k := strings.TrimPrefix(line, "event: ")
			kinds[k]++
			if k == "snapshot" {
				gotSnapshot = true
			}
		}
	}
	if !gotSnapshot || kinds["done"] == 0 || kinds["round"] == 0 {
		t.Fatalf("sse kinds = %v", kinds)
	}
	// wait for the run to finish (the stream may end slightly before the status update)
	deadline := time.Now().Add(10 * time.Second)
	var runDoc struct {
		Run      store.Run   `json:"run"`
		Children []store.Run `json:"children"`
		Running  bool        `json:"running"`
	}
	for time.Now().Before(deadline) {
		e.json("GET", "/api/runs/"+id, nil, &runDoc)
		if runDoc.Run.Status == core.StatusDone && !runDoc.Running {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	if runDoc.Run.Status != core.StatusDone || len(runDoc.Children) != 4 {
		t.Fatalf("run = %+v children=%d", runDoc.Run, len(runDoc.Children))
	}
	// summary / families / providers / models / failures
	var sum map[string]any
	if code := e.json("GET", "/api/runs/"+id+"/summary", nil, &sum); code != 200 || sum["summary"] == nil || sum["provider_field"] != "owned_by" {
		t.Fatalf("summary = %d %v", code, sum)
	}
	var fams map[string]any
	if code := e.json("GET", "/api/runs/"+id+"/families?order=stability", nil, &fams); code != 200 || len(fams["families"].([]any)) == 0 {
		t.Fatalf("families = %d", code)
	}
	var provs map[string]any
	e.json("GET", "/api/runs/"+id+"/providers", nil, &provs)
	if provs["field"] != "owned_by" {
		t.Fatalf("providers = %v", provs)
	}
	var models struct {
		Models []map[string]any `json:"models"`
		Total  int              `json:"total"`
	}
	if code := e.json("GET", "/api/runs/"+id+"/models?status=ok&sort=median&order=desc&size=1", nil, &models); code != 200 || models.Total != 2 || len(models.Models) != 1 || models.Models[0]["id"] != "a/slow" {
		t.Fatalf("models = %d %+v", code, models)
	}
	e.json("GET", "/api/runs/"+id+"/models?q=missing", nil, &models)
	if models.Total != 1 || models.Models[0]["provider_name"] != "b" {
		t.Fatalf("search = %+v", models)
	}
	// provider override applied dynamically
	e.json("GET", "/api/runs/"+id+"/models?q=fast", nil, &models)
	if models.Models[0]["provider_name"] != "Provider A" {
		t.Fatalf("override not applied: %v", models.Models[0]["provider_name"])
	}
	var one map[string]any
	if code := e.json("GET", "/api/runs/"+id+"/models/a/fast", nil, &one); code != 200 || one["history"] == nil {
		t.Fatalf("model = %d %v", code, one)
	}
	var fails map[string]any
	e.json("GET", "/api/runs/"+id+"/failures", nil, &fails)
	if len(fails["groups"].([]any)) != 1 {
		t.Fatalf("failures = %v", fails)
	}
	for _, p := range []string{"latency", "capabilities", "quality"} {
		if code, raw := e.do("GET", "/api/runs/"+id+"/"+p, nil); code != 200 {
			t.Fatalf("%s = %d %s", p, code, raw)
		}
	}
	if code, raw := e.do("GET", "/api/runs/"+id+"/load", nil); code != 404 {
		t.Fatalf("load should be absent: %d %s", code, raw)
	}
	code, raw := e.do("GET", "/api/runs/"+id+"/report.md?lang=en&download=1", nil)
	if code != 200 || !strings.Contains(string(raw), "## 1. Summary") || !strings.Contains(string(raw), "## Latency lab") {
		t.Fatalf("report = %d", code)
	}
	if code, raw := e.do("GET", "/api/runs/"+id+"/export.json", nil); code != 200 || !strings.Contains(string(raw), "schema_version") {
		t.Fatal("json export")
	}
	if code, raw := e.do("GET", "/api/runs/"+id+"/export.csv", nil); code != 200 || !strings.HasPrefix(string(raw), "model_id,") {
		t.Fatal("csv export")
	}
	var runs map[string]any
	e.json("GET", "/api/runs?limit=10", nil, &runs)
	if int(runs["total"].(float64)) != 1 {
		t.Fatalf("runs = %v", runs)
	}
	// compare the suite with its own scan child
	var cmp map[string]any
	if code := e.json("GET", "/api/compare?a="+id+"&b="+id, nil, &cmp); code != 200 || cmp["counts"].(map[string]any)["same"] != float64(3) {
		t.Fatalf("compare = %d %v", code, cmp["counts"])
	}
	// misc
	for _, p := range []string{"/healthz", "/api/version", "/api/reasons", "/api/dialects", "/api/presets", "/api/probes", "/api/questions"} {
		if code, _ := e.do("GET", p, nil); code != 200 {
			t.Fatalf("%s = %d", p, code)
		}
	}
	if code, raw := e.do("GET", "/", nil); code != 200 || !strings.Contains(string(raw), "LLM Speedtest") {
		t.Fatalf("static placeholder = %d", code)
	}
	if code, _ := e.do("GET", "/api/nope", nil); code != 404 {
		t.Fatal("unknown api")
	}
	// delete
	if code, _ := e.do("DELETE", "/api/runs/"+id, nil); code != 200 {
		t.Fatal("delete run")
	}
	if code, _ := e.do("GET", "/api/runs/"+id, nil); code != 404 {
		t.Fatal("deleted run should be 404")
	}
	if code, _ := e.do("DELETE", "/api/endpoints/"+ep.ID, nil); code != 200 {
		t.Fatal("delete endpoint")
	}
}

func TestAuthAndValidation(t *testing.T) {
	e := newEnv(t, "secret")
	saved := e.tok
	e.tok = ""
	if code, _ := e.do("GET", "/api/version", nil); code != 401 {
		t.Fatalf("unauthenticated = %d", code)
	}
	if code, _ := e.do("GET", "/healthz", nil); code != 200 {
		t.Fatal("healthz must be public")
	}
	if code, _ := e.do("GET", "/api/version?token=secret", nil); code != 200 {
		t.Fatal("query token")
	}
	e.tok = saved
	if code, _ := e.do("POST", "/api/endpoints", map[string]any{"name": "x"}); code != 400 {
		t.Fatal("base_url required")
	}
	if code, _ := e.do("POST", "/api/runs", map[string]any{"kind": "load", "endpoint": map[string]any{"base_url": e.base}, "models": []string{"a/fast"}}); code != 409 {
		t.Fatal("load needs confirm")
	}
	if code, _ := e.do("POST", "/api/runs", map[string]any{"kind": "scan", "endpoint_id": "nope"}); code != 404 {
		t.Fatal("unknown endpoint")
	}
	req, _ := http.NewRequest("POST", e.api.URL+"/api/endpoints", strings.NewReader("{bad"))
	req.Header.Set("Authorization", "Bearer secret")
	resp, _ := http.DefaultClient.Do(req)
	if resp.StatusCode != 400 {
		t.Fatal("bad json")
	}
	resp.Body.Close()
	if code, _ := e.do("GET", "/api/compare?a=x", nil); code != 400 {
		t.Fatal("compare validation")
	}
	fmt.Fprint(io.Discard, "")
}
