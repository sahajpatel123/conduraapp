package daemon

import (
	"context"
	"log/slog"
	"net/http"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/anomaly"
	"github.com/sahajpatel123/conduraapp/internal/api_key"
	"github.com/sahajpatel123/conduraapp/internal/config"
	"github.com/sahajpatel123/conduraapp/internal/halt"
	"github.com/sahajpatel123/conduraapp/internal/llm"
)

// buildProvidersFromConfig reads cfg.LLM.Providers and, for every
// enabled entry, fetches the stored API key and registers a
// provider with the registry. Returns the number registered.
//
// If netGuard is non-nil, every provider's HTTP transport is wrapped
// by the guard so the kill switch's Layer 3 (network isolation) takes
// effect for outbound LLM traffic.
//
// If anomalyDet is non-nil, the same HTTP transports are wrapped
// with anomaly.RecordingTransport so the fifth §5.6 trigger
// (new-endpoint detection) fires on every outbound request. The
// recorder sits OUTSIDE the guard so a request blocked by the guard
// does not consume a "seen host" entry — only requests that
// actually reach the network are recorded.
//
// Phase 17, Fix #4 (B1): we ALSO auto-enable any provider that has a
// stored API key in the api_key.Manager but is disabled in the YAML
// map. This makes `apikeys.set` self-sufficient — the user adds a key
// via the GUI, and the provider becomes routable without requiring
// them to also edit config.yaml. cfg.LLM.Providers is treated as a
// defaults source, not a hard allowlist. The canonical provider name
// list is iterated so we pick up keys for any provider we know how
// to build, regardless of whether the user explicitly added it to
// the YAML map.
func buildProvidersFromConfig(log *slog.Logger, registry *llm.Registry, cfg *config.Config, akm *api_key.Manager, netGuard halt.NetworkGuard, anomalyDet *anomaly.Detector) int {
	if cfg.LLM.Providers == nil {
		cfg.LLM.Providers = map[string]config.ProviderConfig{}
	}
	for _, name := range knownProviders() {
		keys, err := akm.ListByProvider(context.Background(), name)
		if err != nil || len(keys) == 0 {
			continue
		}
		entry, ok := cfg.LLM.Providers[name]
		if !ok || !entry.Enabled {
			entry.Enabled = true
			cfg.LLM.Providers[name] = entry
			log.Info("auto-enabled provider from stored api key", "provider", name)
		}
	}
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
		// Phase 14I: wrap the provider's HTTP transport with the
		// network guard so Layer 3 (kill switch network isolation)
		// applies to all outbound LLM traffic. We do this through
		// a small adapter that calls into the LLM provider's
		// settable HTTPClient field.
		//
		// 2026-06-29 audit (P0-2): also wrap with the anomaly
		// recorder so the fifth §5.6 trigger (new-endpoint detection)
		// fires when the agent pivots to a host it has not used
		// before. Recorder wraps OUTSIDE the guard so a guard-blocked
		// request is not counted as "seen host".
		wrapProviderHTTPClient(prov, netGuard, anomalyDet)
		registry.Register(prov)
		count++
	}
	return count
}

// knownProviders returns the canonical list of provider names this
// daemon can register. Mirrors the cases in buildProvider(). Used
// by buildProvidersFromConfig to discover which providers have
// stored API keys without requiring the user to enumerate them in
// config.yaml first.
func knownProviders() []string {
	return []string{
		config.ProviderAnthropic,
		config.ProviderOpenAI,
		config.ProviderGoogle,
		config.ProviderXAI,
		config.ProviderMistral,
		config.ProviderDeepSeek,
		config.ProviderOpenRouter,
		config.ProviderGroq,
		config.ProviderTogether,
		config.ProviderFireworks,
		config.ProviderOllama,
		config.ProviderLocalAI,
		config.ProviderLMStudio,
		config.ProviderVLLM,
	}
}

// wrapProviderHTTPClient attaches the net guard's transport to the
// provider's HTTP client, if the provider exposes one. The OpenAI-
// compat providers (OpenAI, xAI, mistral, deepseek, openrouter,
// groq, together, fireworks, ollama, localai, lmstudio, vllm)
// all share the OpenAICompat struct which has an exported
// HTTPClient field. The Anthropic and Google native providers have
// their own struct; for those, the guard's policy is enforced
// through the in-process guard's WrapTransport applied at the
// http.Client level when those clients are built.
//
// The LLM clients in internal/llm read p.HTTPClient (or
// p.HTTPClient field for OpenAICompat) at request time, so wrapping
// the field takes effect on the next request without rebuilding
// the provider.
func wrapProviderHTTPClient(prov llm.Provider, guard halt.NetworkGuard, anomalyDet *anomaly.Detector) {
	if prov == nil {
		return
	}
	// OpenAICompat and friends all expose a settable *http.Client
	// via a method. We use a small interface to discover it.
	type clientBearer interface {
		GetHTTPClient() *http.Client
	}
	b, ok := prov.(clientBearer)
	if !ok {
		return
	}
	hc := b.GetHTTPClient()
	if hc == nil {
		hc = &http.Client{Timeout: 5 * time.Minute}
	}
	// Compose the transport chain. The recorder sits OUTSIDE the
	// guard so a guard-rejected request does not record the host
	// as "seen". The recorder also sits outside the existing
	// transport so the recorder captures the actual outbound URL
	// without the guard rewriting it.
	var transport http.RoundTripper
	transport = hc.Transport
	if guard != nil {
		transport = guard.WrapTransport(transport)
	}
	if anomalyDet != nil {
		transport = anomaly.NewRecordingTransport(anomalyDet, transport)
	}
	hc.Transport = transport
}

// wrapProvidersWithRecorder walks every registered provider in
// `reg` and re-wraps its HTTP client transport with an
// anomaly.RecordingTransport. It is called AFTER the safety
// subsystem is constructed, because the recorder needs the
// anomaly detector (which is part of the safety layer).
//
// Idempotent: a transport that already has a *RecordingTransport
// is left as-is. This avoids double-recording if the daemon is
// restarted within a test process.
func wrapProvidersWithRecorder(reg *llm.Registry, det *anomaly.Detector) {
	if reg == nil || det == nil {
		return
	}
	type clientBearer interface {
		GetHTTPClient() *http.Client
	}
	for _, prov := range reg.List() {
		b, ok := prov.(clientBearer)
		if !ok {
			continue
		}
		hc := b.GetHTTPClient()
		if hc == nil {
			continue
		}
		// Already wrapped? Skip.
		if _, ok := hc.Transport.(*anomaly.RecordingTransport); ok {
			continue
		}
		hc.Transport = anomaly.NewRecordingTransport(det, hc.Transport)
	}
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
