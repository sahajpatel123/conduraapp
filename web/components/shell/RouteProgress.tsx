"use client";

import { usePathname } from "next/navigation";
import { motion } from "motion/react";
import { useReducedMotion } from "@/hooks/useReducedMotion";
import { EASE_OUT } from "@/lib/motion";

/**
 * Top-edge progress flash on route change. Keyed by pathname so we never need
 * setState inside an effect (which can fight React strict mode / HMR).
 */
export default function RouteProgress() {
  const pathname = usePathname();
  const reduced = useReducedMotion();

  if (reduced) return null;

  return (
    <motion.div
      key={pathname}
      className="pointer-events-none fixed left-0 top-0 z-[300] h-[2px] w-full origin-left bg-[var(--color-synapse)] shadow-[0_0_8px_rgba(11,61,46,0.6)]"
      initial={{ scaleX: 0 }}
      animate={{ scaleX: 1, opacity: [1, 1, 0] }}
      transition={{ duration: 0.5, ease: EASE_OUT, times: [0, 0.7, 1] }}
      aria-hidden
    />
  );
}
