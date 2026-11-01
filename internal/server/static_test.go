package server

import (
	"net/http"
	"strings"
	"testing"

	"github.com/AnnaofArendelle/llm-speedtest/web"
)

// TestStaticUI checks that the embedded single-page app (or the placeholder
// when the frontend has not been built) is served for every non-API path.
func TestStaticUI(t *testing.T) {
	e := newEnv(t, "")
	for _, p := range []string{"/", "/runs/does-not-exist", "/settings"} {
		code, body := e.do(http.MethodGet, p, nil)
		if code != http.StatusOK {
			t.Fatalf("GET %s: status %d", p, code)
		}
		if web.Built() {
			if !strings.Contains(string(body), `<div id="root">`) {
				t.Fatalf("GET %s: embedded index not served: %.200s", p, body)
			}
		} else if !strings.Contains(string(body), "LLM Speedtest API is running") {
			t.Fatalf("GET %s: placeholder not served: %.200s", p, body)
		}
	}
	if code, _ := e.do(http.MethodGet, "/api/nope", nil); code != http.StatusNotFound {
		t.Fatalf("unknown API path: status %d", code)
	}
	if web.Built() {
		code, body := e.do(http.MethodGet, "/favicon.svg", nil)
		if code != http.StatusOK || !strings.Contains(string(body), "<svg") {
			t.Fatalf("favicon: status %d", code)
		}
	}
}
