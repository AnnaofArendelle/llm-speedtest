// Package refscan exposes the embedded reference scan to tests.
package refscan

import (
	"testing"

	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/refdata"
)

// Load returns the reference models and attempts.
func Load(t testing.TB) ([]core.RunModel, []core.Attempt) {
	t.Helper()
	models, attempts, err := refdata.Load()
	if err != nil {
		t.Fatal(err)
	}
	return models, attempts
}

// List returns the reference model list.
func List(t testing.TB) []refdata.ListItem {
	t.Helper()
	items, err := refdata.List()
	if err != nil {
		t.Fatal(err)
	}
	return items
}
