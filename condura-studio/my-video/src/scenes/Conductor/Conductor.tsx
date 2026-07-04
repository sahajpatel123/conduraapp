// src/scenes/Conductor/Conductor.tsx
//
// Scene 3 — The Conductor. 10 seconds.
//
// Beat sheet:
//   •  0–48f   Vertical rail scales from 0 height to full
//   • 56–168f  Four numbered steps appear (staggered 28f)
//   • 56–168f  Azurite square hops each step as it appears
//   • 168–232f Display headline resolves
//   • 232–268f Body subline resolves
//   • 268–285f BREATH BEAT
//   • 285–300f Exit (aligned to transition)
//
// HUMAN MOMENT: the azurite square leaves a 3-frame "ghost trail" —
// three small azurite dots at decreasing opacity at its previous
// positions. Reads as motion blur, but with restraint. No other scene
// in the film does this; the audience feels an intentional gesture.
//
// LAYOUT FIX (was overflowing at 1260px in the previous version):
// RAIL_HEIGHT = 540 placed the rail 180px off-canvas. RAIL_HEIGHT is
// now 360 with RAIL_TOP = 580, fitting comfortably between the headline
// (which ends ~370) and the bottom HUD safe zone (~980).

import { AbsoluteFill, interpolate, useCurrentFrame } from "remotion";

import { videoConfig } from "../../config/video.config";
import { CornerBrackets } from "../../components/CornerBrackets";
import { FilmGrain } from "../../components/FilmGrain";
import { Body, Display } from "../../components/Typography";
import type { SceneProps } from "../types";

const STEPS = [
  { n: "01", label: "Context", body: "What's already on your screen" },
  { n: "02", label: "Route", body: "The right model or agent" },
  { n: "03", label: "Action", body: "Prepared, never hidden" },
  { n: "04", label: "Receipt", body: "What happened, saved clearly" },
] as const;

const PAGE_INSET = 96;
const RAIL_LEFT = 96 + 80;
const RAIL_TOP = 580; // FIXED — was 720 (rail overflowed canvas)
const RAIL_HEIGHT = 360; // FIXED — was 540 (rail overflowed canvas)
const STEP_HEIGHT = 90;

export const Conductor = ({ durationInFrames }: SceneProps) => {
  const frame = useCurrentFrame();
  const TRANSITION = 15;
  const exit = interpolate(
    frame,
    [durationInFrames - TRANSITION, durationInFrames],
    [1, 0],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" },
  );

  const rail = interpolate(frame, [0, 48], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const railOp = rail * exit;

  const stepAppear = (i: number) =>
    interpolate(frame, [56 + i * 28, 92 + i * 28], [0, 1], {
      extrapolateLeft: "clamp",
      extrapolateRight: "clamp",
    });

  // Azurite square hops
  const activeStep = Math.min(
    STEPS.length - 1,
    Math.max(0, Math.floor((frame - 56) / 28)),
  );
  const activeFrac = Math.min(1, Math.max(0, ((frame - 56) % 28) / 28));
  const squareY = interpolate(
    activeStep + activeFrac,
    [0, STEPS.length - 1],
    [0, RAIL_HEIGHT - 12],
  );

  // HUMAN MOMENT: trailing dots at the square's previous position
  // 3 frames ago and 6 frames ago. Decreasing opacity (0.32 → 0.12).
  const trailY1 = interpolate(
    Math.max(0, activeStep + Math.max(0, ((frame - 56 - 3) % 28) / 28) - 1),
    [0, STEPS.length - 1],
    [0, RAIL_HEIGHT - 12],
  );
  const trailY2 = interpolate(
    Math.max(0, activeStep + Math.max(0, ((frame - 56 - 6) % 28) / 28) - 2),
    [0, STEPS.length - 1],
    [0, RAIL_HEIGHT - 12],
  );

  const headlineRise = interpolate(frame, [168, 232], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const headlineY = (1 - headlineRise) * 12;
  const headlineOp = headlineRise * exit;

  const bodyOp = interpolate(frame, [204, 240], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });

  const bracketIn = interpolate(frame, [0, 24], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const bracketOp = bracketIn * exit;

  return (
    <AbsoluteFill style={{ background: videoConfig.palette.bg }}>
      <CornerBrackets inset={PAGE_INSET} style={{ opacity: bracketOp }} />

      <div
        style={{
          position: "absolute",
          left: PAGE_INSET,
          right: PAGE_INSET,
          top: PAGE_INSET + 4,
          opacity: bracketIn,
          display: "flex",
          justifyContent: "space-between",
        }}
      >
        <span style={eyebrowStyle}>Section III · Conductor</span>
        <span style={eyebrowStyle}>04 steps</span>
      </div>

      {/* Headline */}
      <div
        style={{
          position: "absolute",
          left: PAGE_INSET,
          right: PAGE_INSET,
          top: 200,
          opacity: headlineOp,
          transform: `translateY(${headlineY}px)`,
        }}
      >
        <Display style={{ fontSize: 124, lineHeight: 1.0 }}>
          One request. Many hands.
        </Display>
        <Body style={{ marginTop: 28, maxWidth: 720, fontSize: 22, opacity: bodyOp }}>
          The user gives one prompt. Condura splits the work, picks the
          tools, and returns a single receipt.
        </Body>
      </div>

      {/* Timeline rail — FIXED position to fit canvas */}
      <div
        style={{
          position: "absolute",
          left: RAIL_LEFT,
          top: RAIL_TOP,
          width: 2,
          height: RAIL_HEIGHT,
          background: videoConfig.palette.line,
          opacity: railOp,
        }}
        aria-hidden
      >
        {/* Trailing dots (HUMAN MOMENT) */}
        <div
          aria-hidden
          style={{
            position: "absolute",
            top: trailY1,
            left: -2,
            width: 6,
            height: 6,
            background: videoConfig.palette.accent,
            opacity: 0.32,
          }}
        />
        <div
          aria-hidden
          style={{
            position: "absolute",
            top: trailY2,
            left: -1,
            width: 4,
            height: 4,
            background: videoConfig.palette.accent,
            opacity: 0.12,
          }}
        />
        {/* Active square */}
        <div
          style={{
            position: "absolute",
            top: squareY,
            left: -5,
            width: 12,
            height: 12,
            background: videoConfig.palette.accent,
          }}
        />
      </div>

      {/* Step rows — FIXED: bounded by STEP_HEIGHT × STEPS.length */}
      <div
        style={{
          position: "absolute",
          left: RAIL_LEFT + 32,
          top: RAIL_TOP,
          width: 760,
        }}
      >
        {STEPS.map((s, i) => {
          const op = stepAppear(i);
          const isActive = i <= activeStep;
          return (
            <div
              key={s.n}
              style={{
                height: STEP_HEIGHT,
                display: "grid",
                gridTemplateColumns: "60px 1fr 1.5fr",
                columnGap: 32,
                alignItems: "center",
                borderBottom:
                  i < STEPS.length - 1
                    ? `1px solid ${videoConfig.palette.line}`
                    : undefined,
                opacity: op,
                transform: `translateX(${interpolate(op, [0, 1], [16, 0])}px)`,
              }}
            >
              <span
                style={{
                  fontFamily: videoConfig.type.monoFamily,
                  fontSize: 13,
                  fontWeight: 600,
                  letterSpacing: videoConfig.type.tracking.eyebrow,
                  color: isActive
                    ? videoConfig.palette.accent
                    : videoConfig.palette.inkMutedSoft,
                }}
              >
                {s.n}
              </span>
              <span
                style={{
                  fontFamily: videoConfig.type.displayFamily,
                  fontSize: 32,
                  fontWeight: 200,
                  letterSpacing: videoConfig.type.tracking.display,
                  color: videoConfig.palette.ink,
                }}
              >
                {s.label}
              </span>
              <span
                style={{
                  fontFamily: videoConfig.type.bodyFamily,
                  fontSize: 18,
                  lineHeight: 1.4,
                  color: videoConfig.palette.inkMuted,
                }}
              >
                {s.body}
              </span>
            </div>
          );
        })}
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
