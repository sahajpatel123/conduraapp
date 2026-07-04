"use client";

import { motion } from "motion/react";
import { type ReactNode } from "react";
import { EASE_OUT } from "@/lib/motion";

/**
 * Reveal — a scroll-triggered ink reveal. Blurs in from slightly down.
 * Use `delay` to stagger within a group. Respects reduced motion via
 * the global CSS rule on `.reveal-init`.
 */
export default function Reveal({
  children,
  delay = 0,
  y = 24,
  className = "",
  as = "div",
}: {
  children: ReactNode;
  delay?: number;
  y?: number;
  className?: string;
  as?: "div" | "section" | "li" | "span" | "p" | "h2" | "h3";
}) {
  const MotionTag = motion[as] as typeof motion.div;
  return (
    <MotionTag
      initial={{ opacity: 0, y, filter: "blur(6px)" }}
      whileInView={{ opacity: 1, y: 0, filter: "blur(0px)" }}
      viewport={{ once: true, margin: "-12%" }}
      transition={{ duration: 0.8, ease: EASE_OUT, delay }}
      className={className}
    >
      {children}
    </MotionTag>
  );
}
