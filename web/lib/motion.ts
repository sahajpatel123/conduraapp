/** Shared motion tokens — one vocabulary for the whole site. */
export const EASE_OUT = [0.22, 1, 0.36, 1] as const;
export const EASE_IN_OUT = [0.65, 0, 0.35, 1] as const;

export const springSnappy = { type: "spring" as const, stiffness: 420, damping: 32 };
export const springSoft = { type: "spring" as const, stiffness: 260, damping: 28 };
export const springBouncy = { type: "spring" as const, stiffness: 380, damping: 18, mass: 0.8 };

/** Cross-route enter/exit — short enough to feel instant, long enough to read. */
export const PAGE_TRANSITION = {
  duration: 0.32,
  ease: EASE_OUT,
} as const;

export const pageMotion = {
  initial: { opacity: 0, y: 14 },
  animate: { opacity: 1, y: 0 },
  exit: { opacity: 0, y: -8 },
} as const;
