package store

import (
	"errors"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/secrets"
)

func open(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	codec, err := secrets.Load(filepath.Join(dir, ".secret"))
	if err != nil {
		t.Fatal(err)
	}
	s, err := Open(filepath.Join(dir, "data", "test.db"), codec)
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { s.Close() })
	return s
}

func TestMigrateIdempotent(t *testing.T) {
	s := open(t)
	if err := s.migrate(); err != nil {
		t.Fatal(err)
	}
	s2, err := Open(s.Path(), nil)
	if err != nil {
		t.Fatal(err)
	}
	s2.Close()
}

func TestEndpointsAndOverrides(t *testing.T) {
	s := open(t)
	ep := &Endpoint{Name: "local", BaseURL: "http://127.0.0.1:20128/v1", APIKey: "sk-abcdefgh12345", Headers: map[string]string{"X-A": "1"}}
	if err := s.SaveEndpoint(ep); err != nil {
		t.Fatal(err)
	}
	if ep.ID == "" || !ep.HasKey || ep.KeyHint != "sk-a••••2345" {
		t.Fatalf("saved = %+v", ep)
	}
	got, err := s.GetEndpoint(ep.ID)
	if err != nil || got.APIKey != "" || !got.HasKey || got.Headers["X-A"] != "1" {
		t.Fatalf("get = %+v %v", got, err)
	}
	wk, err := s.GetEndpointWithKey(ep.ID)
	if err != nil || wk.APIKey != "sk-abcdefgh12345" {
		t.Fatalf("with key = %+v %v", wk, err)
	}
	// update without key keeps the key
	ep.APIKey = ""
	ep.Name = "renamed"
	if err := s.SaveEndpoint(ep); err != nil {
		t.Fatal(err)
	}
	wk, _ = s.GetEndpointWithKey(ep.ID)
	if wk.Name != "renamed" || wk.APIKey != "sk-abcdefgh12345" {
		t.Fatalf("update kept key = %+v", wk)
	}
	if err := s.ClearEndpointKey(ep.ID); err != nil {
		t.Fatal(err)
	}
	wk, _ = s.GetEndpointWithKey(ep.ID)
	if wk.APIKey != "" || wk.HasKey {
		t.Fatal("clear key")
	}
	list, err := s.ListEndpoints()
	if err != nil || len(list) != 1 {
		t.Fatal("list")
	}
	if _, err := s.GetEndpoint("nope"); !errors.Is(err, ErrNotFound) {
		t.Fatal("not found")
	}
	// overrides
	if err := s.SetProviderOverride(ep.ID, ProviderOverrideRow{RawKey: "kilo", DisplayName: "Kilo Gateway"}); err != nil {
		t.Fatal(err)
	}
	if err := s.SetProviderOverride(ep.ID, ProviderOverrideRow{RawKey: "k2", MergedInto: "kilo"}); err != nil {
		t.Fatal(err)
	}
	m, rows, err := s.ProviderOverrides(ep.ID)
	if err != nil || len(rows) != 2 || m["kilo"].DisplayName != "Kilo Gateway" || m["k2"].MergedInto != "kilo" {
		t.Fatalf("provider overrides = %+v %v", rows, err)
	}
	if err := s.SetProviderOverride(ep.ID, ProviderOverrideRow{RawKey: "k2"}); err != nil {
		t.Fatal(err)
	}
	if _, rows, _ := s.ProviderOverrides(ep.ID); len(rows) != 1 {
		t.Fatal("delete override")
	}
	if err := s.SetFamilyOverride(ep.ID, FamilyOverrideRow{ModelID: "x/y", FamilyKey: "a|b|1", Canon: "a-b-1"}); err != nil {
		t.Fatal(err)
	}
	fm, frows, err := s.FamilyOverrides(ep.ID)
	if err != nil || len(frows) != 1 || fm["x/y"].Key != "a|b|1" {
		t.Fatal("family overrides")
	}
	if err := s.DeleteEndpoint(ep.ID); err != nil {
		t.Fatal(err)
	}
	if _, rows, _ := s.ProviderOverrides(ep.ID); len(rows) != 0 {
		t.Fatal("cascade")
	}
}

func TestRunsAndMeasurements(t *testing.T) {
	s := open(t)
	run := &Run{Kind: core.KindScan, EndpointName: "e", BaseURL: "http://x/v1", Preset: "quick", Spec: []byte(`{"a":1}`)}
	if err := s.CreateRun(run); err != nil {
		t.Fatal(err)
	}
	if err := s.SetRunStatus(run.ID, core.StatusRunning, ""); err != nil {
		t.Fatal(err)
	}
	models := []core.RunModel{{ID: "a/x", OwnedBy: "a", Upstream: "x", ProviderKey: "a", ProviderName: "a", FamilyKey: "x||", FamilyCanon: "x", FamilySource: "rule"}, {ID: "b/y", OwnedBy: "b"}}
	if err := s.PutRunModels(run.ID, models); err != nil {
		t.Fatal(err)
	}
	w := s.AttemptWriter(run.ID)
	var wg sync.WaitGroup
	for i := 0; i < 120; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			_ = w.Add(core.Attempt{ModelID: "a/x", Round: 1 + i%3, At: time.Now(), OK: i%2 == 0, Reason: "ok", Latency: time.Duration(i) * time.Millisecond, Message: "m"})
		}(i)
	}
	wg.Wait()
	if err := w.Flush(); err != nil {
		t.Fatal(err)
	}
	as, err := s.ListAttempts(run.ID)
	if err != nil || len(as) != 120 {
		t.Fatalf("attempts = %d %v", len(as), err)
	}
	if as[0].Round != 1 || as[len(as)-1].Round != 3 {
		t.Fatal("ordering by round")
	}
	// timer-based flush
	w2 := s.AttemptWriter(run.ID)
	_ = w2.Add(core.Attempt{ModelID: "b/y", Round: 1, At: time.Now()})
	time.Sleep(400 * time.Millisecond)
	as, _ = s.ListAttempts(run.ID)
	if len(as) != 121 {
		t.Fatalf("timer flush: %d", len(as))
	}
	ms, err := s.ListRunModels(run.ID)
	if err != nil || len(ms) != 2 || ms[0].ProviderName != "a" {
		t.Fatalf("models = %+v %v", ms, err)
	}
	ms[1].ProviderName = "B!"
	if err := s.UpdateRunModelAssignments(run.ID, ms); err != nil {
		t.Fatal(err)
	}
	ms, _ = s.ListRunModels(run.ID)
	if ms[1].ProviderName != "B!" {
		t.Fatal("assignment update")
	}
	// latency / load / capability / quality
	if err := s.InsertLatencySamples(run.ID, []core.LatencySample{{ModelID: "a/x", Index: 0, OK: true, TTFT: 120 * time.Millisecond, E2E: time.Second, TokPerSec: 42, OutputTokens: 40, TokensExact: true, At: time.Now()}}); err != nil {
		t.Fatal(err)
	}
	ls, err := s.ListLatencySamples(run.ID)
	if err != nil || len(ls) != 1 || ls[0].TTFT != 120*time.Millisecond || !ls[0].TokensExact {
		t.Fatalf("latency = %+v %v", ls, err)
	}
	if err := s.InsertLoadRequests(run.ID, []core.LoadRequest{{Level: 2, Index: 0, OK: true, E2E: 500 * time.Millisecond, At: time.Now()}}); err != nil {
		t.Fatal(err)
	}
	lr, _ := s.ListLoadRequests(run.ID)
	if len(lr) != 1 || lr[0].Level != 2 {
		t.Fatal("load")
	}
	if err := s.PutCapabilityResults(run.ID, []core.CapabilityResult{{ModelID: "a/x", Probe: "tools", Status: "supported"}}); err != nil {
		t.Fatal(err)
	}
	if err := s.PutCapabilityResults(run.ID, []core.CapabilityResult{{ModelID: "a/x", Probe: "tools", Status: "error", Evidence: "x"}}); err != nil {
		t.Fatal(err)
	}
	cr, _ := s.ListCapabilityResults(run.ID)
	if len(cr) != 1 || cr[0].Status != "error" {
		t.Fatalf("cap upsert = %+v", cr)
	}
	if err := s.InsertQualityResults(run.ID, []core.QualityResult{{ModelID: "a/x", QuestionID: "q1", Trial: 1, Pass: true, Score: 1}}); err != nil {
		t.Fatal(err)
	}
	qr, _ := s.ListQualityResults(run.ID)
	if len(qr) != 1 || !qr[0].Pass {
		t.Fatal("quality")
	}
	// summary / status / list
	if err := s.SetRunSummary(run.ID, map[string]int{"ok": 1}); err != nil {
		t.Fatal(err)
	}
	if err := s.SetRunProgress(run.ID, map[string]int{"done": 3}); err != nil {
		t.Fatal(err)
	}
	if err := s.SetRunStatus(run.ID, core.StatusDone, ""); err != nil {
		t.Fatal(err)
	}
	got, err := s.GetRun(run.ID)
	if err != nil || got.Status != core.StatusDone || string(got.Summary) != `{"ok":1}` || got.StartedAt.IsZero() || got.FinishedAt.IsZero() || string(got.Spec) != `{"a":1}` {
		t.Fatalf("run = %+v %v", got, err)
	}
	child := &Run{Kind: core.KindLatency, ParentID: run.ID}
	if err := s.CreateRun(child); err != nil {
		t.Fatal(err)
	}
	list, total, err := s.ListRuns("", "", 10, 0)
	if err != nil || total != 1 || len(list) != 1 {
		t.Fatalf("list top-level = %d %v", total, err)
	}
	if kids, _ := s.ChildRuns(run.ID); len(kids) != 1 {
		t.Fatal("children")
	}
	if list, _, _ := s.ListRuns(core.KindLatency, "", 10, 0); len(list) != 1 {
		t.Fatal("kind filter")
	}
	// interrupted marking
	r2 := &Run{Kind: core.KindScan}
	_ = s.CreateRun(r2)
	_ = s.SetRunStatus(r2.ID, core.StatusRunning, "")
	n, err := s.MarkInterrupted() // the pending child run and r2
	if err != nil || n != 2 {
		t.Fatalf("interrupted = %d %v", n, err)
	}
	// delete cascade
	if err := s.DeleteRun(run.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.GetRun(child.ID); !errors.Is(err, ErrNotFound) {
		t.Fatal("child deleted")
	}
	if as, _ := s.ListAttempts(run.ID); len(as) != 0 {
		t.Fatal("attempts deleted")
	}
}

func TestBatchErrorSticks(t *testing.T) {
	calls := 0
	b := NewBatch(2, time.Hour, func(rows []int) error {
		calls++
		if calls == 1 {
			return errBoom
		}
		return nil
	})
	_ = b.Add(1)
	if err := b.Add(2); !errors.Is(err, errBoom) {
		t.Fatalf("first flush error = %v", err)
	}
	if err := b.Add(3); !errors.Is(err, errBoom) {
		t.Fatal("error must stick")
	}
}

var errBoom = &boom{}

type boom struct{}

func (*boom) Error() string { return "boom" }
