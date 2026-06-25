"use client";

import { useEffect, useRef } from "react";
import { motion, useScroll, useSpring, useTransform } from "motion/react";

/**
 * ScrollThread — a vertical synapse thread fixed to the left margin
 * that draws itself as you scroll the page. A pollen node rides along
 * the thread at the current scroll position.
 *
 * Hidden on narrow viewports (it needs margin to breathe) and under
 * reduced motion (a static faded line is rendered instead).
 */
export default function ScrollThread() {
  const ref = useRef<HTMLDivElement | null>(null);
  const { scrollYProgress } = useScroll();
  const draw = useSpring(scrollYProgress, { stiffness: 120, damping: 30, mass: 0.4 });
  const nodeY = useTransform(draw, [0, 1], ["0%", "100%"]);

  useEffect(() => {
    const prefersReduced = window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;
    if (prefersReduced && ref.current) ref.current.style.display = "none";
  }, []);

  return (
    <div
      ref={ref}
      aria-hidden
      className="pointer-events-none fixed left-4 top-0 z-30 hidden h-screen w-8 lg:block"
    >
      <svg viewBox="0 0 8 1000" className="h-full w-full" preserveAspectRatio="none">
        {/* base faded track */}
        <line
          x1="4"
          y1="0"
          x2="4"
          y2="1000"
          stroke="rgba(20,17,11,0.10)"
          strokeWidth="1"
        />
        {/* drawn synapse */}
        <motion.line
          x1="4"
          y1="0"
          x2="4"
          y2="1000"
          className="synapse-thread"
          strokeWidth="1.5"
          style={{ pathLength: draw, opacity: 0.9 }}
        />
      </svg>
      {/* riding pollen node */}
      <motion.div
        className="absolute left-1/2 h-2.5 w-2.5 -translate-x-1/2 rounded-full bg-[var(--color-pollen)]"
        style={{ top: nodeY, y: "-50%", boxShadow: "0 0 12px rgba(201,123,46,0.8)" }}
      />
    </div>
  );
}
