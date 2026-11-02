// Package mcpserver exposes llm-speedtest to AI agents through the Model
// Context Protocol (stdio transport).
package mcpserver

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"time"

	"github.com/modelcontextprotocol/go-sdk/mcp"

	"github.com/AnnaofArendelle/llm-speedtest/internal/app"
	"github.com/AnnaofArendelle/llm-speedtest/internal/capability"
	"github.com/AnnaofArendelle/llm-speedtest/internal/classify"
	"github.com/AnnaofArendelle/llm-speedtest/internal/config"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/latency"
	"github.com/AnnaofArendelle/llm-speedtest/internal/quality"
	"github.com/AnnaofArendelle/llm-speedtest/internal/skill"
	"github.com/AnnaofArendelle/llm-speedtest/internal/store"
)

// Run serves MCP over stdio until ctx is cancelled.
func Run(ctx context.Context, svc *app.Service) error {
	return NewServer(svc).Run(ctx, &mcp.StdioTransport{})
}

// EndpointArgs selects a stored endpoint or describes one inline.
type EndpointArgs struct {
	EndpointID string `json:"endpoint_id,omitempty" jsonschema:"id of a saved endpoint"`
	BaseURL    string `json:"base_url,omitempty" jsonschema:"inline endpoint base URL such as https://host/v1"`
	APIKey     string `json:"api_key,omitempty" jsonschema:"API key for an inline endpoint (optional)"`
	Dialect    string `json:"dialect,omitempty" jsonschema:"gateway dialect; auto-detected when empty"`
	Name       string `json:"name,omitempty" jsonschema:"display name for an inline endpoint"`
}

func (e EndpointArgs) inline() *config.Endpoint {
	if e.BaseURL == "" {
		return nil
	}
	name := e.Name
	if name == "" {
		name = e.BaseURL
	}
	return &config.Endpoint{Name: name, BaseURL: e.BaseURL, APIKey: e.APIKey, Dialect: e.Dialect}
}

type filterArgs struct {
	IncludeOwners []string `json:"include_owners,omitempty" jsonschema:"only models whose owned_by is in this list"`
	ExcludeOwners []string `json:"exclude_owners,omitempty" jsonschema:"skip these owned_by values (default combo)"`
	IDRegex       string   `json:"id_regex,omitempty" jsonschema:"only model ids matching this regular expression"`
	Limit         int      `json:"limit,omitempty" jsonschema:"test at most this many models"`
	Models        []string `json:"models,omitempty" jsonschema:"explicit model ids"`
}

func (f filterArgs) filter() config.Filter {
	out := config.DefaultFilter()
	if len(f.ExcludeOwners) > 0 {
		out.ExcludeOwners = f.ExcludeOwners
	}
	out.IncludeOwners, out.IDRegex, out.Limit, out.Models = f.IncludeOwners, f.IDRegex, f.Limit, f.Models
	return out
}

func text(v any) *mcp.CallToolResult {
	raw, _ := json.MarshalIndent(v, "", " ")
	return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(raw)}}}
}

func fail(err error) (*mcp.CallToolResult, any, error) {
	return &mcp.CallToolResult{IsError: true, Content: []mcp.Content{&mcp.TextContent{Text: err.Error()}}}, nil, nil
}

// NewServer builds the MCP server with every tool, resource and prompt.
func NewServer(svc *app.Service) *mcp.Server {
	s := mcp.NewServer(&mcp.Implementation{Name: "llm-speedtest", Version: svc.Version}, nil)

	type empty struct{}
	mcp.AddTool(s, &mcp.Tool{Name: "list_endpoints", Description: "List saved endpoints (API keys are masked)."}, func(ctx context.Context, _ *mcp.CallToolRequest, _ empty) (*mcp.CallToolResult, any, error) {
		eps, err := svc.Store.ListEndpoints()
		if err != nil {
			return fail(err)
		}
		if eps == nil {
			eps = []*store.Endpoint{}
		}
		return text(map[string]any{"endpoints": eps}), nil, nil
	})

	type upsertArgs struct {
		ID      string `json:"id,omitempty" jsonschema:"existing endpoint id to update"`
		Name    string `json:"name" jsonschema:"display name"`
		BaseURL string `json:"base_url" jsonschema:"base URL such as https://host/v1"`
		APIKey  string `json:"api_key,omitempty" jsonschema:"API key (stored encrypted; omit to keep the existing key)"`
		Dialect string `json:"dialect,omitempty" jsonschema:"gateway dialect or empty for auto-detect"`
	}
	mcp.AddTool(s, &mcp.Tool{Name: "upsert_endpoint", Description: "Create or update a saved endpoint."}, func(ctx context.Context, _ *mcp.CallToolRequest, in upsertArgs) (*mcp.CallToolResult, any, error) {
		if in.BaseURL == "" {
			return fail(errors.New("base_url is required"))
		}
		ep := &store.Endpoint{ID: in.ID, Name: in.Name, BaseURL: in.BaseURL, APIKey: in.APIKey, Dialect: in.Dialect}
		if ep.Name == "" {
			ep.Name = ep.BaseURL
		}
		if err := svc.Store.SaveEndpoint(ep); err != nil {
			return fail(err)
		}
		out, _ := svc.Store.GetEndpoint(ep.ID)
		return text(out), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{Name: "check_endpoint", Description: "Check connectivity, authentication, model count and gateway dialect of an endpoint."}, func(ctx context.Context, _ *mcp.CallToolRequest, in EndpointArgs) (*mcp.CallToolResult, any, error) {
		ep, err := resolve(svc, in)
		if err != nil {
			return fail(err)
		}
		return text(svc.Check(ctx, ep)), nil, nil
	})

	type discoverArgs struct {
		EndpointArgs
		filterArgs
		MaxModels int `json:"max_models,omitempty" jsonschema:"number of models to include in the response (default 50)"`
	}
	mcp.AddTool(s, &mcp.Tool{Name: "discover_models", Description: "List the models an endpoint exposes with resolved providers and families."}, func(ctx context.Context, _ *mcp.CallToolRequest, in discoverArgs) (*mcp.CallToolResult, any, error) {
		ep, err := resolve(svc, in.EndpointArgs)
		if err != nil {
			return fail(err)
		}
		res, err := svc.Discover(ctx, ep, in.EndpointID, in.filter())
		if err != nil {
			return fail(err)
		}
		n := in.MaxModels
		if n <= 0 {
			n = 50
		}
		models := res.Models
		if len(models) > n {
			models = models[:n]
		}
		return text(map[string]any{"total": res.Total, "filtered": res.Filtered, "provider_field": res.Field, "provider_rationale": res.Rationale, "dialect": res.Dialect, "providers": res.Providers, "families": res.Families, "models": models}), nil, nil
	})

	type startArgs struct {
		EndpointArgs
		filterArgs
		Kind    string   `json:"kind" jsonschema:"suite | scan | latency | load | capability | quality"`
		Profile string   `json:"profile,omitempty" jsonschema:"quick | standard | deep (default standard)"`
		Models  []string `json:"models,omitempty" jsonschema:"explicit models for latency/capability/quality/load"`
		Confirm bool     `json:"confirm,omitempty" jsonschema:"must be true for load tests (they send sustained traffic)"`
		Note    string   `json:"note,omitempty"`
		Lang    string   `json:"lang,omitempty" jsonschema:"report language zh or en"`
	}
	mcp.AddTool(s, &mcp.Tool{Name: "start_run", Description: "Start a benchmark run in the background and return its id. Use wait_run to block until it finishes."}, func(ctx context.Context, _ *mcp.CallToolRequest, in startArgs) (*mcp.CallToolResult, any, error) {
		if in.Kind == "" {
			in.Kind = core.KindSuite
		}
		suite := config.DefaultSuite(in.Profile)
		suite.Filter = in.filter()
		if in.Lang != "" {
			suite.Lang = in.Lang
		}
		if in.Kind == core.KindLoad && len(in.Models) > 0 {
			suite.Specs.Load.Model = in.Models[0]
		}
		req := app.Request{Kind: in.Kind, EndpointID: in.EndpointID, Endpoint: in.inline(), Suite: suite, Models: in.Models, Note: in.Note, Confirm: in.Confirm}
		id, err := svc.Start(req)
		if err != nil {
			return fail(err)
		}
		return text(map[string]any{"run_id": id, "kind": in.Kind, "profile": suite.Profile}), nil, nil
	})

	type runArgs struct {
		RunID string `json:"run_id" jsonschema:"run id"`
	}
	mcp.AddTool(s, &mcp.Tool{Name: "get_run", Description: "Get the status, children and latest progress of a run."}, func(ctx context.Context, _ *mcp.CallToolRequest, in runArgs) (*mcp.CallToolResult, any, error) {
		run, err := svc.Store.GetRun(in.RunID)
		if err != nil {
			return fail(err)
		}
		kids, _ := svc.Store.ChildRuns(in.RunID)
		latest, _ := svc.Bus.Latest(in.RunID)
		return text(map[string]any{"run": run, "children": kids, "running": svc.IsRunning(in.RunID), "latest": latest}), nil, nil
	})

	type waitArgs struct {
		RunID    string `json:"run_id"`
		TimeoutS int    `json:"timeout_s,omitempty" jsonschema:"maximum seconds to wait (default 600)"`
	}
	mcp.AddTool(s, &mcp.Tool{Name: "wait_run", Description: "Wait until a run finishes (or the timeout elapses) and return it."}, func(ctx context.Context, _ *mcp.CallToolRequest, in waitArgs) (*mcp.CallToolResult, any, error) {
		timeout := time.Duration(in.TimeoutS) * time.Second
		if timeout <= 0 {
			timeout = 10 * time.Minute
		}
		deadline := time.Now().Add(timeout)
		for {
			run, err := svc.Store.GetRun(in.RunID)
			if err != nil {
				return fail(err)
			}
			if !svc.IsRunning(in.RunID) && run.Status != core.StatusPending && run.Status != core.StatusRunning && run.Status != core.StatusDiscovering {
				kids, _ := svc.Store.ChildRuns(in.RunID)
				return text(map[string]any{"run": run, "children": kids, "finished": true}), nil, nil
			}
			if time.Now().After(deadline) {
				latest, _ := svc.Bus.Latest(in.RunID)
				return text(map[string]any{"run": run, "finished": false, "latest": latest}), nil, nil
			}
			select {
			case <-ctx.Done():
				return fail(ctx.Err())
			case <-time.After(time.Second):
			}
		}
	})

	mcp.AddTool(s, &mcp.Tool{Name: "cancel_run", Description: "Cancel a running benchmark."}, func(ctx context.Context, _ *mcp.CallToolRequest, in runArgs) (*mcp.CallToolResult, any, error) {
		return text(map[string]any{"cancelled": svc.Cancel(in.RunID)}), nil, nil
	})

	type listArgs struct {
		Kind  string `json:"kind,omitempty"`
		Limit int    `json:"limit,omitempty"`
	}
	type skipArgs struct {
		RunID string `json:"run_id" jsonschema:"run id"`
		Model string `json:"model,omitempty" jsonschema:"model id to skip (omit with all=true)"`
		All   bool   `json:"all,omitempty" jsonschema:"skip every model currently in flight"`
	}
	mcp.AddTool(s, &mcp.Tool{Name: "skip_model", Description: "Skip a model while a run is in progress: cancels its in-flight request and leaves it out of later rounds and stages. Use all=true to skip everything currently in flight."}, func(ctx context.Context, _ *mcp.CallToolRequest, in skipArgs) (*mcp.CallToolResult, any, error) {
		ids, ok := svc.Skip(in.RunID, in.Model, in.All)
		if !ok {
			return fail(fmt.Errorf("run %s is not in progress", in.RunID))
		}
		return text(map[string]any{"skipped": ids, "inflight": svc.Inflight(in.RunID)}), nil, nil
	})
	mcp.AddTool(s, &mcp.Tool{Name: "list_runs", Description: "List stored runs, newest first."}, func(ctx context.Context, _ *mcp.CallToolRequest, in listArgs) (*mcp.CallToolResult, any, error) {
		runs, total, err := svc.Store.ListRuns(in.Kind, "", in.Limit, 0)
		if err != nil {
			return fail(err)
		}
		if runs == nil {
			runs = []*store.Run{}
		}
		return text(map[string]any{"total": total, "runs": runs}), nil, nil
	})

	type rankArgs struct {
		RunID string `json:"run_id"`
		Order string `json:"order,omitempty" jsonschema:"composite (default) or stability"`
		Limit int    `json:"limit,omitempty" jsonschema:"number of families (default 20)"`
	}
	mcp.AddTool(s, &mcp.Tool{Name: "get_ranking", Description: "Family ranking of a scan or suite run with summary numbers."}, func(ctx context.Context, _ *mcp.CallToolRequest, in rankArgs) (*mcp.CallToolResult, any, error) {
		res, _, err := svc.ScanResult(in.RunID)
		if err != nil {
			return fail(err)
		}
		fams := res.Families
		if in.Order == "stability" {
			fams = res.FamiliesByStability
		}
		n := in.Limit
		if n <= 0 {
			n = 20
		}
		if len(fams) > n {
			fams = fams[:n]
		}
		unavail := res.UnavailableFamilies
		if len(unavail) > n {
			unavail = unavail[:n]
		}
		return text(map[string]any{"summary": res.Summary, "provider_field": res.ProviderField, "families": fams, "unavailable_families": unavail}), nil, nil
	})

	type modelsArgs struct {
		RunID    string `json:"run_id"`
		Status   string `json:"status,omitempty" jsonschema:"ok | fail"`
		Provider string `json:"provider,omitempty"`
		Query    string `json:"q,omitempty" jsonschema:"substring filter on id, provider or family"`
		Page     int    `json:"page,omitempty"`
		Size     int    `json:"size,omitempty" jsonschema:"page size (default 50)"`
	}
	mcp.AddTool(s, &mcp.Tool{Name: "get_models", Description: "Models of a scan run ranked by stability tier then median latency, with filters and paging."}, func(ctx context.Context, _ *mcp.CallToolRequest, in modelsArgs) (*mcp.CallToolResult, any, error) {
		res, _, err := svc.ScanResult(in.RunID)
		if err != nil {
			return fail(err)
		}
		var list []any
		q := strings.ToLower(in.Query)
		for _, m := range res.Models {
			if in.Status != "" && m.Status != in.Status {
				continue
			}
			if in.Provider != "" && m.ProviderKey != in.Provider && m.ProviderName != in.Provider {
				continue
			}
			if q != "" && !strings.Contains(strings.ToLower(m.ID), q) && !strings.Contains(strings.ToLower(m.ProviderName), q) && !strings.Contains(strings.ToLower(m.FamilyCanon), q) {
				continue
			}
			list = append(list, m)
		}
		size := in.Size
		if size <= 0 {
			size = 50
		}
		page := in.Page
		if page <= 0 {
			page = 1
		}
		start := (page - 1) * size
		if start > len(list) {
			start = len(list)
		}
		end := start + size
		if end > len(list) {
			end = len(list)
		}
		if list == nil {
			list = []any{}
		}
		return text(map[string]any{"total": len(list), "page": page, "size": size, "models": list[start:end]}), nil, nil
	})

	type failuresArgs struct {
		RunID         string `json:"run_id"`
		LimitPerGroup int    `json:"limit_per_group,omitempty" jsonschema:"models listed per reason/provider group (default 5)"`
	}
	mcp.AddTool(s, &mcp.Tool{Name: "get_failures", Description: "Unavailable models grouped by reason and provider, with explanations and example messages."}, func(ctx context.Context, _ *mcp.CallToolRequest, in failuresArgs) (*mcp.CallToolResult, any, error) {
		res, _, err := svc.ScanResult(in.RunID)
		if err != nil {
			return fail(err)
		}
		limit := in.LimitPerGroup
		if limit <= 0 {
			limit = 5
		}
		type grp struct {
			Reason     string              `json:"reason"`
			Info       classify.ReasonInfo `json:"info"`
			Count      int                 `json:"count"`
			ByProvider map[string]int      `json:"by_provider"`
			Examples   []map[string]any    `json:"examples"`
		}
		groups := map[string]*grp{}
		for _, m := range res.Models {
			if m.Status == "ok" {
				continue
			}
			g := groups[m.Reason]
			if g == nil {
				g = &grp{Reason: m.Reason, Info: classify.Info(m.Reason), ByProvider: map[string]int{}}
				groups[m.Reason] = g
			}
			g.Count++
			g.ByProvider[m.ProviderName]++
			if len(g.Examples) < limit {
				g.Examples = append(g.Examples, map[string]any{"id": m.ID, "provider": m.ProviderName, "http": m.HTTP, "inner": m.Inner, "message": m.Message})
			}
		}
		var out []*grp
		for _, code := range classify.Order {
			if g, ok := groups[code]; ok {
				out = append(out, g)
			}
		}
		if out == nil {
			out = []*grp{}
		}
		return text(map[string]any{"groups": out}), nil, nil
	})

	mcp.AddTool(s, &mcp.Tool{Name: "get_providers", Description: "Provider table of a scan run with the evidence used to identify providers."}, func(ctx context.Context, _ *mcp.CallToolRequest, in runArgs) (*mcp.CallToolResult, any, error) {
		res, _, err := svc.ScanResult(in.RunID)
		if err != nil {
			return fail(err)
		}
		return text(map[string]any{"field": res.ProviderField, "rationale": res.ProviderRationale, "suggestions": res.Suggestions, "providers": res.Providers}), nil, nil
	})

	stage := func(name, kind, desc string) {
		mcp.AddTool(s, &mcp.Tool{Name: name, Description: desc}, func(ctx context.Context, _ *mcp.CallToolRequest, in runArgs) (*mcp.CallToolResult, any, error) {
			run, err := childOf(svc, in.RunID, kind)
			if err != nil {
				return fail(err)
			}
			switch kind {
			case core.KindLatency:
				samples, err := svc.Store.ListLatencySamples(run.ID)
				if err != nil {
					return fail(err)
				}
				return text(map[string]any{"run_id": run.ID, "stats": latency.Summarize(samples)}), nil, nil
			case core.KindLoad:
				return text(map[string]any{"run_id": run.ID, "summary": json.RawMessage(run.Summary)}), nil, nil
			case core.KindCapability:
				results, err := svc.Store.ListCapabilityResults(run.ID)
				if err != nil {
					return fail(err)
				}
				return text(map[string]any{"run_id": run.ID, "matrix": capability.Summarize(results, nil)}), nil, nil
			case core.KindQuality:
				results, err := svc.Store.ListQualityResults(run.ID)
				if err != nil {
					return fail(err)
				}
				return text(map[string]any{"run_id": run.ID, "scores": quality.Summarize(results, nil)}), nil, nil
			}
			return fail(errors.New("unknown stage"))
		})
	}
	stage("get_latency", core.KindLatency, "Latency lab statistics (TTFT, TPOT, tokens/s percentiles) of a run.")
	stage("get_load", core.KindLoad, "Load test levels (throughput, error rate, latency percentiles, goodput) of a run.")
	stage("get_capabilities", core.KindCapability, "Capability matrix (tools, JSON mode, vision, long context, ...) of a run.")
	stage("get_quality", core.KindQuality, "Quality spot-check scores of a run.")

	type exportArgs struct {
		RunID   string `json:"run_id"`
		Format  string `json:"format,omitempty" jsonschema:"markdown (default) | json | csv"`
		Lang    string `json:"lang,omitempty" jsonschema:"zh (default) | en"`
		OutPath string `json:"out_path,omitempty" jsonschema:"write to this file instead of returning the content"`
	}
	mcp.AddTool(s, &mcp.Tool{Name: "export_report", Description: "Render the report of a run as Markdown, JSON or CSV; returns the content or writes it to out_path."}, func(ctx context.Context, _ *mcp.CallToolRequest, in exportArgs) (*mcp.CallToolResult, any, error) {
		var raw []byte
		var err error
		switch in.Format {
		case "", "markdown", "md":
			var md string
			md, err = svc.Report(in.RunID, in.Lang)
			raw = []byte(md)
		case "json":
			raw, err = svc.ExportJSON(in.RunID)
		case "csv":
			raw, err = svc.ExportCSV(in.RunID)
		default:
			err = fmt.Errorf("unknown format %q", in.Format)
		}
		if err != nil {
			return fail(err)
		}
		if in.OutPath != "" {
			if err := os.WriteFile(in.OutPath, raw, 0o644); err != nil { //nolint:gosec // reports are meant to be shared
				return fail(err)
			}
			return text(map[string]any{"written": in.OutPath, "bytes": len(raw)}), nil, nil
		}
		return &mcp.CallToolResult{Content: []mcp.Content{&mcp.TextContent{Text: string(raw)}}}, nil, nil
	})

	type compareArgs struct {
		A string `json:"a"`
		B string `json:"b"`
	}
	mcp.AddTool(s, &mcp.Tool{Name: "compare_runs", Description: "Compare two scan runs: models whose availability changed and headline deltas."}, func(ctx context.Context, _ *mcp.CallToolRequest, in compareArgs) (*mcp.CallToolResult, any, error) {
		ra, _, err := svc.ScanResult(in.A)
		if err != nil {
			return fail(err)
		}
		rb, _, err := svc.ScanResult(in.B)
		if err != nil {
			return fail(err)
		}
		var changed []map[string]any
		improved, regressed := 0, 0
		for _, ma := range ra.Models {
			mb := rb.Model(ma.ID)
			if mb == nil || ma.Status == mb.Status {
				continue
			}
			if mb.Status == "ok" {
				improved++
			} else {
				regressed++
			}
			changed = append(changed, map[string]any{"id": ma.ID, "provider": ma.ProviderName, "status_a": ma.Status, "status_b": mb.Status, "median_a": ma.Med, "median_b": mb.Med, "reason_b": mb.Reason})
		}
		if changed == nil {
			changed = []map[string]any{}
		}
		return text(map[string]any{"summary_a": ra.Summary, "summary_b": rb.Summary, "improved": improved, "regressed": regressed, "changed": changed}), nil, nil
	})

	// resources
	s.AddResource(&mcp.Resource{URI: "speedtest://metrics-glossary", Name: "metrics-glossary", MIMEType: "text/markdown", Description: "Definitions of every metric and ranking rule."}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		raw, err := readSkillFile("references/metrics.md")
		if err != nil {
			return nil, err
		}
		return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{URI: req.Params.URI, MIMEType: "text/markdown", Text: raw}}}, nil
	})
	s.AddResource(&mcp.Resource{URI: "speedtest://skill", Name: "skill", MIMEType: "text/markdown", Description: "The llm-speedtest agent skill (workflow and interpretation guide)."}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{URI: req.Params.URI, MIMEType: "text/markdown", Text: skill.Manifest()}}}, nil
	})
	s.AddResourceTemplate(&mcp.ResourceTemplate{URITemplate: "speedtest://runs/{id}/summary", Name: "run-summary", MIMEType: "application/json", Description: "Summary document of a run."}, func(ctx context.Context, req *mcp.ReadResourceRequest) (*mcp.ReadResourceResult, error) {
		id := strings.TrimSuffix(strings.TrimPrefix(req.Params.URI, "speedtest://runs/"), "/summary")
		run, err := svc.Store.GetRun(id)
		if err != nil {
			return nil, err
		}
		return &mcp.ReadResourceResult{Contents: []*mcp.ResourceContents{{URI: req.Params.URI, MIMEType: "application/json", Text: string(run.Summary)}}}, nil
	})

	// prompt
	s.AddPrompt(&mcp.Prompt{Name: "summarize_run", Description: "Summarise a benchmark run for the user.", Arguments: []*mcp.PromptArgument{{Name: "run_id", Description: "run id", Required: true}}}, func(ctx context.Context, req *mcp.GetPromptRequest) (*mcp.GetPromptResult, error) {
		id := req.Params.Arguments["run_id"]
		tmpl, _ := readSkillFile("assets/report-template.md")
		msg := fmt.Sprintf("Call get_ranking, get_failures and (if present) get_latency/get_capabilities/get_quality for run %s, then write a concise summary for the user following this template:\n\n%s", id, tmpl)
		return &mcp.GetPromptResult{Description: "Summarise run " + id, Messages: []*mcp.PromptMessage{{Role: "user", Content: &mcp.TextContent{Text: msg}}}}, nil
	})
	return s
}

func readSkillFile(name string) (string, error) {
	f, err := skill.Files().Open(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	var sb strings.Builder
	buf := make([]byte, 32*1024)
	for {
		n, err := f.Read(buf)
		sb.Write(buf[:n])
		if err != nil {
			break
		}
	}
	return sb.String(), nil
}

func resolve(svc *app.Service, in EndpointArgs) (config.Endpoint, error) {
	if ep := in.inline(); ep != nil {
		return *ep, nil
	}
	if in.EndpointID == "" {
		return config.Endpoint{}, errors.New("endpoint_id or base_url is required")
	}
	st, err := svc.Store.GetEndpointWithKey(in.EndpointID)
	if err != nil {
		return config.Endpoint{}, err
	}
	return config.Endpoint{Name: st.Name, BaseURL: st.BaseURL, APIKey: st.APIKey, Dialect: st.Dialect, Headers: st.Headers, Insecure: st.Insecure, Proxy: st.Proxy}, nil
}

func childOf(svc *app.Service, id, kind string) (*store.Run, error) {
	run, err := svc.Store.GetRun(id)
	if err != nil {
		return nil, err
	}
	if run.Kind == kind {
		return run, nil
	}
	if run.Kind == core.KindSuite {
		kids, _ := svc.Store.ChildRuns(id)
		for _, k := range kids {
			if k.Kind == kind {
				return k, nil
			}
		}
	}
	return nil, fmt.Errorf("run %s has no %s stage", id, kind)
}
