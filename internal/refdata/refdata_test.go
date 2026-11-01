package refdata

import "testing"

func TestEmbeddedReferenceData(t *testing.T) {
	list, err := List()
	if err != nil {
		t.Fatal(err)
	}
	if len(list) < 2000 {
		t.Fatalf("expected the full reference model list, got %d", len(list))
	}
	models, attempts, err := Load()
	if err != nil {
		t.Fatal(err)
	}
	if len(models) != len(list) {
		t.Fatalf("models %d != list %d", len(models), len(list))
	}
	if len(attempts) < len(models) {
		t.Fatalf("expected at least one attempt per model, got %d attempts for %d models", len(attempts), len(models))
	}
	ids := map[string]bool{}
	for _, m := range models {
		if m.ID == "" || m.OwnedBy == "" {
			t.Fatalf("model with empty id/owner: %+v", m)
		}
		ids[m.ID] = true
	}
	for _, a := range attempts {
		if !ids[a.ModelID] {
			t.Fatalf("attempt for unknown model %q", a.ModelID)
		}
		if a.Round <= 0 {
			t.Fatalf("attempt without round: %+v", a)
		}
	}
}
