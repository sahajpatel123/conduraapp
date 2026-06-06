// Package llm is the LLM provider abstraction for Synaptic.
//
// Every LLM API (OpenAI, Anthropic, Google, OpenRouter, Together, Groq,
// Fireworks, DeepSeek, xAI, Mistral, Ollama, custom) is wrapped behind a
// common Provider interface. The router and failover layers only see this
// interface; per-provider quirks are encapsulated in the implementations.
//
// Wire format:
//   - ChatRequest / ChatResponse model the OpenAI chat completions shape.
//     Anthropic and Google are translated at the edge of their respective
//     packages.
//   - Streaming is a channel of StreamEvents; the caller closes via the
//     returned cancel function.
//
// Cost:
//   - Per-model input/output USD/MTok is hard-coded in this file
//     (model_pricing.go) so the failover layer can estimate spend without
//     calling out to the model. Update these as pricing changes.
package llm

import (
	"bytes"
	"context"
	"errors"
	"io"
)

// Role is the role of a message in a chat conversation.
type Role string

const (
	RoleSystem    Role = "system"
	RoleUser      Role = "user"
	RoleAssistant Role = "assistant"
	RoleTool      Role = "tool"
)

// IsValid reports whether r is a known role.
func (r Role) IsValid() bool {
	switch r {
	case RoleSystem, RoleUser, RoleAssistant, RoleTool:
		return true
	}
	return false
}

// FinishReason indicates why a chat response ended.
type FinishReason string

const (
	FinishStop          FinishReason = "stop"
	FinishLength        FinishReason = "length"
	FinishToolCalls     FinishReason = "tool_calls"
	FinishContentFilter FinishReason = "content_filter"
	FinishError         FinishReason = "error"
)

// Usage tracks token counts for a single Chat call.
type Usage struct {
	InputTokens  int
	OutputTokens int
	TotalTokens  int
}

// Add sums u into the receiver.
func (u *Usage) Add(other Usage) {
	u.InputTokens += other.InputTokens
	u.OutputTokens += other.OutputTokens
	u.TotalTokens += other.TotalTokens
}

// ToolCall is a model-initiated tool invocation. The model asks us to call
// a tool and we (the agent) execute it; the result is then sent back.
type ToolCall struct {
	ID       string `json:"id"`
	Type     string `json:"type"` // always "function" today
	Function struct {
		Name      string `json:"name"`
		Arguments string `json:"arguments"` // JSON-encoded
	} `json:"function"`
}

// ToolDefinition describes a tool the model can call.
type ToolDefinition struct {
	Type     string `json:"type"` // always "function"
	Function struct {
		Name        string `json:"name"`
		Description string `json:"description"`
		// Parameters is a JSON Schema object describing the arguments.
		Parameters map[string]any `json:"parameters,omitempty"`
	} `json:"function"`
}

// Message is one turn in a conversation.
type Message struct {
	Role       Role       `json:"role"`
	Content    string     `json:"content"`
	Name       string     `json:"name,omitempty"`         // for tool messages
	ToolCallID string     `json:"tool_call_id,omitempty"` // for tool messages
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`   // for assistant messages
}

// ChatRequest is the input to a Chat call.
type ChatRequest struct {
	// Model is the provider-specific model identifier (e.g. "gpt-4o-mini",
	// "claude-3-5-sonnet-20241022", "gemini-1.5-flash").
	Model string
	// Messages is the conversation so far. The first message should usually
	// be a system message.
	Messages []Message
	// Tools available to the model.
	Tools []ToolDefinition
	// ToolChoice: "auto", "required", "none", or {"type":"function","function":{"name":"..."}}.
	ToolChoice any
	// Temperature 0..2; 0 = deterministic, 2 = very random.
	Temperature *float64
	// TopP for nucleus sampling.
	TopP *float64
	// MaxTokens limits the response length.
	MaxTokens *int
	// Stop sequences.
	Stop []string
	// Stream is true if the caller wants a streaming response.
	Stream bool
	// User is an opaque end-user identifier for abuse tracking.
	User string
	// Metadata is passed through to providers that support it.
	Metadata map[string]string
}

// ChatResponse is the result of a non-streaming Chat call.
type ChatResponse struct {
	ID           string
	Model        string
	Message      Message
	FinishReason FinishReason
	Usage        Usage
	// Raw is the raw provider response (for debugging/audit).
	Raw []byte
}

// StreamEvent is one chunk of a streaming response.
type StreamEvent struct {
	// Delta is the incremental content (text or tool call fragment).
	Delta Message
	// Usage is populated on the final event.
	Usage Usage
	// FinishReason is populated on the final event.
	FinishReason FinishReason
	// Err is non-nil if the stream errored.
	Err error
	// Done is true for the terminal event (after Err or FinishReason).
	Done bool
}

// ModelInfo describes a model the provider offers.
type ModelInfo struct {
	ID                string
	DisplayName       string
	ContextWindow     int
	InputCostPerMTok  float64 // USD per 1M input tokens
	OutputCostPerMTok float64
	SupportsTools     bool
	SupportsVision    bool
	SupportsStream    bool
}

// Common errors.
var (
	ErrNoModel        = errors.New("llm: no model specified")
	ErrNoMessages     = errors.New("llm: no messages")
	ErrNoAPIKey       = errors.New("llm: no api key available")
	ErrNoProvider     = errors.New("llm: provider not registered")
	ErrNotImplemented = errors.New("llm: not implemented")
	ErrResponseShape  = errors.New("llm: unexpected response shape")
)

// Provider is the interface every LLM backend implements.
type Provider interface {
	// Name returns the canonical provider name (matches internal/api_key).
	Name() string
	// Chat performs a non-streaming chat completion.
	Chat(ctx context.Context, req ChatRequest) (ChatResponse, error)
	// Stream performs a streaming chat completion. The returned channel
	// receives StreamEvent values; the final event has Done=true.
	// Callers should drain the channel and call cancel() when done.
	Stream(ctx context.Context, req ChatRequest) (<-chan StreamEvent, func(), error)
	// Models returns the list of models this provider serves.
	Models() []ModelInfo
	// DefaultModel returns the recommended model for a given task.
	// task is one of "chat", "embedding", "vision", "tool", "code".
	DefaultModel(task string) string
}

// -----------------------------------------------------------------------------
// Auth / credentials accessor — supplied by internal/api_key.
// -----------------------------------------------------------------------------

// Authenticator is the subset of internal/api_key consumed here.
// We re-declare it to avoid an import cycle.
type Authenticator interface {
	GetByLabel(ctx context.Context, provider, label string) (apiKey, error)
	ListByProvider(ctx context.Context, provider string) ([]apiKey, error)
	Touch(ctx context.Context, id int64) error
}

// apiKey is the minimal key shape we need; matches internal/api_key.Key.
type apiKey struct {
	ID        int64
	Provider  string
	Label     string
	AuthKind  string
	Secret    string
	Refresh   string
	ExpiresAt string
}

// Match the shape with internal/api_key.Key via a public adapter.
// We use a free function to break the import cycle: callers pass an
// adapter at registry-construction time. See Adapter in adapter.go.

// -----------------------------------------------------------------------------
// Cost estimation
// -----------------------------------------------------------------------------

// EstimateCost returns the USD cost for the given usage against a model.
// Returns 0 if the model is unknown.
func EstimateCost(model string, u Usage) float64 {
	info, ok := LookupModel(model)
	if !ok {
		return 0
	}
	return float64(u.InputTokens)/1_000_000*info.InputCostPerMTok +
		float64(u.OutputTokens)/1_000_000*info.OutputCostPerMTok
}

// EstimateCostFromInfo returns the USD cost for the given usage against an
// explicit ModelInfo. Useful for custom models registered at runtime.
func EstimateCostFromInfo(info ModelInfo, u Usage) float64 {
	return float64(u.InputTokens)/1_000_000*info.InputCostPerMTok +
		float64(u.OutputTokens)/1_000_000*info.OutputCostPerMTok
}

// modelRegistry is a name → ModelInfo map used by EstimateCost.
// Populated by init() in model_pricing.go.
var modelRegistry = map[string]ModelInfo{}

// LookupModel returns the ModelInfo for the given model name.
func LookupModel(model string) (ModelInfo, bool) {
	m, ok := modelRegistry[model]
	return m, ok
}

// RegisterModel adds a model to the global pricing registry. Intended for
// tests and runtime extensions.
func RegisterModel(m ModelInfo) {
	modelRegistry[m.ID] = m
}

// UnregisterModel removes a model from the global pricing registry.
func UnregisterModel(id string) {
	delete(modelRegistry, id)
}

// -----------------------------------------------------------------------------
// Helper: copy a body for audit
// -----------------------------------------------------------------------------

// CopyBody reads a body and returns both the bytes and a new ReadCloser
// that yields the same bytes. Useful when the body needs to be inspected
// (for audit/debug) and also consumed by the request.
func CopyBody(b io.ReadCloser) ([]byte, io.ReadCloser, error) {
	if b == nil {
		return nil, nil, nil
	}
	data, err := io.ReadAll(b)
	_ = b.Close()
	if err != nil {
		return nil, nil, err
	}
	return data, io.NopCloser(bytes.NewReader(data)), nil
}
