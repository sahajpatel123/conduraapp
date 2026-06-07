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

// OpenAICompat is the base for all OpenAI-compatible chat completions APIs.
// Most LLM providers expose an endpoint matching the OpenAI chat schema;
// the only per-provider differences are base URL, model list, and auth.
//
// Concrete providers (openrouter, together, groq, fireworks, deepseek,
// xai, mistral, ollama, custom) instantiate OpenAICompat with their
// own config and list of models.
type OpenAICompat struct {
	NameVal    string
	BaseURL    string // e.g. "https://api.openai.com/v1"
	APIKey     string
	HTTPClient *http.Client
	ModelsList []ModelInfo
	// AuthHeader is the header used for the API key. Default "Authorization".
	AuthHeader string
	// AuthPrefix is the value prefix (default "Bearer "). Set to "" for
	// providers that send the raw key in a custom header (e.g. OpenRouter
	// uses the same Bearer format but Together uses Bearer too).
	AuthPrefix string
	// ExtraHeaders are added to every request (e.g. OpenRouter app name).
	ExtraHeaders map[string]string
}

// NewOpenAICompat returns an OpenAICompat with sane defaults.
func NewOpenAICompat(name, baseURL, apiKey string) *OpenAICompat {
	return &OpenAICompat{
		NameVal:    name,
		BaseURL:    strings.TrimRight(baseURL, "/"),
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
	}
}

func (p *OpenAICompat) Name() string { return p.NameVal }

func (p *OpenAICompat) Models() []ModelInfo { return p.ModelsList }

func (p *OpenAICompat) DefaultModel(task string) string {
	if len(p.ModelsList) == 0 {
		return ""
	}
	// Filter out models that don't support the task.
	var candidates []ModelInfo
	for _, m := range p.ModelsList {
		if task == "vision" && !m.SupportsVision {
			continue
		}
		if task == "tool" && !m.SupportsTools {
			continue
		}
		candidates = append(candidates, m)
	}
	if len(candidates) == 0 {
		return p.ModelsList[0].ID
	}
	// Pick the cheapest qualifying model.
	best := candidates[0]
	bestCost := best.InputCostPerMTok + best.OutputCostPerMTok
	for _, m := range candidates[1:] {
		c := m.InputCostPerMTok + m.OutputCostPerMTok
		if c < bestCost {
			best, bestCost = m, c
		}
	}
	return best.ID
}

// -----------------------------------------------------------------------------
// Wire types (OpenAI chat completions)
// -----------------------------------------------------------------------------

type oaiRequest struct {
	Model       string       `json:"model"`
	Messages    []oaiMessage `json:"messages"`
	Tools       []oaiTool    `json:"tools,omitempty"`
	ToolChoice  any          `json:"tool_choice,omitempty"`
	Temperature *float64     `json:"temperature,omitempty"`
	TopP        *float64     `json:"top_p,omitempty"`
	MaxTokens   *int         `json:"max_tokens,omitempty"`
	Stop        []string     `json:"stop,omitempty"`
	Stream      bool         `json:"stream,omitempty"`
	User        string       `json:"user,omitempty"`
	// Some providers accept a metadata object; most ignore unknown fields.
	Metadata map[string]string `json:"metadata,omitempty"`
}

type oaiMessage struct {
	Role       string     `json:"role"`
	Content    string     `json:"content"`
	Name       string     `json:"name,omitempty"`
	ToolCallID string     `json:"tool_call_id,omitempty"`
	ToolCalls  []ToolCall `json:"tool_calls,omitempty"`
}

type oaiTool struct {
	Type     string `json:"type"`
	Function struct {
		Name        string         `json:"name"`
		Description string         `json:"description"`
		Parameters  map[string]any `json:"parameters,omitempty"`
	} `json:"function"`
}

type oaiResponse struct {
	ID      string `json:"id"`
	Object  string `json:"object"`
	Created int64  `json:"created"`
	Model   string `json:"model"`
	Choices []struct {
		Index        int        `json:"index"`
		Message      oaiMessage `json:"message"`
		FinishReason string     `json:"finish_reason"`
	} `json:"choices"`
	Usage struct {
		PromptTokens     int `json:"prompt_tokens"`
		CompletionTokens int `json:"completion_tokens"`
		TotalTokens      int `json:"total_tokens"`
	} `json:"usage"`
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    any    `json:"code"`
	} `json:"error,omitempty"`
}

// -----------------------------------------------------------------------------
// Request conversion
// -----------------------------------------------------------------------------

func toOAIRequest(req ChatRequest) oaiRequest {
	out := oaiRequest{
		Model:       req.Model,
		Temperature: req.Temperature,
		TopP:        req.TopP,
		MaxTokens:   req.MaxTokens,
		Stop:        req.Stop,
		Stream:      req.Stream,
		User:        req.User,
		Metadata:    req.Metadata,
		ToolChoice:  req.ToolChoice,
	}
	for _, m := range req.Messages {
		out.Messages = append(out.Messages, oaiMessage{
			Role:       string(m.Role),
			Content:    m.Content,
			Name:       m.Name,
			ToolCallID: m.ToolCallID,
			ToolCalls:  m.ToolCalls,
		})
	}
	for _, t := range req.Tools {
		ot := oaiTool{Type: t.Type}
		ot.Function.Name = t.Function.Name
		ot.Function.Description = t.Function.Description
		ot.Function.Parameters = t.Function.Parameters
		out.Tools = append(out.Tools, ot)
	}
	return out
}

func fromOAIResponse(r oaiResponse) (ChatResponse, error) {
	if len(r.Choices) == 0 {
		return ChatResponse{}, fmt.Errorf("%w: no choices", ErrResponseShape)
	}
	c := r.Choices[0]
	resp := ChatResponse{
		ID:    r.ID,
		Model: r.Model,
		Message: Message{
			Role:      Role(c.Message.Role),
			Content:   c.Message.Content,
			ToolCalls: c.Message.ToolCalls,
			Name:      c.Message.Name,
		},
		FinishReason: FinishReason(c.FinishReason),
		Usage: Usage{
			InputTokens:  r.Usage.PromptTokens,
			OutputTokens: r.Usage.CompletionTokens,
			TotalTokens:  r.Usage.TotalTokens,
		},
	}
	if resp.FinishReason == "" {
		resp.FinishReason = FinishStop
	}
	if resp.Message.Role == "" {
		resp.Message.Role = RoleAssistant
	}
	return resp, nil
}

// -----------------------------------------------------------------------------
// HTTP plumbing
// -----------------------------------------------------------------------------

// HTTPDoer is the subset of *http.Client used by OpenAICompat. Tests can
// inject a custom transport.
type HTTPDoer interface {
	Do(*http.Request) (*http.Response, error)
}

func (p *OpenAICompat) client() HTTPDoer {
	if p.HTTPClient != nil {
		return p.HTTPClient
	}
	return &http.Client{Timeout: 5 * time.Minute}
}

func (p *OpenAICompat) buildRequest(ctx context.Context, method, path string, body any) (*http.Request, error) {
	var rdr io.Reader
	if body != nil {
		data, err := json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("llm: marshal: %w", err)
		}
		rdr = bytes.NewReader(data)
	}
	req, err := http.NewRequestWithContext(ctx, method, p.BaseURL+path, rdr)
	if err != nil {
		return nil, err
	}
	if p.APIKey != "" {
		req.Header.Set(p.AuthHeader, p.AuthPrefix+p.APIKey)
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")
	for k, v := range p.ExtraHeaders {
		req.Header.Set(k, v)
	}
	return req, nil
}

func (p *OpenAICompat) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	if req.Model == "" {
		return ChatResponse{}, ErrNoModel
	}
	if len(req.Messages) == 0 {
		return ChatResponse{}, ErrNoMessages
	}
	if req.Stream {
		// We still return a final response by draining the stream.
		ch, cancel, err := p.Stream(ctx, req)
		if err != nil {
			return ChatResponse{}, err
		}
		defer cancel()
		var (
			content strings.Builder
			finish  FinishReason
			usage   Usage
			role    Role
		)
		for ev := range ch {
			if ev.Err != nil {
				return ChatResponse{}, ev.Err
			}
			content.WriteString(ev.Delta.Content)
			if ev.Delta.Role != "" {
				role = ev.Delta.Role
			}
			if ev.FinishReason != "" {
				finish = ev.FinishReason
			}
			if !ev.Done {
				ev.Usage.Add(usage)
			} else {
				usage = ev.Usage
			}
		}
		if finish == "" {
			return ChatResponse{}, ErrResponseShape
		}
		return ChatResponse{
			Message:      Message{Role: role, Content: content.String()},
			FinishReason: finish,
			Usage:        usage,
		}, nil
	}
	oai := toOAIRequest(req)
	httpReq, err := p.buildRequest(ctx, http.MethodPost, "/chat/completions", oai)
	if err != nil {
		return ChatResponse{}, err
	}
	resp, err := p.client().Do(httpReq)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("llm: http: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return ChatResponse{}, fmt.Errorf("llm: %s: %d: %s", p.NameVal, resp.StatusCode, string(body))
	}
	var r oaiResponse
	if err := json.Unmarshal(body, &r); err != nil {
		return ChatResponse{}, fmt.Errorf("%w: %v: %s", ErrResponseShape, err, string(body))
	}
	if r.Error != nil {
		return ChatResponse{}, fmt.Errorf("llm: %s: %s", p.NameVal, r.Error.Message)
	}
	cr, err := fromOAIResponse(r)
	if err != nil {
		return ChatResponse{}, err
	}
	cr.Raw = body
	return cr, nil
}

// Stream implements SSE streaming per the OpenAI chat spec.
func (p *OpenAICompat) Stream(ctx context.Context, req ChatRequest) (<-chan StreamEvent, func(), error) {
	if req.Model == "" {
		return nil, nil, ErrNoModel
	}
	if len(req.Messages) == 0 {
		return nil, nil, ErrNoMessages
	}
	oai := toOAIRequest(req)
	oai.Stream = true
	httpReq, err := p.buildRequest(ctx, http.MethodPost, "/chat/completions", oai)
	if err != nil {
		return nil, nil, err
	}
	resp, err := p.client().Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("llm: http: %w", err)
	}
	if resp.StatusCode >= 400 {
		body, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, nil, fmt.Errorf("llm: %s: %d: %s", p.NameVal, resp.StatusCode, string(body))
	}

	out := make(chan StreamEvent, 16)
	cancel := make(chan struct{})
	go func() {
		defer close(out)
		defer func() { _ = resp.Body.Close() }()
		reader := bufio.NewReaderSize(resp.Body, 64*1024)
		var (
			accumulated  strings.Builder
			finishReason FinishReason
			usage        Usage
		)
		for {
			select {
			case <-cancel:
				return
			default:
			}
			line, err := reader.ReadString('\n')
			if err != nil {
				if !errors.Is(err, io.EOF) {
					out <- StreamEvent{Err: fmt.Errorf("llm: read: %w", err), Done: true}
				}
				return
			}
			line = strings.TrimRight(line, "\r\n")
			if line == "" {
				continue
			}
			if !strings.HasPrefix(line, "data: ") {
				continue
			}
			payload := strings.TrimPrefix(line, "data: ")
			if payload == "[DONE]" {
				// Emit a final terminal event with usage + finish reason.
				// Content was already streamed via per-delta events.
				out <- StreamEvent{
					FinishReason: finishReason,
					Usage:        usage,
					Done:         true,
				}
				return
			}
			var chunk struct {
				Choices []struct {
					Delta        oaiMessage `json:"delta"`
					FinishReason string     `json:"finish_reason"`
				} `json:"choices"`
				Usage *struct {
					PromptTokens     int `json:"prompt_tokens"`
					CompletionTokens int `json:"completion_tokens"`
					TotalTokens      int `json:"total_tokens"`
				} `json:"usage"`
			}
			if err := json.Unmarshal([]byte(payload), &chunk); err != nil {
				out <- StreamEvent{Err: fmt.Errorf("llm: parse chunk: %w", err), Done: true}
				return
			}
			if len(chunk.Choices) > 0 {
				c := chunk.Choices[0]
				accumulated.WriteString(c.Delta.Content)
				if c.FinishReason != "" {
					finishReason = FinishReason(c.FinishReason)
				}
				out <- StreamEvent{
					Delta: Message{
						Role:      Role(c.Delta.Role),
						Content:   c.Delta.Content,
						ToolCalls: c.Delta.ToolCalls,
					},
				}
			}
			if chunk.Usage != nil {
				usage = Usage{
					InputTokens:  chunk.Usage.PromptTokens,
					OutputTokens: chunk.Usage.CompletionTokens,
					TotalTokens:  chunk.Usage.TotalTokens,
				}
			}
		}
	}()
	return out, func() { close(cancel) }, nil
}

// -----------------------------------------------------------------------------
// Provider-specific factories
// -----------------------------------------------------------------------------

// NewOpenAI returns a Provider for OpenAI.
func NewOpenAI(apiKey string, models []ModelInfo) *OpenAICompat {
	return &OpenAICompat{
		NameVal:    "openai",
		BaseURL:    "https://api.openai.com/v1",
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		ModelsList: models,
	}
}

// NewOpenRouter returns a Provider for OpenRouter.
func NewOpenRouter(apiKey string, models []ModelInfo) *OpenAICompat {
	return &OpenAICompat{
		NameVal:    "openrouter",
		BaseURL:    "https://openrouter.ai/api/v1",
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		ExtraHeaders: map[string]string{
			"HTTP-Referer": "https://synaptic.app",
			"X-Title":      "Synaptic",
		},
		ModelsList: models,
	}
}

// NewTogether returns a Provider for Together.
func NewTogether(apiKey string, models []ModelInfo) *OpenAICompat {
	return &OpenAICompat{
		NameVal:    "together",
		BaseURL:    "https://api.together.xyz/v1",
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		ModelsList: models,
	}
}

// NewGroq returns a Provider for Groq.
func NewGroq(apiKey string, models []ModelInfo) *OpenAICompat {
	return &OpenAICompat{
		NameVal:    "groq",
		BaseURL:    "https://api.groq.com/openai/v1",
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		ModelsList: models,
	}
}

// NewFireworks returns a Provider for Fireworks.
func NewFireworks(apiKey string, models []ModelInfo) *OpenAICompat {
	return &OpenAICompat{
		NameVal:    "fireworks",
		BaseURL:    "https://api.fireworks.ai/inference/v1",
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		ModelsList: models,
	}
}

// NewDeepSeek returns a Provider for DeepSeek.
func NewDeepSeek(apiKey string, models []ModelInfo) *OpenAICompat {
	return &OpenAICompat{
		NameVal:    "deepseek",
		BaseURL:    "https://api.deepseek.com/v1",
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		ModelsList: models,
	}
}

// NewXAI returns a Provider for xAI (Grok).
func NewXAI(apiKey string, models []ModelInfo) *OpenAICompat {
	return &OpenAICompat{
		NameVal:    "xai",
		BaseURL:    "https://api.x.ai/v1",
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		ModelsList: models,
	}
}

// NewMistral returns a Provider for Mistral.
func NewMistral(apiKey string, models []ModelInfo) *OpenAICompat {
	return &OpenAICompat{
		NameVal:    "mistral",
		BaseURL:    "https://api.mistral.ai/v1",
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		ModelsList: models,
	}
}

// NewOllama returns a Provider for Ollama (no API key required).
func NewOllama(baseURL string, models []ModelInfo) *OpenAICompat {
	if baseURL == "" {
		baseURL = "http://localhost:11434/v1"
	}
	return &OpenAICompat{
		NameVal:    "ollama",
		BaseURL:    baseURL,
		APIKey:     "ollama", // Ollama ignores; some setups require any non-empty value
		HTTPClient: &http.Client{Timeout: 10 * time.Minute},
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		ModelsList: models,
	}
}

// NewCustom returns a Provider for a custom OpenAI-compatible endpoint.
func NewCustom(name, baseURL, apiKey string, models []ModelInfo) *OpenAICompat {
	return &OpenAICompat{
		NameVal:    name,
		BaseURL:    baseURL,
		APIKey:     apiKey,
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
		AuthHeader: "Authorization",
		AuthPrefix: "Bearer ",
		ModelsList: models,
	}
}
