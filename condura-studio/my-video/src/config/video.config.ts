// src/config/video.config.ts
//
// Single source of truth for the launch film — White Room direction.
//
// Aesthetic: museum / architectural / Dieter-Rams-meets-Lin-Helge.
//   • One accent (azurite) earns its presence by being alone.
//   • Display type is hairline (Inter/Helvetica Neue Thin @ 200px).
//   • Background is pure white — every element earns its ink.
//
// If you change a duration, the net 52s auto-recomputes. If you change
// an accent, every accent in every scene changes with it.

import type {
  CompositionSpec,
  EasingToken,
  SpringToken,
  TransitionKind,
  VideoConfig,
} from "./types";

// Re-export tokens so consumers only need one entry-point.
export type { EasingToken, SpringToken, TransitionKind };

// ----------------------------------------------------------------------
// Composition — 6 scenes, 1635 source frames, 75 overlap → 52.0s net.
//
// transitionKind:  the cross-dissolve is the default; we use "dissolve"
// on the two pivots where the next scene has a different *mood* —
// the dissolve adds a soft glow at the meeting line.
//   SystemEmergence → Conductor     : mood shifts from abstract to active
//   Sovereignty    → Constellation : mood shifts from principles to lineup
// ----------------------------------------------------------------------

const composition: CompositionSpec = {
  id: "ConduraLaunchFilm",
  width: 1920,
  height: 1080,
  fps: 30,
  scenes: [
    { id: "cold-open", title: "Cold Open", durationInFrames: 240 },
    {
      id: "system-emergence",
      title: "System Emergence",
      durationInFrames: 300,
      transitionKind: "dissolve",
    },
    {
      id: "conductor",
      title: "The Conductor",
      durationInFrames: 300,
    },
    {
      id: "sovereignty",
      title: "Local Sovereignty",
      durationInFrames: 300,
      transitionKind: "dissolve",
    },
    {
      id: "constellation",
      title: "The Constellation",
      durationInFrames: 300,
    },
    { id: "close", title: "Brand Close", durationInFrames: 195 },
  ],
};

// ----------------------------------------------------------------------
// Palette — pure white canvas, near-black ink, single azurite accent.
//
// Line colors are intentionally low-contrast hairlines. On white,
// anything 0.06+ reads as furniture; >0.18 reads as a wall.
// ----------------------------------------------------------------------
const palette: VideoConfig["palette"] = {
  bg: "#FFFFFF", // paper-white canvas
  bgInk: "#0A0A0B", // reserved for negative variants only
  ink: "#0F0F10", // near-ink
  inkMuted: "rgba(15, 15, 16, 0.62)",
  inkMutedSoft: "rgba(15, 15, 16, 0.42)",
  line: "rgba(15, 15, 16, 0.14)", // hairline rule
  lineStrong: "rgba(15, 15, 16, 0.28)", // heavier rule
  accent: "#1B4DFF", // azurite — the *only* chromatic element
  accentSoft: "rgba(27, 77, 255, 0.16)",
  markLine: "#0F0F10", // mark accent bar (was olive in the warm variant)
};

// ----------------------------------------------------------------------
// Typography — thin sans for display, Inter for body, mono for labels.
//
// fontWeight 200 is the soul of this direction. Helvetica Neue is
// present on every macOS/iOS device; Inter is present via Microsoft
// and most Linux desktops. Falls back to system-ui which has at least
// a Light variant on every modern OS.
// ----------------------------------------------------------------------
const type: VideoConfig["type"] = {
  displayFamily:
    '"Helvetica Neue", "Inter", -apple-system, BlinkMacSystemFont, "Segoe UI Light", system-ui, sans-serif',
  bodyFamily:
    '"Inter", -apple-system, BlinkMacSystemFont, "Helvetica Neue", "Segoe UI", system-ui, sans-serif',
  monoFamily:
    '"SF Mono", ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, monospace',
  tracking: {
    // Negative tracking pulls hairline display type into a confident
    // block. Positive tracking would look fussy at 184px.
    display: "-0.02em",
    body: "0",
    // Wide tracking on eyebrows — the wide-open monospace label is the
    // single most reliable "this is engineered" signal we have.
    eyebrow: "0.32em",
  },
};

// ----------------------------------------------------------------------
// Motion — all springs are overdamped. Architectural glide, not settle.
// We want movement that stops *precisely*, not movement that *arrives*.
// ----------------------------------------------------------------------
const motion: VideoConfig["motion"] = {
  springs: {
    // Ratio = damping / (2·√(mass·stiffness))
    //   subtle  : 28 / 2√90     ≈ 1.48  (architectural glide)
    //   soft    : 26 / 2√110    ≈ 1.24  (settle)
    //   gentle  : 24 / 2√130    ≈ 1.05  (just overdamped, snappy)
    subtle: { damping: 28, mass: 1, stiffness: 90 },
    soft: { damping: 26, mass: 1, stiffness: 110 },
    gentle: { damping: 24, mass: 1, stiffness: 130 },
  },
  easings: {
    // linear → confidence, settle → arrival, soft → fade
    entrance: [0.2, 0, 0, 1], // engineering linear-in-out
    settle: [0.16, 1, 0.3, 1], // classic ease-out-expo
    soft: [0.4, 0, 0.2, 1], // material-style soft fade
  },
  defaultTransitionFrames: 15,
};

const grain: VideoConfig["grain"] = {
  // Lower than before — on white, grain reads ~2× louder.
  // baseFrequency dialed down so per-frame fractalNoise is cheaper.
  intensity: 0.05,
  seed: 1729,
  baseFrequency: 0.72,
};

// ----------------------------------------------------------------------
// Aggregate + derived constants
// ----------------------------------------------------------------------

export const videoConfig: VideoConfig = {
  meta: {
    project: "Condura",
    title: "Condura — Launch Film",
    subtitle: "52s · 1920×1080 · 30 fps",
    author: "Condura Brand",
  },
  composition,
  palette,
  type,
  motion,
  grain,
};

export const SCENES_BY_ID = Object.fromEntries(
  composition.scenes.map((s) => [s.id, s]),
) as Readonly<Record<string, (typeof composition.scenes)[number]>>;

export const SCENE_LIST = composition.scenes;

export const TOTAL_SOURCE_FRAMES = composition.scenes.reduce(
  (acc, s) => acc + s.durationInFrames,
  0,
);

export const TOTAL_OVERLAP_FRAMES =
  (composition.scenes.length - 1) * motion.defaultTransitionFrames;

export const NET_DURATION_FRAMES =
  TOTAL_SOURCE_FRAMES - TOTAL_OVERLAP_FRAMES;

export const NET_DURATION_SECONDS = NET_DURATION_FRAMES / composition.fps;
