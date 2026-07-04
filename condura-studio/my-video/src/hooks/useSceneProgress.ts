// src/hooks/useSceneProgress.ts
//
// Inside a TransitionSeries.Sequence, `useCurrentFrame()` returns 0 at
// the scene's local 0. This hook turns that into a clamped 0..1 progress
// plus a few derived fields every scene actually wants.

import {
  type EasingFunction,
  interpolate,
  useCurrentFrame,
  useVideoConfig,
} from "remotion";

import { type EasingToken, videoConfig } from "../config/video.config";

export type SceneProgress = {
  /** Local frame inside the wrapping Sequence (0 at scene start). */
  readonly frame: number;
  /** Clamped 0..1 progress through the scene. */
  readonly t: number;
  /** Project-wide fps. */
  readonly fps: number;
  /** Source frames inside the sequence. */
  readonly durationInFrames: number;
  /** Apply the named easing curve (entrance | settle | soft) to `t`. */
  readonly eased: (token?: EasingToken) => number;
  /**
   * Compute a 0..1 value for a window inside the scene.
   * Equivalent to `interpolate(frame, [from, to], [0, 1], clamped)`
   * but with an optional easing applied.
   */
  readonly window: (from: number, to: number, easing?: EasingFunction) => number;
};

export function useSceneProgress(durationInFrames: number): SceneProgress {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const t = Math.min(
    1,
    Math.max(0, durationInFrames === 0 ? 1 : frame / durationInFrames),
  );

  const eased = (token: EasingToken = "entrance") => {
    const bezier = videoConfig.motion.easings[token];
    return cubicBezier(t, bezier[0], bezier[1], bezier[2], bezier[3]);
  };

  const window = (from: number, to: number, easing?: EasingFunction) => {
    return interpolate(frame, [from, to], [0, 1], {
      extrapolateLeft: "clamp",
      extrapolateRight: "clamp",
      easing,
    });
  };

  return { frame, t, fps, durationInFrames, eased, window };
}

// ---------- cubic-bezier() implemented locally to avoid an extra import. --
//
// Same shape as the CSS spec; t in [0,1] → eased value in [0,1].
// Keep here rather than another file — it's only used by `eased`.

function cubicBezier(t: number, x1: number, y1: number, x2: number, y2: number): number {
  if (t <= 0) return 0;
  if (t >= 1) return 1;

  // Solve for the parametric u in x = B(u) via Newton-Raphson, then
  // evaluate y at that u. 4 iterations is plenty for our curves.
  let u = t;
  for (let i = 0; i < 4; i++) {
    const x = bezier(u, x1, x2) - t;
    const dx = bezierDeriv(u, x1, x2);
    if (Math.abs(dx) < 1e-6) break;
    u = u - x / dx;
  }

  return bezier(u, y1, y2);
}

function bezier(t: number, a: number, b: number): number {
  // 3·(1−t)²·t·a + 3·(1−t)·t²·b + t³
  const omt = 1 - t;
  return 3 * omt * omt * t * a + 3 * omt * t * t * b + t * t * t;
}

function bezierDeriv(t: number, a: number, b: number): number {
  const omt = 1 - t;
  return 3 * omt * omt * a + 6 * omt * t * (b - a) + 3 * t * t * (1 - b);
}
