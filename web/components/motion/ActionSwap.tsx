"use client";

import { AnimatePresence, motion } from "motion/react";
import { type ReactNode } from "react";

interface ActionSwapProps {
  primary: ReactNode;
  secondary: ReactNode;
  active: boolean;
  className?: string;
}

/** Cross-fades between two action labels — e.g. Download ↔ Copy link. */
export default function ActionSwap({ primary, secondary, active, className = "" }: ActionSwapProps) {
  return (
    <span className={`relative inline-grid overflow-hidden ${className}`}>
      <AnimatePresence mode="popLayout" initial={false}>
        <motion.span
          key={active ? "secondary" : "primary"}
          initial={{ opacity: 0, y: 10 }}
          animate={{ opacity: 1, y: 0 }}
          exit={{ opacity: 0, y: -10 }}
          transition={{ duration: 0.18 }}
          className="col-start-1 row-start-1"
        >
          {active ? secondary : primary}
        </motion.span>
      </AnimatePresence>
    </span>
  );
}
