package llm

import (
	"bufio"
	"bytes"
	"context"
	"encoding/json"
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

func (a *Anthropic) Name() string { return "anthropic" }

func (a *Anthropic) Models() []ModelInfo { return a.ModelsList }

func (a *Anthropic) DefaultModel(task string) string {
	for _, m := range a.ModelsList {
		if m.ID == "claude-3-5-sonnet-20241022" {
			return m.ID
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

func (a *Anthropic) fromResponse(r anthResponse) (ChatResponse, error) {
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
	return resp, nil
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
	defer resp.Body.Close()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return ChatResponse{}, fmt.Errorf("llm/anthropic: %d: %s", resp.StatusCode, string(raw))
	}
	var r anthResponse
	if err := json.Unmarshal(raw, &r); err != nil {
		return ChatResponse{}, fmt.Errorf("%w: %v", ErrResponseShape, err)
	}
	if r.Error != nil {
		return ChatResponse{}, fmt.Errorf("llm/anthropic: %s: %s", r.Error.Type, r.Error.Message)
	}
	cr, err := a.fromResponse(r)
	if err != nil {
		return ChatResponse{}, err
	}
	cr.Raw = raw
	return cr, nil
}

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
	go func() {
		defer close(out)
		defer resp.Body.Close()
		reader := bufio.NewReaderSize(resp.Body, 64*1024)
		var (
			accumulated  strings.Builder
			finishReason FinishReason
			usage        Usage
			event        strings.Builder
		)
		flush := func() {
			if event.Len() == 0 {
				return
			}
			var ev struct {
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
			if err := json.Unmarshal([]byte(event.String()), &ev); err != nil {
				event.Reset()
				return
			}
			event.Reset()
			switch ev.Type {
			case "content_block_delta":
				if ev.Delta.Type == "text_delta" {
					accumulated.WriteString(ev.Delta.Text)
					out <- StreamEvent{Delta: Message{Role: RoleAssistant, Content: ev.Delta.Text}}
				}
			case "message_delta":
				// Stop reason update; we'll capture it on message_stop.
				_ = ev
			case "message_start":
				usage.InputTokens = ev.Message.Usage.InputTokens
				usage.OutputTokens = ev.Message.Usage.OutputTokens
			case "message_stop":
				// Emit a terminal event with usage but no delta content
				// (the per-delta events already streamed it).
				out <- StreamEvent{
					FinishReason: finishReason,
					Usage:        usage,
					Done:         true,
				}
			}
		}
		for {
			select {
			case <-cancel:
				return
			default:
			}
			line, err := reader.ReadString('\n')
			if err != nil {
				flush()
				if err != io.EOF {
					out <- StreamEvent{Err: fmt.Errorf("llm/anthropic: read: %w", err), Done: true}
				}
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if line == "" {
				flush()
				continue
			}
			if strings.HasPrefix(line, "event: ") {
				// For Phase 1 we don't need the event type to drive parsing;
				// we just rely on the data payload type.
				continue
			}
			if strings.HasPrefix(line, "data: ") {
				event.WriteString(strings.TrimPrefix(line, "data: "))
			}
		}
	}()
	return out, func() { close(cancel) }, nil
}
