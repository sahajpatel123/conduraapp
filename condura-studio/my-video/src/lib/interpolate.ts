// src/lib/interpolate.ts
//
// Tiny helpers that encode *our* look — gate pre-fade and post-fade
// windows, plus an alpha helper. Avoids scene files all redefining the
// same `interpolate()` calls.

import { type EasingFunction, interpolate } from "remotion";

export type FadeSpec = {
  /** Frames at scene start until fully visible. */
  readonly inStart: number;
  readonly inEnd: number;
  /** Frames at scene end where the fade-out begins. */
  readonly outStart: number;
  readonly outEnd: number;
};

/**
 * Returns an opacity [0..1] from a scene-local `frame`, given the
 * scene's fade-in/out windows. Optionally pass an easing per direction.
 */
export function sceneOpacity(
  frame: number,
  spec: FadeSpec,
  inEasing?: EasingFunction,
  outEasing?: EasingFunction,
): number {
  const fadeIn = interpolate(frame, [spec.inStart, spec.inEnd], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
    easing: inEasing,
  });
  const fadeOut = interpolate(frame, [spec.outStart, spec.outEnd], [1, 0], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
    easing: outEasing,
  });
  return Math.min(fadeIn, fadeOut);
}

/** Translate a 0..1 progress to a vertical rise in pixels. */
export function riseFor(t: number, distance = 16): number {
  return (1 - t) * distance;
}

/** Linear interpolate by name — wrapped for clarity at call sites. */
export const lerp = (
  frame: number,
  from: [number, number],
  to: [number, number],
  easing?: EasingFunction,
) =>
  interpolate(frame, [from[0], from[1]], [to[0], to[1]], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
    easing,
  });
