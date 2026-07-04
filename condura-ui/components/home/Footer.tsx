"use client";

import Link from "next/link";
import { useEffect, useRef } from "react";
import { SITE } from "@/lib/site";

/**
 * Footer — a quiet editorial sign-off. The Condura wordmark set large,
 * a single self-drawing thread underneath, the nav in a column, and a
 * final line: "Made by a human and an AI, in partnership."
 */
export default function Footer() {
  const threadRef = useRef<SVGPathElement | null>(null);

  useEffect(() => {
    const prefersReduced = window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;
    if (prefersReduced) return;
    const path = threadRef.current;
    if (!path) return;
    const io = new IntersectionObserver(
      (entries) => {
        for (const e of entries) {
          if (e.isIntersecting) {
            const len = path.getTotalLength();
            path.style.strokeDasharray = `${len}`;
            path.style.strokeDashoffset = `${len}`;
            path.getBoundingClientRect();
            path.style.transition = "stroke-dashoffset 2.6s cubic-bezier(0.22,1,0.36,1)";
            path.style.strokeDashoffset = "0";
            io.unobserve(e.target);
          }
        }
      },
      { threshold: 0.3 }
    );
    io.observe(path);
    return () => io.disconnect();
  }, []);

  return (
    <footer className="relative mt-20 border-t border-[rgba(20,17,11,0.12)] bg-[var(--color-paper-warm)]">
      <div className="mx-auto max-w-[1180px] px-6 py-20">
        {/* big wordmark + thread */}
        <div className="flex flex-col items-center text-center">
          <Link href="/" prefetch className="group inline-flex items-center gap-3" aria-label={`${SITE.name} home`}>
            <svg width="34" height="34" viewBox="0 0 24 24" fill="none" aria-hidden>
              <path className="synapse-thread" d="M3 18 C 8 12, 14 16, 21 7" strokeWidth="1.2" />
              <path className="synapse-thread" d="M3 6 C 9 14, 13 4, 21 14" strokeWidth="1.2" style={{ opacity: 0.4 }} />
              <circle className="synapse-node" cx="12" cy="11.5" r="2.6" />
              <circle className="synapse-node-ring" cx="12" cy="11.5" r="5" />
            </svg>
            <span className="font-display text-[clamp(48px,8vw,96px)] leading-none tracking-[-0.045em] text-[var(--color-ink)]">
              {SITE.name}
            </span>
          </Link>

          <svg className="mt-6 h-6 w-full max-w-[420px]" viewBox="0 0 420 24" aria-hidden>
            <path
              ref={threadRef}
              className="synapse-thread"
              d="M 4 12 C 80 2, 160 22, 210 12 C 260 2, 340 22, 416 12"
              strokeWidth="1.4"
            />
            <circle cx="210" cy="12" r="3" className="synapse-node" />
          </svg>

          <p className="mt-6 max-w-[42ch] text-lead text-[var(--color-ink-soft)] text-pretty">
            {SITE.tagline}
          </p>
        </div>

        {/* nav grid */}
        <div className="mt-16 grid grid-cols-2 gap-8 sm:grid-cols-4">
          {FOOTER_GROUPS.map((g) => (
            <div key={g.title}>
              <p className="text-mono-label mb-4">{g.title}</p>
              <ul className="space-y-2.5">
                {g.links.map((l) => (
                  <li key={l.href}>
                    <Link
                      href={l.href}
                      prefetch
                      className="thread-link text-[14px] text-[var(--color-ink-soft)]"
                    >
                      {l.label}
                    </Link>
                  </li>
                ))}
              </ul>
            </div>
          ))}
        </div>

        {/* bottom line */}
        <div className="rule-ink my-10" />
        <div className="flex flex-col items-center justify-between gap-4 text-[12px] text-[var(--color-ink-mute)] sm:flex-row">
          <p className="font-mono">
            © {new Date().getFullYear()} {SITE.name}. Free for personal & commercial use.
          </p>
          <p className="font-mono">
            Made by a human and an AI, in partnership.
          </p>
          <div className="flex gap-4 font-mono">
            <a href={SITE.github} target="_blank" rel="noopener noreferrer" className="thread-link !text-[var(--color-ink-mute)]">GitHub</a>
            <a href={SITE.discord} target="_blank" rel="noopener noreferrer" className="thread-link !text-[var(--color-ink-mute)]">Discord</a>
          </div>
        </div>
      </div>
    </footer>
  );
}

const FOOTER_GROUPS = [
  {
    title: "Product",
    links: [
      { label: "Download", href: "/download" },
      { label: "How it works", href: "/orchestration" },
      { label: "Integrations", href: "/ecosystem" },
      { label: "Security", href: "/security" },
    ],
  },
  {
    title: "Project",
    links: [
      { label: "Mission", href: "/manifesto" },
      { label: "Changelog", href: "/changelog" },
      { label: "Legal", href: "/legal" },
      { label: "Privacy", href: "/privacy" },
    ],
  },
  {
    title: "Community",
    links: [
      { label: "GitHub", href: SITE.github },
      { label: "Discord", href: SITE.discord },
    ],
  },
] as const;
