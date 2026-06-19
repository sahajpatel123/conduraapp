"use client";

import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/**
 * OverlayPreview — the hero's right side, finally understandable.
 *
 * Not a terminal. The actual product UI a user will see: a floating chat
 * overlay with a real conversation cycling through relatable examples.
 * Each example shows a user message, which agent Condura routed it to,
 * and the response — so a first-time visitor instantly gets it:
 *
 *   "Oh. I press a hotkey. This window appears. I ask anything. It uses
 *    the AI I already have. And it's safe."
 *
 * Below the chat, four clear badges state what makes it different:
 * Free · Runs on your machine · Uses your AI subscriptions · Every action safety-checked
 *
 * Minimal. No glow, no aurora. Just the product, shown honestly.
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

const BADGES = [
  { label: "Free forever" },
  { label: "Runs on your machine" },
  { label: "Uses your AI subscriptions" },
  { label: "Every action safety-checked" },
];

export default function OverlayPreview({ active }: { active: boolean }) {
  const [idx, setIdx] = useState(0);

  useEffect(() => {
    if (!active) return;
    const timer = setInterval(() => setIdx((p) => (p + 1) % CONVERSATIONS.length), 5500);
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
        <div className="h-9 border-b border-white/[0.06] flex items-center px-4 bg-[#161616]">
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

      {/* ── Feature badges — what makes it different, in plain words ── */}
      <div className="grid grid-cols-2 gap-2">
        {BADGES.map((b, i) => (
          <motion.div
            key={b.label}
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 * i + 0.2, duration: 0.4, ease: EASE_OUT }}
            className="flex items-center gap-2.5 rounded-xl border border-white/[0.06] bg-white/[0.015] px-3 py-2.5"
          >
            <span className="flex h-5 w-5 shrink-0 items-center justify-center rounded-md border border-white/10 bg-white/[0.03]">
              <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2.5" strokeLinecap="round" strokeLinejoin="round" className="text-white/55">
                <path d="M5 12l5 5 9-10" />
              </svg>
            </span>
            <span className="font-body-mature text-[11.5px] text-white/65 leading-tight">{b.label}</span>
          </motion.div>
        ))}
      </div>
    </div>
  );
}