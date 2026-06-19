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
   SiteDock — quick-access only. Bottom-center, every page.

   Philosophy: the dock is for things a user reaches for
   constantly. Reference pages (How it works, Security,
   Mission, Changelog, Legal) live in the footer — they're
   for browsing, not quick access.

   Five things, nothing more:
     Home · Download · ⌘K · GitHub · Discord

   Interaction:
   - Magnify-on-hover with distance falloff.
   - Shared-layout active pill glides to the current route.
   - Active dot under the current page's icon.
   - Tooltips on every item (hover + keyboard focus).
   - Scroll-to-top slides in on the far left after first viewport.
   - Reduced-motion: instant, no magnify, instant scroll.
   ──────────────────────────────────────────────────────────── */

type Entry = {
  href: string;
  label: string;
  icon: IconKey;
  action?: "command" | "external";
  target?: string;
};

// Only quick-access actions. Reference content is in the footer.
const ENTRIES: Entry[] = [
  { href: "/", label: "Home", icon: "home" },
  { href: "/download", label: "Download", icon: "download" },
  { href: "#command-palette", label: "Command palette  ⌘K", icon: "command", action: "command" },
  { href: SITE.github, label: "GitHub", icon: "github", action: "external", target: "_blank" },
  { href: SITE.discord, label: "Discord", icon: "discord", action: "external", target: "_blank" },
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

  return (
    <nav
      aria-label="Quick access"
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
          {ENTRIES.map((entry, index) => {
            const active = !entry.action && isActive(entry.href);
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
                    prefetch
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
          })}
        </div>
      </LayoutGroup>
    </nav>
  );
}