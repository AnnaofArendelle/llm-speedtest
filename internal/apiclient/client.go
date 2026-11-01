// Package apiclient is a small OpenAI-compatible HTTP client focused on
// measurement: it records timings, caps body sizes, never retries on its own
// and exposes raw status/body for classification.
package apiclient

import (
	"bytes"
	"context"
	"crypto/tls"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"
)

const (
	maxErrorBody  = 64 * 1024
	maxModelsBody = 32 * 1024 * 1024
	maxStreamText = 1024 * 1024
)

// Options configures a Client.
type Options struct {
	Headers         map[string]string
	UserAgent       string
	Insecure        bool
	Proxy           string // explicit proxy URL; empty uses the environment
	MaxConnsPerHost int
}

// Client talks to one endpoint.
type Client struct {
	base    *url.URL
	key     string
	headers map[string]string
	ua      string
	hc      *http.Client
}

// Normalize trims whitespace and trailing slashes from a base URL.
func Normalize(base string) string {
	b := strings.TrimSpace(base)
	for strings.HasSuffix(b, "/") {
		b = strings.TrimSuffix(b, "/")
	}
	return b
}

// New creates a client for base (e.g. https://host/v1).
func New(base, key string, opts Options) (*Client, error) {
	b := Normalize(base)
	u, err := url.Parse(b)
	if err != nil || u.Scheme == "" || u.Host == "" {
		return nil, fmt.Errorf("invalid endpoint URL %q", base)
	}
	if u.Scheme != "http" && u.Scheme != "https" {
		return nil, fmt.Errorf("unsupported scheme %q", u.Scheme)
	}
	tr := &http.Transport{
		Proxy:                 http.ProxyFromEnvironment,
		DialContext:           (&net.Dialer{Timeout: 15 * time.Second, KeepAlive: 30 * time.Second}).DialContext,
		MaxIdleConns:          256,
		MaxIdleConnsPerHost:   64,
		IdleConnTimeout:       90 * time.Second,
		TLSHandshakeTimeout:   15 * time.Second,
		ExpectContinueTimeout: time.Second,
		ForceAttemptHTTP2:     true,
	}
	if opts.MaxConnsPerHost > 0 {
		tr.MaxIdleConnsPerHost = opts.MaxConnsPerHost
		tr.MaxConnsPerHost = opts.MaxConnsPerHost
	}
	if opts.Insecure {
		tr.TLSClientConfig = &tls.Config{InsecureSkipVerify: true} //nolint:gosec // explicit opt-in
	}
	if opts.Proxy != "" {
		pu, err := url.Parse(opts.Proxy)
		if err != nil {
			return nil, fmt.Errorf("invalid proxy URL: %w", err)
		}
		tr.Proxy = http.ProxyURL(pu)
	}
	ua := opts.UserAgent
	if ua == "" {
		ua = "llm-speedtest"
	}
	return &Client{base: u, key: key, headers: opts.Headers, ua: ua, hc: &http.Client{Transport: tr}}, nil
}

// Base returns the normalised base URL.
func (c *Client) Base() string { return c.base.String() }

// Origin returns scheme://host of the endpoint (used by dialect probes).
func (c *Client) Origin() string { return c.base.Scheme + "://" + c.base.Host }

// HTTPClient exposes the underlying client for dialect probes.
func (c *Client) HTTPClient() *http.Client { return c.hc }

// WithBase returns a shallow copy of the client pointing at another base URL
// on the same transport (used when probing /v1 variants).
func (c *Client) WithBase(base string) (*Client, error) {
	u, err := url.Parse(Normalize(base))
	if err != nil || u.Host == "" {
		return nil, fmt.Errorf("invalid endpoint URL %q", base)
	}
	cp := *c
	cp.base = u
	return &cp, nil
}

func (c *Client) newRequest(ctx context.Context, method, path string, body []byte) (*http.Request, error) {
	u := *c.base
	u.Path = strings.TrimSuffix(u.Path, "/") + path
	var rd io.Reader
	if body != nil {
		rd = bytes.NewReader(body)
	}
	req, err := http.NewRequestWithContext(ctx, method, u.String(), rd)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", c.ua)
	req.Header.Set("Accept", "application/json")
	if body != nil {
		req.Header.Set("Content-Type", "application/json")
	}
	if c.key != "" {
		req.Header.Set("Authorization", "Bearer "+c.key)
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	return req, nil
}

// Get performs a GET against an absolute URL with the client's auth headers
// and returns status and a size-capped body.
func (c *Client) Get(ctx context.Context, absURL string, limit int64) (int, []byte, http.Header, error) {
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, absURL, nil)
	if err != nil {
		return 0, nil, nil, err
	}
	req.Header.Set("User-Agent", c.ua)
	req.Header.Set("Accept", "application/json")
	if c.key != "" {
		req.Header.Set("Authorization", "Bearer "+c.key)
	}
	for k, v := range c.headers {
		req.Header.Set(k, v)
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return 0, nil, nil, err
	}
	defer resp.Body.Close()
	b, err := io.ReadAll(io.LimitReader(resp.Body, limit))
	return resp.StatusCode, b, resp.Header, err
}

// Model is one entry of GET /models.
type Model struct {
	ID            string          `json:"id"`
	OwnedBy       string          `json:"owned_by"`
	Object        string          `json:"object,omitempty"`
	Created       int64           `json:"created,omitempty"`
	ContextLength int             `json:"context_length,omitempty"`
	MaxOutput     int             `json:"max_output,omitempty"`
	Raw           json.RawMessage `json:"raw,omitempty"`
}

// ErrStatus is returned when an endpoint answers with a non-2xx status.
type ErrStatus struct {
	Status int
	Body   []byte
}

func (e *ErrStatus) Error() string {
	b := strings.TrimSpace(string(e.Body))
	if len(b) > 200 {
		b = b[:200] + "…"
	}
	return fmt.Sprintf("HTTP %d: %s", e.Status, b)
}

// ListModels fetches /models and extracts the fields the benchmark needs.
func (c *Client) ListModels(ctx context.Context) ([]Model, error) {
	req, err := c.newRequest(ctx, http.MethodGet, "/models", nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.hc.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, maxModelsBody))
	if err != nil {
		return nil, err
	}
	if resp.StatusCode/100 != 2 {
		return nil, &ErrStatus{Status: resp.StatusCode, Body: capBody(body, maxErrorBody)}
	}
	return ParseModels(body)
}

// ParseModels decodes a models response ({"data":[...]} or a bare list).
func ParseModels(body []byte) ([]Model, error) {
	var env struct {
		Data   []json.RawMessage `json:"data"`
		Models []json.RawMessage `json:"models"`
	}
	var items []json.RawMessage
	if err := json.Unmarshal(body, &env); err == nil && (len(env.Data) > 0 || len(env.Models) > 0) {
		items = env.Data
		if len(items) == 0 {
			items = env.Models
		}
	} else if err := json.Unmarshal(body, &items); err != nil {
		return nil, fmt.Errorf("models response is neither an object with data[] nor a list: %w", err)
	}
	out := make([]Model, 0, len(items))
	for _, raw := range items {
		var m map[string]any
		if err := json.Unmarshal(raw, &m); err != nil {
			continue
		}
		id, _ := m["id"].(string)
		if id == "" {
			if n, ok := m["name"].(string); ok {
				id = n
			}
		}
		if id == "" {
			continue
		}
		md := Model{ID: id, Raw: raw}
		md.OwnedBy, _ = m["owned_by"].(string)
		md.Object, _ = m["object"].(string)
		if v, ok := m["created"].(float64); ok {
			md.Created = int64(v)
		}
		md.ContextLength = firstInt(m, "context_length", "context_window", "max_context_length", "max_model_len")
		if md.ContextLength == 0 {
			if caps, ok := m["capabilities"].(map[string]any); ok {
				md.ContextLength = firstInt(caps, "contextWindow", "context_length")
				md.MaxOutput = firstInt(caps, "maxOutput", "max_output_tokens")
			}
		}
		if md.ContextLength == 0 {
			if tp, ok := m["top_provider"].(map[string]any); ok {
				md.ContextLength = firstInt(tp, "context_length")
				if md.MaxOutput == 0 {
					md.MaxOutput = firstInt(tp, "max_completion_tokens")
				}
			}
		}
		if md.MaxOutput == 0 {
			md.MaxOutput = firstInt(m, "max_completion_tokens", "max_output_tokens", "max_tokens")
		}
		out = append(out, md)
	}
	return out, nil
}

func firstInt(m map[string]any, keys ...string) int {
	for _, k := range keys {
		switch v := m[k].(type) {
		case float64:
			if v > 0 {
				return int(v)
			}
		case string:
			var n int
			if _, err := fmt.Sscanf(v, "%d", &n); err == nil && n > 0 {
				return n
			}
		}
	}
	return 0
}

// Message is a chat message.
type Message struct {
	Role    string `json:"role"`
	Content any    `json:"content"`
}

// ChatRequest is the subset of the chat completions request the tool uses.
// Extra is merged into the JSON body verbatim (for dialect/provider-specific
// parameters such as reasoning effort or stream options).
type ChatRequest struct {
	Model          string         `json:"model"`
	Messages       []Message      `json:"messages"`
	MaxTokens      int            `json:"max_tokens,omitempty"`
	Temperature    *float64       `json:"temperature,omitempty"`
	Stop           []string       `json:"stop,omitempty"`
	Tools          []any          `json:"tools,omitempty"`
	ToolChoice     any            `json:"tool_choice,omitempty"`
	ResponseFormat any            `json:"response_format,omitempty"`
	Seed           *int           `json:"seed,omitempty"`
	Extra          map[string]any `json:"-"`
}

func (r ChatRequest) encode(stream bool) ([]byte, error) {
	type alias ChatRequest
	raw, err := json.Marshal(alias(r))
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := json.Unmarshal(raw, &m); err != nil {
		return nil, err
	}
	if stream {
		m["stream"] = true
		m["stream_options"] = map[string]any{"include_usage": true}
	} else {
		m["stream"] = false
	}
	for k, v := range r.Extra {
		m[k] = v
	}
	return json.Marshal(m)
}

// Usage mirrors the OpenAI usage object.
type Usage struct {
	PromptTokens     int `json:"prompt_tokens"`
	CompletionTokens int `json:"completion_tokens"`
	TotalTokens      int `json:"total_tokens"`
}

// ToolCall is a function call emitted by the model.
type ToolCall struct {
	ID        string `json:"id"`
	Name      string `json:"name"`
	Arguments string `json:"arguments"`
}

// ChatResult is the outcome of a non-streaming chat request. Status 0 means
// the request failed before a response arrived; Body then holds a short
// description suitable for classification ("client timeout after 60s").
type ChatResult struct {
	Status       int           `json:"status"`
	Latency      time.Duration `json:"latency"`
	Headers      http.Header   `json:"-"`
	Body         []byte        `json:"-"`
	Model        string        `json:"model,omitempty"`
	FinishReason string        `json:"finish_reason,omitempty"`
	Content      string        `json:"content,omitempty"`
	Reasoning    string        `json:"reasoning,omitempty"`
	ToolCalls    []ToolCall    `json:"tool_calls,omitempty"`
	Usage        *Usage        `json:"usage,omitempty"`
	Err          error         `json:"-"`
}

// OK reports whether the response was a chat completion.
func (r ChatResult) OK() bool {
	return r.Status == 200 && r.Err == nil && bytes.Contains(r.Body, []byte(`"choices"`))
}

// Chat performs a non-streaming chat completion.
func (c *Client) Chat(ctx context.Context, req ChatRequest) ChatResult {
	start := time.Now()
	out := ChatResult{}
	body, err := req.encode(false)
	if err != nil {
		out.Err = err
		out.Body = []byte("encode error: " + err.Error())
		return out
	}
	hreq, err := c.newRequest(ctx, http.MethodPost, "/chat/completions", body)
	if err != nil {
		out.Err = err
		out.Body = []byte("request error: " + err.Error())
		return out
	}
	resp, err := c.hc.Do(hreq)
	if err != nil {
		out.Latency = time.Since(start)
		out.Err = err
		out.Body = []byte(transportMessage(err, ctx, out.Latency))
		return out
	}
	defer resp.Body.Close()
	raw, rerr := io.ReadAll(io.LimitReader(resp.Body, maxModelsBody))
	out.Latency = time.Since(start)
	out.Status = resp.StatusCode
	out.Headers = resp.Header
	if rerr != nil {
		out.Err = rerr
		out.Status = 0
		out.Body = []byte(transportMessage(rerr, ctx, out.Latency))
		return out
	}
	out.Body = capBody(raw, maxErrorBody)
	if resp.StatusCode == 200 {
		parseCompletion(raw, &out)
	}
	return out
}

func parseCompletion(raw []byte, out *ChatResult) {
	var cr struct {
		Model   string `json:"model"`
		Choices []struct {
			FinishReason string `json:"finish_reason"`
			Message      struct {
				Content    any    `json:"content"`
				Reasoning  string `json:"reasoning_content"`
				Reasoning2 string `json:"reasoning"`
				ToolCalls  []struct {
					ID       string `json:"id"`
					Function struct {
						Name      string `json:"name"`
						Arguments string `json:"arguments"`
					} `json:"function"`
				} `json:"tool_calls"`
			} `json:"message"`
		} `json:"choices"`
		Usage *Usage `json:"usage"`
	}
	if err := json.Unmarshal(raw, &cr); err != nil {
		return
	}
	out.Model = cr.Model
	out.Usage = cr.Usage
	if len(cr.Choices) > 0 {
		ch := cr.Choices[0]
		out.FinishReason = ch.FinishReason
		out.Content = contentToString(ch.Message.Content)
		out.Reasoning = ch.Message.Reasoning
		if out.Reasoning == "" {
			out.Reasoning = ch.Message.Reasoning2
		}
		for _, tc := range ch.Message.ToolCalls {
			out.ToolCalls = append(out.ToolCalls, ToolCall{ID: tc.ID, Name: tc.Function.Name, Arguments: tc.Function.Arguments})
		}
	}
}

func contentToString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case []any:
		var sb strings.Builder
		for _, part := range t {
			if pm, ok := part.(map[string]any); ok {
				if s, ok := pm["text"].(string); ok {
					sb.WriteString(s)
				}
			}
		}
		return sb.String()
	}
	return ""
}

// EmbedResult is the outcome of an embeddings request.
type EmbedResult struct {
	Status     int
	Latency    time.Duration
	Body       []byte
	Dimensions int
	Count      int
	Err        error
}

// Embeddings performs POST /embeddings.
func (c *Client) Embeddings(ctx context.Context, model string, input []string) EmbedResult {
	start := time.Now()
	out := EmbedResult{}
	body, _ := json.Marshal(map[string]any{"model": model, "input": input})
	hreq, err := c.newRequest(ctx, http.MethodPost, "/embeddings", body)
	if err != nil {
		out.Err = err
		return out
	}
	resp, err := c.hc.Do(hreq)
	if err != nil {
		out.Latency = time.Since(start)
		out.Err = err
		out.Body = []byte(transportMessage(err, ctx, out.Latency))
		return out
	}
	defer resp.Body.Close()
	raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxModelsBody))
	out.Latency = time.Since(start)
	out.Status = resp.StatusCode
	out.Body = capBody(raw, maxErrorBody)
	if resp.StatusCode == 200 {
		var er struct {
			Data []struct {
				Embedding []float64 `json:"embedding"`
			} `json:"data"`
		}
		if json.Unmarshal(raw, &er) == nil {
			out.Count = len(er.Data)
			if out.Count > 0 {
				out.Dimensions = len(er.Data[0].Embedding)
			}
		}
	}
	return out
}

func capBody(b []byte, n int) []byte {
	if len(b) <= n {
		return b
	}
	return b[:n]
}

// transportMessage renders a transport failure as text for classification.
func transportMessage(err error, ctx context.Context, elapsed time.Duration) string {
	var ne net.Error
	switch {
	case errors.Is(err, context.DeadlineExceeded) || (errors.As(err, &ne) && ne.Timeout()):
		return fmt.Sprintf("client timeout after %.1fs", elapsed.Seconds())
	case errors.Is(err, context.Canceled) || ctx.Err() != nil:
		return "request cancelled"
	}
	return "network error: " + err.Error()
}
