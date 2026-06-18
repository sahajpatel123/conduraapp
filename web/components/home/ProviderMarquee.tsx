"use client";

import { motion } from "motion/react";

/* ────────────────────────────────────────────────────────────
   ProviderMarquee — infinite rolling banner of supported
   companies/runtimes. Sits directly below the hero.

   Design intent:
   - Centered uppercase label, widely tracked, zinc-muted.
   - Two side-by-side tracks scrolling in OPPOSITE directions
     for a mature, layered feel (not a single flat strip).
   - Bold provider names in zinc-300, generous letter-spacing.
   - Soft edge masks so names fade in/out rather than hard-cut.
   - Pause on hover/focus so users can read a name.
   - Prefers-reduced-motion: static, centered row.
   ──────────────────────────────────────────────────────────── */

const PROVIDERS = [
  "Anthropic",
  "OpenAI",
  "Ollama",
  "Mistral",
  "Groq",
  "LocalAI",
  "Google",
  "xAI",
  "DeepSeek",
  "OpenRouter",
  "Together",
  "Fireworks",
];

export default function ProviderMarquee() {
  // Duplicate enough for seamless infinite scroll on each track.
  const row = [...PROVIDERS, ...PROVIDERS];

  return (
    <section
      aria-label="Supported providers and runtimes"
      className="relative w-full border-t border-white/[0.08] bg-black py-14"
    >
      {/* Label */}
      <p className="mb-10 text-center font-mono text-[11px] uppercase tracking-[0.3em] text-white/35">
        Compatible Models &amp; Runtimes
      </p>

      {/* Marquee viewport */}
      <div className="relative w-full overflow-hidden">
        {/* Edge fades */}
        <div className="pointer-events-none absolute inset-y-0 left-0 z-10 w-32 bg-gradient-to-r from-black to-transparent" />
        <div className="pointer-events-none absolute inset-y-0 right-0 z-10 w-32 bg-gradient-to-l from-black to-transparent" />

        {/* Track 1 — scrolls left */}
        <div className="flex w-max gap-12 py-2 will-change-transform animate-[provider-marquee_40s_linear_infinite] hover:[animation-play-state:paused] motion-reduce:animate-none">
          {row.map((name, i) => (
            <ProviderName key={`t1-${i}`} name={name} />
          ))}
        </div>
      </div>

      <style jsx global>{`
        @keyframes provider-marquee {
          0% { transform: translateX(0); }
          100% { transform: translateX(-50%); }
        }
      `}</style>
    </section>
  );
}

function ProviderName({ name }: { name: string }) {
  return (
    <motion.span
      whileHover={{ color: "rgba(255,255,255,0.95)" }}
      transition={{ duration: 0.3 }}
      className="select-none whitespace-nowrap font-display text-[clamp(20px,2.4vw,28px)] font-bold tracking-[-0.01em] text-white/45 transition-colors"
    >
      {name}
    </motion.span>
  );
}
