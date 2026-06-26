"use client";

import Link from "next/link";
import { useEffect, useRef } from "react";
import Reveal from "@/components/motion/Reveal";
import MagneticButton from "@/components/motion/MagneticButton";
import { Icon, type IconKey } from "@/components/motion/Icon";
import { usePlatform } from "@/hooks/usePlatform";
import { DOWNLOADS } from "@/lib/downloads";
import { SITE, PLATFORMS, type PlatformKey } from "@/lib/site";

const PLATFORM_ICONS: Record<PlatformKey, IconKey> = {
  mac: "mac",
  windows: "windows",
  linux: "linux",
};

/**
 * DownloadCTA — closing call to action on an ink panel.
 * Muted on-ink typography + vertical platform picks that download directly.
 */
export default function DownloadCTA() {
  const knotRef = useRef<SVGGElement | null>(null);
  const detected = usePlatform();

  useEffect(() => {
    const prefersReduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
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
    <section data-section="download" className="relative mx-auto max-w-[1180px] px-6 py-28 sm:py-36">
      <div className="surface-ink relative overflow-hidden p-8 sm:p-14 lg:p-20">
        <svg
          className="pointer-events-none absolute -right-20 -top-20 h-[420px] w-[420px] opacity-30"
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
                stroke="rgba(168, 198, 182, 0.35)"
                strokeWidth="0.8"
              />
            ))}
            <circle cx="100" cy="100" r="3.5" fill="rgba(184, 175, 152, 0.45)" />
          </g>
        </svg>

        <div className="relative z-10 grid gap-12 lg:grid-cols-[1.15fr_0.85fr] lg:items-start">
          <div>
            <Reveal>
              <p className="text-mono-label text-on-ink-eyebrow">— Free forever</p>
            </Reveal>
            <Reveal as="h2" delay={0.05} className="font-display mt-4 text-[clamp(36px,5vw,68px)] leading-[0.98] tracking-[-0.035em] text-on-ink-headline text-balance">
              Bring it home.
              <br />
              <span className="italic text-on-ink-emphasis">It&apos;s yours.</span>
            </Reveal>
            <Reveal delay={0.15} as="p" className="text-on-ink-caption mt-6 max-w-[48ch] text-pretty">
              No account. No subscription. No telemetry by default. One download, one
              hotkey, and every AI you own finally talks to each other.
            </Reveal>

            <Reveal delay={0.25}>
              <ul className="mt-8 space-y-2.5">
                {PLATFORMS.map((p) => {
                  const download = DOWNLOADS[p.key];
                  return (
                    <li key={p.key}>
                      <a
                        href={download.primary.href}
                        className={`platform-pick w-full ${detected === p.key ? "platform-pick--detected" : ""}`}
                      >
                        <span className="platform-pick__icon" aria-hidden>
                          <Icon name={PLATFORM_ICONS[p.key]} size={18} />
                        </span>
                        <span className="min-w-0 flex-1 text-left">
                          <span className="platform-pick__title">{p.name}</span>
                          <span className="platform-pick__sub">{p.requirement}</span>
                        </span>
                        <span className="platform-pick__action">
                          {download.primary.label}
                          <span aria-hidden>↓</span>
                        </span>
                      </a>
                    </li>
                  );
                })}
              </ul>
            </Reveal>
          </div>

          <Reveal delay={0.2}>
            <div className="flex flex-col items-start gap-5 lg:items-end lg:pt-6">
              <MagneticButton strength={0.45} radius={120}>
                <Link href="/download" prefetch className="btn-on-ink group">
                  <span className="relative h-2 w-2" aria-hidden>
                    <span className="absolute inset-0 rounded-full bg-[var(--color-ink)] opacity-80" />
                    <span className="absolute inset-0 animate-[breathe_2s_ease-in-out_infinite] rounded-full bg-[var(--color-synapse-glow)] opacity-40" />
                  </span>
                  Download Condura
                  <svg
                    width="14"
                    height="14"
                    viewBox="0 0 12 12"
                    fill="none"
                    className="transition-transform duration-300 group-hover:translate-x-[2px] group-hover:-translate-y-[2px]"
                    aria-hidden
                  >
                    <path
                      d="M3 9L9 3M9 3H4M9 3V8"
                      stroke="currentColor"
                      strokeWidth="1.5"
                      strokeLinecap="round"
                      strokeLinejoin="round"
                    />
                  </svg>
                </Link>
              </MagneticButton>
              <p className="text-on-ink-meta lg:text-right">
                v0.1.1 · ~12 MB · signed + notarized
              </p>
              <p className="max-w-[28ch] text-[13px] leading-relaxed text-on-ink-body lg:text-right">
                Prefer to compare builds first?{" "}
                <Link href="/download" prefetch className="text-on-ink-link underline-offset-4 hover:underline">
                  Open the download page
                </Link>
                .
              </p>
            </div>
          </Reveal>
        </div>

        <div className="rule-ink-on-ink relative z-10 mt-12" />
        <div className="relative z-10 mt-6 flex flex-wrap items-center gap-x-6 gap-y-2">
          <span className="text-on-ink-meta">Also</span>
          <a href={SITE.github} target="_blank" rel="noopener noreferrer" className="text-on-ink-link text-[13px]">
            GitHub
          </a>
          <a href={SITE.discord} target="_blank" rel="noopener noreferrer" className="text-on-ink-link text-[13px]">
            Discord
          </a>
          <Link href="/manifesto" prefetch className="text-on-ink-link text-[13px]">
            The mission
          </Link>
          <Link href="/changelog" prefetch className="text-on-ink-link text-[13px]">
            Changelog
          </Link>
        </div>
      </div>
    </section>
  );
}
