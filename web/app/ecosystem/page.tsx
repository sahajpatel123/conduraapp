"use client";

import { motion } from "motion/react";
import Link from "next/link";
import PageHeader from "@/components/shell/PageHeader";
import Reveal from "@/components/motion/Reveal";
import { TOOL_ROSTER } from "@/lib/site";

const LLM_PROVIDERS = [
  { name: "Anthropic", models: "Claude Opus 4.7, Sonnet 4.5, Haiku 4.5", auth: "API key or Claude Pro OAuth*" },
  { name: "OpenAI", models: "GPT-5.5, o3, o4-mini, gpt-image-2", auth: "API key or ChatGPT Plus OAuth*" },
  { name: "Google", models: "Gemini 3.5 Flash, 2.5 Pro", auth: "API key or Google AI Pro OAuth*" },
  { name: "xAI", models: "Grok-4.3, Grok-4.3-fast", auth: "API key or SuperGrok OAuth*" },
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
  { name: "Ollama", desc: "Local model runner", cmd: "Direct HTTP — no subprocess" },
];

export default function EcosystemPage() {
  return (
    <PageHeader
      eyebrow="Integrations"
      title="Works with every AI"
      titleAccent="you already use."
      description={`Condura doesn't replace your tools — it conducts them. One hotkey routes work across ${LLM_PROVIDERS.length} LLM providers and ${AGENT_CLIS.length} agent CLIs. Bring your own keys, your own models, your own workflow.`}
    >
      {/* ── LLM provider grid ── */}
      <section className="mt-8">
        <Reveal>
          <p className="text-eyebrow mb-4">— AI providers</p>
          <h2 className="text-display text-[var(--color-ink)] max-w-[16ch] text-balance">Connect what you have.</h2>
          <p className="text-lead mt-5 max-w-[54ch] text-[var(--color-ink-soft)] text-pretty">
            Use API keys or, where supported, your existing subscriptions. Condura never stores keys on a server — they stay encrypted on your machine.
          </p>
        </Reveal>

        <div className="mt-10 grid gap-3.5 sm:grid-cols-2 lg:grid-cols-3">
          {LLM_PROVIDERS.map((provider, i) => (
            <motion.div
              key={provider.name}
              initial={{ opacity: 0, y: 16 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-10%" }}
              transition={{ delay: i * 0.04, duration: 0.5, ease: [0.22, 1, 0.36, 1] }}
              className="surface-card p-5 transition-all hover:-translate-y-0.5 hover:bg-[var(--color-paper-deep)]"
            >
              <div className="mb-3 flex items-center justify-between">
                <h3 className="font-display text-[16px] text-[var(--color-ink)]">{provider.name}</h3>
                <span className="font-mono text-[9.5px] uppercase tracking-wider text-[var(--color-ink-faint)]">{provider.auth}</span>
              </div>
              <p className="font-mono text-[12.5px] leading-relaxed text-[var(--color-ink-mute)]">{provider.models}</p>
            </motion.div>
          ))}
        </div>
        <p className="mt-5 text-small text-[var(--color-ink-faint)]">* Subscription OAuth is on the v0.2.0 roadmap. v0.1.x uses API keys.</p>
      </section>

      {/* ── Agent CLIs ── */}
      <section className="mt-28">
        <Reveal>
          <p className="text-eyebrow mb-4">— Agent CLIs</p>
          <h2 className="text-display text-[var(--color-ink)] max-w-[16ch] text-balance">Sub-agents on your $PATH.</h2>
          <p className="text-lead mt-5 max-w-[54ch] text-[var(--color-ink-soft)] text-pretty">
            Condura auto-detects every agent CLI you have installed and spawns them as sub-agents. Each runs in its own sandbox with model isolation. Missing a CLI? It simply doesn&apos;t appear — no installs forced.
          </p>
        </Reveal>

        <div className="mt-10 grid gap-3 sm:grid-cols-2 lg:grid-cols-4">
          {AGENT_CLIS.map((cli, i) => (
            <motion.div
              key={cli.name}
              initial={{ opacity: 0, scale: 0.96 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true, margin: "-10%" }}
              transition={{ delay: i * 0.06, duration: 0.4 }}
              className="surface-card p-4"
            >
              <div className="mb-3 grid h-9 w-9 place-items-center rounded-lg border border-[rgba(20,17,11,0.12)] bg-[var(--color-paper)]">
                <span className="font-mono text-[12px] font-medium text-[var(--color-ink-soft)]">{cli.name[0]}</span>
              </div>
              <h3 className="text-[14.5px] font-semibold text-[var(--color-ink)]">{cli.name}</h3>
              <p className="mt-1 text-[12px] leading-relaxed text-[var(--color-ink-mute)]">{cli.desc}</p>
              <code className="mt-2 block truncate font-mono text-[10px] text-[var(--color-ink-faint)]">{cli.cmd}</code>
            </motion.div>
          ))}
        </div>
      </section>

      {/* ── Routing ── */}
      <section className="mt-28">
        <Reveal>
          <div className="mb-12 text-center">
            <p className="text-eyebrow mb-3">— Routing</p>
            <h2 className="text-display text-[var(--color-ink)] max-w-[16ch] mx-auto text-balance">Every model. One interface.</h2>
            <p className="text-lead mt-5 max-w-2xl mx-auto text-[var(--color-ink-soft)] text-pretty">
              The hybrid router learns which model works best for each task. Start with the cheapest, escalate on failure, bias toward what has worked before.
            </p>
          </div>
        </Reveal>
        <div className="grid gap-5 md:grid-cols-3">
          {[
            { title: "Cascade", desc: "Try the cheapest model first. If it fails the quality gate, escalate to the next tier. No wasted spend on trivial tasks.", icon: "cascade" },
            { title: "Memory bias", desc: "After enough samples, the router learns which model succeeds most for your specific coding, writing, or research patterns.", icon: "memory" },
            { title: "User override", desc: "Pin specific providers to specific task types. Claude for code, Gemini for research, local for drafts — your preference wins.", icon: "lock" },
          ].map((card, i) => (
            <Reveal key={card.title} delay={i * 0.1}>
              <div className="surface-card h-full p-7">
                <div className="mb-5 grid h-10 w-10 place-items-center rounded-full border border-[rgba(11,61,46,0.3)] bg-[rgba(11,61,46,0.08)] text-[var(--color-synapse)]">
                  <RouteIcon name={card.icon} />
                </div>
                <h3 className="font-display text-[20px] leading-tight text-[var(--color-ink)]">{card.title}</h3>
                <p className="mt-2.5 text-body text-[var(--color-ink-mute)]">{card.desc}</p>
              </div>
            </Reveal>
          ))}
        </div>
      </section>

      {/* ── Bring your own ── */}
      <Reveal>
        <section className="mt-28">
          <div className="surface-ink p-10 text-center sm:p-14">
            <p className="text-mono-label !text-[var(--color-pollen)]">— Bring your own everything</p>
            <h2 className="mt-4 font-display text-[clamp(28px,4vw,44px)] leading-tight text-[var(--color-paper)] text-balance">
              Run entirely offline. Or point it at a custom endpoint.
            </h2>
            <p className="text-lead mt-5 max-w-2xl mx-auto text-[rgba(244,239,228,0.7)] text-pretty">
              Bring your own API keys from any provider. Use Ollama for everything. The choice is yours — and yours alone.
            </p>
            <div className="mt-9 flex flex-wrap justify-center gap-2.5">
              {TOOL_ROSTER.map((tool) => (
                <span key={tool} className="rounded-full border border-[rgba(244,239,228,0.16)] bg-[rgba(244,239,228,0.06)] px-4 py-2 font-mono text-[12.5px] text-[rgba(244,239,228,0.8)]">
                  {tool}
                </span>
              ))}
            </div>
            <p className="mt-6 font-mono text-[12px] text-[rgba(244,239,228,0.45)]">
              Auto-detected on your <span className="text-[rgba(244,239,228,0.7)]">$PATH</span> — no config needed
            </p>
            <Link href="/download" prefetch className="btn btn-pollen mt-9 inline-flex">
              Download Condura
              <span aria-hidden>→</span>
            </Link>
          </div>
        </section>
      </Reveal>
    </PageHeader>
  );
}

function RouteIcon({ name }: { name: string }) {
  if (name === "cascade")
    return (
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
        <path d="M12 2L2 7l10 5 10-5-10-5z" /><path d="M2 17l10 5 10-5" /><path d="M2 12l10 5 10-5" />
      </svg>
    );
  if (name === "memory")
    return (
      <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
        <circle cx="12" cy="12" r="10" /><path d="M12 6v6l4 2" />
      </svg>
    );
  return (
    <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
      <rect x="3" y="11" width="18" height="11" rx="2" /><path d="M7 11V7a5 5 0 0110 0v4" />
    </svg>
  );
}
