package report

import (
	"bytes"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/AnnaofArendelle/llm-speedtest/internal/analyze"
)

// SchemaVersion is bumped whenever the JSON export shape changes.
const SchemaVersion = 1

// JSON renders the analysis as a stable, versioned JSON document.
func JSON(res *analyze.Result, meta Meta) ([]byte, error) {
	doc := map[string]any{
		"schema_version": SchemaVersion,
		"meta": map[string]any{
			"title": meta.Title, "endpoint": meta.Endpoint, "dialect": meta.Dialect, "tool_version": meta.Version,
			"generated": meta.Generated, "started": meta.Started, "finished": meta.Finished, "preset": meta.Preset,
		},
		"summary":               res.Summary,
		"scoring":               res.Scoring,
		"provider_field":        res.ProviderField,
		"provider_rationale":    res.ProviderRationale,
		"provider_suggestions":  res.Suggestions,
		"providers":             res.Providers,
		"families":              res.Families,
		"families_by_stability": res.FamiliesByStability,
		"unavailable_families":  res.UnavailableFamilies,
		"models":                res.Models,
	}
	return json.MarshalIndent(doc, "", " ")
}

// CSV renders one row per model.
func CSV(res *analyze.Result) ([]byte, error) {
	var buf bytes.Buffer
	w := csv.NewWriter(&buf)
	header := []string{"model_id", "provider", "owned_by", "family", "family_key", "family_source", "status", "tier", "attempts", "ok", "success_rate",
		"median_s", "min_s", "max_s", "jitter", "served_model", "reason", "http", "inner", "message", "context_length"}
	if err := w.Write(header); err != nil {
		return nil, err
	}
	for _, m := range res.Models {
		row := []string{m.ID, m.ProviderName, m.OwnedBy, m.FamilyCanon, m.FamilyKey, m.FamilySource, m.Status, m.Tier,
			fmt.Sprint(m.Attempts), fmt.Sprint(m.OK), fmt.Sprintf("%.3f", m.Rate), f3(m.Med), f3(m.Min), f3(m.Max), f3(m.Jitter), m.Served, m.Reason,
			fmt.Sprint(m.HTTP), fmt.Sprint(m.Inner), strings.ReplaceAll(m.Message, "\n", " "), fmt.Sprint(m.ContextLength)}
		if err := w.Write(row); err != nil {
			return nil, err
		}
	}
	w.Flush()
	return buf.Bytes(), w.Error()
}

func f3(x float64) string {
	if x == 0 {
		return ""
	}
	return fmt.Sprintf("%.3f", x)
}
