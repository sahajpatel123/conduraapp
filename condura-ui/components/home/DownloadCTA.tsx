"use client";

import Link from "next/link";
import { useEffect, useRef } from "react";
import { motion } from "motion/react";
import Reveal from "@/components/motion/Reveal";
import { Icon, type IconKey } from "@/components/motion/Icon";
import { usePlatform } from "@/hooks/usePlatform";
import { useReducedMotion } from "@/hooks/useReducedMotion";
import { DOWNLOADS } from "@/lib/downloads";
import { SITE, PLATFORMS, type PlatformKey } from "@/lib/site";

const PLATFORM_ICONS: Record<PlatformKey, IconKey> = {
  mac: "mac",
  windows: "windows",
  linux: "linux",
};

/** Specs without repeating the OS name — avoids "macOSmacOS 13+" layout bugs. */
const PLATFORM_SPECS: Record<PlatformKey, string> = {
  mac: "13+ · Apple silicon & Intel",
  windows: "10+ · x64",
  linux: "glibc 2.31+ · x64",
};

const TRUST_PILLS = [
  "No account",
  "No subscription",
  "Local-first",
  "Unsigned preview builds",
] as const;

/**
 * DownloadCTA — closing call to action on an ink panel.
 * Trust chips + three platform download tiles.
 */
export default function DownloadCTA() {
  const knotRef = useRef<SVGGElement | null>(null);
  const detected = usePlatform();
  const reduced = useReducedMotion();

  useEffect(() => {
    if (reduced) return;
    const g = knotRef.current;
    if (!g) return;
    let raf = 0;
    let t = 0;
    const loop = () => {
      t += 0.005;
      g.style.transform = `rotate(${Math.sin(t) * 5}deg) scale(${1 + Math.sin(t * 0.7) * 0.02})`;
      raf = requestAnimationFrame(loop);
    };
    raf = requestAnimationFrame(loop);
    return () => cancelAnimationFrame(raf);
  }, [reduced]);

  return (
    <section data-section="download" className="relative mx-auto max-w-[1180px] px-6 py-28 sm:py-36">
      <div className="download-cta surface-ink relative overflow-hidden rounded-[var(--radius-xl)] p-8 sm:p-12 lg:p-16">
        {/* Ambient layers */}
        <div className="download-cta__glow download-cta__glow--left" aria-hidden />
        <div className="download-cta__glow download-cta__glow--right" aria-hidden />
        <p className="download-cta__watermark" aria-hidden>
          v0.1.1
        </p>

        <svg
          className="pointer-events-none absolute -right-16 -top-16 h-[min(52vw,420px)] w-[min(52vw,420px)] opacity-[0.22]"
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
                stroke="rgba(168, 198, 182, 0.4)"
                strokeWidth="0.9"
              />
            ))}
            <circle cx="100" cy="100" r="4" fill="rgba(184, 175, 152, 0.5)" />
          </g>
        </svg>

        <div className="relative z-10">
          {/* Header */}
          <div className="max-w-[62ch]">
            <Reveal>
              <p className="text-mono-label text-on-ink-eyebrow">— Free forever</p>
            </Reveal>
            <Reveal
              as="h2"
              delay={0.05}
              className="font-display mt-4 text-[clamp(38px,5.2vw,72px)] leading-[0.96] tracking-[-0.038em] text-on-ink-headline text-balance"
            >
              Bring it home.
              <br />
              <span className="italic text-on-ink-emphasis">It&apos;s yours.</span>
            </Reveal>
            <Reveal delay={0.12} as="p" className="text-on-ink-caption mt-6 text-pretty">
              No account. No subscription. Local-first. One download, one hotkey on macOS
              and Windows, and a single chat window for the provider you configure.
            </Reveal>
          </div>

          <Reveal delay={0.18}>
            <ul className="mt-8 flex flex-wrap gap-2.5" aria-label="What you get">
              {TRUST_PILLS.map((pill) => (
                <li key={pill} className="download-cta-chip">
                  <span className="download-cta-chip__dot" aria-hidden />
                  {pill}
                </li>
              ))}
            </ul>
          </Reveal>

          {/* All platforms */}
          <Reveal delay={0.22}>
            <div className="mt-8">
              <p className="text-on-ink-meta mb-4">All platforms</p>
              <ul className="grid gap-3 sm:grid-cols-3 sm:gap-4">
                {PLATFORMS.map((p, i) => {
                  const download = DOWNLOADS[p.key];
                  const isDetected = p.key === detected;
                  return (
                    <li key={p.key}>
                      <motion.a
                        href={download.primary.href}
                        className={`download-platform ${isDetected ? "download-platform--active" : ""}`}
                        initial={reduced ? false : { opacity: 0, y: 14 }}
                        whileInView={{ opacity: 1, y: 0 }}
                        viewport={{ once: true, margin: "-40px" }}
                        transition={{ delay: reduced ? 0 : i * 0.06, duration: 0.45 }}
                        whileHover={reduced ? undefined : { y: -3 }}
                      >
                        {isDetected && (
                          <span className="download-platform__badge">Your OS</span>
                        )}
                        <span className="download-platform__icon" aria-hidden>
                          <Icon name={PLATFORM_ICONS[p.key]} size={20} />
                        </span>
                        <span className="download-platform__text">
                          <span className="download-platform__name">{p.name}</span>
                          <span className="download-platform__spec">{PLATFORM_SPECS[p.key]}</span>
                        </span>
                        <span className="download-platform__action">
                          {download.primary.label}
                          <Icon name="download" size={14} strokeWidth={2} />
                        </span>
                      </motion.a>
                    </li>
                  );
                })}
              </ul>
            </div>
          </Reveal>

          <div className="rule-ink-on-ink mt-12" />
          <div className="mt-6 flex flex-col gap-4 sm:flex-row sm:items-center sm:justify-between">
            <p className="text-on-ink-body max-w-[42ch] text-[14px] leading-relaxed">
              Prefer checksums, release notes, or alternate builds?{" "}
              <Link href="/download" prefetch className="text-on-ink-link underline-offset-4 hover:underline">
                Open the full download page
              </Link>
              .
            </p>
            <div className="flex flex-wrap items-center gap-x-5 gap-y-2">
              <a
                href={SITE.github}
                target="_blank"
                rel="noopener noreferrer"
                className="download-cta-footer-link"
              >
                <Icon name="github" size={15} />
                GitHub
              </a>
              <a
                href={SITE.discord}
                target="_blank"
                rel="noopener noreferrer"
                className="download-cta-footer-link"
              >
                <Icon name="discord" size={15} />
                Discord
              </a>
              <Link href="/manifesto" prefetch className="download-cta-footer-link">
                Mission
              </Link>
              <Link href="/changelog" prefetch className="download-cta-footer-link">
                Changelog
              </Link>
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}
