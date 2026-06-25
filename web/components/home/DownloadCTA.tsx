"use client";

import Link from "next/link";
import { motion } from "motion/react";
import { useEffect, useRef } from "react";
import Reveal from "@/components/motion/Reveal";
import MagneticButton from "@/components/motion/MagneticButton";
import { SITE, PLATFORMS } from "@/lib/site";
import { EASE_OUT } from "@/lib/motion";

/**
 * DownloadCTA — the closing call to action.
 * A large ink panel with a breathing synapse knot and the three
 * platform cards. The knot is a hand-drawn tangle of threads that
 * slowly draws itself and breathes.
 */
export default function DownloadCTA() {
  const knotRef = useRef<SVGGElement | null>(null);

  useEffect(() => {
    const prefersReduced = window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;
    if (prefersReduced) return;
    const g = knotRef.current;
    if (!g) return;
    let raf = 0;
    let t = 0;
    const loop = () => {
      t += 0.006;
      g.style.transform = `rotate(${Math.sin(t) * 4}deg)`;
      raf = requestAnimationFrame(loop);
    };
    raf = requestAnimationFrame(loop);
    return () => cancelAnimationFrame(raf);
  }, []);

  return (
    <section className="relative mx-auto max-w-[1180px] px-6 py-28 sm:py-36">
      <div className="surface-ink relative overflow-hidden p-8 sm:p-14 lg:p-20">
        {/* breathing knot in the corner */}
        <svg
          className="pointer-events-none absolute -right-20 -top-20 h-[420px] w-[420px] opacity-40"
          viewBox="0 0 200 200"
          aria-hidden
        >
          <g ref={knotRef} style={{ transformOrigin: "100px 100px" }}>
            {[
              "M 30 100 C 60 40, 140 160, 170 100",
              "M 30 100 C 60 160, 140 40, 170 100",
              "M 100 30 C 40 60, 160 140, 100 170",
              "M 100 30 C 160 60, 40 140, 100 170",
            ].map((d, i) => (
              <path
                key={i}
                d={d}
                className="synapse-thread"
                stroke="var(--color-synapse-light)"
                strokeWidth="0.8"
                opacity={0.5}
              />
            ))}
            <circle cx="100" cy="100" r="4" className="synapse-node" />
          </g>
        </svg>

        <div className="relative z-10 grid gap-10 md:grid-cols-[1.2fr_1fr] md:items-center">
          <div>
            <Reveal>
              <p className="text-mono-label !text-[var(--color-synapse-light)]">
                — Free forever
              </p>
            </Reveal>
            <Reveal as="h2" delay={0.05} className="font-display mt-4 text-[clamp(36px,5vw,68px)] leading-[0.98] tracking-[-0.035em] text-[var(--color-paper)] text-balance">
              Bring it home.
              <br />
              <span className="italic text-[var(--color-synapse-light)]">
                It&apos;s yours.
              </span>
            </Reveal>
            <Reveal delay={0.15} as="p" className="text-lead mt-6 max-w-[48ch] text-[rgba(244,239,228,0.75)] text-pretty">
              No account. No subscription. No telemetry by default. One
              download, one hotkey, and every AI you own finally talks to each
              other.
            </Reveal>

            <Reveal delay={0.25}>
              <div className="mt-8 flex flex-wrap gap-3">
                {PLATFORMS.map((p) => (
                  <Link
                    key={p.key}
                    href={`/download?platform=${p.key}`}
                    prefetch
                    className="group inline-flex items-center gap-3 rounded-full border border-[rgba(244,239,228,0.16)] bg-[rgba(244,239,228,0.06)] px-5 py-3 text-[14px] font-medium text-[var(--color-paper)] transition-all hover:-translate-y-0.5 hover:border-[rgba(244,239,228,0.3)] hover:bg-[rgba(244,239,228,0.1)]"
                  >
                    <PlatformIcon platform={p.key} />
                    <span>{p.name}</span>
                    <span className="font-mono text-[11px] text-[rgba(244,239,228,0.5)]">
                      {p.requirement}
                    </span>
                  </Link>
                ))}
              </div>
            </Reveal>
          </div>

          {/* big single CTA */}
          <Reveal delay={0.2}>
            <div className="flex flex-col items-center gap-5 md:items-end">
              <MagneticButton strength={0.45} radius={120}>
                <Link
                  href="/download"
                  prefetch
                  className="group relative inline-flex items-center gap-3 rounded-full bg-[var(--color-pollen)] px-7 py-4 text-[16px] font-semibold text-[var(--color-ink)] transition-all hover:-translate-y-0.5 hover:bg-[var(--color-pollen-deep)] hover:text-[var(--color-paper)]"
                >
                  <span className="relative h-2 w-2">
                    <span className="absolute inset-0 rounded-full bg-[var(--color-ink)]" />
                    <span className="absolute inset-0 animate-[breathe_2s_ease-in-out_infinite] rounded-full bg-[var(--color-paper)]" />
                  </span>
                  Download Condura
                  <svg width="14" height="14" viewBox="0 0 12 12" fill="none" className="transition-transform duration-300 group-hover:translate-x-[2px] group-hover:-translate-y-[2px]">
                    <path d="M3 9L9 3M9 3H4M9 3V8" stroke="currentColor" strokeWidth="1.5" strokeLinecap="round" strokeLinejoin="round" />
                  </svg>
                </Link>
              </MagneticButton>
              <p className="text-mono-label !text-[rgba(244,239,228,0.5)]">
                v0.1.1 · ~12 MB · signed + notarized
              </p>
            </div>
          </Reveal>
        </div>

        {/* bottom hairline + roster */}
        <div className="relative z-10 mt-12 border-t border-[rgba(244,239,228,0.1)] pt-6">
          <div className="flex flex-wrap items-center gap-x-6 gap-y-2 text-[12px] text-[rgba(244,239,228,0.55)]">
            <span className="font-mono uppercase tracking-[0.18em]">Also</span>
            <a href={SITE.github} target="_blank" rel="noopener noreferrer" className="thread-link !text-[rgba(244,239,228,0.8)]">GitHub</a>
            <a href={SITE.discord} target="_blank" rel="noopener noreferrer" className="thread-link !text-[rgba(244,239,228,0.8)]">Discord</a>
            <Link href="/manifesto" prefetch className="thread-link !text-[rgba(244,239,228,0.8)]">The mission</Link>
            <Link href="/changelog" prefetch className="thread-link !text-[rgba(244,239,228,0.8)]">Changelog</Link>
          </div>
        </div>
      </div>
    </section>
  );
}

function PlatformIcon({ platform }: { platform: string }) {
  if (platform === "mac")
    return (
      <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" aria-hidden>
        <path d="M17.05 20.28c-.98.95-2.05.8-3.08.35-1.09-.46-2.09-.48-3.24 0-1.44.62-2.2.44-3.06-.35C2.79 15.25 3.51 7.59 9.05 7.31c1.35.07 2.29.74 3.08.8 1.18-.24 2.31-.93 3.57-.84 1.51.12 2.65.72 3.4 1.8-3.08 1.85-2.38 5.98.48 7.13-.57 1.5-1.31 2.99-2.54 4.09l.01-.01zM12 7.25c-.15-2.23 1.66-4.07 3.74-4.25.29 2.58-2.34 4.5-3.74 4.25z" />
      </svg>
    );
  if (platform === "windows")
    return (
      <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" aria-hidden>
        <path d="M3 5.1L10.4 4v7.3H3V5.1zM10.4 12.6v7.3L3 18.9v-6.3h7.4zM11.6 3.8L21 2.5v8.8h-9.4V3.8zM21 12.6v8.8l-9.4-1.3v-7.5H21z" />
      </svg>
    );
  return (
    <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor" aria-hidden>
      <path d="M12 2C6.5 2 2 6.5 2 12c0 4.4 2.9 8.2 6.8 9.5.5.1.7-.2.7-.5v-1.7c-2.8.6-3.4-1.4-3.4-1.4-.5-1.1-1.1-1.4-1.1-1.4-.9-.6.1-.6.1-.6 1 .1 1.5 1 1.5 1 .9 1.5 2.3 1.1 2.9.8.1-.6.3-1.1.6-1.4-2.2-.3-4.5-1.1-4.5-5 0-1.1.4-2 1-2.7-.1-.3-.4-1.3.1-2.7 0 0 .8-.3 2.7 1 .8-.2 1.7-.3 2.5-.3s1.7.1 2.5.3c1.9-1.3 2.7-1 2.7-1 .5 1.4.2 2.4.1 2.7.6.7 1 1.6 1 2.7 0 3.9-2.3 4.7-4.5 5 .3.3.6.9.6 1.8v2.6c0 .3.2.6.7.5C19.1 20.2 22 16.4 22 12c0-5.5-4.5-10-10-10z" />
    </svg>
  );
}
