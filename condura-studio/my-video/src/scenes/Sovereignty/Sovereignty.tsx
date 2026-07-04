// src/scenes/Sovereignty/Sovereignty.tsx
//
// Scene 4 — Local Sovereignty. 10 seconds.
//
// Beat sheet:
//   •  0–48f   Left vertical azurite line scales from y=0 to full height
//   • 48–112f  Display headline "Local by default." resolves
//   • 112–240f Three manifesto rows appear (staggered 36f)
//   • 220–256f Azurite marker on row 3 settles in
//   • 256–285f BREATH BEAT
//   • 285–300f Exit aligned to transition
//
// HUMAN MOMENT: the third row's hairline rule is *azurite*, not gray,
// and a 4×12 tick extends 1px below it — a printer's registration mark.
// The eye finds it without being told, and once it does, the rule
// itself reads as authored, not auto-generated. Plus: a tiny "APPROVE"
// eyebrow label appears in azurite above row 3.

import { AbsoluteFill, interpolate, useCurrentFrame } from "remotion";

import { videoConfig } from "../../config/video.config";
import { CornerBrackets } from "../../components/CornerBrackets";
import { FilmGrain } from "../../components/FilmGrain";
import { Body, Display } from "../../components/Typography";
import type { SceneProps } from "../types";

const PRINCIPLES = [
  { n: "01", line: "Local when it can." },
  { n: "02", line: "Cloud only when you choose." },
  { n: "03", line: "Approval before any action." },
] as const;

const PAGE_INSET = 96;
const PRINCIPLES_LEFT = 96 + 96;

export const Sovereignty = ({ durationInFrames }: SceneProps) => {
  const frame = useCurrentFrame();
  const TRANSITION = 15;
  const exit = interpolate(
    frame,
    [durationInFrames - TRANSITION, durationInFrames],
    [1, 0],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" },
  );

  const verticalRise = interpolate(frame, [0, 48], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const verticalOp = verticalRise * exit;

  const headlineRise = interpolate(frame, [48, 112], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const headlineY = (1 - headlineRise) * 14;
  const headlineOp = headlineRise * exit;

  const rowRise = (i: number) =>
    interpolate(frame, [120 + i * 36, 156 + i * 36], [0, 1], {
      extrapolateLeft: "clamp",
      extrapolateRight: "clamp",
    });

  const thirdRowAccent = interpolate(frame, [220, 256], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const thirdRowAccentOp = thirdRowAccent * exit;

  // Approval eyebrow (HUMAN MOMENT) on row 3
  const approveEyebrow = interpolate(frame, [232, 268], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const approveOp = approveEyebrow * exit;

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
        <span style={eyebrowStyle}>Section IV · Sovereignty</span>
        <span style={eyebrowStyle}>3 principles</span>
      </div>

      {/* Vertical azurite line */}
      <div
        aria-hidden
        style={{
          position: "absolute",
          left: PRINCIPLES_LEFT,
          top: 220,
          width: 1,
          height: 480,
          background: videoConfig.palette.accent,
          transform: `scaleY(${verticalRise})`,
          transformOrigin: "top center",
          opacity: verticalOp,
        }}
      />

      {/* Display headline */}
      <div
        style={{
          position: "absolute",
          left: PRINCIPLES_LEFT,
          top: 180,
          opacity: headlineOp,
          transform: `translateY(${headlineY}px)`,
        }}
      >
        <Display style={{ fontSize: 124 }}>Local by default.</Display>
        <Body style={{ marginTop: 28, fontSize: 22, maxWidth: 660 }}>
          A command layer that respects the machine it's running on.
        </Body>
      </div>

      {/* Manifesto rows */}
      <div
        style={{
          position: "absolute",
          left: PRINCIPLES_LEFT,
          right: PAGE_INSET,
          top: 520,
        }}
      >
        {PRINCIPLES.map((p, i) => {
          const op = rowRise(i) * exit;
          const isApproved = i === PRINCIPLES.length - 1;
          return (
            <div
              key={p.n}
              style={{
                position: "relative",
                padding: "44px 0",
                borderBottom:
                  i < PRINCIPLES.length - 1
                    ? `1px solid ${videoConfig.palette.line}`
                    : undefined,
                opacity: op,
                transform: `translateY(${interpolate(op, [0, 1], [16, 0])}px)`,
              }}
            >
              {/* HUMAN-MOMENT: third row's hairline is azurite + has a
                  4×12 tick extending below — printer's registration mark */}
              {isApproved ? (
                <>
                  <div
                    aria-hidden
                    style={{
                      position: "absolute",
                      left: -36,
                      right: 0,
                      bottom: 0,
                      height: 1,
                      background: videoConfig.palette.accent,
                      opacity: thirdRowAccentOp,
                    }}
                  />
                  <div
                    aria-hidden
                    style={{
                      position: "absolute",
                      left: -36,
                      bottom: 0,
                      width: 4,
                      height: 12,
                      background: videoConfig.palette.accent,
                      opacity: thirdRowAccentOp,
                    }}
                  />
                  {/* APPROVE eyebrow above the row */}
                  <div
                    style={{
                      position: "absolute",
                      left: 88,
                      top: -28,
                      fontFamily: videoConfig.type.monoFamily,
                      fontSize: 10,
                      fontWeight: 600,
                      letterSpacing: videoConfig.type.tracking.eyebrow,
                      textTransform: "uppercase",
                      color: videoConfig.palette.accent,
                      opacity: approveOp,
                    }}
                  >
                    Approve
                  </div>
                </>
              ) : null}
              <div
                style={{
                  display: "grid",
                  gridTemplateColumns: "60px 1fr",
                  alignItems: "baseline",
                  columnGap: 28,
                }}
              >
                <span
                  style={{
                    fontFamily: videoConfig.type.monoFamily,
                    fontSize: 12,
                    fontWeight: 600,
                    letterSpacing: videoConfig.type.tracking.eyebrow,
                    color: isApproved
                      ? videoConfig.palette.accent
                      : videoConfig.palette.inkMutedSoft,
                  }}
                >
                  {p.n}
                </span>
                <span
                  style={{
                    fontFamily: videoConfig.type.displayFamily,
                    fontSize: 56,
                    fontWeight: 200,
                    lineHeight: 1.1,
                    letterSpacing: videoConfig.type.tracking.display,
                    color: videoConfig.palette.ink,
                  }}
                >
                  {p.line}
                </span>
              </div>
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
