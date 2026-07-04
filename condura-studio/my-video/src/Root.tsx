// src/Root.tsx
//
// Composition registry. One composition — the launch film — whose
// dimensions, fps, and duration come from the config. Adding future
// compositions (e.g. teaser cuts at 9:16) is a new <Composition> block,
// not a new root.

import { Composition } from "remotion";

import { Video, VIDEO_DURATION_IN_FRAMES } from "./Video";
import {
  videoConfig,
  NET_DURATION_SECONDS,
  TOTAL_SOURCE_FRAMES,
  NET_DURATION_FRAMES,
} from "./config/video.config";
import "./styles/launch.css";

export const RemotionRoot: React.FC = () => {
  return (
    <>
      <Composition
        id={videoConfig.composition.id}
        component={Video}
        durationInFrames={VIDEO_DURATION_IN_FRAMES}
        fps={videoConfig.composition.fps}
        width={videoConfig.composition.width}
        height={videoConfig.composition.height}
      />
    </>
  );
};

// Sanity-check the math at module load — fails loudly on bad config.
if (process.env.NODE_ENV !== "production") {
  console.info(
    `[condura-launch] source=${TOTAL_SOURCE_FRAMES}f ` +
      `overlap=${TOTAL_SOURCE_FRAMES - NET_DURATION_FRAMES}f ` +
      `net=${NET_DURATION_FRAMES}f ` +
      `(${NET_DURATION_SECONDS.toFixed(2)}s)`,
  );
}
