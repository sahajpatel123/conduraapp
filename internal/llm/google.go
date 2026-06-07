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

// Google implements the Provider interface for Google's Gemini API.
// Endpoint: https://generativelanguage.googleapis.com/v1beta
//
// Auth: ?key=API_KEY query parameter, or Bearer token for OAuth (Phase 2).
// Wire format: Gemini's generateContent schema is different from OpenAI;
// we translate at the edge.
type Google struct {
	APIKey     string
	BaseURL    string
	HTTPClient *http.Client
	ModelsList []ModelInfo
}

// NewGoogle returns a Google Gemini provider.
func NewGoogle(apiKey string, models []ModelInfo) *Google {
	return &Google{
		APIKey:     apiKey,
		BaseURL:    "https://generativelanguage.googleapis.com",
		HTTPClient: &http.Client{Timeout: 5 * time.Minute},
		ModelsList: models,
	}
}

func (g *Google) Name() string { return "google" }

func (g *Google) Models() []ModelInfo { return g.ModelsList }

func (g *Google) DefaultModel(task string) string {
	for _, m := range g.ModelsList {
		if m.ID == "gemini-1.5-flash" {
			return m.ID
		}
	}
	if len(g.ModelsList) == 0 {
		return ""
	}
	return g.ModelsList[0].ID
}

// -----------------------------------------------------------------------------
// Wire types (Gemini generateContent)
// -----------------------------------------------------------------------------

type gemRequest struct {
	Contents          []gemContent  `json:"contents"`
	SystemInstruction *gemContent   `json:"systemInstruction,omitempty"`
	Tools             []gemTool     `json:"tools,omitempty"`
	GenerationConfig  *gemGenConfig `json:"generationConfig,omitempty"`
}

type gemContent struct {
	Role  string    `json:"role,omitempty"` // "user" or "model"
	Parts []gemPart `json:"parts"`
}

type gemPart struct {
	Text         string           `json:"text,omitempty"`
	FunctionCall *gemFunctionCall `json:"functionCall,omitempty"`
	FunctionResp *gemFunctionResp `json:"functionResponse,omitempty"`
}

type gemFunctionCall struct {
	Name string         `json:"name"`
	Args map[string]any `json:"args,omitempty"`
}

type gemFunctionResp struct {
	Name     string         `json:"name"`
	Response map[string]any `json:"response"`
}

type gemTool struct {
	FunctionDeclarations []gemFunctionDecl `json:"functionDeclarations"`
}

type gemFunctionDecl struct {
	Name        string         `json:"name"`
	Description string         `json:"description"`
	Parameters  map[string]any `json:"parameters,omitempty"`
}

type gemGenConfig struct {
	Temperature     *float64 `json:"temperature,omitempty"`
	TopP            *float64 `json:"topP,omitempty"`
	MaxOutputTokens *int     `json:"maxOutputTokens,omitempty"`
	StopSequences   []string `json:"stopSequences,omitempty"`
}

type gemResponse struct {
	Candidates []struct {
		Content      gemContent `json:"content"`
		FinishReason string     `json:"finishReason"`
		Index        int        `json:"index"`
	} `json:"candidates"`
	UsageMetadata struct {
		PromptTokenCount     int `json:"promptTokenCount"`
		CandidatesTokenCount int `json:"candidatesTokenCount"`
		TotalTokenCount      int `json:"totalTokenCount"`
	} `json:"usageMetadata"`
	Error *struct {
		Code    int    `json:"code"`
		Message string `json:"message"`
		Status  string `json:"status"`
	} `json:"error,omitempty"`
}

// -----------------------------------------------------------------------------
// Request conversion
// -----------------------------------------------------------------------------

func (g *Google) toRequest(req ChatRequest) gemRequest {
	out := gemRequest{}
	if req.Temperature != nil || req.TopP != nil || req.MaxTokens != nil || len(req.Stop) > 0 {
		out.GenerationConfig = &gemGenConfig{
			Temperature:     req.Temperature,
			TopP:            req.TopP,
			MaxOutputTokens: req.MaxTokens,
			StopSequences:   req.Stop,
		}
	}
	for _, t := range req.Tools {
		out.Tools = append(out.Tools, gemTool{
			FunctionDeclarations: []gemFunctionDecl{{
				Name:        t.Function.Name,
				Description: t.Function.Description,
				Parameters:  t.Function.Parameters,
			}},
		})
	}
	for _, m := range req.Messages {
		if m.Role == RoleSystem {
			out.SystemInstruction = &gemContent{
				Parts: []gemPart{{Text: m.Content}},
			}
			continue
		}
		role := "user"
		if m.Role == RoleAssistant {
			role = "model"
		}
		if m.Role == RoleTool {
			// Tool results are a "user" turn with a functionResponse part.
			out.Contents = append(out.Contents, gemContent{
				Role: "user",
				Parts: []gemPart{{
					FunctionResp: &gemFunctionResp{
						Name:     m.Name,
						Response: map[string]any{"result": m.Content},
					},
				}},
			})
			continue
		}
		var parts []gemPart
		if m.Content != "" {
			parts = append(parts, gemPart{Text: m.Content})
		}
		for _, tc := range m.ToolCalls {
			var args map[string]any
			_ = json.Unmarshal([]byte(tc.Function.Arguments), &args)
			parts = append(parts, gemPart{
				FunctionCall: &gemFunctionCall{
					Name: tc.Function.Name,
					Args: args,
				},
			})
		}
		if len(parts) == 0 {
			continue
		}
		out.Contents = append(out.Contents, gemContent{Role: role, Parts: parts})
	}
	return out
}

func (g *Google) fromResponse(r gemResponse) (ChatResponse, error) {
	if len(r.Candidates) == 0 {
		return ChatResponse{}, fmt.Errorf("%w: no candidates", ErrResponseShape)
	}
	c := r.Candidates[0]
	resp := ChatResponse{
		Model:        r.Candidates[0].Content.Parts[0].Text, // best-effort
		FinishReason: mapGemFinishReason(c.FinishReason),
		Usage: Usage{
			InputTokens:  r.UsageMetadata.PromptTokenCount,
			OutputTokens: r.UsageMetadata.CandidatesTokenCount,
			TotalTokens:  r.UsageMetadata.TotalTokenCount,
		},
		Message: Message{Role: RoleAssistant},
	}
	for _, p := range c.Content.Parts {
		if p.Text != "" {
			resp.Message.Content += p.Text
		}
		if p.FunctionCall != nil {
			args, _ := json.Marshal(p.FunctionCall.Args)
			tc := ToolCall{Type: "function"}
			tc.Function.Name = p.FunctionCall.Name
			tc.Function.Arguments = string(args)
			resp.Message.ToolCalls = append(resp.Message.ToolCalls, tc)
		}
	}
	return resp, nil
}

func mapGemFinishReason(s string) FinishReason {
	switch s {
	case "STOP":
		return FinishStop
	case "MAX_TOKENS":
		return FinishLength
	case "SAFETY":
		return FinishContentFilter
	case "RECITATION":
		return FinishContentFilter
	default:
		return FinishReason(s)
	}
}

// -----------------------------------------------------------------------------
// HTTP
// -----------------------------------------------------------------------------

func (g *Google) buildRequest(ctx context.Context, model string, body any, stream bool) (*http.Request, error) {
	data, err := json.Marshal(body)
	if err != nil {
		return nil, err
	}
	method := "generateContent"
	if stream {
		method = "streamGenerateContent"
	}
	url := fmt.Sprintf("%s/v1beta/models/%s:%s?key=%s", g.BaseURL, model, method, g.APIKey)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(data))
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	return req, nil
}

func (g *Google) Chat(ctx context.Context, req ChatRequest) (ChatResponse, error) {
	if req.Model == "" {
		return ChatResponse{}, ErrNoModel
	}
	if len(req.Messages) == 0 {
		return ChatResponse{}, ErrNoMessages
	}
	if req.Stream {
		ch, cancel, err := g.Stream(ctx, req)
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
	body := g.toRequest(req)
	httpReq, err := g.buildRequest(ctx, req.Model, body, false)
	if err != nil {
		return ChatResponse{}, err
	}
	resp, err := g.HTTPClient.Do(httpReq)
	if err != nil {
		return ChatResponse{}, fmt.Errorf("llm/google: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()
	raw, _ := io.ReadAll(resp.Body)
	if resp.StatusCode >= 400 {
		return ChatResponse{}, fmt.Errorf("llm/google: %d: %s", resp.StatusCode, string(raw))
	}
	var r gemResponse
	if err := json.Unmarshal(raw, &r); err != nil {
		return ChatResponse{}, fmt.Errorf("%w: %v", ErrResponseShape, err)
	}
	if r.Error != nil {
		return ChatResponse{}, fmt.Errorf("llm/google: %d %s: %s", r.Error.Code, r.Error.Status, r.Error.Message)
	}
	cr, err := g.fromResponse(r)
	if err != nil {
		return ChatResponse{}, err
	}
	cr.Raw = raw
	return cr, nil
}

func (g *Google) Stream(ctx context.Context, req ChatRequest) (<-chan StreamEvent, func(), error) {
	if req.Model == "" {
		return nil, nil, ErrNoModel
	}
	if len(req.Messages) == 0 {
		return nil, nil, ErrNoMessages
	}
	body := g.toRequest(req)
	httpReq, err := g.buildRequest(ctx, req.Model, body, true)
	if err != nil {
		return nil, nil, err
	}
	resp, err := g.HTTPClient.Do(httpReq)
	if err != nil {
		return nil, nil, fmt.Errorf("llm/google: %w", err)
	}
	if resp.StatusCode >= 400 {
		b, _ := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		return nil, nil, fmt.Errorf("llm/google: %d: %s", resp.StatusCode, string(b))
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
		// Gemini streams either a JSON array of objects, or one
		// newline-delimited JSON object per chunk. We use a state machine
		// driven by json.Decoder to read individual JSON values.
		dec := json.NewDecoder(reader)
		// Read the opening '[' if present, then loop over values until ']'.
		tok, err := dec.Token()
		if err == nil {
			if delim, ok := tok.(json.Delim); ok && delim == '[' {
				// Array form.
				for dec.More() {
					var r gemResponse
					if err := dec.Decode(&r); err != nil {
						out <- StreamEvent{Err: fmt.Errorf("llm/google: parse: %w", err), Done: true}
						return
					}
					emitGemResponse(&r, out, &accumulated, &finishReason, &usage)
				}
			} else {
				// Single-object form. Re-feed the token by decoding it.
				if tok != nil {
					// Reconstruct the value as a single gemResponse.
					// Easier: read the rest of the body and parse.
					rest, _ := io.ReadAll(reader)
					combined := append(append([]byte{}, fmt.Sprintf("%v", tok)...), rest...)
					var r gemResponse
					if err := json.Unmarshal(combined, &r); err != nil {
						out <- StreamEvent{Err: fmt.Errorf("llm/google: parse: %w", err), Done: true}
						return
					}
					emitGemResponse(&r, out, &accumulated, &finishReason, &usage)
				}
			}
		}
		out <- StreamEvent{
			FinishReason: finishReason,
			Usage:        usage,
			Done:         true,
		}
	}()
	return out, func() { close(cancel) }, nil
}

func emitGemResponse(r *gemResponse, out chan<- StreamEvent, accumulated *strings.Builder, finishReason *FinishReason, usage *Usage) {
	for _, c := range r.Candidates {
		for _, p := range c.Content.Parts {
			if p.Text != "" {
				accumulated.WriteString(p.Text)
				out <- StreamEvent{Delta: Message{Role: RoleAssistant, Content: p.Text}}
			}
		}
		if c.FinishReason != "" {
			*finishReason = mapGemFinishReason(c.FinishReason)
		}
	}
	if r.UsageMetadata.TotalTokenCount > 0 {
		*usage = Usage{
			InputTokens:  r.UsageMetadata.PromptTokenCount,
			OutputTokens: r.UsageMetadata.CandidatesTokenCount,
			TotalTokens:  r.UsageMetadata.TotalTokenCount,
		}
	}
}
