package app

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/config"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/scan"
	"github.com/AnnaofArendelle/llm-speedtest/internal/secrets"
	"github.com/AnnaofArendelle/llm-speedtest/internal/store"
	"github.com/AnnaofArendelle/llm-speedtest/internal/testutil/mockllm"
)

func newService(t *testing.T) *Service {
	t.Helper()
	dir := t.TempDir()
	codec, err := secrets.Load(filepath.Join(dir, ".secret"))
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(filepath.Join(dir, "db.sqlite"), codec)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { st.Close() })
	return New(st, "test")
}

func mockModels() []mockllm.ModelSpec {
	return []mockllm.ModelSpec{
		{ID: "a/fast", OwnedBy: "a", ContextLength: 128000, Behavior: mockllm.Behavior{Latency: 20 * time.Millisecond, Content: "one two three four five six", ChunkDelay: 5 * time.Millisecond, ToolCall: true, JSONMode: true, Vision: true, Embedding: true, MaxContext: 20000}},
		{ID: "a/flaky", OwnedBy: "a", Behavior: mockllm.Behavior{Status: 503, Inner: 503, Error: "A server error occurred. Please try again.", FailFirst: 1, Content: "ok"}},
		{ID: "b/missing", OwnedBy: "b", Behavior: mockllm.Behavior{Status: 404, Error: "The model 'missing' does not exist"}},
		{ID: "c/limited", OwnedBy: "c", Behavior: mockllm.Behavior{Status: 429, Inner: 429, Error: "Rate limit exceeded", ResetAfter: 800 * time.Millisecond, FailFirst: 1, Content: "ok"}},
		{ID: "combo/agg", OwnedBy: "combo", Behavior: mockllm.Behavior{}},
	}
}

func fastSuite() config.Suite {
	s := config.DefaultSuite(config.ProfileStandard)
	s.Specs.Scan = scan.Spec{First: scan.RoundSpec{Concurrency: 4, PerProvider: 2, Timeout: 2 * time.Second}, Retry: scan.RoundSpec{Concurrency: 2, PerProvider: 1, Timeout: 2 * time.Second},
		Final: scan.RoundSpec{Concurrency: 2, PerProvider: 1, Timeout: 2 * time.Second}, Resample: 1, MaxCooldownWait: 3 * time.Second}
	s.Specs.Latency.Samples, s.Specs.Latency.Warmup, s.Specs.Latency.InputTokens, s.Specs.Latency.OutputTokens = 2, 1, 30, 16
	s.Specs.Latency.Timeout = 5 * time.Second
	s.Specs.Capability.Timeout = 5 * time.Second
	s.Specs.Capability.LongContextTiers = []int{4000}
	s.Specs.Quality.Trials = 1
	s.Specs.Quality.Timeout = 5 * time.Second
	s.Specs.Load.Levels = []int{1, 2}
	s.Specs.Load.Duration = 300 * time.Millisecond
	s.Specs.Load.InputTokens, s.Specs.Load.OutputTokens = 30, 8
	s.LatencyTop, s.CapabilityTop, s.QualityTop = 2, 2, 2
	s.Load = true
	s.Normalize()
	return s
}

func TestSuiteEndToEnd(t *testing.T) {
	srv := mockllm.New(mockModels())
	defer srv.Close()
	srv.Dialect = "9router"
	svc := newService(t)
	ep := &config.Endpoint{Name: "mock", BaseURL: srv.URL + "/v1"}

	check := svc.Check(context.Background(), *ep)
	if !check.Reachable || !check.AuthOK || check.Models != 5 || check.Dialect.Name != "9router" {
		t.Fatalf("check = %+v", check)
	}
	disc, err := svc.Discover(context.Background(), *ep, "", config.DefaultFilter())
	if err != nil || disc.Total != 5 || disc.Filtered != 4 || disc.Field != "owned_by" {
		t.Fatalf("discover = %+v %v", disc, err)
	}

	// Subscribe before starting so that we see every forwarded event.
	var events []core.Event
	var emu sync.Mutex
	req := Request{Kind: core.KindSuite, Endpoint: ep, Suite: fastSuite(), Confirm: true}
	id, err := svc.RunSyncWithID(context.Background(), req, func(runID string) {
		ch, _ := svc.Bus.Subscribe(runID, 2048)
		go func() {
			for ev := range ch {
				emu.Lock()
				events = append(events, ev)
				emu.Unlock()
			}
		}()
	})
	if err != nil {
		t.Fatal(err)
	}
	run, err := svc.Store.GetRun(id)
	if err != nil || run.Status != core.StatusDone {
		t.Fatalf("run = %+v %v", run, err)
	}
	kids, _ := svc.Store.ChildRuns(id)
	kinds := map[string]*store.Run{}
	for _, k := range kids {
		kinds[k.Kind] = k
	}
	for _, k := range []string{core.KindScan, core.KindLatency, core.KindCapability, core.KindQuality, core.KindLoad} {
		if kinds[k] == nil || kinds[k].Status != core.StatusDone {
			t.Fatalf("child %s = %+v", k, kinds[k])
		}
	}
	res, _, err := svc.ScanResult(id)
	if err != nil {
		t.Fatal(err)
	}
	// fast ok; flaky recovers in retry; limited recovers after cooldown; missing fails; combo excluded
	if res.Summary.Models != 4 || res.Summary.OK != 3 || res.Summary.Fail != 1 || res.Summary.Reasons["model_unavailable"] != 1 {
		t.Fatalf("scan summary = %+v", res.Summary)
	}
	if m := res.Model("a/flaky"); m == nil || m.Status != "ok" || m.Flaky["upstream_5xx"] != 1 {
		t.Fatalf("flaky = %+v", m)
	}
	if m := res.Model("c/limited"); m == nil || m.Status != "ok" {
		t.Fatalf("limited = %+v", m)
	}
	if res.ProviderField != "owned_by" || res.Model("a/fast").ProviderName != "a" {
		t.Fatalf("provider = %s %s", res.ProviderField, res.Model("a/fast").ProviderName)
	}
	// latency child measured the top-2 models
	samples, _ := svc.Store.ListLatencySamples(kinds[core.KindLatency].ID)
	if len(samples) < 3 {
		t.Fatalf("latency samples = %d", len(samples))
	}
	caps, _ := svc.Store.ListCapabilityResults(kinds[core.KindCapability].ID)
	if len(caps) == 0 {
		t.Fatal("no capability results")
	}
	qs, _ := svc.Store.ListQualityResults(kinds[core.KindQuality].ID)
	if len(qs) == 0 {
		t.Fatal("no quality results")
	}
	lr, _ := svc.Store.ListLoadRequests(kinds[core.KindLoad].ID)
	if len(lr) == 0 {
		t.Fatal("no load requests")
	}
	var loadSum struct {
		Model string `json:"model"`
	}
	_ = json.Unmarshal(kinds[core.KindLoad].Summary, &loadSum)
	if loadSum.Model == "" {
		t.Fatalf("load summary = %s", kinds[core.KindLoad].Summary)
	}
	// reports
	md, err := svc.Report(id, "zh")
	if err != nil {
		t.Fatal(err)
	}
	for _, must := range []string{"## 1. 结论摘要", "### 4.1 综合排名", "## 延迟实验室", "## 能力矩阵", "## 质量抽检", "## 负载测试", "局限性"} {
		if !strings.Contains(md, must) {
			t.Errorf("report missing %q", must)
		}
	}
	en, _ := svc.Report(kinds[core.KindScan].ID, "en")
	if !strings.Contains(en, "## 1. Summary") {
		t.Error("english scan report")
	}
	single, _ := svc.Report(kinds[core.KindLatency].ID, "en")
	if !strings.Contains(single, "## Latency lab") {
		t.Error("latency-only report")
	}
	if js, err := svc.ExportJSON(id); err != nil || !strings.Contains(string(js), `"schema_version"`) {
		t.Fatalf("json export %v", err)
	}
	if csvb, err := svc.ExportCSV(id); err != nil || !strings.HasPrefix(string(csvb), "model_id,") {
		t.Fatalf("csv export %v", err)
	}
	// events: the suite run received forwarded stage events and a final done
	time.Sleep(50 * time.Millisecond)
	emu.Lock()
	defer emu.Unlock()
	stages := map[string]bool{}
	var done bool
	for _, ev := range events {
		stages[ev.Stage] = true
		if ev.Kind == core.EventDone && ev.Stage == "suite" {
			done = true
		}
	}
	for _, st := range []string{"scan", "latency", "capability", "quality", "load"} {
		if !stages[st] {
			t.Errorf("no events for stage %s", st)
		}
	}
	if !done {
		t.Error("no suite done event")
	}
	if ev, ok := svc.Bus.Latest(id); !ok || ev.Seq == 0 {
		t.Error("latest event")
	}
	if h := svc.Bus.History(id, 0); len(h) == 0 {
		t.Error("history")
	}
}

func TestStartCancelAndLoadConfirm(t *testing.T) {
	srv := mockllm.New([]mockllm.ModelSpec{{ID: "h/x", OwnedBy: "h", Behavior: mockllm.Behavior{Hang: 20 * time.Second}}})
	defer srv.Close()
	svc := newService(t)
	ep := &config.Endpoint{Name: "mock", BaseURL: srv.URL + "/v1"}
	suite := config.DefaultSuite(config.ProfileQuick)
	suite.Specs.Scan.First.Timeout = 30 * time.Second
	if _, err := svc.Start(Request{Kind: core.KindLoad, Endpoint: ep, Suite: suite, Models: []string{"h/x"}}); !errors.Is(err, ErrConfirmRequired) {
		t.Fatalf("load without confirm: %v", err)
	}
	id, err := svc.Start(Request{Kind: core.KindScan, Endpoint: ep, Suite: suite})
	if err != nil {
		t.Fatal(err)
	}
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if r, _ := svc.Store.GetRun(id); r != nil && r.Status == core.StatusRunning {
			break
		}
		time.Sleep(20 * time.Millisecond)
	}
	if !svc.IsRunning(id) {
		t.Fatal("run should be in progress")
	}
	if !svc.Cancel(id) {
		t.Fatal("cancel")
	}
	deadline = time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if r, _ := svc.Store.GetRun(id); r != nil && r.Status == core.StatusCancelled {
			return
		}
		time.Sleep(20 * time.Millisecond)
	}
	r, _ := svc.Store.GetRun(id)
	t.Fatalf("run not cancelled: %+v", r)
}

func TestBus(t *testing.T) {
	b := NewBus()
	ch, unsub := b.Subscribe("r", 2)
	defer unsub()
	for i := 0; i < 5; i++ {
		b.Publish("r", core.Event{Kind: core.EventProgress, Done: i})
	}
	b.Publish("r", core.Event{Kind: core.EventDone})
	// buffer of 2: the newest events survive
	var got []core.Event
	for len(ch) > 0 {
		got = append(got, <-ch)
	}
	if len(got) != 2 || got[len(got)-1].Kind != core.EventDone {
		t.Fatalf("got = %+v", got)
	}
	if ev, ok := b.Latest("r"); !ok || ev.Done != 4 {
		t.Fatalf("latest = %+v", ev)
	}
	if h := b.History("r", 0); len(h) != 1 || h[0].Kind != core.EventDone || h[0].Seq != 6 {
		t.Fatalf("history = %+v", h)
	}
	if h := b.History("r", 6); len(h) != 0 {
		t.Fatal("history after")
	}
	b.Forget("r")
	if _, ok := b.Latest("r"); ok {
		t.Fatal("forget")
	}
	unsub()
	unsub() // idempotent
}
