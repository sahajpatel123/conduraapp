"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState, useEffect } from "react";
import { LayoutGroup, motion } from "motion/react";
import Tooltip from "@/components/motion/Tooltip";
import { Icon, type IconKey } from "@/components/motion/Icon";
import { useReducedMotion } from "@/hooks/useReducedMotion";
import { SITE } from "@/lib/site";
import { springSnappy } from "@/lib/motion";

/* ────────────────────────────────────────────────────────────
   SiteDock — the single source of truth for navigation.

   Three segments, one bar:
     [Brand]  |  Home · How it works · Security · Mission ·
               Changelog · Download · Legal  |  ⌘K · GitHub · [CTA]

   Why this can replace the top nav:
   - Brand: the Condura wordmark in a glass chip (left).
   - Every site destination is present, deduplicated, ordered
     explore → trust → reference → act.
   - Primary CTA: a mature-button "Get Condura" pill (right),
     visually distinct from the icon buttons.
   - Quick actions: ⌘K command palette + GitHub, grouped with CTA.

   Interaction:
   - Magnify-on-hover with distance falloff (classic dock).
   - Shared-layout active pill glides to the current route.
   - Active dot under the current page's icon.
   - Tooltips on every item (hover + keyboard focus).
   - Scroll-to-top slides in on the far left after first viewport.
   - Reduced-motion: instant, no magnify, instant scroll.
   ──────────────────────────────────────────────────────────── */

type NavEntry = { href: string; label: string; icon: IconKey };
type ActionEntry = {
  href: string;
  label: string;
  icon: IconKey;
  action?: "command" | "external";
  target?: string;
};

// One unified set of every destination on the site.
// Order: explore → trust → reference → act.
const NAV_ENTRIES: NavEntry[] = [
  { href: "/", label: "Home", icon: "home" },
  { href: "/orchestration", label: "How it works", icon: "layers" },
  { href: "/security", label: "Security", icon: "shield" },
  { href: "/manifesto", label: "Mission", icon: "compass" },
  { href: "/changelog", label: "Changelog", icon: "clock" },
  { href: "/download", label: "Download", icon: "download" },
  { href: "/legal", label: "Legal", icon: "scale" },
];

const ACTIONS: ActionEntry[] = [
  { href: "#command-palette", label: "Command palette  ⌘K", icon: "command", action: "command" },
  { href: SITE.github, label: "GitHub", icon: "github", action: "external", target: "_blank" },
];

export default function SiteDock() {
  const pathname = usePathname();
  const reduced = useReducedMotion();
  const [hovered, setHovered] = useState<number | null>(null);
  const [showTop, setShowTop] = useState(false);

  useEffect(() => {
    const onScroll = () => setShowTop(window.scrollY > window.innerHeight * 0.8);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  const openCommandPalette = () => {
    window.dispatchEvent(
      new KeyboardEvent("keydown", {
        key: "k",
        metaKey: true,
        bubbles: true,
        cancelable: true,
      })
    );
  };

  const scrollToTop = () => {
    window.scrollTo({ top: 0, behavior: reduced ? "auto" : "smooth" });
  };

  const isActive = (href: string) =>
    href === "/" ? pathname === "/" : pathname === href || pathname.startsWith(`${href}/`);

  const renderNavEntry = (entry: NavEntry, index: number) => {
    const active = isActive(entry.href);
    const distance = hovered === null ? null : Math.abs(hovered - index);
    const scale = reduced
      ? 1
      : distance === null
        ? 1
        : distance === 0
          ? 1.2
          : distance === 1
            ? 1.08
            : 1;

    return (
      <Tooltip key={entry.href} label={entry.label} side="top">
        <Link
          href={entry.href}
          aria-label={entry.label}
          aria-current={active ? "page" : undefined}
          onMouseEnter={() => setHovered(index)}
          onMouseLeave={() => setHovered(null)}
          onFocus={() => setHovered(index)}
          onBlur={() => setHovered(null)}
          className="relative flex items-center justify-center rounded-2xl text-white/55 transition-colors hover:text-white"
        >
          <motion.span
            animate={{ scale }}
            transition={reduced ? { duration: 0 } : { type: "spring", stiffness: 500, damping: 30 }}
            className="relative flex h-11 w-11 items-center justify-center"
          >
            {active && (
              <motion.span
                layoutId="dock-active"
                className="absolute inset-0 rounded-2xl bg-white/[0.10] shadow-[inset_0_1px_0_rgba(255,255,255,0.10)]"
                transition={reduced ? { duration: 0 } : springSnappy}
              />
            )}
            {active && (
              <span className="absolute -bottom-1.5 left-1/2 h-1 w-1 -translate-x-1/2 rounded-full bg-white/60 shadow-[0_0_6px_rgba(255,255,255,0.5)]" />
            )}
            <Icon
              name={entry.icon}
              size={20}
              className={`relative z-10 transition-colors ${
                active ? "text-white" : "text-white/55"
              }`}
            />
          </motion.span>
        </Link>
      </Tooltip>
    );
  };

  const renderAction = (entry: ActionEntry, index: number) => {
    // Offset index so magnify falloff continues naturally from the nav row.
    const flatIndex = NAV_ENTRIES.length + index;
    const distance = hovered === null ? null : Math.abs(hovered - flatIndex);
    const scale = reduced
      ? 1
      : distance === null
        ? 1
        : distance === 0
          ? 1.15
          : distance === 1
            ? 1.06
            : 1;

    const handleClick =
      entry.action === "command"
        ? (e: React.MouseEvent) => {
            e.preventDefault();
            openCommandPalette();
          }
        : undefined;

    const inner = (
      <motion.span
        animate={{ scale }}
        transition={reduced ? { duration: 0 } : { type: "spring", stiffness: 500, damping: 30 }}
        className="relative flex h-11 w-11 items-center justify-center"
      >
        <Icon name={entry.icon} size={18} className="relative z-10 text-white/55" />
      </motion.span>
    );

    return (
      <Tooltip key={entry.href + entry.label} label={entry.label} side="top">
        {entry.action === "external" ? (
          <a
            href={entry.href}
            target={entry.target}
            rel="noopener noreferrer"
            aria-label={entry.label}
            onMouseEnter={() => setHovered(flatIndex)}
            onMouseLeave={() => setHovered(null)}
            onFocus={() => setHovered(flatIndex)}
            onBlur={() => setHovered(null)}
            className="relative flex items-center justify-center rounded-2xl text-white/55 transition-colors hover:text-white"
          >
            {inner}
          </a>
        ) : (
          <Link
            href={entry.href}
            aria-label={entry.label}
            onClick={handleClick}
            onMouseEnter={() => setHovered(flatIndex)}
            onMouseLeave={() => setHovered(null)}
            onFocus={() => setHovered(flatIndex)}
            onBlur={() => setHovered(null)}
            className="relative flex items-center justify-center rounded-2xl text-white/55 transition-colors hover:text-white"
          >
            {inner}
          </Link>
        )}
      </Tooltip>
    );
  };

  return (
    <nav
      aria-label="Primary"
      className="fixed bottom-5 left-1/2 z-[160] flex -translate-x-1/2 items-end gap-2"
    >
      {/* Scroll-to-top — slides in only after the user scrolls down. */}
      <LayoutGroup id="dock-top">
        <motion.button
          type="button"
          onClick={scrollToTop}
          aria-label="Scroll to top"
          initial={false}
          animate={{
            width: showTop ? 40 : 0,
            opacity: showTop ? 1 : 0,
            marginRight: showTop ? 8 : 0,
          }}
          transition={reduced ? { duration: 0 } : springSnappy}
          className="flex h-11 shrink-0 items-center justify-center overflow-hidden rounded-2xl border border-white/[0.08] bg-[#111113]/92 text-white/55 shadow-[0_12px_40px_rgba(0,0,0,0.45)] backdrop-blur-xl hover:text-white"
        >
          <Icon name="arrowUp" size={18} />
        </motion.button>
      </LayoutGroup>

      {/* Main dock surface */}
      <LayoutGroup id="site-dock">
        <div className="flex items-center gap-1 rounded-3xl border border-white/[0.08] bg-[#111113]/92 p-1.5 shadow-[0_12px_40px_rgba(0,0,0,0.45)] backdrop-blur-xl">
          {/* ── Brand chip ── */}
          <Link
            href="/"
            aria-label={`${SITE.name} home`}
            className="group relative flex h-11 items-center gap-2 rounded-2xl px-3"
          >
            <span className="flex h-6 w-6 items-center justify-center rounded-lg border border-white/15 bg-white/[0.05]">
              <span className="font-display text-[13px] font-bold leading-none text-white">
                C
              </span>
            </span>
            <span className="font-display text-[15px] font-semibold leading-none tracking-[-0.03em] text-white/85 transition-colors group-hover:text-white">
              {SITE.name}
            </span>
          </Link>

          {/* Divider after brand */}
          <span className="mx-0.5 h-7 w-px bg-white/[0.08]" />

          {/* ── Nav destinations ── */}
          {NAV_ENTRIES.map(renderNavEntry)}

          {/* Divider before actions */}
          <span className="mx-0.5 h-7 w-px bg-white/[0.08]" />

          {/* ── Quick actions ── */}
          {ACTIONS.map(renderAction)}

          {/* ── Primary CTA — visually distinct ── */}
          <Link
            href="/download"
            aria-label="Get Condura — download"
            onMouseEnter={() => setHovered(NAV_ENTRIES.length + ACTIONS.length)}
            onMouseLeave={() => setHovered(null)}
            onFocus={() => setHovered(NAV_ENTRIES.length + ACTIONS.length)}
            onBlur={() => setHovered(null)}
            className="group relative ml-1 flex h-11 items-center gap-2 overflow-hidden rounded-2xl border border-white/20 bg-white/[0.08] px-4 font-body-mature text-[13px] font-semibold text-white transition-colors hover:bg-white/[0.14]"
          >
            {/* Shimmer sweep on hover */}
            <span className="absolute inset-y-0 -left-10 w-10 rotate-12 bg-white/20 opacity-0 blur-sm transition-all duration-700 group-hover:left-[120%] group-hover:opacity-100" />
            <span className="relative z-10 flex items-center gap-1.5">
              <Icon name="rocket" size={15} className="text-white/80" />
              Get Condura
            </span>
          </Link>
        </div>
      </LayoutGroup>
    </nav>
  );
}