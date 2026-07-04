// src/components/Hud.tsx
//
// Bottom chrome, white-room treatment. Project name left, scene title
// center (small caps, wide tracking), azurite progress dot inside the
// hairline rule, seconds counter right.
//
// The Hud reads `useCurrentFrame()` — its time reflects the *global*
// composition clock, which is what the renderer is doing. Within
// TransitionSeries.Sequence, frames are local; outside a sequence they
// are global. The Hud is rendered inside the AbsoluteFill *outside* any
// sequence, so its frame is global.

import { useCurrentFrame } from "remotion";

import {
  TOTAL_SOURCE_FRAMES,
  videoConfig,
} from "../config/video.config";

export function Hud({ sceneTitle }: { sceneTitle?: string }) {
  const frame = useCurrentFrame();
  const seconds = (frame / videoConfig.composition.fps).toFixed(1);
  const progress = frame / TOTAL_SOURCE_FRAMES;

  return (
    <div
      style={{
        position: "absolute",
        left: 56,
        right: 56,
        bottom: 40,
        zIndex: 50,
        display: "grid",
        gridTemplateColumns: "auto 1fr auto",
        alignItems: "center",
        gap: 32,
        fontFamily: videoConfig.type.monoFamily,
        fontSize: 10,
        fontWeight: 600,
        letterSpacing: videoConfig.type.tracking.eyebrow,
        textTransform: "uppercase",
        color: videoConfig.palette.inkMutedSoft,
      }}
    >
      <span style={{ whiteSpace: "nowrap" }}>{videoConfig.meta.project}</span>
      <div
        style={{
          position: "relative",
          height: 1,
          background: videoConfig.palette.line,
        }}
      >
        {/* Single azurite dot rides the rule */}
        <span
          aria-hidden
          style={{
            position: "absolute",
            left: `${progress * 100}%`,
            top: "-3px",
            width: 7,
            height: 7,
            borderRadius: "50%",
            background: videoConfig.palette.accent,
            transform: "translateX(-50%)",
          }}
        />
        {sceneTitle ? (
          <span
            style={{
              position: "absolute",
              left: "50%",
              top: -22,
              transform: "translateX(-50%)",
              whiteSpace: "nowrap",
              color: videoConfig.palette.inkMuted,
            }}
          >
            {sceneTitle}
          </span>
        ) : null}
      </div>
      <span style={{ whiteSpace: "nowrap" }}>{`${seconds}s`}</span>
    </div>
  );
}
