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
		if len(keys) == 0 {
			log.Debug("no api key for provider, skipping", "provider", name)
			continue
		}
		// Use the first key (we currently support one key per
		// provider per label; future versions can support multiple).
		key := keys[0].Secret
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
var allModels = []struct {
	provider string
	id       string
}{
	{config.ProviderAnthropic, "claude-3-5-sonnet-20241022"},
	{config.ProviderAnthropic, "claude-3-5-haiku-20241022"},
	{config.ProviderAnthropic, "claude-3-opus-20240229"},
	{config.ProviderOpenAI, "gpt-4o"},
	{config.ProviderOpenAI, "gpt-4o-mini"},
	{config.ProviderOpenAI, "o1-preview"},
	{config.ProviderOpenAI, "o1-mini"},
	{config.ProviderGoogle, "gemini-1.5-pro"},
	{config.ProviderGoogle, "gemini-1.5-flash"},
	{config.ProviderXAI, "grok-2"},
	{config.ProviderMistral, "mistral-large-latest"},
	{config.ProviderDeepSeek, "deepseek-chat"},
	{config.ProviderOpenRouter, "openrouter/auto"},
	{config.ProviderGroq, "llama-3.3-70b-versatile"},
	{config.ProviderTogether, "meta-llama/Llama-3.3-70B-Instruct-Turbo"},
	{config.ProviderFireworks, "accounts/fireworks/models/llama-v3p3-70b-instruct"},
	{config.ProviderOllama, "llama3.2"},
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
