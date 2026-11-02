// Package core defines the data records shared by the engines, the store,
// the analysis layer and the API.
package core

import (
	"encoding/json"
	"time"
)

// RunKind enumerates the kinds of runs.
const (
	KindScan       = "scan"
	KindLatency    = "latency"
	KindLoad       = "load"
	KindCapability = "capability"
	KindQuality    = "quality"
	KindSuite      = "suite"
)

// Run statuses.
const (
	StatusPending     = "pending"
	StatusDiscovering = "discovering"
	StatusRunning     = "running"
	StatusDone        = "done"
	StatusCancelled   = "cancelled"
	StatusFailed      = "failed"
	StatusInterrupted = "interrupted"
)

// RunModel is a model that took part in a run, with its resolved provider
// and family.
type RunModel struct {
	ID             string          `json:"id"`
	OwnedBy        string          `json:"owned_by"`
	Upstream       string          `json:"upstream"` // id without the provider prefix
	ContextLength  int             `json:"context_length,omitempty"`
	MaxOutput      int             `json:"max_output,omitempty"`
	ProviderKey    string          `json:"provider_key"`
	ProviderName   string          `json:"provider_name"`
	FamilyKey      string          `json:"family_key"`
	FamilyCanon    string          `json:"family_canon"`
	FamilySource   string          `json:"family_source"`
	FamilyEvidence string          `json:"family_evidence,omitempty"`
	Raw            json.RawMessage `json:"raw,omitempty"`
}

// Attempt is one availability probe of one model in one round.
type Attempt struct {
	ModelID          string        `json:"model_id"`
	Round            int           `json:"round"`
	At               time.Time     `json:"at"`
	HTTP             int           `json:"http"`
	Inner            int           `json:"inner"`
	Conn             string        `json:"conn,omitempty"`
	Latency          time.Duration `json:"latency"`
	OK               bool          `json:"ok"`
	Class            string        `json:"class"`
	Reason           string        `json:"reason"`
	Message          string        `json:"message,omitempty"`
	ServedModel      string        `json:"served_model,omitempty"`
	FinishReason     string        `json:"finish_reason,omitempty"`
	PromptTokens     int           `json:"prompt_tokens,omitempty"`
	CompletionTokens int           `json:"completion_tokens,omitempty"`
}

// LatencySample is one streamed request measured by the latency lab.
type LatencySample struct {
	ModelID      string        `json:"model_id"`
	Index        int           `json:"index"`
	Warm         bool          `json:"warm"`
	At           time.Time     `json:"at"`
	OK           bool          `json:"ok"`
	Reason       string        `json:"reason,omitempty"`
	Message      string        `json:"message,omitempty"`
	TTFT         time.Duration `json:"ttft"`
	TTST         time.Duration `json:"ttst"`
	TPOT         time.Duration `json:"tpot"`
	E2E          time.Duration `json:"e2e"`
	TokPerSec    float64       `json:"tok_per_s"`
	InputTokens  int           `json:"input_tokens"`
	OutputTokens int           `json:"output_tokens"`
	TokensExact  bool          `json:"tokens_exact"`
	ServedModel  string        `json:"served_model,omitempty"`
}

// LoadRequest is one request issued by the load generator.
type LoadRequest struct {
	Level        int           `json:"level"` // concurrency level or rate bucket index
	Index        int           `json:"index"`
	At           time.Time     `json:"at"`
	OK           bool          `json:"ok"`
	Reason       string        `json:"reason,omitempty"`
	TTFT         time.Duration `json:"ttft"`
	E2E          time.Duration `json:"e2e"`
	ITL          time.Duration `json:"itl"` // mean inter-token latency
	OutputTokens int           `json:"output_tokens"`
	QueueWait    time.Duration `json:"queue_wait,omitempty"` // open-loop dispatcher delay
}

// CapabilityResult records one capability probe.
type CapabilityResult struct {
	ModelID  string        `json:"model_id"`
	Probe    string        `json:"probe"`
	Status   string        `json:"status"` // supported | unsupported | error | skipped
	Evidence string        `json:"evidence,omitempty"`
	Latency  time.Duration `json:"latency,omitempty"`
}

// QualityResult records one quality question trial.
type QualityResult struct {
	ModelID    string        `json:"model_id"`
	QuestionID string        `json:"question_id"`
	Trial      int           `json:"trial"`
	Pass       bool          `json:"pass"`
	Score      float64       `json:"score"`
	Output     string        `json:"output,omitempty"`
	Latency    time.Duration `json:"latency,omitempty"`
	Error      string        `json:"error,omitempty"`
}

// EventKind enumerates progress events.
const (
	EventStart    = "start" // a request for one model began (Attempt carries ModelID/Round)
	EventRound    = "round"
	EventProgress = "progress"
	EventAttempt  = "attempt"
	EventSample   = "sample"
	EventLevel    = "level"
	EventCooldown = "cooldown"
	EventLog      = "log"
	EventDone     = "done"
	EventError    = "error"
)

// Event is a progress notification emitted by the engines.
type Event struct {
	Seq     int64          `json:"seq,omitempty"`
	Kind    string         `json:"kind"`
	RunID   string         `json:"run_id,omitempty"`
	Stage   string         `json:"stage,omitempty"` // scan | latency | load | capability | quality
	Round   int            `json:"round,omitempty"`
	Rounds  int            `json:"rounds,omitempty"`
	Done    int            `json:"done"`
	Total   int            `json:"total"`
	OK      int            `json:"ok"`
	Fail    int            `json:"fail"`
	ETA     float64        `json:"eta_s,omitempty"`
	Message string         `json:"message,omitempty"`
	Attempt *Attempt       `json:"attempt,omitempty"`
	Sample  *LatencySample `json:"sample,omitempty"`
	Data    any            `json:"data,omitempty"`
	At      time.Time      `json:"at"`
}
