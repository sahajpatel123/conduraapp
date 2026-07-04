// src/scenes/Close/Close.tsx
//
// Scene 6 — Brand Close. ~6.5 seconds.
//
// Beat sheet:
//   •  0–28f   Brackets sweep in
//   • 24–72f   Mark expands via gentle spring (slight 1.06 ratio)
//   • 72–128f  "Condura." headline resolves
//   • 124–168f Azurite dot + URL resolves
//   • 152–180f BREATH BEAT (everything stops)
//   • 180–195f Exit aligned to transition
//
// HUMAN MOMENT: the azurite dot above the URL "breathes" — a 60-frame
// sine envelope on its opacity (0.7 → 1.0 → 0.7). The only motion in
// the final beat. Reads as a heartbeat — present, alive, not a logo.

import { AbsoluteFill, interpolate, useCurrentFrame } from "remotion";

import { videoConfig } from "../../config/video.config";
import { CornerBrackets } from "../../components/CornerBrackets";
import { FilmGrain } from "../../components/FilmGrain";
import { Mark } from "../../components/Mark";
import { Display } from "../../components/Typography";
import { useRefinedSpring } from "../../hooks/useRefinedSpring";
import type { SceneProps } from "../types";

const PAGE_INSET = 96;

export const Close = ({ durationInFrames }: SceneProps) => {
  const frame = useCurrentFrame();
  const TRANSITION = 15;
  const exit = interpolate(
    frame,
    [durationInFrames - TRANSITION, durationInFrames],
    [1, 0],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" },
  );

  // Mark expands with overdamped gentle spring
  const settle = useRefinedSpring({ token: "gentle", delayFrames: 14 });

  const headlineRise = interpolate(frame, [72, 128], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const headlineY = (1 - headlineRise) * 12;
  const headlineOp = headlineRise * exit;

  const urlRise = interpolate(frame, [124, 168], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const urlY = (1 - urlRise) * 8;
  const urlOp = urlRise * exit;

  // HUMAN MOMENT: azurite dot breathes (60-frame period)
  const breathPhase = (frame - 124) / 60;
  const breathing = (frame > 124 && frame < 180)
    ? 0.7 + 0.3 * Math.sin(breathPhase * Math.PI * 2)
    : 1;

  const accentDot = interpolate(frame, [124, 168], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const accentOp = accentDot * urlOp * breathing;

  const bracketIn = interpolate(frame, [0, 28], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const bracketOp = bracketIn * exit;

  const markExit = interpolate(
    frame,
    [durationInFrames - 18, durationInFrames - 4],
    [1, 0.5],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" },
  );
  const headlineExit = interpolate(
    frame,
    [durationInFrames - 18, durationInFrames - 4],
    [1, 0.92],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" },
  );

  return (
    <AbsoluteFill style={{ background: videoConfig.palette.bg }}>
      <CornerBrackets inset={PAGE_INSET} style={{ opacity: bracketOp }} />

      <div
        style={{
          position: "absolute",
          left: PAGE_INSET,
          right: PAGE_INSET,
          top: PAGE_INSET + 4,
          opacity: bracketOp,
          display: "flex",
          justifyContent: "space-between",
        }}
      >
        <span style={eyebrowStyle}>Section VI · Close</span>
        <span style={{ ...eyebrowStyle, color: videoConfig.palette.accent }}>
          condura.app
        </span>
      </div>

      {/* Mark */}
      <div
        style={{
          position: "absolute",
          left: "50%",
          top: 380,
          transform: `translate(-50%, -50%) scale(${0.94 + settle * 0.06})`,
          opacity: settle * markExit,
        }}
      >
        <Mark />
      </div>

      {/* Final word */}
      <div
        style={{
          position: "absolute",
          left: PAGE_INSET,
          right: PAGE_INSET,
          top: 560,
          textAlign: "center",
          opacity: headlineOp * headlineExit,
          transform: `translateY(${headlineY}px)`,
        }}
      >
        <Display style={{ fontSize: 200, lineHeight: 0.96, letterSpacing: "-0.04em" }}>
          Condura.
        </Display>
      </div>

      {/* Azurite dot + URL (the breathing dot is the human moment) */}
      <div
        style={{
          position: "absolute",
          left: "50%",
          bottom: 240,
          transform: `translateX(-50%) translateY(${urlY}px)`,
          opacity: urlOp,
          display: "grid",
          justifyItems: "center",
          gap: 18,
        }}
      >
        <span
          aria-hidden
          style={{
            width: 6,
            height: 6,
            borderRadius: "50%",
            background: videoConfig.palette.accent,
            opacity: accentOp,
            transform: `scale(${0.92 + breathing * 0.16})`,
          }}
        />
        <span
          style={{
            fontFamily: videoConfig.type.monoFamily,
            fontSize: 16,
            fontWeight: 600,
            letterSpacing: videoConfig.type.tracking.eyebrow,
            textTransform: "uppercase",
            color: videoConfig.palette.ink,
          }}
        >
          condura.app
        </span>
      </div>

      <FilmGrain />
    </AbsoluteFill>
  );
};

const eyebrowStyle: React.CSSProperties = {
  fontFamily: videoConfig.type.monoFamily,
  fontSize: 13,
  fontWeight: 600,
  lineHeight: 1,
  letterSpacing: videoConfig.type.tracking.eyebrow,
  textTransform: "uppercase",
  color: videoConfig.palette.inkMutedSoft,
};
