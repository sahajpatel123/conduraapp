package daemon

import (
	"context"
	"log/slog"

	"github.com/sahajpatel123/synapticapp/internal/api_key"
	"github.com/sahajpatel123/synapticapp/internal/config"
	"github.com/sahajpatel123/synapticapp/internal/llm"
)

// buildProvidersFromConfig reads cfg.LLM.Providers and, for every
// enabled entry, fetches the stored API key and registers a
// provider with the registry. Returns the number registered.
func buildProvidersFromConfig(log *slog.Logger, registry *llm.Registry, cfg *config.Config, akm *api_key.Manager) int {
	count := 0
	for name, p := range cfg.LLM.Providers {
		if !p.Enabled {
			continue
		}
		models := modelsForProvider(name)
		keys, err := akm.ListByProvider(context.Background(), name)
		if err != nil {
			log.Warn("list keys failed", "provider", name, "err", err)
			continue
		}
		var key string
		if len(keys) > 0 {
			key = keys[0].Secret
		}
		// Ollama is a keyless local provider — register it
		// without requiring an API key.
		if len(keys) == 0 && name != config.ProviderOllama {
			log.Debug("no api key for provider, skipping", "provider", name)
			continue
		}
		prov := buildProvider(name, key, p.BaseURL, models)
		if prov == nil {
			log.Warn("unknown provider in config", "provider", name)
			continue
		}
		registry.Register(prov)
		count++
	}
	return count
}

// buildProvider returns a registered llm.Provider for the given name.
// Returns nil if the name is unknown. baseURL is an optional override
// (e.g. for local proxies or custom endpoints).
func buildProvider(name, key, baseURL string, models []llm.ModelInfo) llm.Provider {
	switch name {
	case config.ProviderAnthropic:
		return llm.NewAnthropic(key, models)
	case config.ProviderOpenAI:
		return llm.NewOpenAI(key, models)
	case config.ProviderGoogle:
		return llm.NewGoogle(key, models)
	case config.ProviderXAI:
		return llm.NewOpenAICompat(config.ProviderXAI, pickURL(baseURL, "https://api.x.ai/v1"), key)
	case config.ProviderMistral:
		return llm.NewOpenAICompat(config.ProviderMistral, pickURL(baseURL, "https://api.mistral.ai/v1"), key)
	case config.ProviderDeepSeek:
		return llm.NewOpenAICompat(config.ProviderDeepSeek, pickURL(baseURL, "https://api.deepseek.com/v1"), key)
	case config.ProviderOpenRouter:
		return llm.NewOpenAICompat(config.ProviderOpenRouter, pickURL(baseURL, "https://openrouter.ai/api/v1"), key)
	case config.ProviderGroq:
		return llm.NewOpenAICompat(config.ProviderGroq, pickURL(baseURL, "https://api.groq.com/openai/v1"), key)
	case config.ProviderTogether:
		return llm.NewOpenAICompat(config.ProviderTogether, pickURL(baseURL, "https://api.together.xyz/v1"), key)
	case config.ProviderFireworks:
		return llm.NewOpenAICompat(config.ProviderFireworks, pickURL(baseURL, "https://api.fireworks.ai/inference/v1"), key)
	case config.ProviderOllama:
		// Ollama uses no key; the API key field is ignored.
		return llm.NewOpenAICompat(config.ProviderOllama, pickURL(baseURL, "http://127.0.0.1:11434/v1"), "")
	case config.ProviderLocalAI:
		// LocalAI is keyless by default; users can configure an API key
		// at the LocalAI server if they want one.
		return llm.NewOpenAICompat(config.ProviderLocalAI, pickURL(baseURL, "http://127.0.0.1:8080/v1"), "")
	case config.ProviderLMStudio:
		// LM Studio is keyless; users configure auth in the LM Studio UI.
		return llm.NewOpenAICompat(config.ProviderLMStudio, pickURL(baseURL, "http://127.0.0.1:1234/v1"), "")
	case config.ProviderVLLM:
		// vLLM is keyless by default; pass --api-key on the server to require one.
		return llm.NewOpenAICompat(config.ProviderVLLM, pickURL(baseURL, "http://127.0.0.1:8000/v1"), "")
	}
	return nil
}

// pickURL returns baseURL if non-empty, otherwise fallback. Used to
// allow per-provider BaseURL overrides in the config while keeping
// sane defaults.
func pickURL(baseURL, fallback string) string {
	if baseURL != "" {
		return baseURL
	}
	return fallback
}

// allModels is the catalog of well-known models per provider.
// The full model list (with prices) lives in internal/llm/model_pricing.go;
// this is a smaller, opinionated subset used when the user has not
// configured custom models.
//
// Marketing-aligned defaults come first; legacy IDs are kept for users
// upgrading from earlier builds.
var allModels = []struct {
	provider string
	id       string
}{
	// Anthropic — current gen (marketing defaults)
	{config.ProviderAnthropic, "claude-opus-4-7"},
	{config.ProviderAnthropic, "claude-sonnet-4-5"},
	{config.ProviderAnthropic, "claude-haiku-4-5"},
	// Anthropic — legacy
	{config.ProviderAnthropic, "claude-3-5-sonnet-20241022"},
	{config.ProviderAnthropic, "claude-3-5-haiku-20241022"},
	{config.ProviderAnthropic, "claude-3-opus-20240229"},

	// OpenAI — current gen (marketing defaults)
	{config.ProviderOpenAI, "gpt-5.5"},
	{config.ProviderOpenAI, "o3"},
	{config.ProviderOpenAI, "o4-mini"},
	// OpenAI — legacy
	{config.ProviderOpenAI, "gpt-4o"},
	{config.ProviderOpenAI, "gpt-4o-mini"},
	{config.ProviderOpenAI, "o1-mini"},

	// Google — current gen (marketing defaults)
	{config.ProviderGoogle, "gemini-3.5-flash"},
	{config.ProviderGoogle, "gemini-3.1-pro"},
	// Google — legacy
	{config.ProviderGoogle, "gemini-1.5-pro"},
	{config.ProviderGoogle, "gemini-1.5-flash"},

	// xAI — current gen (marketing defaults)
	{config.ProviderXAI, "grok-4.3"},
	{config.ProviderXAI, "grok-4.3-fast"},
	// xAI — legacy
	{config.ProviderXAI, "grok-2"},

	// Mistral — current gen (marketing defaults)
	{config.ProviderMistral, "mistral-large-3"},
	{config.ProviderMistral, "codestral-latest"},
	// Mistral — legacy
	{config.ProviderMistral, "mistral-large-latest"},

	// DeepSeek — current gen (marketing defaults)
	{config.ProviderDeepSeek, "deepseek-v4"},
	{config.ProviderDeepSeek, "deepseek-r1"},
	// DeepSeek — legacy
	{config.ProviderDeepSeek, "deepseek-chat"},

	// OpenRouter (300+ models available via API)
	{config.ProviderOpenRouter, "openrouter/auto"},

	// Groq — current gen (marketing defaults; Whisper is STT-only)
	{config.ProviderGroq, "llama-4-70b-versatile"},
	{config.ProviderGroq, "llama-4-8b-instant"},
	// Groq — legacy
	{config.ProviderGroq, "llama-3.3-70b-versatile"},

	// Together — current gen
	{config.ProviderTogether, "meta-llama/Llama-4-70B-Instruct"},
	{config.ProviderTogether, "Qwen/Qwen3-72B-Instruct"},
	{config.ProviderTogether, "mistralai/Mixtral-8x22B-Instruct-v0.1"},

	// Fireworks — current gen
	{config.ProviderFireworks, "accounts/fireworks/models/llama-4-70b-instruct"},
	{config.ProviderFireworks, "accounts/fireworks/models/qwen3-72b-instruct"},
	{config.ProviderFireworks, "accounts/fireworks/models/deepseek-v4-instruct"},

	// Ollama / LocalAI / LM Studio / vLLM (local; user picks actual model)
	{config.ProviderOllama, "llama3.2"},
	{config.ProviderLocalAI, "llama-3.1-8b-instruct"},
	{config.ProviderLMStudio, "qwen2.5-7b-instruct"},
	{config.ProviderVLLM, "meta-llama/Llama-3.1-8B-Instruct"},
}

// modelsForProvider returns the well-known model IDs for a provider.
func modelsForProvider(name string) []llm.ModelInfo {
	var out []llm.ModelInfo
	for _, m := range allModels {
		if m.provider == name {
			out = append(out, llm.ModelInfo{ID: m.id})
		}
	}
	return out
}
