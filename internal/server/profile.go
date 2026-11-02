package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"math"
	"net/http"
	"sort"

	"github.com/AnnaofArendelle/llm-speedtest/internal/analyze"
	"github.com/AnnaofArendelle/llm-speedtest/internal/capability"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/latency"
	"github.com/AnnaofArendelle/llm-speedtest/internal/load"
	"github.com/AnnaofArendelle/llm-speedtest/internal/metrics"
	"github.com/AnnaofArendelle/llm-speedtest/internal/quality"
)

// profileDimension is one axis of the capability radar.
type profileDimension struct {
	Key     string  `json:"key"`
	LabelZH string  `json:"label_zh"`
	LabelEN string  `json:"label_en"`
	Score   float64 `json:"score"` // 0..100
	Covered bool    `json:"covered"`
	Detail  string  `json:"detail"`
}

// profileDoc is the response of the model/family profile endpoints.
type profileDoc struct {
	ID         string             `json:"id"`
	Kind       string             `json:"kind"` // model | family
	Display    string             `json:"display"`
	Dimensions []profileDimension `json:"dimensions"`
	Coverage   map[string]bool    `json:"coverage"` // stage -> any data
	Measured   map[string]int     `json:"measured"` // stage -> members measured (families)
	Members    int                `json:"members"`
}

var profileDims = []struct{ key, zh, en string }{
	{"stability", "稳定性", "Stability"},
	{"speed", "响应速度", "Speed"},
	{"breadth", "提供商广度", "Breadth"},
	{"ttft", "首 token", "TTFT"},
	{"throughput", "解码吞吐", "Throughput"},
	{"capability", "能力覆盖", "Capabilities"},
	{"quality", "回答质量", "Quality"},
	{"load", "负载 Goodput", "Load"},
}

// stageData holds the per-model results of the stages attached to a run.
type stageData struct {
	latency     map[string]latency.ModelStats
	tpsRef      float64 // p90 of per-model median tok/s in the latency stage
	caps        *capability.Matrix
	quality     map[string]quality.ModelScore
	loadModel   string
	loadGoodput float64
	hasLoad     bool
}

func (s *Server) stageData(runID string) stageData {
	sd := stageData{latency: map[string]latency.ModelStats{}, quality: map[string]quality.ModelScore{}}
	if run, err := s.childOf(runID, core.KindLatency); err == nil {
		if samples, err := s.svc.Store.ListLatencySamples(run.ID); err == nil {
			var tps []float64
			for _, st := range latency.Summarize(samples) {
				sd.latency[st.ModelID] = st
				if st.OK > 0 {
					tps = append(tps, st.TokPerSec.Median)
				}
			}
			if len(tps) > 0 {
				sd.tpsRef = metrics.Summarise(tps).P90
			}
		}
	}
	if run, err := s.childOf(runID, core.KindCapability); err == nil {
		if results, err := s.svc.Store.ListCapabilityResults(run.ID); err == nil && len(results) > 0 {
			m := capability.Summarize(results, nil)
			sd.caps = &m
		}
	}
	if run, err := s.childOf(runID, core.KindQuality); err == nil {
		if results, err := s.svc.Store.ListQualityResults(run.ID); err == nil {
			for _, sc := range quality.Summarize(results, nil) {
				sd.quality[sc.ModelID] = sc
			}
		}
	}
	if run, err := s.childOf(runID, core.KindLoad); err == nil {
		var sum struct {
			Load  []load.LevelResult `json:"load"`
			Model string             `json:"model"`
		}
		if json.Unmarshal(run.Summary, &sum) == nil && len(sum.Load) > 0 {
			sd.hasLoad, sd.loadModel = true, sum.Model
			best := sum.Load[0]
			for _, l := range sum.Load {
				if l.Level > best.Level {
					best = l
				}
			}
			sd.loadGoodput = best.Goodput
		}
	}
	return sd
}

func clamp(v float64) float64 {
	if math.IsNaN(v) || v < 0 {
		return 0
	}
	if v > 100 {
		return 100
	}
	return metrics.Round(v, 1)
}

// modelDims computes the eight radar axes of one model.
func modelDims(m *analyze.ModelStat, fam *analyze.FamilyStat, sc analyze.Scoring, sd stageData) ([]profileDimension, map[string]bool) {
	cov := map[string]bool{"scan": m.Attempts > 0, "latency": false, "capability": false, "quality": false, "load": false}
	out := make([]profileDimension, 0, len(profileDims))
	for _, d := range profileDims {
		pd := profileDimension{Key: d.key, LabelZH: d.zh, LabelEN: d.en}
		switch d.key {
		case "stability":
			if m.Attempts > 0 {
				pd.Covered, pd.Score = true, clamp(m.Rate*100)
				pd.Detail = fmt.Sprintf("%d/%d", m.OK, m.Attempts)
			}
		case "speed":
			if m.Status == "ok" {
				pd.Covered, pd.Score = true, clamp(metrics.LogScore(m.Med, sc.SpeedLo, sc.SpeedHi))
				pd.Detail = fmt.Sprintf("median %.2fs", m.Med)
			}
		case "breadth":
			if fam != nil && fam.OK > 0 {
				pd.Covered, pd.Score = true, clamp(fam.SBreadth)
				pd.Detail = fmt.Sprintf("%d providers", fam.NProvOK)
			}
		case "ttft":
			if st, ok := sd.latency[m.ID]; ok && st.OK > 0 {
				cov["latency"] = true
				pd.Covered, pd.Score = true, clamp(metrics.LogScore(st.TTFT.Median, 0.2, 5))
				pd.Detail = fmt.Sprintf("p50 %.2fs", st.TTFT.Median)
			}
		case "throughput":
			if st, ok := sd.latency[m.ID]; ok && st.OK > 0 && sd.tpsRef > 0 {
				cov["latency"] = true
				pd.Covered, pd.Score = true, clamp(100*st.TokPerSec.Median/sd.tpsRef)
				pd.Detail = fmt.Sprintf("%.1f tok/s", st.TokPerSec.Median)
			}
		case "capability":
			if sd.caps != nil {
				if cells, ok := sd.caps.Cells[m.ID]; ok && len(sd.caps.Probes) > 0 {
					cov["capability"] = true
					pd.Covered, pd.Score = true, clamp(100*float64(sd.caps.Support[m.ID])/float64(len(sd.caps.Probes)))
					pd.Detail = fmt.Sprintf("%d/%d probes", sd.caps.Support[m.ID], len(cells))
				}
			}
		case "quality":
			if q, ok := sd.quality[m.ID]; ok && q.Answered > 0 {
				cov["quality"] = true
				pd.Covered, pd.Score = true, clamp(q.Score)
				pd.Detail = fmt.Sprintf("%.0f/100, consistency %.0f%%", q.Score, q.Consistency*100)
			}
		case "load":
			if sd.hasLoad && sd.loadModel == m.ID {
				cov["load"] = true
				pd.Covered, pd.Score = true, clamp(sd.loadGoodput*100)
				pd.Detail = fmt.Sprintf("goodput %.0f%%", sd.loadGoodput*100)
			}
		}
		out = append(out, pd)
	}
	return out, cov
}

func (s *Server) modelProfile(w http.ResponseWriter, r *http.Request) {
	res, run, err := s.scanResult(r.PathValue("id"))
	if err != nil {
		notFoundOr(w, err)
		return
	}
	id := r.PathValue("model")
	m := res.Model(id)
	if m == nil {
		writeErr(w, 404, errors.New("model not found"))
		return
	}
	parent := r.PathValue("id")
	_ = run
	dims, cov := modelDims(m, res.Family(m.FamilyKey), res.Scoring, s.stageData(parent))
	measured := map[string]int{}
	for stage, ok := range cov {
		if ok {
			measured[stage] = 1
		}
	}
	writeJSON(w, 200, profileDoc{ID: id, Kind: "model", Display: id, Dimensions: dims, Coverage: cov, Measured: measured, Members: 1})
}

func (s *Server) familyProfile(w http.ResponseWriter, r *http.Request) {
	res, _, err := s.scanResult(r.PathValue("id"))
	if err != nil {
		notFoundOr(w, err)
		return
	}
	f := res.Family(r.PathValue("key"))
	if f == nil {
		writeErr(w, 404, errors.New("family not found"))
		return
	}
	sd := s.stageData(r.PathValue("id"))
	best := make([]profileDimension, len(profileDims))
	bestModel := make([]string, len(profileDims))
	for i, d := range profileDims {
		best[i] = profileDimension{Key: d.key, LabelZH: d.zh, LabelEN: d.en}
	}
	cov := map[string]bool{"scan": false, "latency": false, "capability": false, "quality": false, "load": false}
	measured := map[string]int{}
	members := f.Members
	sort.Slice(members, func(i, j int) bool { return members[i].ID < members[j].ID })
	for _, m := range members {
		dims, c := modelDims(m, f, res.Scoring, sd)
		for stage, ok := range c {
			if ok {
				cov[stage] = true
				measured[stage]++
			}
		}
		for i, d := range dims {
			if d.Covered && (!best[i].Covered || d.Score > best[i].Score) {
				best[i].Covered, best[i].Score, best[i].Detail = true, d.Score, d.Detail
				bestModel[i] = m.ID
			}
		}
	}
	for i := range best {
		if best[i].Covered && len(members) > 1 {
			best[i].Detail = fmt.Sprintf("%s · best: %s", best[i].Detail, bestModel[i])
		}
	}
	// Stability and breadth are family-level figures rather than the best member.
	for i := range best {
		switch best[i].Key {
		case "stability":
			if f.Attempts > 0 {
				best[i].Covered, best[i].Score, best[i].Detail = true, clamp(f.Rate*100), fmt.Sprintf("%d/%d attempts · %d/%d members available", f.Successes, f.Attempts, f.OK, f.N)
			}
		case "breadth":
			if f.OK > 0 {
				best[i].Covered, best[i].Score, best[i].Detail = true, clamp(f.SBreadth), fmt.Sprintf("%d providers available", f.NProvOK)
			}
		}
	}
	writeJSON(w, 200, profileDoc{ID: f.Key, Kind: "family", Display: f.Display, Dimensions: best, Coverage: cov, Measured: measured, Members: len(members)})
}
