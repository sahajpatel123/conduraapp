package llm

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// -----------------------------------------------------------------------------
// LocalAI, LM Studio, vLLM — coverage for the new local provider constructors.
// All three are keyless OpenAI-compatible servers.
// -----------------------------------------------------------------------------

func TestNewLocalAI_DefaultURL(t *testing.T) {
	p := NewLocalAI("", nil)
	assert.Equal(t, "localai", p.Name())
	assert.Equal(t, "http://localhost:8080/v1", p.BaseURL)
	assert.NotNil(t, p.HTTPClient)
}

func TestNewLocalAI_CustomURL(t *testing.T) {
	p := NewLocalAI("http://mybox:9000/v1", nil)
	assert.Equal(t, "http://mybox:9000/v1", p.BaseURL)
}

func TestNewLocalAI_Models(t *testing.T) {
	p := NewLocalAI("", []ModelInfo{{ID: "a"}, {ID: "b"}})
	assert.Len(t, p.Models(), 2)
}

func TestNewLMStudio_DefaultURL(t *testing.T) {
	p := NewLMStudio("", nil)
	assert.Equal(t, "lmstudio", p.Name())
	assert.Equal(t, "http://localhost:1234/v1", p.BaseURL)
}

func TestNewLMStudio_CustomURL(t *testing.T) {
	p := NewLMStudio("http://gpu-rig:1234/v1", nil)
	assert.Equal(t, "http://gpu-rig:1234/v1", p.BaseURL)
}

func TestNewVLLM_DefaultURL(t *testing.T) {
	p := NewVLLM("", nil)
	assert.Equal(t, "vllm", p.Name())
	assert.Equal(t, "http://localhost:8000/v1", p.BaseURL)
}

func TestNewVLLM_CustomURL(t *testing.T) {
	p := NewVLLM("http://server:9999/v1", nil)
	assert.Equal(t, "http://server:9999/v1", p.BaseURL)
}

// -----------------------------------------------------------------------------
// Marketing-aligned model registry.
//
// Every model name listed on the website (/ecosystem page) must be
// registered in the pricing catalog so the cost estimator and the
// model-suggestion UI can reference it. The IDs follow the providers'
// conventional naming patterns; if a provider rejects an ID at runtime,
// the upstream API will return an error and the failover layer routes
// around it.
// -----------------------------------------------------------------------------

func TestMarketingModels_AllRegistered(t *testing.T) {
	// Each tuple is (provider, model ID) as it appears on the marketing site.
	required := []struct {
		provider string
		modelID  string
	}{
		// Anthropic
		{"anthropic", "claude-opus-4-7"},
		{"anthropic", "claude-sonnet-4-5"},
		{"anthropic", "claude-haiku-4-5"},
		// OpenAI
		{"openai", "gpt-5.5"},
		{"openai", "o3"},
		{"openai", "o4-mini"},
		// Google
		{"google", "gemini-3.5-flash"},
		{"google", "gemini-3.1-pro"},
		// xAI
		{"xai", "grok-4.3"},
		{"xai", "grok-4.3-fast"},
		// Mistral
		{"mistral", "mistral-large-3"},
		{"mistral", "codestral-latest"},
		// DeepSeek
		{"deepseek", "deepseek-v4"},
		{"deepseek", "deepseek-r1"},
		// OpenRouter
		{"openrouter", "openrouter/auto"},
		// Groq (Whisper is STT-only, handled by voice subsystem)
		{"groq", "llama-4-70b-versatile"},
		{"groq", "llama-4-8b-instant"},
		// Together
		{"together", "meta-llama/Llama-4-70B-Instruct"},
		{"together", "Qwen/Qwen3-72B-Instruct"},
		{"together", "mistralai/Mixtral-8x22B-Instruct-v0.1"},
		// Fireworks
		{"fireworks", "accounts/fireworks/models/llama-4-70b-instruct"},
		{"fireworks", "accounts/fireworks/models/qwen3-72b-instruct"},
		{"fireworks", "accounts/fireworks/models/deepseek-v4-instruct"},
		// Local / Ollama / LocalAI / LM Studio / vLLM
		{"ollama", "llama3.2"},
		{"localai", "llama-3.1-8b-instruct"},
		{"lmstudio", "qwen2.5-7b-instruct"},
		{"vllm", "meta-llama/Llama-3.1-8B-Instruct"},
	}
	for _, m := range required {
		_, ok := LookupModel(m.modelID)
		assert.Truef(t, ok, "marketing-listed model not in pricing registry: provider=%s modelID=%s", m.provider, m.modelID)
	}
}

func TestLegacyModels_StillRegistered(t *testing.T) {
	// Backward-compatibility: users upgrading from earlier builds must still
	// find their installed models in the registry.
	legacy := []string{
		"gpt-4o", "gpt-4o-mini", "o1", "o1-mini", "o3-mini", "gpt-4-turbo",
		"claude-3-5-sonnet-20241022", "claude-3-5-haiku-20241022", "claude-3-opus-20240229",
		"gemini-1.5-pro", "gemini-1.5-flash", "gemini-2.0-flash-exp",
		"grok-2", "grok-2-mini",
		"mistral-large-latest", "mistral-small-latest",
		"deepseek-chat", "deepseek-reasoner",
		"llama-3.3-70b-versatile", "llama-3.1-8b-instant", "mixtral-8x7b-32768",
		"meta-llama/Llama-3.3-70B-Instruct-Turbo", "Qwen/Qwen2.5-72B-Instruct-Turbo",
		"accounts/fireworks/models/llama-v3p3-70b-instruct",
		"qwen2.5",
	}
	for _, id := range legacy {
		_, ok := LookupModel(id)
		assert.Truef(t, ok, "legacy model missing from pricing registry: %s", id)
	}
}

func TestEstimateCost_MarketingModels(t *testing.T) {
	// Spot-check that EstimateCost returns a non-negative number for every
	// marketing-listed model. Pricing may be 0.0 (unknown) for new releases;
	// the failover layer handles zero-cost estimation gracefully.
	ids := []string{
		"claude-opus-4-7", "claude-sonnet-4-5", "claude-haiku-4-5",
		"gpt-5.5", "o3", "o4-mini",
		"gemini-3.5-flash", "gemini-3.1-pro",
		"grok-4.3", "grok-4.3-fast",
		"mistral-large-3", "codestral-latest",
		"deepseek-v4", "deepseek-r1",
		"openrouter/auto",
		"llama-4-70b-versatile", "llama-4-8b-instant",
		"meta-llama/Llama-4-70B-Instruct",
		"accounts/fireworks/models/llama-4-70b-instruct",
	}
	for _, id := range ids {
		cost := EstimateCost(id, Usage{InputTokens: 1000, OutputTokens: 500})
		assert.GreaterOrEqualf(t, cost, 0.0, "estimate must be non-negative: %s", id)
	}
}
