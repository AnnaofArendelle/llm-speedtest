// Package mockllm is a scriptable OpenAI-compatible server used by tests and
// by the demo mode. Behaviours are configured per model: latency, error
// envelopes (optionally wrapped like a gateway), cooldown locks, flaky
// sequences, streaming, tool calls, JSON mode, vision and context limits.
package mockllm

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"time"
)

// Behavior configures how a model answers.
type Behavior struct {
	Latency      time.Duration // base latency before answering
	Jitter       time.Duration // random extra latency in [0, Jitter)
	Status       int           // 0 = success; otherwise the HTTP status of the error
	Error        string        // error message (plain)
	Inner        int           // when > 0 the error is wrapped gateway-style with this upstream status
	Conn         string        // connection identifier for the wrapper
	ResetAfter   time.Duration // lock the model after a 401/403/429 for this long
	FailFirst    int           // the first N requests fail, later ones succeed
	ServedModel  string        // model name reported in responses (default: the upstream id)
	Content      string        // completion text (default "OK")
	StreamChunks []string      // explicit streamed pieces; default splits Content into words
	ChunkDelay   time.Duration // delay between streamed chunks
	NoUsage      bool          // omit usage from responses
	ToolCall     bool          // answer with a tool call when tools are supplied
	JSONMode     bool          // answer with a JSON object when response_format is json
	Vision       bool          // accept image parts
	MaxContext   int           // reject prompts larger than this many (estimated) tokens
	Embedding    bool          // serve /embeddings for this model
	IgnoreStream bool          // answer streaming requests with a plain JSON body
	Hang         time.Duration // sleep this long before answering (timeouts)
}

// ModelSpec declares a model served by the mock.
type ModelSpec struct {
	ID            string
	OwnedBy       string
	ContextLength int
	Behavior      Behavior
}

// Server is the mock HTTP server.
type Server struct {
	*httptest.Server
	mu          sync.Mutex
	models      []ModelSpec
	byID        map[string]*ModelSpec
	counts      map[string]int
	inflight    map[string]int
	maxInflight map[string]int
	locks       map[string]lock
	RequireKey  string
	Dialect     string // "", "9router", "oneapi"
	ModelsDelay time.Duration
	Version     string
	rnd         *rand.Rand
}

type lock struct {
	until  time.Time
	status int
	body   string
}

// New starts a mock server with the given models.
func New(models []ModelSpec) *Server {
	s := &Server{models: models, byID: map[string]*ModelSpec{}, counts: map[string]int{}, inflight: map[string]int{},
		maxInflight: map[string]int{}, locks: map[string]lock{}, Version: "0.0.0-mock", rnd: rand.New(rand.NewSource(1))}
	for i := range models {
		s.byID[models[i].ID] = &models[i]
	}
	mux := http.NewServeMux()
	mux.HandleFunc("/v1/models", s.handleModels)
	mux.HandleFunc("/models", s.handleModels)
	mux.HandleFunc("/v1/chat/completions", s.handleChat)
	mux.HandleFunc("/chat/completions", s.handleChat)
	mux.HandleFunc("/v1/embeddings", s.handleEmbeddings)
	mux.HandleFunc("/api/version", s.handleVersion)
	mux.HandleFunc("/api/health", s.handleHealth)
	mux.HandleFunc("/api/status", s.handleStatus)
	s.Server = httptest.NewServer(mux)
	return s
}

// Counts returns the number of chat requests seen per model.
func (s *Server) Counts() map[string]int {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]int, len(s.counts))
	for k, v := range s.counts {
		out[k] = v
	}
	return out
}

// MaxInflight returns the highest number of concurrent chat requests
// observed per owned_by.
func (s *Server) MaxInflight() map[string]int {
	s.mu.Lock()
	defer s.mu.Unlock()
	out := make(map[string]int, len(s.maxInflight))
	for k, v := range s.maxInflight {
		out[k] = v
	}
	return out
}

// TotalRequests returns the total number of chat requests.
func (s *Server) TotalRequests() int {
	s.mu.Lock()
	defer s.mu.Unlock()
	n := 0
	for _, v := range s.counts {
		n += v
	}
	return n
}

// SetBehavior replaces the behaviour of a model at runtime.
func (s *Server) SetBehavior(id string, b Behavior) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if m, ok := s.byID[id]; ok {
		m.Behavior = b
	}
}

func (s *Server) auth(w http.ResponseWriter, r *http.Request) bool {
	if s.RequireKey == "" {
		return true
	}
	if r.Header.Get("Authorization") != "Bearer "+s.RequireKey {
		writeErr(w, 401, "Invalid API key", "invalid_request_error", "invalid_api_key")
		return false
	}
	return true
}

func writeErr(w http.ResponseWriter, status int, msg, typ, code string) {
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(status)
	_ = json.NewEncoder(w).Encode(map[string]any{"error": map[string]any{"message": msg, "type": typ, "code": code}})
}

func (s *Server) handleModels(w http.ResponseWriter, r *http.Request) {
	if !s.auth(w, r) {
		return
	}
	if s.ModelsDelay > 0 {
		time.Sleep(s.ModelsDelay)
	}
	data := make([]map[string]any, 0, len(s.models))
	for _, m := range s.models {
		e := map[string]any{"id": m.ID, "object": "model", "owned_by": m.OwnedBy}
		if m.ContextLength > 0 {
			e["context_length"] = m.ContextLength
		}
		data = append(data, e)
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"object": "list", "data": data})
}

func (s *Server) handleVersion(w http.ResponseWriter, _ *http.Request) {
	if s.Dialect != "9router" {
		http.NotFound(w, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"currentVersion": s.Version, "latestVersion": s.Version, "hasUpdate": false})
}

func (s *Server) handleHealth(w http.ResponseWriter, _ *http.Request) {
	if s.Dialect != "9router" {
		http.NotFound(w, nil)
		return
	}
	w.WriteHeader(200)
	_, _ = w.Write([]byte(`{"status":"ok"}`))
}

func (s *Server) handleStatus(w http.ResponseWriter, _ *http.Request) {
	if s.Dialect != "oneapi" {
		http.NotFound(w, nil)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"success": true, "message": "", "data": map[string]any{"version": s.Version, "system_name": "mock one-api"}})
}

type chatReq struct {
	Model          string           `json:"model"`
	Messages       []map[string]any `json:"messages"`
	Stream         bool             `json:"stream"`
	MaxTokens      int              `json:"max_tokens"`
	Tools          []any            `json:"tools"`
	ResponseFormat map[string]any   `json:"response_format"`
}

func upstreamName(m *ModelSpec) string {
	if m.Behavior.ServedModel != "" {
		return m.Behavior.ServedModel
	}
	id := m.ID
	if m.OwnedBy != "" && strings.HasPrefix(id, m.OwnedBy+"/") {
		id = strings.TrimPrefix(id, m.OwnedBy+"/")
	}
	return id
}

func (s *Server) errorBody(m *ModelSpec, b Behavior, suffix string) (int, string) {
	msg := b.Error
	if msg == "" {
		msg = http.StatusText(b.Status)
	}
	status := b.Status
	if b.Inner > 0 {
		inner, _ := json.Marshal(map[string]any{"error": map[string]any{"message": msg, "type": "upstream_error"}})
		conn := b.Conn
		if conn == "" {
			conn = "mock-conn"
		}
		msg = fmt.Sprintf("[%s/%s] [%d]: %s", conn, upstreamName(m), b.Inner, inner)
		if status == 0 || status == b.Inner {
			status = 503
		}
	}
	msg += suffix
	body, _ := json.Marshal(map[string]any{"error": map[string]any{"message": msg, "type": "invalid_request_error", "code": "bad_request"}})
	return status, string(body)
}

func (s *Server) errorBodyNoSuffix(m *ModelSpec, b Behavior) string {
	_, body := s.errorBody(m, b, "")
	return body
}

func (s *Server) handleChat(w http.ResponseWriter, r *http.Request) {
	if !s.auth(w, r) {
		return
	}
	var req chatReq
	if err := json.NewDecoder(http.MaxBytesReader(w, r.Body, 8<<20)).Decode(&req); err != nil {
		writeErr(w, 400, "invalid JSON body", "invalid_request_error", "bad_request")
		return
	}
	if req.Model == "" {
		writeErr(w, 400, "Missing model", "invalid_request_error", "bad_request")
		return
	}
	s.mu.Lock()
	m := s.byID[req.Model]
	if m == nil {
		s.mu.Unlock()
		writeErr(w, 404, fmt.Sprintf("The model '%s' does not exist", req.Model), "invalid_request_error", "model_not_found")
		return
	}
	s.counts[m.ID]++
	count := s.counts[m.ID]
	s.inflight[m.OwnedBy]++
	if s.inflight[m.OwnedBy] > s.maxInflight[m.OwnedBy] {
		s.maxInflight[m.OwnedBy] = s.inflight[m.OwnedBy]
	}
	b := m.Behavior
	lk, locked := s.locks[m.ID]
	if locked && time.Now().After(lk.until) {
		delete(s.locks, m.ID)
		locked = false
	}
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.inflight[m.OwnedBy]--
		s.mu.Unlock()
	}()

	if locked {
		remaining := time.Until(lk.until).Round(time.Second)
		if remaining < time.Second {
			remaining = time.Second
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(lk.status)
		_, _ = w.Write([]byte(strings.TrimSuffix(lk.body, "\"}}") + fmt.Sprintf(" (reset after %ds)\"}}", int(remaining.Seconds()))))
		return
	}
	if b.Hang > 0 {
		select {
		case <-time.After(b.Hang):
		case <-r.Context().Done():
			return
		}
	}
	delay := b.Latency
	if b.Jitter > 0 {
		s.mu.Lock()
		delay += time.Duration(s.rnd.Int63n(int64(b.Jitter)))
		s.mu.Unlock()
	}
	if delay > 0 {
		select {
		case <-time.After(delay):
		case <-r.Context().Done():
			return
		}
	}
	shouldFail := b.Status != 0 && (b.FailFirst == 0 || count <= b.FailFirst)
	if shouldFail {
		lockable := b.ResetAfter > 0 && (b.Inner == 401 || b.Inner == 403 || b.Inner == 429 || b.Status == 401 || b.Status == 403 || b.Status == 429)
		suffix := ""
		if lockable {
			// Gateways report the cooldown on the very first failure as well.
			secs := int((b.ResetAfter + time.Second - 1) / time.Second)
			suffix = fmt.Sprintf(" (reset after %ds)", secs)
		}
		status, body := s.errorBody(m, b, suffix)
		if lockable {
			s.mu.Lock()
			s.locks[m.ID] = lock{until: time.Now().Add(b.ResetAfter), status: status, body: s.errorBodyNoSuffix(m, b)}
			s.mu.Unlock()
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
		return
	}
	// capability checks
	if !b.Vision {
		for _, msg := range req.Messages {
			if parts, ok := msg["content"].([]any); ok {
				for _, p := range parts {
					if pm, ok := p.(map[string]any); ok && pm["type"] == "image_url" {
						writeErr(w, 400, "This model does not support image input", "invalid_request_error", "unsupported_input")
						return
					}
				}
			}
		}
	}
	if b.MaxContext > 0 {
		chars := 0
		for _, msg := range req.Messages {
			if sc, ok := msg["content"].(string); ok {
				chars += len(sc)
			}
		}
		if chars/4 > b.MaxContext {
			writeErr(w, 400, fmt.Sprintf("This model's maximum context length is %d tokens", b.MaxContext), "invalid_request_error", "context_length_exceeded")
			return
		}
	}
	content := b.Content
	if content == "" {
		content = "OK"
	}
	if b.JSONMode && req.ResponseFormat != nil {
		content = jsonForFormat(req.ResponseFormat)
	}
	usage := map[string]any{"prompt_tokens": 12, "completion_tokens": len(strings.Fields(content)), "total_tokens": 12 + len(strings.Fields(content))}
	if req.MaxTokens > 0 && req.MaxTokens < len(strings.Fields(content)) {
		content = strings.Join(strings.Fields(content)[:req.MaxTokens], " ")
		usage["completion_tokens"] = req.MaxTokens
	}
	var toolCalls []map[string]any
	if b.ToolCall && len(req.Tools) > 0 {
		toolCalls = []map[string]any{{"id": "call_1", "type": "function", "function": map[string]any{"name": "get_weather", "arguments": `{"city":"Paris"}`}}}
		content = ""
	}
	if req.Stream && !b.IgnoreStream {
		s.stream(w, r, m, b, content, toolCalls, usage)
		return
	}
	msg := map[string]any{"role": "assistant", "content": content}
	if toolCalls != nil {
		msg["tool_calls"] = toolCalls
	}
	resp := map[string]any{"id": "chatcmpl-mock", "object": "chat.completion", "created": time.Now().Unix(), "model": upstreamName(m),
		"choices": []map[string]any{{"index": 0, "message": msg, "finish_reason": "stop"}}}
	if !b.NoUsage {
		resp["usage"] = usage
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(resp)
}

func (s *Server) stream(w http.ResponseWriter, r *http.Request, m *ModelSpec, b Behavior, content string, toolCalls []map[string]any, usage map[string]any) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.WriteHeader(200)
	fl, _ := w.(http.Flusher)
	send := func(v any) bool {
		raw, _ := json.Marshal(v)
		if _, err := fmt.Fprintf(w, "data: %s\n\n", raw); err != nil {
			return false
		}
		if fl != nil {
			fl.Flush()
		}
		return true
	}
	model := upstreamName(m)
	send(map[string]any{"id": "chatcmpl-mock", "object": "chat.completion.chunk", "model": model, "choices": []map[string]any{{"index": 0, "delta": map[string]any{"role": "assistant"}}}})
	chunks := b.StreamChunks
	if len(chunks) == 0 {
		for i, wd := range strings.Fields(content) {
			if i > 0 {
				wd = " " + wd
			}
			chunks = append(chunks, wd)
		}
	}
	if toolCalls != nil {
		chunks = nil
		send(map[string]any{"choices": []map[string]any{{"index": 0, "delta": map[string]any{"tool_calls": []map[string]any{{"index": 0, "id": "call_1", "type": "function", "function": map[string]any{"name": "get_weather", "arguments": `{"city":"Paris"}`}}}}}}})
	}
	for i, ch := range chunks {
		if b.ChunkDelay > 0 {
			select {
			case <-time.After(b.ChunkDelay):
			case <-r.Context().Done():
				return
			}
		}
		delta := map[string]any{"content": ch}
		ev := map[string]any{"choices": []map[string]any{{"index": 0, "delta": delta}}}
		if i == len(chunks)-1 {
			ev["choices"].([]map[string]any)[0]["finish_reason"] = "stop"
		}
		if !send(ev) {
			return
		}
	}
	if toolCalls != nil {
		send(map[string]any{"choices": []map[string]any{{"index": 0, "delta": map[string]any{}, "finish_reason": "tool_calls"}}})
	}
	if !b.NoUsage {
		if n := len(chunks); n > 0 {
			usage["completion_tokens"] = n
			usage["total_tokens"] = 12 + n
		}
		send(map[string]any{"choices": []map[string]any{}, "usage": usage})
	}
	_, _ = fmt.Fprint(w, "data: [DONE]\n\n")
	if fl != nil {
		fl.Flush()
	}
}

// jsonForFormat builds an object that satisfies the required properties of a
// json_schema response_format (strings -> "OK", numbers -> 1, booleans ->
// true); plain json_object mode answers {"answer":"OK"}.
func jsonForFormat(rf map[string]any) string {
	obj := map[string]any{"answer": "OK"}
	if js, ok := rf["json_schema"].(map[string]any); ok {
		if schema, ok := js["schema"].(map[string]any); ok {
			props, _ := schema["properties"].(map[string]any)
			for name, p := range props {
				pm, _ := p.(map[string]any)
				switch pm["type"] {
				case "number", "integer":
					obj[name] = 1
				case "boolean":
					obj[name] = true
				default:
					obj[name] = "OK"
				}
			}
		}
	}
	raw, _ := json.Marshal(obj)
	return string(raw)
}

func (s *Server) handleEmbeddings(w http.ResponseWriter, r *http.Request) {
	if !s.auth(w, r) {
		return
	}
	var req struct {
		Model string `json:"model"`
		Input any    `json:"input"`
	}
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		writeErr(w, 400, "invalid JSON body", "invalid_request_error", "bad_request")
		return
	}
	s.mu.Lock()
	m := s.byID[req.Model]
	s.mu.Unlock()
	if m == nil || !m.Behavior.Embedding {
		writeErr(w, 404, fmt.Sprintf("The model '%s' does not support embeddings", req.Model), "invalid_request_error", "model_not_found")
		return
	}
	n := 1
	if arr, ok := req.Input.([]any); ok {
		n = len(arr)
	}
	data := make([]map[string]any, 0, n)
	for i := 0; i < n; i++ {
		data = append(data, map[string]any{"object": "embedding", "index": i, "embedding": []float64{0.1, 0.2, 0.3, 0.4}})
	}
	w.Header().Set("Content-Type", "application/json")
	_ = json.NewEncoder(w).Encode(map[string]any{"object": "list", "data": data, "model": req.Model, "usage": map[string]any{"prompt_tokens": 4 * n, "total_tokens": 4 * n}})
}
