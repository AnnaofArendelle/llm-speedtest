package server

import (
	"context"
	"net/http"
	"net/url"
	"testing"

	"github.com/AnnaofArendelle/llm-speedtest/internal/app"
	"github.com/AnnaofArendelle/llm-speedtest/internal/config"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
)

type profileResp struct {
	ID         string             `json:"id"`
	Kind       string             `json:"kind"`
	Dimensions []profileDimension `json:"dimensions"`
	Coverage   map[string]bool    `json:"coverage"`
	Measured   map[string]int     `json:"measured"`
	Members    int                `json:"members"`
}

func dim(p profileResp, key string) profileDimension {
	for _, d := range p.Dimensions {
		if d.Key == key {
			return d
		}
	}
	return profileDimension{}
}

func TestProfileEndpoints(t *testing.T) {
	e := newEnv(t, "")
	ctx := context.Background()
	suiteID, err := e.svc.RunSync(ctx, app.Request{Kind: core.KindSuite, Endpoint: &config.Endpoint{Name: "mock", BaseURL: e.base}, Suite: quickSuite()})
	if err != nil {
		t.Fatal(err)
	}
	scanID, err := e.svc.RunSync(ctx, app.Request{Kind: core.KindScan, Endpoint: &config.Endpoint{Name: "mock", BaseURL: e.base}, Suite: quickSuite()})
	if err != nil {
		t.Fatal(err)
	}

	var p profileResp
	if code := e.json(http.MethodGet, "/api/runs/"+suiteID+"/profile/model/a/fast", nil, &p); code != 200 {
		t.Fatalf("model profile: %d", code)
	}
	if p.Kind != "model" || len(p.Dimensions) != 8 || p.Members != 1 {
		t.Fatalf("profile = %+v", p)
	}
	if d := dim(p, "stability"); !d.Covered || d.Score != 100 {
		t.Fatalf("stability = %+v", d)
	}
	for _, key := range []string{"speed", "breadth", "ttft", "throughput", "capability", "quality"} {
		if d := dim(p, key); !d.Covered || d.Score < 0 || d.Score > 100 || d.Detail == "" {
			t.Fatalf("%s = %+v (suite should cover it)", key, d)
		}
	}
	if d := dim(p, "load"); d.Covered {
		t.Fatalf("load must not be covered without a load stage: %+v", d)
	}
	if !p.Coverage["scan"] || !p.Coverage["latency"] || !p.Coverage["capability"] || !p.Coverage["quality"] || p.Coverage["load"] {
		t.Fatalf("coverage = %v", p.Coverage)
	}

	// The failing model has a stability score of 0 and no stage coverage.
	if code := e.json(http.MethodGet, "/api/runs/"+suiteID+"/profile/model/b/missing", nil, &p); code != 200 {
		t.Fatalf("failing model profile: %d", code)
	}
	if d := dim(p, "stability"); !d.Covered || d.Score != 0 {
		t.Fatalf("failing stability = %+v", d)
	}
	if d := dim(p, "ttft"); d.Covered {
		t.Fatalf("failing model must not have latency coverage: %+v", d)
	}

	// Scan-only runs cover only scan dimensions.
	if code := e.json(http.MethodGet, "/api/runs/"+scanID+"/profile/model/a/fast", nil, &p); code != 200 {
		t.Fatalf("scan profile: %d", code)
	}
	if !dim(p, "speed").Covered || dim(p, "ttft").Covered || dim(p, "capability").Covered {
		t.Fatalf("scan-only coverage wrong: %+v", p.Dimensions)
	}

	// Family profile aggregates members.
	var fams struct {
		Families []struct {
			Key string `json:"key"`
		} `json:"families"`
	}
	e.json(http.MethodGet, "/api/runs/"+suiteID+"/families", nil, &fams)
	if len(fams.Families) == 0 {
		t.Fatal("no families")
	}
	if code := e.json(http.MethodGet, "/api/runs/"+suiteID+"/profile/family/"+url.PathEscape(fams.Families[0].Key), nil, &p); code != 200 {
		t.Fatalf("family profile: %d", code)
	}
	if p.Kind != "family" || p.Members < 1 || len(p.Dimensions) != 8 || !dim(p, "stability").Covered {
		t.Fatalf("family profile = %+v", p)
	}
	if code, _ := e.do(http.MethodGet, "/api/runs/"+suiteID+"/profile/model/nope", nil); code != 404 {
		t.Fatalf("unknown model: %d", code)
	}
	if code, _ := e.do(http.MethodGet, "/api/runs/"+suiteID+"/profile/family/nope", nil); code != 404 {
		t.Fatalf("unknown family: %d", code)
	}
}
