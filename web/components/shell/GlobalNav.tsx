"use client";

import Link from "next/link";
import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "motion/react";
import { useReducedMotion } from "@/hooks/useReducedMotion";
import { SITE } from "@/lib/site";

const NAV_ITEMS = [
  { label: "Download", href: "/download" },
  { label: "How it works", href: "/orchestration" },
  { label: "Security", href: "/security" },
  { label: "Mission", href: "/manifesto" },
];

export default function GlobalNav() {
  const [hoveredIndex, setHoveredIndex] = useState<number | null>(null);
  const [hidden, setHidden] = useState(false);
  const reduced = useReducedMotion();

  // Hide on scroll-down, reveal on scroll-up. The nav slides up
  // out of view and fades; on scroll-up it slides back in. The
  // site dock is a separate fixed element and is never affected.
  useEffect(() => {
    if (reduced) return;
    let lastY = window.scrollY;
    let ticking = false;

    const onScroll = () => {
      if (ticking) return;
      ticking = true;
      requestAnimationFrame(() => {
        const y = window.scrollY;
        // Always show at the very top of the page.
        if (y < 80) {
          setHidden(false);
        } else if (y > lastY + 6) {
          // Scrolling down → hide.
          setHidden(true);
        } else if (y < lastY - 6) {
          // Scrolling up → reveal.
          setHidden(false);
        }
        lastY = y;
        ticking = false;
      });
    };

    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, [reduced]);

  const handleScrollToSection = (id: string) => {
    const el = document.getElementById(id);
    if (el) el.scrollIntoView({ behavior: "smooth" });
  };

  return (
    <motion.nav
      initial={{ y: -100, opacity: 0 }}
      animate={{
        y: hidden ? -110 : 0,
        opacity: hidden ? 0 : 1,
      }}
      transition={{
        y: { duration: 0.4, ease: [0.16, 1, 0.3, 1] },
        opacity: { duration: 0.3 },
      }}
      className="fixed left-1/2 top-4 z-[90] w-[calc(100%-24px)] max-w-6xl -translate-x-1/2 sm:w-[calc(100%-40px)]"
      aria-label="Primary"
      aria-hidden={hidden}
    >
      <div className="liquid-glass relative grid h-[60px] w-full grid-cols-[1fr_auto_1fr] items-center overflow-hidden rounded-[28px] px-3 sm:px-4 lg:h-[64px]">
        <div className="pointer-events-none absolute inset-0 rounded-[28px] bg-[linear-gradient(116deg,rgba(255,255,255,0.13),rgba(255,255,255,0.035)_32%,rgba(255,255,255,0.075)_64%,rgba(255,255,255,0.03)),radial-gradient(circle_at_18%_0%,rgba(255,255,255,0.13),transparent_34%),radial-gradient(circle_at_84%_100%,rgba(255,255,255,0.06),transparent_36%)]" />
        <div className="pointer-events-none absolute inset-[1px] rounded-[27px] border border-white/[0.055] bg-black/[0.18]" />
        <div className="pointer-events-none absolute -left-16 top-1/2 h-24 w-56 -translate-y-1/2 rotate-[-10deg] rounded-full bg-white/[0.035] blur-2xl" />
        <div className="pointer-events-none absolute -right-20 top-1/2 h-24 w-64 -translate-y-1/2 rotate-[8deg] rounded-full bg-white/[0.035] blur-2xl" />
        
        <Link
          href="/"
          aria-label={`${SITE.name} home`}
          className="group relative z-10 col-start-1 flex min-w-0 items-center justify-self-start rounded-full px-4 py-2 outline-none transition-colors focus-visible:ring-2 focus-visible:ring-white/45"
        >
          <span className="relative inline-flex overflow-hidden font-display text-[20px] font-semibold leading-none tracking-[-0.04em] text-white sm:text-[22px]">
            <span className="absolute inset-x-0 bottom-0 h-px origin-left scale-x-0 bg-gradient-to-r from-white/0 via-white/65 to-white/0 transition-transform duration-500 group-hover:scale-x-100" />
            <span className="bg-gradient-to-b from-white via-white to-white/56 bg-clip-text text-transparent drop-shadow-[0_1px_16px_rgba(255,255,255,0.08)]">
              {SITE.name}
            </span>
          </span>
        </Link>

        <div 
          className="relative z-10 col-start-2 hidden items-center gap-1 justify-self-center rounded-[18px] border border-white/[0.075] bg-black/20 p-1 shadow-[inset_0_1px_0_rgba(255,255,255,0.05)] backdrop-blur-xl md:flex"
          onMouseLeave={() => setHoveredIndex(null)}
        >
          {NAV_ITEMS.map((item, i) => (
            <Link
              href={item.href}
              key={item.label}
              className="relative h-10 rounded-[14px] px-4 text-left outline-none transition-colors focus-visible:ring-2 focus-visible:ring-white/40 flex items-center"
              onMouseEnter={() => setHoveredIndex(i)}
              onFocus={() => setHoveredIndex(i)}
            >
              <AnimatePresence>
                {hoveredIndex === i && (
                  <motion.div
                    layoutId="nav-hover"
                    initial={{ opacity: 0, scale: 0.92 }}
                    animate={{ opacity: 1, scale: 1 }}
                    exit={{ opacity: 0 }}
                    transition={{ type: "spring", bounce: 0.2, duration: 0.6 }}
                    className="absolute inset-0 rounded-[14px] border border-white/[0.12] bg-white/[0.11] shadow-[inset_0_1px_0_rgba(255,255,255,0.18)]"
                  />
                )}
              </AnimatePresence>
              
              <span className={`relative z-20 font-body-mature text-[13px] transition-colors duration-300 ${
                hoveredIndex === i ? "text-[#ffffff]" : "text-[#a1a1aa]"
              }`}>
                {item.label}
              </span>
            </Link>
          ))}
        </div>

        <div className="z-10 col-start-3 flex items-center gap-2 justify-self-end sm:gap-3">
          <a
            href={SITE.github}
            target="_blank"
            rel="noopener noreferrer"
            className="hidden rounded-full px-3 py-2 font-body-mature text-[13px] text-[#a1a1aa] outline-none transition-colors hover:text-[#ffffff] focus-visible:ring-2 focus-visible:ring-white/40 sm:inline"
          >
            GitHub
          </a>
          <button 
            onClick={() => handleScrollToSection("download-tile")}
            className="group relative h-10 overflow-hidden rounded-full border border-white/18 bg-white/[0.08] px-4 font-body-mature text-[13px] font-semibold text-white outline-none transition-colors hover:bg-white/[0.14] focus-visible:ring-2 focus-visible:ring-white/45 sm:px-5"
          >
            <span className="absolute inset-y-0 -left-10 w-10 rotate-12 bg-white/25 opacity-0 blur-sm transition-all duration-700 group-hover:left-[110%] group-hover:opacity-100" />
            <span className="relative z-10 hidden sm:inline">Download v0.1.0</span>
            <span className="relative z-10 sm:hidden">Download</span>
          </button>
        </div>

      </div>
    </motion.nav>
  );
}
