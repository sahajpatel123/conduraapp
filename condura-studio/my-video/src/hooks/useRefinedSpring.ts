// src/hooks/useRefinedSpring.ts
//
// A spring that *never* overshoots. Wraps `remotion.spring` and resolves
// a token from `video.config.ts` — so the whole project shares the same
// vocabulary (subtle / soft / gentle) and you can't accidentally
// reintroduce bounce by hand-rolling a config in a scene file.

import { spring, useCurrentFrame, useVideoConfig } from "remotion";

import { type SpringToken, videoConfig } from "../config/video.config";

type Options = {
  /** Spring preset from video.config. Defaults to "soft". */
  token?: SpringToken;
  /** Frames to delay before the spring fires. */
  delayFrames?: number;
  /** Stop iterating once movement is below this threshold. */
  durationRestThreshold?: number;
};

export function useRefinedSpring({
  token = "soft",
  delayFrames = 0,
  durationRestThreshold = 0.0005,
}: Options = {}): number {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  return spring({
    frame: frame - delayFrames,
    fps,
    config: videoConfig.motion.springs[token],
    durationRestThreshold,
  });
}
