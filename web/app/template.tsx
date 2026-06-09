"use client";

/*
  Page-enter movement: each route rises gently out of the ink.
  Templates remount per navigation, which is exactly what we want here.
*/
import { m } from "motion/react";
import { EASE } from "@/lib/motion";

export default function Template({ children }: { children: React.ReactNode }) {
  return (
    <m.div
      initial={{ opacity: 0, y: 18 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.55, ease: EASE }}
    >
      {children}
    </m.div>
  );
}
