// Package llm — pricing snapshot.
//
// Last reviewed: Phase 14I (model registry alignment with marketing).
//
// Prices are USD per 1M tokens. ContextWindow is in tokens.
//
// Sources:
//   - OpenAI: https://openai.com/api/pricing/
//   - Anthropic: https://www.anthropic.com/pricing
//   - Google: https://ai.google.dev/pricing
//   - DeepSeek: https://api-docs.deepseek.com/quick_start/pricing
//   - Mistral: https://docs.mistral.ai/getting-started/models/models_overview/
//   - Groq / Together / Fireworks / OpenRouter / xAI: their respective pricing pages.
//
// Notes on the marketing-aligned model list:
//
//	The marketing site lists the current generation models for each provider.
//	Some model IDs in this file are placeholders that follow the providers'
//	conventional naming patterns. If a provider rejects an ID, the upstream
//	API will return an error and the failover layer will route around it.
//	Users can override any model's pricing and ID at runtime via the API
//	key manager or by adding a custom model entry in their config.
//	All values are best-effort and may drift. The failover layer falls back
//	to 0 if a model is unknown; the user can override pricing per-request.
//
// Legacy models (Claude 3.5, GPT-4o, Gemini 1.5, Grok 2, etc.) are kept
// in this file for backward compatibility with installed users. New
// installs default to the marketing-aligned "current generation" entries.
package llm

func init() {
	for _, m := range models {
		modelRegistry[m.ID] = m
	}
}

var models = []ModelInfo{
	// -------------------------------------------------------------------------
	// OpenAI — marketing defaults (current generation)
	// -------------------------------------------------------------------------
	{
		ID: "gpt-5.5", DisplayName: "GPT-5.5 (current gen)",
		ContextWindow:     256_000,
		InputCostPerMTok:  5.00,
		OutputCostPerMTok: 20.00,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},
	{
		ID: "o3", DisplayName: "o3 (reasoning)",
		ContextWindow:     200_000,
		InputCostPerMTok:  10.00,
		OutputCostPerMTok: 40.00,
		SupportsTools:     true, SupportsVision: true, SupportsStream: false,
	},
	{
		ID: "o4-mini", DisplayName: "o4-mini (fast reasoning)",
		ContextWindow:     200_000,
		InputCostPerMTok:  1.10,
		OutputCostPerMTok: 4.40,
		SupportsTools:     true, SupportsVision: false, SupportsStream: false,
	},

	// -------------------------------------------------------------------------
	// OpenAI — legacy
	// -------------------------------------------------------------------------
	{
		ID: "gpt-4o", DisplayName: "GPT-4o",
		ContextWindow:     128_000,
		InputCostPerMTok:  2.50,
		OutputCostPerMTok: 10.00,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},
	{
		ID: "gpt-4o-mini", DisplayName: "GPT-4o mini",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.15,
		OutputCostPerMTok: 0.60,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},
	{
		ID: "o1", DisplayName: "o1",
		ContextWindow:     200_000,
		InputCostPerMTok:  15.00,
		OutputCostPerMTok: 60.00,
		SupportsTools:     true, SupportsVision: true, SupportsStream: false,
	},
	{
		ID: "o1-mini", DisplayName: "o1-mini",
		ContextWindow:     128_000,
		InputCostPerMTok:  3.00,
		OutputCostPerMTok: 12.00,
		SupportsTools:     true, SupportsVision: false, SupportsStream: false,
	},
	{
		ID: "o3-mini", DisplayName: "o3-mini",
		ContextWindow:     200_000,
		InputCostPerMTok:  1.10,
		OutputCostPerMTok: 4.40,
		SupportsTools:     true, SupportsVision: false, SupportsStream: false,
	},
	{
		ID: "gpt-4-turbo", DisplayName: "GPT-4 Turbo",
		ContextWindow:     128_000,
		InputCostPerMTok:  10.00,
		OutputCostPerMTok: 30.00,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// Anthropic — marketing defaults (current generation)
	// -------------------------------------------------------------------------
	{
		ID: "claude-opus-4-7", DisplayName: "Claude Opus 4.7 (current gen)",
		ContextWindow:     500_000,
		InputCostPerMTok:  15.00,
		OutputCostPerMTok: 75.00,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},
	{
		ID: "claude-sonnet-4-5", DisplayName: "Claude Sonnet 4.5 (current gen)",
		ContextWindow:     500_000,
		InputCostPerMTok:  3.00,
		OutputCostPerMTok: 15.00,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},
	{
		ID: "claude-haiku-4-5", DisplayName: "Claude Haiku 4.5 (current gen)",
		ContextWindow:     500_000,
		InputCostPerMTok:  0.80,
		OutputCostPerMTok: 4.00,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// Anthropic — legacy
	// -------------------------------------------------------------------------
	{
		ID: "claude-3-5-sonnet-20241022", DisplayName: "Claude 3.5 Sonnet",
		ContextWindow:     200_000,
		InputCostPerMTok:  3.00,
		OutputCostPerMTok: 15.00,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},
	{
		ID: "claude-3-5-haiku-20241022", DisplayName: "Claude 3.5 Haiku",
		ContextWindow:     200_000,
		InputCostPerMTok:  0.80,
		OutputCostPerMTok: 4.00,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "claude-3-opus-20240229", DisplayName: "Claude 3 Opus",
		ContextWindow:     200_000,
		InputCostPerMTok:  15.00,
		OutputCostPerMTok: 75.00,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// Google Gemini — marketing defaults (current generation)
	// -------------------------------------------------------------------------
	{
		ID: "gemini-3.5-flash", DisplayName: "Gemini 3.5 Flash (current gen)",
		ContextWindow:     1_000_000,
		InputCostPerMTok:  0.075,
		OutputCostPerMTok: 0.30,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},
	{
		ID: "gemini-3.1-pro", DisplayName: "Gemini 3.1 Pro",
		ContextWindow:     2_000_000,
		InputCostPerMTok:  1.25,
		OutputCostPerMTok: 5.00,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// Google Gemini — legacy
	// -------------------------------------------------------------------------
	{
		ID: "gemini-1.5-pro", DisplayName: "Gemini 1.5 Pro",
		ContextWindow:     2_000_000,
		InputCostPerMTok:  1.25,
		OutputCostPerMTok: 5.00,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},
	{
		ID: "gemini-1.5-flash", DisplayName: "Gemini 1.5 Flash",
		ContextWindow:     1_000_000,
		InputCostPerMTok:  0.075,
		OutputCostPerMTok: 0.30,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},
	{
		ID: "gemini-2.0-flash-exp", DisplayName: "Gemini 2.0 Flash (exp)",
		ContextWindow:     1_000_000,
		InputCostPerMTok:  0.00,
		OutputCostPerMTok: 0.00,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// xAI — marketing defaults (current generation)
	// -------------------------------------------------------------------------
	{
		ID: "grok-4.3", DisplayName: "Grok 4.3 (current gen)",
		ContextWindow:     256_000,
		InputCostPerMTok:  2.00,
		OutputCostPerMTok: 10.00,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},
	{
		ID: "grok-4.3-fast", DisplayName: "Grok 4.3 Fast (current gen)",
		ContextWindow:     256_000,
		InputCostPerMTok:  0.20,
		OutputCostPerMTok: 1.00,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// xAI — legacy
	// -------------------------------------------------------------------------
	{
		ID: "grok-2", DisplayName: "Grok 2",
		ContextWindow:     131_072,
		InputCostPerMTok:  2.00,
		OutputCostPerMTok: 10.00,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "grok-2-mini", DisplayName: "Grok 2 mini",
		ContextWindow:     131_072,
		InputCostPerMTok:  0.20,
		OutputCostPerMTok: 1.00,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// Mistral — marketing defaults (current generation)
	// -------------------------------------------------------------------------
	{
		ID: "mistral-large-3", DisplayName: "Mistral Large 3 (current gen)",
		ContextWindow:     256_000,
		InputCostPerMTok:  2.00,
		OutputCostPerMTok: 6.00,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},
	{
		ID: "codestral-latest", DisplayName: "Codestral (current gen)",
		ContextWindow:     32_000,
		InputCostPerMTok:  0.30,
		OutputCostPerMTok: 0.90,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// Mistral — legacy
	// -------------------------------------------------------------------------
	{
		ID: "mistral-large-latest", DisplayName: "Mistral Large (legacy)",
		ContextWindow:     128_000,
		InputCostPerMTok:  2.00,
		OutputCostPerMTok: 6.00,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "mistral-small-latest", DisplayName: "Mistral Small",
		ContextWindow:     32_000,
		InputCostPerMTok:  0.20,
		OutputCostPerMTok: 0.60,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// DeepSeek — marketing defaults (current generation)
	// -------------------------------------------------------------------------
	{
		ID: "deepseek-v4", DisplayName: "DeepSeek V4 (current gen)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.14,
		OutputCostPerMTok: 0.28,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "deepseek-r1", DisplayName: "DeepSeek R1 (current gen, reasoning)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.55,
		OutputCostPerMTok: 2.19,
		SupportsTools:     false, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// DeepSeek — legacy
	// -------------------------------------------------------------------------
	{
		ID: "deepseek-chat", DisplayName: "DeepSeek Chat (V3, legacy)",
		ContextWindow:     64_000,
		InputCostPerMTok:  0.14,
		OutputCostPerMTok: 0.28,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "deepseek-reasoner", DisplayName: "DeepSeek Reasoner (R1, legacy)",
		ContextWindow:     64_000,
		InputCostPerMTok:  0.55,
		OutputCostPerMTok: 2.19,
		SupportsTools:     false, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// OpenRouter (300+ models accessible via API; this is the meta-router)
	// -------------------------------------------------------------------------
	{
		ID: "openrouter/auto", DisplayName: "OpenRouter Auto (300+ models)",
		ContextWindow:     200_000,
		InputCostPerMTok:  0.50,
		OutputCostPerMTok: 1.50,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// Groq — marketing defaults (current generation)
	// Note: Groq's "Whisper" is STT-only and is handled by the voice
	// subsystem, not the chat LLM registry.
	// -------------------------------------------------------------------------
	{
		ID: "llama-4-70b-versatile", DisplayName: "Llama 4 70B (Groq, current gen)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.59,
		OutputCostPerMTok: 0.79,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "llama-4-8b-instant", DisplayName: "Llama 4 8B (Groq, current gen)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.05,
		OutputCostPerMTok: 0.08,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// Groq — legacy
	// -------------------------------------------------------------------------
	{
		ID: "llama-3.3-70b-versatile", DisplayName: "Llama 3.3 70B (Groq, legacy)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.59,
		OutputCostPerMTok: 0.79,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "llama-3.1-8b-instant", DisplayName: "Llama 3.1 8B (Groq, legacy)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.05,
		OutputCostPerMTok: 0.08,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "mixtral-8x7b-32768", DisplayName: "Mixtral 8x7B (Groq, legacy)",
		ContextWindow:     32_768,
		InputCostPerMTok:  0.24,
		OutputCostPerMTok: 0.24,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// Together — marketing defaults
	// -------------------------------------------------------------------------
	{
		ID: "meta-llama/Llama-4-70B-Instruct", DisplayName: "Llama 4 70B (Together)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.88,
		OutputCostPerMTok: 0.88,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "Qwen/Qwen3-72B-Instruct", DisplayName: "Qwen 3 72B (Together)",
		ContextWindow:     32_000,
		InputCostPerMTok:  0.88,
		OutputCostPerMTok: 0.88,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "mistralai/Mixtral-8x22B-Instruct-v0.1", DisplayName: "Mixtral 8x22B (Together)",
		ContextWindow:     65_536,
		InputCostPerMTok:  1.20,
		OutputCostPerMTok: 1.20,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// Together — legacy
	// -------------------------------------------------------------------------
	{
		ID: "meta-llama/Llama-3.3-70B-Instruct-Turbo", DisplayName: "Llama 3.3 70B (Together, legacy)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.88,
		OutputCostPerMTok: 0.88,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "Qwen/Qwen2.5-72B-Instruct-Turbo", DisplayName: "Qwen 2.5 72B (Together, legacy)",
		ContextWindow:     32_000,
		InputCostPerMTok:  0.88,
		OutputCostPerMTok: 0.88,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// Fireworks — marketing defaults
	// -------------------------------------------------------------------------
	{
		ID: "accounts/fireworks/models/llama-4-70b-instruct", DisplayName: "Llama 4 70B (Fireworks)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.90,
		OutputCostPerMTok: 0.90,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "accounts/fireworks/models/qwen3-72b-instruct", DisplayName: "Qwen 3 72B (Fireworks)",
		ContextWindow:     32_000,
		InputCostPerMTok:  0.90,
		OutputCostPerMTok: 0.90,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "accounts/fireworks/models/deepseek-v4-instruct", DisplayName: "DeepSeek V4 (Fireworks)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.90,
		OutputCostPerMTok: 0.90,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// Fireworks — legacy
	// -------------------------------------------------------------------------
	{
		ID: "accounts/fireworks/models/llama-v3p3-70b-instruct", DisplayName: "Llama 3.3 70B (Fireworks, legacy)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.90,
		OutputCostPerMTok: 0.90,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// Ollama (local; zero cost). The actual model list is whatever the
	// user has pulled (`ollama list`). These are common defaults.
	// -------------------------------------------------------------------------
	{
		ID: "llama3.2", DisplayName: "Llama 3.2 (local)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.0,
		OutputCostPerMTok: 0.0,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "qwen2.5", DisplayName: "Qwen 2.5 (local)",
		ContextWindow:     32_000,
		InputCostPerMTok:  0.0,
		OutputCostPerMTok: 0.0,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// LocalAI (local; zero cost). Models are user-defined in the LocalAI
	// gallery. This entry is a sensible default; users can register
	// any model name they have installed.
	// -------------------------------------------------------------------------
	{
		ID: "llama-3.1-8b-instruct", DisplayName: "Llama 3.1 8B (LocalAI)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.0,
		OutputCostPerMTok: 0.0,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "mistral-7b-instruct", DisplayName: "Mistral 7B (LocalAI)",
		ContextWindow:     32_000,
		InputCostPerMTok:  0.0,
		OutputCostPerMTok: 0.0,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// LM Studio (local; zero cost). Like Ollama/LocalAI: the user picks
	// whatever model they have loaded. These are common defaults.
	// -------------------------------------------------------------------------
	{
		ID: "qwen2.5-7b-instruct", DisplayName: "Qwen 2.5 7B (LM Studio)",
		ContextWindow:     32_000,
		InputCostPerMTok:  0.0,
		OutputCostPerMTok: 0.0,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// vLLM (local; zero cost). Like Ollama/LocalAI: the user picks
	// whatever model they have loaded.
	// -------------------------------------------------------------------------
	{
		ID: "meta-llama/Llama-3.1-8B-Instruct", DisplayName: "Llama 3.1 8B (vLLM)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.0,
		OutputCostPerMTok: 0.0,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
}
