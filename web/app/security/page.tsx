"use client";

import { motion, AnimatePresence } from "motion/react";
import { useState, useEffect } from "react";
import AnimatedBadge from "@/components/motion/AnimatedBadge";
import TiltCard from "@/components/motion/TiltCard";
import { EASE_OUT } from "@/lib/motion";

/* ────────────────────────────────────────────────────────────
   SECURITY — The Gatekeeper
   Capability without the blank check. Condura draws a hard
   line between thinking and acting with a deterministic
   permission gatekeeper that no model can override.
   ──────────────────────────────────────────────────────────── */

const BLAST_RADII = [
  { class: "READ", desc: "Screenshot, copy text, inspect file", color: "#3b82f6", risk: "Low" },
  { class: "WRITE", desc: "Edit file, type text, paste content", color: "#f59e0b", risk: "Medium" },
  { class: "NETWORK", desc: "Click link, submit form, send message", color: "#ec4899", risk: "High" },
  { class: "DESTRUCTIVE", desc: "Delete, format, transfer, purchase", color: "#ef4444", risk: "Critical" },
];

const GATEKEEPER_FLOW = [
  { step: "01", title: "Strategist proposes", desc: "The model says: \"I need to click 'Send Email' in Gmail.\"" },
  { step: "02", title: "Blast radius classified", desc: "Action tagged as NETWORK on a messaging app. Risk: HIGH." },
  { step: "03", title: "Gatekeeper evaluates", desc: "Policy check against ~/.condura/policy.yaml. No matching allow rule." },
  { step: "04", title: "Consent required", desc: "Native macOS dialog blocks until you click. No silent execution." },
  { step: "05", title: "Audit logged", desc: "HMAC-chained entry written. Forever. Tamper-resistant." },
];

const KILL_SWITCHES = [
  { layer: "Layer 1", name: "Hard hotkey", desc: "Cmd+Shift+Escape kills the process instantly. Hardware-level." },
  { layer: "Layer 2", name: "Watchdog timer", desc: "If N seconds pass without verification, auto-pause. No model can stop it." },
  { layer: "Layer 3", name: "Network isolation", desc: "A separate OS process owns a pf/netsh rule blocking all egress. The agent cannot touch it." },
  { layer: "Layer 4", name: "Menu bar kill", desc: "One click in the system tray. Always visible, always available." },
];

const AUDIT_ENTRIES = [
  { ts: "14:32:01.234", action: "READ", target: "screenshot://primary", verdict: "allow", model: "claude-sonnet-4.5" },
  { ts: "14:32:02.891", action: "READ", target: "axtree://Safari", verdict: "allow", model: "claude-sonnet-4.5" },
  { ts: "14:32:04.102", action: "WRITE", target: "/Users/sahaj/Desktop/notes.txt", verdict: "allow", model: "codex" },
  { ts: "14:32:06.340", action: "NETWORK", target: "https://gmail.com/send", verdict: "require_consent", model: "claude-sonnet-4.5" },
  { ts: "14:32:08.000", action: "CONSENT", target: "native_dialog://gmail-send", verdict: "approved_by_user", model: "—" },
  { ts: "14:32:08.500", action: "NETWORK", target: "https://gmail.com/send", verdict: "executed", model: "claude-sonnet-4.5" },
  { ts: "14:32:09.100", action: "DESTRUCTIVE", target: "rm -rf /tmp/cache", verdict: "denied", model: "ollama" },
];

const VERDICT_COLORS: Record<string, string> = {
  allow: "#10a37f",
  require_consent: "#f59e0b",
  approved_by_user: "#3b82f6",
  executed: "#10a37f",
  denied: "#ef4444",
};

export default function SecurityPage() {
  return (
    <main className="relative w-full bg-black text-white overflow-hidden">
      <SecurityHero />
      <VaultSection />
      <BlastRadiusSection />
      <GatekeeperFlowSection />
      <KillSwitchSection />
      <AuditLogSection />
      <InvariantSummarySection />
      <ClosingCTA />
    </main>
  );
}

/* ════════════════════════════════════════════════════════════
   1. HERO
   ════════════════════════════════════════════════════════════ */

function SecurityHero() {
  const [mounted, setMounted] = useState(false);
  useEffect(() => { const t = setTimeout(() => setMounted(true), 200); return () => clearTimeout(t); }, []);

  return (
    <section className="relative min-h-screen flex flex-col items-center justify-center px-6 overflow-hidden">
      <div className="absolute inset-0 bg-grid-dark opacity-20" />

      {/* Shield glow */}
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <motion.div
          animate={{ scale: [1, 1.1, 1], opacity: [0.05, 0.1, 0.05] }}
          transition={{ duration: 4, repeat: Infinity }}
          className="w-[400px] h-[400px] rounded-full bg-white blur-[120px]"
        />
      </div>

      {/* Rotating shield rings */}
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 50, repeat: Infinity, ease: "linear" }}
          className="w-[500px] h-[500px] rounded-full border border-white/[0.06]"
        >
          {[0, 120, 240].map((deg) => (
            <div
              key={deg}
              style={{ transform: `rotate(${deg}deg)` }}
              className="absolute top-0 left-1/2 w-1.5 h-1.5 -translate-x-1/2 rounded-full bg-white/20"
            />
          ))}
        </motion.div>
        <motion.div
          animate={{ rotate: -360 }}
          transition={{ duration: 30, repeat: Infinity, ease: "linear" }}
          className="absolute w-[350px] h-[350px] rounded-full border border-dashed border-white/[0.08]"
        />
      </div>

      <div className="relative z-10 max-w-4xl text-center">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: mounted ? 1 : 0, y: mounted ? 0 : 20 }}
          transition={{ duration: 1, ease: EASE_OUT }}
        >
          <div className="mb-8 flex justify-center">
            <AnimatedBadge tone="neutral" pulse>Zero Trust</AnimatedBadge>
          </div>

          <h1 className="font-display text-[clamp(2.5rem,7vw,5rem)] font-semibold leading-[1.05] tracking-[-0.04em]">
            Capability without
            <br />
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-white via-white to-white/30">
              the blank check.
            </span>
          </h1>

          <p className="mt-8 mx-auto max-w-2xl font-lead-airy">
            Models hallucinate. Prompts get injected. Some actions cannot be undone. Condura draws
            a hard line between thinking and acting — a deterministic permission gatekeeper that no
            model output can bypass, override, or shortcut.
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: mounted ? 1 : 0 }}
          transition={{ delay: 0.8, duration: 1 }}
          className="mt-12 flex flex-wrap items-center justify-center gap-6"
        >
          {[
            { label: "Safety modules", value: "7" },
            { label: "Kill switches", value: "4" },
            { label: "Model bypass", value: "0" },
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
   2. VAULT — Interactive Lock/Unlock
   ════════════════════════════════════════════════════════════ */

function VaultSection() {
  const [unlocked, setUnlocked] = useState(false);

  return (
    <section className="relative w-full py-[160px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-20 max-w-3xl">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">The Gatekeeper</span>
          <h2 className="mt-4 font-display text-[clamp(2rem,5vw,3.5rem)] font-semibold tracking-[-0.03em] leading-[1.1]">
            The strategist decides what.
            <br />
            The gatekeeper decides if.
          </h2>
          <p className="mt-6 font-lead-airy">
            Two separate systems. The strategist is a model — fallible, promptable, probabilistic.
            The gatekeeper is deterministic Go code — static rules, no neural nets, no prompt
            injection surface. They are never the same, never merged, never shortcut.
          </p>
        </div>

        {/* Interactive vault */}
        <div
          className="relative mx-auto flex h-[500px] w-full max-w-2xl items-center justify-center overflow-hidden rounded-3xl border border-white/10 bg-[#050505] cursor-pointer"
          onMouseEnter={() => setUnlocked(true)}
          onMouseLeave={() => setUnlocked(false)}
          onClick={() => setUnlocked(!unlocked)}
        >
          {/* Background glow */}
          <motion.div
            animate={{ opacity: unlocked ? 0.3 : 0.08, scale: unlocked ? 1.3 : 1 }}
            transition={{ duration: 1 }}
            className="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(255,255,255,0.15),transparent_50%)]"
          />

          {/* Outer rotating ring */}
          <motion.div
            animate={{ rotate: unlocked ? 180 : 0 }}
            transition={{ type: "spring", stiffness: 50, damping: 20 }}
            className="absolute h-[380px] w-[380px] rounded-full border-2 border-dashed border-white/15"
          />

          {/* Middle rotating ring */}
          <motion.div
            animate={{ rotate: unlocked ? -90 : 0 }}
            transition={{ type: "spring", stiffness: 60, damping: 25 }}
            className="absolute h-[280px] w-[280px] rounded-full border border-white/10 flex items-center justify-center"
          >
            {/* Locking pins */}
            {[0, 90, 180, 270].map((deg) => (
              <motion.div
                key={deg}
                style={{ rotate: deg }}
                className="absolute flex h-full w-full justify-between"
              >
                <motion.div
                  animate={{ scaleX: unlocked ? 0 : 1 }}
                  transition={{ duration: 0.3 }}
                  className="h-[2px] w-5 bg-white/40 origin-left"
                />
                <motion.div
                  animate={{ scaleX: unlocked ? 0 : 1 }}
                  transition={{ duration: 0.3 }}
                  className="h-[2px] w-5 bg-white/40 origin-right"
                />
              </motion.div>
            ))}
          </motion.div>

          {/* Core vault */}
          <motion.div
            animate={{
              scale: unlocked ? 1.05 : 1,
              boxShadow: unlocked
                ? "0 0 80px rgba(255,255,255,0.12), inset 0 0 20px rgba(255,255,255,0.05)"
                : "0 0 0px rgba(255,255,255,0), inset 0 0 0px rgba(255,255,255,0)",
            }}
            transition={{ duration: 0.5 }}
            className="relative flex h-44 w-44 flex-col items-center justify-center rounded-full border border-white/20 bg-black z-20"
          >
            <span className="font-mono text-[10px] uppercase tracking-widest text-white/40 mb-2">Status</span>
            <motion.span
              animate={{ color: unlocked ? "rgba(255,255,255,1)" : "rgba(255,255,255,0.4)" }}
              className="text-xl font-medium tracking-tight"
            >
              {unlocked ? "UNLOCKED" : "LOCKED"}
            </motion.span>

            <AnimatePresence>
              {unlocked && (
                <motion.div
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -10 }}
                  className="absolute -bottom-20 flex gap-2"
                >
                  <AnimatedBadge tone="neutral">Read FS</AnimatedBadge>
                  <AnimatedBadge tone="neutral">Port 3000</AnimatedBadge>
                </motion.div>
              )}
            </AnimatePresence>
          </motion.div>

          {/* Hint */}
          <div className="absolute bottom-6 left-1/2 -translate-x-1/2 font-mono text-[10px] text-white/20">
            {unlocked ? "hover to re-lock" : "hover to unlock"}
          </div>
        </div>

        {/* Separation explainer */}
        <div className="mt-12 grid md:grid-cols-2 gap-6">
          <motion.div
            initial={{ opacity: 0, x: -20 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6"
          >
            <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Strategist</span>
            <h3 className="mt-3 font-body-mature text-[16px] font-semibold text-white">A model. Fallible.</h3>
            <p className="mt-2 font-body-mature text-[14px] text-white/45 leading-relaxed">
              Any LLM. Proposes actions. Can be wrong. Can be tricked. Can hallucinate. That&apos;s
              fine — it&apos;s not allowed to touch the keyboard.
            </p>
          </motion.div>
          <motion.div
            initial={{ opacity: 0, x: 20 }}
            whileInView={{ opacity: 1, x: 0 }}
            viewport={{ once: true }}
            className="rounded-2xl border border-white/[0.10] bg-white/[0.03] p-6"
          >
            <span className="font-mono text-[11px] uppercase tracking-widest text-white/40">Gatekeeper</span>
            <h3 className="mt-3 font-body-mature text-[16px] font-semibold text-white">Pure Go. Deterministic.</h3>
            <p className="mt-2 font-body-mature text-[14px] text-white/50 leading-relaxed">
              Static YAML policy. No neural nets. No prompts. No &ldquo;trust me, the model said
              it&apos;s safe.&rdquo; The only path to a click, keystroke, or shell exec.
            </p>
          </motion.div>
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   3. BLAST RADIUS — Action Classification
   ════════════════════════════════════════════════════════════ */

function BlastRadiusSection() {
  return (
    <section className="relative w-full py-[160px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-20 max-w-3xl">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Blast Radius</span>
          <h2 className="mt-4 font-display text-[clamp(2rem,5vw,3.5rem)] font-semibold tracking-[-0.03em] leading-[1.1]">
            Every action has a radius.
          </h2>
          <p className="mt-6 font-lead-airy">
            Before any action reaches the gatekeeper, it is classified by blast radius. The
            classification is deterministic and instant — no model involvement, no ambiguity.
            Higher radius means more friction. Destructive actions require a real human at the keyboard.
          </p>
        </div>

        {/* Blast radius cards */}
        <div className="space-y-4">
          {BLAST_RADII.map((radius, i) => (
            <motion.div
              key={radius.class}
              initial={{ opacity: 0, x: -20 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.08 }}
              className="group relative overflow-hidden rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6"
            >
              {/* Risk bar on left */}
              <div
                className="absolute left-0 top-0 bottom-0 w-1"
                style={{ background: radius.color }}
              />

              <div className="flex items-center justify-between gap-6 pl-4">
                <div className="flex items-center gap-4">
                  <div
                    className="flex h-12 w-12 items-center justify-center rounded-xl border"
                    style={{ borderColor: `${radius.color}40`, background: `${radius.color}10` }}
                  >
                    <span className="font-mono text-[12px]" style={{ color: radius.color }}>
                      {radius.risk[0]}
                    </span>
                  </div>
                  <div>
                    <h3 className="font-body-mature text-[18px] font-semibold text-white">
                      {radius.class}
                    </h3>
                    <p className="font-body-mature text-[14px] text-white/45">{radius.desc}</p>
                  </div>
                </div>

                <div className="flex items-center gap-3">
                  <span
                    className="rounded-full border px-3 py-1 font-mono text-[11px]"
                    style={{ borderColor: `${radius.color}30`, color: radius.color, background: `${radius.color}08` }}
                  >
                    {radius.risk}
                  </span>
                  {/* Risk meter */}
                  <div className="hidden md:flex gap-1">
                    {[0, 1, 2, 3].map((idx) => (
                      <div
                        key={idx}
                        className="h-4 w-1.5 rounded-full"
                        style={{
                          background: idx <= i ? radius.color : "rgba(255,255,255,0.08)",
                        }}
                      />
                    ))}
                  </div>
                </div>
              </div>
            </motion.div>
          ))}
        </div>

        {/* Policy example */}
        <div className="mt-12 overflow-hidden rounded-2xl border border-white/[0.10] bg-[#0a0a0a]">
          <div className="flex items-center gap-2 border-b border-white/[0.06] bg-white/[0.02] px-5 py-3">
            <span className="font-mono text-[11px] text-white/30">~/.condura/policy.yaml</span>
          </div>
          <pre className="p-6 font-mono text-[12px] leading-relaxed text-white/50 overflow-x-auto">
{`rules:
  - match: { class: READ }
    decide: allow

  - match: { class: WRITE, target_app: ["Code", "Terminal"] }
    decide: allow

  - match: { class: NETWORK }
    decide: require_consent
    consent:
      type: native_dialog
      timeout_seconds: 300
      on_timeout: queue

  - match: { class: DESTRUCTIVE }
    decide: require_presence_and_consent
    consent:
      type: native_dialog
      require_user_active: true
      on_user_absent: queue

  - match: { target_app: ["1Password", "Keychain Access"] }
    decide: deny`}
          </pre>
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   4. GATEKEEPER FLOW — Decision Pipeline
   ════════════════════════════════════════════════════════════ */

function GatekeeperFlowSection() {
  return (
    <section className="relative w-full py-[160px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-20 max-w-3xl">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Decision Flow</span>
          <h2 className="mt-4 font-display text-[clamp(2rem,5vw,3.5rem)] font-semibold tracking-[-0.03em] leading-[1.1]">
            Five steps. Zero shortcuts.
          </h2>
          <p className="mt-6 font-lead-airy">
            From the moment a model proposes an action to the moment it executes (or doesn&apos;t),
            the pipeline is fully deterministic, fully logged, and fully auditable.
          </p>
        </div>

        {/* Flow steps */}
        <div className="space-y-4">
          {GATEKEEPER_FLOW.map((step, i) => (
            <motion.div
              key={step.step}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-50px" }}
              transition={{ delay: i * 0.08 }}
              className="flex items-start gap-6"
            >
              {/* Step number */}
              <div className="relative flex h-12 w-12 shrink-0 items-center justify-center rounded-xl border border-white/15 bg-white/[0.03]">
                <span className="font-mono text-[13px] text-white/50">{step.step}</span>
                {/* Connector line */}
                {i < GATEKEEPER_FLOW.length - 1 && (
                  <div className="absolute top-full left-1/2 h-4 w-[1px] -translate-x-1/2 bg-white/10" />
                )}
              </div>

              {/* Content */}
              <div className="flex-1 rounded-2xl border border-white/[0.08] bg-white/[0.02] p-5">
                <h3 className="font-body-mature text-[16px] font-semibold text-white">{step.title}</h3>
                <p className="mt-2 font-body-mature text-[14px] text-white/45 leading-relaxed">{step.desc}</p>
              </div>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   5. KILL SWITCH — Four Independent Mechanisms
   ════════════════════════════════════════════════════════════ */

function KillSwitchSection() {
  return (
    <section className="relative w-full py-[160px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-20 max-w-3xl">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Kill Switch</span>
          <h2 className="mt-4 font-display text-[clamp(2rem,5vw,3.5rem)] font-semibold tracking-[-0.03em] leading-[1.1]">
            You can always stop it.
          </h2>
          <p className="mt-6 font-lead-airy">
            Four independent mechanisms. The agent cannot disable any of them. Not by prompt, not
            by code, not by &ldquo;I&apos;m pretty sure this is safe.&rdquo; If one fails, the other
            three are still there.
          </p>
        </div>

        <div className="grid md:grid-cols-2 gap-6">
          {KILL_SWITCHES.map((kill, i) => (
            <TiltCard key={kill.layer} maxRotate={5} className="h-full">
              <motion.div
                initial={{ opacity: 0, scale: 0.95 }}
                whileInView={{ opacity: 1, scale: 1 }}
                viewport={{ once: true }}
                transition={{ delay: i * 0.08, type: "spring", stiffness: 100, damping: 15 }}
                className="relative h-full overflow-hidden rounded-2xl border border-white/[0.10] bg-white/[0.02] p-6 backdrop-blur-md"
              >
                {/* Red accent */}
                <div className="absolute top-0 left-0 right-0 h-[2px] bg-red-500/30" />

                <div className="flex items-center justify-between mb-4">
                  <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">{kill.layer}</span>
                  <motion.div
                    animate={{ scale: [1, 1.2, 1] }}
                    transition={{ duration: 2, repeat: Infinity, delay: i * 0.3 }}
                    className="flex h-3 w-3 items-center justify-center rounded-full border border-red-400/30"
                  >
                    <span className="h-1 w-1 rounded-full bg-red-400/50" />
                  </motion.div>
                </div>

                <h3 className="font-body-mature text-[18px] font-semibold text-white">{kill.name}</h3>
                <p className="mt-2 font-body-mature text-[14px] text-white/45 leading-relaxed">{kill.desc}</p>
              </motion.div>
            </TiltCard>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   6. AUDIT LOG — HMAC-Chained, Append-Only
   ════════════════════════════════════════════════════════════ */

function AuditLogSection() {
  const [visibleEntries, setVisibleEntries] = useState(0);

  useEffect(() => {
    if (visibleEntries >= AUDIT_ENTRIES.length) {
      const reset = setTimeout(() => setVisibleEntries(0), 5000);
      return () => clearTimeout(reset);
    }
    const t = setTimeout(() => setVisibleEntries((p) => p + 1), 700);
    return () => clearTimeout(t);
  }, [visibleEntries]);

  return (
    <section className="relative w-full py-[160px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-20 max-w-3xl">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Audit Trail</span>
          <h2 className="mt-4 font-display text-[clamp(2rem,5vw,3.5rem)] font-semibold tracking-[-0.03em] leading-[1.1]">
            Everything is logged. Forever.
          </h2>
          <p className="mt-6 font-lead-airy">
            Every screenshot, every accessibility tree dump, every model decision, every API call,
            every click coordinate — written to an HMAC-chained, append-only log. Not for debugging.
            For forensics. If something goes wrong, you can prove exactly what happened.
          </p>
        </div>

        {/* Audit log terminal */}
        <div className="overflow-hidden rounded-2xl border border-white/[0.10] bg-[#0a0a0a] shadow-[0_40px_80px_rgba(0,0,0,0.5)]">
          <div className="flex items-center justify-between border-b border-white/[0.06] bg-white/[0.02] px-5 py-3">
            <div className="flex items-center gap-2">
              <div className="w-2.5 h-2.5 rounded-full bg-[#ff5f57]" />
              <div className="w-2.5 h-2.5 rounded-full bg-[#febc2e]" />
              <div className="w-2.5 h-2.5 rounded-full bg-[#28c840]" />
            </div>
            <span className="font-mono text-[11px] text-white/30">audit.log — HMAC chain verified ✓</span>
            <span className="flex items-center gap-1.5 font-mono text-[10px] text-white/30">
              <motion.span
                animate={{ opacity: [1, 0.3, 1] }}
                transition={{ duration: 1.5, repeat: Infinity }}
                className="w-1.5 h-1.5 rounded-full bg-green-400/60"
              />
              live
            </span>
          </div>

          {/* Log entries */}
          <div className="p-6 min-h-[360px] font-mono text-[12px]">
            {/* Header */}
            <div className="flex items-center gap-4 pb-3 mb-3 border-b border-white/[0.06] text-white/25 text-[10px] uppercase tracking-wider">
              <span className="w-24">Timestamp</span>
              <span className="w-20">Action</span>
              <span className="flex-1">Target</span>
              <span className="w-32">Verdict</span>
              <span className="w-32">Model</span>
            </div>

            <AnimatePresence>
              {AUDIT_ENTRIES.slice(0, visibleEntries).map((entry, i) => (
                <motion.div
                  key={i}
                  initial={{ opacity: 0, y: 6 }}
                  animate={{ opacity: 1, y: 0 }}
                  className="flex items-center gap-4 py-1.5"
                >
                  <span className="w-24 text-white/30">{entry.ts}</span>
                  <span className="w-20 text-white/50">{entry.action}</span>
                  <span className="flex-1 truncate text-white/45">{entry.target}</span>
                  <span
                    className="w-32 rounded px-1.5 py-0.5 text-[10px]"
                    style={{ color: VERDICT_COLORS[entry.verdict], background: `${VERDICT_COLORS[entry.verdict]}15` }}
                  >
                    {entry.verdict}
                  </span>
                  <span className="w-32 text-white/30">{entry.model}</span>
                </motion.div>
              ))}
            </AnimatePresence>

            {visibleEntries < AUDIT_ENTRIES.length && (
              <div className="flex items-center gap-2 pt-2">
                <span className="text-white/30">❯</span>
                <motion.span
                  animate={{ opacity: [1, 0] }}
                  transition={{ repeat: Infinity, duration: 0.9 }}
                  className="inline-block w-[7px] h-[13px] bg-white/30"
                />
              </div>
            )}
          </div>
        </div>

        {/* Audit properties */}
        <div className="mt-8 grid md:grid-cols-3 gap-6">
          {[
            { title: "HMAC-chained", desc: "Each entry includes a hash of the previous. Tampering breaks the chain visibly." },
            { title: "Append-only", desc: "Entries are never deleted, never edited. 90-day retention, configurable." },
            { title: "Secret redaction", desc: "API keys, tokens, and passwords are stripped before logging. Always." },
          ].map((prop, i) => (
            <motion.div
              key={prop.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.08 }}
              className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-5"
            >
              <h3 className="font-body-mature text-[15px] font-semibold text-white">{prop.title}</h3>
              <p className="mt-2 font-body-mature text-[13px] text-white/45 leading-relaxed">{prop.desc}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   7. INVARIANT SUMMARY — The Seven Rules
   ════════════════════════════════════════════════════════════ */

function InvariantSummarySection() {
  const invariants = [
    "The Strategist and the Gatekeeper are separate systems.",
    "The Gatekeeper is the only path to physical action.",
    "Destructive actions require a real human at the keyboard.",
    "You can always stop the agent.",
    "Every action is auditable, in a tamper-resistant log.",
    "The agent is a guest, not an owner.",
    "OS permissions are granted by you, on your machine.",
  ];

  return (
    <section className="relative w-full py-[160px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-16 text-center">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">The Seven Invariants</span>
          <h2 className="mt-4 font-display text-[clamp(2rem,5vw,3.5rem)] font-semibold tracking-[-0.03em] leading-[1.1]">
            If a feature conflicts, the feature is wrong.
          </h2>
        </div>

        <div className="space-y-3">
          {invariants.map((inv, i) => (
            <motion.div
              key={i}
              initial={{ opacity: 0, x: -20 }}
              whileInView={{ opacity: 1, x: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.06 }}
              className="group flex items-center gap-6 rounded-2xl border border-white/[0.06] bg-white/[0.01] p-5 transition-colors hover:bg-white/[0.03]"
            >
              <span className="font-mono text-[22px] font-light text-white/30 tabular-nums w-8">
                {String(i + 1).padStart(2, "0")}
              </span>
              <p className="flex-1 font-body-mature text-[16px] text-white/60 group-hover:text-white/80 transition-colors">
                {inv}
              </p>
              <span className="font-mono text-[10px] uppercase tracking-widest text-white/20">
                non-negotiable
              </span>
            </motion.div>
          ))}
        </div>

        <div className="mt-12 text-center">
          <a
            href="/manifesto"
            className="font-body-mature text-[14px] text-white/50 underline decoration-white/20 underline-offset-4 hover:text-white hover:decoration-white/50 transition-colors"
          >
            Read the full manifesto →
          </a>
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   8. CLOSING CTA
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
            An agent safe enough to hand to anyone.
          </h2>
          <p className="mt-8 font-lead-airy mx-auto max-w-xl">
            Download Condura. Set your hotkey. The gatekeeper is sealed by default — you open it,
            you decide how far, you can always close it.
          </p>
          <div className="mt-12 flex flex-col sm:flex-row items-center justify-center gap-4">
            <a
              href="/download"
              className="mature-button inline-flex items-center gap-2 px-8 py-4 font-body-mature text-[15px] font-semibold"
            >
              Download v0.1.0 →
            </a>
            <a
              href="/manifesto"
              className="mature-button-secondary inline-flex items-center gap-2 px-6 py-4 font-body-mature text-[14px]"
            >
              Read the manifesto
            </a>
          </div>
        </motion.div>
      </div>
    </section>
  );
}
