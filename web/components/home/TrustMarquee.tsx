"use client";

import { motion } from "motion/react";

const BADGES = [
  "Anthropic", "OpenAI", "Ollama", "Google", "Claude Code", "Codex",
  "OpenCode", "Gemini", "DeepSeek", "Together", "Groq", "Fireworks",
];

export default function TrustMarquee() {
  const doubled = [...BADGES, ...BADGES];

  return (
    <section className="relative overflow-hidden bg-[#050505] py-16">
      <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/[0.06] to-transparent" />

      <div className="mx-auto max-w-5xl px-6 mb-8">
        <motion.div
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true, amount: 0.5 }}
          transition={{ duration: 0.5 }}
          className="text-center"
        >
          <p className="text-[13px] font-medium uppercase tracking-widest text-white/25">
            Works with everything you already use
          </p>
        </motion.div>
      </div>

      <div className="relative">
        <div className="pointer-events-none absolute inset-y-0 left-0 z-10 w-24 bg-gradient-to-r from-[#050505] to-transparent" />
        <div className="pointer-events-none absolute inset-y-0 right-0 z-10 w-24 bg-gradient-to-l from-[#050505] to-transparent" />

        <div className="animate-marquee flex w-max items-center gap-6">
          {doubled.map((badge, i) => (
            <span
              key={`${badge}-${i}`}
              className="shrink-0 inline-flex items-center rounded-full border border-white/[0.06] bg-white/[0.02] px-5 py-2 text-[13px] font-mono font-medium uppercase tracking-wide text-white/30 transition-all duration-200 hover:border-white/[0.12] hover:bg-white/[0.04] hover:text-white/50"
            >
              {badge}
            </span>
          ))}
        </div>
      </div>

      <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/[0.06] to-transparent" />
    </section>
  );
}
