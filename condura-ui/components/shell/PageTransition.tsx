"use client";

import { useEffect, useRef, type ReactNode } from "react";
import { usePathname } from "next/navigation";
import { motion } from "motion/react";
import { useReducedMotion } from "@/hooks/useReducedMotion";
import { PAGE_TRANSITION, pageMotion } from "@/lib/motion";

/**
 * Enter-only route fade. Avoids AnimatePresence mode="wait", which blocks the
 * next page until exit finishes — that stacks on top of Turbopack compile time
 * and feels like the app is stuck "Compiling…".
 */
export default function PageTransition({ children }: { children: ReactNode }) {
  const pathname = usePathname();
  const reduced = useReducedMotion();
  const isFirst = useRef(true);

  useEffect(() => {
    if (isFirst.current) {
      isFirst.current = false;
      return;
    }
    window.scrollTo({ top: 0, left: 0, behavior: "auto" });
  }, [pathname]);

  if (reduced) {
    return <div className="min-h-screen">{children}</div>;
  }

  return (
    <motion.div
      key={pathname}
      className="min-h-screen"
      initial={pageMotion.initial}
      animate={pageMotion.animate}
      transition={PAGE_TRANSITION}
    >
      {children}
    </motion.div>
  );
}
