"use client";

import Link from "next/link";
import { motion, useInView } from "motion/react";
import { useRef } from "react";
import SynapseGarden from "./SynapseGarden";
import MagneticButton from "@/components/motion/MagneticButton";
import { SITE } from "@/lib/site";
import { EASE_OUT } from "@/lib/motion";

/**
 * HeroSection — the opening statement of the site.
 *
 * Layout: full-viewport SynapseGarden scene with the editorial headline
 * set into the sky (negative space, per the image brief), a subhead, two
 * CTAs, and a thin "scroll" cue at the bottom. Everything reveals with a
 * staggered ink-drying motion.
 */
export default function HeroSection() {
  const ref = useRef<HTMLDivElement | null>(null);
  const inView = useInView(ref, { once: true, margin: "-10%" });

  const line = (text: string, delay: number) => ({
    initial: { y: "110%" },
    animate: inView ? { y: "0%" } : { y: "110%" },
    transition: { duration: 0.9, ease: EASE_OUT, delay },
  });

  return (
    <section
      ref={ref}
      className="relative min-h-[100svh] w-full overflow-hidden"
    >
      <SynapseGarden />

      {/* Top eyebrow */}
      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 8 }}
        transition={{ duration: 0.8, ease: EASE_OUT, delay: 0.1 }}
        className="absolute left-1/2 top-[18vh] z-10 -translate-x-1/2 px-6 text-center"
      >
        <span className="text-eyebrow">Condura · v0.1.1 · Free forever</span>
      </motion.div>

      {/* Headline block — set into the sky */}
      <div className="absolute inset-x-0 top-[26vh] z-10 mx-auto flex max-w-[1100px] flex-col items-center px-6 text-center">
        <h1 className="text-hero text-[var(--color-ink)] text-balance">
          <span className="block overflow-hidden">
            <motion.span className="block" {...line("Your computer,", 0.18)} />
          </span>
          <span className="block overflow-hidden">
            <motion.span
              className="block italic text-[var(--color-synapse)]"
              {...line("alive.", 0.34)}
            />
          </span>
        </h1>

        <motion.p
          initial={{ opacity: 0, y: 14, filter: "blur(6px)" }}
          animate={
            inView
              ? { opacity: 1, y: 0, filter: "blur(0px)" }
              : { opacity: 0, y: 14, filter: "blur(6px)" }
          }
          transition={{ duration: 0.9, ease: EASE_OUT, delay: 0.55 }}
          className="text-lead mt-7 max-w-[52ch] text-[var(--color-ink-soft)] text-pretty"
        >
          One hotkey summons every AI tool on your machine. Condura is the
          conductor that makes Claude, Codex, Ollama, and your subscriptions
          play together — clicking, typing, and shipping while you stay in
          control. Free, local, private.
        </motion.p>

        <motion.div
          initial={{ opacity: 0, y: 14 }}
          animate={inView ? { opacity: 1, y: 0 } : { opacity: 0, y: 14 }}
          transition={{ duration: 0.8, ease: EASE_OUT, delay: 0.75 }}
          className="mt-9 flex flex-col items-center gap-3 sm:flex-row"
        >
          <MagneticButton strength={0.4}>
            <Link href="/download" prefetch className="btn btn-primary group">
              <span className="relative h-1.5 w-1.5">
                <span className="absolute inset-0 rounded-full bg-[var(--color-synapse-light)]" />
                <span className="absolute inset-0 animate-[breathe_2.4s_ease-in-out_infinite] rounded-full bg-[var(--color-synapse-glow)]" />
              </span>
              Download for free
              <svg width="13" height="13" viewBox="0 0 12 12" fill="none" className="transition-transform duration-300 group-hover:translate-x-[2px] group-hover:-translate-y-[2px]">
                <path d="M3 9L9 3M9 3H4M9 3V8" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round" />
              </svg>
            </Link>
          </MagneticButton>
          <Link href="/orchestration" prefetch className="btn btn-ghost">
            See how it works
          </Link>
        </motion.div>
      </div>

      {/* Bottom strip — the roster + scroll cue */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={inView ? { opacity: 1 } : { opacity: 0 }}
        transition={{ duration: 1, ease: EASE_OUT, delay: 1 }}
        className="absolute inset-x-0 bottom-0 z-10"
      >
        <div className="mx-auto flex max-w-[1100px] flex-col items-center gap-4 px-6 pb-7">
          <div className="flex flex-wrap items-center justify-center gap-x-5 gap-y-1.5 text-[11px] uppercase tracking-[0.16em] text-[var(--color-ink-mute)]">
            <span className="font-mono">Conducts</span>
            {["Claude Code", "Codex", "Antigravity", "Ollama", "Gemini", "OpenCode", "Hermes", "Kilo"].map(
              (t, i) => (
                <span key={t} className="flex items-center gap-2.5">
                  {i > 0 && <span className="h-1 w-1 rounded-full bg-[var(--color-ink-faint)]" />}
                  <span className="font-mono text-[var(--color-ink-soft)]">{t}</span>
                </span>
              )
            )}
          </div>
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
      <span className="text-[10px] font-mono uppercase tracking-[0.2em]">Scroll</span>
      <span className="relative h-9 w-px overflow-hidden bg-[rgba(20,17,11,0.15)]">
        <span className="absolute inset-x-0 top-0 h-3 animate-[thread-draw_1.8s_var(--thread-ease)_infinite] bg-[var(--color-synapse)]" />
      </span>
    </a>
  );
}
