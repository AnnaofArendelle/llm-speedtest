// Package app orchestrates runs end to end: endpoint checks, discovery, the
// benchmark engines, persistence, progress events and report generation.
// It is shared by the HTTP server, the CLI and the MCP server.
package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/analyze"
	"github.com/AnnaofArendelle/llm-speedtest/internal/apiclient"
	"github.com/AnnaofArendelle/llm-speedtest/internal/capability"
	"github.com/AnnaofArendelle/llm-speedtest/internal/config"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/dialect"
	"github.com/AnnaofArendelle/llm-speedtest/internal/family"
	"github.com/AnnaofArendelle/llm-speedtest/internal/latency"
	"github.com/AnnaofArendelle/llm-speedtest/internal/load"
	"github.com/AnnaofArendelle/llm-speedtest/internal/modelid"
	"github.com/AnnaofArendelle/llm-speedtest/internal/provider"
	"github.com/AnnaofArendelle/llm-speedtest/internal/quality"
	"github.com/AnnaofArendelle/llm-speedtest/internal/report"
	"github.com/AnnaofArendelle/llm-speedtest/internal/scan"
	"github.com/AnnaofArendelle/llm-speedtest/internal/store"
)

// Service runs benchmarks.
type Service struct {
	Store   *store.Store
	Bus     *Bus
	Version string

	mu       sync.Mutex
	running  map[string]context.CancelFunc
	skippers map[string]*core.Skipper
}

// New creates a service.
func New(st *store.Store, version string) *Service {
	return &Service{Store: st, Bus: NewBus(), Version: version, running: map[string]context.CancelFunc{}, skippers: map[string]*core.Skipper{}}
}

// Request describes a run to execute.
type Request struct {
	Kind       string           `json:"kind"`
	EndpointID string           `json:"endpoint_id,omitempty"` // stored endpoint (key loaded from the store)
	Endpoint   *config.Endpoint `json:"endpoint,omitempty"`    // inline endpoint (CLI / ad-hoc)
	Suite      config.Suite     `json:"suite"`
	Models     []string         `json:"models,omitempty"` // explicit models for latency/capability/quality/load
	Note       string           `json:"note,omitempty"`
	Confirm    bool             `json:"confirm,omitempty"` // required for load tests started through automation
}

// ErrConfirmRequired is returned when a load test is started without confirmation.
var ErrConfirmRequired = errors.New("load tests send sustained traffic to the endpoint; set confirm=true to proceed")

// resolveEndpoint loads the endpoint from the store or the request.
func (s *Service) resolveEndpoint(req Request) (config.Endpoint, string, error) {
	if req.Endpoint != nil {
		ep := *req.Endpoint
		if ep.Name == "" {
			ep.Name = ep.BaseURL
		}
		return ep, req.EndpointID, nil
	}
	if req.EndpointID == "" {
		return config.Endpoint{}, "", errors.New("endpoint is required")
	}
	st, err := s.Store.GetEndpointWithKey(req.EndpointID)
	if err != nil {
		return config.Endpoint{}, "", err
	}
	return config.Endpoint{Name: st.Name, BaseURL: st.BaseURL, APIKey: st.APIKey, Dialect: st.Dialect, Headers: st.Headers, Insecure: st.Insecure, Proxy: st.Proxy}, st.ID, nil
}

func clientFor(ep config.Endpoint) (*apiclient.Client, error) {
	return apiclient.New(ep.BaseURL, ep.APIKey, apiclient.Options{Headers: ep.Headers, Insecure: ep.Insecure, Proxy: ep.Proxy, UserAgent: "llm-speedtest"})
}

// CheckResult is the outcome of an endpoint check.
type CheckResult struct {
	BaseURL     string       `json:"base_url"`
	Reachable   bool         `json:"reachable"`
	AuthOK      bool         `json:"auth_ok"`
	Models      int          `json:"models"`
	Owners      int          `json:"owners"`
	Dialect     dialect.Info `json:"dialect"`
	Latency     float64      `json:"latency_s"`
	Error       string       `json:"error,omitempty"`
	SuggestedV1 string       `json:"suggested_base_url,omitempty"`
}

// Check verifies connectivity, authentication and discovery for an endpoint.
func (s *Service) Check(ctx context.Context, ep config.Endpoint) CheckResult {
	res := CheckResult{BaseURL: apiclient.Normalize(ep.BaseURL)}
	cl, err := clientFor(ep)
	if err != nil {
		res.Error = err.Error()
		return res
	}
	start := time.Now()
	cctx, cancel := context.WithTimeout(ctx, 180*time.Second)
	defer cancel()
	models, err := cl.ListModels(cctx)
	res.Latency = time.Since(start).Seconds()
	var es *apiclient.ErrStatus
	if err != nil && errors.As(err, &es) && es.Status == 404 && !strings.HasSuffix(res.BaseURL, "/v1") {
		if alt, err2 := cl.WithBase(res.BaseURL + "/v1"); err2 == nil {
			if models2, err3 := alt.ListModels(cctx); err3 == nil {
				res.SuggestedV1 = res.BaseURL + "/v1"
				models, err = models2, nil
				cl = alt
			}
		}
	}
	if err != nil {
		if errors.As(err, &es) {
			res.Reachable = true
			res.AuthOK = es.Status != 401 && es.Status != 403
			res.Error = err.Error()
		} else {
			res.Error = err.Error()
		}
	} else {
		res.Reachable, res.AuthOK = true, true
		res.Models = len(models)
		owners := map[string]bool{}
		for _, m := range models {
			owners[m.OwnedBy] = true
		}
		res.Owners = len(owners)
	}
	if ep.Dialect != "" && ep.Dialect != dialect.Generic {
		res.Dialect = dialect.Health(ctx, ep.Dialect, cl)
	} else {
		res.Dialect = dialect.Detect(ctx, cl)
	}
	return res
}

// DiscoverResult is a model preview.
type DiscoverResult struct {
	Models    []core.RunModel              `json:"models"`
	Total     int                          `json:"total"`
	Filtered  int                          `json:"filtered"`
	Providers []provider.Provider          `json:"providers"`
	Field     string                       `json:"provider_field"`
	Rationale string                       `json:"provider_rationale"`
	Families  int                          `json:"families"`
	Dialect   dialect.Info                 `json:"dialect"`
	Overrides map[string]provider.Override `json:"-"`
}

// Discover lists and filters the models of an endpoint and resolves
// providers/families from the listing alone.
func (s *Service) Discover(ctx context.Context, ep config.Endpoint, endpointID string, filter config.Filter) (*DiscoverResult, error) {
	cl, err := clientFor(ep)
	if err != nil {
		return nil, err
	}
	dctx, cancel := context.WithTimeout(ctx, 180*time.Second)
	defer cancel()
	models, err := cl.ListModels(dctx)
	if err != nil {
		// one retry for slow gateways
		models, err = cl.ListModels(dctx)
		if err != nil {
			return nil, fmt.Errorf("list models: %w", err)
		}
	}
	res := &DiscoverResult{Total: len(models)}
	if ep.Dialect != "" && ep.Dialect != dialect.Generic {
		res.Dialect = dialect.Info{Name: ep.Dialect}
	} else {
		res.Dialect = dialect.Detect(ctx, cl)
	}
	var rms []core.RunModel
	for _, m := range models {
		ok, err := filter.Match(m.ID, m.OwnedBy)
		if err != nil {
			return nil, err
		}
		if !ok {
			continue
		}
		rms = append(rms, core.RunModel{ID: m.ID, OwnedBy: m.OwnedBy, Upstream: modelid.Upstream(m.ID, m.OwnedBy), ContextLength: m.ContextLength, MaxOutput: m.MaxOutput, Raw: m.Raw})
	}
	sort.Slice(rms, func(i, j int) bool { return rms[i].ID < rms[j].ID })
	if filter.Limit > 0 && len(rms) > filter.Limit {
		rms = rms[:filter.Limit]
	}
	res.Filtered = len(rms)
	opts := s.assignmentOptions(ep.Name, endpointID)
	pres := analyze.Enrich(rms, nil, opts)
	res.Models = rms
	res.Providers = pres.Providers
	res.Field, res.Rationale = pres.Field, pres.Rationale
	fams := map[string]bool{}
	for _, m := range rms {
		fams[m.FamilyKey] = true
	}
	res.Families = len(fams)
	return res, nil
}

func (s *Service) assignmentOptions(name, endpointID string) analyze.Options {
	opts := analyze.Options{EndpointName: name}
	if endpointID != "" && s.Store != nil {
		if po, _, err := s.Store.ProviderOverrides(endpointID); err == nil {
			opts.ProviderOverrides = po
		}
		if fo, _, err := s.Store.FamilyOverrides(endpointID); err == nil {
			opts.FamilyOverrides = fo
		}
	}
	return opts
}

// Start creates the run and executes it in the background.
func (s *Service) Start(req Request) (string, error) {
	req.Suite.Normalize()
	if req.Kind == core.KindLoad && !req.Confirm {
		return "", ErrConfirmRequired
	}
	ep, epID, err := s.resolveEndpoint(req)
	if err != nil {
		return "", err
	}
	run, err := s.createRun(req, ep, epID, "")
	if err != nil {
		return "", err
	}
	ctx, cancel := context.WithCancel(context.Background())
	s.mu.Lock()
	s.running[run.ID] = cancel
	s.skippers[run.ID] = core.NewSkipper()
	s.mu.Unlock()
	go func() {
		defer func() {
			s.mu.Lock()
			delete(s.running, run.ID)
			delete(s.skippers, run.ID)
			s.mu.Unlock()
			cancel()
		}()
		_ = s.execute(ctx, run, req, ep, epID)
	}()
	return run.ID, nil
}

// RunSync executes a run in the foreground (CLI).
func (s *Service) RunSync(ctx context.Context, req Request) (string, error) {
	return s.RunSyncWithID(ctx, req, nil)
}

// RunSyncWithID is RunSync with a callback that receives the run id before
// execution starts (so callers can subscribe to its events).
func (s *Service) RunSyncWithID(ctx context.Context, req Request, onStart func(runID string)) (string, error) {
	req.Suite.Normalize()
	ep, epID, err := s.resolveEndpoint(req)
	if err != nil {
		return "", err
	}
	run, err := s.createRun(req, ep, epID, "")
	if err != nil {
		return "", err
	}
	if onStart != nil {
		onStart(run.ID)
	}
	ctx, cancel := context.WithCancel(ctx)
	s.mu.Lock()
	s.running[run.ID] = cancel
	s.skippers[run.ID] = core.NewSkipper()
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		delete(s.running, run.ID)
		delete(s.skippers, run.ID)
		s.mu.Unlock()
		cancel()
	}()
	return run.ID, s.execute(ctx, run, req, ep, epID)
}

// Cancel stops a running run.
func (s *Service) Cancel(runID string) bool {
	s.mu.Lock()
	cancel, ok := s.running[runID]
	s.mu.Unlock()
	if ok {
		cancel()
	}
	return ok
}

// skipperFor returns the skipper of a run, or of its parent suite for child runs.
func (s *Service) skipperFor(run *store.Run) *core.Skipper {
	s.mu.Lock()
	defer s.mu.Unlock()
	if sk := s.skippers[run.ID]; sk != nil {
		return sk
	}
	if run.ParentID != "" {
		return s.skippers[run.ParentID]
	}
	return nil
}

// Skip skips one model of a running run (or every model currently in flight
// when all is true). It reports false when the run is not in progress.
func (s *Service) Skip(runID, model string, all bool) ([]string, bool) {
	s.mu.Lock()
	sk := s.skippers[runID]
	s.mu.Unlock()
	if sk == nil {
		return nil, false
	}
	var ids []string
	switch {
	case all:
		ids = sk.SkipAll()
	case model != "":
		sk.Skip(model)
		ids = []string{model}
	default:
		ids = []string{}
	}
	if len(ids) > 0 {
		s.Bus.Publish(runID, core.Event{Kind: core.EventLog, Message: fmt.Sprintf("skipped %s", strings.Join(ids, ", "))})
	}
	return ids, true
}

// Inflight lists the requests currently in progress for a run.
func (s *Service) Inflight(runID string) []core.InflightItem {
	s.mu.Lock()
	sk := s.skippers[runID]
	s.mu.Unlock()
	if sk == nil {
		return []core.InflightItem{}
	}
	return sk.Inflight()
}

// IsRunning reports whether a run is in progress.
func (s *Service) IsRunning(runID string) bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	_, ok := s.running[runID]
	return ok
}

func (s *Service) createRun(req Request, ep config.Endpoint, epID, parent string) (*store.Run, error) {
	spec, _ := json.Marshal(map[string]any{"suite": req.Suite, "models": req.Models, "kind": req.Kind})
	run := &store.Run{Kind: req.Kind, ParentID: parent, EndpointID: epID, EndpointName: ep.Name, BaseURL: apiclient.Normalize(ep.BaseURL), Dialect: ep.Dialect,
		Status: core.StatusPending, Preset: req.Suite.Profile, Spec: spec, Note: req.Note}
	if err := s.Store.CreateRun(run); err != nil {
		return nil, err
	}
	return run, nil
}

func (s *Service) fail(run *store.Run, err error) error {
	status := core.StatusFailed
	if errors.Is(err, context.Canceled) {
		status = core.StatusCancelled
	}
	_ = s.Store.SetRunStatus(run.ID, status, err.Error())
	s.Bus.Publish(run.ID, core.Event{Kind: core.EventError, Message: err.Error()})
	return err
}

func (s *Service) reporter(runID string) func(core.Event) {
	var mu sync.Mutex
	last := time.Time{}
	return func(ev core.Event) {
		s.Bus.Publish(runID, ev)
		if ev.Kind == core.EventAttempt || ev.Kind == core.EventProgress || ev.Kind == core.EventSample {
			mu.Lock()
			due := time.Since(last) > 2*time.Second
			if due {
				last = time.Now()
			}
			mu.Unlock()
			if due {
				_ = s.Store.SetRunProgress(runID, progressOf(ev))
			}
		}
	}
}

func progressOf(ev core.Event) map[string]any {
	return map[string]any{"stage": ev.Stage, "round": ev.Round, "rounds": ev.Rounds, "done": ev.Done, "total": ev.Total, "ok": ev.OK, "fail": ev.Fail, "eta_s": ev.ETA, "at": ev.At}
}

func (s *Service) execute(ctx context.Context, run *store.Run, req Request, ep config.Endpoint, epID string) error {
	switch req.Kind {
	case core.KindScan:
		_, err := s.runScan(ctx, run, req, ep, epID)
		return err
	case core.KindLatency:
		return s.runLatency(ctx, run, req, ep, epID, req.Models)
	case core.KindLoad:
		return s.runLoad(ctx, run, req, ep, epID)
	case core.KindCapability:
		return s.runCapability(ctx, run, req, ep, epID, req.Models)
	case core.KindQuality:
		return s.runQuality(ctx, run, req, ep, epID, req.Models)
	case core.KindSuite:
		return s.runSuite(ctx, run, req, ep, epID)
	}
	return s.fail(run, fmt.Errorf("unknown run kind %q", req.Kind))
}

// ScanOutcome is what a scan produces for the suite.
type ScanOutcome struct {
	Models   []core.RunModel
	Attempts []core.Attempt
	Result   *analyze.Result
	Dialect  dialect.Info
}

func (s *Service) runScan(ctx context.Context, run *store.Run, req Request, ep config.Endpoint, epID string) (*ScanOutcome, error) {
	if err := s.Store.SetRunStatus(run.ID, core.StatusDiscovering, ""); err != nil {
		return nil, err
	}
	s.Bus.Publish(run.ID, core.Event{Kind: core.EventLog, Stage: "scan", Message: "discovering models"})
	disc, err := s.Discover(ctx, ep, epID, req.Suite.Filter)
	if err != nil {
		return nil, s.fail(run, err)
	}
	if disc.Dialect.Name != "" {
		_ = s.Store.SetRunDialect(run.ID, disc.Dialect.Name)
	}
	if len(disc.Models) == 0 {
		return nil, s.fail(run, errors.New("no models matched the filter"))
	}
	if err := s.Store.PutRunModels(run.ID, disc.Models); err != nil {
		return nil, s.fail(run, err)
	}
	if err := s.Store.SetRunStatus(run.ID, core.StatusRunning, ""); err != nil {
		return nil, err
	}
	cl, err := clientFor(ep)
	if err != nil {
		return nil, s.fail(run, err)
	}
	targets := make([]scan.Target, 0, len(disc.Models))
	for _, m := range disc.Models {
		targets = append(targets, scan.Target{ID: m.ID, Provider: m.ProviderKey, Name: modelid.Name(m.ID, m.OwnedBy)})
	}
	prior, _ := s.Store.ListAttempts(run.ID)
	writer := s.Store.AttemptWriter(run.ID)
	runner := &scan.Runner{Skipper: s.skipperFor(run), Client: cl, Spec: req.Suite.Specs.Scan, Sink: scan.SinkFunc(func(_ context.Context, a core.Attempt) error { return writer.Add(a) }), Report: s.reporter(run.ID), Prior: prior}
	sum, runErr := runner.Run(ctx, targets)
	if ferr := writer.Flush(); ferr != nil && runErr == nil {
		runErr = ferr
	}
	if runErr != nil {
		return nil, s.fail(run, runErr)
	}
	attempts, err := s.Store.ListAttempts(run.ID)
	if err != nil {
		return nil, s.fail(run, err)
	}
	models := disc.Models
	pres := analyze.Enrich(models, attempts, s.assignmentOptions(ep.Name, epID))
	if err := s.Store.UpdateRunModelAssignments(run.ID, models); err != nil {
		return nil, s.fail(run, err)
	}
	res := analyze.Build(models, attempts, req.Suite.Specs.Scoring, &pres)
	summary := map[string]any{"summary": res.Summary, "rounds": sum.Rounds, "elapsed_s": sum.Elapsed.Seconds(), "dialect": disc.Dialect, "provider_field": pres.Field}
	if err := s.Store.SetRunSummary(run.ID, summary); err != nil {
		return nil, s.fail(run, err)
	}
	if err := s.Store.SetRunStatus(run.ID, core.StatusDone, ""); err != nil {
		return nil, err
	}
	s.Bus.Publish(run.ID, core.Event{Kind: core.EventDone, Stage: "scan", Done: sum.Attempts, Total: sum.Attempts, OK: res.Summary.OK, Fail: res.Summary.Fail, Message: "done", Data: res.Summary})
	return &ScanOutcome{Models: models, Attempts: attempts, Result: res, Dialect: disc.Dialect}, nil
}

func (s *Service) targetsFor(ctx context.Context, ep config.Endpoint, epID string, ids []string) ([]core.RunModel, error) {
	if len(ids) == 0 {
		return nil, errors.New("no models specified")
	}
	models := make([]core.RunModel, 0, len(ids))
	for _, id := range ids {
		models = append(models, core.RunModel{ID: id, Upstream: id})
	}
	// Try to enrich owned_by from discovery (best effort, short timeout).
	cl, err := clientFor(ep)
	if err == nil {
		dctx, cancel := context.WithTimeout(ctx, 60*time.Second)
		defer cancel()
		if list, err := cl.ListModels(dctx); err == nil {
			by := map[string]apiclient.Model{}
			for _, m := range list {
				by[m.ID] = m
			}
			for i := range models {
				if m, ok := by[models[i].ID]; ok {
					models[i].OwnedBy = m.OwnedBy
					models[i].Upstream = modelid.Upstream(m.ID, m.OwnedBy)
					models[i].ContextLength = m.ContextLength
				}
			}
		}
	}
	analyze.Enrich(models, nil, s.assignmentOptions(ep.Name, epID))
	return models, nil
}

func (s *Service) runLatency(ctx context.Context, run *store.Run, req Request, ep config.Endpoint, epID string, ids []string) error {
	models, err := s.targetsFor(ctx, ep, epID, ids)
	if err != nil {
		return s.fail(run, err)
	}
	_ = s.Store.PutRunModels(run.ID, models)
	_ = s.Store.SetRunStatus(run.ID, core.StatusRunning, "")
	cl, err := clientFor(ep)
	if err != nil {
		return s.fail(run, err)
	}
	targets := make([]latency.Target, 0, len(models))
	for _, m := range models {
		targets = append(targets, latency.Target{ID: m.ID, Provider: m.ProviderKey, Name: modelid.Name(m.ID, m.OwnedBy)})
	}
	writer := s.Store.LatencyWriter(run.ID)
	runner := &latency.Runner{Skipper: s.skipperFor(run), Client: cl, Spec: req.Suite.Specs.Latency, Sink: latency.SinkFunc(func(_ context.Context, smp core.LatencySample) error { return writer.Add(smp) }), Report: s.reporter(run.ID)}
	samples, runErr := runner.Run(ctx, targets)
	if ferr := writer.Flush(); ferr != nil && runErr == nil {
		runErr = ferr
	}
	if runErr != nil {
		return s.fail(run, runErr)
	}
	stats := latency.Summarize(samples)
	if err := s.Store.SetRunSummary(run.ID, map[string]any{"latency": stats}); err != nil {
		return s.fail(run, err)
	}
	_ = s.Store.SetRunStatus(run.ID, core.StatusDone, "")
	s.Bus.Publish(run.ID, core.Event{Kind: core.EventDone, Stage: "latency", Message: "done", Data: stats})
	return nil
}

func (s *Service) runLoad(ctx context.Context, run *store.Run, req Request, ep config.Endpoint, epID string) error {
	spec := req.Suite.Specs.Load
	if spec.Model == "" && len(req.Models) > 0 {
		spec.Model = req.Models[0]
	}
	if spec.Model == "" {
		return s.fail(run, errors.New("load test needs a model"))
	}
	models, _ := s.targetsFor(ctx, ep, epID, []string{spec.Model})
	_ = s.Store.PutRunModels(run.ID, models)
	_ = s.Store.SetRunStatus(run.ID, core.StatusRunning, "")
	cl, err := clientFor(ep)
	if err != nil {
		return s.fail(run, err)
	}
	writer := s.Store.LoadWriter(run.ID)
	name := spec.Model
	if len(models) > 0 {
		name = modelid.Name(models[0].ID, models[0].OwnedBy)
	}
	runner := &load.Runner{Client: cl, Spec: spec, Name: name, Sink: load.SinkFunc(func(_ context.Context, r core.LoadRequest) error { return writer.Add(r) }), Report: s.reporter(run.ID)}
	levels, _, runErr := runner.Run(ctx)
	if ferr := writer.Flush(); ferr != nil && runErr == nil {
		runErr = ferr
	}
	if runErr != nil {
		return s.fail(run, runErr)
	}
	if err := s.Store.SetRunSummary(run.ID, map[string]any{"load": levels, "model": spec.Model, "mode": spec.Mode}); err != nil {
		return s.fail(run, err)
	}
	_ = s.Store.SetRunStatus(run.ID, core.StatusDone, "")
	s.Bus.Publish(run.ID, core.Event{Kind: core.EventDone, Stage: "load", Message: "done", Data: levels})
	return nil
}

func (s *Service) runCapability(ctx context.Context, run *store.Run, req Request, ep config.Endpoint, epID string, ids []string) error {
	models, err := s.targetsFor(ctx, ep, epID, ids)
	if err != nil {
		return s.fail(run, err)
	}
	_ = s.Store.PutRunModels(run.ID, models)
	_ = s.Store.SetRunStatus(run.ID, core.StatusRunning, "")
	cl, err := clientFor(ep)
	if err != nil {
		return s.fail(run, err)
	}
	targets := make([]capability.Target, 0, len(models))
	for _, m := range models {
		targets = append(targets, capability.Target{ID: m.ID, Provider: m.ProviderKey, Name: modelid.Name(m.ID, m.OwnedBy)})
	}
	var mu sync.Mutex
	var buf []core.CapabilityResult
	runner := &capability.Runner{Skipper: s.skipperFor(run), Client: cl, Spec: req.Suite.Specs.Capability, Sink: capability.SinkFunc(func(_ context.Context, r core.CapabilityResult) error {
		mu.Lock()
		buf = append(buf, r)
		n := len(buf)
		mu.Unlock()
		if n%20 == 0 {
			mu.Lock()
			rows := buf
			buf = nil
			mu.Unlock()
			return s.Store.PutCapabilityResults(run.ID, rows)
		}
		return nil
	}), Report: s.reporter(run.ID)}
	results, runErr := runner.Run(ctx, targets)
	mu.Lock()
	rest := buf
	buf = nil
	mu.Unlock()
	if err := s.Store.PutCapabilityResults(run.ID, rest); err != nil && runErr == nil {
		runErr = err
	}
	if runErr != nil {
		return s.fail(run, runErr)
	}
	mx := capability.Summarize(results, req.Suite.Specs.Capability.Probes)
	if err := s.Store.SetRunSummary(run.ID, map[string]any{"capability": mx}); err != nil {
		return s.fail(run, err)
	}
	_ = s.Store.SetRunStatus(run.ID, core.StatusDone, "")
	s.Bus.Publish(run.ID, core.Event{Kind: core.EventDone, Stage: "capability", Message: "done"})
	return nil
}

func (s *Service) runQuality(ctx context.Context, run *store.Run, req Request, ep config.Endpoint, epID string, ids []string) error {
	models, err := s.targetsFor(ctx, ep, epID, ids)
	if err != nil {
		return s.fail(run, err)
	}
	_ = s.Store.PutRunModels(run.ID, models)
	_ = s.Store.SetRunStatus(run.ID, core.StatusRunning, "")
	cl, err := clientFor(ep)
	if err != nil {
		return s.fail(run, err)
	}
	var set *quality.Set
	if req.Suite.Specs.Quality.Questions != "" {
		set, err = quality.Load(req.Suite.Specs.Quality.Questions)
		if err != nil {
			return s.fail(run, err)
		}
	}
	targets := make([]quality.Target, 0, len(models))
	for _, m := range models {
		targets = append(targets, quality.Target{ID: m.ID, Provider: m.ProviderKey, Name: modelid.Name(m.ID, m.OwnedBy)})
	}
	var mu sync.Mutex
	var buf []core.QualityResult
	runner := &quality.Runner{Skipper: s.skipperFor(run), Client: cl, Spec: req.Suite.Specs.Quality, Set: set, Sink: quality.SinkFunc(func(_ context.Context, r core.QualityResult) error {
		mu.Lock()
		buf = append(buf, r)
		var rows []core.QualityResult
		if len(buf) >= 25 {
			rows = buf
			buf = nil
		}
		mu.Unlock()
		if rows != nil {
			return s.Store.InsertQualityResults(run.ID, rows)
		}
		return nil
	}), Report: s.reporter(run.ID)}
	results, runErr := runner.Run(ctx, targets)
	mu.Lock()
	rest := buf
	buf = nil
	mu.Unlock()
	if err := s.Store.InsertQualityResults(run.ID, rest); err != nil && runErr == nil {
		runErr = err
	}
	if runErr != nil {
		return s.fail(run, runErr)
	}
	scores := quality.Summarize(results, runner.Set)
	if err := s.Store.SetRunSummary(run.ID, map[string]any{"quality": scores}); err != nil {
		return s.fail(run, err)
	}
	_ = s.Store.SetRunStatus(run.ID, core.StatusDone, "")
	s.Bus.Publish(run.ID, core.Event{Kind: core.EventDone, Stage: "quality", Message: "done", Data: scores})
	return nil
}

func (s *Service) runSuite(ctx context.Context, run *store.Run, req Request, ep config.Endpoint, epID string) error {
	_ = s.Store.SetRunStatus(run.ID, core.StatusRunning, "")
	children := map[string]string{}
	child := func(kind string, models []string) (*store.Run, Request) {
		creq := req
		creq.Kind = kind
		creq.Models = models
		cr, _ := s.createRun(creq, ep, epID, run.ID)
		children[kind] = cr.ID
		s.Bus.Publish(run.ID, core.Event{Kind: core.EventLog, Stage: kind, Message: "stage started", Data: map[string]string{"child": cr.ID, "kind": kind}})
		return cr, creq
	}
	forward := func(childID string) func() {
		ch, unsub := s.Bus.Subscribe(childID, 512)
		done := make(chan struct{})
		go func() {
			defer close(done)
			for ev := range ch {
				ev.RunID = ""
				ev.Data = nil
				s.Bus.Publish(run.ID, ev)
			}
		}()
		return func() {
			unsub()
		}
	}
	var scanRes *ScanOutcome
	if req.Suite.Scan {
		cr, creq := child(core.KindScan, nil)
		stop := forward(cr.ID)
		out, err := s.runScan(ctx, cr, creq, ep, epID)
		stop()
		if err != nil {
			return s.fail(run, err)
		}
		scanRes = out
	}
	pick := func(n int) []string {
		if scanRes == nil {
			return req.Models
		}
		var ids []string
		for _, m := range scanRes.Result.Available {
			if len(ids) >= n {
				break
			}
			ids = append(ids, m.ID)
		}
		return ids
	}
	if req.Suite.Latency {
		if ids := pick(req.Suite.LatencyTop); len(ids) > 0 {
			cr, creq := child(core.KindLatency, ids)
			stop := forward(cr.ID)
			err := s.runLatency(ctx, cr, creq, ep, epID, ids)
			stop()
			if err != nil {
				return s.fail(run, err)
			}
		}
	}
	if req.Suite.Capability {
		if ids := pick(req.Suite.CapabilityTop); len(ids) > 0 {
			cr, creq := child(core.KindCapability, ids)
			stop := forward(cr.ID)
			err := s.runCapability(ctx, cr, creq, ep, epID, ids)
			stop()
			if err != nil {
				return s.fail(run, err)
			}
		}
	}
	if req.Suite.Quality {
		if ids := pick(req.Suite.QualityTop); len(ids) > 0 {
			cr, creq := child(core.KindQuality, ids)
			stop := forward(cr.ID)
			err := s.runQuality(ctx, cr, creq, ep, epID, ids)
			stop()
			if err != nil {
				return s.fail(run, err)
			}
		}
	}
	if req.Suite.Load {
		model := req.Suite.LoadModel
		if model == "" {
			if ids := pick(1); len(ids) > 0 {
				model = ids[0]
			}
		}
		if model != "" {
			creq := req
			creq.Suite.Specs.Load.Model = model
			cr, creq := child(core.KindLoad, []string{model})
			creq.Suite.Specs.Load.Model = model
			stop := forward(cr.ID)
			err := s.runLoad(ctx, cr, creq, ep, epID)
			stop()
			if err != nil {
				return s.fail(run, err)
			}
		}
	}
	summary := map[string]any{"children": children}
	if scanRes != nil {
		summary["summary"] = scanRes.Result.Summary
		summary["dialect"] = scanRes.Dialect
	}
	if err := s.Store.SetRunSummary(run.ID, summary); err != nil {
		return s.fail(run, err)
	}
	_ = s.Store.SetRunStatus(run.ID, core.StatusDone, "")
	s.Bus.Publish(run.ID, core.Event{Kind: core.EventDone, Stage: "suite", Message: "suite complete", Data: children})
	return nil
}

// ScanResult rebuilds the analysis of a scan run from the store.
func (s *Service) ScanResult(runID string) (*analyze.Result, *store.Run, error) {
	run, err := s.Store.GetRun(runID)
	if err != nil {
		return nil, nil, err
	}
	if run.Kind == core.KindSuite {
		kids, _ := s.Store.ChildRuns(runID)
		for _, k := range kids {
			if k.Kind == core.KindScan {
				run = k
				break
			}
		}
	}
	if run.Kind != core.KindScan {
		return nil, run, fmt.Errorf("run %s is a %s run, not a scan", runID, run.Kind)
	}
	models, err := s.Store.ListRunModels(run.ID)
	if err != nil {
		return nil, run, err
	}
	attempts, err := s.Store.ListAttempts(run.ID)
	if err != nil {
		return nil, run, err
	}
	var scoring analyze.Scoring
	var spec struct {
		Suite config.Suite `json:"suite"`
	}
	if json.Unmarshal(run.Spec, &spec) == nil {
		scoring = spec.Suite.Specs.Scoring
	}
	pres := analyze.Enrich(models, attempts, s.assignmentOptions(run.EndpointName, run.EndpointID))
	res := analyze.Build(models, attempts, scoring, &pres)
	return res, run, nil
}

// Report renders the Markdown report of a run (scan or suite).
func (s *Service) Report(runID, lang string) (string, error) {
	run, err := s.Store.GetRun(runID)
	if err != nil {
		return "", err
	}
	meta := report.Meta{Endpoint: run.BaseURL, Dialect: run.Dialect, Version: s.Version, Generated: time.Now(), Started: run.StartedAt, Finished: run.FinishedAt, Preset: run.Preset, Lang: lang}
	if run.EndpointName != "" && run.EndpointName != run.BaseURL {
		meta.Title = run.EndpointName
		if lang == "en" {
			meta.Title += " — LLM API Speed Test Report"
		} else {
			meta.Title += " — LLM API 测速报告"
		}
	}
	var scanRun *store.Run
	children := []*store.Run{}
	switch run.Kind {
	case core.KindScan:
		scanRun = run
	case core.KindSuite:
		children, _ = s.Store.ChildRuns(runID)
		for _, k := range children {
			if k.Kind == core.KindScan {
				scanRun = k
			}
		}
	default:
		children = []*store.Run{run}
	}
	for _, k := range children {
		if k.Kind == core.KindScan {
			continue
		}
		sec, err := s.sectionFor(k, lang)
		if err == nil && sec != nil {
			meta.Extra = append(meta.Extra, *sec)
		}
	}
	if scanRun == nil {
		// non-scan run: render only the extra sections under a minimal header
		var b strings.Builder
		if lang == "en" {
			fmt.Fprintf(&b, "# LLM API Speed Test Report\n\n- Endpoint: `%s`\n- Generated: %s\n\n", run.BaseURL, time.Now().Format("2006-01-02 15:04:05"))
		} else {
			fmt.Fprintf(&b, "# LLM API 测速报告\n\n- 端点：`%s`\n- 生成时间：%s\n\n", run.BaseURL, time.Now().Format("2006-01-02 15:04:05"))
		}
		for _, sec := range meta.Extra {
			fmt.Fprintf(&b, "## %s\n\n%s\n\n", sec.Title, strings.TrimRight(sec.Body, "\n"))
		}
		return b.String(), nil
	}
	res, _, err := s.ScanResult(scanRun.ID)
	if err != nil {
		return "", err
	}
	var sum struct {
		Rounds []scan.RoundSummary `json:"rounds"`
	}
	if json.Unmarshal(scanRun.Summary, &sum) == nil {
		for _, r := range sum.Rounds {
			meta.Rounds = append(meta.Rounds, report.RoundLine{Round: r.Round, Kind: r.Kind, Targets: r.Targets, OK: r.OK, Fail: r.Fail})
		}
	}
	meta.Started, meta.Finished = scanRun.StartedAt, scanRun.FinishedAt
	if run.Kind == core.KindSuite && !run.FinishedAt.IsZero() {
		meta.Finished = run.FinishedAt
	}
	return report.Markdown(res, meta), nil
}

func (s *Service) sectionFor(run *store.Run, lang string) (*report.Section, error) {
	switch run.Kind {
	case core.KindLatency:
		samples, err := s.Store.ListLatencySamples(run.ID)
		if err != nil {
			return nil, err
		}
		sec := report.LatencySection(latency.Summarize(samples), lang)
		return &sec, nil
	case core.KindLoad:
		var sum struct {
			Load  []load.LevelResult `json:"load"`
			Model string             `json:"model"`
		}
		_ = json.Unmarshal(run.Summary, &sum)
		sec := report.LoadSection(sum.Model, sum.Load, lang)
		return &sec, nil
	case core.KindCapability:
		results, err := s.Store.ListCapabilityResults(run.ID)
		if err != nil {
			return nil, err
		}
		sec := report.CapabilitySection(capability.Summarize(results, nil), lang)
		return &sec, nil
	case core.KindQuality:
		results, err := s.Store.ListQualityResults(run.ID)
		if err != nil {
			return nil, err
		}
		sec := report.QualitySection(quality.Summarize(results, nil), lang)
		return &sec, nil
	}
	return nil, nil
}

// ExportJSON renders the JSON export of a scan run.
func (s *Service) ExportJSON(runID string) ([]byte, error) {
	res, run, err := s.ScanResult(runID)
	if err != nil {
		return nil, err
	}
	return report.JSON(res, report.Meta{Endpoint: run.BaseURL, Dialect: run.Dialect, Version: s.Version, Generated: time.Now(), Started: run.StartedAt, Finished: run.FinishedAt, Preset: run.Preset})
}

// ExportCSV renders the CSV export of a scan run.
func (s *Service) ExportCSV(runID string) ([]byte, error) {
	res, _, err := s.ScanResult(runID)
	if err != nil {
		return nil, err
	}
	return report.CSV(res)
}

// FamilyOverridesFor exposes family overrides of an endpoint for the API.
func (s *Service) FamilyOverridesFor(endpointID string) (map[string]family.Override, error) {
	m, _, err := s.Store.FamilyOverrides(endpointID)
	return m, err
}
