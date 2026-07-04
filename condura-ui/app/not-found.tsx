"use client";

import Link from "next/link";
import { useEffect, useRef } from "react";
import { motion, useInView } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/**
 * 404 — "the lost thread."
 * A synapse path wanders off the canvas. A node pulses where the page
 * would have been. The headline is set editorially on paper.
 */
export default function NotFound() {
  const ref = useRef<HTMLDivElement | null>(null);
  const inView = useInView(ref, { once: true });
  const pathRef = useRef<SVGPathElement | null>(null);

  useEffect(() => {
    if (!inView) return;
    const p = pathRef.current;
    if (!p) return;
    const len = p.getTotalLength();
    p.style.strokeDasharray = `${len}`;
    p.style.strokeDashoffset = `${len}`;
    p.getBoundingClientRect();
    p.style.transition = "stroke-dashoffset 2.6s cubic-bezier(0.22,1,0.36,1)";
    p.style.strokeDashoffset = "0";
  }, [inView]);

  return (
    <div ref={ref} className="surface-paper relative flex min-h-screen flex-col items-center justify-center overflow-hidden px-6">
      <div className="paper-grain absolute inset-0" />

      {/* the lost thread — a wandering synapse that exits the frame */}
      <svg className="absolute inset-0 h-full w-full" viewBox="0 0 100 100" preserveAspectRatio="none" aria-hidden>
        <path
          ref={pathRef}
          className="synapse-thread"
          d="M -4 70 C 22 30, 40 78, 58 50 C 72 28, 88 64, 104 40"
          strokeWidth="0.4"
        />
        <circle cx="58" cy="50" r="0.9" className="synapse-node">
          <animate attributeName="r" values="0.7;1.4;0.7" dur="2.4s" repeatCount="indefinite" />
        </circle>
      </svg>

      <div className="relative z-10 max-w-xl text-center">
        <motion.p
          initial={{ opacity: 0, y: 8 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 8 }}
          transition={{ duration: 0.7, ease: EASE_OUT }}
          className="text-eyebrow"
        >
          — Lost thread
        </motion.p>

        <motion.div
          initial={{ opacity: 0 }}
          animate={inView ? { opacity: 1 } : { opacity: 0 }}
          transition={{ duration: 0.8, delay: 0.1 }}
          className="my-6 font-display text-[clamp(80px,18vw,200px)] leading-none tracking-[-0.05em] text-[var(--color-ink-faint)]"
        >
          404
        </motion.div>

        <h1 className="font-display text-[clamp(28px,5vw,44px)] leading-tight tracking-[-0.03em] text-[var(--color-ink)] text-balance">
          This page wandered off the canvas.
        </h1>
        <p className="text-lead mt-5 text-[var(--color-ink-soft)] text-pretty">
          The thread you followed doesn&apos;t lead anywhere — yet. Let&apos;s get you back to the garden.
        </p>

        <div className="mt-9 flex flex-col items-center justify-center gap-3 sm:flex-row">
          <Link href="/" prefetch className="btn btn-primary group">
            Back to home
            <span aria-hidden className="transition-transform duration-300 group-hover:translate-x-[2px] group-hover:-translate-y-[2px]">→</span>
          </Link>
          <Link href="/download" prefetch className="btn btn-ghost">
            Download Condura
          </Link>
        </div>
      </div>
    </div>
  );
}
