"use client";

import PageChrome from "@/components/shell/PageChrome";
import { motion } from "motion/react";
import { TOOL_ROSTER, SITE } from "@/lib/site";

const LLM_PROVIDERS = [
  { name: "Anthropic", models: "Claude Opus 4.7, Sonnet 4.5, Haiku 4.5", auth: "API key or Claude Pro OAuth" },
  { name: "OpenAI", models: "GPT-5.5, o3, o4-mini, gpt-image-2", auth: "API key or ChatGPT Plus OAuth" },
  { name: "Google", models: "Gemini 3.5 Flash, 2.5 Pro", auth: "API key or Google AI Pro OAuth" },
  { name: "xAI", models: "Grok-4.3, Grok-4.3-fast", auth: "API key or SuperGrok OAuth" },
  { name: "Mistral", models: "Mistral Large 3, Codestral, Pixtral", auth: "API key" },
  { name: "DeepSeek", models: "DeepSeek-V4, R1", auth: "API key" },
  { name: "OpenRouter", models: "300+ models", auth: "API key" },
  { name: "Together", models: "Llama, Qwen, Mixtral", auth: "API key" },
  { name: "Groq", models: "Llama 4, Mixtral, Whisper", auth: "API key" },
  { name: "Fireworks", models: "Llama, Qwen, DeepSeek", auth: "API key" },
  { name: "Local", models: "Ollama, LM Studio, vLLM, llama.cpp", auth: "None — runs locally" },
  { name: "Custom", models: "Any OpenAI-compatible endpoint", auth: "API key + base URL" },
];

const AGENT_CLIS = [
  { name: "Claude Code", desc: "Anthropic's terminal coding agent", cmd: "claude --print --output-format stream-json" },
  { name: "Codex", desc: "OpenAI's terminal coding agent", cmd: "codex --json" },
  { name: "Antigravity", desc: "Open-source agent framework", cmd: "agy --output-format json" },
  { name: "OpenCode", desc: "Terminal coding assistant", cmd: "opencode --format json" },
  { name: "Kilo Code", desc: "Agentic coding CLI", cmd: "kilo --json" },
  { name: "Hermes Agent", desc: "Multi-tool autonomous agent", cmd: "hermes --format json" },
  { name: "Gemini CLI", desc: "Google's terminal agent", cmd: "gemini --output-format json" },
  { name: "Ollama", desc: "Local model runner", cmd: "Direct HTTP — no subprocess needed" },
];

export default function EcosystemPage() {
  return (
    <div className="bg-black text-white min-h-screen">
      <PageChrome
        eyebrow="Integrations"
        title="Works with every AI you already use."
        description={`Condura doesn't replace your tools — it conducts them. One hotkey routes work across ${LLM_PROVIDERS.length} LLM providers and ${AGENT_CLIS.length} agent CLIs. Bring your own keys, your own models, your own workflow.`}
        badge="Ecosystem"
      >
        {/* --- LLM Provider Grid --- */}
        <div className="mt-24">
          <div className="max-w-6xl mx-auto px-8">
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-5xl font-semibold tracking-tight mb-4 text-white">
                AI providers
              </h2>
              <p className="text-lg text-white/50 max-w-2xl mx-auto">
                Connect your existing subscriptions — API keys or OAuth. Condura never stores keys on a server; they stay encrypted on your machine.
              </p>
            </div>

            <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
              {LLM_PROVIDERS.map((provider, i) => (
                <motion.div
                  key={provider.name}
                  initial={{ opacity: 0, y: 20 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ delay: i * 0.05, duration: 0.5 }}
                  className="rounded-2xl border border-white/[0.08] bg-[#0a0a0a] p-5 hover:border-white/[0.15] transition-colors"
                >
                  <div className="flex items-center justify-between mb-3">
                    <h3 className="text-white font-medium text-[15px]">{provider.name}</h3>
                    <span className="text-[10px] font-mono text-white/30 uppercase tracking-wider">{provider.auth}</span>
                  </div>
                  <p className="text-white/50 text-[13px] leading-relaxed font-mono">{provider.models}</p>
                </motion.div>
              ))}
            </div>
          </div>
        </div>

        {/* --- Agent CLI Section --- */}
        <div className="mt-40 mb-32">
          <div className="max-w-6xl mx-auto px-8">
            <div className="text-center mb-16">
              <h2 className="text-3xl md:text-5xl font-semibold tracking-tight mb-4 text-white">
                Agent CLIs
              </h2>
              <p className="text-lg text-white/50 max-w-2xl mx-auto">
                Condura auto-detects every agent CLI on your <code className="text-white/70">$PATH</code> and spawns them as sub-agents. Each runs in its own sandbox with model isolation.
              </p>
            </div>

            <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-3">
              {AGENT_CLIS.map((cli, i) => (
                <motion.div
                  key={cli.name}
                  initial={{ opacity: 0, scale: 0.95 }}
                  animate={{ opacity: 1, scale: 1 }}
                  transition={{ delay: i * 0.08, duration: 0.4 }}
                  className="rounded-xl border border-white/[0.06] bg-[#050505] p-4"
                >
                  <div className="w-8 h-8 rounded-lg bg-white/[0.06] flex items-center justify-center mb-3 border border-white/[0.06]">
                    <span className="text-white/70 text-[11px] font-mono font-medium">{cli.name[0]}</span>
                  </div>
                  <h3 className="text-white text-[14px] font-medium mb-1">{cli.name}</h3>
                  <p className="text-white/40 text-[12px] leading-relaxed mb-2">{cli.desc}</p>
                  <code className="block text-[10px] font-mono text-white/25 truncate">{cli.cmd}</code>
                </motion.div>
              ))}
            </div>
          </div>
        </div>

        {/* --- How Routing Works --- */}
        <div className="mt-32 max-w-5xl mx-auto px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl md:text-5xl font-semibold tracking-tight mb-4 text-white">
              Every model. One interface.
            </h2>
            <p className="text-lg text-white/50 max-w-2xl mx-auto">
              Condura's hybrid router learns which model works best for each task. Start with the cheapest, escalate on failure, and bias toward what has worked before.
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-6">
            <div className="p-8 rounded-[32px] bg-[#050505] border border-white/10">
              <div className="w-10 h-10 mb-6 rounded-full bg-white/10 flex items-center justify-center">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="text-white/60"><path d="M12 2L2 7l10 5 10-5-10-5z"/><path d="M2 17l10 5 10-5"/><path d="M2 12l10 5 10-5"/></svg>
              </div>
              <h3 className="text-xl font-medium text-white mb-3">Cascade</h3>
              <p className="text-white/40 text-sm leading-relaxed">Try the cheapest model first. If it fails the quality gate, escalate to the next tier. No wasted spend on trivial tasks.</p>
            </div>
            <div className="p-8 rounded-[32px] bg-[#050505] border border-white/10">
              <div className="w-10 h-10 mb-6 rounded-full bg-white/10 flex items-center justify-center">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="text-white/60"><circle cx="12" cy="12" r="10"/><path d="M12 6v6l4 2"/></svg>
              </div>
              <h3 className="text-xl font-medium text-white mb-3">Memory bias</h3>
              <p className="text-white/40 text-sm leading-relaxed">After enough samples, the router learns which model succeeds most for your specific coding, writing, or research patterns.</p>
            </div>
            <div className="p-8 rounded-[32px] bg-[#050505] border border-white/10">
              <div className="w-10 h-10 mb-6 rounded-full bg-white/10 flex items-center justify-center">
                <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" className="text-white/60"><rect x="3" y="11" width="18" height="11" rx="2"/><path d="M7 11V7a5 5 0 0110 0v4"/></svg>
              </div>
              <h3 className="text-xl font-medium text-white mb-3">User override</h3>
              <p className="text-white/40 text-sm leading-relaxed">Pin specific providers to specific task types. Claude for code, Gemini for research, local for drafts — your preference is the strongest signal.</p>
            </div>
          </div>
        </div>

        {/* --- Bring your own --- */}
        <div className="mt-40 mb-32 max-w-5xl mx-auto px-8 text-center">
          <div className="rounded-[32px] border border-white/[0.08] bg-[#050505] p-12">
            <h2 className="text-3xl md:text-4xl font-semibold tracking-tight mb-6 text-white">
              Bring your own everything
            </h2>
            <p className="text-lg text-white/50 max-w-2xl mx-auto mb-8">
              Run entirely offline with Ollama. Use your ChatGPT Plus or Claude Pro subscription instead of API keys. Point Condura at a custom endpoint. The choice is yours — and yours alone.
            </p>
            <div className="flex flex-wrap justify-center gap-3">
              {TOOL_ROSTER.map((tool) => (
                <span
                  key={tool}
                  className="rounded-full border border-white/[0.08] bg-white/[0.04] px-4 py-2 text-[13px] text-white/60 font-mono"
                >
                  {tool}
                </span>
              ))}
            </div>
            <p className="mt-6 text-[13px] text-white/30 font-mono">
              Auto-detected on your <code className="text-white/50">$PATH</code> — no config needed
            </p>
          </div>
        </div>

      </PageChrome>
    </div>
  );
}
