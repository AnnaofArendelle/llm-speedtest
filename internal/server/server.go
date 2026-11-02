// Package server exposes the HTTP API, server-sent events and the embedded
// web UI.
package server

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"io/fs"
	"net/http"
	"path"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/analyze"
	"github.com/AnnaofArendelle/llm-speedtest/internal/app"
	"github.com/AnnaofArendelle/llm-speedtest/internal/capability"
	"github.com/AnnaofArendelle/llm-speedtest/internal/classify"
	"github.com/AnnaofArendelle/llm-speedtest/internal/config"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/dialect"
	"github.com/AnnaofArendelle/llm-speedtest/internal/latency"
	"github.com/AnnaofArendelle/llm-speedtest/internal/provider"
	"github.com/AnnaofArendelle/llm-speedtest/internal/quality"
	"github.com/AnnaofArendelle/llm-speedtest/internal/store"
	"github.com/AnnaofArendelle/llm-speedtest/internal/version"
	"github.com/AnnaofArendelle/llm-speedtest/web"
)

// Options configures the server.
type Options struct {
	AccessToken string
	MaxBody     int64
}

// Server serves the API and UI.
type Server struct {
	svc   *app.Service
	opts  Options
	mux   *http.ServeMux
	cache sync.Map // runID -> cachedResult
}

type cachedResult struct {
	res *analyze.Result
	run *store.Run
	at  time.Time
}

// New creates a server.
func New(svc *app.Service, opts Options) *Server {
	if opts.MaxBody <= 0 {
		opts.MaxBody = 1 << 20
	}
	s := &Server{svc: svc, opts: opts, mux: http.NewServeMux()}
	s.routes()
	return s
}

// Handler returns the root handler.
func (s *Server) Handler() http.Handler { return s.mux }

func (s *Server) routes() {
	m := s.mux
	m.HandleFunc("GET /healthz", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, map[string]any{"ok": true}) })
	api := func(pattern string, h http.HandlerFunc) { m.HandleFunc(pattern, s.auth(h)) }
	api("GET /api/version", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, version.Get()) })
	api("GET /api/reasons", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, classify.All()) })
	api("GET /api/dialects", func(w http.ResponseWriter, _ *http.Request) { writeJSON(w, 200, dialect.Names()) })
	api("GET /api/probes", func(w http.ResponseWriter, _ *http.Request) {
		var out []map[string]string
		for _, id := range capability.Probes {
			out = append(out, map[string]string{"id": id, "zh": capability.Describe(id, "zh"), "en": capability.Describe(id, "en")})
		}
		writeJSON(w, 200, out)
	})
	api("GET /api/presets", func(w http.ResponseWriter, _ *http.Request) {
		writeJSON(w, 200, map[string]config.Suite{config.ProfileQuick: config.DefaultSuite(config.ProfileQuick), config.ProfileStandard: config.DefaultSuite(config.ProfileStandard), config.ProfileDeep: config.DefaultSuite(config.ProfileDeep)})
	})
	api("GET /api/questions", func(w http.ResponseWriter, _ *http.Request) {
		set, err := quality.Default()
		if err != nil {
			writeErr(w, 500, err)
			return
		}
		writeJSON(w, 200, set)
	})

	// endpoints
	api("GET /api/endpoints", s.listEndpoints)
	api("POST /api/endpoints", s.saveEndpoint)
	api("GET /api/endpoints/{id}", s.getEndpoint)
	api("PUT /api/endpoints/{id}", s.saveEndpoint)
	api("DELETE /api/endpoints/{id}", s.deleteEndpoint)
	api("POST /api/endpoints/{id}/check", s.checkEndpoint)
	api("POST /api/endpoints/{id}/discover", s.discoverEndpoint)
	api("GET /api/endpoints/{id}/overrides", s.getOverrides)
	api("PUT /api/endpoints/{id}/overrides/provider", s.putProviderOverride)
	api("PUT /api/endpoints/{id}/overrides/family", s.putFamilyOverride)
	api("POST /api/check", s.checkAdhoc)
	api("POST /api/discover", s.discoverAdhoc)

	// runs
	api("POST /api/runs", s.startRun)
	api("GET /api/runs", s.listRuns)
	api("GET /api/runs/{id}", s.getRun)
	api("DELETE /api/runs/{id}", s.deleteRun)
	api("POST /api/runs/{id}/cancel", s.cancelRun)
	api("POST /api/runs/{id}/skip", s.skipRun)
	api("GET /api/runs/{id}/events", s.events)
	api("GET /api/runs/{id}/summary", s.runSummary)
	api("GET /api/runs/{id}/families", s.runFamilies)
	api("GET /api/runs/{id}/families/{key}", s.runFamily)
	api("GET /api/runs/{id}/providers", s.runProviders)
	api("GET /api/runs/{id}/models", s.runModels)
	api("GET /api/runs/{id}/models/{model...}", s.runModel)
	api("GET /api/runs/{id}/failures", s.runFailures)
	api("GET /api/runs/{id}/profile/model/{model...}", s.modelProfile)
	api("GET /api/runs/{id}/profile/family/{key}", s.familyProfile)
	api("GET /api/runs/{id}/latency", s.runLatency)
	api("GET /api/runs/{id}/load", s.runLoad)
	api("GET /api/runs/{id}/capabilities", s.runCapabilities)
	api("GET /api/runs/{id}/quality", s.runQuality)
	api("GET /api/runs/{id}/report.md", s.runReport)
	api("GET /api/runs/{id}/export.json", s.runExportJSON)
	api("GET /api/runs/{id}/export.csv", s.runExportCSV)
	api("GET /api/compare", s.compare)

	// UI
	m.HandleFunc("/", s.static)
}

// ---------------------------------------------------------------- helpers

func writeJSON(w http.ResponseWriter, status int, v any) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(v)
}

func writeErr(w http.ResponseWriter, status int, err error) {
	writeJSON(w, status, map[string]any{"error": err.Error()})
}

func (s *Server) decode(w http.ResponseWriter, r *http.Request, v any) bool {
	body := http.MaxBytesReader(w, r.Body, s.opts.MaxBody)
	dec := json.NewDecoder(body)
	if err := dec.Decode(v); err != nil {
		writeErr(w, 400, fmt.Errorf("invalid JSON body: %w", err))
		return false
	}
	return true
}

func (s *Server) auth(h http.HandlerFunc) http.HandlerFunc {
	if s.opts.AccessToken == "" {
		return h
	}
	return func(w http.ResponseWriter, r *http.Request) {
		tok := strings.TrimPrefix(r.Header.Get("Authorization"), "Bearer ")
		if tok == "" {
			tok = r.URL.Query().Get("token")
		}
		if tok != s.opts.AccessToken {
			writeErr(w, 401, errors.New("unauthorized"))
			return
		}
		h(w, r)
	}
}

// nonNilDiscover guarantees that the collection fields of a discovery result
// serialise as JSON arrays rather than null.
func nonNilDiscover(res *app.DiscoverResult) *app.DiscoverResult {
	if res.Models == nil {
		res.Models = []core.RunModel{}
	}
	if res.Providers == nil {
		res.Providers = []provider.Provider{}
	}
	return res
}

func notFoundOr(w http.ResponseWriter, err error) {
	if errors.Is(err, store.ErrNotFound) {
		writeErr(w, 404, err)
		return
	}
	writeErr(w, 500, err)
}

func qInt(r *http.Request, key string, def int) int {
	if v, err := strconv.Atoi(r.URL.Query().Get(key)); err == nil {
		return v
	}
	return def
}

// ---------------------------------------------------------------- endpoints

type endpointBody struct {
	Name     string            `json:"name"`
	BaseURL  string            `json:"base_url"`
	APIKey   string            `json:"api_key"`
	ClearKey bool              `json:"clear_key"`
	Dialect  string            `json:"dialect"`
	Headers  map[string]string `json:"headers"`
	Insecure bool              `json:"insecure"`
	Proxy    string            `json:"proxy"`
}

func (s *Server) listEndpoints(w http.ResponseWriter, _ *http.Request) {
	eps, err := s.svc.Store.ListEndpoints()
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	if eps == nil {
		eps = []*store.Endpoint{}
	}
	writeJSON(w, 200, eps)
}

func (s *Server) saveEndpoint(w http.ResponseWriter, r *http.Request) {
	var b endpointBody
	if !s.decode(w, r, &b) {
		return
	}
	if strings.TrimSpace(b.BaseURL) == "" {
		writeErr(w, 400, errors.New("base_url is required"))
		return
	}
	ep := &store.Endpoint{ID: r.PathValue("id"), Name: b.Name, BaseURL: b.BaseURL, Dialect: b.Dialect, Headers: b.Headers, APIKey: b.APIKey, Insecure: b.Insecure, Proxy: b.Proxy}
	if ep.ID != "" {
		existing, err := s.svc.Store.GetEndpoint(ep.ID)
		if err != nil {
			notFoundOr(w, err)
			return
		}
		ep.CreatedAt = existing.CreatedAt
		ep.KeyHint = existing.KeyHint
	}
	if ep.Name == "" {
		ep.Name = ep.BaseURL
	}
	if err := s.svc.Store.SaveEndpoint(ep); err != nil {
		writeErr(w, 500, err)
		return
	}
	if b.ClearKey {
		_ = s.svc.Store.ClearEndpointKey(ep.ID)
	}
	out, err := s.svc.Store.GetEndpoint(ep.ID)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, out)
}

func (s *Server) getEndpoint(w http.ResponseWriter, r *http.Request) {
	ep, err := s.svc.Store.GetEndpoint(r.PathValue("id"))
	if err != nil {
		notFoundOr(w, err)
		return
	}
	writeJSON(w, 200, ep)
}

func (s *Server) deleteEndpoint(w http.ResponseWriter, r *http.Request) {
	if err := s.svc.Store.DeleteEndpoint(r.PathValue("id")); err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (s *Server) endpointConfig(id string) (config.Endpoint, error) {
	ep, err := s.svc.Store.GetEndpointWithKey(id)
	if err != nil {
		return config.Endpoint{}, err
	}
	return config.Endpoint{Name: ep.Name, BaseURL: ep.BaseURL, APIKey: ep.APIKey, Dialect: ep.Dialect, Headers: ep.Headers, Insecure: ep.Insecure, Proxy: ep.Proxy}, nil
}

func (s *Server) checkEndpoint(w http.ResponseWriter, r *http.Request) {
	ep, err := s.endpointConfig(r.PathValue("id"))
	if err != nil {
		notFoundOr(w, err)
		return
	}
	writeJSON(w, 200, s.svc.Check(r.Context(), ep))
}

type adhocBody struct {
	Endpoint config.Endpoint `json:"endpoint"`
	APIKey   string          `json:"api_key"`
	Filter   *config.Filter  `json:"filter"`
}

func (s *Server) checkAdhoc(w http.ResponseWriter, r *http.Request) {
	var b adhocBody
	if !s.decode(w, r, &b) {
		return
	}
	b.Endpoint.APIKey = b.APIKey
	writeJSON(w, 200, s.svc.Check(r.Context(), b.Endpoint))
}

func (s *Server) discoverEndpoint(w http.ResponseWriter, r *http.Request) {
	ep, err := s.endpointConfig(r.PathValue("id"))
	if err != nil {
		notFoundOr(w, err)
		return
	}
	var b struct {
		Filter *config.Filter `json:"filter"`
	}
	if r.ContentLength > 0 && !s.decode(w, r, &b) {
		return
	}
	filter := config.DefaultFilter()
	if b.Filter != nil {
		filter = *b.Filter
	}
	res, err := s.svc.Discover(r.Context(), ep, r.PathValue("id"), filter)
	if err != nil {
		writeErr(w, 502, err)
		return
	}
	writeJSON(w, 200, nonNilDiscover(res))
}

func (s *Server) discoverAdhoc(w http.ResponseWriter, r *http.Request) {
	var b adhocBody
	if !s.decode(w, r, &b) {
		return
	}
	b.Endpoint.APIKey = b.APIKey
	filter := config.DefaultFilter()
	if b.Filter != nil {
		filter = *b.Filter
	}
	res, err := s.svc.Discover(r.Context(), b.Endpoint, "", filter)
	if err != nil {
		writeErr(w, 502, err)
		return
	}
	writeJSON(w, 200, nonNilDiscover(res))
}

func (s *Server) getOverrides(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	_, prows, err := s.svc.Store.ProviderOverrides(id)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	_, frows, err := s.svc.Store.FamilyOverrides(id)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	if prows == nil {
		prows = []store.ProviderOverrideRow{}
	}
	if frows == nil {
		frows = []store.FamilyOverrideRow{}
	}
	writeJSON(w, 200, map[string]any{"providers": prows, "families": frows})
}

func (s *Server) putProviderOverride(w http.ResponseWriter, r *http.Request) {
	var row store.ProviderOverrideRow
	if !s.decode(w, r, &row) {
		return
	}
	if row.RawKey == "" {
		writeErr(w, 400, errors.New("raw_key is required"))
		return
	}
	if err := s.svc.Store.SetProviderOverride(r.PathValue("id"), row); err != nil {
		writeErr(w, 500, err)
		return
	}
	s.cache.Range(func(k, _ any) bool { s.cache.Delete(k); return true })
	writeJSON(w, 200, map[string]any{"ok": true})
}

func (s *Server) putFamilyOverride(w http.ResponseWriter, r *http.Request) {
	var row store.FamilyOverrideRow
	if !s.decode(w, r, &row) {
		return
	}
	if row.ModelID == "" {
		writeErr(w, 400, errors.New("model_id is required"))
		return
	}
	if err := s.svc.Store.SetFamilyOverride(r.PathValue("id"), row); err != nil {
		writeErr(w, 500, err)
		return
	}
	s.cache.Range(func(k, _ any) bool { s.cache.Delete(k); return true })
	writeJSON(w, 200, map[string]any{"ok": true})
}

// ---------------------------------------------------------------- runs

type startBody struct {
	Kind       string           `json:"kind"`
	EndpointID string           `json:"endpoint_id"`
	Endpoint   *config.Endpoint `json:"endpoint"`
	APIKey     string           `json:"api_key"`
	Profile    string           `json:"profile"`
	Suite      *config.Suite    `json:"suite"`
	Models     []string         `json:"models"`
	Note       string           `json:"note"`
	Confirm    bool             `json:"confirm"`
}

func (s *Server) startRun(w http.ResponseWriter, r *http.Request) {
	var b startBody
	if !s.decode(w, r, &b) {
		return
	}
	if b.Kind == "" {
		b.Kind = core.KindSuite
	}
	suite := config.DefaultSuite(b.Profile)
	if b.Suite != nil {
		suite = *b.Suite
		if suite.Profile == "" {
			suite.Profile = b.Profile
		}
	}
	suite.Normalize()
	req := app.Request{Kind: b.Kind, EndpointID: b.EndpointID, Suite: suite, Models: b.Models, Note: b.Note, Confirm: b.Confirm}
	if b.Endpoint != nil {
		ep := *b.Endpoint
		ep.APIKey = b.APIKey
		req.Endpoint = &ep
	}
	id, err := s.svc.Start(req)
	if err != nil {
		if errors.Is(err, app.ErrConfirmRequired) {
			writeErr(w, 409, err)
			return
		}
		if errors.Is(err, store.ErrNotFound) {
			writeErr(w, 404, err)
			return
		}
		writeErr(w, 400, err)
		return
	}
	writeJSON(w, 202, map[string]any{"id": id})
}

func (s *Server) listRuns(w http.ResponseWriter, r *http.Request) {
	runs, total, err := s.svc.Store.ListRuns(r.URL.Query().Get("kind"), r.URL.Query().Get("parent"), qInt(r, "limit", 50), qInt(r, "offset", 0))
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	if runs == nil {
		runs = []*store.Run{}
	}
	type item struct {
		*store.Run
		Running bool `json:"running"`
	}
	out := make([]item, 0, len(runs))
	for _, run := range runs {
		out = append(out, item{run, s.svc.IsRunning(run.ID)})
	}
	writeJSON(w, 200, map[string]any{"runs": out, "total": total})
}

func (s *Server) getRun(w http.ResponseWriter, r *http.Request) {
	run, err := s.svc.Store.GetRun(r.PathValue("id"))
	if err != nil {
		notFoundOr(w, err)
		return
	}
	kids, _ := s.svc.Store.ChildRuns(run.ID)
	if kids == nil {
		kids = []*store.Run{}
	}
	latest, _ := s.svc.Bus.Latest(run.ID)
	writeJSON(w, 200, map[string]any{"run": run, "children": kids, "running": s.svc.IsRunning(run.ID), "latest": latest, "inflight": s.svc.Inflight(run.ID)})
}

func (s *Server) deleteRun(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	if s.svc.IsRunning(id) {
		writeErr(w, 409, errors.New("run is in progress; cancel it first"))
		return
	}
	if err := s.svc.Store.DeleteRun(id); err != nil {
		writeErr(w, 500, err)
		return
	}
	s.cache.Delete(id)
	writeJSON(w, 200, map[string]any{"ok": true})
}

// skipRun skips one model (or every in-flight model) of a run in progress.
func (s *Server) skipRun(w http.ResponseWriter, r *http.Request) {
	var b struct {
		Model string `json:"model"`
		All   bool   `json:"all"`
	}
	if r.ContentLength > 0 && !s.decode(w, r, &b) {
		return
	}
	ids, ok := s.svc.Skip(r.PathValue("id"), b.Model, b.All)
	if !ok {
		writeErr(w, 409, errors.New("run is not in progress"))
		return
	}
	if ids == nil {
		ids = []string{}
	}
	writeJSON(w, 200, map[string]any{"skipped": ids})
}

func (s *Server) cancelRun(w http.ResponseWriter, r *http.Request) {
	ok := s.svc.Cancel(r.PathValue("id"))
	writeJSON(w, 200, map[string]any{"cancelled": ok})
}

func (s *Server) events(w http.ResponseWriter, r *http.Request) {
	id := r.PathValue("id")
	run, err := s.svc.Store.GetRun(id)
	if err != nil {
		notFoundOr(w, err)
		return
	}
	fl, ok := w.(http.Flusher)
	if !ok {
		writeErr(w, 500, errors.New("streaming unsupported"))
		return
	}
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("X-Accel-Buffering", "no")
	w.WriteHeader(200)
	send := func(ev core.Event) bool {
		raw, _ := json.Marshal(ev)
		if ev.Seq > 0 {
			fmt.Fprintf(w, "id: %d\n", ev.Seq)
		}
		if _, err := fmt.Fprintf(w, "event: %s\ndata: %s\n\n", ev.Kind, raw); err != nil {
			return false
		}
		fl.Flush()
		return true
	}
	var after int64
	if v := r.Header.Get("Last-Event-ID"); v != "" {
		after, _ = strconv.ParseInt(v, 10, 64)
	}
	// subscribe first so no event is lost between snapshot and stream
	ch, unsub := s.svc.Bus.Subscribe(id, 512)
	defer unsub()
	latest, _ := s.svc.Bus.Latest(id)
	snapshot := core.Event{Kind: "snapshot", RunID: id, At: time.Now(), Data: map[string]any{"status": run.Status, "running": s.svc.IsRunning(id), "latest": latest, "inflight": s.svc.Inflight(id)}}
	if !send(snapshot) {
		return
	}
	for _, ev := range s.svc.Bus.History(id, after) {
		if !send(ev) {
			return
		}
	}
	if !s.svc.IsRunning(id) {
		send(core.Event{Kind: core.EventDone, RunID: id, At: time.Now(), Message: "run is not in progress", Data: map[string]any{"status": run.Status}})
		return
	}
	ping := time.NewTicker(15 * time.Second)
	defer ping.Stop()
	for {
		select {
		case <-r.Context().Done():
			return
		case <-ping.C:
			if _, err := fmt.Fprint(w, ": ping\n\n"); err != nil {
				return
			}
			fl.Flush()
		case ev := <-ch:
			if !send(ev) {
				return
			}
			if ev.Kind == core.EventDone && (ev.Stage == "suite" || ev.Stage == "" || run.Kind != core.KindSuite) && ev.RunID == id {
				if !s.svc.IsRunning(id) {
					return
				}
			}
			if ev.Kind == core.EventError {
				return
			}
		}
	}
}

func (s *Server) scanResult(id string) (*analyze.Result, *store.Run, error) {
	if c, ok := s.cache.Load(id); ok {
		cr := c.(cachedResult)
		if time.Since(cr.at) < 30*time.Second && !s.svc.IsRunning(id) && !s.svc.IsRunning(cr.run.ID) {
			return cr.res, cr.run, nil
		}
	}
	res, run, err := s.svc.ScanResult(id)
	if err != nil {
		return nil, run, err
	}
	// Never cache a result computed while the run was still in progress:
	// it would be served as the final result for up to 30 s after completion.
	if !s.svc.IsRunning(id) && !s.svc.IsRunning(run.ID) {
		s.cache.Store(id, cachedResult{res: res, run: run, at: time.Now()})
	}
	return res, run, nil
}

func (s *Server) runSummary(w http.ResponseWriter, r *http.Request) {
	res, run, err := s.scanResult(r.PathValue("id"))
	if err != nil {
		if run != nil && !errors.Is(err, store.ErrNotFound) {
			writeJSON(w, 200, map[string]any{"run": run, "summary": json.RawMessage(run.Summary), "partial": true, "status": run.Status, "kind": run.Kind})
			return
		}
		notFoundOr(w, err)
		return
	}
	top := res.Families
	if len(top) > 15 {
		top = top[:15]
	}
	topModels := res.Available
	if len(topModels) > 20 {
		topModels = topModels[:20]
	}
	partial := run.Status != core.StatusDone || s.svc.IsRunning(run.ID)
	writeJSON(w, 200, map[string]any{"run": run, "summary": res.Summary, "scoring": res.Scoring, "provider_field": res.ProviderField, "provider_rationale": res.ProviderRationale,
		"provider_suggestions": res.Suggestions, "top_families": top, "top_models": topModels, "providers": res.Providers, "partial": partial, "status": run.Status, "kind": run.Kind})
}

func (s *Server) runFamilies(w http.ResponseWriter, r *http.Request) {
	res, _, err := s.scanResult(r.PathValue("id"))
	if err != nil {
		notFoundOr(w, err)
		return
	}
	order := r.URL.Query().Get("order")
	switch order {
	case "stability":
		writeJSON(w, 200, map[string]any{"families": res.FamiliesByStability, "unavailable": res.UnavailableFamilies})
	default:
		writeJSON(w, 200, map[string]any{"families": res.Families, "unavailable": res.UnavailableFamilies})
	}
}

func (s *Server) runFamily(w http.ResponseWriter, r *http.Request) {
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
	members := f.Members
	if members == nil {
		members = []*analyze.ModelStat{}
	}
	writeJSON(w, 200, map[string]any{"family": f, "members": members})
}

func (s *Server) runProviders(w http.ResponseWriter, r *http.Request) {
	res, _, err := s.scanResult(r.PathValue("id"))
	if err != nil {
		notFoundOr(w, err)
		return
	}
	writeJSON(w, 200, map[string]any{"providers": res.Providers, "field": res.ProviderField, "rationale": res.ProviderRationale, "suggestions": res.Suggestions})
}

func (s *Server) runModels(w http.ResponseWriter, r *http.Request) {
	res, _, err := s.scanResult(r.PathValue("id"))
	if err != nil {
		notFoundOr(w, err)
		return
	}
	q := r.URL.Query()
	status, prov, reason, tier, fam, search := q.Get("status"), q.Get("provider"), q.Get("reason"), q.Get("tier"), q.Get("family"), strings.ToLower(q.Get("q"))
	var list []*analyze.ModelStat
	for _, m := range res.Models {
		if status != "" && m.Status != status {
			continue
		}
		if prov != "" && m.ProviderKey != prov && m.ProviderName != prov {
			continue
		}
		if reason != "" && m.Reason != reason {
			continue
		}
		if tier != "" && m.Tier != tier {
			continue
		}
		if fam != "" && m.FamilyKey != fam {
			continue
		}
		if search != "" && !strings.Contains(strings.ToLower(m.ID), search) && !strings.Contains(strings.ToLower(m.ProviderName), search) && !strings.Contains(strings.ToLower(m.FamilyCanon), search) {
			continue
		}
		list = append(list, m)
	}
	sortModels(list, q.Get("sort"), q.Get("order"))
	total := len(list)
	page, size := qInt(r, "page", 1), qInt(r, "size", 100)
	if size <= 0 || size > 5000 {
		size = 100
	}
	if page <= 0 {
		page = 1
	}
	start := (page - 1) * size
	if start > total {
		start = total
	}
	end := start + size
	if end > total {
		end = total
	}
	if list == nil {
		list = []*analyze.ModelStat{}
	}
	writeJSON(w, 200, map[string]any{"models": list[start:end], "total": total, "page": page, "size": size})
}

func sortModels(list []*analyze.ModelStat, key, order string) {
	desc := order == "desc"
	less := func(i, j int) bool {
		a, b := list[i], list[j]
		var r bool
		switch key {
		case "id":
			r = a.ID < b.ID
		case "provider":
			r = a.ProviderName < b.ProviderName
		case "family":
			r = a.FamilyCanon < b.FamilyCanon
		case "median":
			r = a.Med < b.Med
		case "min":
			r = a.Min < b.Min
		case "max":
			r = a.Max < b.Max
		case "rate":
			r = a.Rate < b.Rate
		case "attempts":
			r = a.Attempts < b.Attempts
		case "reason":
			r = a.Reason < b.Reason
		case "status":
			r = a.Status < b.Status
		default:
			return false
		}
		if desc {
			return !r && a.ID != b.ID
		}
		return r
	}
	if key != "" {
		sort.SliceStable(list, less)
	}
}

func (s *Server) runModel(w http.ResponseWriter, r *http.Request) {
	res, _, err := s.scanResult(r.PathValue("id"))
	if err != nil {
		notFoundOr(w, err)
		return
	}
	m := res.Model(r.PathValue("model"))
	if m == nil {
		writeErr(w, 404, errors.New("model not found"))
		return
	}
	history := m.History
	if history == nil {
		history = []core.Attempt{}
	}
	writeJSON(w, 200, map[string]any{"model": m, "history": history})
}

func (s *Server) runFailures(w http.ResponseWriter, r *http.Request) {
	res, _, err := s.scanResult(r.PathValue("id"))
	if err != nil {
		notFoundOr(w, err)
		return
	}
	type group struct {
		Reason     string                          `json:"reason"`
		Info       classify.ReasonInfo             `json:"info"`
		Count      int                             `json:"count"`
		ByProvider map[string][]*analyze.ModelStat `json:"by_provider"`
	}
	groups := map[string]*group{}
	for _, m := range res.Models {
		if m.Status == "ok" {
			continue
		}
		g := groups[m.Reason]
		if g == nil {
			g = &group{Reason: m.Reason, Info: classify.Info(m.Reason), ByProvider: map[string][]*analyze.ModelStat{}}
			groups[m.Reason] = g
		}
		g.Count++
		g.ByProvider[m.ProviderName] = append(g.ByProvider[m.ProviderName], m)
	}
	var out []*group
	for _, code := range classify.Order {
		if g, ok := groups[code]; ok {
			out = append(out, g)
		}
	}
	if out == nil {
		out = []*group{}
	}
	writeJSON(w, 200, map[string]any{"groups": out})
}

func (s *Server) childOf(id, kind string) (*store.Run, error) {
	run, err := s.svc.Store.GetRun(id)
	if err != nil {
		return nil, err
	}
	if run.Kind == kind {
		return run, nil
	}
	if run.Kind == core.KindSuite {
		kids, _ := s.svc.Store.ChildRuns(id)
		for _, k := range kids {
			if k.Kind == kind {
				return k, nil
			}
		}
	}
	return nil, store.ErrNotFound
}

func (s *Server) runLatency(w http.ResponseWriter, r *http.Request) {
	run, err := s.childOf(r.PathValue("id"), core.KindLatency)
	if err != nil {
		notFoundOr(w, err)
		return
	}
	samples, err := s.svc.Store.ListLatencySamples(run.ID)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	if samples == nil {
		samples = []core.LatencySample{}
	}
	writeJSON(w, 200, map[string]any{"run": run, "stats": latency.Summarize(samples), "samples": samples})
}

func (s *Server) runLoad(w http.ResponseWriter, r *http.Request) {
	run, err := s.childOf(r.PathValue("id"), core.KindLoad)
	if err != nil {
		notFoundOr(w, err)
		return
	}
	reqs, err := s.svc.Store.ListLoadRequests(run.ID)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	if reqs == nil {
		reqs = []core.LoadRequest{}
	}
	writeJSON(w, 200, map[string]any{"run": run, "summary": json.RawMessage(run.Summary), "requests": reqs})
}

func (s *Server) runCapabilities(w http.ResponseWriter, r *http.Request) {
	run, err := s.childOf(r.PathValue("id"), core.KindCapability)
	if err != nil {
		notFoundOr(w, err)
		return
	}
	results, err := s.svc.Store.ListCapabilityResults(run.ID)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	writeJSON(w, 200, map[string]any{"run": run, "matrix": capability.Summarize(results, nil)})
}

func (s *Server) runQuality(w http.ResponseWriter, r *http.Request) {
	run, err := s.childOf(r.PathValue("id"), core.KindQuality)
	if err != nil {
		notFoundOr(w, err)
		return
	}
	results, err := s.svc.Store.ListQualityResults(run.ID)
	if err != nil {
		writeErr(w, 500, err)
		return
	}
	if results == nil {
		results = []core.QualityResult{}
	}
	writeJSON(w, 200, map[string]any{"run": run, "scores": quality.Summarize(results, nil), "results": results})
}

func (s *Server) runReport(w http.ResponseWriter, r *http.Request) {
	lang := r.URL.Query().Get("lang")
	md, err := s.svc.Report(r.PathValue("id"), lang)
	if err != nil {
		notFoundOr(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/markdown; charset=utf-8")
	if r.URL.Query().Get("download") != "" {
		w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="llm-speedtest-%s.md"`, r.PathValue("id")))
	}
	_, _ = io.WriteString(w, md) //nolint:gosec // Markdown served as text/markdown (download disposition), never rendered as HTML here
}

func (s *Server) runExportJSON(w http.ResponseWriter, r *http.Request) {
	raw, err := s.svc.ExportJSON(r.PathValue("id"))
	if err != nil {
		notFoundOr(w, err)
		return
	}
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="llm-speedtest-%s.json"`, r.PathValue("id")))
	_, _ = w.Write(raw)
}

func (s *Server) runExportCSV(w http.ResponseWriter, r *http.Request) {
	raw, err := s.svc.ExportCSV(r.PathValue("id"))
	if err != nil {
		notFoundOr(w, err)
		return
	}
	w.Header().Set("Content-Type", "text/csv; charset=utf-8")
	w.Header().Set("Content-Disposition", fmt.Sprintf(`attachment; filename="llm-speedtest-%s.csv"`, r.PathValue("id")))
	_, _ = w.Write(raw)
}

// ---------------------------------------------------------------- compare

type modelDiff struct {
	ID       string  `json:"id"`
	Provider string  `json:"provider"`
	Family   string  `json:"family"`
	StatusA  string  `json:"status_a"`
	StatusB  string  `json:"status_b"`
	TierA    string  `json:"tier_a,omitempty"`
	TierB    string  `json:"tier_b,omitempty"`
	MedA     float64 `json:"median_a"`
	MedB     float64 `json:"median_b"`
	DeltaMed float64 `json:"delta_median"`
	ReasonA  string  `json:"reason_a,omitempty"`
	ReasonB  string  `json:"reason_b,omitempty"`
	Change   string  `json:"change"` // improved | regressed | same | new | gone
}

type familyDiff struct {
	Key     string  `json:"key"`
	Display string  `json:"display"`
	ScoreA  float64 `json:"score_a"`
	ScoreB  float64 `json:"score_b"`
	RankA   int     `json:"rank_a"`
	RankB   int     `json:"rank_b"`
	MedA    float64 `json:"median_a"`
	MedB    float64 `json:"median_b"`
	OKA     int     `json:"ok_a"`
	OKB     int     `json:"ok_b"`
}

func (s *Server) compare(w http.ResponseWriter, r *http.Request) {
	a, b := r.URL.Query().Get("a"), r.URL.Query().Get("b")
	if a == "" || b == "" {
		writeErr(w, 400, errors.New("a and b are required"))
		return
	}
	ra, runA, err := s.scanResult(a)
	if err != nil {
		notFoundOr(w, err)
		return
	}
	rb, runB, err := s.scanResult(b)
	if err != nil {
		notFoundOr(w, err)
		return
	}
	var models []modelDiff
	seen := map[string]bool{}
	for _, ma := range ra.Models {
		seen[ma.ID] = true
		d := modelDiff{ID: ma.ID, Provider: ma.ProviderName, Family: ma.FamilyCanon, StatusA: ma.Status, TierA: ma.Tier, MedA: ma.Med, ReasonA: ma.Reason}
		if mb := rb.Model(ma.ID); mb != nil {
			d.StatusB, d.TierB, d.MedB, d.ReasonB = mb.Status, mb.Tier, mb.Med, mb.Reason
			d.DeltaMed = mb.Med - ma.Med
			switch {
			case ma.Status != "ok" && mb.Status == "ok":
				d.Change = "improved"
			case ma.Status == "ok" && mb.Status != "ok":
				d.Change = "regressed"
			case ma.Status == "ok" && mb.Status == "ok" && ma.Med > 0 && mb.Med > ma.Med*1.25:
				d.Change = "regressed"
			case ma.Status == "ok" && mb.Status == "ok" && mb.Med > 0 && mb.Med < ma.Med*0.8:
				d.Change = "improved"
			default:
				d.Change = "same"
			}
		} else {
			d.Change = "gone"
		}
		models = append(models, d)
	}
	for _, mb := range rb.Models {
		if !seen[mb.ID] {
			models = append(models, modelDiff{ID: mb.ID, Provider: mb.ProviderName, Family: mb.FamilyCanon, StatusB: mb.Status, TierB: mb.Tier, MedB: mb.Med, ReasonB: mb.Reason, Change: "new"})
		}
	}
	sort.Slice(models, func(i, j int) bool {
		rank := map[string]int{"regressed": 0, "improved": 1, "new": 2, "gone": 3, "same": 4}
		if rank[models[i].Change] != rank[models[j].Change] {
			return rank[models[i].Change] < rank[models[j].Change]
		}
		return models[i].ID < models[j].ID
	})
	var fams []familyDiff
	for _, fa := range ra.Families {
		d := familyDiff{Key: fa.Key, Display: fa.Display, ScoreA: fa.Score, RankA: fa.RankComposite, MedA: fa.FMed, OKA: fa.OK}
		if fb := rb.Family(fa.Key); fb != nil {
			d.ScoreB, d.RankB, d.MedB, d.OKB = fb.Score, fb.RankComposite, fb.FMed, fb.OK
		}
		fams = append(fams, d)
	}
	for _, fb := range rb.Families {
		if ra.Family(fb.Key) == nil {
			fams = append(fams, familyDiff{Key: fb.Key, Display: fb.Display, ScoreB: fb.Score, RankB: fb.RankComposite, MedB: fb.FMed, OKB: fb.OK})
		}
	}
	counts := map[string]int{}
	for _, m := range models {
		counts[m.Change]++
	}
	if models == nil {
		models = []modelDiff{}
	}
	if fams == nil {
		fams = []familyDiff{}
	}
	writeJSON(w, 200, map[string]any{"a": runA, "b": runB, "summary_a": ra.Summary, "summary_b": rb.Summary, "counts": counts, "models": models, "families": fams})
}

// ---------------------------------------------------------------- static UI

const placeholder = `<!doctype html><html><head><meta charset="utf-8"><title>LLM Speedtest</title>
<style>body{font-family:system-ui,sans-serif;background:#0B0F19;color:#E3E6EC;display:flex;align-items:center;justify-content:center;height:100vh;margin:0}
main{max-width:560px;padding:32px;border-radius:20px;background:#121826}code{background:#1A2234;padding:2px 6px;border-radius:6px}</style></head>
<body><main><h1>LLM Speedtest API is running</h1><p>The web UI was not embedded in this build. Build it with <code>make web</code> and rebuild the binary, or use the HTTP API under <code>/api/</code>.</p></main></body></html>`

func (s *Server) static(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/api/") {
		writeErr(w, 404, errors.New("not found"))
		return
	}
	if !web.Built() {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = io.WriteString(w, placeholder)
		return
	}
	dist := web.Dist()
	p := strings.TrimPrefix(path.Clean(r.URL.Path), "/")
	if p == "" {
		p = "index.html"
	}
	if f, err := dist.Open(p); err == nil {
		st, _ := f.Stat()
		_ = f.Close()
		if st != nil && !st.IsDir() {
			if strings.HasPrefix(p, "assets/") {
				w.Header().Set("Cache-Control", "public, max-age=31536000, immutable")
			} else {
				w.Header().Set("Cache-Control", "no-cache")
			}
			http.ServeFileFS(w, r, dist, p) //nolint:gosec // p is path.Clean-ed and resolved inside the embedded read-only FS
			return
		}
	}
	// SPA fallback
	w.Header().Set("Cache-Control", "no-cache")
	http.ServeFileFS(w, r, dist, "index.html")
}

// ListenAndServe runs the server until ctx is cancelled.
func ListenAndServe(ctx context.Context, addr string, h http.Handler) error {
	srv := &http.Server{Addr: addr, Handler: h, ReadHeaderTimeout: 10 * time.Second}
	errCh := make(chan error, 1)
	go func() { errCh <- srv.ListenAndServe() }()
	select {
	case <-ctx.Done():
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		return srv.Shutdown(shutdownCtx)
	case err := <-errCh:
		if errors.Is(err, http.ErrServerClosed) {
			return nil
		}
		return err
	}
}

var _ fs.FS = (fs.FS)(nil)
