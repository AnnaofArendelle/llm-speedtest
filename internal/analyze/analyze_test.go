package analyze

import (
	"testing"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/testutil/refscan"
)

func TestBuildReferenceScan(t *testing.T) {
	models, attempts := refscan.Load(t)
	if len(attempts) != 2939 {
		t.Fatalf("attempts = %d", len(attempts))
	}
	pres := Enrich(models, attempts, Options{EndpointName: "9router"})
	if pres.Field != "owned_by" {
		t.Fatalf("provider field = %s (%s)", pres.Field, pres.Rationale)
	}
	res := Build(models, attempts, DefaultScoring(), &pres)
	s := res.Summary
	if s.Models != 2306 || s.OK != 138 || s.Fail != 2168 {
		t.Fatalf("summary = %+v", s)
	}
	if s.TierCounts["A"] != 110 || s.TierCounts["B"] != 8 || s.TierCounts["C"] != 20 {
		t.Fatalf("tiers = %+v", s.TierCounts)
	}
	want := map[string]int{"auth": 717, "quota_billing": 597, "no_credentials": 255, "quota_exhausted": 156, "plan_restriction": 150,
		"model_unavailable": 99, "not_chat_model": 43, "not_chat_model_inferred": 32, "upstream_5xx": 24, "model_retired": 23, "rate_limit": 19,
		"no_channel": 19, "empty_response": 14, "no_capacity": 7, "method_not_allowed": 3, "param_incompatible": 3, "upstream_network": 3,
		"terms_required": 2, "timeout": 1, "bad_request": 1}
	for k, v := range want {
		if s.Reasons[k] != v {
			t.Errorf("reason %s = %d, want %d", k, s.Reasons[k], v)
		}
	}
	if s.ProviderLevelFail != 1877 {
		t.Errorf("provider-level failures = %d", s.ProviderLevelFail)
	}
	if s.Families < 495 || s.Families > 515 || s.FamiliesOK < 70 || s.FamiliesOK > 75 {
		t.Errorf("families = %d (ok %d)", s.Families, s.FamiliesOK)
	}
	if s.Providers != 40 {
		t.Errorf("providers = %d", s.Providers)
	}
	// top of the composite ranking must be a stable, multi-provider family
	top := res.Families[0]
	if top.Tier != "A" || top.NProvOK < 2 || top.Score < 75 {
		t.Errorf("top family = %+v", top)
	}
	if res.Families[0].RankComposite != 1 || res.FamiliesByStability[0].RankStability != 1 {
		t.Error("ranks")
	}
	ds := res.Family("deepseek|flash|4")
	if ds == nil || ds.N < 75 || ds.OK != 8 || ds.NProvOK != 5 || ds.Tier != "B" {
		t.Fatalf("deepseek family = %+v", ds)
	}
	// the fastest stable model is first in the model ranking
	if res.Available[0].Tier != "A" || res.Available[0].Med > res.Available[len(res.Available)-1].Med && res.Available[len(res.Available)-1].Tier == "A" {
		t.Errorf("model ranking order")
	}
	// provider stats: kilo is entirely account-level unavailable
	var kilo *ProviderStat
	for _, p := range res.Providers {
		if p.Key == "kilo" {
			kilo = p
		}
	}
	if kilo == nil || kilo.Models != 380 || kilo.OK != 0 || !kilo.AccountLevelAll || kilo.TopReason != "auth" {
		t.Fatalf("kilo = %+v", kilo)
	}
	// deterministic
	res2 := Build(models, attempts, DefaultScoring(), &pres)
	for i := range res.Families {
		if res.Families[i].Key != res2.Families[i].Key {
			t.Fatal("non-deterministic family order")
		}
	}
	if m := res.Model("thai/deepseek-v4.1-flash:free"); m == nil || m.Status != "ok" || m.ProviderName != "thai" {
		t.Fatalf("model lookup = %+v", m)
	}
}

func TestScoringNormalize(t *testing.T) {
	var s Scoring
	s.Normalize()
	d := DefaultScoring()
	if s.WStab != d.WStab || s.SpeedHi != d.SpeedHi || s.ModelTierB != d.ModelTierB {
		t.Fatalf("defaults = %+v", s)
	}
	s = Scoring{WStab: 2, WBreadth: 1, WSpeed: 1}
	s.Normalize()
	if s.WStab != 0.5 || s.WBreadth != 0.25 {
		t.Fatalf("rescale = %+v", s)
	}
	if modelTier(1, d) != "A" || modelTier(0.7, d) != "B" || modelTier(0.3, d) != "C" || familyTier(0.96, d) != "A" {
		t.Fatal("tiers")
	}
	if TierName("A", "zh") == "" || TierName("B", "en") == "" {
		t.Fatal("names")
	}
}

func TestBuildEdgeCases(t *testing.T) {
	models := []core.RunModel{{ID: "a/x", OwnedBy: "a"}, {ID: "a/y", OwnedBy: "a"}, {ID: "b/z", OwnedBy: "b"}}
	attempts := []core.Attempt{
		{ModelID: "a/x", Round: 1, OK: true, Latency: time.Second, Reason: "ok", Class: "ok", ServedModel: "x-2"},
		{ModelID: "a/x", Round: 3, OK: false, Reason: "rate_limit", Class: "rate_limit", Latency: time.Second},
		{ModelID: "a/y", Round: 1, OK: false, Reason: "auth", Class: "auth", HTTP: 503, Inner: 401, Message: "bad token"},
	}
	Enrich(models, attempts, Options{})
	res := Build(models, attempts, Scoring{}, nil)
	x := res.Model("a/x")
	if x.Status != "ok" || x.Tier != "C" || x.Flaky["rate_limit"] != 1 || x.Served != "x-2" || x.Attempts != 2 {
		t.Fatalf("x = %+v", x)
	}
	if y := res.Model("a/y"); y.Status != "fail" || y.Reason != "auth" || y.HTTP != 503 || y.Inner != 401 {
		t.Fatalf("y = %+v", y)
	}
	if z := res.Model("b/z"); z.Status != "untested" || z.Reason != "untested" {
		t.Fatalf("z = %+v", z)
	}
	if res.Summary.OK != 1 || res.Summary.Fail != 2 || len(res.Summary.Buckets) != 6 || res.Summary.Buckets[1].Count != 1 {
		t.Fatalf("summary = %+v", res.Summary)
	}
}
