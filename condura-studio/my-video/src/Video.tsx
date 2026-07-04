// src/Video.tsx
//
// The single composition. Reads sceneRegistry for its scenes and
// videoConfig for its dimensions. TransitionSeries handles the
// cross-dissolve overlaps; we just declare the per-scene frames.
//
// Each scene's `transitionKind` chooses its presentation:
//   • "fade"      → soft cross-dissolve (default)
//   • "dissolve"  → dissolve presentation with a soft inner glow at the
//                    meeting line — used for the two pivots where the
//                    next scene has a different *mood*
//
// Hud is rendered *outside* the TransitionSeries so it reads the
// global frame counter (the one the renderer is iterating) — not the
// local scene frame. That keeps the seconds counter and progress bar
// exactly aligned with the rendered image.

import { Fragment } from "react";
import { AbsoluteFill } from "remotion";
import {
  TransitionSeries,
  linearTiming,
  type TransitionPresentation,
} from "@remotion/transitions";
import { fade } from "@remotion/transitions/fade";
import { dissolve } from "@remotion/transitions/dissolve";

import { sceneRegistry } from "./scenes";
import { Hud } from "./components/Hud";
import {
  TOTAL_SOURCE_FRAMES,
  videoConfig,
  type TransitionKind,
} from "./config/video.config";

const PRESENTATION_BY_KIND: Record<
  TransitionKind,
  () => TransitionPresentation<Record<string, unknown>>
> = {
  fade: () => fade(),
  // Dissolve — single hairline burn between scenes. lineWidth kept
  // tiny so the effect reads as a refined glow, not a flame.
  dissolve: () => dissolve({ lineWidth: 4, intensity: 0.6 }),
};

export const Video: React.FC = () => {
  return (
    <AbsoluteFill style={{ background: videoConfig.palette.bg }}>
      <TransitionSeries
        style={{
          translate: "5px 0px",
        }}
      >
        {sceneRegistry.map(({ Component, meta }, i) => {
          const isLast = i === sceneRegistry.length - 1;
          const kind: TransitionKind = meta.transitionKind ?? "fade";
          return (
            <Fragment key={meta.id}>
              <TransitionSeries.Sequence
                durationInFrames={meta.durationInFrames}
              >
                <Component durationInFrames={meta.durationInFrames} />
              </TransitionSeries.Sequence>
              {!isLast && (
                <TransitionSeries.Transition
                  presentation={PRESENTATION_BY_KIND[kind]()}
                  timing={linearTiming({
                    durationInFrames:
                      videoConfig.motion.defaultTransitionFrames,
                  })}
                />
              )}
            </Fragment>
          );
        })}
      </TransitionSeries>
      <Hud sceneTitle={videoConfig.meta.title} />
    </AbsoluteFill>
  );
};

/** Composition duration used by `Root.tsx`. Exported for the registry. */
export const VIDEO_DURATION_IN_FRAMES = TOTAL_SOURCE_FRAMES;
