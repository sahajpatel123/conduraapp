"use client";

import Link from "next/link";
import { motion, useInView } from "motion/react";
import { useRef } from "react";
import SynapseGarden from "./SynapseGarden";
import HeroThread from "./HeroThread";
import HeroConductor from "./HeroConductor";
import HeroMarginalia from "./HeroMarginalia";
import HeroProof from "./HeroProof";
import MagneticButton from "@/components/motion/MagneticButton";
import WordReveal from "@/components/motion/WordReveal";
import { EASE_OUT } from "@/lib/motion";

/**
 * HeroSection — the opening statement.
 *
 * Editorial stack: headline → living conductor card → proof chips → CTAs.
 * The garden breathes behind; marginalia frames the stage on wide screens.
 */
export default function HeroSection() {
  const ref = useRef<HTMLDivElement | null>(null);
  const inView = useInView(ref, { once: true, margin: "-8%" });

  return (
    <section
      ref={ref}
      data-section="hero"
      className="relative min-h-[100svh] w-full overflow-hidden"
    >
      <SynapseGarden />
      <HeroThread />
      <HeroMarginalia inView={inView} />

      <div className="relative z-10 mx-auto flex min-h-[100svh] max-w-[1100px] flex-col items-center px-6 pb-6 pt-[13vh] sm:pt-[15vh]">
        {/* ── Headline ── */}
        <div className="text-center">
          <WordReveal
            as="h1"
            text="Your computer,"
            className="text-hero text-[var(--color-ink)] leading-[0.95]"
            delay={0.12}
            stagger={0.055}
          />
          <h1 className="text-hero -mt-[0.08em] overflow-hidden text-[var(--color-ink)] leading-[0.95]">
            <motion.span
              className="inline-block italic text-[var(--color-synapse)]"
              initial={{ y: "110%", opacity: 0 }}
              animate={
                inView ? { y: "0%", opacity: 1 } : { y: "110%", opacity: 0 }
              }
              transition={{ duration: 0.95, ease: EASE_OUT, delay: 0.38 }}
            >
              alive.
            </motion.span>
          </h1>
        </div>

        {/* ── Subhead — sits above the card so CTAs stay in view ── */}
        <motion.p
          initial={{ opacity: 0, y: 12 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
          transition={{ duration: 0.85, ease: EASE_OUT, delay: 0.42 }}
          className="text-lead mt-5 max-w-[40ch] text-center text-[var(--color-ink-soft)] text-pretty sm:mt-6"
        >
          One hotkey conducts every AI on your desk — local, private, under
          your control.
        </motion.p>

        {/* ── Living centerpiece ── */}
        <div className="mt-5 w-full sm:mt-6">
          <HeroConductor inView={inView} delay={0.5} />
        </div>

        <HeroProof inView={inView} delay={0.62} />

        {/* ── CTAs ── */}
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
          transition={{ duration: 0.8, ease: EASE_OUT, delay: 0.74 }}
          className="mt-6 flex flex-col items-center gap-3 sm:mt-7 sm:flex-row"
        >
          <MagneticButton strength={0.4}>
            <Link href="/download" prefetch className="btn btn-primary group">
              <span className="relative h-1.5 w-1.5">
                <span className="absolute inset-0 rounded-full bg-[var(--color-synapse-light)]" />
                <span className="absolute inset-0 animate-[breathe_2.4s_ease-in-out_infinite] rounded-full bg-[var(--color-synapse-glow)]" />
              </span>
              Download for free
              <svg
                width="13"
                height="13"
                viewBox="0 0 12 12"
                fill="none"
                className="transition-transform duration-300 group-hover:translate-x-[2px] group-hover:-translate-y-[2px]"
              >
                <path
                  d="M3 9L9 3M9 3H4M9 3V8"
                  stroke="currentColor"
                  strokeWidth="1.4"
                  strokeLinecap="round"
                  strokeLinejoin="round"
                />
              </svg>
            </Link>
          </MagneticButton>
          <Link href="/orchestration" prefetch className="btn btn-ghost">
            See how it works
          </Link>
        </motion.div>

        {/* ── Spacer pushes roster to the foot ── */}
        <div className="flex-1 min-h-[4vh]" />

        {/* ── Roster + scroll cue ── */}
        <motion.div
          initial={{ opacity: 0 }}
          animate={inView ? { opacity: 1 } : { opacity: 0 }}
          transition={{ duration: 1, ease: EASE_OUT, delay: 0.95 }}
          className="flex w-full flex-col items-center gap-4 pb-4"
        >
          <div className="flex flex-wrap items-center justify-center gap-x-5 gap-y-1.5 text-[11px] uppercase tracking-[0.16em] text-[var(--color-ink-mute)]">
            <span className="font-mono">Conducts</span>
            {[
              "Claude Code",
              "Codex",
              "Antigravity",
              "Ollama",
              "Gemini",
              "OpenCode",
              "Hermes",
              "Kilo",
            ].map((t, i) => (
              <span key={t} className="flex items-center gap-2.5">
                {i > 0 && (
                  <span className="h-1 w-1 rounded-full bg-[var(--color-ink-faint)]" />
                )}
                <span className="font-mono text-[var(--color-ink-soft)]">
                  {t}
                </span>
              </span>
            ))}
          </div>
          <ScrollCue />
        </motion.div>
      </div>
    </section>
  );
}

function ScrollCue() {
  return (
    <a
      href="#manifesto-start"
      className="group flex flex-col items-center gap-2 text-[var(--color-ink-mute)] transition-colors hover:text-[var(--color-ink)]"
      aria-label="Scroll to content"
    >
      <span className="text-[10px] font-mono uppercase tracking-[0.2em]">
        Scroll
      </span>
      <span className="relative h-9 w-px overflow-hidden bg-[rgba(20,17,11,0.15)]">
        <span className="absolute inset-x-0 top-0 h-3 animate-[thread-draw_1.8s_var(--thread-ease)_infinite] bg-[var(--color-synapse)]" />
      </span>
    </a>
  );
}
