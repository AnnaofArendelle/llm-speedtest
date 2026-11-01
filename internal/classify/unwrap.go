package classify

import (
	"encoding/json"
	"regexp"
	"strconv"
	"strings"
	"time"
	"unicode/utf8"
)

// Gateways such as 9router wrap upstream failures as
//
//	[<connection>/<model>] [<upstream status>]: <upstream body>
//
// and nest them when the upstream is itself a gateway. The patterns below
// are structural (no vendor-specific vocabulary) and are safe to apply to any
// error text: when nothing matches the message is returned unchanged.
var (
	wrapA   = regexp.MustCompile(`(?s)^\s*\[([^\]\n]+)\]\s*\[(\d{3})\]:\s*(.*)$`)
	wrapB   = regexp.MustCompile(`(?s)^\s*\[(\d{3})\]:\s*(.*)$`)
	wrapC   = regexp.MustCompile(`(?s)^\s*\[([^\]\n]+/[^\]\n]*)\]\s*(.*)$`)
	wrapBPr = regexp.MustCompile(`^\s*\[(\d{3})\]:`)
	tagRE   = regexp.MustCompile(`<[^>]+>`)
	titleRE = regexp.MustCompile(`(?is)<title>(.*?)</title>`)
	resetRE = regexp.MustCompile(`\(reset after ([^)]*)\)`)
	reqIDRE = regexp.MustCompile(`\s*\(request id: [^)]*\)`)
	durRE   = regexp.MustCompile(`(\d+)\s*([hms])`)
)

// ExtractError derives the human readable error message from an HTTP error
// response body. It understands the OpenAI error envelope ({"error":
// {"message": ...}}), bare {"error": "..."} / {"message": ...} / {"detail":
// ...} shapes and falls back to the raw text or "HTTP <status>".
func ExtractError(status int, body []byte) string {
	text := strings.TrimSpace(string(body))
	var j map[string]any
	if len(text) > 0 && text[0] == '{' {
		if err := json.Unmarshal([]byte(text), &j); err == nil {
			var m any
			if e, ok := j["error"]; ok && e != nil {
				if em, ok := e.(map[string]any); ok {
					m = em["message"]
				} else {
					m = e
				}
			} else if v, ok := j["message"]; ok && v != nil && v != "" {
				m = v
			} else if v, ok := j["detail"]; ok && v != nil && v != "" {
				m = v
			}
			if m != nil {
				return anyToString(m)
			}
		}
	}
	if text == "" {
		return "HTTP " + strconv.Itoa(status)
	}
	return text
}

func anyToString(v any) string {
	switch t := v.(type) {
	case string:
		return t
	case nil:
		return ""
	default:
		b, err := json.Marshal(t)
		if err != nil {
			return ""
		}
		return string(b)
	}
}

// extractMsg converts an (upstream) error body into a short human message.
// It mirrors the behaviour of the reference implementation so that the
// recorded samples in testdata classify identically.
func extractMsg(body string) string {
	b := strings.TrimSpace(body)
	if b == "" {
		return ""
	}
	tail := ""
	var parsed any
	if b[0] == '{' || b[0] == '[' {
		if err := json.Unmarshal([]byte(b), &parsed); err != nil {
			parsed = nil
			if end := strings.LastIndex(b, "}"); end > 0 {
				var p2 any
				if err2 := json.Unmarshal([]byte(b[:end+1]), &p2); err2 == nil {
					parsed = p2
					tail = strings.TrimSpace(b[end+1:])
				}
			}
		}
	}
	switch j := parsed.(type) {
	case map[string]any:
		var out string
		found := false
		switch e := j["error"].(type) {
		case map[string]any:
			for _, k := range []string{"message", "msg", "detail", "description"} {
				if v, ok := e[k]; ok && truthy(v) {
					out, found = anyToString(v), true
					break
				}
			}
			if !found {
				raw, _ := json.Marshal(e)
				out = string(raw)
			}
		case string:
			out = e
		default:
			for _, k := range []string{"message", "detail", "msg", "error_description"} {
				if v, ok := j[k]; ok && truthy(v) {
					out, found = anyToString(v), true
					break
				}
			}
			if !found {
				if title, ok := j["title"]; ok && truthy(title) {
					detail := anyToString(j["detail"])
					out = strings.Trim(anyToString(title)+": "+detail, ": ")
					found = true
				}
			}
			if !found {
				out = b
			}
		}
		if tail != "" {
			out += " " + tail
		}
		return truncRunes(strings.TrimSpace(out), 500)
	case []any:
		if len(j) > 0 {
			if first, ok := j[0].(map[string]any); ok {
				raw, _ := json.Marshal(first)
				return extractMsg(string(raw))
			}
		}
	}
	low := strings.ToLower(b)
	if strings.Contains(low, "<html") || strings.Contains(low, "<title") {
		if m := titleRE.FindStringSubmatch(b); m != nil {
			return truncRunes("HTML: "+strings.TrimSpace(m[1]), 300)
		}
		return truncRunes("HTML: "+truncRunes(strings.TrimSpace(tagRE.ReplaceAllString(b, " ")), 120), 300)
	}
	return truncRunes(b, 500)
}

func truthy(v any) bool {
	switch t := v.(type) {
	case nil:
		return false
	case string:
		return t != ""
	case bool:
		return t
	case float64:
		return t != 0
	case []any:
		return len(t) > 0
	case map[string]any:
		return len(t) > 0
	}
	return true
}

func truncRunes(s string, n int) string {
	if utf8.RuneCountInString(s) <= n {
		return s
	}
	r := []rune(s)
	return string(r[:n])
}

// Unwrap peels one level of gateway wrapping from an error message and
// returns the connection identifier (if any), the embedded upstream status
// (0 when absent) and the human readable message.
func Unwrap(msg string) (conn string, code int, human string) {
	body := msg
	if m := wrapA.FindStringSubmatch(msg); m != nil {
		conn = strings.TrimSpace(m[1])
		code, _ = strconv.Atoi(m[2])
		body = m[3]
	} else if m := wrapB.FindStringSubmatch(msg); m != nil {
		code, _ = strconv.Atoi(m[1])
		body = m[2]
	} else if m := wrapC.FindStringSubmatch(msg); m != nil {
		conn = strings.TrimSpace(m[1])
		body = m[2]
	}
	human = extractMsg(body)
	if m2 := wrapB.FindStringSubmatch(human); m2 != nil {
		if code == 0 {
			code, _ = strconv.Atoi(m2[1])
		}
		human = extractMsg(m2[2])
	}
	return conn, code, human
}

// DeepUnwrap repeatedly unwraps nested gateway envelopes (at most three
// levels) and returns the innermost message, the deepest embedded status
// (falling back to inner) and the outermost connection identifier.
func DeepUnwrap(msg string, inner int) (human string, code int, conn string) {
	code = inner
	m := msg
	for i := 0; i < 3; i++ {
		if !wrapA.MatchString(m) && !wrapBPr.MatchString(m) {
			break
		}
		c, cd, h := Unwrap(m)
		if conn == "" {
			conn = c
		}
		if cd != 0 {
			code = cd
		}
		if h == "" || h == m {
			break
		}
		m = h
	}
	return m, code, conn
}

// ResetHint parses a gateway cooldown hint such as "(reset after 1h 2m 3s)"
// and returns the duration.
func ResetHint(msg string) (time.Duration, bool) {
	m := resetRE.FindAllStringSubmatch(msg, -1)
	if m == nil {
		return 0, false
	}
	spec := strings.TrimSpace(m[len(m)-1][1])
	var total time.Duration
	matched := false
	for _, part := range durRE.FindAllStringSubmatch(spec, -1) {
		n, _ := strconv.Atoi(part[1])
		matched = true
		switch part[2] {
		case "h":
			total += time.Duration(n) * time.Hour
		case "m":
			total += time.Duration(n) * time.Minute
		case "s":
			total += time.Duration(n) * time.Second
		}
	}
	return total, matched
}

// CleanMessage removes volatile decorations (request ids, repeated cooldown
// hints) from a message for display, keeping the last cooldown hint.
func CleanMessage(msg string) string {
	resets := resetRE.FindAllString(msg, -1)
	out := reqIDRE.ReplaceAllString(msg, "")
	if len(resets) > 0 {
		out = strings.TrimSpace(resetRE.ReplaceAllString(out, "")) + " " + resets[len(resets)-1]
	}
	return strings.TrimSpace(out)
}
