package llm

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// Google — additional coverage
// -----------------------------------------------------------------------------

func TestGoogle_Models(t *testing.T) {
	p := NewGoogle("k", []ModelInfo{{ID: "a"}, {ID: "b"}})
	assert.Len(t, p.Models(), 2)
}

func TestGoogle_DefaultModel_Empty(t *testing.T) {
	p := NewGoogle("k", []ModelInfo{{ID: "only"}})
	assert.Equal(t, "only", p.DefaultModel("chat"))
}

func TestGoogle_DefaultModel_NoFlash(t *testing.T) {
	p := NewGoogle("k", []ModelInfo{{ID: "gemini-1.5-pro"}})
	assert.Equal(t, "gemini-1.5-pro", p.DefaultModel("chat"))
}

func TestGoogle_DefaultModel_EmptyList(t *testing.T) {
	p := NewGoogle("k", nil)
	assert.Equal(t, "", p.DefaultModel("chat"))
}

func TestGoogle_Stream_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, ":streamGenerateContent")
		// Gemini returns an array of objects. The parser splits on "},{".
		// Keep the test data simple so the split is unambiguous.
		_, _ = w.Write([]byte(`[`))
		_, _ = w.Write([]byte(`{"candidates":[{"content":{"role":"model","parts":[{"text":"Hi"}]},"finishReason":""}],"usageMetadata":{"promptTokenCount":2,"candidatesTokenCount":1,"totalTokenCount":3}}`))
		_, _ = w.Write([]byte(`,{"candidates":[{"content":{"role":"model","parts":[]},"finishReason":"STOP"}]}`))
		_, _ = w.Write([]byte(`]`))
	}))
	defer srv.Close()
	p := NewGoogle("k", nil)
	p.BaseURL = srv.URL
	ch, cancel, err := p.Stream(context.Background(), ChatRequest{
		Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}},
	})
	require.NoError(t, err)
	defer cancel()
	var content strings.Builder
	var finish FinishReason
	for ev := range ch {
		if ev.Err != nil {
			t.Fatal(ev.Err)
		}
		content.WriteString(ev.Delta.Content)
		if ev.FinishReason != "" {
			finish = ev.FinishReason
		}
	}
	assert.Equal(t, "Hi", content.String())
	assert.Equal(t, FinishStop, finish)
}

func TestGoogle_Stream_4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`forbidden`))
	}))
	defer srv.Close()
	p := NewGoogle("k", nil)
	p.BaseURL = srv.URL
	_, _, err := p.Stream(context.Background(), ChatRequest{
		Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}},
	})
	assert.Error(t, err)
}

func TestGoogle_Stream_NoModel(t *testing.T) {
	p := NewGoogle("k", nil)
	_, _, err := p.Stream(context.Background(), ChatRequest{Messages: []Message{{Role: RoleUser, Content: "x"}}})
	assert.ErrorIs(t, err, ErrNoModel)
}

func TestGoogle_Stream_NoMessages(t *testing.T) {
	p := NewGoogle("k", nil)
	_, _, err := p.Stream(context.Background(), ChatRequest{Model: "m"})
	assert.ErrorIs(t, err, ErrNoMessages)
}

func TestGoogle_Chat_InlineError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"error":{"code":429,"message":"rate","status":"RESOURCE_EXHAUSTED"}}`))
	}))
	defer srv.Close()
	p := NewGoogle("k", nil)
	p.BaseURL = srv.URL
	_, err := p.Chat(context.Background(), ChatRequest{
		Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}},
	})
	assert.Error(t, err)
}

func TestGoogle_Chat_ToolUse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(gemResponse{
			Candidates: []struct {
				Content      gemContent `json:"content"`
				FinishReason string     `json:"finishReason"`
				Index        int        `json:"index"`
			}{{
				Content: gemContent{
					Role: "model",
					Parts: []gemPart{{
						FunctionCall: &gemFunctionCall{
							Name: "get_weather",
							Args: map[string]any{"location": "Tokyo"},
						},
					}},
				},
				FinishReason: "STOP",
			}},
		})
	}))
	defer srv.Close()
	p := NewGoogle("k", nil)
	p.BaseURL = srv.URL
	resp, err := p.Chat(context.Background(), ChatRequest{
		Model: "m", Messages: []Message{{Role: RoleUser, Content: "weather?"}},
	})
	require.NoError(t, err)
	require.Len(t, resp.Message.ToolCalls, 1)
	assert.Equal(t, "get_weather", resp.Message.ToolCalls[0].Function.Name)
	assert.Contains(t, resp.Message.ToolCalls[0].Function.Arguments, "Tokyo")
}

func TestGoogle_Chat_MaxTokensReason(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(gemResponse{
			Candidates: []struct {
				Content      gemContent `json:"content"`
				FinishReason string     `json:"finishReason"`
				Index        int        `json:"index"`
			}{{
				Content:      gemContent{Parts: []gemPart{{Text: "x"}}},
				FinishReason: "MAX_TOKENS",
			}},
		})
	}))
	defer srv.Close()
	p := NewGoogle("k", nil)
	p.BaseURL = srv.URL
	resp, err := p.Chat(context.Background(), ChatRequest{
		Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}},
	})
	require.NoError(t, err)
	assert.Equal(t, FinishLength, resp.FinishReason)
}

func TestGoogle_Chat_SafetyReason(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(gemResponse{
			Candidates: []struct {
				Content      gemContent `json:"content"`
				FinishReason string     `json:"finishReason"`
				Index        int        `json:"index"`
			}{{
				Content:      gemContent{Parts: []gemPart{{Text: ""}}},
				FinishReason: "SAFETY",
			}},
		})
	}))
	defer srv.Close()
	p := NewGoogle("k", nil)
	p.BaseURL = srv.URL
	resp, err := p.Chat(context.Background(), ChatRequest{
		Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}},
	})
	require.NoError(t, err)
	assert.Equal(t, FinishContentFilter, resp.FinishReason)
}

func TestGoogle_ToolMessage_Translation(t *testing.T) {
	captured := gemRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		_ = json.NewEncoder(w).Encode(gemResponse{
			Candidates: []struct {
				Content      gemContent `json:"content"`
				FinishReason string     `json:"finishReason"`
				Index        int        `json:"index"`
			}{{
				Content:      gemContent{Parts: []gemPart{{Text: "ok"}}},
				FinishReason: "STOP",
			}},
		})
	}))
	defer srv.Close()
	p := NewGoogle("k", nil)
	p.BaseURL = srv.URL
	_, err := p.Chat(context.Background(), ChatRequest{
		Model: "m",
		Messages: []Message{
			{Role: RoleUser, Content: "weather?"},
			{Role: RoleTool, Name: "get_weather", Content: "sunny"},
		},
	})
	require.NoError(t, err)
	// Tool message is a "user" turn with functionResponse.
	require.GreaterOrEqual(t, len(captured.Contents), 2)
	last := captured.Contents[len(captured.Contents)-1]
	assert.Equal(t, "user", last.Role)
	require.Len(t, last.Parts, 1)
	require.NotNil(t, last.Parts[0].FunctionResp)
	assert.Equal(t, "get_weather", last.Parts[0].FunctionResp.Name)
}

func TestGoogle_AssistantToolCalls_Translation(t *testing.T) {
	captured := gemRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		_ = json.NewEncoder(w).Encode(gemResponse{
			Candidates: []struct {
				Content      gemContent `json:"content"`
				FinishReason string     `json:"finishReason"`
				Index        int        `json:"index"`
			}{{
				Content:      gemContent{Parts: []gemPart{{Text: "ok"}}},
				FinishReason: "STOP",
			}},
		})
	}))
	defer srv.Close()
	p := NewGoogle("k", nil)
	p.BaseURL = srv.URL
	_, err := p.Chat(context.Background(), ChatRequest{
		Model: "m",
		Messages: []Message{
			{Role: RoleUser, Content: "x"},
			{Role: RoleAssistant, ToolCalls: []ToolCall{{
				ID: "t1",
				Function: struct {
					Name      string `json:"name"`
					Arguments string `json:"arguments"`
				}{Name: "fn", Arguments: `{"a":1}`},
			}}},
		},
	})
	require.NoError(t, err)
	require.GreaterOrEqual(t, len(captured.Contents), 2)
	asst := captured.Contents[1]
	assert.Equal(t, "model", asst.Role)
	require.GreaterOrEqual(t, len(asst.Parts), 1)
	require.NotNil(t, asst.Parts[0].FunctionCall)
	assert.Equal(t, "fn", asst.Parts[0].FunctionCall.Name)
}

func TestGoogle_Tools_Translation(t *testing.T) {
	captured := gemRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		_ = json.NewEncoder(w).Encode(gemResponse{
			Candidates: []struct {
				Content      gemContent `json:"content"`
				FinishReason string     `json:"finishReason"`
				Index        int        `json:"index"`
			}{{
				Content:      gemContent{Parts: []gemPart{{Text: "x"}}},
				FinishReason: "STOP",
			}},
		})
	}))
	defer srv.Close()
	p := NewGoogle("k", nil)
	p.BaseURL = srv.URL
	td := ToolDefinition{Type: "function"}
	td.Function.Name = "fn"
	td.Function.Description = "do thing"
	td.Function.Parameters = map[string]any{"type": "object"}
	_, err := p.Chat(context.Background(), ChatRequest{
		Model:    "m",
		Tools:    []ToolDefinition{td},
		Messages: []Message{{Role: RoleUser, Content: "x"}},
	})
	require.NoError(t, err)
	require.NotNil(t, captured.Tools)
	require.Len(t, captured.Tools, 1)
	assert.Equal(t, "fn", captured.Tools[0].FunctionDeclarations[0].Name)
}

func TestGoogle_GenerationConfig(t *testing.T) {
	captured := gemRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		_ = json.NewEncoder(w).Encode(gemResponse{
			Candidates: []struct {
				Content      gemContent `json:"content"`
				FinishReason string     `json:"finishReason"`
				Index        int        `json:"index"`
			}{{
				Content:      gemContent{Parts: []gemPart{{Text: "x"}}},
				FinishReason: "STOP",
			}},
		})
	}))
	defer srv.Close()
	p := NewGoogle("k", nil)
	p.BaseURL = srv.URL
	temp := 0.5
	maxT := 256
	_, err := p.Chat(context.Background(), ChatRequest{
		Model:       "m",
		Temperature: &temp,
		MaxTokens:   &maxT,
		Stop:        []string{"END"},
		Messages:    []Message{{Role: RoleUser, Content: "x"}},
	})
	require.NoError(t, err)
	require.NotNil(t, captured.GenerationConfig)
	assert.Equal(t, 0.5, *captured.GenerationConfig.Temperature)
	assert.Equal(t, 256, *captured.GenerationConfig.MaxOutputTokens)
	assert.Equal(t, []string{"END"}, captured.GenerationConfig.StopSequences)
}

func TestGoogle_SystemInstruction(t *testing.T) {
	captured := gemRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		_ = json.NewEncoder(w).Encode(gemResponse{
			Candidates: []struct {
				Content      gemContent `json:"content"`
				FinishReason string     `json:"finishReason"`
				Index        int        `json:"index"`
			}{{
				Content:      gemContent{Parts: []gemPart{{Text: "x"}}},
				FinishReason: "STOP",
			}},
		})
	}))
	defer srv.Close()
	p := NewGoogle("k", nil)
	p.BaseURL = srv.URL
	_, err := p.Chat(context.Background(), ChatRequest{
		Model: "m",
		Messages: []Message{
			{Role: RoleSystem, Content: "Be brief."},
			{Role: RoleUser, Content: "x"},
		},
	})
	require.NoError(t, err)
	require.NotNil(t, captured.SystemInstruction)
	assert.Equal(t, "Be brief.", captured.SystemInstruction.Parts[0].Text)
}

// -----------------------------------------------------------------------------
// Anthropic — additional coverage
// -----------------------------------------------------------------------------

func TestAnthropic_DefaultModel_NoList(t *testing.T) {
	p := NewAnthropic("k", nil)
	assert.Equal(t, "", p.DefaultModel("chat"))
}

func TestAnthropic_Models(t *testing.T) {
	p := NewAnthropic("k", []ModelInfo{{ID: "x"}})
	assert.Len(t, p.Models(), 1)
}

func TestAnthropic_MapStopReason(t *testing.T) {
	assert.Equal(t, FinishStop, mapAnthStopReason("end_turn"))
	assert.Equal(t, FinishLength, mapAnthStopReason("max_tokens"))
	assert.Equal(t, FinishStop, mapAnthStopReason("stop_sequence"))
	assert.Equal(t, FinishToolCalls, mapAnthStopReason("tool_use"))
	assert.Equal(t, FinishReason("weird"), mapAnthStopReason("weird"))
}

func TestAnthropic_Stream_4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`unauthorized`))
	}))
	defer srv.Close()
	p := NewAnthropic("k", nil)
	p.BaseURL = srv.URL
	_, _, err := p.Stream(context.Background(), ChatRequest{
		Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}},
	})
	assert.Error(t, err)
}

func TestAnthropic_Stream_NoModel(t *testing.T) {
	p := NewAnthropic("k", nil)
	_, _, err := p.Stream(context.Background(), ChatRequest{Messages: []Message{{Role: RoleUser, Content: "x"}}})
	assert.ErrorIs(t, err, ErrNoModel)
}

func TestAnthropic_Stream_NoMessages(t *testing.T) {
	p := NewAnthropic("k", nil)
	_, _, err := p.Stream(context.Background(), ChatRequest{Model: "m"})
	assert.ErrorIs(t, err, ErrNoMessages)
}

func TestAnthropic_Tools_Translation(t *testing.T) {
	captured := anthRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		_ = json.NewEncoder(w).Encode(anthResponse{
			ID:         "x",
			StopReason: "end_turn",
			Content: []struct {
				Type  string `json:"type"`
				Text  string `json:"text,omitempty"`
				ID    string `json:"id,omitempty"`
				Name  string `json:"name,omitempty"`
				Input any    `json:"input,omitempty"`
			}{{
				Type: "text",
				Text: "ok",
			}},
		})
	}))
	defer srv.Close()
	p := NewAnthropic("k", nil)
	p.BaseURL = srv.URL
	td := ToolDefinition{Type: "function"}
	td.Function.Name = "fn"
	td.Function.Description = "d"
	td.Function.Parameters = map[string]any{"type": "object"}
	maxTokens := 100
	_, err := p.Chat(context.Background(), ChatRequest{
		Model:     "m",
		Tools:     []ToolDefinition{td},
		MaxTokens: &maxTokens,
		Messages:  []Message{{Role: RoleUser, Content: "x"}},
	})
	require.NoError(t, err)
	require.Len(t, captured.Tools, 1)
	assert.Equal(t, "fn", captured.Tools[0].Name)
	assert.Equal(t, 100, captured.MaxTokens)
}

func TestAnthropic_StopSequences(t *testing.T) {
	captured := anthRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		_ = json.NewEncoder(w).Encode(anthResponse{
			ID: "x", StopReason: "end_turn",
			Content: []struct {
				Type  string `json:"type"`
				Text  string `json:"text,omitempty"`
				ID    string `json:"id,omitempty"`
				Name  string `json:"name,omitempty"`
				Input any    `json:"input,omitempty"`
			}{{
				Type: "text", Text: "x",
			}},
		})
	}))
	defer srv.Close()
	p := NewAnthropic("k", nil)
	p.BaseURL = srv.URL
	_, err := p.Chat(context.Background(), ChatRequest{
		Model:    "m",
		Stop:     []string{"STOP1", "STOP2"},
		Messages: []Message{{Role: RoleUser, Content: "x"}},
	})
	require.NoError(t, err)
	assert.Equal(t, []string{"STOP1", "STOP2"}, captured.StopSeqs)
}

// -----------------------------------------------------------------------------
// OpenAI-compat — additional coverage
// -----------------------------------------------------------------------------

func TestOpenAICompat_Tools_Translation(t *testing.T) {
	captured := oaiRequest{}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		require.NoError(t, json.NewDecoder(r.Body).Decode(&captured))
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer srv.Close()
	p := NewOpenAICompat("t", srv.URL, "k")
	p.ModelsList = []ModelInfo{{ID: "m"}}
	td := ToolDefinition{Type: "function"}
	td.Function.Name = "fn"
	td.Function.Description = "d"
	td.Function.Parameters = map[string]any{"type": "object"}
	temp := 0.3
	_, err := p.Chat(context.Background(), ChatRequest{
		Model:       "m",
		Tools:       []ToolDefinition{td},
		Temperature: &temp,
		ToolChoice:  "auto",
		Messages:    []Message{{Role: RoleUser, Content: "x"}},
	})
	require.NoError(t, err)
	require.Len(t, captured.Tools, 1)
	assert.Equal(t, "fn", captured.Tools[0].Function.Name)
	assert.Equal(t, "auto", captured.ToolChoice)
	assert.Equal(t, 0.3, *captured.Temperature)
}

func TestOpenAICompat_StreamError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Write a partial line and then close abruptly.
		_, _ = w.Write([]byte(`data: {"choices":[{"delta":{"content":"hi"}}]}` + "\n"))
		if f, ok := w.(http.Flusher); ok {
			f.Flush()
		}
		// Hijack + close to force an error in the reader.
		if hj, ok := w.(http.Hijacker); ok {
			conn, _, _ := hj.Hijack()
			_ = conn.Close()
			return
		}
	}))
	defer srv.Close()
	p := NewOpenAICompat("t", srv.URL, "k")
	p.ModelsList = []ModelInfo{{ID: "m"}}
	ch, cancel, err := p.Stream(context.Background(), ChatRequest{Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}}})
	require.NoError(t, err)
	defer cancel()
	var sawError bool
	for ev := range ch {
		if ev.Err != nil {
			sawError = true
		}
	}
	assert.True(t, sawError, "stream should report error when connection drops")
}

func TestOpenAICompat_StreamCancelMidway(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		flusher, _ := w.(http.Flusher)
		for i := 0; i < 10; i++ {
			_, _ = w.Write([]byte(`data: {"choices":[{"delta":{"content":"chunk"}}]}` + "\n\n"))
			if flusher != nil {
				flusher.Flush()
			}
		}
	}))
	defer srv.Close()
	p := NewOpenAICompat("t", srv.URL, "k")
	p.ModelsList = []ModelInfo{{ID: "m"}}
	ch, cancel, err := p.Stream(context.Background(), ChatRequest{Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}}})
	require.NoError(t, err)
	// Read one event then cancel.
	<-ch
	cancel()
	for range ch {
	}
}

func TestOpenAICompat_StreamBadChunk(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`data: not-valid-json` + "\n\n"))
		_, _ = w.Write([]byte(`data: [DONE]` + "\n\n"))
	}))
	defer srv.Close()
	p := NewOpenAICompat("t", srv.URL, "k")
	p.ModelsList = []ModelInfo{{ID: "m"}}
	ch, cancel, err := p.Stream(context.Background(), ChatRequest{Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}}})
	require.NoError(t, err)
	defer cancel()
	var sawErr bool
	for ev := range ch {
		if ev.Err != nil {
			sawErr = true
		}
	}
	assert.True(t, sawErr)
}

func TestOpenAICompat_Chat_NoStream(t *testing.T) {
	// Stream=true should still work via the drain path.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		flusher, _ := w.(http.Flusher)
		_, _ = w.Write([]byte(`data: {"choices":[{"delta":{"content":"hi "}}]}` + "\n\n"))
		_, _ = w.Write([]byte(`data: {"choices":[{"delta":{"content":"there"},"finish_reason":"stop"}]}` + "\n\n"))
		_, _ = w.Write([]byte(`data: [DONE]` + "\n\n"))
		if flusher != nil {
			flusher.Flush()
		}
	}))
	defer srv.Close()
	p := NewOpenAICompat("t", srv.URL, "k")
	p.ModelsList = []ModelInfo{{ID: "m"}}
	resp, err := p.Chat(context.Background(), ChatRequest{
		Model: "m", Stream: true, Messages: []Message{{Role: RoleUser, Content: "x"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "hi there", resp.Message.Content)
}

func TestOpenAICompat_Stream_NoModel(t *testing.T) {
	p := NewOpenAICompat("t", "http://x", "k")
	_, _, err := p.Stream(context.Background(), ChatRequest{Messages: []Message{{Role: RoleUser, Content: "x"}}})
	assert.ErrorIs(t, err, ErrNoModel)
}

func TestOpenAICompat_Stream_NoMessages(t *testing.T) {
	p := NewOpenAICompat("t", "http://x", "k")
	_, _, err := p.Stream(context.Background(), ChatRequest{Model: "m"})
	assert.ErrorIs(t, err, ErrNoMessages)
}

func TestOpenAICompat_NameAndModels(t *testing.T) {
	p := NewOpenAICompat("custom", "http://x", "k")
	p.ModelsList = []ModelInfo{{ID: "a"}}
	assert.Equal(t, "custom", p.Name())
	assert.Len(t, p.Models(), 1)
}

func TestOpenAICompat_DefaultModel_NoQualifying(t *testing.T) {
	p := NewOpenAICompat("t", "http://x", "k")
	p.ModelsList = []ModelInfo{
		{ID: "a", SupportsVision: false},
	}
	// No vision-supporting model; fall back to first.
	assert.Equal(t, "a", p.DefaultModel("vision"))
}

func TestOpenAICompat_Client_Default(t *testing.T) {
	// When HTTPClient is nil, client() returns a default one.
	p := NewOpenAICompat("t", "http://x", "k")
	p.HTTPClient = nil
	c := p.client()
	assert.NotNil(t, c)
}

// -----------------------------------------------------------------------------
// Registry — additional
// -----------------------------------------------------------------------------

func TestRegistry_Chat_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[{"message":{"role":"assistant","content":"ok"},"finish_reason":"stop"}],"usage":{"prompt_tokens":1,"completion_tokens":1,"total_tokens":2}}`))
	}))
	defer srv.Close()
	r := NewRegistry()
	provider := NewOpenAICompat("test", srv.URL, "k")
	provider.ModelsList = []ModelInfo{{ID: "m"}}
	r.Register(provider)
	resp, err := r.Chat(context.Background(), "test", ChatRequest{
		Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "ok", resp.Message.Content)
}

func TestRegistry_Stream_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`data: {"choices":[{"delta":{"content":"x"}}]}` + "\n\n"))
		_, _ = w.Write([]byte(`data: [DONE]` + "\n\n"))
	}))
	defer srv.Close()
	r := NewRegistry()
	provider := NewOpenAICompat("test", srv.URL, "k")
	provider.ModelsList = []ModelInfo{{ID: "m"}}
	r.Register(provider)
	ch, cancel, err := r.Stream(context.Background(), "test", ChatRequest{
		Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}},
	})
	require.NoError(t, err)
	defer cancel()
	<-ch
}
