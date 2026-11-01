// Package cli implements the command line interface.
package cli

import (
	"context"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/signal"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/app"
	"github.com/AnnaofArendelle/llm-speedtest/internal/config"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/demo"
	"github.com/AnnaofArendelle/llm-speedtest/internal/load"
	"github.com/AnnaofArendelle/llm-speedtest/internal/mcpserver"
	"github.com/AnnaofArendelle/llm-speedtest/internal/secrets"
	"github.com/AnnaofArendelle/llm-speedtest/internal/server"
	"github.com/AnnaofArendelle/llm-speedtest/internal/skill"
	"github.com/AnnaofArendelle/llm-speedtest/internal/store"
	"github.com/AnnaofArendelle/llm-speedtest/internal/version"
)

// Exit codes.
const (
	ExitOK        = 0
	ExitError     = 1
	ExitUsage     = 2
	ExitEndpoint  = 3
	ExitCancelled = 4
	ExitThreshold = 5
)

// Main runs the CLI and returns the exit code.
func Main(args []string, stdout, stderr io.Writer) int {
	if len(args) == 0 || args[0] == "-h" || args[0] == "--help" || args[0] == "help" {
		usage(stdout)
		return ExitOK
	}
	cmd, rest := args[0], args[1:]
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()
	c := &cli{out: stdout, err: stderr, ctx: ctx}
	switch cmd {
	case "version":
		return c.version(rest)
	case "serve":
		return c.serve(rest)
	case "suite", "scan", "latency", "load", "cap", "capability", "quality":
		return c.run(cmd, rest)
	case "discover":
		return c.discover(rest)
	case "check":
		return c.check(rest)
	case "report":
		return c.report(rest)
	case "compare":
		return c.compare(rest)
	case "runs":
		return c.runs(rest)
	case "config":
		return c.configCmd(rest)
	case "mcp":
		return c.mcp(rest)
	case "skill":
		return c.skill(rest)
	}
	fmt.Fprintf(stderr, "unknown command %q\n\n", cmd)
	usage(stderr)
	return ExitUsage
}

func usage(w io.Writer) {
	fmt.Fprint(w, `llm-speedtest — availability, latency, load, capability and quality benchmarks for any OpenAI-compatible API

Usage:
  llm-speedtest <command> [flags]

Commands:
  serve      Start the web dashboard and HTTP API (default 127.0.0.1:8787)
  suite      One-shot full benchmark: scan → latency → capability → quality [→ load]
  scan       Availability scan of every model with multi-round probing and ranking
  latency    Streaming latency lab (TTFT, tokens/s, ...) for selected models
  load       Concurrency / request-rate load test for one model (requires --confirm)
  cap        Capability probes (tools, JSON mode, vision, long context, ...)
  quality    Deterministic quality spot-check
  discover   List the models an endpoint exposes with resolved providers/families
  check      Check connectivity, authentication and dialect of an endpoint
  report     Render the report of a stored run (markdown, json, csv)
  compare    Compare two stored scan runs
  runs       List stored runs
  config     Print an example configuration file
  mcp        Run the Model Context Protocol server over stdio (for AI agents)
  skill      Install or export the AI skill package
  version    Print version information

Run "llm-speedtest <command> -h" for the flags of a command.
Environment: LLM_SPEEDTEST_KEY (API key), LLM_SPEEDTEST_DATA (data directory).
`)
}

type cli struct {
	out, err io.Writer
	ctx      context.Context
}

func (c *cli) openStore(dataDir string) (*store.Store, error) {
	if dataDir == "" {
		dataDir = os.Getenv("LLM_SPEEDTEST_DATA")
	}
	if dataDir == "" {
		dataDir = "./data"
	}
	codec, err := secrets.Load(filepath.Join(dataDir, ".secret"))
	if err != nil {
		return nil, err
	}
	return store.Open(filepath.Join(dataDir, "llm-speedtest.db"), codec)
}

func (c *cli) fail(code int, err error) int {
	fmt.Fprintf(c.err, "error: %v\n", err)
	return code
}

type multiFlag []string

func (m *multiFlag) String() string     { return strings.Join(*m, ",") }
func (m *multiFlag) Set(v string) error { *m = append(*m, v); return nil }

// ---------------------------------------------------------------- version

func (c *cli) version(args []string) int {
	fs := flag.NewFlagSet("version", flag.ContinueOnError)
	fs.SetOutput(c.err)
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}
	if *jsonOut {
		_ = json.NewEncoder(c.out).Encode(version.Get())
		return ExitOK
	}
	fmt.Fprintln(c.out, version.String())
	return ExitOK
}

// ---------------------------------------------------------------- serve

func (c *cli) serve(args []string) int {
	fs := flag.NewFlagSet("serve", flag.ContinueOnError)
	fs.SetOutput(c.err)
	listen := fs.String("listen", "127.0.0.1:8787", "address to listen on")
	dataDir := fs.String("data-dir", "", "data directory (default ./data or $LLM_SPEEDTEST_DATA)")
	token := fs.String("access-token", "", "require this bearer token for the API (recommended when exposing beyond localhost)")
	demoMode := fs.Bool("demo", false, "seed the database with demo data (reference scan) on start")
	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}
	st, err := c.openStore(*dataDir)
	if err != nil {
		return c.fail(ExitError, err)
	}
	defer st.Close()
	if n, err := st.MarkInterrupted(); err == nil && n > 0 {
		fmt.Fprintf(c.err, "marked %d unfinished run(s) as interrupted\n", n)
	}
	svc := app.New(st, version.Version)
	if *demoMode {
		if id, err := demo.Seed(st); err != nil {
			fmt.Fprintf(c.err, "demo seed failed: %v\n", err)
		} else if id != "" {
			fmt.Fprintf(c.err, "demo data seeded (run %s)\n", id)
		}
	}
	srv := server.New(svc, server.Options{AccessToken: *token})
	fmt.Fprintf(c.err, "llm-speedtest %s listening on http://%s\n", version.Version, *listen)
	if err := server.ListenAndServe(c.ctx, *listen, srv.Handler()); err != nil {
		return c.fail(ExitError, err)
	}
	return ExitOK
}

// ---------------------------------------------------------------- runs (suite/scan/...)

type runFlags struct {
	url, key, name, dialect, proxy string
	headers                        multiFlag
	insecure                       bool
	profile, configPath, lang      string
	out, jsonOut, csvOut, note     string
	dataDir                        string
	includeOwners, excludeOwners   multiFlag
	idRegex, excludeRegex          string
	limit                          int
	models                         multiFlag
	failUnder                      float64
	quiet, verbose                 bool
	// latency
	samples, warmup, inputTokens, outputTokens int
	// load
	mode, levels      string
	rate              float64
	duration          time.Duration
	requests          int
	noStream, confirm bool
	// suite toggles
	noLatency, noCapability, noQuality, withLoad bool
	loadModel                                    string
	top                                          int
	trials                                       int
	longContext                                  string
	noStore                                      bool
}

func (c *cli) run(kind string, args []string) int {
	if kind == "cap" {
		kind = core.KindCapability
	}
	fs := flag.NewFlagSet(kind, flag.ContinueOnError)
	fs.SetOutput(c.err)
	var f runFlags
	fs.StringVar(&f.url, "url", "", "endpoint base URL, e.g. https://host/v1 (required unless --config)")
	fs.StringVar(&f.key, "key", "", "API key (or $LLM_SPEEDTEST_KEY)")
	fs.StringVar(&f.name, "name", "", "endpoint display name")
	fs.StringVar(&f.dialect, "dialect", "", "gateway dialect (auto-detect when empty)")
	fs.Var(&f.headers, "header", "extra request header k=v (repeatable)")
	fs.BoolVar(&f.insecure, "insecure", false, "skip TLS verification")
	fs.StringVar(&f.proxy, "proxy", "", "proxy URL (defaults to environment)")
	fs.StringVar(&f.profile, "profile", "standard", "quick | standard | deep")
	fs.StringVar(&f.configPath, "config", "", "YAML config file (see `llm-speedtest config`)")
	fs.StringVar(&f.lang, "lang", "zh", "report language: zh | en")
	fs.StringVar(&f.out, "out", "", "write the Markdown report to this file")
	fs.StringVar(&f.jsonOut, "json-out", "", "write the JSON export to this file")
	fs.StringVar(&f.csvOut, "csv-out", "", "write the CSV export to this file")
	fs.StringVar(&f.note, "note", "", "note stored with the run")
	fs.StringVar(&f.dataDir, "data-dir", "", "data directory for the results database")
	fs.Var(&f.includeOwners, "include-owner", "only models with this owned_by (repeatable)")
	fs.Var(&f.excludeOwners, "exclude-owner", "skip models with this owned_by (repeatable, default combo)")
	fs.StringVar(&f.idRegex, "id-regex", "", "only model ids matching this regex")
	fs.StringVar(&f.excludeRegex, "exclude-regex", "", "skip model ids matching this regex")
	fs.IntVar(&f.limit, "limit", 0, "test at most N models")
	fs.Var(&f.models, "model", "explicit model id (repeatable)")
	fs.Float64Var(&f.failUnder, "fail-under", 0, "exit 5 when the available share is below this ratio (0..1)")
	fs.BoolVar(&f.quiet, "quiet", false, "no progress output")
	fs.BoolVar(&f.verbose, "verbose", false, "print every attempt")
	fs.IntVar(&f.samples, "samples", 0, "latency samples per model")
	fs.IntVar(&f.warmup, "warmup", -1, "latency warm-up samples per model")
	fs.IntVar(&f.inputTokens, "input-tokens", 0, "prompt size in tokens (latency/load)")
	fs.IntVar(&f.outputTokens, "output-tokens", 0, "max_tokens (latency/load)")
	fs.StringVar(&f.mode, "mode", "", "load mode: concurrency | rate")
	fs.StringVar(&f.levels, "levels", "", "load concurrency levels, e.g. 1,2,4,8")
	fs.Float64Var(&f.rate, "rate", 0, "load requests per second (rate mode)")
	fs.DurationVar(&f.duration, "duration", 0, "load duration per level")
	fs.IntVar(&f.requests, "requests", 0, "load request cap per level")
	fs.BoolVar(&f.noStream, "no-stream", false, "load: use non-streaming requests")
	fs.BoolVar(&f.confirm, "confirm", false, "acknowledge that a load test sends sustained traffic")
	fs.BoolVar(&f.noLatency, "no-latency", false, "suite: skip the latency lab")
	fs.BoolVar(&f.noCapability, "no-capability", false, "suite: skip capability probes")
	fs.BoolVar(&f.noQuality, "no-quality", false, "suite: skip the quality spot-check")
	fs.BoolVar(&f.withLoad, "load", false, "suite: also run a load test on the best model (implies --confirm)")
	fs.StringVar(&f.loadModel, "load-model", "", "suite: model for the load test")
	fs.IntVar(&f.top, "top", 0, "suite: number of top models for latency/capability/quality stages")
	fs.IntVar(&f.trials, "trials", 0, "quality trials per question")
	fs.StringVar(&f.longContext, "long-context", "", "capability: context tiers in tokens, e.g. 8000,32000,128000")
	fs.BoolVar(&f.noStore, "no-store", false, "use a temporary database (results are not kept)")
	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}
	ep, suite, err := c.buildRequest(&f)
	if err != nil {
		return c.fail(ExitUsage, err)
	}
	if kind == core.KindLoad && !f.confirm {
		return c.fail(ExitUsage, app.ErrConfirmRequired)
	}
	if f.withLoad {
		suite.Load = true
		suite.LoadModel = f.loadModel
	}
	if kind != core.KindScan && kind != core.KindSuite && len(f.models) == 0 {
		return c.fail(ExitUsage, fmt.Errorf("%s needs at least one --model", kind))
	}
	dataDir := f.dataDir
	if f.noStore {
		tmp, err := os.MkdirTemp("", "llm-speedtest-*")
		if err != nil {
			return c.fail(ExitError, err)
		}
		defer os.RemoveAll(tmp)
		dataDir = tmp
	}
	st, err := c.openStore(dataDir)
	if err != nil {
		return c.fail(ExitError, err)
	}
	defer st.Close()
	svc := app.New(st, version.Version)
	req := app.Request{Kind: kind, Endpoint: &ep, Suite: suite, Models: f.models, Note: f.note, Confirm: f.confirm || f.withLoad}
	if kind == core.KindLoad && len(f.models) > 0 {
		req.Suite.Specs.Load.Model = f.models[0]
	}
	start := time.Now()
	var progressDone chan struct{}
	id, err := svc.RunSyncWithID(c.ctx, req, func(runID string) {
		if !f.quiet {
			fmt.Fprintf(c.err, "run %s started (%s, profile %s)\n", runID, kind, suite.Profile)
			progressDone = make(chan struct{})
			go func() {
				defer close(progressDone)
				c.progress(svc, runID, kind, f.verbose)
			}()
		}
	})
	if progressDone != nil {
		// Drop the subscription and wait for the progress printer so that the
		// summary below never interleaves with (or races against) its output.
		svc.Bus.Forget(id)
		select {
		case <-progressDone:
		case <-time.After(2 * time.Second):
		}
	}
	if err != nil {
		if errors.Is(err, context.Canceled) {
			fmt.Fprintf(c.err, "cancelled (run %s)\n", id)
			return ExitCancelled
		}
		code := ExitError
		if strings.Contains(err.Error(), "list models") || strings.Contains(err.Error(), "no models matched") {
			code = ExitEndpoint
		}
		return c.fail(code, err)
	}
	if !f.quiet {
		fmt.Fprintf(c.err, "run %s finished in %s\n", id, time.Since(start).Round(time.Second))
	}
	return c.finish(svc, id, kind, &f)
}

func (c *cli) buildRequest(f *runFlags) (config.Endpoint, config.Suite, error) {
	var ep config.Endpoint
	suite := config.DefaultSuite(f.profile)
	if f.configPath != "" {
		file, err := config.LoadFile(f.configPath)
		if err != nil {
			return ep, suite, err
		}
		ep, suite = file.Endpoint, file.Suite
	}
	if f.url != "" {
		ep.BaseURL = f.url
	}
	if ep.BaseURL == "" {
		return ep, suite, errors.New("--url is required (or --config with endpoint.base_url)")
	}
	if f.key != "" {
		ep.APIKey = f.key
	}
	if ep.APIKey == "" {
		ep.APIKey = os.Getenv("LLM_SPEEDTEST_KEY")
	}
	if f.name != "" {
		ep.Name = f.name
	}
	if ep.Name == "" {
		ep.Name = ep.BaseURL
	}
	if f.dialect != "" {
		ep.Dialect = f.dialect
	}
	if f.insecure {
		ep.Insecure = true
	}
	if f.proxy != "" {
		ep.Proxy = f.proxy
	}
	for _, h := range f.headers {
		k, v, ok := strings.Cut(h, "=")
		if !ok {
			return ep, suite, fmt.Errorf("bad --header %q, want k=v", h)
		}
		if ep.Headers == nil {
			ep.Headers = map[string]string{}
		}
		ep.Headers[strings.TrimSpace(k)] = strings.TrimSpace(v)
	}
	if f.configPath == "" || f.profile != "standard" {
		suite.Profile = f.profile
		if f.configPath == "" {
			suite = config.DefaultSuite(f.profile)
		}
	}
	if f.lang != "" {
		suite.Lang = f.lang
	}
	if len(f.includeOwners) > 0 {
		suite.Filter.IncludeOwners = f.includeOwners
	}
	if len(f.excludeOwners) > 0 {
		suite.Filter.ExcludeOwners = f.excludeOwners
	}
	if f.idRegex != "" {
		suite.Filter.IDRegex = f.idRegex
	}
	if f.excludeRegex != "" {
		suite.Filter.Exclude = f.excludeRegex
	}
	if f.limit > 0 {
		suite.Filter.Limit = f.limit
	}
	if len(f.models) > 0 {
		suite.Filter.Models = f.models
	}
	if f.samples > 0 {
		suite.Specs.Latency.Samples = f.samples
	}
	if f.warmup >= 0 {
		suite.Specs.Latency.Warmup = f.warmup
	}
	if f.inputTokens > 0 {
		suite.Specs.Latency.InputTokens = f.inputTokens
		suite.Specs.Load.InputTokens = f.inputTokens
	}
	if f.outputTokens > 0 {
		suite.Specs.Latency.OutputTokens = f.outputTokens
		suite.Specs.Load.OutputTokens = f.outputTokens
	}
	if f.mode != "" {
		suite.Specs.Load.Mode = f.mode
	}
	if f.levels != "" {
		var lv []int
		for _, p := range strings.Split(f.levels, ",") {
			n, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil || n <= 0 {
				return ep, suite, fmt.Errorf("bad --levels %q", f.levels)
			}
			lv = append(lv, n)
		}
		suite.Specs.Load.Levels = lv
	}
	if f.rate > 0 {
		suite.Specs.Load.Rate = f.rate
		if f.mode == "" {
			suite.Specs.Load.Mode = load.ModeRate
		}
	}
	if f.duration > 0 {
		suite.Specs.Load.Duration = f.duration
	}
	if f.requests > 0 {
		suite.Specs.Load.Requests = f.requests
	}
	if f.noStream {
		suite.Specs.Load.Stream = false
	}
	if f.noLatency {
		suite.Latency = false
	}
	if f.noCapability {
		suite.Capability = false
	}
	if f.noQuality {
		suite.Quality = false
	}
	if f.top > 0 {
		suite.LatencyTop, suite.CapabilityTop, suite.QualityTop = f.top, f.top, f.top
	}
	if f.trials > 0 {
		suite.Specs.Quality.Trials = f.trials
	}
	if f.longContext != "" {
		var tiers []int
		for _, p := range strings.Split(f.longContext, ",") {
			n, err := strconv.Atoi(strings.TrimSpace(p))
			if err != nil || n <= 0 {
				return ep, suite, fmt.Errorf("bad --long-context %q", f.longContext)
			}
			tiers = append(tiers, n)
		}
		suite.Specs.Capability.LongContextTiers = tiers
	}
	suite.Normalize()
	return ep, suite, nil
}

func (c *cli) progress(svc *app.Service, runID, kind string, verbose bool) {
	ch, unsub := svc.Bus.Subscribe(runID, 1024)
	defer unsub()
	terminal := func(ev core.Event) bool { // the run's own done event (child stages of a suite keep going)
		return ev.Stage == "" || ev.Stage == kind || ev.Stage == core.KindSuite
	}
	last := time.Time{}
	for ev := range ch {
		switch ev.Kind {
		case core.EventRound, core.EventLevel:
			fmt.Fprintf(c.err, "[%s] %s %d/%d: %d targets\n", ev.Stage, ev.Kind, ev.Round, ev.Rounds, ev.Total)
		case core.EventAttempt, core.EventSample, core.EventProgress:
			if verbose && ev.Attempt != nil {
				a := ev.Attempt
				status := "ok"
				if !a.OK {
					status = a.Reason
				}
				fmt.Fprintf(c.err, "  %-60s %-22s %6.2fs %s\n", a.ModelID, status, a.Latency.Seconds(), strings.TrimSpace(a.Message))
				continue
			}
			if time.Since(last) < time.Second && ev.Done != ev.Total {
				continue
			}
			last = time.Now()
			eta := ""
			if ev.ETA > 0 {
				eta = fmt.Sprintf(" eta %s", (time.Duration(ev.ETA) * time.Second).Round(time.Second))
			}
			fmt.Fprintf(c.err, "[%s] %d/%d ok=%d fail=%d%s\n", ev.Stage, ev.Done, ev.Total, ev.OK, ev.Fail, eta)
		case core.EventCooldown:
			if verbose {
				fmt.Fprintf(c.err, "[%s] waiting %.0fs for cooldown of %s\n", ev.Stage, ev.ETA, ev.Message)
			}
		case core.EventLog:
			fmt.Fprintf(c.err, "[%s] %s\n", ev.Stage, ev.Message)
		case core.EventError:
			fmt.Fprintf(c.err, "[%s] error: %s\n", ev.Stage, ev.Message)
			return
		case core.EventDone:
			fmt.Fprintf(c.err, "[%s] %s\n", ev.Stage, ev.Message)
			if terminal(ev) {
				return
			}
		}
	}
}

func (c *cli) finish(svc *app.Service, id, kind string, f *runFlags) int {
	exit := ExitOK
	if kind == core.KindScan || kind == core.KindSuite {
		res, _, err := svc.ScanResult(id)
		if err == nil {
			s := res.Summary
			fmt.Fprintf(c.err, "models %d · available %d (%.1f%%) · A/B/C %d/%d/%d · families %d (%d available) · median %.2fs\n",
				s.Models, s.OK, 100*float64(s.OK)/float64(max(s.Models, 1)), s.TierCounts["A"], s.TierCounts["B"], s.TierCounts["C"], s.Families, s.FamiliesOK, s.LatencyMedian)
			if len(res.Families) > 0 {
				fmt.Fprintln(c.err, "top families:")
				for i, fam := range res.Families {
					if i >= 5 {
						break
					}
					fmt.Fprintf(c.err, "  %d. %-28s score %.1f  providers %d  stability %s %d/%d  median %.2fs  best %s @ %s\n", i+1, fam.Display, fam.Score, fam.NProvOK, fam.Tier, fam.Successes, fam.Attempts, fam.FMed, fam.Best.ID, fam.Best.ProviderName)
				}
			}
			if f.failUnder > 0 && float64(s.OK)/float64(max(s.Models, 1)) < f.failUnder {
				fmt.Fprintf(c.err, "available share %.3f is below --fail-under %.3f\n", float64(s.OK)/float64(max(s.Models, 1)), f.failUnder)
				exit = ExitThreshold
			}
		}
	}
	if f.out != "" {
		md, err := svc.Report(id, f.lang)
		if err != nil {
			return c.fail(ExitError, err)
		}
		if err := os.WriteFile(f.out, []byte(md), 0o644); err != nil { //nolint:gosec // reports are meant to be shared
			return c.fail(ExitError, err)
		}
		fmt.Fprintf(c.err, "report written to %s\n", f.out)
	}
	if f.jsonOut != "" {
		raw, err := c.exportJSON(svc, id, kind)
		if err != nil {
			return c.fail(ExitError, err)
		}
		if err := os.WriteFile(f.jsonOut, raw, 0o644); err != nil { //nolint:gosec // exports are meant to be shared
			return c.fail(ExitError, err)
		}
		fmt.Fprintf(c.err, "json written to %s\n", f.jsonOut)
	}
	if f.csvOut != "" && (kind == core.KindScan || kind == core.KindSuite) {
		raw, err := svc.ExportCSV(id)
		if err != nil {
			return c.fail(ExitError, err)
		}
		if err := os.WriteFile(f.csvOut, raw, 0o644); err != nil { //nolint:gosec // exports are meant to be shared
			return c.fail(ExitError, err)
		}
		fmt.Fprintf(c.err, "csv written to %s\n", f.csvOut)
	}
	if f.out == "" && f.jsonOut == "" {
		// default: print the run summary document as JSON on stdout for scripting
		run, err := svc.Store.GetRun(id)
		if err == nil {
			_ = json.NewEncoder(c.out).Encode(map[string]any{"schema_version": 1, "run_id": id, "kind": kind, "status": run.Status, "summary": run.Summary})
		}
	}
	return exit
}

func (c *cli) exportJSON(svc *app.Service, id, kind string) ([]byte, error) {
	if kind == core.KindScan || kind == core.KindSuite {
		return svc.ExportJSON(id)
	}
	run, err := svc.Store.GetRun(id)
	if err != nil {
		return nil, err
	}
	return json.MarshalIndent(map[string]any{"schema_version": 1, "run": run, "summary": run.Summary}, "", " ")
}

// ---------------------------------------------------------------- discover / check

func (c *cli) endpointFlags(fs *flag.FlagSet) *runFlags {
	var f runFlags
	fs.StringVar(&f.url, "url", "", "endpoint base URL (required)")
	fs.StringVar(&f.key, "key", "", "API key (or $LLM_SPEEDTEST_KEY)")
	fs.StringVar(&f.dialect, "dialect", "", "gateway dialect (auto-detect when empty)")
	fs.Var(&f.headers, "header", "extra request header k=v (repeatable)")
	fs.BoolVar(&f.insecure, "insecure", false, "skip TLS verification")
	fs.StringVar(&f.proxy, "proxy", "", "proxy URL")
	fs.StringVar(&f.configPath, "config", "", "YAML config file")
	fs.Var(&f.includeOwners, "include-owner", "only this owned_by (repeatable)")
	fs.Var(&f.excludeOwners, "exclude-owner", "skip this owned_by (repeatable)")
	fs.StringVar(&f.idRegex, "id-regex", "", "only ids matching this regex")
	fs.IntVar(&f.limit, "limit", 0, "at most N models")
	return &f
}

func (c *cli) discover(args []string) int {
	fs := flag.NewFlagSet("discover", flag.ContinueOnError)
	fs.SetOutput(c.err)
	f := c.endpointFlags(fs)
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}
	ep, suite, err := c.buildRequest(f)
	if err != nil {
		return c.fail(ExitUsage, err)
	}
	svc := app.New(nil, version.Version)
	res, err := svc.Discover(c.ctx, ep, "", suite.Filter)
	if err != nil {
		return c.fail(ExitEndpoint, err)
	}
	if *jsonOut {
		_ = json.NewEncoder(c.out).Encode(map[string]any{"schema_version": 1, "total": res.Total, "filtered": res.Filtered, "provider_field": res.Field, "provider_rationale": res.Rationale, "dialect": res.Dialect, "providers": res.Providers, "families": res.Families, "models": res.Models})
		return ExitOK
	}
	fmt.Fprintf(c.out, "%d models (%d after filter), %d providers (field %s), %d families, dialect %s\n", res.Total, res.Filtered, len(res.Providers), res.Field, res.Families, res.Dialect.Name)
	for _, p := range res.Providers {
		fmt.Fprintf(c.out, "  %-30s %d models\n", p.Name, p.Models)
	}
	return ExitOK
}

func (c *cli) check(args []string) int {
	fs := flag.NewFlagSet("check", flag.ContinueOnError)
	fs.SetOutput(c.err)
	f := c.endpointFlags(fs)
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}
	ep, _, err := c.buildRequest(f)
	if err != nil {
		return c.fail(ExitUsage, err)
	}
	svc := app.New(nil, version.Version)
	res := svc.Check(c.ctx, ep)
	if *jsonOut {
		_ = json.NewEncoder(c.out).Encode(res)
	} else {
		fmt.Fprintf(c.out, "endpoint   %s\nreachable  %v\nauth       %v\nmodels     %d (%d owners)\ndialect    %s %s\nlatency    %.2fs\n", res.BaseURL, res.Reachable, res.AuthOK, res.Models, res.Owners, res.Dialect.Name, res.Dialect.Version, res.Latency)
		if res.SuggestedV1 != "" {
			fmt.Fprintf(c.out, "hint       use %s\n", res.SuggestedV1)
		}
		if res.Error != "" {
			fmt.Fprintf(c.out, "error      %s\n", res.Error)
		}
	}
	if !res.Reachable || !res.AuthOK {
		return ExitEndpoint
	}
	return ExitOK
}

// ---------------------------------------------------------------- report / compare / runs

func (c *cli) report(args []string) int {
	fs := flag.NewFlagSet("report", flag.ContinueOnError)
	fs.SetOutput(c.err)
	run := fs.String("run", "", "run id (default: latest)")
	out := fs.String("out", "", "output file (default stdout)")
	lang := fs.String("lang", "zh", "zh | en")
	format := fs.String("format", "md", "md | json | csv")
	dataDir := fs.String("data-dir", "", "data directory")
	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}
	st, err := c.openStore(*dataDir)
	if err != nil {
		return c.fail(ExitError, err)
	}
	defer st.Close()
	svc := app.New(st, version.Version)
	id := *run
	if id == "" {
		runs, _, err := st.ListRuns("", "", 1, 0)
		if err != nil || len(runs) == 0 {
			return c.fail(ExitError, errors.New("no runs stored"))
		}
		id = runs[0].ID
	}
	var raw []byte
	switch *format {
	case "md":
		md, err := svc.Report(id, *lang)
		if err != nil {
			return c.fail(ExitError, err)
		}
		raw = []byte(md)
	case "json":
		raw, err = svc.ExportJSON(id)
	case "csv":
		raw, err = svc.ExportCSV(id)
	default:
		return c.fail(ExitUsage, fmt.Errorf("unknown format %q", *format))
	}
	if err != nil {
		return c.fail(ExitError, err)
	}
	if *out == "" {
		_, _ = c.out.Write(raw)
		return ExitOK
	}
	if err := os.WriteFile(*out, raw, 0o644); err != nil { //nolint:gosec // reports are meant to be shared
		return c.fail(ExitError, err)
	}
	fmt.Fprintf(c.err, "written to %s\n", *out)
	return ExitOK
}

func (c *cli) compare(args []string) int {
	fs := flag.NewFlagSet("compare", flag.ContinueOnError)
	fs.SetOutput(c.err)
	a := fs.String("a", "", "first run id")
	b := fs.String("b", "", "second run id")
	dataDir := fs.String("data-dir", "", "data directory")
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}
	if *a == "" || *b == "" {
		return c.fail(ExitUsage, errors.New("--a and --b are required"))
	}
	st, err := c.openStore(*dataDir)
	if err != nil {
		return c.fail(ExitError, err)
	}
	defer st.Close()
	svc := app.New(st, version.Version)
	ra, _, err := svc.ScanResult(*a)
	if err != nil {
		return c.fail(ExitError, err)
	}
	rb, _, err := svc.ScanResult(*b)
	if err != nil {
		return c.fail(ExitError, err)
	}
	type row struct {
		ID      string  `json:"id"`
		StatusA string  `json:"status_a"`
		StatusB string  `json:"status_b"`
		MedA    float64 `json:"median_a"`
		MedB    float64 `json:"median_b"`
	}
	var rows []row
	improved, regressed := 0, 0
	for _, ma := range ra.Models {
		mb := rb.Model(ma.ID)
		if mb == nil {
			continue
		}
		r := row{ID: ma.ID, StatusA: ma.Status, StatusB: mb.Status, MedA: ma.Med, MedB: mb.Med}
		if ma.Status != "ok" && mb.Status == "ok" {
			improved++
		}
		if ma.Status == "ok" && mb.Status != "ok" {
			regressed++
		}
		rows = append(rows, r)
	}
	if *jsonOut {
		_ = json.NewEncoder(c.out).Encode(map[string]any{"schema_version": 1, "a": ra.Summary, "b": rb.Summary, "improved": improved, "regressed": regressed, "models": rows})
		return ExitOK
	}
	fmt.Fprintf(c.out, "A: %d models, %d available · B: %d models, %d available · improved %d · regressed %d\n", ra.Summary.Models, ra.Summary.OK, rb.Summary.Models, rb.Summary.OK, improved, regressed)
	sort.Slice(rows, func(i, j int) bool { return rows[i].ID < rows[j].ID })
	for _, r := range rows {
		if r.StatusA != r.StatusB {
			fmt.Fprintf(c.out, "  %-60s %s -> %s\n", r.ID, r.StatusA, r.StatusB)
		}
	}
	return ExitOK
}

func (c *cli) runs(args []string) int {
	fs := flag.NewFlagSet("runs", flag.ContinueOnError)
	fs.SetOutput(c.err)
	dataDir := fs.String("data-dir", "", "data directory")
	limit := fs.Int("limit", 20, "number of runs")
	jsonOut := fs.Bool("json", false, "print JSON")
	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}
	st, err := c.openStore(*dataDir)
	if err != nil {
		return c.fail(ExitError, err)
	}
	defer st.Close()
	runs, total, err := st.ListRuns("", "", *limit, 0)
	if err != nil {
		return c.fail(ExitError, err)
	}
	if *jsonOut {
		_ = json.NewEncoder(c.out).Encode(map[string]any{"schema_version": 1, "total": total, "runs": runs})
		return ExitOK
	}
	for _, r := range runs {
		fmt.Fprintf(c.out, "%-32s %-10s %-11s %-9s %s\n", r.ID, r.Kind, r.Status, r.Preset, r.EndpointName)
	}
	return ExitOK
}

func (c *cli) configCmd(args []string) int {
	if len(args) > 0 && args[0] != "example" {
		fmt.Fprintln(c.err, "usage: llm-speedtest config example")
		return ExitUsage
	}
	fmt.Fprint(c.out, config.Example())
	return ExitOK
}

// ---------------------------------------------------------------- mcp / skill

func (c *cli) mcp(args []string) int {
	fs := flag.NewFlagSet("mcp", flag.ContinueOnError)
	fs.SetOutput(c.err)
	dataDir := fs.String("data-dir", "", "data directory shared with `serve`")
	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}
	st, err := c.openStore(*dataDir)
	if err != nil {
		return c.fail(ExitError, err)
	}
	defer st.Close()
	svc := app.New(st, version.Version)
	if err := mcpserver.Run(c.ctx, svc); err != nil && !errors.Is(err, context.Canceled) {
		return c.fail(ExitError, err)
	}
	return ExitOK
}

func (c *cli) skill(args []string) int {
	fs := flag.NewFlagSet("skill", flag.ContinueOnError)
	fs.SetOutput(c.err)
	target := fs.String("target", "claude", "install target: claude | codex | cursor | <directory>")
	force := fs.Bool("force", false, "overwrite an existing installation")
	// The sub-command comes first so that flags may follow it:
	//   skill install --target codex | skill export [file.zip] | skill show
	sub := "install"
	if len(args) > 0 && !strings.HasPrefix(args[0], "-") {
		sub, args = args[0], args[1:]
	}
	if err := fs.Parse(args); err != nil {
		return ExitUsage
	}
	switch sub {
	case "install":
		dir, err := skill.Install(*target, *force)
		if err != nil {
			return c.fail(ExitError, err)
		}
		fmt.Fprintf(c.out, "skill installed to %s\n", dir)
		return ExitOK
	case "export":
		out := "llm-speedtest-skill.zip"
		if fs.NArg() > 0 {
			out = fs.Arg(0)
		}
		if err := skill.ExportZip(out); err != nil {
			return c.fail(ExitError, err)
		}
		fmt.Fprintf(c.out, "skill package written to %s\n", out)
		return ExitOK
	case "show":
		fmt.Fprint(c.out, skill.Manifest())
		return ExitOK
	}
	fmt.Fprintln(c.err, "usage: llm-speedtest skill [install --target claude|codex|cursor|<dir>] | export [file.zip] | show")
	return ExitUsage
}
