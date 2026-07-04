"use client";

import Link from "next/link";
import { motion, useInView } from "motion/react";
import { useRef } from "react";
import SynapseGarden from "./SynapseGarden";
import HeroThread from "./HeroThread";
import HeroPulse from "./HeroPulse";
import MagneticButton from "@/components/motion/MagneticButton";
import WordReveal from "@/components/motion/WordReveal";
import { EASE_OUT } from "@/lib/motion";

/**
 * HeroSection — minimal opening.
 *
 * Headline in the sky, one line of copy, a synapse pulse, two buttons.
 * The garden does the storytelling; nothing else competes with it.
 */
export default function HeroSection() {
  const ref = useRef<HTMLDivElement | null>(null);
  const inView = useInView(ref, { once: true, margin: "-10%" });

  return (
    <section
      ref={ref}
      data-section="hero"
      className="relative min-h-[100svh] w-full overflow-hidden"
    >
      <SynapseGarden />
      <HeroThread />

      {/* Copy — set into the sky, lots of air */}
      <div className="absolute inset-x-0 top-[24vh] z-10 mx-auto flex max-w-[900px] flex-col items-center px-6 text-center sm:top-[26vh]">
        <div className="text-balance">
          <WordReveal
            as="h1"
            text="Your computer,"
            className="text-hero text-[var(--color-ink)]"
            delay={0.15}
            stagger={0.05}
          />
          <h1 className="text-hero -mt-[0.06em] overflow-hidden text-[var(--color-ink)]">
            <motion.span
              className="inline-block italic text-[var(--color-synapse)]"
              initial={{ y: "110%" }}
              animate={inView ? { y: "0%" } : { y: "110%" }}
              transition={{ duration: 0.9, ease: EASE_OUT, delay: 0.32 }}
            >
              alive.
            </motion.span>
          </h1>
        </div>

        <motion.p
          initial={{ opacity: 0, y: 12, filter: "blur(5px)" }}
          animate={
            inView
              ? { opacity: 1, y: 0, filter: "blur(0px)" }
              : { opacity: 0, y: 12, filter: "blur(5px)" }
          }
          transition={{ duration: 0.85, ease: EASE_OUT, delay: 0.48 }}
          className="text-lead mt-7 max-w-[48ch] text-[var(--color-ink-soft)] text-pretty"
        >
          A local chat overlay on macOS — free, private, and under your control.
          Windows and Linux use the terminal UI today; GUI overlays for both are
          v0.2.0.
        </motion.p>

        <HeroPulse inView={inView} delay={0.62} />

        <motion.div
          initial={{ opacity: 0, y: 12 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 12 }}
          transition={{ duration: 0.8, ease: EASE_OUT, delay: 0.72 }}
          className="mt-8 flex flex-col items-center gap-3 sm:flex-row"
        >
          <MagneticButton strength={0.35}>
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
      </div>

      {/* Quiet footer strip */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={inView ? { opacity: 1 } : { opacity: 0 }}
        transition={{ duration: 1, ease: EASE_OUT, delay: 0.95 }}
        className="absolute inset-x-0 bottom-0 z-10"
      >
        <div className="mx-auto flex max-w-[1100px] flex-col items-center px-6 pb-7">
          <ScrollCue />
        </div>
      </motion.div>
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
