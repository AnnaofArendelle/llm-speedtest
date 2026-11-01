// Package classify turns HTTP probe results into a stable failure taxonomy.
// Classification is status-code first and keyword second; unknown shapes fall
// into "unknown" while always preserving the original message.
package classify

import (
	"fmt"
	"regexp"
	"strings"
)

// Outcome is the classified result of a single probe.
type Outcome struct {
	OK      bool   `json:"ok"`
	Class   string `json:"class"`  // coarse class from status + keywords
	Reason  string `json:"reason"` // refined reason code (see Reasons)
	HTTP    int    `json:"http"`   // status returned by the endpoint
	Inner   int    `json:"inner"`  // deepest upstream status embedded in the message (0 if none)
	Conn    string `json:"conn"`   // gateway connection identifier when present
	Message string `json:"message"`
}

var nonChatRE = regexp.MustCompile(`(?i)(embed|rerank|\btts\b|tts-|whisper|transcribe|speech|audio|voice|eleven_|elevenlabs|scribe|flux|sdxl|` +
	`stable-diffusion|\bsd3|dall-e|dalle|imagen|ideogram|recraft|hidream|nano-banana|cogview|z-image|p-image|` +
	`seedream|seedance|dreamina|kling|\bveo\b|veo-|vidu|\bsora|wan2|video|image|illustrious|pony|animagine|` +
	`dreamshaper|majicmix|chroma|lucid-origin|omnivoice|voyage|jina|\bbge\b|bge-|minilm|\be5-|gte-|\bclip|siglip|` +
	`moderation|realistic|anything-v|counterfeit|juggernaut|dreamlike|playground-v|kandinsky|riverflow|\bltx|` +
	`hunyuan-video|hailuo|music|\bsuno|\budio|tunix|photon|recraft|text-embedding)`)

// LooksNonChat reports whether a model name suggests a non-chat modality
// (image, video, audio, embedding, ...). It is only used to refine ambiguous
// "model does not exist or is not supported" errors.
func LooksNonChat(name string) bool { return nonChatRE.MatchString(name) }

func has(m string, keys ...string) bool {
	for _, k := range keys {
		if strings.Contains(m, k) {
			return true
		}
	}
	return false
}

// Class derives the coarse failure class from the outer HTTP status, the
// deepest embedded upstream status and the unwrapped message. outer == 0
// means the request never produced a response (timeout / network).
func Class(outer, inner int, msg string) string {
	m := strings.ToLower(msg)
	if outer == 0 {
		if has(m, "timeout", "timed out", "deadline exceeded") {
			return "timeout"
		}
		return "network"
	}
	if has(m, "no active credentials") {
		return "no_credentials"
	}
	s := inner
	if s == 0 {
		s = outer
	}
	switch {
	case has(m, "not a chat model", "call /v1/"):
		return "not_chat_model"
	case has(m, "no longer available", "is no longer", "decommissioned", "deprecated", "has been retired", "retired"):
		return "model_retired"
	case s == 429 || has(m, "rate limit", "rate_limit", "ratelimit", "too many requests", "throttl", "usage limit reached", "limit reached", "requests per", "concurrency limit"):
		return "rate_limit"
	case s == 402 || has(m, "insufficient", "balance", "credit", "billing", "payment required", "out of funds", "欠费", "余额", "not included in your free usage", "requires more credits", "top up", "purchase", "quota exceeded", "exceeded your"):
		return "quota_billing"
	case s == 401 || s == 403 || has(m, "unauthorized", "invalid api key", "invalid_api_key", "authentication", "forbidden", "bearer token", "invalid token", "api key", "api-key", "permission denied", "not authorized", "access denied", "invalid key"):
		return "auth"
	case s == 404 || s == 410 || has(m, "model is unavailable", "model unavailable", "model not found", "does not exist", "not exist", "no such model", "unknown model", "invalid model", "model_not_found", "unsupported model", "not supported", "is not available", "could not find model", "no endpoints found", "check the model", "not found", "unsupported"):
		return "model_unavailable"
	case has(m, "empty turn", "empty response", "returned empty", "no content"):
		return "empty_response"
	case has(m, "timeout", "timed out", "deadline", "fetch failed", "econnreset", "econnrefused", "socket hang up", "network error", "connection error", "unreachable", "getaddrinfo"):
		return "upstream_network"
	case s == 400 || s == 422:
		return "bad_request"
	case s == 405:
		return "method_not_allowed"
	case s >= 500 && s < 600:
		return "upstream_5xx"
	case s >= 400:
		return fmt.Sprintf("http_%d", s)
	}
	return "unknown"
}

var (
	strongBilling = []string{"insufficient credits", "insufficient funds", "insufficient tokens", "out of credits", "余额不足", "balance is at",
		"balance is empty", "requires more credits", "exhausted its credits", "credit limit is insufficient", "gift balance",
		"预扣费", "billing problem", "remaining credit", "insufficient balance", "not enough credits", "no credits"}
	planWords = []string{"paid model", "free plan only", "free plan is", "only allows free", "not included in your free", "purchase a plan",
		"paying customers", "premium model", "requires an active paid plan", "subscription tier", "仅可使用免费模型", "没有可用模型",
		"membership", "upgrade your plan", "requires real deposited balance", "pay-as-you-go", "free api keys may only use"}
	exhausted = []string{"free usage quota for this model is used up", "额度已用完", "usage limit reached", "free usage limit", "exceeded your current quota",
		"quota exceeded", "gift", "赠送", "daily limit", "free requests", "quota has been", "reached its daily", "额度不足"}
	weakBilling = []string{"top up", "充值", "add credits", "add usage credits", "please add credit"}
)

// Refine maps a coarse class plus the message (and the model name for
// modality inference) onto the final reason code.
func Refine(cls, msg, name string) string {
	m := strings.ToLower(msg)
	n := strings.ToLower(name)
	switch cls {
	case "ok", "no_credentials", "network", "exception", "untested":
		return cls
	}
	switch {
	case has(m, "image generation model", "does not support /v1/chat", "use /v1/images", "not a chat model", "does not support chat completions",
		"only serves post /v1/decisions", "answers only /v1/decisions", "use /v1/embeddings", "use /v1/audio", "is an embedding model", "call /v1/"):
		return "not_chat_model"
	case has(m, "requires terms acceptance", "accept the terms"):
		return "terms_required"
	case has(m, "thinking level", "retry with other thinking level", "unsupported parameter", "invalid parameter", "parameter is invalid", "not support the input"):
		return "param_incompatible"
	case has(m, "no available capacity", "no capacity"):
		return "no_capacity"
	case has(m, "无可用渠道", "no available channel", "no available model deployment", "no available provider"):
		return "no_channel"
	case has(m, exhausted...):
		return "quota_exhausted"
	case has(m, strongBilling...):
		return "quota_billing"
	case has(m, planWords...):
		return "plan_restriction"
	case has(m, weakBilling...):
		return "quota_billing"
	case cls == "rate_limit":
		return "rate_limit"
	case cls == "model_retired":
		return cls
	case has(m, "no model found", "model not found", "does not exist", "not exist", "unknown model", "not in our catalog", "not offered here",
		"unsupported model", "check the model you provided", "not supported", "temporarily disabled", "page not found", "is not available",
		"model is unavailable", "not found for account", "model is not found", "not found"):
		if LooksNonChat(n) {
			return "not_chat_model_inferred"
		}
		return "model_unavailable"
	case has(m, "do request failed", "fetch failed", "econnreset", "socket hang up", "bad gateway", "econnrefused", "tls"):
		return "upstream_network"
	case (cls == "model_unavailable" || cls == "bad_request") && LooksNonChat(n):
		return "not_chat_model_inferred"
	}
	return cls
}

// Evaluate classifies a complete HTTP exchange. status==0 denotes a transport
// failure whose description is passed in body (e.g. "client timeout after
// 60s"). name is the cleaned model name used for modality inference.
func Evaluate(status int, body []byte, name string) Outcome {
	o := Outcome{HTTP: status}
	if status == 200 && looksLikeChatSuccess(body) {
		o.OK = true
		o.Class, o.Reason = "ok", "ok"
		return o
	}
	var raw string
	if status == 0 {
		raw = string(body)
	} else {
		raw = ExtractError(status, body)
	}
	human, code, conn := DeepUnwrap(raw, 0)
	o.Inner, o.Conn, o.Message = code, conn, human
	o.Class = Class(status, code, human)
	if status == 200 && (o.Class == "bad_request" || o.Class == "unknown" || o.Class == "http_200") {
		o.Class = "error_in_200"
	}
	o.Reason = Refine(o.Class, human, name)
	return o
}

// looksLikeChatSuccess reports whether a 200 body is a chat completion (as
// opposed to a 200 carrying an error envelope).
func looksLikeChatSuccess(body []byte) bool {
	s := strings.TrimSpace(string(body))
	if s == "" {
		return false
	}
	if !strings.Contains(s, `"choices"`) {
		return false
	}
	// A body with an error envelope and no choices is not a success; a body
	// with choices is treated as success even when it also has "error": null.
	return true
}

// IsTransient reports whether failures with this reason are worth retrying.
func IsTransient(reason string) bool { return transient[reason] }

// IsProviderLevel reports whether the reason is an account/provider problem
// rather than a property of the model itself.
func IsProviderLevel(reason string) bool { return providerLevel[reason] }
