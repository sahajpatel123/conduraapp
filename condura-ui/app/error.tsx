"use client";

import { useEffect } from "react";
import { motion } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/**
 * Error boundary page — a snapped thread.
 * A pollen node where the break happened, with a "try again" that
 * re-spins the thread.
 */
export default function Error({
  error,
  reset,
}: {
  error: Error & { digest?: string };
  reset: () => void;
}) {
  useEffect(() => {
    console.error(error);
  }, [error]);

  return (
    <div className="surface-paper relative flex min-h-[80vh] flex-col items-center justify-center overflow-hidden px-6">
      <div className="paper-grain absolute inset-0" />
      <svg className="absolute inset-x-0 top-1/2 h-40 w-full -translate-y-1/2" viewBox="0 0 100 20" preserveAspectRatio="none" aria-hidden>
        <path className="synapse-thread" d="M 0 10 C 30 10, 42 10, 48 10" strokeWidth="0.3" />
        <path className="synapse-thread" d="M 56 10 C 70 10, 88 10, 104 10" strokeWidth="0.3" style={{ opacity: 0.4 }} />
        <circle cx="52" cy="10" r="1" className="synapse-node">
          <animate attributeName="r" values="0.8;1.6;0.8" dur="1.8s" repeatCount="indefinite" />
        </circle>
      </svg>

      <div className="relative z-10 max-w-lg text-center">
        <motion.p
          initial={{ opacity: 0, y: 8 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 0.6, ease: EASE_OUT }}
          className="text-eyebrow"
        >
          — The thread snapped
        </motion.p>
        <h2 className="font-display text-[clamp(28px,5vw,44px)] leading-tight tracking-[-0.03em] text-[var(--color-ink)] mt-4 text-balance">
          Something broke mid-render.
        </h2>
        <p className="text-body mt-4 text-[var(--color-ink-mute)] text-pretty">
          {error.message || "An unexpected error occurred while loading this page."}
        </p>
        {error.digest && (
          <p className="text-mono-label mt-3">digest: {error.digest}</p>
        )}
        <button type="button" onClick={() => reset()} className="btn btn-primary mt-8 group">
          Try again
          <span aria-hidden className="transition-transform duration-300 group-hover:rotate-180">↻</span>
        </button>
      </div>
    </div>
  );
}
