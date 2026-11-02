package server

import (
	"encoding/json"
	"net/http"
	"testing"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/testutil/mockllm"
)

func TestSkipEndpoint(t *testing.T) {
	e := newEnv(t, "")
	e.llm.SetBehavior("a/slow", mockllm.Behavior{Hang: 30 * time.Second, Content: "ok"})
	suite := quickSuite()
	suite.Specs.Scan.First.Timeout = 20 * time.Second
	var started struct {
		ID string `json:"id"`
	}
	if code := e.json(http.MethodPost, "/api/runs", map[string]any{"kind": core.KindScan, "endpoint": map[string]any{"base_url": e.base}, "suite": suite}, &started); code != 202 {
		t.Fatalf("start: %d", code)
	}
	// Wait until the slow model is in flight.
	deadline := time.Now().Add(10 * time.Second)
	var seen bool
	for time.Now().Before(deadline) && !seen {
		var doc struct {
			Inflight []core.InflightItem `json:"inflight"`
		}
		e.json(http.MethodGet, "/api/runs/"+started.ID, nil, &doc)
		for _, it := range doc.Inflight {
			if it.ID == "a/slow" && it.Stage == "scan" {
				seen = true
			}
		}
		if !seen {
			time.Sleep(20 * time.Millisecond)
		}
	}
	if !seen {
		t.Fatal("slow model never appeared in the in-flight list")
	}
	var skipped struct {
		Skipped []string `json:"skipped"`
	}
	if code := e.json(http.MethodPost, "/api/runs/"+started.ID+"/skip", map[string]any{"model": "a/slow"}, &skipped); code != 200 || len(skipped.Skipped) != 1 || skipped.Skipped[0] != "a/slow" {
		t.Fatalf("skip: %d %+v", code, skipped)
	}
	// Reading results while the run is still in progress must not poison the
	// result cache that serves the finished run (regression).
	e.do(http.MethodGet, "/api/runs/"+started.ID+"/summary", nil)
	e.do(http.MethodGet, "/api/runs/"+started.ID+"/failures", nil)
	deadline = time.Now().Add(15 * time.Second)
	for time.Now().Before(deadline) {
		var doc struct {
			Run struct {
				Status string `json:"status"`
			} `json:"run"`
			Running bool `json:"running"`
		}
		e.json(http.MethodGet, "/api/runs/"+started.ID, nil, &doc)
		if !doc.Running && doc.Run.Status == core.StatusDone {
			break
		}
		time.Sleep(50 * time.Millisecond)
	}
	var models struct {
		Models []struct {
			ID      string `json:"id"`
			Status  string `json:"status"`
			Reason  string `json:"reason"`
			Skipped int    `json:"skipped"`
		} `json:"models"`
	}
	e.json(http.MethodGet, "/api/runs/"+started.ID+"/models", nil, &models)
	byID := map[string]string{}
	for _, m := range models.Models {
		byID[m.ID] = m.Status + "/" + m.Reason
		if m.ID == "a/slow" && m.Skipped != 1 {
			t.Fatalf("a/slow skipped count = %d", m.Skipped)
		}
	}
	if byID["a/slow"] != "fail/skipped" || byID["a/fast"] != "ok/ok" {
		t.Fatalf("models after skip: %v", byID)
	}
	var groups struct {
		Groups []struct {
			Reason string `json:"reason"`
			Count  int    `json:"count"`
		} `json:"groups"`
	}
	e.json(http.MethodGet, "/api/runs/"+started.ID+"/failures", nil, &groups)
	found := false
	for _, g := range groups.Groups {
		if g.Reason == "skipped" && g.Count == 1 {
			found = true
		}
	}
	if !found {
		t.Fatalf("failures should contain a skipped group: %+v", groups.Groups)
	}
	// Skipping a finished run is rejected.
	if code, _ := e.do(http.MethodPost, "/api/runs/"+started.ID+"/skip", map[string]any{"all": true}); code != 409 {
		t.Fatalf("skip after completion: %d", code)
	}
	raw, _ := json.Marshal(models)
	if len(raw) == 0 {
		t.Fatal("unreachable")
	}
}
