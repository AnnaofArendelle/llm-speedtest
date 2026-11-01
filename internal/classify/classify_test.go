package classify

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/AnnaofArendelle/llm-speedtest/internal/modelid"
)

type realError struct {
	ID           string `json:"id"`
	OwnedBy      string `json:"owned_by"`
	Status       int    `json:"status"`
	InnerStatus  int    `json:"inner_status"`
	Conn         string `json:"conn"`
	Message      string `json:"message"`
	ExpectClass  string `json:"expect_class"`
	ExpectReason string `json:"expect_reason"`
	ExpectInner  int    `json:"expect_inner"`
}

// TestRealErrors replays every distinct error observed during the reference
// 9router scan and checks that the Go port reproduces the reference
// classification exactly.
func TestRealErrors(t *testing.T) {
	raw, err := os.ReadFile(filepath.Join("..", "..", "testdata", "real_errors.json"))
	if err != nil {
		t.Fatal(err)
	}
	var samples []realError
	if err := json.Unmarshal(raw, &samples); err != nil {
		t.Fatal(err)
	}
	if len(samples) < 100 {
		t.Fatalf("expected a large sample set, got %d", len(samples))
	}
	mismatch := 0
	for _, s := range samples {
		name := modelid.Name(s.ID, s.OwnedBy)
		human, code, _ := DeepUnwrap(s.Message, s.InnerStatus)
		cls := Class(s.Status, code, human)
		if s.Status == 200 && (cls == "bad_request" || cls == "unknown" || cls == "http_200") {
			cls = "error_in_200"
		}
		reason := Refine(cls, human, name)
		if cls != s.ExpectClass || reason != s.ExpectReason || code != s.ExpectInner {
			mismatch++
			if mismatch <= 15 {
				t.Errorf("%s: got class=%s reason=%s inner=%d, want class=%s reason=%s inner=%d\n  msg=%q", s.ID, cls, reason, code, s.ExpectClass, s.ExpectReason, s.ExpectInner, s.Message)
			}
		}
	}
	if mismatch > 0 {
		t.Fatalf("%d/%d samples mismatch", mismatch, len(samples))
	}
}

func TestExtractError(t *testing.T) {
	cases := []struct {
		status int
		body   string
		want   string
	}{
		{400, `{"error":{"message":"Missing model","type":"invalid_request_error"}}`, "Missing model"},
		{404, `{"error":"No active credentials for provider: bai"}`, "No active credentials for provider: bai"},
		{500, `{"message":"boom"}`, "boom"},
		{503, `{"detail":"nope"}`, "nope"},
		{502, `<html><title>502 Bad Gateway</title></html>`, "<html><title>502 Bad Gateway</title></html>"},
		{504, ``, "HTTP 504"},
		{400, `{"error":{"message":{"nested":1}}}`, `{"nested":1}`},
	}
	for _, c := range cases {
		if got := ExtractError(c.status, []byte(c.body)); got != c.want {
			t.Errorf("ExtractError(%d,%q)=%q want %q", c.status, c.body, got, c.want)
		}
	}
}

func TestUnwrapAndExtract(t *testing.T) {
	conn, code, human := Unwrap(`[openai-compatible-chat-abc/deepseek-v4-flash] [429]: {"error":{"message":"Insufficient credits. Please top up."}} (reset after 20s)`)
	if conn != "openai-compatible-chat-abc/deepseek-v4-flash" || code != 429 || human != "Insufficient credits. Please top up. (reset after 20s)" {
		t.Fatalf("unwrap = %q %d %q", conn, code, human)
	}
	_, code, human = Unwrap(`[400]: {"error":{"type":"server_error","message":"Upstream request failed: Model is unavailable."}}`)
	if code != 400 || human != "Upstream request failed: Model is unavailable." {
		t.Fatalf("wrapB = %d %q", code, human)
	}
	conn, code, human = Unwrap(`[kiro/claude-sonnet-4.5] {"message":"The bearer token included in the request is invalid."}`)
	if conn != "kiro/claude-sonnet-4.5" || code != 0 || human != "The bearer token included in the request is invalid." {
		t.Fatalf("wrapC = %q %d %q", conn, code, human)
	}
	if got := extractMsg(`<html><head><title>401 Authorization Required</title></head></html>`); got != "HTML: 401 Authorization Required" {
		t.Fatalf("html = %q", got)
	}
	if got := extractMsg(`{"status":404,"title":"Not Found","detail":"Function x not found"}`); got != "Function x not found" {
		t.Fatalf("detail wins = %q", got)
	}
	if got := extractMsg(`{"status":404,"title":"Not Found"}`); got != "Not Found" {
		t.Fatalf("title only = %q", got)
	}
	if got := extractMsg(`[{"message":"first"}]`); got != "first" {
		t.Fatalf("list = %q", got)
	}
	if got := extractMsg(`{"error":{"code":"x"}}`); got != `{"code":"x"}` {
		t.Fatalf("dict fallback = %q", got)
	}
	if got := extractMsg(`{"object":"error","message":"Rate limit exceeded"} (reset after 4s)`); got != "Rate limit exceeded (reset after 4s)" {
		t.Fatalf("tail = %q", got)
	}
	// nested gateway: three levels
	human, code, conn = DeepUnwrap(`[occ-1/deepseek-v4-flash] [502]: [occ-2/deepseek-v4-flash] [403]: {"error":{"message":"No permission to access group Discount"}}`, 502)
	if code != 403 || human != "No permission to access group Discount" || conn != "occ-1/deepseek-v4-flash" {
		t.Fatalf("deep = %q %d %q", human, code, conn)
	}
	// plain text is returned unchanged
	human, code, _ = DeepUnwrap("just text", 0)
	if human != "just text" || code != 0 {
		t.Fatal("plain")
	}
}

func TestResetHintAndClean(t *testing.T) {
	d, ok := ResetHint("Insufficient funds (reset after 1h 2m 3s)")
	if !ok || d != time.Hour+2*time.Minute+3*time.Second {
		t.Fatalf("reset = %v %v", d, ok)
	}
	if _, ok := ResetHint("nothing"); ok {
		t.Fatal("no hint")
	}
	d, _ = ResetHint("x (reset after 0s)")
	if d != 0 {
		t.Fatal("zero")
	}
	got := CleanMessage("No permission (request id: abc) (reset after 1m 48s) (reset after 30s)")
	if got != "No permission (reset after 30s)" {
		t.Fatalf("clean = %q", got)
	}
}

func TestEvaluate(t *testing.T) {
	o := Evaluate(200, []byte(`{"choices":[{"message":{"content":"OK"}}]}`), "x")
	if !o.OK || o.Reason != "ok" {
		t.Fatalf("success = %+v", o)
	}
	o = Evaluate(200, []byte(`{"error":{"message":"weird"}}`), "x")
	if o.OK || o.Class != "error_in_200" || o.Reason != "error_in_200" {
		t.Fatalf("error in 200 = %+v", o)
	}
	o = Evaluate(0, []byte("client timeout after 60s"), "x")
	if o.Class != "timeout" || o.Reason != "timeout" {
		t.Fatalf("timeout = %+v", o)
	}
	o = Evaluate(0, []byte("dial tcp: connection refused"), "x")
	if o.Class != "network" {
		t.Fatalf("network = %+v", o)
	}
	o = Evaluate(503, []byte(`{"error":{"message":"[occ/flux-1] [400]: {\"error\":{\"message\":\"The model 'flux-1' does not exist or is not supported for chat completions.\"}}"}}`), "flux-1")
	if o.Reason != "not_chat_model_inferred" || o.Inner != 400 || o.Conn != "occ/flux-1" {
		t.Fatalf("inferred = %+v", o)
	}
	if !IsTransient("rate_limit") || IsTransient("auth") || !IsProviderLevel("auth") || IsProviderLevel("timeout") {
		t.Fatal("flags")
	}
	if Info("auth").Name == "" || Info("nope").Name != "nope" || len(All()) != len(Order) {
		t.Fatal("info")
	}
}

func FuzzDeepUnwrap(f *testing.F) {
	f.Add(`[a/b] [429]: {"error":{"message":"x"}}`, 503)
	f.Add(`[400]: [500]: [a/b] [403]: {"detail":"d"}`, 0)
	f.Add(`<html><title>t</title>`, 502)
	f.Add(`{"error":[1,2]}`, 400)
	f.Fuzz(func(t *testing.T, msg string, inner int) {
		human, code, _ := DeepUnwrap(msg, inner)
		if code < 0 {
			t.Fatal("negative code")
		}
		h2, c2, _ := DeepUnwrap(human, code)
		if h2 != human || c2 != code {
			// A second pass must be a fixed point.
			t.Fatalf("not idempotent: %q/%d -> %q/%d", human, code, h2, c2)
		}
		_ = Refine(Class(200, code, human), human, msg)
	})
}
