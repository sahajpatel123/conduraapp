"use client";

import { motion, AnimatePresence } from "motion/react";
import { useState, useEffect } from "react";
import AnimatedBadge from "@/components/motion/AnimatedBadge";
import TiltCard from "@/components/motion/TiltCard";
import MagneticButton from "@/components/motion/MagneticButton";
import { TOOL_ROSTER } from "@/lib/site";
import { EASE_OUT } from "@/lib/motion";

/* ────────────────────────────────────────────────────────────
   ECOSYSTEM — Plugs Into Everything You Use
   Condura auto-detects installed CLIs and API platforms,
   hooks their standard outputs, and orchestrates them
   through a unified delegation bus.
   ──────────────────────────────────────────────────────────── */

const CLI_TOOLS = [
  { name: "Claude Code", cmd: "claude --print --output-format stream-json", role: "Full-stack coding agent", auth: "API key or Pro OAuth", color: "#d97757" },
  { name: "Codex", cmd: "codex --json --model", role: "GPT-powered code generation", auth: "ChatGPT Plus/Pro OAuth", color: "#10a37f" },
  { name: "Antigravity", cmd: "agy --output-format json", role: "Background-first automation", auth: "API key", color: "#8b5cf6" },
  { name: "OpenCode", cmd: "opencode --format json", role: "Universal code orchestrator", auth: "API key", color: "#3b82f6" },
  { name: "Kilo Code", cmd: "kilo --json", role: "Lightweight code tasks", auth: "API key", color: "#ec4899" },
  { name: "Hermes Agent", cmd: "hermes --format json", role: "Multi-model routing agent", auth: "API key", color: "#f59e0b" },
  { name: "Gemini CLI", cmd: "gemini --output-format json", role: "Google AI code assistant", auth: "Google AI Pro OAuth", color: "#4285f4" },
  { name: "Ollama", cmd: "HTTP localhost:11434", role: "100% local model runner", auth: "None — fully local", color: "#6b7280" },
];

const LLM_PROVIDERS = [
  { name: "Anthropic", models: "Opus 4.7, Sonnet 4.5, Haiku 4.5" },
  { name: "OpenAI", models: "GPT-5.5, o3, o4-mini" },
  { name: "Google", models: "Gemini 3.5 Flash, 3.1 Pro" },
  { name: "xAI", models: "Grok-4.3, Grok-4.3-fast" },
  { name: "Mistral", models: "Large 3, Codestral" },
  { name: "DeepSeek", models: "V4, R1" },
  { name: "OpenRouter", models: "300+ models" },
  { name: "Together", models: "Llama, Qwen, Mixtral" },
  { name: "Groq", models: "Llama 4, Whisper" },
  { name: "Fireworks", models: "Llama, Qwen, DeepSeek" },
  { name: "Local", models: "Ollama, LM Studio, vLLM" },
  { name: "Custom", models: "Any OpenAI-compatible endpoint" },
];

const DETECTION_SEQUENCE = [
  { path: "/usr/local/bin/claude", found: true, label: "Claude Code" },
  { path: "/opt/homebrew/bin/codex", found: true, label: "Codex" },
  { path: "/usr/local/bin/agy", found: true, label: "Antigravity" },
  { path: "/usr/local/bin/opencode", found: true, label: "OpenCode" },
  { path: "/usr/local/bin/kilo", found: false, label: "Kilo Code" },
  { path: "/usr/local/bin/hermes", found: false, label: "Hermes" },
  { path: "/usr/local/bin/gemini", found: true, label: "Gemini CLI" },
  { path: "localhost:11434", found: true, label: "Ollama" },
];

export default function EcosystemPage() {
  return (
    <main className="relative w-full bg-black text-white overflow-hidden">
      <EcosystemHero />
      <ToolGridSection />
      <DetectionDemoSection />
      <ProviderMatrixSection />
      <ProtocolSection />
      <ClosingCTA />
    </main>
  );
}

/* ════════════════════════════════════════════════════════════
   1. HERO
   ════════════════════════════════════════════════════════════ */

function EcosystemHero() {
  const [mounted, setMounted] = useState(false);
  useEffect(() => { const t = setTimeout(() => setMounted(true), 200); return () => clearTimeout(t); }, []);

  return (
    <section className="relative min-h-screen flex flex-col items-center justify-center px-6 overflow-hidden">
      <div className="absolute inset-0 bg-grid-dark opacity-20" />

      {/* Constellation background */}
      <div className="absolute inset-0 pointer-events-none">
        {TOOL_ROSTER.map((tool, i) => {
          const angle = (i / TOOL_ROSTER.length) * Math.PI * 2;
          const radius = 280;
          const x = 50 + (Math.cos(angle) * radius) / 8;
          const y = 50 + (Math.sin(angle) * radius) / 8;
          return (
            <motion.div
              key={tool}
              initial={{ opacity: 0, scale: 0 }}
              animate={{ opacity: mounted ? 0.4 : 0, scale: mounted ? 1 : 0 }}
              transition={{ delay: 0.5 + i * 0.08, duration: 0.6 }}
              style={{ left: `${x}%`, top: `${y}%` }}
              className="absolute flex h-12 w-12 items-center justify-center rounded-xl border border-white/[0.08] bg-white/[0.02] backdrop-blur-sm"
            >
              <span className="font-mono text-[10px] text-white/30">{tool[0]}</span>
            </motion.div>
          );
        })}
        {/* Connecting lines from center */}
        <svg className="absolute inset-0 w-full h-full" preserveAspectRatio="none">
          {TOOL_ROSTER.map((_, i) => {
            const angle = (i / TOOL_ROSTER.length) * Math.PI * 2;
            const radius = 280;
            const x = 50 + (Math.cos(angle) * radius) / 8;
            const y = 50 + (Math.sin(angle) * radius) / 8;
            return (
              <motion.line
                key={i}
                x1="50%"
                y1="50%"
                x2={`${x}%`}
                y2={`${y}%`}
                stroke="rgba(255,255,255,0.06)"
                strokeWidth="1"
                initial={{ pathLength: 0 }}
                animate={{ pathLength: mounted ? 1 : 0 }}
                transition={{ delay: 0.5 + i * 0.08, duration: 0.8 }}
              />
            );
          })}
        </svg>
      </div>

      <div className="relative z-10 max-w-4xl text-center">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: mounted ? 1 : 0, y: mounted ? 0 : 20 }}
          transition={{ duration: 1, ease: EASE_OUT }}
        >
          <div className="mb-8 flex justify-center">
            <AnimatedBadge tone="neutral" pulse>Ecosystem</AnimatedBadge>
          </div>

          <h1 className="font-display text-[clamp(2.5rem,7vw,5rem)] font-semibold leading-[1.05] tracking-[-0.04em]">
            Plugs into everything
            <br />
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-white via-white to-white/30">
              you already use.
            </span>
          </h1>

          <p className="mt-8 mx-auto max-w-2xl font-lead-airy">
            Condura doesn&apos;t ask you to change your stack. It auto-detects installed coding CLIs
            in your <code className="font-mono text-[14px] text-white/60">$PATH</code>, hooks their
            standard JSON outputs, and routes work through 12+ LLM providers — including 100% local
            models via Ollama.
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: mounted ? 1 : 0 }}
          transition={{ delay: 0.8, duration: 1 }}
          className="mt-12 flex flex-wrap items-center justify-center gap-6"
        >
          {[
            { label: "CLI tools", value: "8" },
            { label: "LLM providers", value: "12+" },
            { label: "Local models", value: "unlimited" },
          ].map((stat) => (
            <div key={stat.label} className="flex flex-col items-center">
              <span className="font-mono text-[22px] font-medium text-white">{stat.value}</span>
              <span className="mt-1 font-mono text-[10px] uppercase tracking-widest text-white/25">{stat.label}</span>
            </div>
          ))}
        </motion.div>
      </div>

      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: mounted ? 1 : 0 }}
        transition={{ delay: 1.5, duration: 1 }}
        className="absolute bottom-10 left-1/2 -translate-x-1/2 flex flex-col items-center gap-2"
      >
        <span className="font-mono text-[10px] uppercase tracking-widest text-white/25">Scroll</span>
        <div className="w-[1px] h-10 bg-gradient-to-b from-white/25 to-transparent" />
      </motion.div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   2. TOOL GRID — CLI Sub-Agents
   ════════════════════════════════════════════════════════════ */

function ToolGridSection() {
  return (
    <section className="relative w-full py-[160px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-20 max-w-3xl">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Sub-Agent CLIs</span>
          <h2 className="mt-4 font-display text-[clamp(2rem,5vw,3.5rem)] font-semibold tracking-[-0.03em] leading-[1.1]">
            Eight CLIs. One conductor.
          </h2>
          <p className="mt-6 font-lead-airy">
            Each tool is spawned as a subprocess with sanitized inputs. Condura parses the JSON
            output stream, tracks exit codes, and displays logs natively. If a CLI isn&apos;t
            installed, you get a friendly &ldquo;Install X?&rdquo; prompt with a link to docs.
          </p>
        </div>

        <div className="grid sm:grid-cols-2 gap-6">
          {CLI_TOOLS.map((tool, i) => (
            <TiltCard key={tool.name} maxRotate={6} className="h-full">
              <motion.div
                initial={{ opacity: 0, y: 30 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: i * 0.06, type: "spring", stiffness: 100, damping: 15 }}
                className="group relative h-full overflow-hidden rounded-2xl border border-white/[0.10] bg-white/[0.02] p-6 backdrop-blur-md transition-colors hover:bg-white/[0.04]"
              >
                {/* Color accent bar */}
                <div className="absolute top-0 left-0 right-0 h-[2px]" style={{ background: tool.color, opacity: 0.5 }} />

                <div className="flex items-start justify-between mb-6">
                  <div
                    className="flex h-12 w-12 items-center justify-center rounded-xl border border-white/15"
                    style={{ background: `${tool.color}12` }}
                  >
                    <span className="font-mono text-[16px]" style={{ color: tool.color }}>{tool.name[0]}</span>
                  </div>
                  <span className="rounded-full border border-white/[0.08] bg-white/[0.03] px-2.5 py-0.5 font-mono text-[10px] text-white/40">
                    CLI
                  </span>
                </div>

                <h3 className="font-body-mature text-[17px] font-semibold text-white">{tool.name}</h3>
                <p className="mt-1 font-body-mature text-[14px] text-white/45">{tool.role}</p>

                {/* Command */}
                <div className="mt-4 rounded-lg border border-white/[0.06] bg-black/40 px-3 py-2">
                  <span className="font-mono text-[11px] text-white/35">
                    <span className="text-white/20">$ </span>
                    {tool.cmd}
                  </span>
                </div>

                {/* Auth */}
                <div className="mt-4 flex items-center gap-2">
                  <span className="font-mono text-[10px] uppercase tracking-wider text-white/20">Auth</span>
                  <span className="font-body-mature text-[12px] text-white/50">{tool.auth}</span>
                </div>

                {/* Hover glow */}
                <div
                  className="absolute inset-0 opacity-0 transition-opacity duration-500 group-hover:opacity-100"
                  style={{ background: `radial-gradient(circle at 50% 0%, ${tool.color}08, transparent 70%)` }}
                />
              </motion.div>
            </TiltCard>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   3. DETECTION DEMO — Live $PATH Scan
   ════════════════════════════════════════════════════════════ */

function DetectionDemoSection() {
  const [scanIndex, setScanIndex] = useState(0);
  const [scanned, setScanned] = useState<boolean[]>([]);

  useEffect(() => {
    if (scanIndex >= DETECTION_SEQUENCE.length) {
      const reset = setTimeout(() => {
        setScanIndex(0);
        setScanned([]);
      }, 4000);
      return () => clearTimeout(reset);
    }
    const t = setTimeout(() => {
      setScanned((prev) => [...prev, DETECTION_SEQUENCE[scanIndex].found]);
      setScanIndex((p) => p + 1);
    }, 500);
    return () => clearTimeout(t);
  }, [scanIndex]);

  const foundCount = scanned.filter(Boolean).length;

  return (
    <section className="relative w-full py-[160px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-20 max-w-3xl">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Auto-Detection</span>
          <h2 className="mt-4 font-display text-[clamp(2rem,5vw,3.5rem)] font-semibold tracking-[-0.03em] leading-[1.1]">
            Zero config. Just works.
          </h2>
          <p className="mt-6 font-lead-airy">
            On first launch, Condura walks your <code className="font-mono text-[14px] text-white/60">$PATH</code> and
            pings <code className="font-mono text-[14px] text-white/60">localhost:11434</code> for Ollama.
            Everything it finds is immediately available. Everything it doesn&apos;t gets an install prompt.
          </p>
        </div>

        {/* Detection terminal */}
        <div className="overflow-hidden rounded-2xl border border-white/[0.10] bg-[#0e0e0e] shadow-[0_40px_80px_rgba(0,0,0,0.5)]">
          <div className="h-[40px] border-b border-white/[0.06] bg-[#1a1a1a] flex items-center px-4">
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-[#ff5f57]" />
              <div className="w-3 h-3 rounded-full bg-[#febc2e]" />
              <div className="w-3 h-3 rounded-full bg-[#28c840]" />
            </div>
            <span className="ml-4 font-mono text-[12px] text-white/25">condura detect --verbose</span>
          </div>

          <div className="p-6 min-h-[360px] font-mono text-[13px] space-y-2">
            <p className="text-white/50">
              <span className="text-white/30">❯ </span>
              scanning $PATH for known CLIs…
            </p>

            <AnimatePresence>
              {DETECTION_SEQUENCE.slice(0, scanIndex).map((item) => (
                <motion.div
                  key={item.path}
                  initial={{ opacity: 0, x: -10 }}
                  animate={{ opacity: 1, x: 0 }}
                  className="flex items-center gap-3"
                >
                  <span className="text-white/25 w-6">{item.found ? "✓" : "✗"}</span>
                  <span className={`flex-1 ${item.found ? "text-white/60" : "text-white/25"}`}>
                    {item.path}
                  </span>
                  <span className={item.found ? "text-green-400/60" : "text-white/20"}>
                    {item.found ? "found" : "not found"}
                  </span>
                  {item.found && (
                    <span className="rounded border border-white/[0.06] px-1.5 py-0.5 text-[10px] text-white/40">
                      {item.label}
                    </span>
                  )}
                </motion.div>
              ))}
            </AnimatePresence>

            {scanIndex >= DETECTION_SEQUENCE.length && (
              <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                className="mt-4 border-t border-white/[0.06] pt-4"
              >
                <p className="text-white/50">
                  Detection complete. {foundCount} tools found, {DETECTION_SEQUENCE.length - foundCount} missing.
                </p>
                <p className="mt-2 text-white/30">
                  Missing tools will show an &ldquo;Install?&rdquo; prompt in the dashboard.
                </p>
              </motion.div>
            )}

            {scanIndex < DETECTION_SEQUENCE.length && (
              <div className="flex items-center gap-2 pt-2">
                <span className="text-white/30">❯</span>
                <motion.span
                  animate={{ opacity: [1, 0] }}
                  transition={{ repeat: Infinity, duration: 0.9 }}
                  className="inline-block w-[7px] h-[14px] bg-white/30"
                />
              </div>
            )}
          </div>
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   4. PROVIDER MATRIX — 12+ LLM Providers
   ════════════════════════════════════════════════════════════ */

function ProviderMatrixSection() {
  return (
    <section className="relative w-full py-[160px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-20 max-w-3xl">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">LLM Providers</span>
          <h2 className="mt-4 font-display text-[clamp(2rem,5vw,3.5rem)] font-semibold tracking-[-0.03em] leading-[1.1]">
            Bring your own model.
          </h2>
          <p className="mt-6 font-lead-airy">
            Condura routes through every major provider — by API key or by OAuth into your existing
            subscriptions. No vendor lock-in. No mandatory account. The router picks the cheapest
            model above your quality threshold, and fails over to local Ollama when a provider goes down.
          </p>
        </div>

        {/* Provider grid */}
        <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-4">
          {LLM_PROVIDERS.map((provider, i) => (
            <motion.div
              key={provider.name}
              initial={{ opacity: 0, scale: 0.95 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.04, type: "spring", stiffness: 200, damping: 20 }}
              className="group rounded-xl border border-white/[0.08] bg-white/[0.02] p-5 transition-colors hover:border-white/15 hover:bg-white/[0.04]"
            >
              <div className="flex items-center justify-between mb-3">
                <span className="font-body-mature text-[15px] font-semibold text-white">{provider.name}</span>
                <span className="flex h-2 w-2 rounded-full bg-green-400/40 shadow-[0_0_8px_rgba(74,222,128,0.3)]" />
              </div>
              <p className="font-mono text-[11px] text-white/35 leading-relaxed">{provider.models}</p>
            </motion.div>
          ))}
        </div>

        {/* Router explainer */}
        <div className="mt-16 mature-panel rounded-2xl p-8">
          <h3 className="font-body-mature text-[18px] font-semibold text-white mb-6">
            The hybrid router.
          </h3>
          <div className="grid md:grid-cols-4 gap-4">
            {[
              { step: "01", title: "Cascade", desc: "Try the cheapest model first. Escalate on failure." },
              { step: "02", title: "Pareto", desc: "Pick the cheapest model above your quality bar." },
              { step: "03", title: "Memory bias", desc: "After N samples, prefer what worked for this task type." },
              { step: "04", title: "User override", desc: "Your priority list wins. Always." },
            ].map((s) => (
              <div key={s.step} className="rounded-xl border border-white/[0.06] bg-white/[0.02] p-4">
                <span className="font-mono text-[11px] text-white/30">{s.step}</span>
                <h4 className="mt-2 font-body-mature text-[14px] font-semibold text-white">{s.title}</h4>
                <p className="mt-1 font-body-mature text-[12px] text-white/40 leading-relaxed">{s.desc}</p>
              </div>
            ))}
          </div>
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   5. PROTOCOL — How Condura Talks to CLIs
   ════════════════════════════════════════════════════════════ */

function ProtocolSection() {
  return (
    <section className="relative w-full py-[160px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-20 max-w-3xl">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Protocol</span>
          <h2 className="mt-4 font-display text-[clamp(2rem,5vw,3.5rem)] font-semibold tracking-[-0.03em] leading-[1.1]">
            Model isolation, not just switching.
          </h2>
          <p className="mt-6 font-lead-airy">
            If Claude generates a script and Ollama executes it, Ollama gets no implicit context
            about what Claude intended. Every handoff is explicit and sanitized. Shell commands are
            parsed, validated against an allowlist, and only then passed to a sandboxed executor.
          </p>
        </div>

        {/* Sanitization flow */}
        <div className="mature-panel rounded-2xl p-8">
          <div className="flex flex-col lg:flex-row items-stretch gap-4">
            {[
              { label: "Model A output", sub: "raw text / code", color: "#d97757" },
              { label: "Sanitizer", sub: "AST parse · allowlist · SSRF check", color: "#ffffff" },
              { label: "Sandbox", sub: "isolated executor · no network", color: "#3b82f6" },
              { label: "Result", sub: "structured JSON · audited", color: "#10a37f" },
            ].map((stage, i) => (
              <div key={stage.label} className="flex items-center gap-4 flex-1">
                <motion.div
                  initial={{ opacity: 0, scale: 0.9 }}
                  whileInView={{ opacity: 1, scale: 1 }}
                  viewport={{ once: true }}
                  transition={{ delay: i * 0.15 }}
                  className="flex-1 rounded-xl border border-white/[0.10] bg-white/[0.02] p-5 text-center"
                >
                  <div
                    className="mx-auto mb-3 flex h-10 w-10 items-center justify-center rounded-full border"
                    style={{ borderColor: `${stage.color}40`, background: `${stage.color}10` }}
                  >
                    <span className="font-mono text-[12px]" style={{ color: stage.color }}>{i + 1}</span>
                  </div>
                  <h4 className="font-body-mature text-[14px] font-semibold text-white">{stage.label}</h4>
                  <p className="mt-1 font-mono text-[10px] text-white/35">{stage.sub}</p>
                </motion.div>
                {i < 3 && (
                  <motion.div
                    initial={{ opacity: 0 }}
                    whileInView={{ opacity: 1 }}
                    viewport={{ once: true }}
                    transition={{ delay: i * 0.15 + 0.1 }}
                    className="text-white/20 text-[20px] hidden lg:block"
                  >
                    →
                  </motion.div>
                )}
              </div>
            ))}
          </div>
        </div>

        {/* Sanitizer rules */}
        <div className="mt-12 grid md:grid-cols-2 gap-6">
          {[
            { title: "Shell command sanitizer", rules: ["Allowlist of binaries", "Arg pattern validation", "No shell metacharacters"] },
            { title: "Python script sanitizer", rules: ["AST parse required", "Banned imports blocked", "No subprocess calls"] },
            { title: "File path sanitizer", rules: ["No ../ traversal", "No system paths", "Workspace-scoped"] },
            { title: "URL sanitizer", rules: ["SSRF blocklist", "No localhost / 169.254.x", "HTTPS required"] },
          ].map((sanitizer, i) => (
            <motion.div
              key={sanitizer.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.08 }}
              className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6"
            >
              <h3 className="font-body-mature text-[16px] font-semibold text-white mb-4">{sanitizer.title}</h3>
              <ul className="space-y-2">
                {sanitizer.rules.map((rule) => (
                  <li key={rule} className="flex items-center gap-2 font-body-mature text-[14px] text-white/45">
                    <span className="text-white/30">▸</span>
                    {rule}
                  </li>
                ))}
              </ul>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   6. CLOSING CTA
   ════════════════════════════════════════════════════════════ */

function ClosingCTA() {
  return (
    <section className="relative w-full py-[200px] px-6 border-t border-white/[0.08] overflow-hidden">
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <motion.div
          animate={{ scale: [1, 1.15, 1], opacity: [0.05, 0.1, 0.05] }}
          transition={{ duration: 5, repeat: Infinity }}
          className="w-[600px] h-[300px] rounded-full bg-white blur-[150px]"
        />
      </div>

      <div className="relative z-10 mx-auto max-w-3xl text-center">
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 1, ease: EASE_OUT }}
        >
          <h2 className="font-display text-[clamp(2rem,6vw,4rem)] font-semibold tracking-[-0.04em] leading-[1.05]">
            Your stack, orchestrated.
          </h2>
          <p className="mt-8 font-lead-airy mx-auto max-w-xl">
            Every CLI you installed, every API key you paid for, every local model you downloaded —
            one hotkey away.
          </p>
          <div className="mt-12 flex flex-col sm:flex-row items-center justify-center gap-4">
            <MagneticButton
              href="/download"
              className="mature-button rounded-full px-8 py-4 font-body-mature text-[15px] font-semibold"
            >
              Download v0.1.0 →
            </MagneticButton>
            <MagneticButton
              href="/orchestration"
              className="mature-button-secondary rounded-full px-6 py-4 font-body-mature text-[14px]"
            >
              See the engine
            </MagneticButton>
          </div>
        </motion.div>
      </div>
    </section>
  );
}
