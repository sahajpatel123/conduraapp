"use client";

import { useEffect, useRef } from "react";
import Reveal from "@/components/motion/Reveal";

/**
 * ManifestoOpening — the quiet "why" that follows the hero.
 * A short editorial statement with a self-drawing synapse thread
 * connecting the headline to the body. The thread re-draws whenever
 * the section enters the viewport.
 */
export default function ManifestoOpening() {
  const threadRef = useRef<SVGPathElement | null>(null);
  const sectionRef = useRef<HTMLElement | null>(null);

  useEffect(() => {
    const prefersReduced = window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;
    if (prefersReduced) return;
    const path = threadRef.current;
    const section = sectionRef.current;
    if (!path || !section) return;

    const io = new IntersectionObserver(
      (entries) => {
        for (const e of entries) {
          if (e.isIntersecting) {
            const len = path.getTotalLength();
            path.style.strokeDasharray = `${len}`;
            path.style.strokeDashoffset = `${len}`;
            path.getBoundingClientRect();
            path.style.transition =
              "stroke-dashoffset 2.2s cubic-bezier(0.22,1,0.36,1)";
            path.style.strokeDashoffset = "0";
            io.unobserve(e.target);
          }
        }
      },
      { threshold: 0.4 }
    );
    io.observe(section);
    return () => io.disconnect();
  }, []);

  return (
    <section
      ref={sectionRef}
      id="manifesto-start"
      data-section="manifesto"
      className="relative mx-auto max-w-[1100px] px-6 py-32 sm:py-44"
    >
      <Reveal>
        <p className="text-eyebrow mb-10">— The premise</p>
      </Reveal>

      <div className="grid gap-12 md:grid-cols-[1.4fr_1fr] md:gap-20">
        <div>
          <Reveal as="h2" className="text-display text-[var(--color-ink)] text-balance">
            The best AI tools of this generation don&apos;t talk to each
            other.{" "}
            <span className="text-[var(--color-ink-mute)]">
              They live in separate tabs, separate subscriptions, separate
              silos — and few can actually touch your computer.
            </span>
          </Reveal>

          {/* self-drawing thread */}
          <svg
            className="my-10 h-10 w-full"
            viewBox="0 0 600 40"
            preserveAspectRatio="none"
            aria-hidden
          >
            <path
              ref={threadRef}
              className="synapse-thread"
              d="M 0 20 C 120 4, 220 36, 360 18 C 460 6, 520 28, 600 16"
              strokeWidth="1.4"
            />
            <circle cx="360" cy="18" r="3" className="synapse-node" />
          </svg>

          <Reveal delay={0.15} as="p" className="text-lead text-[var(--color-ink-soft)] text-pretty">
            Condura is the missing conductor. On macOS and Windows, one
            hotkey opens a local chat overlay. Computer-use actions reach a
            gated, audited pipeline when you explicitly invoke them. Sub-agent
            delegation works for single spawns today; parallel orchestration
            is v0.2.0.
          </Reveal>

          <Reveal delay={0.3} as="p" className="text-body mt-6 text-[var(--color-ink-mute)] max-w-[58ch]">
            Not a cloud. Not a subscription. Not another tab. A guest on your
            computer that puts the tools you already pay for behind one hotkey —
            with real orchestration coming in v0.2.0.
          </Reveal>
        </div>

        {/* Right column: a quiet &ldquo;principles&rdquo; stack */}
        <div className="flex flex-col gap-6 md:pt-6">
          {PRINCIPLES.map((p, i) => (
            <Reveal key={p.title} delay={0.1 + i * 0.08}>
              <div className="flex gap-4">
                <span className="mt-1 font-mono text-[12px] text-[var(--color-synapse)]">
                  {p.numeral}
                </span>
                <div>
                  <h3 className="font-display text-[20px] leading-tight text-[var(--color-ink)]">
                    {p.title}
                  </h3>
                  <p className="text-small mt-1.5 text-[var(--color-ink-mute)]">
                    {p.body}
                  </p>
                </div>
              </div>
            </Reveal>
          ))}
        </div>
      </div>
    </section>
  );
}

const PRINCIPLES = [
  {
    numeral: "01",
    title: "Free, forever.",
    body: "No tiers. No nags. A donate button in the menu bar. That's the whole business model.",
  },
  {
    numeral: "02",
    title: "Local, by default.",
    body: "Memory, skills, and audit log live on your disk, encrypted. The only network call is to the LLM you chose.",
  },
  {
    numeral: "03",
    title: "Safe, by structure.",
    body: "A deterministic Gatekeeper — not a model — decides what reaches your keyboard. You can always stop it.",
  },
] as const;
