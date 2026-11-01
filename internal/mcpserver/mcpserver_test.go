package mcpserver

import (
	"context"
	"encoding/json"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/AnnaofArendelle/llm-speedtest/internal/app"
	"github.com/AnnaofArendelle/llm-speedtest/internal/secrets"
	"github.com/AnnaofArendelle/llm-speedtest/internal/store"
	"github.com/AnnaofArendelle/llm-speedtest/internal/testutil/mockllm"
)

func call(t *testing.T, cs *mcp.ClientSession, name string, args map[string]any) map[string]any {
	t.Helper()
	res, err := cs.CallTool(context.Background(), &mcp.CallToolParams{Name: name, Arguments: args})
	if err != nil {
		t.Fatalf("%s: %v", name, err)
	}
	if len(res.Content) == 0 {
		t.Fatalf("%s: no content", name)
	}
	tc, ok := res.Content[0].(*mcp.TextContent)
	if !ok {
		t.Fatalf("%s: content is %T", name, res.Content[0])
	}
	if res.IsError {
		t.Fatalf("%s: tool error: %s", name, tc.Text)
	}
	var out map[string]any
	if err := json.Unmarshal([]byte(tc.Text), &out); err != nil {
		// non-JSON tools (markdown export) return the raw text under "text"
		return map[string]any{"text": tc.Text}
	}
	return out
}

func TestMCPEndToEnd(t *testing.T) {
	dir := t.TempDir()
	codec, _ := secrets.Load(filepath.Join(dir, ".secret"))
	st, err := store.Open(filepath.Join(dir, "db.sqlite"), codec)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()
	svc := app.New(st, "test")
	llm := mockllm.New([]mockllm.ModelSpec{
		{ID: "a/fast", OwnedBy: "a", Behavior: mockllm.Behavior{Latency: 5 * time.Millisecond, Content: "ok"}},
		{ID: "b/missing", OwnedBy: "b", Behavior: mockllm.Behavior{Status: 404, Error: "The model 'missing' does not exist"}},
	})
	defer llm.Close()

	srv := NewServer(svc)
	ct, stTransport := mcp.NewInMemoryTransports()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	ss, err := srv.Connect(ctx, stTransport, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer ss.Close()
	client := mcp.NewClient(&mcp.Implementation{Name: "test-client", Version: "0"}, nil)
	cs, err := client.Connect(ctx, ct, nil)
	if err != nil {
		t.Fatal(err)
	}
	defer cs.Close()

	tools, err := cs.ListTools(ctx, nil)
	if err != nil || len(tools.Tools) < 18 {
		t.Fatalf("tools = %d %v", len(tools.Tools), err)
	}
	if out := call(t, cs, "list_endpoints", nil); len(out["endpoints"].([]any)) != 0 {
		t.Fatalf("endpoints = %v", out)
	}
	ep := call(t, cs, "upsert_endpoint", map[string]any{"name": "mock", "base_url": llm.URL + "/v1", "api_key": "sk-test-1234567"})
	epID, _ := ep["id"].(string)
	if epID == "" || ep["has_key"] != true {
		t.Fatalf("upsert = %v", ep)
	}
	chk := call(t, cs, "check_endpoint", map[string]any{"endpoint_id": epID})
	if chk["reachable"] != true || chk["models"] != float64(2) {
		t.Fatalf("check = %v", chk)
	}
	disc := call(t, cs, "discover_models", map[string]any{"base_url": llm.URL + "/v1", "max_models": 1})
	if disc["total"] != float64(2) || len(disc["models"].([]any)) != 1 {
		t.Fatalf("discover = %v", disc)
	}
	started := call(t, cs, "start_run", map[string]any{"kind": "scan", "endpoint_id": epID, "profile": "quick"})
	runID, _ := started["run_id"].(string)
	if runID == "" {
		t.Fatalf("start = %v", started)
	}
	waited := call(t, cs, "wait_run", map[string]any{"run_id": runID, "timeout_s": 60})
	if waited["finished"] != true {
		t.Fatalf("wait = %v", waited)
	}
	run := waited["run"].(map[string]any)
	if run["status"] != "done" {
		t.Fatalf("status = %v", run["status"])
	}
	rank := call(t, cs, "get_ranking", map[string]any{"run_id": runID})
	if rank["summary"].(map[string]any)["ok"] != float64(1) {
		t.Fatalf("ranking = %v", rank["summary"])
	}
	models := call(t, cs, "get_models", map[string]any{"run_id": runID, "status": "fail"})
	if models["total"] != float64(1) {
		t.Fatalf("models = %v", models)
	}
	fails := call(t, cs, "get_failures", map[string]any{"run_id": runID})
	if len(fails["groups"].([]any)) != 1 {
		t.Fatalf("failures = %v", fails)
	}
	provs := call(t, cs, "get_providers", map[string]any{"run_id": runID})
	if provs["field"] != "owned_by" {
		t.Fatalf("providers = %v", provs)
	}
	md := call(t, cs, "export_report", map[string]any{"run_id": runID, "lang": "en"})
	if !strings.Contains(md["text"].(string), "## 1. Summary") {
		t.Fatalf("report = %.200s", md["text"])
	}
	out := call(t, cs, "export_report", map[string]any{"run_id": runID, "format": "json", "out_path": filepath.Join(dir, "r.json")})
	if out["written"] == nil {
		t.Fatalf("export file = %v", out)
	}
	cmp := call(t, cs, "compare_runs", map[string]any{"a": runID, "b": runID})
	if cmp["improved"] != float64(0) {
		t.Fatalf("compare = %v", cmp)
	}
	list := call(t, cs, "list_runs", nil)
	if list["total"] != float64(1) {
		t.Fatalf("list = %v", list)
	}
	// error paths return IsError rather than failing the call
	res, err := cs.CallTool(ctx, &mcp.CallToolParams{Name: "get_latency", Arguments: map[string]any{"run_id": runID}})
	if err != nil || !res.IsError {
		t.Fatalf("get_latency on a scan run should be a tool error: %v %v", res, err)
	}
	res, err = cs.CallTool(ctx, &mcp.CallToolParams{Name: "start_run", Arguments: map[string]any{"kind": "load", "endpoint_id": epID, "models": []string{"a/fast"}}})
	if err != nil || !res.IsError {
		t.Fatal("load without confirm must be a tool error")
	}
	// resources and prompt
	rr, err := cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "speedtest://metrics-glossary"})
	if err != nil || len(rr.Contents) == 0 || !strings.Contains(rr.Contents[0].Text, "TTFT") {
		t.Fatalf("resource = %v %v", rr, err)
	}
	rr, err = cs.ReadResource(ctx, &mcp.ReadResourceParams{URI: "speedtest://runs/" + runID + "/summary"})
	if err != nil || len(rr.Contents) == 0 || !strings.Contains(rr.Contents[0].Text, "summary") {
		t.Fatalf("run summary resource = %v %v", rr, err)
	}
	pr, err := cs.GetPrompt(ctx, &mcp.GetPromptParams{Name: "summarize_run", Arguments: map[string]string{"run_id": runID}})
	if err != nil || len(pr.Messages) != 1 {
		t.Fatalf("prompt = %v %v", pr, err)
	}
}
