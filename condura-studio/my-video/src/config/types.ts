// src/config/types.ts
//
// Shared shape for the project. Anything that varies per video lives in
// `video.config.ts`, not in this file. This file holds *what a config is*,
// not *what ours contains*.

export type SpringToken = "subtle" | "soft" | "gentle";
export type EasingToken = "entrance" | "settle" | "soft";

export type CubicBezier = readonly [number, number, number, number];

export type SceneSpec = {
  readonly id: string;
  readonly title: string;
  readonly durationInFrames: number;
  /**
   * Which presentation to use for the transition OUT of this scene.
   * Defaults to "fade" when omitted.
   */
  readonly transitionKind?: TransitionKind;
  readonly transitionFrames?: number;
};

/** Discriminated transition types — kept open for future additions. */
export type TransitionKind = "fade" | "dissolve";

export type CompositionSpec = {
  readonly id: string;
  readonly width: number;
  readonly height: number;
  readonly fps: number;
  readonly scenes: readonly SceneSpec[];
};

export type Palette = {
  readonly bg: string;
  readonly bgInk: string;
  readonly ink: string;
  readonly inkMuted: string;
  readonly inkMutedSoft: string;
  readonly line: string;
  readonly lineStrong: string;
  /** The single chromatic accent — used at most once per scene. */
  readonly accent: string;
  readonly accentSoft: string;
  readonly markLine: string;
};

export type Typography = {
  readonly displayFamily: string;
  readonly bodyFamily: string;
  readonly monoFamily: string;
  readonly tracking: {
    readonly display: string;
    readonly body: string;
    readonly eyebrow: string;
  };
};

export type Motion = {
  readonly springs: Record<SpringToken, { damping: number; mass: number; stiffness: number }>;
  readonly easings: Record<EasingToken, CubicBezier>;
  readonly defaultTransitionFrames: number;
};

export type Grain = {
  readonly intensity: number;
  readonly seed: number;
  readonly baseFrequency: number;
};

export type VideoConfig = {
  readonly meta: {
    readonly project: string;
    readonly title: string;
    readonly subtitle: string;
    readonly author: string;
  };
  readonly composition: CompositionSpec;
  readonly palette: Palette;
  readonly type: Typography;
  readonly motion: Motion;
  readonly grain: Grain;
};
