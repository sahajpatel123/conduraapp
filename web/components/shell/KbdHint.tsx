"use client";

import { useEffect, useRef, useState } from "react";
import { AnimatePresence, motion } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/**
 * KbdHint — a quiet, mature nudge to power users.
 *
 * After the user has been idle (no pointer/scroll/keypress) for a few
 * seconds AND has scrolled past the first viewport, a small pill
 * appears at the bottom-right whispering "⌘K to jump." It disappears
 * the moment the user moves, and once dismissed (click or ⌘K) it
 * never returns for the session.
 *
 * It is never a modal, never blocks, never nags more than once.
 * Respectful of reduced motion (instant in/out).
 */
export default function KbdHint() {
  const [show, setShow] = useState(false);
  const [dismissed, setDismissed] = useState(false);
  const idleTimer = useRef<number | null>(null);
  const lastActivity = useRef<number>(Date.now());

  useEffect(() => {
    if (dismissed) return;
    const prefersReduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (prefersReduced) return;

    const arm = () => {
      lastActivity.current = Date.now();
      setShow(false);
      if (idleTimer.current) window.clearTimeout(idleTimer.current);
      idleTimer.current = window.setTimeout(() => {
        // Only show if the user has actually scrolled past the hero
        // and hasn't touched anything for the idle window.
        if (Date.now() - lastActivity.current >= 3800 && window.scrollY > window.innerHeight * 0.6) {
          setShow(true);
        }
      }, 4000);
    };

    const dismiss = () => {
      setShow(false);
      setDismissed(true);
      // Also fire the command palette so the user sees what the hint pointed to.
    };

    const onKey = (e: KeyboardEvent) => {
      if ((e.metaKey || e.ctrlKey) && e.key === "k") {
        dismiss();
      }
      arm();
    };

    window.addEventListener("pointermove", arm, { passive: true });
    window.addEventListener("scroll", arm, { passive: true });
    window.addEventListener("keydown", onKey);
    arm();

    return () => {
      if (idleTimer.current) window.clearTimeout(idleTimer.current);
      window.removeEventListener("pointermove", arm);
      window.removeEventListener("scroll", arm);
      window.removeEventListener("keydown", onKey);
    };
  }, [dismissed]);

  const dismissClick = () => {
    setShow(false);
    setDismissed(true);
  };

  return (
    <AnimatePresence>
      {show && (
        <motion.button
          type="button"
          onClick={dismissClick}
          initial={{ opacity: 0, y: 12, scale: 0.96 }}
          animate={{ opacity: 1, y: 0, scale: 1 }}
          exit={{ opacity: 0, y: 12, scale: 0.96 }}
          transition={{ duration: 0.4, ease: EASE_OUT }}
          className="fixed bottom-6 right-6 z-[80] hidden items-center gap-2.5 rounded-full border border-[rgba(20,17,11,0.12)] bg-[var(--color-paper-warm)] px-4 py-2.5 text-[12.5px] text-[var(--color-ink-soft)] shadow-[var(--shadow-card)] backdrop-blur-xl md:flex"
          aria-label="Press Command K to open the command palette"
        >
          <span className="text-[var(--color-ink-mute)]">Jump to anything</span>
          <kbd className="rounded-md border border-[rgba(20,17,11,0.18)] bg-[var(--color-paper)] px-1.5 py-0.5 font-mono text-[11px] text-[var(--color-ink)] shadow-[0_1px_0_rgba(20,17,11,0.06)]">
            ⌘K
          </kbd>
          <span className="ml-1 text-[var(--color-ink-faint)] transition-colors hover:text-[var(--color-ink)]" aria-hidden>
            ×
          </span>
        </motion.button>
      )}
    </AnimatePresence>
  );
}
