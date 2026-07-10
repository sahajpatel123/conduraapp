/*
 * Synaptic — Motion Tokens (TS)
 *
 * Spring and tween parameters for Svelte's spring/tweened stores. CSS
 * can't express these, so they live here and are imported by components
 * that need physics-based motion (cursor follow, message arrival, etc.).
 *
 * Locked: see docs/design-v1-redesign.md §6.
 */

import { cubicOut } from 'svelte/easing';

/* ---------------------------------------------------------------------- *
 * SPRING PRESETS
 *
 * Use via Svelte's spring() store:
 *   import { spring } from 'svelte/motion';
 *   import { SPRINGS } from '$tokens/motion';
 *   const scale = spring(1, SPRINGS.soft);
 * ---------------------------------------------------------------------- */

export const SPRINGS = {
  /** Cursor follow, message arrival — gentle overshoot */
  soft:   { stiffness: 0.18, damping: 0.24 },
  /** Panel slide, drawer in/out — medium overshoot */
  medium: { stiffness: 0.24, damping: 0.28 },
  /** Snappy interactions, button press — quick response */
  snappy: { stiffness: 0.32, damping: 0.32 },
  /** Ambient breath, idle motion — slow and gentle */
  gentle: { stiffness: 0.14, damping: 0.20 },
} as const;

export type SpringName = keyof typeof SPRINGS;

/* ---------------------------------------------------------------------- *
 * TWEEN EASINGS — for the { duration, easing } tween config
 *
 * svelte/easing helpers re-exported for component convenience.
 * ---------------------------------------------------------------------- */

export const TWEENS = {
  standard:    cubicOut,
  decelerate: cubicOut,
  accelerate: (t: number) => t * t,
  emphasized: cubicOut,
} as const;

/* ---------------------------------------------------------------------- *
 * PULSE STATE — the vital sign's parameters
 *
 * Components use these to compute scale/opacity over time via
 * requestAnimationFrame or spring stores.
 * ---------------------------------------------------------------------- */

export type PulseState = 'idle' | 'thinking' | 'awaiting' | 'error';

export interface PulseParams {
  /** Full cycle period in ms */
  period: number;
  /** Opacity range [min, max] */
  opacity: [number, number];
  /** Scale range [min, max] */
  scale: [number, number];
}

export const PULSE_PARAMS: Record<PulseState, PulseParams> = {
  idle:     { period: 5000, opacity: [0.85, 1.00], scale: [0.98, 1.02] },
  thinking: { period: 7500, opacity: [0.70, 1.00], scale: [0.98, 1.02] },
  awaiting: { period: 3000, opacity: [1.00, 1.00], scale: [1.00, 1.04] },
  error:    { period: 5000, opacity: [0.85, 1.00], scale: [1.00, 1.00] }, // one-shot flash handled separately
};

/* ---------------------------------------------------------------------- *
 * BREAKPOINTS — TS mirror of CSS custom properties
 *
 * Use in JS-driven layout logic (e.g., position the command overlay
 * differently at small viewports). For CSS, use the var(--bp-*) tokens.
 * ---------------------------------------------------------------------- */

export const BREAKPOINTS = {
  xs:  480,
  sm:  768,
  md:  1024,
  lg:  1280,
  xl:  1536,
  '2xl': 1920,
} as const;

export type Breakpoint = keyof typeof BREAKPOINTS;

/** Returns true when the viewport width matches or exceeds the breakpoint. */
export function matchesBreakpoint(width: number, bp: Breakpoint): boolean {
  return width >= BREAKPOINTS[bp];
}

/* ---------------------------------------------------------------------- *
 * PERFORMANCE BUDGETS — per energy mode
 * ---------------------------------------------------------------------- */

export type EnergyMode = 'high' | 'balanced' | 'low';

export interface EnergyConfig {
  ambientBreathAmplitude: number;  // 0..1, multiplier on pulse scale range
  luminanceDrift: boolean;
  staggerMultiplier: number;       // applied to --stagger-* tokens
  linearizeNonCritical: boolean;   // replace non-critical easings with linear
}

export const ENERGY_CONFIG: Record<EnergyMode, EnergyConfig> = {
  high:     { ambientBreathAmplitude: 1.00, luminanceDrift: true,  staggerMultiplier: 1.0, linearizeNonCritical: false },
  balanced: { ambientBreathAmplitude: 0.50, luminanceDrift: false, staggerMultiplier: 0.5, linearizeNonCritical: false },
  low:      { ambientBreathAmplitude: 0.00, luminanceDrift: false, staggerMultiplier: 0.3, linearizeNonCritical: true  },
};

/** Returns true for the few transitions that are NEVER reduced, regardless of energy mode. */
export function isUnreducibleTransition(transitionName: string): boolean {
  return [
    'kill-switch',
    'consent',
    'streaming-text-reveal',
    'pulse-presence',
  ].includes(transitionName);
}