// src/scenes/types.ts
//
// The contract every scene component honors. A scene is a leaf inside a
// TransitionSeries.Sequence; its `useCurrentFrame()` returns 0 at the
// start of that sequence. So scenes never read global frame numbers.

export type SceneProps = {
  /** Source frames inside the wrapping Sequence. */
  readonly durationInFrames: number;
};

/**
 * A scene is a Stateless functional component that owns the
 * AbsoluteFill, Background, typography, and motion for a single beat.
 *
 * It must NOT hardcode durations or colors — read them from
 * `videoConfig` so the timeline is one source of truth.
 */
export type SceneComponent = React.FC<SceneProps>;
