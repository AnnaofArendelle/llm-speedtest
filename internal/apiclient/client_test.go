package apiclient

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

func newServer(t *testing.T, h http.HandlerFunc) (*httptest.Server, *Client) {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	c, err := New(srv.URL+"/v1/", "sk-test", Options{Headers: map[string]string{"X-Custom": "1"}})
	if err != nil {
		t.Fatal(err)
	}
	return srv, c
}

func TestNewAndNormalize(t *testing.T) {
	if Normalize(" https://x/v1// ") != "https://x/v1" {
		t.Fatal("normalize")
	}
	if _, err := New("not a url", "", Options{}); err == nil {
		t.Fatal("expected error")
	}
	if _, err := New("ftp://x/v1", "", Options{}); err == nil {
		t.Fatal("scheme")
	}
	c, err := New("https://api.example.com/v1", "k", Options{Proxy: "http://127.0.0.1:1"})
	if err != nil || c.Origin() != "https://api.example.com" || c.Base() != "https://api.example.com/v1" {
		t.Fatalf("client = %v %v", c, err)
	}
	c2, err := c.WithBase("https://api.example.com")
	if err != nil || c2.Base() != "https://api.example.com" {
		t.Fatal("with base")
	}
}

func TestListModels(t *testing.T) {
	_, c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/models" || r.Header.Get("Authorization") != "Bearer sk-test" || r.Header.Get("X-Custom") != "1" {
			t.Errorf("bad request %s %v", r.URL.Path, r.Header)
		}
		fmt.Fprint(w, `{"object":"list","data":[
			{"id":"a/x","object":"model","owned_by":"a","context_length":128000,"max_completion_tokens":4096},
			{"id":"b/y","owned_by":"b","capabilities":{"contextWindow":262144,"maxOutput":65536}},
			{"id":"c/z","owned_by":"c","top_provider":{"context_length":32768}},
			{"name":"named","owned_by":"d"},
			{"object":"model"}]}`)
	})
	ms, err := c.ListModels(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if len(ms) != 4 || ms[0].ContextLength != 128000 || ms[0].MaxOutput != 4096 || ms[1].ContextLength != 262144 || ms[1].MaxOutput != 65536 || ms[2].ContextLength != 32768 || ms[3].ID != "named" {
		t.Fatalf("models = %+v", ms)
	}
	if ms2, err := ParseModels([]byte(`[{"id":"bare"}]`)); err != nil || len(ms2) != 1 {
		t.Fatal("bare list")
	}
	if _, err := ParseModels([]byte(`"nope"`)); err == nil {
		t.Fatal("expected parse error")
	}
	_, c = newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		fmt.Fprint(w, `{"error":{"message":"Invalid API key"}}`)
	})
	_, err = c.ListModels(context.Background())
	var es *ErrStatus
	if err == nil || !asErrStatus(err, &es) || es.Status != 401 || !strings.Contains(err.Error(), "Invalid API key") {
		t.Fatalf("err = %v", err)
	}
}

func asErrStatus(err error, target **ErrStatus) bool {
	return errors.As(err, target)
}

func TestChat(t *testing.T) {
	_, c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["stream"] != false || body["max_tokens"] != float64(1) || body["custom"] != "x" {
			t.Errorf("body = %v", body)
		}
		fmt.Fprint(w, `{"id":"1","model":"served-x","choices":[{"message":{"role":"assistant","content":"OK","tool_calls":[{"id":"t1","function":{"name":"f","arguments":"{}"}}]},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":1,"total_tokens":4}}`)
	})
	res := c.Chat(context.Background(), ChatRequest{Model: "x", Messages: []Message{{Role: "user", Content: "hi"}}, MaxTokens: 1, Extra: map[string]any{"custom": "x"}})
	if !res.OK() || res.Model != "served-x" || res.Content != "OK" || res.FinishReason != "stop" || res.Usage == nil || res.Usage.CompletionTokens != 1 || len(res.ToolCalls) != 1 || res.ToolCalls[0].Name != "f" {
		t.Fatalf("res = %+v", res)
	}
	_, c = newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(503)
		fmt.Fprint(w, `{"error":{"message":"[occ/x] [429]: {\"error\":{\"message\":\"slow down\"}} (reset after 4s)"}}`)
	})
	res = c.Chat(context.Background(), ChatRequest{Model: "x"})
	if res.OK() || res.Status != 503 || !strings.Contains(string(res.Body), "slow down") {
		t.Fatalf("error res = %+v", res)
	}
	// timeout
	_, c = newServer(t, func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(300 * time.Millisecond)
	})
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	res = c.Chat(ctx, ChatRequest{Model: "x"})
	if res.Status != 0 || !strings.HasPrefix(string(res.Body), "client timeout") {
		t.Fatalf("timeout res = %+v %s", res, res.Body)
	}
	// connection refused
	bad, _ := New("http://127.0.0.1:1/v1", "", Options{})
	res = bad.Chat(context.Background(), ChatRequest{Model: "x"})
	if res.Status != 0 || !strings.HasPrefix(string(res.Body), "network error") {
		t.Fatalf("network res = %s", res.Body)
	}
}

func sse(w http.ResponseWriter, events []string, delay time.Duration) {
	w.Header().Set("Content-Type", "text/event-stream")
	w.WriteHeader(200)
	fl, _ := w.(http.Flusher)
	for _, e := range events {
		fmt.Fprintf(w, "data: %s\n\n", e)
		if fl != nil {
			fl.Flush()
		}
		time.Sleep(delay)
	}
	fmt.Fprint(w, "data: [DONE]\n\n")
	if fl != nil {
		fl.Flush()
	}
}

func TestChatStream(t *testing.T) {
	events := []string{
		`{"model":"m1","choices":[{"delta":{"role":"assistant"}}]}`,
		`{"choices":[{"delta":{"content":"Hel"}}]}`,
		`{"choices":[{"delta":{"content":"lo"}}]}`,
		`{"choices":[{"delta":{"content":" 世界"},"finish_reason":"stop"}]}`,
		`{"choices":[],"usage":{"prompt_tokens":5,"completion_tokens":7,"total_tokens":12}}`,
	}
	_, c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		var body map[string]any
		_ = json.NewDecoder(r.Body).Decode(&body)
		if body["stream"] != true {
			t.Errorf("stream flag missing: %v", body)
		}
		if so, ok := body["stream_options"].(map[string]any); !ok || so["include_usage"] != true {
			t.Errorf("stream_options missing: %v", body)
		}
		sse(w, events, 20*time.Millisecond)
	})
	res := c.ChatStream(context.Background(), ChatRequest{Model: "x", Messages: []Message{{Role: "user", Content: "hi"}}}, StreamOptions{StallTimeout: 2 * time.Second})
	if !res.OK() || res.Content != "Hello 世界" || res.Model != "m1" || res.FinishReason != "stop" {
		t.Fatalf("res = %+v err=%v body=%s", res, res.Err, res.Body)
	}
	if !res.TokensExact || res.OutputTokens != 7 || len(res.Chunks) != 3 {
		t.Fatalf("tokens = %d exact=%v chunks=%d", res.OutputTokens, res.TokensExact, len(res.Chunks))
	}
	if res.TTFT() <= 0 || res.TTST() <= 0 || res.E2E() <= res.TTFT() || res.TPOT() <= 0 || res.TokensPerSecond() <= 0 {
		t.Fatalf("timings ttft=%v ttst=%v e2e=%v tpot=%v tps=%v", res.TTFT(), res.TTST(), res.E2E(), res.TPOT(), res.TokensPerSecond())
	}
	// the role-only first event must not count as the first token
	if res.FirstTokenAt.Sub(res.FirstByteAt) < 10*time.Millisecond {
		t.Fatalf("first token should follow the role-only event: %v", res.FirstTokenAt.Sub(res.FirstByteAt))
	}
	// no usage -> estimate
	_, c = newServer(t, func(w http.ResponseWriter, r *http.Request) {
		sse(w, []string{`{"choices":[{"delta":{"content":"abcdefgh"}}]}`, `{"choices":[{"delta":{"reasoning_content":"think"}}]}`}, 0)
	})
	res = c.ChatStream(context.Background(), ChatRequest{Model: "x"}, StreamOptions{})
	if !res.OK() || res.TokensExact || res.OutputTokens < 2 || res.Reasoning != "think" {
		t.Fatalf("estimate = %+v", res)
	}
	// error status
	_, c = newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(429)
		fmt.Fprint(w, `{"error":{"message":"Rate limit exceeded"}}`)
	})
	res = c.ChatStream(context.Background(), ChatRequest{Model: "x"}, StreamOptions{})
	if res.OK() || res.Status != 429 || !strings.Contains(string(res.Body), "Rate limit") {
		t.Fatalf("err stream = %+v", res)
	}
	// JSON fallback when server ignores stream=true
	_, c = newServer(t, func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `{"model":"j","choices":[{"message":{"content":"plain"},"finish_reason":"stop"}],"usage":{"completion_tokens":2}}`)
	})
	res = c.ChatStream(context.Background(), ChatRequest{Model: "x"}, StreamOptions{})
	if !res.OK() || res.Content != "plain" || res.OutputTokens != 2 || !res.TokensExact {
		t.Fatalf("json fallback = %+v", res)
	}
	// stall
	_, c = newServer(t, func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		fmt.Fprint(w, "data: {\"choices\":[{\"delta\":{\"content\":\"x\"}}]}\n\n")
		if fl, ok := w.(http.Flusher); ok {
			fl.Flush()
		}
		time.Sleep(500 * time.Millisecond)
	})
	res = c.ChatStream(context.Background(), ChatRequest{Model: "x"}, StreamOptions{StallTimeout: 80 * time.Millisecond})
	if !res.Stalled || res.Err == nil {
		t.Fatalf("stall = %+v", res)
	}
}

func TestEmbeddingsAndTokens(t *testing.T) {
	_, c := newServer(t, func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/v1/embeddings" {
			t.Errorf("path %s", r.URL.Path)
		}
		fmt.Fprint(w, `{"data":[{"embedding":[0.1,0.2,0.3]},{"embedding":[0.4,0.5,0.6]}]}`)
	})
	res := c.Embeddings(context.Background(), "e", []string{"a", "b"})
	if res.Status != 200 || res.Count != 2 || res.Dimensions != 3 {
		t.Fatalf("embed = %+v", res)
	}
	if EstimateTokens("") != 0 || EstimateTokens("a") != 1 || EstimateTokens("abcdefgh") != 2 || EstimateTokens("你好") != 2 || EstimateTokens("你好 ab") != 3 {
		t.Fatalf("estimate: %d %d %d %d", EstimateTokens("a"), EstimateTokens("abcdefgh"), EstimateTokens("你好"), EstimateTokens("你好 ab"))
	}
}
