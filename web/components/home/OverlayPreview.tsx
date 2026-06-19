"use client";

import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/**
 * OverlayPreview — the hero's right side: the product, shown honestly.
 *
 * Two layers, both auto-running, both useful:
 *
 *   1. The chat overlay — the actual product UI a user sees. A real
 *      conversation cycles through relatable day-to-day examples, each
 *      showing the user message, which agent it routed to, and the
 *      response. A visitor instantly gets it: press a hotkey, this
 *      window appears, ask anything, it uses the AI you already have,
 *      and it's safe.
 *
 *   2. "Your first 60 seconds" — an auto-stepping timeline beneath the
 *      chat showing the real onboarding journey: Install → Set hotkey →
 *      Grant access → First task → Done. Each step has a tiny visual.
 *      This closes the gap between "I see what it does" and "I can
 *      picture myself doing it" — the last mile before download.
 *
 * Minimal. No glow, no aurora. Just the product and the path to it.
 */

type Conv = {
  user: string;
  agent: string;
  agentKind: "cloud" | "local";
  response: string;
  consent?: string;
};

const CONVERSATIONS: Conv[] = [
  {
    user: "Plan my week based on today's calendar",
    agent: "Gemini",
    agentKind: "cloud",
    response: "5 meetings Mon–Wed. I blocked deep-work on Thu and moved the 3pm Friday to 2pm. Want me to send the updates?",
  },
  {
    user: "Draft a reply to Sarah's email",
    agent: "Claude",
    agentKind: "cloud",
    response: "Drafted. It acknowledges her timeline, proposes Thursday, and keeps it under 120 words. Ready to send?",
    consent: "Send Email · Gmail",
  },
  {
    user: "Summarize this PDF I'm reading",
    agent: "Ollama",
    agentKind: "local",
    response: "12-page lease. Key terms: 2yr, $3,200/mo, tenant pays utilities. Renewal window opens Oct 1. No sublet clause.",
  },
  {
    user: "Fix the bug in my auth.ts file",
    agent: "Codex",
    agentKind: "cloud",
    response: "The token compare used == instead of constant-time. Patched to crypto.subtle.timingSafeEqual. 2 tests added.",
  },
];

const FIRST_60: { step: string; title: string; sub: string }[] = [
  { step: "00s", title: "Install", sub: "Drag to Applications. 11 MB." },
  { step: "15s", title: "Set hotkey", sub: "Press any combo. It's yours." },
  { step: "25s", title: "Grant access", sub: "Accessibility + screen. Reversible." },
  { step: "35s", title: "First task", sub: "Press your hotkey. Ask anything." },
  { step: "60s", title: "Done", sub: "It's running. Your keys. Your machine." },
];

export default function OverlayPreview({ active }: { active: boolean }) {
  const [idx, setIdx] = useState(0);
  const [step, setStep] = useState(0);

  // Cycle chat examples
  useEffect(() => {
    if (!active) return;
    const timer = setInterval(() => setIdx((p) => (p + 1) % CONVERSATIONS.length), 5500);
    return () => clearInterval(timer);
  }, [active]);

  // Cycle first-60 steps (slightly slower, independent of chat)
  useEffect(() => {
    if (!active) return;
    const timer = setInterval(() => setStep((p) => (p + 1) % FIRST_60.length), 2600);
    return () => clearInterval(timer);
  }, [active]);

  const conv = CONVERSATIONS[idx];

  return (
    <div className="flex flex-col gap-5">
      {/* ── The overlay window — the actual product UI ── */}
      <div className="relative rounded-2xl border border-white/[0.10] bg-[#0f0f0f]/95 backdrop-blur-xl shadow-[0_40px_100px_rgba(0,0,0,0.6),0_0_0_1px_rgba(255,255,255,0.04)] overflow-hidden">
        {/* Top hairline */}
        <div className="pointer-events-none absolute inset-x-0 top-0 h-px bg-gradient-to-r from-transparent via-white/20 to-transparent" />

        {/* Title bar */}
        <div className="h-9 border-b border-white/[0.06] flex items-center px-4 bg-[#161616] relative">
          <div className="flex gap-1.5">
            <span className="w-2.5 h-2.5 rounded-full bg-white/15" />
            <span className="w-2.5 h-2.5 rounded-full bg-white/15" />
            <span className="w-2.5 h-2.5 rounded-full bg-white/15" />
          </div>
          <div className="absolute left-1/2 -translate-x-1/2 flex items-center gap-2">
            <span className="w-1.5 h-1.5 rounded-full bg-green-400/60" />
            <span className="text-[10px] text-white/30 font-mono">Condura</span>
          </div>
        </div>

        {/* Chat body */}
        <div className="p-5 min-h-[220px] flex flex-col gap-4">
          <AnimatePresence mode="wait">
            <motion.div
              key={`conv-${idx}`}
              initial={{ opacity: 0, y: 10 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -10 }}
              transition={{ duration: 0.4, ease: EASE_OUT }}
              className="flex flex-col gap-4"
            >
              {/* User message — right-aligned bubble */}
              <div className="flex justify-end">
                <div className="max-w-[80%] rounded-2xl rounded-tr-sm bg-white/[0.08] px-3.5 py-2.5">
                  <p className="font-body-mature text-[13px] text-white/90 leading-snug">
                    {conv.user}
                  </p>
                </div>
              </div>

              {/* Routing badge */}
              <div className="flex items-center gap-2 pl-1">
                <span className={`flex items-center gap-1.5 rounded-full border px-2 py-0.5 ${
                  conv.agentKind === "local"
                    ? "border-violet-400/20 bg-violet-400/10"
                    : "border-sky-400/20 bg-sky-400/10"
                }`}>
                  <span className={`w-1.5 h-1.5 rounded-full ${
                    conv.agentKind === "local" ? "bg-violet-400/70" : "bg-sky-400/70"
                  }`} />
                  <span className={`font-mono text-[10px] ${
                    conv.agentKind === "local" ? "text-violet-400/80" : "text-sky-400/80"
                  }`}>
                    {conv.agentKind === "local" ? "local · " : "routed · "}{conv.agent}
                  </span>
                </span>
                {conv.consent && (
                  <span className="flex items-center gap-1 rounded-full border border-amber-400/20 bg-amber-400/10 px-2 py-0.5">
                    <span className="w-1.5 h-1.5 rounded-full bg-amber-400/70" />
                    <span className="font-mono text-[10px] text-amber-400/80">
                      ask: {conv.consent}
                    </span>
                  </span>
                )}
              </div>

              {/* Agent response — left-aligned with a small avatar */}
              <div className="flex justify-start gap-2.5">
                <div className="w-7 h-7 shrink-0 rounded-lg border border-white/10 bg-white/[0.04] flex items-center justify-center mt-0.5">
                  <span className="text-[11px] font-semibold text-white/60">C</span>
                </div>
                <div className="max-w-[82%]">
                  <p className="font-body-mature text-[13px] text-white/75 leading-relaxed">
                    {conv.response}
                  </p>
                </div>
              </div>
            </motion.div>
          </AnimatePresence>

          {/* Input bar at the bottom — the persistent "ask anything" line */}
          <div className="mt-auto pt-3 border-t border-white/[0.05]">
            <div className="flex items-center gap-2.5 rounded-xl border border-white/[0.08] bg-white/[0.02] px-3.5 py-2.5">
              <span className="text-[12px] text-white/30">Ask anything…</span>
              <span className="ml-auto text-[10px] text-white/20 font-mono border border-white/[0.06] rounded px-1.5 py-0.5">
                ⌘⇧Space
              </span>
            </div>
          </div>
        </div>
      </div>

      {/* ── "Your first 60 seconds" — the onboarding journey, auto-stepping ── */}
      <div className="rounded-2xl border border-white/[0.06] bg-white/[0.015] p-4">
        <div className="flex items-center justify-between mb-3.5 px-1">
          <span className="font-mono text-[10px] uppercase tracking-[0.2em] text-white/35">
            Your first 60 seconds
          </span>
          <span className="font-mono text-[10px] text-white/25">
            {FIRST_60[step].step}
          </span>
        </div>

        {/* Step rail — 5 connected segments */}
        <div className="flex items-center gap-1.5 mb-4 px-1">
          {FIRST_60.map((s, i) => {
            const isActive = i === step;
            const isDone = i < step;
            return (
              <div
                key={s.step}
                className="relative h-1 flex-1 rounded-full overflow-hidden bg-white/[0.06]"
              >
                <motion.div
                  initial={false}
                  animate={{ width: isDone ? "100%" : isActive ? "100%" : "0%" }}
                  transition={{ duration: isActive ? 2.6 : 0.3, ease: "linear" }}
                  className={`absolute inset-y-0 left-0 rounded-full ${isDone ? "bg-white/35" : "bg-white/55"}`}
                />
              </div>
            );
          })}
        </div>

        {/* Active step content — tiny visual + title + sub */}
        <div className="px-1 min-h-[58px]">
          <AnimatePresence mode="wait">
            <motion.div
              key={`step-${step}`}
              initial={{ opacity: 0, y: 6 }}
              animate={{ opacity: 1, y: 0 }}
              exit={{ opacity: 0, y: -6 }}
              transition={{ duration: 0.35, ease: EASE_OUT }}
              className="flex items-center gap-3.5"
            >
              <StepVisual step={step} />
              <div className="flex flex-col">
                <span className="font-body-mature text-[13px] font-semibold text-white leading-tight">
                  {FIRST_60[step].title}
                </span>
                <span className="font-body-mature text-[11.5px] text-white/45 leading-tight mt-0.5">
                  {FIRST_60[step].sub}
                </span>
              </div>
            </motion.div>
          </AnimatePresence>
        </div>
      </div>
    </div>
  );
}

/* ── Per-step minimal visuals ── */

function StepVisual({ step }: { step: number }) {
  const cls = "shrink-0 flex h-9 w-9 items-center justify-center rounded-lg border border-white/10 bg-white/[0.03] text-white/60";

  if (step === 0) {
    // Install — a download arrow into a tray
    return (
      <div className={cls}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M12 3v12" />
          <path d="M7 10l5 5 5-5" />
          <path d="M5 21h14" />
        </svg>
      </div>
    );
  }
  if (step === 1) {
    // Set hotkey — a key cap
    return (
      <div className={cls}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <rect x="3" y="7" width="18" height="11" rx="2.5" />
          <path d="M7 11h2M11 11h2M15 11h2" />
        </svg>
      </div>
    );
  }
  if (step === 2) {
    // Grant access — a shield with a check
    return (
      <div className={cls}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M12 3l8 3v6c0 5-3.5 8-8 9-4.5-1-8-4-8-9V6l8-3z" />
          <path d="M9 12l2 2 4-4" />
        </svg>
      </div>
    );
  }
  if (step === 3) {
    // First task — a chat bubble
    return (
      <div className={cls}>
        <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
          <path d="M21 12a8 8 0 0 1-11.5 7.2L4 21l1.8-5.5A8 8 0 1 1 21 12z" />
          <path d="M8 11h8M8 14h5" />
        </svg>
      </div>
    );
  }
  // Done — a circle with a check
  return (
    <div className={cls}>
      <svg width="16" height="16" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round">
        <circle cx="12" cy="12" r="9" />
        <path d="M8 12l3 3 5-5" />
      </svg>
    </div>
  );
}