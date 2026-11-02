// Package demo seeds a database with a complete demo suite built from the
// embedded reference scan plus deterministic synthetic latency, capability,
// quality and load data, so that the dashboard can be explored (and
// screenshotted) without any endpoint.
package demo

import (
	"encoding/json"
	"math/rand"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/analyze"
	"github.com/AnnaofArendelle/llm-speedtest/internal/capability"
	"github.com/AnnaofArendelle/llm-speedtest/internal/config"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/latency"
	"github.com/AnnaofArendelle/llm-speedtest/internal/load"
	"github.com/AnnaofArendelle/llm-speedtest/internal/quality"
	"github.com/AnnaofArendelle/llm-speedtest/internal/refdata"
	"github.com/AnnaofArendelle/llm-speedtest/internal/store"
)

// Note marks demo runs.
const Note = "demo data (embedded reference scan + synthetic stage data)"

// Seed creates the demo endpoint and suite unless they already exist.
// It returns the suite run id ("" when the demo was already present).
func Seed(st *store.Store) (string, error) {
	runs, _, err := st.ListRuns(core.KindSuite, "", 50, 0)
	if err != nil {
		return "", err
	}
	for _, r := range runs {
		if r.Note == Note {
			return "", nil
		}
	}
	models, attempts, err := refdata.Load()
	if err != nil {
		return "", err
	}
	ep := &store.Endpoint{Name: "demo-gateway", BaseURL: "http://localhost:20128/v1", Dialect: "9router"}
	if err := st.SaveEndpoint(ep); err != nil {
		return "", err
	}
	suite := config.DefaultSuite(config.ProfileStandard)
	spec, _ := json.Marshal(map[string]any{"suite": suite, "kind": core.KindSuite})
	started := attempts[0].At
	finished := attempts[len(attempts)-1].At.Add(2 * time.Second)
	parent := &store.Run{Kind: core.KindSuite, EndpointID: ep.ID, EndpointName: ep.Name, BaseURL: ep.BaseURL, Dialect: "9router", Status: core.StatusRunning, Preset: "standard", Spec: spec, Note: Note, CreatedAt: started}
	if err := st.CreateRun(parent); err != nil {
		return "", err
	}
	child := func(kind string) (*store.Run, error) {
		r := &store.Run{Kind: kind, ParentID: parent.ID, EndpointID: ep.ID, EndpointName: ep.Name, BaseURL: ep.BaseURL, Dialect: "9router", Status: core.StatusRunning, Preset: "standard", Spec: spec, Note: Note, CreatedAt: started}
		if err := st.CreateRun(r); err != nil {
			return nil, err
		}
		_ = st.SetRunStatus(r.ID, core.StatusRunning, "")
		return r, nil
	}
	// ---- scan (real data)
	scanRun, err := child(core.KindScan)
	if err != nil {
		return "", err
	}
	pres := analyze.Enrich(models, attempts, analyze.Options{EndpointName: ep.Name})
	if err := st.PutRunModels(scanRun.ID, models); err != nil {
		return "", err
	}
	if err := st.InsertAttempts(scanRun.ID, attempts); err != nil {
		return "", err
	}
	res := analyze.Build(models, attempts, analyze.DefaultScoring(), &pres)
	rounds := []map[string]any{}
	counts := map[int][3]int{}
	for _, a := range attempts {
		c := counts[a.Round]
		c[0]++
		if a.OK {
			c[1]++
		} else {
			c[2]++
		}
		counts[a.Round] = c
	}
	kinds := []string{"first", "retry", "resample", "resample", "final"}
	for r := 1; r <= 5; r++ {
		c := counts[r]
		rounds = append(rounds, map[string]any{"round": r, "kind": kinds[r-1], "targets": c[0], "ok": c[1], "fail": c[2]})
	}
	if err := st.SetRunSummary(scanRun.ID, map[string]any{"summary": res.Summary, "rounds": rounds, "elapsed_s": finished.Sub(started).Seconds(), "provider_field": pres.Field}); err != nil {
		return "", err
	}
	_ = st.SetRunStatus(scanRun.ID, core.StatusDone, "")

	rnd := rand.New(rand.NewSource(20261101))
	var top []*analyze.ModelStat
	for _, m := range res.Available {
		if len(top) >= 12 {
			break
		}
		top = append(top, m)
	}
	// ---- latency (synthetic from scan medians)
	latRun, err := child(core.KindLatency)
	if err != nil {
		return "", err
	}
	var samples []core.LatencySample
	at := finished
	for _, m := range top {
		base := m.Med
		tps := 40 + rnd.Float64()*80
		for i := 0; i < 6; i++ {
			warm := i > 0
			ttft := base * (0.8 + rnd.Float64()*0.5)
			if !warm {
				ttft *= 1.6
			}
			out := 120 + rnd.Intn(16)
			tpot := 1 / tps
			e2e := ttft + float64(out-1)*tpot
			at = at.Add(time.Duration(e2e*float64(time.Second)) + 50*time.Millisecond)
			samples = append(samples, core.LatencySample{ModelID: m.ID, Index: i, Warm: warm, At: at, OK: true,
				TTFT: time.Duration(ttft * float64(time.Second)), TTST: time.Duration(tpot * 1.3 * float64(time.Second)), TPOT: time.Duration(tpot * float64(time.Second)),
				E2E: time.Duration(e2e * float64(time.Second)), TokPerSec: (float64(out-1) / (e2e - ttft)), InputTokens: 200, OutputTokens: out, TokensExact: rnd.Intn(5) != 0, ServedModel: m.Served})
		}
	}
	if err := st.InsertLatencySamples(latRun.ID, samples); err != nil {
		return "", err
	}
	_ = st.SetRunSummary(latRun.ID, map[string]any{"latency": latency.Summarize(samples)})
	_ = st.SetRunStatus(latRun.ID, core.StatusDone, "")

	// ---- capability (synthetic)
	capRun, err := child(core.KindCapability)
	if err != nil {
		return "", err
	}
	var caps []core.CapabilityResult
	for i, m := range top {
		for j, p := range capability.Probes {
			status := capability.Supported
			ev := "ok"
			switch {
			case p == "embeddings":
				status, ev = capability.Unsupported, "HTTP 404: not an embedding model"
			case p == "vision" && (i+j)%3 == 0:
				status, ev = capability.Unsupported, "HTTP 400: This model does not support image input"
			case p == "json_schema" && (i+j)%4 == 1:
				status, ev = capability.Unsupported, "JSON lacks required keys"
			case p == "long_context":
				ev = "accepted ~32000 tokens"
				if i%5 == 2 {
					ev = "accepted ~8000 tokens; rejected ~32000 tokens: maximum context length exceeded"
				}
			case p == "determinism" && i%4 == 3:
				status, ev = capability.Unsupported, "outputs differ"
			case p == "reasoning_param" && i%6 == 5:
				status, ev = capability.Error, "Upstream 5xx (HTTP 503): A server error occurred"
			}
			caps = append(caps, core.CapabilityResult{ModelID: m.ID, Probe: p, Status: status, Evidence: ev, Latency: time.Duration(300+rnd.Intn(900)) * time.Millisecond})
		}
	}
	if err := st.PutCapabilityResults(capRun.ID, caps); err != nil {
		return "", err
	}
	_ = st.SetRunSummary(capRun.ID, map[string]any{"capability": capability.Summarize(caps, nil)})
	_ = st.SetRunStatus(capRun.ID, core.StatusDone, "")

	// ---- quality (synthetic)
	qRun, err := child(core.KindQuality)
	if err != nil {
		return "", err
	}
	set, _ := quality.Default()
	var qs []core.QualityResult
	for i, m := range top {
		identity := m.Served
		if identity == "" {
			identity = m.Upstream
		}
		qs = append(qs, core.QualityResult{ModelID: m.ID, QuestionID: quality.IdentityQuestionID, Trial: 1, Output: identity})
		skill := 0.55 + 0.45*float64(len(top)-i)/float64(len(top))
		for trial := 1; trial <= 3; trial++ {
			for _, q := range set.Questions {
				pass := rnd.Float64() < skill
				qr := core.QualityResult{ModelID: m.ID, QuestionID: q.ID, Trial: trial, Pass: pass, Latency: time.Duration(200+rnd.Intn(600)) * time.Millisecond}
				if pass {
					qr.Score = q.Weight
					qr.Output = q.Expect
				} else {
					qr.Output = "I'm not sure."
				}
				qs = append(qs, qr)
			}
		}
	}
	if err := st.InsertQualityResults(qRun.ID, qs); err != nil {
		return "", err
	}
	_ = st.SetRunSummary(qRun.ID, map[string]any{"quality": quality.Summarize(qs, set)})
	_ = st.SetRunStatus(qRun.ID, core.StatusDone, "")

	// ---- load (synthetic)
	if len(top) > 0 {
		loadRun, err := child(core.KindLoad)
		if err != nil {
			return "", err
		}
		best := top[0]
		spec := load.Default()
		spec.Model = best.ID
		spec.Levels = []int{1, 2, 4, 8, 16}
		var reqs []core.LoadRequest
		var levels []load.LevelResult
		for _, lvl := range spec.Levels {
			var lr []core.LoadRequest
			t0 := finished.Add(time.Duration(lvl) * time.Minute)
			n := 12 * lvl
			for i := 0; i < n; i++ {
				ttft := best.Med * (1 + float64(lvl)*0.12*rnd.Float64())
				out := 64
				e2e := ttft + float64(out)*(0.012+float64(lvl)*0.0015)
				ok := rnd.Float64() > float64(lvl)*0.01
				r := core.LoadRequest{Level: lvl, Index: i, At: t0.Add(time.Duration(float64(i) / float64(lvl) * e2e * float64(time.Second))), OK: ok,
					TTFT: time.Duration(ttft * float64(time.Second)), E2E: time.Duration(e2e * float64(time.Second)), ITL: time.Duration((e2e - ttft) / float64(out-1) * float64(time.Second)), OutputTokens: out}
				if !ok {
					r.Reason, r.OutputTokens = "rate_limit", 0
				}
				lr = append(lr, r)
			}
			levels = append(levels, load.Summarize(lr, spec, lvl))
			reqs = append(reqs, lr...)
		}
		if err := st.InsertLoadRequests(loadRun.ID, reqs); err != nil {
			return "", err
		}
		_ = st.SetRunSummary(loadRun.ID, map[string]any{"load": levels, "model": best.ID, "mode": spec.Mode})
		_ = st.SetRunStatus(loadRun.ID, core.StatusDone, "")
	}

	_ = st.SetRunSummary(parent.ID, map[string]any{"summary": res.Summary, "children": map[string]string{"scan": scanRun.ID, "latency": latRun.ID, "capability": capRun.ID, "quality": qRun.ID}})
	_ = st.SetRunStatus(parent.ID, core.StatusDone, "")
	if err := seedBroken(st, ep, spec, models, attempts); err != nil {
		return "", err
	}
	_ = st.SetRunTimes(parent.ID, started, finished.Add(25*time.Minute))
	_ = st.SetRunTimes(scanRun.ID, started, finished)
	kids, _ := st.ChildRuns(parent.ID)
	for i, k := range kids {
		if k.Kind != core.KindScan {
			_ = st.SetRunTimes(k.ID, finished.Add(time.Duration(i)*5*time.Minute), finished.Add(time.Duration(i+1)*5*time.Minute))
		}
	}
	return parent.ID, nil
}

// seedBroken adds a failed run (endpoint unreachable) and a cancelled run
// (interrupted after part of the first round) so that the UI's error and
// partial states can be explored and tested.
func seedBroken(st *store.Store, ep *store.Endpoint, spec json.RawMessage, models []core.RunModel, attempts []core.Attempt) error {
	base := attempts[len(attempts)-1].At.Add(time.Hour)
	failed := &store.Run{Kind: core.KindScan, EndpointID: ep.ID, EndpointName: ep.Name, BaseURL: ep.BaseURL, Dialect: "9router", Status: core.StatusPending, Preset: "quick", Spec: spec, Note: Note, CreatedAt: base}
	if err := st.CreateRun(failed); err != nil {
		return err
	}
	_ = st.SetRunTimes(failed.ID, base, base.Add(3*time.Second))
	if err := st.SetRunStatus(failed.ID, core.StatusFailed, "list models: Get \"http://localhost:20128/v1/models\": dial tcp 127.0.0.1:20128: connect: connection refused"); err != nil {
		return err
	}

	cancelled := &store.Run{Kind: core.KindScan, EndpointID: ep.ID, EndpointName: ep.Name, BaseURL: ep.BaseURL, Dialect: "9router", Status: core.StatusPending, Preset: "standard", Spec: spec, Note: Note, CreatedAt: base.Add(10 * time.Minute)}
	if err := st.CreateRun(cancelled); err != nil {
		return err
	}
	subset := models
	if len(subset) > 80 {
		subset = subset[:80]
	}
	if err := st.PutRunModels(cancelled.ID, subset); err != nil {
		return err
	}
	ids := map[string]bool{}
	for i, m := range subset {
		if i < 40 {
			ids[m.ID] = true
		}
	}
	var partial []core.Attempt
	for _, a := range attempts {
		if a.Round == 1 && ids[a.ModelID] {
			partial = append(partial, a)
		}
	}
	if len(partial) > 0 {
		if err := st.InsertAttempts(cancelled.ID, partial); err != nil {
			return err
		}
	}
	_ = st.SetRunTimes(cancelled.ID, base.Add(10*time.Minute), base.Add(12*time.Minute))
	return st.SetRunStatus(cancelled.ID, core.StatusCancelled, "")
}
