"use client";

import Link from "next/link";
import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "motion/react";
import { NAV_LINKS, SITE } from "@/lib/site";

const BREAKPOINT = 768;

export default function GlobalNav() {
  const [mobileOpen, setMobileOpen] = useState(false);
  const [isMobile, setIsMobile] = useState(false);

  useEffect(() => {
    const check = () => setIsMobile(window.innerWidth <= BREAKPOINT);
    check();
    window.addEventListener("resize", check);
    return () => window.removeEventListener("resize", check);
  }, []);

  useEffect(() => {
    document.body.style.overflow = mobileOpen ? "hidden" : "";
    return () => { document.body.style.overflow = ""; };
  }, [mobileOpen]);

  return (
    <>
      <header className="fixed top-0 left-0 right-0 z-[100] h-[48px] border-b border-white/[0.06] bg-[#050505]/80 backdrop-blur-xl">
        <nav className="mx-auto flex h-full max-w-5xl items-center justify-between px-6">
          <Link href="/" className="text-[17px] font-semibold tracking-tight text-white">
            {SITE.name}
          </Link>

          {isMobile ? (
            <button
              aria-label="Toggle navigation"
              aria-expanded={mobileOpen}
              onClick={() => setMobileOpen(!mobileOpen)}
              className="flex flex-col items-center justify-center gap-[5px] p-2 text-white"
            >
              <motion.span
                animate={mobileOpen ? { rotate: 45, y: 4 } : { rotate: 0, y: 0 }}
                className="block h-[1.5px] w-[18px] bg-current origin-center"
              />
              <motion.span
                animate={mobileOpen ? { opacity: 0 } : { opacity: 1 }}
                className="block h-[1.5px] w-[18px] bg-current"
              />
              <motion.span
                animate={mobileOpen ? { rotate: -45, y: -4 } : { rotate: 0, y: 0 }}
                className="block h-[1.5px] w-[18px] bg-current origin-center"
              />
            </button>
          ) : (
            <div className="flex items-center gap-8 text-[13px]">
              {NAV_LINKS.map((l) => (
                <Link
                  key={l.href}
                  href={l.href}
                  className="text-white/40 transition-colors duration-150 hover:text-white/80"
                >
                  {l.label}
                </Link>
              ))}
              <a
                href={SITE.github}
                target="_blank"
                rel="noreferrer"
                className="text-white/40 transition-colors duration-150 hover:text-white/80"
              >
                GitHub
              </a>
            </div>
          )}
        </nav>
      </header>

      <AnimatePresence>
        {mobileOpen && (
          <motion.div
            initial={{ opacity: 0, y: -10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.2 }}
            className="fixed inset-x-0 top-[48px] z-[99] border-b border-white/[0.06] bg-[#050505]/95 backdrop-blur-xl"
          >
            <div className="flex flex-col gap-0 px-6 py-4 pb-8">
              {NAV_LINKS.map((l) => (
                <Link
                  key={l.href}
                  href={l.href}
                  onClick={() => setMobileOpen(false)}
                  className="border-b border-white/[0.04] py-3 text-[16px] font-medium text-white/60 transition-colors hover:text-white"
                >
                  {l.label}
                </Link>
              ))}
              <a
                href={SITE.github}
                target="_blank"
                rel="noreferrer"
                onClick={() => setMobileOpen(false)}
                className="border-b border-white/[0.04] py-3 text-[16px] font-medium text-white/60 transition-colors hover:text-white"
              >
                  GitHub
                </a>
              <a
                href={SITE.discord}
                target="_blank"
                rel="noreferrer"
                onClick={() => setMobileOpen(false)}
                className="py-3 text-[16px] font-medium text-white/60 transition-colors hover:text-white"
              >
                  Discord
                </a>
            </div>
          </motion.div>
        )}
      </AnimatePresence>
    </>
  );
}
