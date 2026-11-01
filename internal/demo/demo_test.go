package demo

import (
	"path/filepath"
	"testing"

	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/secrets"
	"github.com/AnnaofArendelle/llm-speedtest/internal/store"
)

func TestSeed(t *testing.T) {
	dir := t.TempDir()
	codec, err := secrets.Load(filepath.Join(dir, ".secret"))
	if err != nil {
		t.Fatal(err)
	}
	st, err := store.Open(filepath.Join(dir, "db.sqlite"), codec)
	if err != nil {
		t.Fatal(err)
	}
	defer st.Close()

	id, err := Seed(st)
	if err != nil {
		t.Fatal(err)
	}
	run, err := st.GetRun(id)
	if err != nil {
		t.Fatal(err)
	}
	if run.Kind != core.KindSuite || run.Status != core.StatusDone {
		t.Fatalf("suite run = %s/%s", run.Kind, run.Status)
	}
	if run.StartedAt.IsZero() || run.FinishedAt.IsZero() || !run.FinishedAt.After(run.StartedAt) {
		t.Fatalf("suite run should have a realistic time window: %v – %v", run.StartedAt, run.FinishedAt)
	}
	kids, err := st.ChildRuns(id)
	if err != nil {
		t.Fatal(err)
	}
	kinds := map[string]bool{}
	for _, k := range kids {
		kinds[k.Kind] = true
		if k.Status != core.StatusDone {
			t.Fatalf("child %s is %s", k.Kind, k.Status)
		}
	}
	for _, want := range []string{core.KindScan, core.KindLatency, core.KindCapability, core.KindQuality, core.KindLoad} {
		if !kinds[want] {
			t.Fatalf("missing %s child run", want)
		}
	}
	eps, err := st.ListEndpoints()
	if err != nil || len(eps) != 1 {
		t.Fatalf("endpoints: %v %d", err, len(eps))
	}
	var scanID string
	for _, k := range kids {
		if k.Kind == core.KindScan {
			scanID = k.ID
		}
	}
	models, err := st.ListRunModels(scanID)
	if err != nil || len(models) < 2000 {
		t.Fatalf("scan models: %v %d", err, len(models))
	}
	attempts, err := st.ListAttempts(scanID)
	if err != nil || len(attempts) < len(models) {
		t.Fatalf("scan attempts: %v %d", err, len(attempts))
	}
	// Seeding twice must not fail (idempotent demo start-up).
	if _, err := Seed(st); err != nil {
		t.Fatalf("second seed: %v", err)
	}
}
