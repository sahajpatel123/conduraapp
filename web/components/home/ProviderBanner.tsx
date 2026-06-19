"use client";

/* ────────────────────────────────────────────────────────────
   ProviderBanner — infinite rolling banner of supported
   providers and runtimes. Sits directly below the hero, above
   the "Bring your own AI" section.

   Design intent:
   - Centered uppercase label, widely tracked, zinc-muted.
   - Infinite leftward scroll, never pauses, ignores cursor
     and clicks — purely ambient motion.
   - Soft edge masks so names fade in/out rather than hard-cut.
   - Prefers-reduced-motion: static row (no animation).
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

export default function ProviderBanner() {
  // Duplicate enough for seamless infinite scroll.
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

        {/* Track — infinite leftward scroll, no hover pause */}
        <div className="flex w-max gap-12 py-2 will-change-transform animate-[provider-marquee_40s_linear_infinite] motion-reduce:animate-none">
          {row.map((name, i) => (
            <span
              key={`t1-${i}`}
              className="select-none whitespace-nowrap font-display text-[clamp(20px,2.4vw,28px)] font-bold tracking-[-0.01em] text-white/45"
            >
              {name}
            </span>
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