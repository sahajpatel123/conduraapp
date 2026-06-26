"use client";

import { motion } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/**
 * HeroMarginalia — quiet side notes that frame the hero editorially.
 * Desktop only; they never compete with the headline or conductor card.
 */
export default function HeroMarginalia({ inView }: { inView: boolean }) {
  const fade = (delay: number) => ({
    initial: { opacity: 0, x: delay > 0.3 ? 12 : -12 },
    animate: inView ? { opacity: 1, x: 0 } : { opacity: 0, x: delay > 0.3 ? 12 : -12 },
    transition: { duration: 0.9, ease: EASE_OUT, delay },
  });

  return (
  <>
      {/* left — the promise */}
      <motion.aside
        {...fade(0.7)}
        className="pointer-events-none absolute left-6 top-[32vh] z-10 hidden max-w-[140px] xl:left-12 xl:block"
        aria-hidden
      >
        <div className="rule-ink-vertical mx-auto mb-4 h-16 w-px" />
        <p className="font-mono text-[10px] uppercase leading-[1.8] tracking-[0.2em] text-[var(--color-ink-faint)]">
          One key
          <br />
          Every tool
          <br />
          Your machine
        </p>
      </motion.aside>

      {/* right — the facts */}
      <motion.aside
        {...fade(0.85)}
        className="pointer-events-none absolute right-6 top-[36vh] z-10 hidden text-right xl:right-12 xl:block"
        aria-hidden
      >
        <div className="mb-3 flex flex-col items-end gap-2">
          {[
            { n: "12+", label: "tools" },
            { n: "0", label: "accounts" },
            { n: "100%", label: "local" },
          ].map((stat) => (
            <div key={stat.label} className="flex items-baseline gap-2">
              <span className="font-mono text-[10px] uppercase tracking-[0.16em] text-[var(--color-ink-ghost)]">
                {stat.label}
              </span>
              <span className="font-display text-[22px] leading-none tracking-[-0.03em] text-[var(--color-ink-mute)]">
                {stat.n}
              </span>
            </div>
          ))}
        </div>
        <div className="rule-ink-vertical ml-auto h-12 w-px" />
      </motion.aside>
    </>
  );
}
