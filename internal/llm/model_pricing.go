package llm

// Pricing snapshot — last reviewed: Phase 1.
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
// All values are best-effort and may drift. The failover layer falls back
// to 0 if a model is unknown; the user can override pricing per-request.

func init() {
	for _, m := range models {
		modelRegistry[m.ID] = m
	}
}

var models = []ModelInfo{
	// -------------------------------------------------------------------------
	// OpenAI
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
	// Anthropic
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
	// Google Gemini
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
	// xAI
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
	// Mistral
	// -------------------------------------------------------------------------
	{
		ID: "mistral-large-latest", DisplayName: "Mistral Large",
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
	{
		ID: "codestral-latest", DisplayName: "Codestral",
		ContextWindow:     32_000,
		InputCostPerMTok:  0.30,
		OutputCostPerMTok: 0.90,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// DeepSeek
	// -------------------------------------------------------------------------
	{
		ID: "deepseek-chat", DisplayName: "DeepSeek Chat (V3)",
		ContextWindow:     64_000,
		InputCostPerMTok:  0.14,
		OutputCostPerMTok: 0.28,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "deepseek-reasoner", DisplayName: "DeepSeek Reasoner (R1)",
		ContextWindow:     64_000,
		InputCostPerMTok:  0.55,
		OutputCostPerMTok: 2.19,
		SupportsTools:     false, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// Groq (OpenAI-compat)
	// -------------------------------------------------------------------------
	{
		ID: "llama-3.3-70b-versatile", DisplayName: "Llama 3.3 70B (Groq)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.59,
		OutputCostPerMTok: 0.79,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "llama-3.1-8b-instant", DisplayName: "Llama 3.1 8B (Groq)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.05,
		OutputCostPerMTok: 0.08,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "mixtral-8x7b-32768", DisplayName: "Mixtral 8x7B (Groq)",
		ContextWindow:     32_768,
		InputCostPerMTok:  0.24,
		OutputCostPerMTok: 0.24,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// Together (OpenAI-compat)
	// -------------------------------------------------------------------------
	{
		ID: "meta-llama/Llama-3.3-70B-Instruct-Turbo", DisplayName: "Llama 3.3 70B (Together)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.88,
		OutputCostPerMTok: 0.88,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},
	{
		ID: "Qwen/Qwen2.5-72B-Instruct-Turbo", DisplayName: "Qwen 2.5 72B (Together)",
		ContextWindow:     32_000,
		InputCostPerMTok:  0.88,
		OutputCostPerMTok: 0.88,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// Fireworks (OpenAI-compat)
	// -------------------------------------------------------------------------
	{
		ID: "accounts/fireworks/models/llama-v3p3-70b-instruct", DisplayName: "Llama 3.3 70B (Fireworks)",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.90,
		OutputCostPerMTok: 0.90,
		SupportsTools:     true, SupportsVision: false, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// OpenRouter (meta; specific models priced by router)
	// -------------------------------------------------------------------------
	{
		ID: "openrouter/auto", DisplayName: "OpenRouter Auto",
		ContextWindow:     128_000,
		InputCostPerMTok:  0.50,
		OutputCostPerMTok: 1.50,
		SupportsTools:     true, SupportsVision: true, SupportsStream: true,
	},

	// -------------------------------------------------------------------------
	// Ollama (local; zero cost)
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
}
