"use client";

import { motion } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/* ────────────────────────────────────────────────────────────
   BringYourOwnAI — replaces the empty ProviderMarquee.

   The old section was an infinite scroll of company names: pretty,
   useless. A visitor learned nothing. This replacement is a creative,
   scannable layout that shows the actual value of multi-provider
   routing: Condura uses the AI you already pay for or run locally,
   so you don't pay again.

   Three lanes, grouped by *how you connect* — the question a real
   user has — not by vendor alphabet:

     · Subscriptions you already pay for  (no new bill)
     · API keys                          (pay per use, your keys)
     · Runs on your machine              (free, offline, private)

   Each lane lists what you get. A header carries the key insight.
   No cards-in-a-row, no marquee. A structured, useful grid.
   ──────────────────────────────────────────────────────────── */

type Lane = {
  tag: string;
  tagTone: "amber" | "violet";
  headline: string;
  sub: string;
  items: { name: string; note: string }[];
};

const LANES: Lane[] = [
  {
    tag: "API keys",
    tagTone: "amber",
    headline: "Or pay per use, with your own keys",
    sub: "Bring keys from any provider. Spend caps, failover, and full audit logs are built in — your keys never leave your machine and never touch our servers.",
    items: [
      { name: "Anthropic", note: "Opus · Sonnet · Haiku" },
      { name: "OpenAI", note: "GPT-5.5 · o3 · o4-mini" },
      { name: "Google", note: "Gemini 3.5 Flash · 3.1 Pro" },
      { name: "xAI · Mistral · DeepSeek", note: "plus 4 more" },
    ],
  },
  {
    tag: "Runs on your machine",
    tagTone: "violet",
    headline: "Or run fully local, free, and offline",
    sub: "Point Condura at Ollama, LM Studio, or llama.cpp and it works with zero network, zero cost, zero tracking. Your data never leaves the device.",
    items: [
      { name: "Ollama", note: "Llama, Qwen, Mistral" },
      { name: "LM Studio", note: "any GGUF model" },
      { name: "llama.cpp", note: "raw local server" },
      { name: "vLLM", note: "self-hosted inference" },
    ],
  },
];

const TONE: Record<Lane["tagTone"], { dot: string; text: string; border: string; bg: string }> = {
  amber: {
    dot: "bg-amber-400/70",
    text: "text-amber-400/80",
    border: "border-amber-400/20",
    bg: "bg-amber-400/10",
  },
  violet: {
    dot: "bg-violet-400/70",
    text: "text-violet-400/80",
    border: "border-violet-400/20",
    bg: "bg-violet-400/10",
  },
};

export default function BringYourOwnAI() {
  return (
    <section
      aria-label="Bring your own AI — compatible providers and runtimes"
      className="relative w-full border-t border-white/[0.08] bg-black py-24 px-6"
    >
      <div className="mx-auto max-w-6xl">
        {/* ── Header ── */}
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-60px" }}
          transition={{ duration: 0.7, ease: EASE_OUT }}
          className="mb-16 max-w-2xl"
        >
          <div className="flex items-center gap-2.5 mb-5">
            <span className="h-px w-8 bg-white/20" />
            <span className="font-mono text-[11px] uppercase tracking-[0.25em] text-white/35">
              Bring your own AI
            </span>
          </div>
          <h2 className="font-display text-[clamp(1.75rem,3.6vw,2.75rem)] font-semibold leading-[1.1] tracking-[-0.03em] text-white">
            One agent.{" "}
            <span className="text-white/40">Every model you already have.</span>
          </h2>
          <p className="mt-5 font-body-mature text-[15px] text-white/45 leading-relaxed max-w-xl">
            Condura doesn't sell you a model. It drives the ones you already
            have — or the ones running free on your own machine. Bring your own
            API keys. No second account. No vendor lock-in.
          </p>
        </motion.div>

        {/* ── Two lanes ── */}
        <div className="grid md:grid-cols-2 gap-5">
          {LANES.map((lane, li) => {
            const tone = TONE[lane.tagTone];
            return (
              <motion.div
                key={lane.tag}
                initial={{ opacity: 0, y: 20 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true, margin: "-40px" }}
                transition={{ delay: li * 0.1, duration: 0.6, ease: EASE_OUT }}
                className="relative rounded-2xl border border-white/[0.08] bg-white/[0.015] p-6 hover:border-white/[0.14] transition-colors"
              >
                {/* Tag */}
                <div className={`inline-flex items-center gap-1.5 rounded-full border ${tone.border} ${tone.bg} px-2.5 py-0.5 mb-4`}>
                  <span className={`w-1.5 h-1.5 rounded-full ${tone.dot}`} />
                  <span className={`font-mono text-[10px] ${tone.text}`}>
                    {lane.tag}
                  </span>
                </div>

                {/* Headline + sub */}
                <h3 className="font-body-mature text-[16px] font-semibold text-white leading-snug mb-2">
                  {lane.headline}
                </h3>
                <p className="font-body-mature text-[13px] text-white/40 leading-relaxed mb-5">
                  {lane.sub}
                </p>

                {/* Divider */}
                <div className="h-px w-full bg-white/[0.06] mb-4" />

                {/* Items */}
                <ul className="space-y-2.5">
                  {lane.items.map((item, ii) => (
                    <motion.li
                      key={item.name}
                      initial={{ opacity: 0, x: -8 }}
                      whileInView={{ opacity: 1, x: 0 }}
                      viewport={{ once: true }}
                      transition={{ delay: li * 0.1 + ii * 0.05 + 0.15, duration: 0.4, ease: EASE_OUT }}
                      className="flex items-baseline justify-between gap-3"
                    >
                      <span className="font-body-mature text-[13px] text-white/75 shrink-0">
                        {item.name}
                      </span>
                      <span className="font-mono text-[10.5px] text-white/30 text-right">
                        {item.note}
                      </span>
                    </motion.li>
                  ))}
                </ul>
              </motion.div>
            );
          })}
        </div>

        {/* ── Footer line — the honest guarantee ── */}
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ delay: 0.3, duration: 0.6, ease: EASE_OUT }}
          className="mt-12 flex flex-wrap items-center justify-center gap-x-6 gap-y-2 font-mono text-[11px] text-white/30"
        >
          <span className="flex items-center gap-1.5">
            <span className="w-1.5 h-1.5 rounded-full bg-green-400/60" />
            keys never leave your machine
          </span>
          <span className="text-white/15">·</span>
          <span>auto-failover to a backup key or local model</span>
          <span className="text-white/15">·</span>
          <span>per-key spend caps + alerts</span>
          <span className="text-white/15">·</span>
          <span>full audit log of every call</span>
        </motion.div>
      </div>
    </section>
  );
}