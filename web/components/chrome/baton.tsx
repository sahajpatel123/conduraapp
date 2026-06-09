"use client";

/*
  The baton — a one-pixel brass line across the very top of the viewport
  that tracks reading progress through the page.
*/
import { m, useScroll, useSpring } from "motion/react";

export function Baton() {
  const { scrollYProgress } = useScroll();
  const scaleX = useSpring(scrollYProgress, {
    stiffness: 180,
    damping: 30,
    restDelta: 0.001,
  });

  return (
    <m.div
      aria-hidden
      style={{ scaleX }}
      className="fixed inset-x-0 top-0 z-50 h-px origin-left bg-brass"
    />
  );
}
