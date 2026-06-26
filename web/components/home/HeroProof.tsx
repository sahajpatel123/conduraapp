"use client";

import { motion } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

const PROOFS = [
  { label: "Free forever", dot: "var(--color-synapse-glow)" },
  { label: "Local-first", dot: "var(--color-pollen)" },
  { label: "You stay in control", dot: "var(--color-ink-mute)" },
] as const;

/**
 * HeroProof — three quiet trust chips between the conductor and CTAs.
 */
export default function HeroProof({
  inView,
  delay = 0.62,
}: {
  inView: boolean;
  delay?: number;
}) {
  return (
    <motion.ul
      initial={{ opacity: 0, y: 10 }}
      animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 10 }}
      transition={{ duration: 0.8, ease: EASE_OUT, delay }}
      className="mt-6 flex flex-wrap items-center justify-center gap-x-5 gap-y-2"
      aria-label="Product principles"
    >
      {PROOFS.map((p, i) => (
        <li key={p.label} className="flex items-center gap-2">
          {i > 0 && (
            <span className="hidden h-1 w-1 rounded-full bg-[var(--color-ink-faint)] sm:block" />
          )}
          <span
            className="h-1.5 w-1.5 rounded-full"
            style={{ background: p.dot }}
          />
          <span className="font-mono text-[10px] uppercase tracking-[0.16em] text-[var(--color-ink-mute)]">
            {p.label}
          </span>
        </li>
      ))}
    </motion.ul>
  );
}
