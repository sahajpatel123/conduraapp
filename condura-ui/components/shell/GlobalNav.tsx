"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useEffect, useState } from "react";
import { motion, AnimatePresence } from "motion/react";
import { SITE } from "@/lib/site";
import { useReducedMotion } from "@/hooks/useReducedMotion";

const NAV_ITEMS = [
  { label: "How it works", href: "/orchestration" },
  { label: "Integrations", href: "/ecosystem" },
  { label: "Security", href: "/security" },
  { label: "Mission", href: "/manifesto" },
];

export default function GlobalNav() {
  const [hovered, setHovered] = useState<string | null>(null);
  const [hidden, setHidden] = useState(false);
  const pathname = usePathname();
  const reduced = useReducedMotion();

  useEffect(() => {
    if (reduced) return;
    let lastY = window.scrollY;
    let ticking = false;
    const onScroll = () => {
      if (ticking) return;
      ticking = true;
      requestAnimationFrame(() => {
        const y = window.scrollY;
        if (y < 80) setHidden(false);
        else if (y > lastY + 6) setHidden(true);
        else if (y < lastY - 6) setHidden(false);
        lastY = y;
        ticking = false;
      });
    };
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, [reduced]);

  return (
    <motion.nav
      initial={reduced ? false : { y: -100, opacity: 0 }}
      animate={{ y: hidden ? -110 : 0, opacity: hidden ? 0 : 1 }}
      transition={{
        y: { duration: 0.4, ease: [0.16, 1, 0.3, 1] },
        opacity: { duration: 0.3 },
      }}
      className="fixed left-1/2 top-3 z-[90] w-[calc(100%-20px)] max-w-[1180px] -translate-x-1/2"
      aria-label="Primary"
      aria-hidden={hidden}
    >
      <div className="relative grid h-[58px] w-full grid-cols-[auto_1fr_auto] items-center overflow-hidden rounded-full border border-transparent bg-transparent px-2.5 sm:px-3">
        {/* ── Wordmark ── */}
        <Link
          href="/"
          prefetch
          aria-label={`${SITE.name} home`}
          className="group relative col-start-1 flex items-center gap-2.5 rounded-full px-3 py-2"
          onMouseEnter={() => setHovered("home")}
          onMouseLeave={() => setHovered(null)}
        >
          <Wordmark />
          <span className="hidden font-display text-[19px] font-semibold leading-none tracking-[-0.04em] text-[var(--color-ink)] sm:inline">
            {SITE.name}
          </span>
        </Link>

        {/* ── Center nav (desktop) ── */}
        <div
          className="relative col-start-2 hidden items-center justify-self-center md:flex"
          onMouseLeave={() => setHovered(null)}
        >
          {NAV_ITEMS.map((item) => {
            const active = pathname === item.href;
            return (
              <Link
                key={item.href}
                href={item.href}
                prefetch
                className="relative rounded-full px-4 py-2 text-[13.5px] font-semibold tracking-[-0.005em] text-[var(--color-ink-soft)] transition-colors hover:text-[var(--color-ink)]"
                onMouseEnter={() => setHovered(item.href)}
                data-active={active}
              >
                {hovered === item.href && (
                  <motion.span
                    layoutId="nav-pill"
                    className="absolute inset-0 -z-10 rounded-full bg-[rgba(20,17,11,0.06)]"
                    transition={{ type: "spring", stiffness: 420, damping: 34 }}
                  />
                )}
                {active && (
                  <span className="absolute left-1/2 top-1/2 h-1 w-1 -translate-x-1/2 -translate-y-[14px] rounded-full bg-[var(--color-synapse)]" />
                )}
                <span className="relative">{item.label}</span>
              </Link>
            );
          })}
        </div>

        {/* ── Right cluster ── */}
        <div className="col-start-3 flex items-center gap-1.5 justify-self-end sm:gap-2">
          <a
            href={SITE.github}
            target="_blank"
            rel="noopener noreferrer"
            className="hidden rounded-full px-3 py-2 text-[13px] font-semibold text-[var(--color-ink-mute)] transition-colors hover:text-[var(--color-ink)] sm:inline"
          >
            GitHub
          </a>
          <Link href="/download" prefetch className="group btn btn-primary !px-4 !py-2.5 !text-[13px] !font-semibold">
            <span className="relative h-1.5 w-1.5">
              <span className="absolute inset-0 rounded-full bg-[var(--color-synapse-light)]" />
              <span className="absolute inset-0 animate-[breathe_2.4s_ease-in-out_infinite] rounded-full bg-[var(--color-synapse-glow)]" />
            </span>
            <span className="hidden sm:inline">Download</span>
            <span className="sm:hidden">Get</span>
            <Arrow />
          </Link>

          {/* Mobile menu toggle */}
          <MobileMenu />
        </div>
      </div>
    </motion.nav>
  );
}

function MobileMenu() {
  const [open, setOpen] = useState(false);
  const pathname = usePathname();

  // Reset menu state whenever the route changes by remounting.
  return (
    <MobileMenuPanel
      key={pathname}
      open={open}
      setOpen={setOpen}
    />
  );
}

function MobileMenuPanel({
  open,
  setOpen,
}: {
  open: boolean;
  setOpen: (v: boolean) => void;
}) {
  return (
    <>
      <button
        onClick={() => setOpen(!open)}
        aria-label="Toggle navigation menu"
        aria-expanded={open}
        className="flex h-10 w-10 items-center justify-center rounded-full text-[var(--color-ink-soft)] transition-colors hover:bg-[rgba(20,17,11,0.06)] md:hidden"
      >
        <svg width="18" height="18" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="1.6" strokeLinecap="round">
          {open ? <path d="M18 6L6 18M6 6l12 12" /> : <path d="M4 7h16M4 12h16M4 17h16" />}
        </svg>
      </button>

      <AnimatePresence>
        {open && (
          <>
            <motion.div
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              className="fixed inset-0 z-40 bg-[rgba(20,17,11,0.35)] backdrop-blur-[2px] md:hidden"
              onClick={() => setOpen(false)}
            />
            <motion.div
              initial={{ opacity: 0, y: -10, scale: 0.97 }}
              animate={{ opacity: 1, y: 0, scale: 1 }}
              exit={{ opacity: 0, y: -10, scale: 0.97 }}
              transition={{ duration: 0.22, ease: [0.22, 1, 0.36, 1] }}
              className="absolute left-0 right-0 top-[68px] z-50 rounded-3xl border border-[rgba(20,17,11,0.12)] bg-[var(--color-paper-warm)] p-2 shadow-[0_30px_60px_-20px_rgba(20,17,11,0.35)] md:hidden"
            >
              {NAV_ITEMS.map((item) => (
                <Link
                  key={item.href}
                  href={item.href}
                  prefetch
                  onClick={() => setOpen(false)}
                  className="block rounded-2xl px-4 py-3 text-[15px] font-semibold text-[var(--color-ink-soft)] transition-colors hover:bg-[rgba(20,17,11,0.05)] hover:text-[var(--color-ink)]"
                >
                  {item.label}
                </Link>
              ))}
              <div className="rule-ink my-2" />
              <a
                href={SITE.github}
                target="_blank"
                rel="noopener noreferrer"
                onClick={() => setOpen(false)}
                className="block rounded-2xl px-4 py-3 text-[15px] font-semibold text-[var(--color-ink-soft)] transition-colors hover:bg-[rgba(20,17,11,0.05)] hover:text-[var(--color-ink)]"
              >
                GitHub
              </a>
            </motion.div>
          </>
        )}
      </AnimatePresence>
    </>
  );
}

function Wordmark() {
  // A tiny synapse mark — a node with two reaching threads.
  return (
    <svg width="22" height="22" viewBox="0 0 24 24" fill="none" aria-hidden>
      <path
        className="synapse-thread"
        d="M3 18 C 8 12, 14 16, 21 7"
        strokeDasharray="40"
        style={{ strokeDashoffset: 0 }}
      />
      <path
        className="synapse-thread"
        d="M3 6 C 9 14, 13 4, 21 14"
        style={{ opacity: 0.4 }}
      />
      <circle className="synapse-node" cx="12" cy="11.5" r="2.6" />
      <circle className="synapse-node-ring" cx="12" cy="11.5" r="5" />
    </svg>
  );
}

function Arrow() {
  return (
    <svg
      width="12"
      height="12"
      viewBox="0 0 12 12"
      fill="none"
      className="transition-transform duration-300 group-hover:translate-x-[2px] group-hover:-translate-y-[2px]"
    >
      <path d="M3 9L9 3M9 3H4M9 3V8" stroke="currentColor" strokeWidth="1.4" strokeLinecap="round" strokeLinejoin="round" />
    </svg>
  );
}
