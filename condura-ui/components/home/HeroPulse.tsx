"use client";

import { motion } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/**
 * HeroPulse — a single hairline with a traveling synapse dot.
 * The whole metaphor of Condura in ~180px: one thread, always moving.
 */
export default function HeroPulse({
  inView,
  delay = 0.58,
}: {
  inView: boolean;
  delay?: number;
}) {
  return (
    <motion.div
      initial={{ opacity: 0 }}
      animate={inView ? { opacity: 1 } : { opacity: 0 }}
      transition={{ duration: 0.9, ease: EASE_OUT, delay }}
      className="hero-pulse relative mx-auto mt-8 h-px w-[min(200px,40vw)]"
      aria-hidden
    >
      <span className="absolute inset-0 bg-[rgba(20,17,11,0.14)]" />
      <span className="hero-pulse-dot absolute top-1/2 h-1.5 w-1.5 -translate-y-1/2 rounded-full bg-[var(--color-synapse-glow)] shadow-[0_0_10px_rgba(26,138,106,0.45)]" />
    </motion.div>
  );
}
