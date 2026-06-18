"use client";

import Link from "next/link";
import { usePathname } from "next/navigation";
import { useState, useEffect } from "react";
import { LayoutGroup, motion } from "motion/react";
import Tooltip from "@/components/motion/Tooltip";
import { Icon, type IconKey } from "@/components/motion/Icon";
import { useReducedMotion } from "@/hooks/useReducedMotion";
import { NAV_LINKS, SITE } from "@/lib/site";
import { springSnappy } from "@/lib/motion";

/* ────────────────────────────────────────────────────────────
   SiteDock — macOS-style dock, bottom-center, every page.

   Mature, useful, restrained:
   - Mature SVG icons (no Unicode glyphs), currentColor, hover lift.
   - Shared-layout active pill glides to the current route.
   - Magnify-on-hover: neighbors of the hovered item scale up
     subtly (the classic dock feel) without going cartoonish.
   - Two action buttons on the right, behind a divider:
     ⌘K opens the command palette, GitHub links out.
   - Tooltips on every item (hover + keyboard focus).
   - Reduced-motion: instant, no magnify.
   - Scroll-to-top hidden until you scroll past the first viewport.
   ──────────────────────────────────────────────────────────── */

type DockEntry = {
  href: string;
  label: string;
  icon: IconKey;
  action?: "command" | "external";
  target?: string;
};

// Nav items get real SVG icons now.
const ICON_BY_ROUTE: Record<string, IconKey> = {
  "/": "home",
  "/manifesto": "compass",
  "/changelog": "clock",
  "/download": "download",
  "/legal": "scale",
};

const NAV_ENTRIES: DockEntry[] = [
  { href: "/", label: "Home", icon: "home" },
  ...NAV_LINKS.map((l) => ({
    href: l.href,
    label: l.label,
    icon: ICON_BY_ROUTE[l.href] ?? "list",
  })),
];

const ACTIONS: DockEntry[] = [
  { href: "#command-palette", label: "Command palette  ⌘K", icon: "command", action: "command" },
  { href: SITE.github, label: "GitHub", icon: "github", action: "external", target: "_blank" },
];

export default function SiteDock() {
  const pathname = usePathname();
  const reduced = useReducedMotion();
  const [hovered, setHovered] = useState<number | null>(null);
  const [showTop, setShowTop] = useState(false);

  // Show the scroll-to-top affordance only after the user has scrolled
  // past the first viewport. Keeps the dock clean on first load.
  useEffect(() => {
    const onScroll = () => setShowTop(window.scrollY > window.innerHeight * 0.8);
    onScroll();
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  const openCommandPalette = () => {
    // Toggle the palette by dispatching the same shortcut it listens for.
    // Cheapest way to bridge a click → the existing keydown handler
    // without lifting palette state into a context.
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

  const renderEntry = (entry: DockEntry, index: number) => {
    const active =
      entry.href === "/"
        ? pathname === "/"
        : pathname === entry.href || pathname.startsWith(`${entry.href}/`);

    // Dock magnification: the hovered item is largest, immediate neighbors
    // scale up subtly, everything else stays base. Distance-based falloff.
    const distance = hovered === null ? null : Math.abs(hovered - index);
    const scale = reduced
      ? 1
      : distance === null
        ? 1
        : distance === 0
          ? 1.18
          : distance === 1
            ? 1.07
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
        {active && (
          <motion.span
            layoutId="dock-active"
            className="absolute inset-0 rounded-2xl bg-white/[0.10] shadow-[inset_0_1px_0_rgba(255,255,255,0.10)]"
            transition={reduced ? { duration: 0 } : springSnappy}
          />
        )}
        {/* Active dot under the icon */}
        {active && (
          <span className="absolute -bottom-1.5 left-1/2 h-1 w-1 -translate-x-1/2 rounded-full bg-white/60 shadow-[0_0_6px_rgba(255,255,255,0.5)]" />
        )}
        <Icon
          name={entry.icon}
          size={20}
          className={`relative z-10 transition-colors ${
            active ? "text-white" : "text-white/55 hover:text-white"
          }`}
        />
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
            onMouseEnter={() => setHovered(index)}
            onMouseLeave={() => setHovered(null)}
            onFocus={() => setHovered(index)}
            onBlur={() => setHovered(null)}
            className="relative flex items-center justify-center rounded-2xl text-white/55 transition-colors hover:text-white"
          >
            {inner}
          </a>
        ) : (
          <Link
            href={entry.href}
            aria-label={entry.label}
            aria-current={active ? "page" : undefined}
            onClick={handleClick}
            onMouseEnter={() => setHovered(index)}
            onMouseLeave={() => setHovered(null)}
            onFocus={() => setHovered(index)}
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
            width: showTop ? 44 : 0,
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
          {NAV_ENTRIES.map(renderEntry)}

          {/* Divider between nav and actions */}
          <span className="mx-1 h-7 w-px bg-white/[0.08]" />

          {ACTIONS.map(renderEntry)}
        </div>
      </LayoutGroup>
    </nav>
  );
}