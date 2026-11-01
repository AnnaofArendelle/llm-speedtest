package apiclient

import (
	"bufio"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"time"
	"unicode"
)

// ChunkTiming records one streamed event that carried generated text.
type ChunkTiming struct {
	At     time.Time `json:"at"`
	Chars  int       `json:"chars"`
	Tokens int       `json:"tokens"` // estimated tokens in this chunk
}

// StreamResult is the outcome of a streaming chat request with a complete
// timeline. All timestamps are zero when the corresponding event never
// happened.
type StreamResult struct {
	Status        int           `json:"status"`
	Headers       http.Header   `json:"-"`
	Body          []byte        `json:"-"` // error body for non-200 or transport description
	SentAt        time.Time     `json:"sent_at"`
	HeadersAt     time.Time     `json:"headers_at"`
	FirstByteAt   time.Time     `json:"first_byte_at"`
	FirstTokenAt  time.Time     `json:"first_token_at"`
	SecondTokenAt time.Time     `json:"second_token_at"`
	DoneAt        time.Time     `json:"done_at"`
	Chunks        []ChunkTiming `json:"chunks"`
	Content       string        `json:"content"`
	Reasoning     string        `json:"reasoning,omitempty"`
	Model         string        `json:"model,omitempty"`
	FinishReason  string        `json:"finish_reason,omitempty"`
	Usage         *Usage        `json:"usage,omitempty"`
	OutputTokens  int           `json:"output_tokens"` // usage.completion_tokens or an estimate
	TokensExact   bool          `json:"tokens_exact"`  // whether OutputTokens came from usage
	ToolCalls     int           `json:"tool_calls"`
	Err           error         `json:"-"`
	Stalled       bool          `json:"stalled"`
}

// OK reports whether the stream completed successfully.
func (r StreamResult) OK() bool {
	return r.Status == 200 && r.Err == nil && !r.FirstTokenAt.IsZero()
}

// TTFT is the time to first generated token.
func (r StreamResult) TTFT() time.Duration {
	if r.FirstTokenAt.IsZero() {
		return 0
	}
	return r.FirstTokenAt.Sub(r.SentAt)
}

// E2E is the end-to-end latency.
func (r StreamResult) E2E() time.Duration {
	if r.DoneAt.IsZero() {
		return 0
	}
	return r.DoneAt.Sub(r.SentAt)
}

// TTST is the gap between the first and second generated token.
func (r StreamResult) TTST() time.Duration {
	if r.SecondTokenAt.IsZero() || r.FirstTokenAt.IsZero() {
		return 0
	}
	return r.SecondTokenAt.Sub(r.FirstTokenAt)
}

// TPOT is the time per output token excluding the first token
// ((done - first_token) / (N - 1)); 0 when fewer than two tokens arrived.
func (r StreamResult) TPOT() time.Duration {
	if r.OutputTokens < 2 || r.FirstTokenAt.IsZero() || r.DoneAt.IsZero() {
		return 0
	}
	return r.DoneAt.Sub(r.FirstTokenAt) / time.Duration(r.OutputTokens-1)
}

// TokensPerSecond is the decode throughput ((N - 1) / (done - first_token)).
func (r StreamResult) TokensPerSecond() float64 {
	if r.OutputTokens < 2 || r.FirstTokenAt.IsZero() || r.DoneAt.IsZero() {
		return 0
	}
	d := r.DoneAt.Sub(r.FirstTokenAt).Seconds()
	if d <= 0 {
		return 0
	}
	return float64(r.OutputTokens-1) / d
}

// StreamOptions tunes streaming behaviour.
type StreamOptions struct {
	// StallTimeout aborts the stream when no bytes arrive for this long.
	StallTimeout time.Duration
}

type firstByteReader struct {
	r    io.Reader
	seen bool
	at   *time.Time
}

func (f *firstByteReader) Read(p []byte) (int, error) {
	n, err := f.r.Read(p)
	if n > 0 && !f.seen {
		f.seen = true
		*f.at = time.Now()
	}
	return n, err
}

// ChatStream performs a streaming chat completion and records the timeline.
func (c *Client) ChatStream(ctx context.Context, req ChatRequest, opts StreamOptions) StreamResult {
	out := StreamResult{SentAt: time.Now()}
	body, err := req.encode(true)
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
	hreq.Header.Set("Accept", "text/event-stream")
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	hreq = hreq.WithContext(ctx)
	resp, err := c.hc.Do(hreq)
	if err != nil {
		out.Err = err
		out.Body = []byte(transportMessage(err, ctx, time.Since(out.SentAt)))
		return out
	}
	defer resp.Body.Close()
	out.HeadersAt = time.Now()
	out.Status = resp.StatusCode
	out.Headers = resp.Header
	if resp.StatusCode != 200 {
		raw, _ := io.ReadAll(io.LimitReader(resp.Body, maxErrorBody))
		out.Body = raw
		out.DoneAt = time.Now()
		return out
	}
	var stall *time.Timer
	if opts.StallTimeout > 0 {
		stall = time.AfterFunc(opts.StallTimeout, func() {
			out.Stalled = true
			cancel()
		})
		defer stall.Stop()
	}
	fbr := &firstByteReader{r: resp.Body, at: &out.FirstByteAt}
	ct := resp.Header.Get("Content-Type")
	if !strings.Contains(ct, "text/event-stream") {
		// Some servers ignore stream=true and answer with a JSON completion.
		raw, rerr := io.ReadAll(io.LimitReader(fbr, maxModelsBody))
		out.DoneAt = time.Now()
		if rerr != nil {
			out.Err = rerr
			out.Status = 0
			out.Body = []byte(transportMessage(rerr, ctx, time.Since(out.SentAt)))
			return out
		}
		var cr ChatResult
		parseCompletion(raw, &cr)
		out.Body = capBody(raw, maxErrorBody)
		out.Content, out.Reasoning, out.Model, out.FinishReason, out.Usage = cr.Content, cr.Reasoning, cr.Model, cr.FinishReason, cr.Usage
		if out.Content != "" || out.Reasoning != "" || len(cr.ToolCalls) > 0 {
			out.FirstTokenAt = out.FirstByteAt
			out.Chunks = []ChunkTiming{{At: out.FirstByteAt, Chars: len(out.Content), Tokens: EstimateTokens(out.Content + out.Reasoning)}}
		}
		out.ToolCalls = len(cr.ToolCalls)
		finalizeTokens(&out)
		return out
	}
	rd := bufio.NewReaderSize(fbr, 64*1024)
	var content, reasoning strings.Builder
	var dataBuf strings.Builder
	tokenEvents := 0
	toolCalls := map[int]bool{}
	flushEvent := func() {
		data := strings.TrimSpace(dataBuf.String())
		dataBuf.Reset()
		if data == "" {
			return
		}
		if data == "[DONE]" {
			return
		}
		var ev struct {
			Model   string `json:"model"`
			Choices []struct {
				FinishReason *string `json:"finish_reason"`
				Delta        struct {
					Content   any    `json:"content"`
					Reasoning string `json:"reasoning_content"`
					Reason2   string `json:"reasoning"`
					ToolCalls []struct {
						Index int `json:"index"`
					} `json:"tool_calls"`
				} `json:"delta"`
			} `json:"choices"`
			Usage *Usage `json:"usage"`
		}
		if err := json.Unmarshal([]byte(data), &ev); err != nil {
			return
		}
		if ev.Model != "" && out.Model == "" {
			out.Model = ev.Model
		}
		if ev.Usage != nil && (ev.Usage.CompletionTokens > 0 || ev.Usage.PromptTokens > 0) {
			out.Usage = ev.Usage
		}
		for _, ch := range ev.Choices {
			text := contentToString(ch.Delta.Content)
			rs := ch.Delta.Reasoning
			if rs == "" {
				rs = ch.Delta.Reason2
			}
			for _, tc := range ch.Delta.ToolCalls {
				toolCalls[tc.Index] = true
			}
			generated := len(text) > 0 || len(rs) > 0 || len(ch.Delta.ToolCalls) > 0
			if generated {
				now := time.Now()
				tokenEvents++
				switch tokenEvents {
				case 1:
					out.FirstTokenAt = now
				case 2:
					out.SecondTokenAt = now
				}
				if content.Len()+reasoning.Len() < maxStreamText {
					content.WriteString(text)
					reasoning.WriteString(rs)
				}
				if len(out.Chunks) < 200000 {
					out.Chunks = append(out.Chunks, ChunkTiming{At: now, Chars: len(text) + len(rs), Tokens: EstimateTokens(text + rs)})
				}
			}
			if ch.FinishReason != nil && *ch.FinishReason != "" {
				out.FinishReason = *ch.FinishReason
			}
		}
	}
	for {
		line, err := rd.ReadString('\n')
		if stall != nil && len(line) > 0 {
			stall.Reset(opts.StallTimeout)
		}
		if len(line) > 0 {
			l := strings.TrimRight(line, "\r\n")
			switch {
			case l == "":
				flushEvent()
			case strings.HasPrefix(l, "data:"):
				dataBuf.WriteString(strings.TrimPrefix(strings.TrimPrefix(l, "data:"), " "))
				dataBuf.WriteString("\n")
			}
			// any other line (": keep-alive" comments, "event:", "id:") carries no payload
		}
		if err != nil {
			flushEvent()
			if err != io.EOF {
				if out.Stalled {
					out.Err = context.DeadlineExceeded
					out.Body = []byte("stream stalled: no data for " + opts.StallTimeout.String())
				} else if ctx.Err() != nil {
					out.Err = ctx.Err()
					out.Body = []byte(transportMessage(ctx.Err(), ctx, time.Since(out.SentAt)))
				} else {
					out.Err = err
					out.Body = []byte("stream read error: " + err.Error())
				}
				if out.FirstTokenAt.IsZero() {
					out.Status = 0
				}
			}
			break
		}
	}
	out.DoneAt = time.Now()
	out.Content = content.String()
	out.Reasoning = reasoning.String()
	out.ToolCalls = len(toolCalls)
	finalizeTokens(&out)
	return out
}

func finalizeTokens(out *StreamResult) {
	if out.Usage != nil && out.Usage.CompletionTokens > 0 {
		out.OutputTokens = out.Usage.CompletionTokens
		out.TokensExact = true
		return
	}
	n := 0
	for _, c := range out.Chunks {
		n += c.Tokens
	}
	if n == 0 && (out.Content != "" || out.Reasoning != "") {
		n = EstimateTokens(out.Content + out.Reasoning)
	}
	if n < len(out.Chunks) {
		n = len(out.Chunks)
	}
	out.OutputTokens = n
}

// EstimateTokens approximates the token count of text when the endpoint does
// not report usage: CJK characters count as one token each, other
// characters as a quarter token (rounded up, minimum 1 for non-empty text).
func EstimateTokens(s string) int {
	if s == "" {
		return 0
	}
	cjk, other := 0, 0
	for _, r := range s {
		if unicode.Is(unicode.Han, r) || unicode.Is(unicode.Hangul, r) || unicode.Is(unicode.Hiragana, r) || unicode.Is(unicode.Katakana, r) {
			cjk++
		} else {
			other++
		}
	}
	n := cjk + (other+3)/4
	if n < 1 {
		n = 1
	}
	return n
}
