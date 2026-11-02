// Package analyze aggregates scan attempts into model, family and provider
// statistics, assigns stability tiers and composite scores, and produces the
// deterministic orderings used by the dashboard and the reports.
package analyze

import (
	"sort"
	"strings"

	"github.com/AnnaofArendelle/llm-speedtest/internal/classify"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/family"
	"github.com/AnnaofArendelle/llm-speedtest/internal/metrics"
	"github.com/AnnaofArendelle/llm-speedtest/internal/modelid"
	"github.com/AnnaofArendelle/llm-speedtest/internal/provider"
)

// Scoring holds the tunable thresholds and weights.
type Scoring struct {
	WStab       float64 `json:"w_stability" yaml:"w_stability"`
	WBreadth    float64 `json:"w_breadth" yaml:"w_breadth"`
	WSpeed      float64 `json:"w_speed" yaml:"w_speed"`
	SpeedLo     float64 `json:"speed_lo_s" yaml:"speed_lo_s"` // latency scoring 100 at or below
	SpeedHi     float64 `json:"speed_hi_s" yaml:"speed_hi_s"` // latency scoring 0 at or above
	ModelTierA  float64 `json:"model_tier_a" yaml:"model_tier_a"`
	ModelTierB  float64 `json:"model_tier_b" yaml:"model_tier_b"`
	FamilyTierA float64 `json:"family_tier_a" yaml:"family_tier_a"`
	FamilyTierB float64 `json:"family_tier_b" yaml:"family_tier_b"`
}

// DefaultScoring returns the reference weights.
func DefaultScoring() Scoring {
	return Scoring{WStab: 0.4, WBreadth: 0.3, WSpeed: 0.3, SpeedLo: 0.5, SpeedHi: 30, ModelTierA: 0.999, ModelTierB: 0.66, FamilyTierA: 0.95, FamilyTierB: 0.66}
}

// Normalize fills zero weights/thresholds with defaults and rescales the
// weights to sum to one.
func (s *Scoring) Normalize() {
	d := DefaultScoring()
	if s.WStab <= 0 && s.WBreadth <= 0 && s.WSpeed <= 0 {
		s.WStab, s.WBreadth, s.WSpeed = d.WStab, d.WBreadth, d.WSpeed
	}
	sum := s.WStab + s.WBreadth + s.WSpeed
	if sum > 0 {
		s.WStab, s.WBreadth, s.WSpeed = s.WStab/sum, s.WBreadth/sum, s.WSpeed/sum
	}
	if s.SpeedLo <= 0 {
		s.SpeedLo = d.SpeedLo
	}
	if s.SpeedHi <= s.SpeedLo {
		s.SpeedHi = d.SpeedHi
		if s.SpeedHi <= s.SpeedLo {
			s.SpeedHi = s.SpeedLo * 60
		}
	}
	if s.ModelTierA <= 0 {
		s.ModelTierA = d.ModelTierA
	}
	if s.ModelTierB <= 0 {
		s.ModelTierB = d.ModelTierB
	}
	if s.FamilyTierA <= 0 {
		s.FamilyTierA = d.FamilyTierA
	}
	if s.FamilyTierB <= 0 {
		s.FamilyTierB = d.FamilyTierB
	}
}

// Tier ranks.
var tierRank = map[string]int{"A": 0, "B": 1, "C": 2}

// ModelStat is the aggregated result of one model.
type ModelStat struct {
	ID             string         `json:"id"`
	OwnedBy        string         `json:"owned_by"`
	Upstream       string         `json:"upstream"`
	ProviderKey    string         `json:"provider_key"`
	ProviderName   string         `json:"provider_name"`
	FamilyKey      string         `json:"family_key"`
	FamilyCanon    string         `json:"family_canon"`
	FamilySource   string         `json:"family_source"`
	FamilyEvidence string         `json:"family_evidence,omitempty"`
	ContextLength  int            `json:"context_length,omitempty"`
	Status         string         `json:"status"` // ok | fail | untested
	Attempts       int            `json:"attempts"`
	OK             int            `json:"ok"`
	Rate           float64        `json:"rate"`
	Tier           string         `json:"tier,omitempty"`
	Min            float64        `json:"min_s"`
	Med            float64        `json:"median_s"`
	Max            float64        `json:"max_s"`
	Jitter         float64        `json:"jitter"` // max/min
	Served         string         `json:"served_model,omitempty"`
	Flaky          map[string]int `json:"flaky,omitempty"` // failure reasons seen on an available model
	Reason         string         `json:"reason"`          // final reason ("ok" when available)
	FirstReason    string         `json:"first_reason,omitempty"`
	HTTP           int            `json:"http,omitempty"`
	Inner          int            `json:"inner,omitempty"`
	Message        string         `json:"message,omitempty"`
	FailLatency    float64        `json:"fail_latency_s,omitempty"`
	Skipped        int            `json:"skipped,omitempty"` // attempts cancelled by the user (excluded from the statistics)
	History        []core.Attempt `json:"-"`
}

// FamilyStat is the aggregated result of one family.
type FamilyStat struct {
	Key           string         `json:"key"`
	Display       string         `json:"display"`
	Members       []*ModelStat   `json:"-"`
	MemberIDs     []string       `json:"member_ids"`
	N             int            `json:"n"`
	OK            int            `json:"ok"`
	Providers     map[string]int `json:"providers"`
	OKProviders   map[string]int `json:"ok_providers"`
	NProvOK       int            `json:"n_prov_ok"`
	NProvStable   int            `json:"n_prov_stable"`
	Attempts      int            `json:"attempts"`
	Successes     int            `json:"successes"`
	Rate          float64        `json:"rate"`
	Tier          string         `json:"tier,omitempty"`
	FMin          float64        `json:"min_s"`
	FMed          float64        `json:"median_s"`
	FMax          float64        `json:"max_s"`
	BestID        string         `json:"best_id,omitempty"`
	Best          *ModelStat     `json:"-"`
	Stable        []*ModelStat   `json:"-"`
	StableIDs     []string       `json:"stable_ids"`
	Score         float64        `json:"score"`
	SStab         float64        `json:"s_stability"`
	SBreadth      float64        `json:"s_breadth"`
	SSpeed        float64        `json:"s_speed"`
	Reasons       map[string]int `json:"reasons"`
	Sources       map[string]int `json:"sources"` // family assignment sources among members
	RankComposite int            `json:"rank_composite,omitempty"`
	RankStability int            `json:"rank_stability,omitempty"`
}

// ProviderStat is the aggregated result of one provider.
type ProviderStat struct {
	Key             string              `json:"key"`
	Name            string              `json:"name"`
	Models          int                 `json:"models"`
	Families        int                 `json:"families"`
	OK              int                 `json:"ok"`
	Rate            float64             `json:"rate"`
	Med             float64             `json:"median_s"`
	TopReason       string              `json:"top_reason,omitempty"`
	TopReasonCount  int                 `json:"top_reason_count,omitempty"`
	Reasons         map[string]int      `json:"reasons"`
	AccountLevelAll bool                `json:"account_level_all"` // every model fails for the same provider-level reason
	Evidence        []provider.Evidence `json:"evidence,omitempty"`
}

// Bucket is a latency histogram bucket.
type Bucket struct {
	Label string  `json:"label"`
	Lo    float64 `json:"lo"`
	Hi    float64 `json:"hi"`
	Count int     `json:"count"`
}

// Summary holds the headline numbers.
type Summary struct {
	Models            int            `json:"models"`
	OK                int            `json:"ok"`
	Fail              int            `json:"fail"`
	ProviderLevelFail int            `json:"provider_level_fail"`
	Families          int            `json:"families"`
	FamiliesOK        int            `json:"families_ok"`
	FamiliesUnavail   int            `json:"families_unavailable"`
	Providers         int            `json:"providers"`
	TierCounts        map[string]int `json:"tier_counts"`
	FamilyTierCounts  map[string]int `json:"family_tier_counts"`
	LatencyMedian     float64        `json:"latency_median_s"`
	LatencyP90        float64        `json:"latency_p90_s"`
	LatencyMin        float64        `json:"latency_min_s"`
	LatencyMax        float64        `json:"latency_max_s"`
	Buckets           []Bucket       `json:"buckets"`
	Reasons           map[string]int `json:"reasons"`
	Attempts          int            `json:"attempts"`
	Rounds            int            `json:"rounds"`
}

// Result is the complete analysis of a scan.
type Result struct {
	Summary             Summary               `json:"summary"`
	Scoring             Scoring               `json:"scoring"`
	Models              []*ModelStat          `json:"models"`                // all models, available first (tier, median), then failures
	Available           []*ModelStat          `json:"-"`                     // ok models ranked by tier then median
	Families            []*FamilyStat         `json:"families"`              // available families in composite order
	FamiliesByStability []*FamilyStat         `json:"families_by_stability"` // available families by tier then median
	UnavailableFamilies []*FamilyStat         `json:"unavailable_families"`
	Providers           []*ProviderStat       `json:"providers"`
	ProviderField       string                `json:"provider_field"`
	ProviderRationale   string                `json:"provider_rationale"`
	Suggestions         []provider.Suggestion `json:"provider_suggestions,omitempty"`
	familyByKey         map[string]*FamilyStat
	modelByID           map[string]*ModelStat
}

// Family returns the family stat for a key.
func (r *Result) Family(key string) *FamilyStat { return r.familyByKey[key] }

// Model returns the model stat for an id.
func (r *Result) Model(id string) *ModelStat { return r.modelByID[id] }

// Options control provider and family assignment.
type Options struct {
	EndpointName      string
	ProviderOverrides map[string]provider.Override
	FamilyOverrides   map[string]family.Override
	FamilyRules       *family.Rules
}

// Enrich resolves providers and families for the models using the attempts
// as evidence (connection identifiers from errors, served model names from
// successes). It mutates models in place and returns the provider result.
func Enrich(models []core.RunModel, attempts []core.Attempt, opt Options) provider.Result {
	conns := map[string]string{}
	served := map[string]string{}
	for _, a := range attempts {
		if a.Conn != "" {
			if _, ok := conns[a.ModelID]; !ok {
				conns[a.ModelID] = a.Conn
			}
		}
		if a.OK && a.ServedModel != "" {
			if _, ok := served[a.ModelID]; !ok {
				served[a.ModelID] = a.ServedModel
			}
		}
	}
	pm := make([]provider.Model, 0, len(models))
	fm := make([]family.Model, 0, len(models))
	for _, m := range models {
		pm = append(pm, provider.Model{ID: m.ID, OwnedBy: m.OwnedBy})
		fm = append(fm, family.Model{ID: m.ID, OwnedBy: m.OwnedBy, ContextLength: m.ContextLength})
	}
	pres := provider.Resolve(pm, conns, opt.ProviderOverrides, opt.EndpointName)
	fas := family.Learn(opt.FamilyRules, fm, served, opt.FamilyOverrides)
	for i := range models {
		m := &models[i]
		if m.Upstream == "" {
			m.Upstream = modelid.Upstream(m.ID, m.OwnedBy)
		}
		if pa, ok := pres.Models[m.ID]; ok {
			m.ProviderKey, m.ProviderName = pa.Key, pa.Name
		}
		if fa, ok := fas[m.ID]; ok {
			m.FamilyKey, m.FamilyCanon, m.FamilySource, m.FamilyEvidence = fa.Key, fa.Canon, fa.Source, fa.Evidence
		}
	}
	return pres
}

// Build aggregates attempts for the (enriched) models.
func Build(models []core.RunModel, attempts []core.Attempt, sc Scoring, pres *provider.Result) *Result {
	sc.Normalize()
	res := &Result{
		Scoring:             sc,
		familyByKey:         map[string]*FamilyStat{},
		modelByID:           map[string]*ModelStat{},
		Families:            []*FamilyStat{},
		FamiliesByStability: []*FamilyStat{},
		UnavailableFamilies: []*FamilyStat{},
		Available:           []*ModelStat{},
		Models:              []*ModelStat{},
		Providers:           []*ProviderStat{},
	}
	if pres != nil {
		res.ProviderField, res.ProviderRationale, res.Suggestions = pres.Field, pres.Rationale, pres.Suggestions
	}
	byModel := map[string][]core.Attempt{}
	rounds := 0
	for _, a := range attempts {
		byModel[a.ModelID] = append(byModel[a.ModelID], a)
		if a.Round > rounds {
			rounds = a.Round
		}
	}
	res.Summary.Attempts = len(attempts)
	res.Summary.Rounds = rounds
	provEvidence := map[string][]provider.Evidence{}
	if pres != nil {
		for id, a := range pres.Models {
			if _, ok := provEvidence[a.Key]; !ok {
				provEvidence[a.Key] = a.Evidence
			}
			_ = id
		}
	}

	for _, m := range models {
		as := byModel[m.ID]
		sort.SliceStable(as, func(i, j int) bool {
			if as[i].Round != as[j].Round {
				return as[i].Round < as[j].Round
			}
			return as[i].At.Before(as[j].At)
		})
		// Attempts cancelled through "skip" are excluded from the availability
		// and latency statistics; they are only counted.
		full := as
		live := make([]core.Attempt, 0, len(as))
		skipped := 0
		for _, a := range as {
			if a.Reason == classify.ReasonSkipped {
				skipped++
				continue
			}
			live = append(live, a)
		}
		as = live
		ms := &ModelStat{ID: m.ID, OwnedBy: m.OwnedBy, Upstream: m.Upstream, ProviderKey: m.ProviderKey, ProviderName: m.ProviderName,
			FamilyKey: m.FamilyKey, FamilyCanon: m.FamilyCanon, FamilySource: m.FamilySource, FamilyEvidence: m.FamilyEvidence,
			ContextLength: m.ContextLength, Attempts: len(as), History: full, Skipped: skipped}
		if ms.ProviderName == "" {
			ms.ProviderName = ms.ProviderKey
		}
		var lat []float64
		for _, a := range as {
			if a.OK {
				ms.OK++
				lat = append(lat, a.Latency.Seconds())
				if ms.Served == "" {
					ms.Served = a.ServedModel
				}
			}
		}
		switch {
		case ms.OK > 0:
			ms.Status = "ok"
			ms.Reason = "ok"
			st := metrics.Summarise(lat)
			ms.Min, ms.Med, ms.Max = st.Min, st.Median, st.Max
			if ms.Min > 0 {
				ms.Jitter = ms.Max / ms.Min
			}
			ms.Rate = metrics.Ratio(ms.OK, ms.Attempts)
			ms.Tier = modelTier(ms.Rate, sc)
			for _, a := range as {
				if !a.OK {
					if ms.Flaky == nil {
						ms.Flaky = map[string]int{}
					}
					ms.Flaky[a.Reason]++
				}
			}
		case len(as) > 0:
			ms.Status = "fail"
			last := as[len(as)-1]
			ms.Reason, ms.HTTP, ms.Inner, ms.Message, ms.FailLatency = last.Reason, last.HTTP, last.Inner, last.Message, last.Latency.Seconds()
			ms.FirstReason = as[0].Reason
			if ms.Reason == "" {
				ms.Reason = "unknown"
			}
		case ms.Skipped > 0:
			ms.Status, ms.Reason, ms.Message = "fail", classify.ReasonSkipped, "skipped by user"
		default:
			ms.Status = "untested"
			ms.Reason = "untested"
		}
		res.Models = append(res.Models, ms)
		res.modelByID[ms.ID] = ms
	}

	// Families
	for _, ms := range res.Models {
		key := ms.FamilyKey
		if key == "" {
			key = "unknown||"
		}
		f := res.familyByKey[key]
		if f == nil {
			f = &FamilyStat{Key: key, MemberIDs: []string{}, StableIDs: []string{}, Providers: map[string]int{}, OKProviders: map[string]int{}, Reasons: map[string]int{}, Sources: map[string]int{}}
			res.familyByKey[key] = f
		}
		f.Members = append(f.Members, ms)
		f.MemberIDs = append(f.MemberIDs, ms.ID)
		f.N++
		f.Providers[ms.ProviderName]++
		f.Sources[ms.FamilySource]++
		if ms.Status == "ok" {
			f.OK++
			f.OKProviders[ms.ProviderName]++
			f.Attempts += ms.Attempts
			f.Successes += ms.OK
		} else {
			f.Reasons[ms.Reason]++
		}
	}
	for _, f := range res.familyByKey {
		var canons []string
		for _, m := range f.Members {
			canons = append(canons, m.FamilyCanon)
		}
		f.Display = family.Display(canons)
		if f.Display == "" {
			f.Display = strings.ReplaceAll(f.Key, "|", " ")
		}
		if f.OK == 0 {
			continue
		}
		oks := make([]*ModelStat, 0, f.OK)
		for _, m := range f.Members {
			if m.Status == "ok" {
				oks = append(oks, m)
			}
		}
		sortModels(oks)
		meds := make([]float64, 0, len(oks))
		f.FMin, f.FMax = oks[0].Min, 0
		for _, m := range oks {
			meds = append(meds, m.Med)
			if m.Min < f.FMin {
				f.FMin = m.Min
			}
			if m.Max > f.FMax {
				f.FMax = m.Max
			}
		}
		f.FMed = metrics.Median(meds)
		f.Rate = metrics.Ratio(f.Successes, f.Attempts)
		f.Tier = familyTier(f.Rate, sc)
		stableProv := map[string]bool{}
		for _, m := range oks {
			if m.Tier == "A" {
				f.Stable = append(f.Stable, m)
				f.StableIDs = append(f.StableIDs, m.ID)
				stableProv[m.ProviderName] = true
			}
		}
		f.NProvOK = len(f.OKProviders)
		f.NProvStable = len(stableProv)
		if len(f.Stable) > 0 {
			f.Best = f.Stable[0]
		} else {
			f.Best = oks[0]
		}
		f.BestID = f.Best.ID
	}
	avail, unavail := []*FamilyStat{}, []*FamilyStat{}
	maxProv := 1
	for _, f := range res.familyByKey {
		if f.OK > 0 {
			avail = append(avail, f)
			if f.NProvOK > maxProv {
				maxProv = f.NProvOK
			}
		} else {
			unavail = append(unavail, f)
		}
	}
	for _, f := range avail {
		f.SStab = 100 * f.Rate
		f.SBreadth = 100 * float64(f.NProvOK) / float64(maxProv)
		f.SSpeed = metrics.LogScore(f.FMed, sc.SpeedLo, sc.SpeedHi)
		f.Score = metrics.Round(sc.WStab*f.SStab+sc.WBreadth*f.SBreadth+sc.WSpeed*f.SSpeed, 3)
	}
	sort.Slice(avail, func(i, j int) bool {
		a, b := avail[i], avail[j]
		if a.Score != b.Score {
			return a.Score > b.Score
		}
		if a.NProvOK != b.NProvOK {
			return a.NProvOK > b.NProvOK
		}
		if r1, r2 := metrics.Round(a.FMed, 3), metrics.Round(b.FMed, 3); r1 != r2 {
			return r1 < r2
		}
		return a.Key < b.Key
	})
	for i, f := range avail {
		f.RankComposite = i + 1
	}
	res.Families = avail
	byStab := make([]*FamilyStat, len(avail))
	copy(byStab, avail)
	sort.Slice(byStab, func(i, j int) bool {
		a, b := byStab[i], byStab[j]
		if tierRank[a.Tier] != tierRank[b.Tier] {
			return tierRank[a.Tier] < tierRank[b.Tier]
		}
		if r1, r2 := metrics.Round(a.FMed, 3), metrics.Round(b.FMed, 3); r1 != r2 {
			return r1 < r2
		}
		if a.Rate != b.Rate {
			return a.Rate > b.Rate
		}
		return a.Key < b.Key
	})
	for i, f := range byStab {
		f.RankStability = i + 1
	}
	res.FamiliesByStability = byStab
	sort.Slice(unavail, func(i, j int) bool {
		if unavail[i].N != unavail[j].N {
			return unavail[i].N > unavail[j].N
		}
		return unavail[i].Display < unavail[j].Display
	})
	res.UnavailableFamilies = unavail

	// Model rankings
	oks, fails := []*ModelStat{}, []*ModelStat{}
	for _, m := range res.Models {
		if m.Status == "ok" {
			oks = append(oks, m)
		} else {
			fails = append(fails, m)
		}
	}
	sortModels(oks)
	sort.Slice(fails, func(i, j int) bool {
		if fails[i].Reason != fails[j].Reason {
			return fails[i].Reason < fails[j].Reason
		}
		return fails[i].ID < fails[j].ID
	})
	res.Available = oks
	if res.Available == nil {
		res.Available = []*ModelStat{}
	}
	if res.Families == nil {
		res.Families = []*FamilyStat{}
	}
	if res.UnavailableFamilies == nil {
		res.UnavailableFamilies = []*FamilyStat{}
	}
	res.Models = append(append([]*ModelStat{}, oks...), fails...)

	// Providers
	provs := map[string]*ProviderStat{}
	provFam := map[string]map[string]bool{}
	for _, m := range res.Models {
		p := provs[m.ProviderKey]
		if p == nil {
			p = &ProviderStat{Key: m.ProviderKey, Name: m.ProviderName, Reasons: map[string]int{}, Evidence: provEvidence[m.ProviderKey]}
			provs[m.ProviderKey] = p
			provFam[m.ProviderKey] = map[string]bool{}
		}
		p.Models++
		provFam[m.ProviderKey][m.FamilyKey] = true
		if m.Status == "ok" {
			p.OK++
		} else {
			p.Reasons[m.Reason]++
		}
	}
	for key, p := range provs {
		p.Families = len(provFam[key])
		p.Rate = metrics.Ratio(p.OK, p.Models)
		var meds []float64
		for _, m := range oks {
			if m.ProviderKey == key {
				meds = append(meds, m.Med)
			}
		}
		p.Med = metrics.Median(meds)
		for r, n := range p.Reasons {
			if n > p.TopReasonCount || (n == p.TopReasonCount && r < p.TopReason) {
				p.TopReason, p.TopReasonCount = r, n
			}
		}
		p.AccountLevelAll = p.OK == 0 && p.TopReasonCount == p.Models && classify.IsProviderLevel(p.TopReason)
		res.Providers = append(res.Providers, p)
	}
	sort.Slice(res.Providers, func(i, j int) bool {
		a, b := res.Providers[i], res.Providers[j]
		if a.OK != b.OK {
			return a.OK > b.OK
		}
		if a.Models != b.Models {
			return a.Models > b.Models
		}
		return a.Key < b.Key
	})

	if res.Providers == nil {
		res.Providers = []*ProviderStat{}
	}

	// Summary
	s := &res.Summary
	s.Models = len(res.Models)
	s.OK = len(oks)
	s.Fail = len(fails)
	s.Families = len(res.familyByKey)
	s.FamiliesOK = len(avail)
	s.FamiliesUnavail = len(unavail)
	s.Providers = len(res.Providers)
	s.TierCounts = map[string]int{}
	s.FamilyTierCounts = map[string]int{}
	s.Reasons = map[string]int{}
	for _, m := range oks {
		s.TierCounts[m.Tier]++
	}
	for _, f := range avail {
		s.FamilyTierCounts[f.Tier]++
	}
	for _, m := range fails {
		s.Reasons[m.Reason]++
		if classify.IsProviderLevel(m.Reason) {
			s.ProviderLevelFail++
		}
	}
	s.Buckets = []Bucket{{"< 1 s", 0, 1, 0}, {"1 – 2 s", 1, 2, 0}, {"2 – 5 s", 2, 5, 0}, {"5 – 10 s", 5, 10, 0}, {"10 – 30 s", 10, 30, 0}, {"≥ 30 s", 30, 1e9, 0}}
	if len(oks) > 0 {
		meds := make([]float64, 0, len(oks))
		for _, m := range oks {
			meds = append(meds, m.Med)
		}
		st := metrics.Summarise(meds)
		s.LatencyMedian, s.LatencyP90, s.LatencyMin, s.LatencyMax = st.Median, st.P90, st.Min, st.Max
		for _, v := range meds {
			for i := range s.Buckets {
				if v >= s.Buckets[i].Lo && v < s.Buckets[i].Hi {
					s.Buckets[i].Count++
					break
				}
			}
		}
	}
	return res
}

func modelTier(rate float64, sc Scoring) string {
	switch {
	case rate >= sc.ModelTierA:
		return "A"
	case rate >= sc.ModelTierB:
		return "B"
	}
	return "C"
}

func familyTier(rate float64, sc Scoring) string {
	switch {
	case rate >= sc.FamilyTierA:
		return "A"
	case rate >= sc.FamilyTierB:
		return "B"
	}
	return "C"
}

// sortModels orders available models by tier, then median latency, then id.
func sortModels(ms []*ModelStat) {
	sort.Slice(ms, func(i, j int) bool {
		a, b := ms[i], ms[j]
		if tierRank[a.Tier] != tierRank[b.Tier] {
			return tierRank[a.Tier] < tierRank[b.Tier]
		}
		if r1, r2 := metrics.Round(a.Med, 3), metrics.Round(b.Med, 3); r1 != r2 {
			return r1 < r2
		}
		return a.ID < b.ID
	})
}

// TierName returns the display label of a tier.
func TierName(t string, lang string) string {
	zh := map[string]string{"A": "A 稳定", "B": "B 基本稳定", "C": "C 不稳定"}
	en := map[string]string{"A": "A stable", "B": "B mostly stable", "C": "C unstable"}
	if lang == "en" {
		return en[t]
	}
	return zh[t]
}
