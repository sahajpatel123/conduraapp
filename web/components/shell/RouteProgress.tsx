"use client";

import { useEffect, useState } from "react";
import { usePathname } from "next/navigation";
import { motion, AnimatePresence } from "motion/react";

/* ────────────────────────────────────────────────────────────
   RouteProgress — a thin top-edge loading bar that appears the
   instant a navigation starts and completes when the new page
   mounts. Gives the user immediate feedback that the click
   registered, eliminating the "frozen page" feeling.
   ──────────────────────────────────────────────────────────── */

export default function RouteProgress() {
  const pathname = usePathname();
  const [loading, setLoading] = useState(false);

  // Flash the bar whenever the route changes. We use a short
  // timer because Next's client navigation is fast — the bar
  // exists to confirm the click, not to track real byte progress.
  useEffect(() => {
    setLoading(true);
    const t = setTimeout(() => setLoading(false), 400);
    return () => clearTimeout(t);
  }, [pathname]);

  return (
    <AnimatePresence>
      {loading && (
        <motion.div
          className="fixed left-0 top-0 z-[300] h-[2px] w-full bg-white/80 shadow-[0_0_8px_rgba(255,255,255,0.5)]"
          initial={{ scaleX: 0, transformOrigin: "left" }}
          animate={{ scaleX: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: 0.35, ease: [0.16, 1, 0.3, 1] }}
          style={{ transformOrigin: "left" }}
          aria-hidden
        />
      )}
    </AnimatePresence>
  );
}