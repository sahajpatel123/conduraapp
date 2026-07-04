package llm

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// Types
// -----------------------------------------------------------------------------

func TestRole_IsValid(t *testing.T) {
	for _, r := range []Role{RoleSystem, RoleUser, RoleAssistant, RoleTool} {
		assert.True(t, r.IsValid(), r)
	}
	assert.False(t, Role("weird").IsValid())
}

func TestUsage_Add(t *testing.T) {
	u1 := Usage{InputTokens: 10, OutputTokens: 20, TotalTokens: 30}
	u2 := Usage{InputTokens: 1, OutputTokens: 2, TotalTokens: 3}
	u1.Add(u2)
	assert.Equal(t, 11, u1.InputTokens)
	assert.Equal(t, 22, u1.OutputTokens)
	assert.Equal(t, 33, u1.TotalTokens)
}

func TestCopyBody(t *testing.T) {
	// nil body.
	b, r, err := CopyBody(nil)
	assert.NoError(t, err)
	assert.Nil(t, b)
	assert.Nil(t, r)
}

func TestCopyBody_WithBody(t *testing.T) {
	body := io.NopCloser(strings.NewReader("hello world"))
	b, r, err := CopyBody(body)
	require.NoError(t, err)
	assert.Equal(t, []byte("hello world"), b)
	got, _ := readAll(r)
	assert.Equal(t, "hello world", got)
}

func TestCopyBody_Error(t *testing.T) {
	body := &errReadCloser{}
	_, _, err := CopyBody(body)
	assert.Error(t, err)
}

type errReadCloser struct{}

func (errReadCloser) Read(p []byte) (int, error) { return 0, errBoom }
func (errReadCloser) Close() error               { return nil }

type stringErr string

func (e stringErr) Error() string { return string(e) }

var (
	errBoom = stringErr("boom")
)

func readAll(r interface{ Read(p []byte) (int, error) }) (string, error) {
	var b [128]byte
	n, err := r.Read(b[:])
	return string(b[:n]), err
}

// -----------------------------------------------------------------------------
// Model pricing
// -----------------------------------------------------------------------------

func TestLookupModel_Known(t *testing.T) {
	m, ok := LookupModel("gpt-4o")
	require.True(t, ok)
	assert.Equal(t, "GPT-4o", m.DisplayName)
	assert.Equal(t, 2.50, m.InputCostPerMTok)
}

func TestLookupModel_Unknown(t *testing.T) {
	_, ok := LookupModel("not-a-model")
	assert.False(t, ok)
}

func TestEstimateCost(t *testing.T) {
	cost := EstimateCost("gpt-4o", Usage{InputTokens: 1_000_000, OutputTokens: 0})
	assert.InDelta(t, 2.50, cost, 0.0001)
	cost = EstimateCost("gpt-4o", Usage{InputTokens: 0, OutputTokens: 1_000_000})
	assert.InDelta(t, 10.0, cost, 0.0001)
}

func TestEstimateCost_UnknownModel(t *testing.T) {
	assert.Equal(t, 0.0, EstimateCost("nope", Usage{InputTokens: 100}))
}

func TestEstimateCostFromInfo(t *testing.T) {
	info := ModelInfo{InputCostPerMTok: 1.0, OutputCostPerMTok: 2.0}
	cost := EstimateCostFromInfo(info, Usage{InputTokens: 500_000, OutputTokens: 500_000})
	assert.InDelta(t, 1.5, cost, 0.0001)
}

func TestRegisterUnregister(t *testing.T) {
	m := ModelInfo{ID: "test-model", InputCostPerMTok: 1, OutputCostPerMTok: 2}
	RegisterModel(m)
	got, ok := LookupModel("test-model")
	require.True(t, ok)
	assert.Equal(t, 1.0, got.InputCostPerMTok)
	UnregisterModel("test-model")
	_, ok = LookupModel("test-model")
	assert.False(t, ok)
}

// -----------------------------------------------------------------------------
// OpenAI-compat: Chat
// -----------------------------------------------------------------------------

func TestOpenAICompat_Chat_OK(t *testing.T) {
	var gotAuth string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("Authorization")
		assert.Equal(t, "/v1/chat/completions", r.URL.Path)
		var body oaiRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "test-model", body.Model)
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(oaiResponse{
			ID:    "chatcmpl-1",
			Model: "test-model",
			Choices: []struct {
				Index        int        `json:"index"`
				Message      oaiMessage `json:"message"`
				FinishReason string     `json:"finish_reason"`
			}{{
				Index: 0,
				Message: oaiMessage{
					Role:    "assistant",
					Content: "hello back",
				},
				FinishReason: "stop",
			}},
			Usage: struct {
				PromptTokens     int `json:"prompt_tokens"`
				CompletionTokens int `json:"completion_tokens"`
				TotalTokens      int `json:"total_tokens"`
			}{PromptTokens: 5, CompletionTokens: 2, TotalTokens: 7},
		})
	}))
	defer srv.Close()

	p := NewOpenAICompat("test", srv.URL+"/v1", "sk-test")
	p.ModelsList = []ModelInfo{{ID: "test-model"}}
	resp, err := p.Chat(context.Background(), ChatRequest{
		Model:    "test-model",
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "hello back", resp.Message.Content)
	assert.Equal(t, "stop", string(resp.FinishReason))
	assert.Equal(t, 5, resp.Usage.InputTokens)
	assert.Equal(t, 2, resp.Usage.OutputTokens)
	assert.Equal(t, "Bearer sk-test", gotAuth)
}

func TestOpenAICompat_Chat_4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(401)
		_, _ = w.Write([]byte(`{"error":{"message":"bad key"}}`))
	}))
	defer srv.Close()
	p := NewOpenAICompat("test", srv.URL, "k")
	p.ModelsList = []ModelInfo{{ID: "m"}}
	_, err := p.Chat(context.Background(), ChatRequest{Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}}})
	assert.Error(t, err)
}

func TestOpenAICompat_Chat_InlineError(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(oaiResponse{
			Error: &struct {
				Message string `json:"message"`
				Type    string `json:"type"`
				Code    any    `json:"code"`
			}{Message: "no quota", Type: "rate_limit"},
		})
	}))
	defer srv.Close()
	p := NewOpenAICompat("test", srv.URL, "k")
	p.ModelsList = []ModelInfo{{ID: "m"}}
	_, err := p.Chat(context.Background(), ChatRequest{Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}}})
	assert.Error(t, err)
	assert.Contains(t, err.Error(), "no quota")
}

func TestOpenAICompat_Chat_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()
	p := NewOpenAICompat("test", srv.URL, "k")
	p.ModelsList = []ModelInfo{{ID: "m"}}
	_, err := p.Chat(context.Background(), ChatRequest{Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}}})
	assert.Error(t, err)
}

func TestOpenAICompat_Chat_NoChoices(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"choices":[]}`))
	}))
	defer srv.Close()
	p := NewOpenAICompat("test", srv.URL, "k")
	p.ModelsList = []ModelInfo{{ID: "m"}}
	_, err := p.Chat(context.Background(), ChatRequest{Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}}})
	assert.Error(t, err)
}

func TestOpenAICompat_NoModel(t *testing.T) {
	p := NewOpenAICompat("test", "http://localhost", "k")
	_, err := p.Chat(context.Background(), ChatRequest{Messages: []Message{{Role: RoleUser, Content: "x"}}})
	assert.ErrorIs(t, err, ErrNoModel)
}

func TestOpenAICompat_NoMessages(t *testing.T) {
	p := NewOpenAICompat("test", "http://localhost", "k")
	_, err := p.Chat(context.Background(), ChatRequest{Model: "m"})
	assert.ErrorIs(t, err, ErrNoMessages)
}

func TestOpenAICompat_Stream_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		flusher, _ := w.(http.Flusher)
		chunks := []string{
			`data: {"choices":[{"delta":{"role":"assistant","content":"Hel"}}]}`,
			``,
			`data: {"choices":[{"delta":{"content":"lo "}}]}`,
			``,
			`data: {"choices":[{"delta":{"content":"world"},"finish_reason":"stop"}],"usage":{"prompt_tokens":3,"completion_tokens":2,"total_tokens":5}}`,
			``,
			`data: [DONE]`,
			``,
		}
		for _, c := range chunks {
			_, _ = w.Write([]byte(c + "\n"))
			if flusher != nil {
				flusher.Flush()
			}
		}
	}))
	defer srv.Close()
	p := NewOpenAICompat("test", srv.URL, "k")
	p.ModelsList = []ModelInfo{{ID: "m"}}
	ch, cancel, err := p.Stream(context.Background(), ChatRequest{Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}}})
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
	assert.Equal(t, "Hello world", content.String())
	assert.Equal(t, FinishStop, finish)
}

func TestOpenAICompat_Stream_4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`forbidden`))
	}))
	defer srv.Close()
	p := NewOpenAICompat("test", srv.URL, "k")
	p.ModelsList = []ModelInfo{{ID: "m"}}
	_, _, err := p.Stream(context.Background(), ChatRequest{Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}}})
	assert.Error(t, err)
}

func TestOpenAICompat_DefaultModel(t *testing.T) {
	p := NewOpenAICompat("t", "http://x", "k")
	p.ModelsList = []ModelInfo{
		{ID: "a", InputCostPerMTok: 1, OutputCostPerMTok: 2},
		{ID: "b", InputCostPerMTok: 0.1, OutputCostPerMTok: 0.2},
	}
	assert.Equal(t, "b", p.DefaultModel("chat"))
}

func TestOpenAICompat_DefaultModel_VisionFilter(t *testing.T) {
	p := NewOpenAICompat("t", "http://x", "k")
	p.ModelsList = []ModelInfo{
		{ID: "a", InputCostPerMTok: 1, OutputCostPerMTok: 2, SupportsVision: false},
		{ID: "b", InputCostPerMTok: 5, OutputCostPerMTok: 5, SupportsVision: true},
	}
	assert.Equal(t, "b", p.DefaultModel("vision"))
}

func TestOpenAICompat_DefaultModel_ToolFilter(t *testing.T) {
	p := NewOpenAICompat("t", "http://x", "k")
	p.ModelsList = []ModelInfo{
		{ID: "a", InputCostPerMTok: 1, OutputCostPerMTok: 2, SupportsTools: false},
		{ID: "b", InputCostPerMTok: 5, OutputCostPerMTok: 5, SupportsTools: true},
	}
	assert.Equal(t, "b", p.DefaultModel("tool"))
}

// -----------------------------------------------------------------------------
// Concrete providers: factory tests
// -----------------------------------------------------------------------------

func TestNewOpenAI(t *testing.T) {
	p := NewOpenAI("k", nil)
	assert.Equal(t, "openai", p.Name())
	assert.Equal(t, "https://api.openai.com/v1", p.BaseURL)
}

func TestNewOpenRouter(t *testing.T) {
	p := NewOpenRouter("k", nil)
	assert.Equal(t, "openrouter", p.Name())
	assert.Contains(t, p.ExtraHeaders, "HTTP-Referer")
}

func TestNewTogether(t *testing.T) {
	p := NewTogether("k", nil)
	assert.Equal(t, "together", p.Name())
}

func TestNewGroq(t *testing.T) {
	p := NewGroq("k", nil)
	assert.Equal(t, "groq", p.Name())
}

func TestNewFireworks(t *testing.T) {
	p := NewFireworks("k", nil)
	assert.Equal(t, "fireworks", p.Name())
}

func TestNewDeepSeek(t *testing.T) {
	p := NewDeepSeek("k", nil)
	assert.Equal(t, "deepseek", p.Name())
}

func TestNewXAI(t *testing.T) {
	p := NewXAI("k", nil)
	assert.Equal(t, "xai", p.Name())
}

func TestNewMistral(t *testing.T) {
	p := NewMistral("k", nil)
	assert.Equal(t, "mistral", p.Name())
}

func TestNewOllama(t *testing.T) {
	p := NewOllama("", nil)
	assert.Equal(t, "ollama", p.Name())
	assert.Equal(t, "http://localhost:11434/v1", p.BaseURL)
}

func TestNewOllama_CustomURL(t *testing.T) {
	p := NewOllama("http://myserver:9999/v1", nil)
	assert.Equal(t, "http://myserver:9999/v1", p.BaseURL)
}

func TestNewCustom(t *testing.T) {
	p := NewCustom("custom", "https://myllm.example/v1", "k", nil)
	assert.Equal(t, "custom", p.Name())
	assert.Equal(t, "https://myllm.example/v1", p.BaseURL)
}

// -----------------------------------------------------------------------------
// Anthropic
// -----------------------------------------------------------------------------

func TestAnthropic_Name_DefaultModel(t *testing.T) {
	p := NewAnthropic("k", []ModelInfo{
		{ID: "claude-3-opus-20240229"},
		{ID: "claude-3-5-sonnet-20241022"},
	})
	assert.Equal(t, "anthropic", p.Name())
	assert.Equal(t, "claude-3-5-sonnet-20241022", p.DefaultModel("chat"))
}

func TestAnthropic_Chat_OK(t *testing.T) {
	var gotAuth, gotVersion string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotAuth = r.Header.Get("x-api-key")
		gotVersion = r.Header.Get("anthropic-version")
		var body anthRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		assert.Equal(t, "claude-3-5-sonnet-20241022", body.Model)
		assert.Equal(t, "You are helpful", body.System)
		// System is hoisted out; the messages array contains only user turns.
		assert.Equal(t, 1, len(body.Messages))
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(anthResponse{
			ID:         "msg_1",
			Model:      "claude-3-5-sonnet-20241022",
			StopReason: "end_turn",
			Content: []struct {
				Type  string `json:"type"`
				Text  string `json:"text,omitempty"`
				ID    string `json:"id,omitempty"`
				Name  string `json:"name,omitempty"`
				Input any    `json:"input,omitempty"`
			}{{
				Type: "text",
				Text: "Hi there!",
			}},
			Usage: struct {
				InputTokens  int `json:"input_tokens"`
				OutputTokens int `json:"output_tokens"`
			}{InputTokens: 10, OutputTokens: 5},
		})
	}))
	defer srv.Close()
	p := NewAnthropic("sk-ant-x", nil)
	p.BaseURL = srv.URL
	resp, err := p.Chat(context.Background(), ChatRequest{
		Model: "claude-3-5-sonnet-20241022",
		Messages: []Message{
			{Role: RoleSystem, Content: "You are helpful"},
			{Role: RoleUser, Content: "hi"},
		},
	})
	require.NoError(t, err)
	assert.Equal(t, "Hi there!", resp.Message.Content)
	assert.Equal(t, FinishStop, resp.FinishReason)
	assert.Equal(t, "sk-ant-x", gotAuth)
	assert.Equal(t, "2023-06-01", gotVersion)
}

func TestAnthropic_Chat_ToolUse(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_ = json.NewEncoder(w).Encode(anthResponse{
			ID:         "msg_2",
			Model:      "claude-3-5-sonnet-20241022",
			StopReason: "tool_use",
			Content: []struct {
				Type  string `json:"type"`
				Text  string `json:"text,omitempty"`
				ID    string `json:"id,omitempty"`
				Name  string `json:"name,omitempty"`
				Input any    `json:"input,omitempty"`
			}{{
				Type: "tool_use",
				ID:   "toolu_1",
				Name: "get_weather",
				Input: map[string]any{
					"location": "Paris",
				},
			}},
		})
	}))
	defer srv.Close()
	p := NewAnthropic("k", nil)
	p.BaseURL = srv.URL
	resp, err := p.Chat(context.Background(), ChatRequest{
		Model:    "claude-3-5-sonnet-20241022",
		Messages: []Message{{Role: RoleUser, Content: "weather?"}},
	})
	require.NoError(t, err)
	assert.Equal(t, FinishToolCalls, resp.FinishReason)
	require.Len(t, resp.Message.ToolCalls, 1)
	assert.Equal(t, "get_weather", resp.Message.ToolCalls[0].Function.Name)
	assert.Contains(t, resp.Message.ToolCalls[0].Function.Arguments, "Paris")
}

func TestAnthropic_Chat_4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte(`{"error":{"type":"bad_request","message":"x"}}`))
	}))
	defer srv.Close()
	p := NewAnthropic("k", nil)
	p.BaseURL = srv.URL
	_, err := p.Chat(context.Background(), ChatRequest{
		Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}},
	})
	assert.Error(t, err)
}

func TestAnthropic_Chat_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()
	p := NewAnthropic("k", nil)
	p.BaseURL = srv.URL
	_, err := p.Chat(context.Background(), ChatRequest{
		Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}},
	})
	assert.Error(t, err)
}

func TestAnthropic_NoModel(t *testing.T) {
	p := NewAnthropic("k", nil)
	_, err := p.Chat(context.Background(), ChatRequest{Messages: []Message{{Role: RoleUser, Content: "x"}}})
	assert.ErrorIs(t, err, ErrNoModel)
}

func TestAnthropic_NoMessages(t *testing.T) {
	p := NewAnthropic("k", nil)
	_, err := p.Chat(context.Background(), ChatRequest{Model: "m"})
	assert.ErrorIs(t, err, ErrNoMessages)
}

func TestAnthropic_Stream_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/event-stream")
		w.WriteHeader(200)
		flusher, _ := w.(http.Flusher)
		lines := []string{
			"event: message_start",
			`data: {"type":"message_start","message":{"usage":{"input_tokens":3,"output_tokens":0}}}`,
			"",
			"event: content_block_start",
			`data: {"type":"content_block_start"}`,
			"",
			"event: content_block_delta",
			`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"Hel"}}`,
			"",
			"event: content_block_delta",
			`data: {"type":"content_block_delta","delta":{"type":"text_delta","text":"lo"}}`,
			"",
			"event: message_stop",
			`data: {"type":"message_stop"}`,
			"",
		}
		for _, l := range lines {
			_, _ = w.Write([]byte(l + "\n"))
			if flusher != nil {
				flusher.Flush()
			}
		}
	}))
	defer srv.Close()
	p := NewAnthropic("k", nil)
	p.BaseURL = srv.URL
	ch, cancel, err := p.Stream(context.Background(), ChatRequest{
		Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}},
	})
	require.NoError(t, err)
	defer cancel()
	var content strings.Builder
	for ev := range ch {
		if ev.Err != nil {
			t.Fatal(ev.Err)
		}
		content.WriteString(ev.Delta.Content)
	}
	assert.Equal(t, "Hello", content.String())
}

// -----------------------------------------------------------------------------
// Google
// -----------------------------------------------------------------------------

func TestGoogle_Name_DefaultModel(t *testing.T) {
	p := NewGoogle("k", []ModelInfo{
		{ID: "gemini-1.5-pro"},
		{ID: "gemini-1.5-flash"},
	})
	assert.Equal(t, "google", p.Name())
	assert.Equal(t, "gemini-1.5-flash", p.DefaultModel("chat"))
}

func TestGoogle_Chat_OK(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.RawQuery, "key=k")
		assert.Contains(t, r.URL.Path, "/models/gemini-1.5-flash:generateContent")
		var body gemRequest
		require.NoError(t, json.NewDecoder(r.Body).Decode(&body))
		_ = json.NewEncoder(w).Encode(gemResponse{
			Candidates: []struct {
				Content      gemContent `json:"content"`
				FinishReason string     `json:"finishReason"`
				Index        int        `json:"index"`
			}{{
				Content: gemContent{
					Role:  "model",
					Parts: []gemPart{{Text: "Gemini says hi"}},
				},
				FinishReason: "STOP",
			}},
			UsageMetadata: struct {
				PromptTokenCount     int `json:"promptTokenCount"`
				CandidatesTokenCount int `json:"candidatesTokenCount"`
				TotalTokenCount      int `json:"totalTokenCount"`
			}{PromptTokenCount: 4, CandidatesTokenCount: 3, TotalTokenCount: 7},
		})
	}))
	defer srv.Close()
	p := NewGoogle("k", nil)
	p.BaseURL = srv.URL
	resp, err := p.Chat(context.Background(), ChatRequest{
		Model:    "gemini-1.5-flash",
		Messages: []Message{{Role: RoleUser, Content: "hi"}},
	})
	require.NoError(t, err)
	assert.Equal(t, "Gemini says hi", resp.Message.Content)
	assert.Equal(t, FinishStop, resp.FinishReason)
}

func TestGoogle_Chat_4xx(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`{"error":{"code":403,"message":"forbidden","status":"PERMISSION_DENIED"}}`))
	}))
	defer srv.Close()
	p := NewGoogle("k", nil)
	p.BaseURL = srv.URL
	_, err := p.Chat(context.Background(), ChatRequest{
		Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}},
	})
	assert.Error(t, err)
}

func TestGoogle_Chat_BadJSON(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`not json`))
	}))
	defer srv.Close()
	p := NewGoogle("k", nil)
	p.BaseURL = srv.URL
	_, err := p.Chat(context.Background(), ChatRequest{
		Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}},
	})
	assert.Error(t, err)
}

func TestGoogle_Chat_NoCandidates(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"candidates":[]}`))
	}))
	defer srv.Close()
	p := NewGoogle("k", nil)
	p.BaseURL = srv.URL
	_, err := p.Chat(context.Background(), ChatRequest{
		Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}},
	})
	assert.Error(t, err)
}

func TestGoogle_NoModel(t *testing.T) {
	p := NewGoogle("k", nil)
	_, err := p.Chat(context.Background(), ChatRequest{Messages: []Message{{Role: RoleUser, Content: "x"}}})
	assert.ErrorIs(t, err, ErrNoModel)
}

func TestGoogle_NoMessages(t *testing.T) {
	p := NewGoogle("k", nil)
	_, err := p.Chat(context.Background(), ChatRequest{Model: "m"})
	assert.ErrorIs(t, err, ErrNoMessages)
}

// -----------------------------------------------------------------------------
// Registry
// -----------------------------------------------------------------------------

func TestRegistry_RegisterGet(t *testing.T) {
	r := NewRegistry()
	p := NewOpenAI("k", nil)
	r.Register(p)
	got, ok := r.Get("openai")
	require.True(t, ok)
	assert.Equal(t, p, got)
}

func TestRegistry_GetMissing(t *testing.T) {
	r := NewRegistry()
	_, ok := r.Get("nope")
	assert.False(t, ok)
}

func TestRegistry_MustGet(t *testing.T) {
	r := NewRegistry()
	r.Register(NewOpenAI("k", nil))
	assert.NotPanics(t, func() { r.MustGet("openai") })
	assert.Panics(t, func() { r.MustGet("nope") })
}

func TestRegistry_Delete(t *testing.T) {
	r := NewRegistry()
	r.Register(NewOpenAI("k", nil))
	assert.True(t, r.Delete("openai"))
	assert.False(t, r.Delete("openai"))
}

func TestRegistry_Names(t *testing.T) {
	r := NewRegistry()
	r.Register(NewOpenAI("k", nil))
	r.Register(NewAnthropic("k", nil))
	r.Register(NewGoogle("k", nil))
	assert.Equal(t, []string{"anthropic", "google", "openai"}, r.Names())
}

func TestRegistry_List(t *testing.T) {
	r := NewRegistry()
	r.Register(NewOpenAI("k", nil))
	r.Register(NewGoogle("k", nil))
	ps := r.List()
	assert.Len(t, ps, 2)
}

func TestRegistry_Len(t *testing.T) {
	r := NewRegistry()
	assert.Equal(t, 0, r.Len())
	r.Register(NewOpenAI("k", nil))
	assert.Equal(t, 1, r.Len())
}

func TestRegistry_Chat_NotFound(t *testing.T) {
	r := NewRegistry()
	_, err := r.Chat(context.Background(), "nope", ChatRequest{Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}}})
	assert.ErrorIs(t, err, ErrNoProvider)
}

func TestRegistry_Stream_NotFound(t *testing.T) {
	r := NewRegistry()
	_, _, err := r.Stream(context.Background(), "nope", ChatRequest{Model: "m", Messages: []Message{{Role: RoleUser, Content: "x"}}})
	assert.ErrorIs(t, err, ErrNoProvider)
}

func TestRegistry_AllModels(t *testing.T) {
	r := NewRegistry()
	r.Register(NewOpenAI("k", []ModelInfo{{ID: "gpt-4o"}}))
	r.Register(NewAnthropic("k", []ModelInfo{{ID: "claude-3-5-sonnet-20241022"}}))
	all := r.AllModels()
	assert.Contains(t, all, "openai")
	assert.Contains(t, all, "anthropic")
	assert.Equal(t, "gpt-4o", all["openai"][0].ID)
}

func TestRegistry_RegisterReplace(t *testing.T) {
	r := NewRegistry()
	r.Register(NewOpenAI("k1", nil))
	r.Register(NewOpenAI("k2", nil))
	assert.Equal(t, 1, r.Len())
}

// Sanity: atomic counter for streaming test.
var _ atomic.Int32
