package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"
)

// Anthropic implements the Provider interface for Anthropic's Messages API.
// Endpoint: https://api.anthropic.com/v1/messages
//
// Differences from the OpenAI API:
//   - System messages are passed as a top-level "system" field, not as
//     a message in the "messages" array.
//   - Tools are "input_schema" instead of "parameters".
//   - max_tokens is required.
//   - "anthropic-version" header is required.
//   - SSE event format differs (event: + data: lines).
type Anthropic struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	Version    string
	ModelsList []ModelInfo
}

// NewAnthropic returns an Anthropic provider.
func NewAnthropic(apiKey string, models []ModelInfo) *Anthropic {
	return &Anthropic{
		APIKey:     apiKey,
		BaseURL:    "https://api.anthropic.com",
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
		Version:    "2023-06-01",
		ModelsList: models,
	}
}

// Name returns the canonical provider name.
func (a *Anthropic) Name() string { return "anthropic" }

// GetHTTPClient returns the Anthropic provider's underlying HTTP
// client. Used by the daemon to wrap the transport with the
// network guard.
func (a *Anthropic) GetHTTPClient() *http.Client { return a.HTTPClient }

// Models returns the list of model IDs this provider can serve.
func (a *Anthropic) Models() []ModelInfo { return a.ModelsList }

// DefaultModel returns the recommended model for a given task
// (e.g. "chat", "code", "vision"). Prefers the marketing-aligned
// current generation; falls back to the first model in the registry.
func (a *Anthropic) DefaultModel(task string) string {
	// Preference order: current gen, then legacy Claude 3.5, then first.
	for _, id := range []string{"claude-sonnet-4-5", "claude-3-5-sonnet-20241022"} {
		for _, m := range a.ModelsList {
			if m.ID == id {
				return m.ID
			}
		}
	}
	if len(a.ModelsList) == 0 {
		return ""
	}
	return a.ModelsList[0].ID
}

// -----------------------------------------------------------------------------
// Wire types
// -----------------------------------------------------------------------------

type anthRequest struct {
	Model       string            `json:"model"`
	Messages    []anthMessage     `json:"messages"`
	System      string            `json:"system,omitempty"`
	Tools       []anthTool        `json:"tools,omitempty"`
	MaxTokens   int               `json:"max_tokens"`
	Temperature *float64          `json:"temperature,omitempty"`
	TopP        *float64          `json:"top_p,omitempty"`
	StopSeqs    []string          `json:"stop_sequences,omitempty"`
	Stream      bool              `json:"stream,omitempty"`
	Metadata    map[string]string `json:"metadata,omitempty"`
}

type anthMessage struct {
	Role    string `json:"role"`
	Content any    `json:"content"` // string or []block
}

type anthTool struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	InputSchema map[string]any `json:"input_schema"`
}

type anthResponse struct {
	ID           string `json:"id"`
	Type         string `json:"type"`
	Model        string `json:"model"`
	StopReason   string `json:"stop_reason"`
	StopSequence string `json:"stop_sequence,omitempty"`
	Content      []struct {
		Type  string `json:"type"` // "text" or "tool_use"
		Text  string `json:"text,omitempty"`
		ID    string `json:"id,omitempty"`
		Name  string `json:"name,omitempty"`
		Input any    `json:"input,omitempty"`
	} `json:"content"`
	Usage struct {
		InputTokens  int `json:"input_tokens"`
		OutputTokens int `json:"output_tokens"`
	} `json:"usage"`
	Error *struct {
		Type    string `json:"type"`
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

// -----------------------------------------------------------------------------
// Request conversion
// -----------------------------------------------------------------------------

func (a *Anthropic) toRequest(req ChatRequest) anthRequest {
	out := anthRequest{
		Model:       req.Model,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		StopSeqs:    req.Stop,
		Stream:      req.Stream,
		Metadata:    req.Metadata,
		// Anthropic requires max_tokens; default to 4096.
		MaxTokens: 4096,
	}
	if req.MaxTokens != nil {
		out.MaxTokens = *req.MaxTokens
	}
	var sysPrompt string
	for _, m := range req.Messages {
		if m.Role == RoleSystem {
			sysPrompt += m.Content + "\n"
			continue
		}
		// Anthropic's "user" and "assistant" roles map directly.
		// Tool results must be sent as a "user" message with content
		// blocks of type "tool_result"; we keep the simple case (a flat
		// string) for now and let callers send tool results as strings.
		out.Messages = append(out.Messages, anthMessage{
			Role:    string(m.Role),
			Content: m.Content,
		})
	}
	out.System = strings.TrimRight(sysPrompt, "\n")
	for _, t := range req.Tools {
		out.Tools = append(out.Tools, anthTool{
			Name:        t.Function.Name,
			Description: t.Function.Description,
			InputSchema: t.Function.Parameters,
		})
	}
	return out
}

func (a *Anthropic) fromResponse(r anthResponse) ChatResponse {
	resp := ChatResponse{
		ID:           r.ID,
		Model:        r.Model,
		FinishReason: mapAnthStopReason(r.StopReason),
		Usage: Usage{
			InputTokens:  r.Usage.InputTokens,
			OutputTokens: r.Usage.OutputTokens,
			TotalTokens:  r.Usage.InputTokens + r.Usage.OutputTokens,
		},
	}
	for _, c := range r.Content {
		switch c.Type {
		case "text":
			resp.Message.Content += c.Text
		case "tool_use":
			args, _ := json.Marshal(c.Input)
			tc := ToolCall{Type: "function"}
			tc.ID = c.ID
			tc.Function.Name = c.Name
			tc.Function.Arguments = string(args)
			resp.Message.ToolCalls = append(resp.Message.ToolCalls, tc)
		}
	}
	resp.Message.Role = RoleAssistant
	return resp
}

func mapAnthStopReason(s string) FinishReason {
	switch s {
	case "end_turn":
		return FinishStop
	case "max_tokens":
		return FinishLength
	case "stop_sequence":
		return FinishStop
	case "tool_use":
		return FinishToolCalls
	default:
		return FinishReason(s)
	}
}

// -----------------------------------------------------------------------------
// HTTP
// -----------------------------------------------------------------------------

func (a *Anthropic) buildRequest(ctx context.Context, body any, stream bool) (*http.Request, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, a.BaseURL+"/v1/messages", bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("x-api-key", a.APIKey)
	req.Header.Set("anthropic-version", a.Version)
	if stream {
		req.Header.Set("Accept", "text/event-stream")
	}
	return req, nil
}

// Chat sends a non-streaming request to the Anthropic Messages API and
// returns the assembled response.
func (a *Anthropic) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	if req.Model == "" {
		return ChatResponse{}, ErrNoModel
	}
	if len(req.Messages) == 0 {
		return ChatResponse{}, ErrNoMessages
	}
	if req.Stream {
		ch, cancel, err := a.Stream(ctx, req)
		if err != nil {
			return ChatResponse{}, err
		}
		defer cancel()
		var last StreamEvent
		for ev := range ch {
			last = ev
			if ev.Err != nil {
				return ChatResponse{}, ev.Err
			}
		}
		return ChatResponse{
			Message:      last.Delta,
			FinishReason: last.FinishReason,
			Usage:        last.Usage,
		}, nil
	}
	body := a.toRequest(req)
	httpReq, err := a.buildRequest(ctx, body, false)
	if err != nil {
		return ChatResponse{}, err
	}
	resp, err := a.HTTPClient.Do(httpReq)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("llm/anthropic: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return ChatResponse{}, fmt.Errorf("llm/anthropic: %d: %s", resp.StatusCode, string(raw))
	}
	var r anthResponse
	if err := json.Unmarshal(raw, &r); err != nil {
		return ChatResponse{}, fmt.Errorf("%w: %w", ErrResponseShape, err)
	}
	if r.Error != nil {
		return ChatResponse{}, fmt.Errorf("llm/anthropic: %s: %s", r.Error.Type, r.Error.Message)
	}
	cr := a.fromResponse(r)
	cr.Raw = raw
	return cr, nil
}

// Stream returns a channel of incremental events from the Anthropic SSE
// stream. The cancel function aborts the in-flight request.
func (a *Anthropic) Stream(ctx context.Context, req ChatRequest) (<-chan StreamEvent, func(), error) {
	if req.Model == "" {
		return nil, nil, ErrNoModel
	}
	if len(req.Messages) == 0 {
		return nil, nil, ErrNoMessages
	}
	body := a.toRequest(req)
	httpReq, err := a.buildRequest(ctx, body, true)
	if err != nil {
		return nil, nil, err
	}
	//nolint:bodyclose // body is closed in the streaming goroutine below
	resp, err := a.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("llm/anthropic: %w", err)
	}
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, nil, fmt.Errorf("llm/anthropic: %d: %s", resp.StatusCode, string(b))
	}
	out := make(chan StreamEvent, 16)
	cancel := make(chan struct{})
	go a.streamAnthropicEvents(out, cancel, resp.Body)
	return out, func() { close(cancel) }, nil
}

// streamAnthropicEvents is the inner loop of Stream: it reads SSE lines
// from body, accumulates multi-line `data:` payloads, and dispatches each
// complete event to a per-type handler.
func (a *Anthropic) streamAnthropicEvents(out chan<- StreamEvent, cancel <-chan struct{}, body io.ReadCloser) {
	defer close(out)
	defer func() { _ = body.Close() }()

	reader := bufio.NewReaderSize(body, 64*1024)
	state := newAnthropicStreamState()

	for {
		select {
		case <-cancel:
			return
		default:
		}
		line, err := reader.ReadString('\n')
		if err != nil {
			state.flush(out)
			if !errors.Is(err, io.EOF) {
				out <- StreamEvent{Err: fmt.Errorf("llm/anthropic: read: %w", err), Done: true}
			}
			return
		}
		line = strings.TrimRight(line, "\r\n")
		if line == "" {
			state.flush(out)
			continue
		}
		if strings.HasPrefix(line, "event: ") {
			// We don't drive parsing by the SSE event name; the JSON
			// payload's "type" field is the source of truth.
			continue
		}
		if strings.HasPrefix(line, "data: ") {
			state.event.WriteString(strings.TrimPrefix(line, "data: "))
		}
	}
}

// anthropicStreamState holds the per-stream accumulator and finished-ness.
// Methods are not concurrency-safe; one Stream owns one state.
type anthropicStreamState struct {
	accumulated  strings.Builder
	finishReason FinishReason
	usage        Usage
	event        strings.Builder
}

// newAnthropicStreamState returns an empty state.
func newAnthropicStreamState() *anthropicStreamState {
	return &anthropicStreamState{}
}

// anthStreamEvent is the subset of the Anthropic SSE event payload we read.
type anthStreamEvent struct {
	Type  string `json:"type"`
	Delta struct {
		Type string `json:"type"`
		Text string `json:"text"`
	} `json:"delta"`
	Message struct {
		Usage struct {
			InputTokens  int `json:"input_tokens"`
			OutputTokens int `json:"output_tokens"`
		} `json:"usage"`
	} `json:"message"`
}

// flush parses any accumulated `data:` payload and dispatches it to the
// per-type handler. It is safe to call on an empty buffer.
func (s *anthropicStreamState) flush(out chan<- StreamEvent) {
	if s.event.Len() == 0 {
		return
	}
	var ev anthStreamEvent
	if err := json.Unmarshal([]byte(s.event.String()), &ev); err != nil {
		s.event.Reset()
		return
	}
	s.event.Reset()
	s.dispatch(out, ev)
}

// dispatch routes one parsed event to the appropriate branch.
func (s *anthropicStreamState) dispatch(out chan<- StreamEvent, ev anthStreamEvent) {
	switch ev.Type {
	case "content_block_delta":
		if ev.Delta.Type == "text_delta" {
			s.accumulated.WriteString(ev.Delta.Text)
			out <- StreamEvent{Delta: Message{Role: RoleAssistant, Content: ev.Delta.Text}}
		}
	case "message_delta":
		// Stop reason update; we capture it on message_stop.
	case "message_start":
		s.usage.InputTokens = ev.Message.Usage.InputTokens
		s.usage.OutputTokens = ev.Message.Usage.OutputTokens
	case "message_stop":
		out <- StreamEvent{
			FinishReason: s.finishReason,
			Usage:        s.usage,
			Done:         true,
		}
	}
}
