// Package capability probes what a model supports beyond plain chat
// (streaming, usage reporting, tool calls, JSON mode, structured output,
// vision, system prompt adherence, max_tokens / stop handling, deterministic
// decoding, reasoning parameters, long context, embeddings).
package capability

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"golang.org/x/sync/errgroup"
	"golang.org/x/sync/semaphore"

	"github.com/AnnaofArendelle/llm-speedtest/internal/apiclient"
	"github.com/AnnaofArendelle/llm-speedtest/internal/classify"
	"github.com/AnnaofArendelle/llm-speedtest/internal/core"
	"github.com/AnnaofArendelle/llm-speedtest/internal/latency"
)

// Statuses.
const (
	Supported   = "supported"
	Unsupported = "unsupported"
	Error       = "error"
	Skipped     = "skipped"
)

// Probe names in presentation order.
var Probes = []string{"streaming", "usage_in_stream", "tools", "json_mode", "json_schema", "vision", "system_prompt", "max_tokens", "stop", "determinism", "reasoning_param", "long_context", "embeddings"}

// Describe returns a short bilingual description of a probe.
func Describe(probe, lang string) string {
	zh := map[string]string{ //nolint:gosec // UI descriptions, not credentials
		"streaming": "流式输出（SSE）", "usage_in_stream": "流式响应附带 usage", "tools": "工具调用（function calling）", "json_mode": "JSON 模式", "json_schema": "结构化输出（json_schema）",
		"vision": "视觉输入（图片）", "system_prompt": "系统提示遵循", "max_tokens": "遵守 max_tokens", "stop": "遵守 stop 序列", "determinism": "temperature=0 确定性",
		"reasoning_param": "接受推理参数（reasoning_effort）", "long_context": "长上下文", "embeddings": "嵌入接口（/embeddings）",
	}
	en := map[string]string{ //nolint:gosec // UI descriptions, not credentials
		"streaming": "Streaming (SSE)", "usage_in_stream": "Usage in stream", "tools": "Tool calling", "json_mode": "JSON mode", "json_schema": "Structured output (json_schema)",
		"vision": "Vision input", "system_prompt": "System prompt adherence", "max_tokens": "Honours max_tokens", "stop": "Honours stop sequences", "determinism": "Deterministic at temperature 0",
		"reasoning_param": "Accepts reasoning_effort", "long_context": "Long context", "embeddings": "Embeddings endpoint",
	}
	if lang == "en" {
		return en[probe]
	}
	return zh[probe]
}

// Spec configures the probes.
type Spec struct {
	Probes           []string      `json:"probes" yaml:"probes"`
	Concurrency      int           `json:"concurrency" yaml:"concurrency"`
	PerProvider      int           `json:"per_provider" yaml:"per_provider"`
	Timeout          time.Duration `json:"timeout" yaml:"timeout"`
	LongContextTiers []int         `json:"long_context_tiers" yaml:"long_context_tiers"` // tokens
}

// Default returns the standard probe set (long context up to 32k).
func Default() Spec {
	return Spec{Probes: Probes, Concurrency: 2, PerProvider: 1, Timeout: 90 * time.Second, LongContextTiers: []int{8000, 32000}}
}

// Normalize fills zero values.
func (s *Spec) Normalize() {
	d := Default()
	if len(s.Probes) == 0 {
		s.Probes = d.Probes
	}
	if s.Concurrency <= 0 {
		s.Concurrency = d.Concurrency
	}
	if s.PerProvider <= 0 {
		s.PerProvider = 1
	}
	if s.Timeout <= 0 {
		s.Timeout = d.Timeout
	}
	if len(s.LongContextTiers) == 0 {
		s.LongContextTiers = d.LongContextTiers
	}
}

// Target is a model to probe.
type Target struct {
	ID       string
	Provider string
	Name     string
}

// Sink receives results.
type Sink interface {
	Capability(ctx context.Context, r core.CapabilityResult) error
}

// SinkFunc adapts a function to Sink.
type SinkFunc func(ctx context.Context, r core.CapabilityResult) error

// Capability implements Sink.
func (f SinkFunc) Capability(ctx context.Context, r core.CapabilityResult) error { return f(ctx, r) }

// Runner executes probes.
type Runner struct {
	Client *apiclient.Client
	Spec   Spec
	Sink   Sink
	Report func(core.Event)
	Stage  string
	// Skipper (optional) lets users skip models while the run is in progress.
	Skipper *core.Skipper
}

// Run probes every target.
func (r *Runner) Run(ctx context.Context, targets []Target) ([]core.CapabilityResult, error) {
	r.Spec.Normalize()
	if r.Stage == "" {
		r.Stage = "capability"
	}
	total := len(targets) * len(r.Spec.Probes)
	var mu sync.Mutex
	var all []core.CapabilityResult
	done := 0
	emit := func(ev core.Event) {
		if r.Report != nil {
			ev.At = time.Now()
			ev.Stage = r.Stage
			r.Report(ev)
		}
	}
	emit(core.Event{Kind: core.EventRound, Round: 1, Rounds: 1, Total: total, Message: "capability"})
	global := semaphore.NewWeighted(int64(r.Spec.Concurrency))
	var pmu sync.Mutex
	perProv := map[string]*semaphore.Weighted{}
	g, gctx := errgroup.WithContext(ctx)
	for _, t := range targets {
		t := t
		g.Go(func() error {
			pmu.Lock()
			ps, ok := perProv[t.Provider]
			if !ok {
				ps = semaphore.NewWeighted(int64(r.Spec.PerProvider))
				perProv[t.Provider] = ps
			}
			pmu.Unlock()
			if err := ps.Acquire(gctx, 1); err != nil {
				return err
			}
			defer ps.Release(1)
			if err := global.Acquire(gctx, 1); err != nil {
				return err
			}
			defer global.Release(1)
			tctx, release := r.Skipper.Watch(gctx, r.Stage, t.ID, 0)
			defer release()
			for _, p := range r.Spec.Probes {
				if gctx.Err() != nil {
					return gctx.Err()
				}
				start := time.Now()
				var status, evidence string
				if r.Skipper.IsSkipped(t.ID) {
					status, evidence = "skipped", "skipped by user"
				} else {
					emit(core.Event{Kind: core.EventStart, Stage: r.Stage, Attempt: &core.Attempt{ModelID: t.ID, At: start}, Message: t.ID + " " + p})
					status, evidence = r.probe(tctx, p, t)
					if gctx.Err() != nil {
						return gctx.Err()
					}
					if r.Skipper.IsSkipped(t.ID) {
						status, evidence = "skipped", "skipped by user"
					}
				}
				res := core.CapabilityResult{ModelID: t.ID, Probe: p, Status: status, Evidence: evidence, Latency: time.Since(start)}
				if r.Sink != nil {
					if err := r.Sink.Capability(gctx, res); err != nil {
						return err
					}
				}
				mu.Lock()
				all = append(all, res)
				done++
				d := done
				mu.Unlock()
				emit(core.Event{Kind: core.EventProgress, Done: d, Total: total, Message: t.ID + " " + p, Data: res})
			}
			return nil
		})
	}
	err := g.Wait()
	sort.Slice(all, func(i, j int) bool {
		if all[i].ModelID != all[j].ModelID {
			return all[i].ModelID < all[j].ModelID
		}
		return probeIndex(all[i].Probe) < probeIndex(all[j].Probe)
	})
	if err == nil {
		emit(core.Event{Kind: core.EventDone, Done: done, Total: total, Message: "capability complete"})
	}
	return all, err
}

func probeIndex(p string) int {
	for i, q := range Probes {
		if q == p {
			return i
		}
	}
	return len(Probes)
}

var (
	jsonObjRE = regexp.MustCompile(`(?s)\{.*\}`)
	numRE     = regexp.MustCompile(`\d+`)
)

func (r *Runner) chat(ctx context.Context, req apiclient.ChatRequest) apiclient.ChatResult {
	pctx, cancel := context.WithTimeout(ctx, r.Spec.Timeout)
	defer cancel()
	return r.Client.Chat(pctx, req)
}

func user(s string) []apiclient.Message { return []apiclient.Message{{Role: "user", Content: s}} }

// errorStatus maps a failed chat result to a probe status: parameter /
// feature rejections are "unsupported", everything else "error".
func errorStatus(res apiclient.ChatResult, name string) (string, string) {
	out := classify.Evaluate(res.Status, res.Body, name)
	msg := classify.CleanMessage(out.Message)
	low := strings.ToLower(msg)
	if res.Status == 400 || res.Status == 422 || out.Reason == "param_incompatible" || out.Reason == "bad_request" || out.Reason == "not_chat_model" {
		for _, k := range []string{"not support", "unsupported", "invalid", "unknown parameter", "does not", "unrecognized", "not allowed", "cannot",
			"context length", "context_length", "maximum context", "too long", "too many tokens", "exceeds", "exceeded", "reduce the length", "max_tokens"} {
			if strings.Contains(low, k) {
				return Unsupported, "HTTP " + fmt.Sprint(res.Status) + ": " + msg
			}
		}
	}
	return Error, fmt.Sprintf("%s (HTTP %d): %s", classify.Info(out.Reason).NameEN, res.Status, msg)
}

func (r *Runner) probe(ctx context.Context, p string, t Target) (string, string) {
	switch p {
	case "streaming", "usage_in_stream":
		pctx, cancel := context.WithTimeout(ctx, r.Spec.Timeout)
		defer cancel()
		res := r.Client.ChatStream(pctx, apiclient.ChatRequest{Model: t.ID, Messages: user("Reply with the word OK."), MaxTokens: 8}, apiclient.StreamOptions{StallTimeout: 30 * time.Second})
		if !res.OK() {
			status := res.Status
			if res.Err != nil && status == 200 {
				status = 0
			}
			cr := apiclient.ChatResult{Status: status, Body: res.Body}
			return errorStatus(cr, t.Name)
		}
		isSSE := strings.Contains(res.Headers.Get("Content-Type"), "text/event-stream")
		if p == "streaming" {
			if isSSE && len(res.Chunks) > 0 {
				return Supported, fmt.Sprintf("%d chunks, first token after %s", len(res.Chunks), res.TTFT().Round(time.Millisecond))
			}
			return Unsupported, "server answered stream=true with a non-streaming JSON body"
		}
		if res.TokensExact {
			return Supported, fmt.Sprintf("usage.completion_tokens=%d", res.OutputTokens)
		}
		return Unsupported, "no usage object in the stream (stream_options.include_usage ignored)"
	case "tools":
		tools := []any{map[string]any{"type": "function", "function": map[string]any{"name": "get_weather", "description": "Get the current weather for a city",
			"parameters": map[string]any{"type": "object", "properties": map[string]any{"city": map[string]any{"type": "string"}}, "required": []string{"city"}}}}}
		res := r.chat(ctx, apiclient.ChatRequest{Model: t.ID, Messages: user("What is the weather in Paris right now? You must call the get_weather tool."), Tools: tools, ToolChoice: "auto", MaxTokens: 128})
		if !res.OK() {
			return errorStatus(res, t.Name)
		}
		if len(res.ToolCalls) > 0 {
			var args map[string]any
			if json.Unmarshal([]byte(res.ToolCalls[0].Arguments), &args) == nil {
				return Supported, fmt.Sprintf("called %s(%s)", res.ToolCalls[0].Name, strings.TrimSpace(res.ToolCalls[0].Arguments))
			}
			return Supported, "tool call emitted with non-JSON arguments: " + trunc(res.ToolCalls[0].Arguments, 80)
		}
		return Unsupported, "answered in text without calling the tool: " + trunc(res.Content, 80)
	case "json_mode":
		res := r.chat(ctx, apiclient.ChatRequest{Model: t.ID, Messages: user(`Return a JSON object with a single key "answer" whose value is the string "ok". Output JSON only.`), ResponseFormat: map[string]any{"type": "json_object"}, MaxTokens: 64})
		if !res.OK() {
			return errorStatus(res, t.Name)
		}
		if obj := parseJSONObject(res.Content); obj != nil {
			return Supported, "valid JSON object: " + trunc(res.Content, 80)
		}
		return Unsupported, "response is not a JSON object: " + trunc(res.Content, 80)
	case "json_schema":
		schema := map[string]any{"type": "json_schema", "json_schema": map[string]any{"name": "answer", "strict": true,
			"schema": map[string]any{"type": "object", "properties": map[string]any{"answer": map[string]any{"type": "string"}, "confidence": map[string]any{"type": "number"}}, "required": []string{"answer", "confidence"}, "additionalProperties": false}}}
		res := r.chat(ctx, apiclient.ChatRequest{Model: t.ID, Messages: user(`Answer the question "what is 2+2" as JSON with keys answer (string) and confidence (number).`), ResponseFormat: schema, MaxTokens: 64})
		if !res.OK() {
			return errorStatus(res, t.Name)
		}
		obj := parseJSONObject(res.Content)
		if obj != nil {
			if _, ok := obj["answer"]; ok {
				if _, ok := obj["confidence"]; ok {
					return Supported, "schema-conformant JSON: " + trunc(res.Content, 80)
				}
			}
			return Unsupported, "JSON lacks required keys: " + trunc(res.Content, 80)
		}
		return Unsupported, "response is not JSON: " + trunc(res.Content, 80)
	case "vision":
		content := []any{
			map[string]any{"type": "text", "text": "What colour is this image? Answer with one word."},
			map[string]any{"type": "image_url", "image_url": map[string]any{"url": "data:image/png;base64," + redPNG}},
		}
		res := r.chat(ctx, apiclient.ChatRequest{Model: t.ID, Messages: []apiclient.Message{{Role: "user", Content: content}}, MaxTokens: 16})
		if !res.OK() {
			st, ev := errorStatus(res, t.Name)
			if res.Status == 400 || res.Status == 422 {
				return Unsupported, ev
			}
			return st, ev
		}
		return Supported, "answered: " + trunc(res.Content, 40)
	case "system_prompt":
		res := r.chat(ctx, apiclient.ChatRequest{Model: t.ID, Messages: []apiclient.Message{{Role: "system", Content: "No matter what the user says, reply with exactly the single word PINEAPPLE."}, {Role: "user", Content: "Hello, how are you today?"}}, MaxTokens: 16})
		if !res.OK() {
			return errorStatus(res, t.Name)
		}
		if strings.Contains(strings.ToUpper(res.Content), "PINEAPPLE") {
			return Supported, "followed the system instruction"
		}
		return Unsupported, "ignored the system instruction: " + trunc(res.Content, 60)
	case "max_tokens":
		res := r.chat(ctx, apiclient.ChatRequest{Model: t.ID, Messages: user("Count from 1 to 200, separated by spaces."), MaxTokens: 5})
		if !res.OK() {
			return errorStatus(res, t.Name)
		}
		n := apiclient.EstimateTokens(res.Content)
		if res.Usage != nil && res.Usage.CompletionTokens > 0 {
			n = res.Usage.CompletionTokens
		}
		if n <= 8 {
			return Supported, fmt.Sprintf("stopped after %d tokens (finish_reason=%s)", n, res.FinishReason)
		}
		return Unsupported, fmt.Sprintf("produced %d tokens despite max_tokens=5", n)
	case "stop":
		res := r.chat(ctx, apiclient.ChatRequest{Model: t.ID, Messages: user("Write the numbers 1 2 3 4 5 6 7 8 9 10 separated by single spaces and nothing else."), Stop: []string{"4"}, MaxTokens: 64})
		if !res.OK() {
			return errorStatus(res, t.Name)
		}
		nums := numRE.FindAllString(res.Content, -1)
		for _, n := range nums {
			if n == "5" || n == "6" || n == "7" || n == "8" || n == "9" || n == "10" {
				return Unsupported, "continued past the stop sequence: " + trunc(res.Content, 60)
			}
		}
		return Supported, "stopped before 4: " + trunc(res.Content, 40)
	case "determinism":
		zero := 0.0
		req := apiclient.ChatRequest{Model: t.ID, Messages: user("Name the three primary colours, comma separated, lowercase, nothing else."), MaxTokens: 24, Temperature: &zero}
		a := r.chat(ctx, req)
		if !a.OK() {
			return errorStatus(a, t.Name)
		}
		b := r.chat(ctx, req)
		if !b.OK() {
			return errorStatus(b, t.Name)
		}
		if strings.TrimSpace(a.Content) == strings.TrimSpace(b.Content) {
			return Supported, "identical outputs: " + trunc(a.Content, 40)
		}
		return Unsupported, "outputs differ: " + trunc(a.Content, 30) + " / " + trunc(b.Content, 30)
	case "reasoning_param":
		res := r.chat(ctx, apiclient.ChatRequest{Model: t.ID, Messages: user("Reply with OK."), MaxTokens: 8, Extra: map[string]any{"reasoning_effort": "low"}})
		if !res.OK() {
			return errorStatus(res, t.Name)
		}
		return Supported, "accepted reasoning_effort=low"
	case "long_context":
		best := 0
		evidence := ""
		for _, tier := range r.Spec.LongContextTiers {
			prompt := latency.Synthesize(tier, fmt.Sprintf("ctx%d", tier)) + "\n\nIgnore the text above and reply with the word OK."
			res := r.chat(ctx, apiclient.ChatRequest{Model: t.ID, Messages: user(prompt), MaxTokens: 8})
			if res.OK() {
				best = tier
				continue
			}
			st, ev := errorStatus(res, t.Name)
			if st == Error && best == 0 {
				return Error, ev
			}
			evidence = fmt.Sprintf("rejected ~%d tokens: %s", tier, trunc(ev, 80))
			break
		}
		if best == 0 {
			return Unsupported, evidence
		}
		if evidence != "" {
			return Supported, fmt.Sprintf("accepted ~%d tokens; %s", best, evidence)
		}
		return Supported, fmt.Sprintf("accepted ~%d tokens", best)
	case "embeddings":
		pctx, cancel := context.WithTimeout(ctx, r.Spec.Timeout)
		defer cancel()
		res := r.Client.Embeddings(pctx, t.ID, []string{"hello world"})
		if res.Status == 200 && res.Count > 0 {
			return Supported, fmt.Sprintf("%d dimensions", res.Dimensions)
		}
		if res.Status == 0 {
			return Error, string(res.Body)
		}
		return Unsupported, fmt.Sprintf("HTTP %d: %s", res.Status, trunc(classify.CleanMessage(classify.ExtractError(res.Status, res.Body)), 80))
	}
	return Skipped, "unknown probe"
}

func parseJSONObject(s string) map[string]any {
	m := jsonObjRE.FindString(s)
	if m == "" {
		return nil
	}
	var obj map[string]any
	if json.Unmarshal([]byte(m), &obj) != nil {
		return nil
	}
	return obj
}

func trunc(s string, n int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

// redPNG is a 2x2 solid red PNG.
const redPNG = "iVBORw0KGgoAAAANSUhEUgAAAAIAAAACCAIAAAD91JpzAAAAEklEQVR4nGP4z8DAwMDAAMEADbYD/4JQOq4AAAAASUVORK5CYII="

// Matrix summarises results as model -> probe -> status.
type Matrix struct {
	Models  []string                   `json:"models"`
	Probes  []string                   `json:"probes"`
	Cells   map[string]map[string]Cell `json:"cells"`
	Support map[string]int             `json:"support"` // per model: number of supported probes
}

// Cell is one matrix entry.
type Cell struct {
	Status   string  `json:"status"`
	Evidence string  `json:"evidence,omitempty"`
	Latency  float64 `json:"latency_s,omitempty"`
}

// Summarize builds the matrix.
func Summarize(results []core.CapabilityResult, probes []string) Matrix {
	if len(probes) == 0 {
		probes = Probes
	}
	m := Matrix{Models: []string{}, Probes: probes, Cells: map[string]map[string]Cell{}, Support: map[string]int{}}
	seen := map[string]bool{}
	for _, r := range results {
		if !seen[r.ModelID] {
			seen[r.ModelID] = true
			m.Models = append(m.Models, r.ModelID)
			m.Cells[r.ModelID] = map[string]Cell{}
		}
		m.Cells[r.ModelID][r.Probe] = Cell{Status: r.Status, Evidence: r.Evidence, Latency: r.Latency.Seconds()}
		if r.Status == Supported {
			m.Support[r.ModelID]++
		}
	}
	sort.SliceStable(m.Models, func(i, j int) bool {
		if m.Support[m.Models[i]] != m.Support[m.Models[j]] {
			return m.Support[m.Models[i]] > m.Support[m.Models[j]]
		}
		return m.Models[i] < m.Models[j]
	})
	return m
}
