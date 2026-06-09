/*
  Shared motion vocabulary. Every animation on the site speaks this dialect:
  one ease, a small set of durations, mask-reveals over fades wherever
  the element is typographic.
*/
import type { Variants } from "motion/react";

export const EASE: [number, number, number, number] = [0.16, 1, 0.3, 1];

export const DUR = {
  micro: 0.25,
  short: 0.6,
  long: 1.1,
} as const;

/** Fade-and-rise for blocks of content. */
export const rise: Variants = {
  hidden: { opacity: 0, y: 28 },
  visible: {
    opacity: 1,
    y: 0,
    transition: { duration: DUR.long, ease: EASE },
  },
};

/** Mask-reveal for a single line of display type (wrap in overflow-hidden). */
export const lineReveal: Variants = {
  hidden: { y: "110%" },
  visible: {
    y: "0%",
    transition: { duration: DUR.long, ease: EASE },
  },
};

/** Parent that staggers its children. */
export const stagger = (delayChildren = 0, staggerChildren = 0.09): Variants => ({
  hidden: {},
  visible: {
    transition: { delayChildren, staggerChildren },
  },
});

/** Hairline rule that draws itself in. */
export const ruleDraw: Variants = {
  hidden: { scaleX: 0 },
  visible: {
    scaleX: 1,
    transition: { duration: DUR.long, ease: EASE },
  },
};

export const VIEWPORT = { once: true, amount: 0.35 } as const;
export const VIEWPORT_EAGER = { once: true, amount: 0.15 } as const;
