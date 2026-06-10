"use client";

/*
  Top chrome. Recedes when you read down the score, returns the moment
  you scroll back. Mobile gets a full-stage overlay with staggered serif
  links; while it is up, the rest of the page is inert and Escape (or
  back/forward, or growing past the breakpoint) dismisses it.
*/
import Link from "next/link";
import { usePathname } from "next/navigation";
import { AnimatePresence, m, useMotionValueEvent, useScroll } from "motion/react";
import { useEffect, useRef, useState } from "react";
import { NAV_LINKS } from "@/lib/site";
import { DUR, EASE } from "@/lib/motion";
import { usePalette } from "./palette";

export function Nav() {
  const pathname = usePathname();
  const { setOpen } = usePalette();
  const { scrollY } = useScroll();
  const [hidden, setHidden] = useState(false);
  const [menuOpen, setMenuOpen] = useState(false);
  const toggleRef = useRef<HTMLButtonElement>(null);

  useMotionValueEvent(scrollY, "change", (y) => {
    const prev = scrollY.getPrevious() ?? 0;
    setHidden(y > prev && y > 160 && !menuOpen);
  });

  // The overlay must never outlive its context: close on history
  // navigation, on Escape, and when the viewport grows past mobile.
  useEffect(() => {
    if (!menuOpen) return;
    const closeMenu = () => setMenuOpen(false);
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") {
        closeMenu();
        toggleRef.current?.focus();
      }
    };
    const mql = window.matchMedia("(min-width: 768px)");
    const onResize = () => mql.matches && closeMenu();
    window.addEventListener("popstate", closeMenu);
    window.addEventListener("keydown", onKey);
    mql.addEventListener("change", onResize);
    return () => {
      window.removeEventListener("popstate", closeMenu);
      window.removeEventListener("keydown", onKey);
      mql.removeEventListener("change", onResize);
    };
  }, [menuOpen]);

  // Lock scroll and make the page behind the overlay inert.
  useEffect(() => {
    if (!menuOpen) return;
    document.documentElement.style.overflow = "hidden";
    const main = document.getElementById("main");
    const footer = document.querySelector("footer");
    main?.setAttribute("inert", "");
    footer?.setAttribute("inert", "");
    return () => {
      document.documentElement.style.overflow = "";
      main?.removeAttribute("inert");
      footer?.removeAttribute("inert");
    };
  }, [menuOpen]);

  return (
    <>
      <m.header
        animate={{ y: hidden ? "-100%" : "0%" }}
        transition={{ duration: 0.45, ease: EASE }}
        className="fixed inset-x-0 top-0 z-40 border-b border-line bg-ink/80 backdrop-blur-md"
      >
        <nav
          aria-label="Primary"
          className="mx-auto flex h-14 max-w-6xl items-center justify-between px-5 md:px-8"
        >
          <Link href="/" className="group flex items-center gap-2">
            <svg
              aria-hidden
              viewBox="0 0 24 32"
              className="h-5 w-auto transition-transform duration-300 ease-out group-hover:-rotate-12"
            >
              <path
                d="M8,5 C8,9 5,10.5 4,14 A8.5,8.5 0 1,0 20,14 C19,10.5 16,9 16,5 Z"
                fill="var(--t-glow)"
                fillOpacity="0.55"
                stroke="currentColor"
                strokeWidth="1.8"
              />
              <rect x="7.5" y="1" width="9" height="5" rx="1.5" fill="var(--t-bg-3)" stroke="currentColor" strokeWidth="1.5" />
            </svg>
            <span className="font-display text-lg font-bold tracking-tight">Synaptic</span>
          </Link>

          <div className="hidden items-center gap-7 md:flex">
            {NAV_LINKS.map((link) => (
              <Link
                key={link.href}
                href={link.href}
                aria-current={pathname === link.href ? "page" : undefined}
                className={`annotation !tracking-[0.14em] transition-colors duration-200 hover:!text-ivory ${
                  pathname === link.href ? "!text-brass" : ""
                }`}
              >
                {link.label}
              </Link>
            ))}
            <button
              type="button"
              onClick={() => setOpen(true)}
              className="annotation flex items-center gap-1.5 border border-line px-2.5 py-1.5 transition-colors duration-200 hover:border-line-strong hover:!text-ivory"
            >
              <kbd className="font-mono">⌘K</kbd>
            </button>
          </div>

          <button
            ref={toggleRef}
            type="button"
            onClick={() => setMenuOpen((v) => !v)}
            aria-expanded={menuOpen}
            aria-controls="mobile-menu"
            aria-label={menuOpen ? "Close menu" : "Open menu"}
            className="annotation border border-line px-3 py-1.5 md:hidden"
          >
            {menuOpen ? "Close" : "Menu"}
          </button>
        </nav>
      </m.header>

      <AnimatePresence>
        {menuOpen && (
          <m.div
            id="mobile-menu"
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            exit={{ opacity: 0 }}
            transition={{ duration: DUR.micro }}
            className="fixed inset-0 z-30 bg-ink md:hidden"
          >
            <div className="staff absolute inset-0" aria-hidden />
            <nav
              aria-label="Mobile"
              className="relative flex h-full flex-col justify-center gap-2 px-8"
            >
              {[{ href: "/", label: "Overture" }, ...NAV_LINKS].map((link, i) => (
                <span key={link.href} className="block overflow-hidden">
                  <m.span
                    initial={{ y: "110%" }}
                    animate={{ y: "0%" }}
                    exit={{ y: "110%" }}
                    transition={{ duration: DUR.short, ease: EASE, delay: 0.05 * i }}
                    className="block"
                  >
                    <Link
                      href={link.href}
                      onClick={() => setMenuOpen(false)}
                      aria-current={pathname === link.href ? "page" : undefined}
                      className={`display text-5xl ${
                        pathname === link.href ? "text-brass" : "text-ivory"
                      }`}
                    >
                      {link.label}
                    </Link>
                  </m.span>
                </span>
              ))}
              <m.p
                initial={{ opacity: 0 }}
                animate={{ opacity: 1 }}
                transition={{ delay: 0.4, duration: DUR.short }}
                className="annotation mt-10"
              >
                Free forever · No telemetry · Yours
              </m.p>
            </nav>
          </m.div>
        )}
      </AnimatePresence>
    </>
  );
}
