package server

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/app"
	"github.com/AnnaofArendelle/llm-speedtest/internal/config"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/store"
	"github.com/AnnaofArendelle/llm-speedtest/internal/testutil/mockllm"
)

var updateFixtures = flag.Bool("update-fixtures", false, "write API responses to web/src/test/fixtures")

// collectionFields are JSON keys whose values must always be arrays/objects,
// never null, because the web UI iterates over them.
var collectionFields = map[string]bool{
	"top_models": true, "top_families": true, "providers": true, "families": true, "unavailable": true, "models": true,
	"groups": true, "by_provider": true, "stats": true, "samples": true, "scores": true, "results": true, "member_ids": true,
	"stable_ids": true, "buckets": true, "probes": true, "children": true, "runs": true, "reasons": true, "tier_counts": true,
	"family_tier_counts": true, "cells": true, "support": true, "questions": true, "categories": true, "ok_providers": true,
	"sources": true, "inflight": true, "skipped": true, "requests": true, "load": true, "trial_scores": true, "counts": true,
	"history": true, "members": true, "endpoints": true, "keys": true, "dimensions": true,
}

func assertNoNullCollections(t *testing.T, where string, raw []byte) {
	t.Helper()
	var doc any
	if err := json.Unmarshal(raw, &doc); err != nil {
		t.Fatalf("%s: invalid JSON: %v", where, err)
	}
	// Echoed request inputs (run.spec, filters, presets) may legitimately contain
	// null optional lists; only outputs are checked.
	ignoreSubtree := map[string]bool{"spec": true, "filter": true}
	var walk func(path string, v any)
	walk = func(path string, v any) {
		switch x := v.(type) {
		case map[string]any:
			for k, val := range x {
				if ignoreSubtree[k] {
					continue
				}
				if val == nil && collectionFields[k] {
					t.Errorf("%s: %s.%s is null (must be an empty collection)", where, path, k)
				}
				walk(path+"."+k, val)
			}
		case []any:
			for i, val := range x {
				walk(fmt.Sprintf("%s[%d]", path, i), val)
			}
		}
	}
	walk("$", doc)
}

type scenario struct {
	name  string
	runID string
	paths map[string]string // fixture name -> API path
}

func mustCreateRun(t *testing.T, st *store.Store, ep *store.Endpoint, base, status string, models []core.RunModel, attempts []core.Attempt, errMsg string) *store.Run {
	t.Helper()
	run := &store.Run{Kind: core.KindScan, EndpointID: ep.ID, EndpointName: ep.Name, BaseURL: base, Dialect: "generic", Status: core.StatusPending, Preset: "quick", Spec: json.RawMessage(`{"kind":"scan","suite":{"profile":"quick"}}`), CreatedAt: time.Now()}
	if err := st.CreateRun(run); err != nil {
		t.Fatal(err)
	}
	if len(models) > 0 {
		if err := st.PutRunModels(run.ID, models); err != nil {
			t.Fatal(err)
		}
	}
	if len(attempts) > 0 {
		if err := st.InsertAttempts(run.ID, attempts); err != nil {
			t.Fatal(err)
		}
	}
	if err := st.SetRunStatus(run.ID, status, errMsg); err != nil {
		t.Fatal(err)
	}
	return run
}

func scanPaths(id string) map[string]string {
	return map[string]string{
		"run":                "/api/runs/" + id,
		"summary":            "/api/runs/" + id + "/summary",
		"families-composite": "/api/runs/" + id + "/families?order=composite",
		"families-stability": "/api/runs/" + id + "/families?order=stability",
		"providers":          "/api/runs/" + id + "/providers",
		"models":             "/api/runs/" + id + "/models?size=50",
		"failures":           "/api/runs/" + id + "/failures",
		"export":             "/api/runs/" + id + "/export.json",
	}
}

// TestAPIContract exercises every read endpoint against runs in each
// lifecycle state and asserts that collection fields are never null. With
// -update-fixtures it also writes the responses for the web UI tests.
func TestAPIContract(t *testing.T) {
	e := newEnv(t, "")
	st := e.svc.Store
	var ep store.Endpoint
	if code := e.json("POST", "/api/endpoints", map[string]any{"name": "mock", "base_url": e.base}, &ep); code != 200 {
		t.Fatalf("create endpoint: %d", code)
	}
	models := []core.RunModel{
		{ID: "a/fast", OwnedBy: "a", Upstream: "fast", ProviderKey: "a", ProviderName: "a", FamilyKey: "fast||", FamilyCanon: "fast", FamilySource: "rule"},
		{ID: "a/slow", OwnedBy: "a", Upstream: "slow", ProviderKey: "a", ProviderName: "a", FamilyKey: "slow||", FamilyCanon: "slow", FamilySource: "rule"},
		{ID: "b/missing", OwnedBy: "b", Upstream: "missing", ProviderKey: "b", ProviderName: "b", FamilyKey: "missing||", FamilyCanon: "missing", FamilySource: "rule"},
	}
	now := time.Now()
	okAttempt := core.Attempt{ModelID: "a/fast", Round: 1, At: now, HTTP: 200, Latency: 120 * time.Millisecond, OK: true, Class: "ok", Reason: "ok", ServedModel: "fast"}
	failAttempt := core.Attempt{ModelID: "b/missing", Round: 1, At: now, HTTP: 404, Latency: 30 * time.Millisecond, Class: "model_unavailable", Reason: "model_unavailable", Message: "The model 'missing' does not exist"}
	skippedAttempt := core.Attempt{ModelID: "a/slow", Round: 1, At: now, Class: "skipped", Reason: "skipped", Message: "skipped by user"}

	discovering := mustCreateRun(t, st, &ep, e.base, core.StatusDiscovering, nil, nil, "")
	running := mustCreateRun(t, st, &ep, e.base, core.StatusRunning, models, []core.Attempt{okAttempt}, "")
	failed := mustCreateRun(t, st, &ep, e.base, core.StatusFailed, models, []core.Attempt{failAttempt, skippedAttempt}, "list models: connection refused")

	// Real runs through the engine: a scan where every model fails, a full
	// quick suite, and single-stage latency/load runs.
	dead := mockllm.New([]mockllm.ModelSpec{
		{ID: "x/gone", OwnedBy: "x", Behavior: mockllm.Behavior{Status: 404, Error: "The model 'gone' does not exist"}},
		{ID: "x/locked", OwnedBy: "x", Behavior: mockllm.Behavior{Status: 401, Error: "Invalid API key"}},
	})
	t.Cleanup(dead.Close)
	ctx := context.Background()
	allFailID, err := e.svc.RunSync(ctx, app.Request{Kind: core.KindScan, Endpoint: &config.Endpoint{Name: "dead", BaseURL: dead.URL + "/v1"}, Suite: quickSuite()})
	if err != nil {
		t.Fatalf("all-fail scan: %v", err)
	}
	suiteID, err := e.svc.RunSync(ctx, app.Request{Kind: core.KindSuite, EndpointID: ep.ID, Suite: quickSuite()})
	if err != nil {
		t.Fatalf("suite: %v", err)
	}
	latencyID, err := e.svc.RunSync(ctx, app.Request{Kind: core.KindLatency, EndpointID: ep.ID, Suite: quickSuite(), Models: []string{"a/fast"}})
	if err != nil {
		t.Fatalf("latency: %v", err)
	}
	loadSuite := quickSuite()
	loadSuite.Specs.Load.Levels, loadSuite.Specs.Load.Duration, loadSuite.Specs.Load.Requests, loadSuite.Specs.Load.OutputTokens = []int{1}, 300*time.Millisecond, 2, 4
	loadID, err := e.svc.RunSync(ctx, app.Request{Kind: core.KindLoad, EndpointID: ep.ID, Suite: loadSuite, Models: []string{"a/fast"}, Confirm: true})
	if err != nil {
		t.Fatalf("load: %v", err)
	}

	scenarios := []scenario{
		{name: "scan-discovering", runID: discovering.ID, paths: scanPaths(discovering.ID)},
		{name: "scan-running", runID: running.ID, paths: scanPaths(running.ID)},
		{name: "scan-failed", runID: failed.ID, paths: scanPaths(failed.ID)},
		{name: "scan-all-fail", runID: allFailID, paths: scanPaths(allFailID)},
		{name: "suite", runID: suiteID, paths: scanPaths(suiteID)},
		{name: "latency-only", runID: latencyID, paths: map[string]string{"run": "/api/runs/" + latencyID, "summary": "/api/runs/" + latencyID + "/summary", "latency": "/api/runs/" + latencyID + "/latency"}},
		{name: "load-only", runID: loadID, paths: map[string]string{"run": "/api/runs/" + loadID, "summary": "/api/runs/" + loadID + "/summary", "load": "/api/runs/" + loadID + "/load"}},
	}
	for _, p := range []string{"latency", "capabilities", "quality"} { // the quick suite has no load stage
		scenarios[4].paths[p] = "/api/runs/" + suiteID + "/" + p
	}
	// Drill-down endpoints for the first family / model of each scan scenario.
	for i := range scenarios[:5] {
		sc := &scenarios[i]
		var fam struct {
			Families []struct {
				Key string `json:"key"`
			} `json:"families"`
			Unavailable []struct {
				Key string `json:"key"`
			} `json:"unavailable"`
		}
		e.json("GET", sc.paths["families-composite"], nil, &fam)
		if len(fam.Families) > 0 {
			sc.paths["family"] = "/api/runs/" + sc.runID + "/families/" + fam.Families[0].Key
		} else if len(fam.Unavailable) > 0 {
			sc.paths["family"] = "/api/runs/" + sc.runID + "/families/" + fam.Unavailable[0].Key
		}
		var ms struct {
			Models []struct {
				ID string `json:"id"`
			} `json:"models"`
		}
		e.json("GET", sc.paths["models"], nil, &ms)
		if len(ms.Models) > 0 {
			sc.paths["model"] = "/api/runs/" + sc.runID + "/models/" + ms.Models[0].ID
			sc.paths["profile-model"] = "/api/runs/" + sc.runID + "/profile/model/" + ms.Models[0].ID
		}
		if fp, ok := sc.paths["family"]; ok {
			sc.paths["profile-family"] = strings.Replace(fp, "/families/", "/profile/family/", 1)
		}
	}
	global := map[string]string{
		"version": "/api/version", "reasons": "/api/reasons", "probes": "/api/probes", "presets": "/api/presets", "dialects": "/api/dialects",
		"endpoints": "/api/endpoints", "runs": "/api/runs?limit=50", "compare": "/api/compare?a=" + suiteID + "&b=" + allFailID,
	}

	fixtureRoot := filepath.Join("..", "..", "web", "src", "test", "fixtures")
	write := func(dir, name string, raw []byte) {
		if !*updateFixtures {
			return
		}
		if err := os.MkdirAll(filepath.Join(fixtureRoot, dir), 0o750); err != nil {
			t.Fatal(err)
		}
		var pretty json.RawMessage = raw
		if out, err := json.MarshalIndent(json.RawMessage(raw), "", "  "); err == nil {
			pretty = out
		}
		if err := os.WriteFile(filepath.Join(fixtureRoot, dir, name+".json"), append(pretty, '\n'), 0o644); err != nil { //nolint:gosec // test fixtures
			t.Fatal(err)
		}
	}
	check := func(dir, name, path string) {
		code, raw := e.do(http.MethodGet, path, nil)
		if code != 200 {
			t.Errorf("%s/%s GET %s: status %d: %s", dir, name, path, code, raw)
			return
		}
		assertNoNullCollections(t, dir+"/"+name+" ("+path+")", raw)
		write(dir, name, raw)
	}
	for _, sc := range scenarios {
		for name, path := range sc.paths {
			check(sc.name, name, path)
		}
		// Markdown report must render for every state (not JSON, so only the status is checked).
		if code, raw := e.do(http.MethodGet, "/api/runs/"+sc.runID+"/report.md?lang=en", nil); code != 200 || len(raw) == 0 {
			t.Errorf("%s report.md: status %d", sc.name, code)
		} else if *updateFixtures {
			_ = os.MkdirAll(filepath.Join(fixtureRoot, sc.name), 0o750)
			_ = os.WriteFile(filepath.Join(fixtureRoot, sc.name, "report.md"), raw, 0o644) //nolint:gosec // test fixtures
		}
	}
	for name, path := range global {
		check("global", name, path)
	}
	// Discover and check against the mock endpoint.
	for name, path := range map[string]string{"check": "/api/endpoints/" + ep.ID + "/check", "discover": "/api/endpoints/" + ep.ID + "/discover"} {
		code, raw := e.do(http.MethodPost, path, nil)
		if code != 200 {
			t.Errorf("global/%s: status %d: %s", name, code, raw)
			continue
		}
		assertNoNullCollections(t, "global/"+name, raw)
		write("global", name, raw)
	}
	if *updateFixtures {
		var index strings.Builder
		for _, sc := range scenarios {
			fmt.Fprintf(&index, "%s\t%s\n", sc.name, sc.runID)
		}
		_ = os.WriteFile(filepath.Join(fixtureRoot, "scenarios.txt"), []byte(index.String()), 0o644) //nolint:gosec // test fixtures
	}
}
