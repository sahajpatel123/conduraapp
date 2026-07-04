// src/scenes/ColdOpen/ColdOpen.tsx
//
// Scene 1 — Cold Open. 8 seconds.
//
// Beat sheet (all windows snap to 12-frame boundaries — no fractional
// timing, easier to scrub):
//   •  0–36f   Corner brackets sweep
//   • 24–72f   Center hairline rule scales from middle
//   • 60–120f  Display headline resolves, rises 12px
//   • 108–156f Body subline fades up
//   • 144–180f Azurite square settles (small 4f "click" beat at peak)
//   • 168–216f Mark fades + scales from 0.92 → 1
//   • 192–216f BREATH BEAT (no motion) — eye reads
//   • 216–240f Exit aligned with transition overlap (15f)
//   •  60f     Print-mark "02 / 06" appears (human moment #1)
//
// HUMAN MOMENT: a tiny page-fraction label "02 / 06" sits in the
// upper-right of the canvas — a printer's signature. Off-grid from
// the headline rhythm. Reads as a hand, not a system.

import { AbsoluteFill, interpolate, useCurrentFrame } from "remotion";

import { videoConfig } from "../../config/video.config";
import { CornerBrackets } from "../../components/CornerBrackets";
import { FilmGrain } from "../../components/FilmGrain";
import { Mark } from "../../components/Mark";
import { Body, Display } from "../../components/Typography";
import type { SceneProps } from "../types";

const PAGE_INSET = 96;
const TRANSITION_FRAMES = 15; // matches videoConfig.motion.defaultTransitionFrames

export const ColdOpen = ({ durationInFrames }: SceneProps) => {
  const frame = useCurrentFrame();

  // ---- Exit: aligned with transition overlap (15 frames) ----
  const exit = interpolate(
    frame,
    [durationInFrames - TRANSITION_FRAMES, durationInFrames],
    [1, 0],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" },
  );

  // ---- Brackets sweep ----
  const bracketSweep = interpolate(frame, [0, 36], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const bracketOp = bracketSweep * exit;

  // ---- Center hairline rule ----
  const rule = interpolate(frame, [24, 72], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });

  // ---- Display headline ----
  const headlineRise = interpolate(frame, [60, 120], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const headlineY = (1 - headlineRise) * 12;
  const headlineOp = headlineRise * exit;

  // ---- Body subline ----
  const bodyOp = interpolate(frame, [108, 156], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });

  // ---- Azurite square: settles with a 4-frame "click" at peak ----
  // The "click" is implemented by snapping opacity from 0.95 → 1.0
  // (over 4 frames) right at the peak. Subliminal, almost imperceptible
  // — but it gives the square the feel of being placed, not faded in.
  const accentPre = interpolate(frame, [144, 176], [0, 0.95], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const accentSettle = interpolate(frame, [176, 180], [0.95, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const accentSquare = accentPre + accentSettle;

  // ---- Mark ----
  const markAppear = interpolate(frame, [168, 216], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const markScale = 0.92 + markAppear * 0.08;
  const markOp = markAppear * exit;

  // ---- Print-mark "02 / 06" (human moment #1) ----
  const printMark = interpolate(frame, [60, 120], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const printMarkOp = printMark * exit;

  return (
    <AbsoluteFill style={{ background: videoConfig.palette.bg }}>
      <CornerBrackets inset={PAGE_INSET} style={{ opacity: bracketOp }} />

      {/* Top eyebrow + page-fraction (HUMAN MOMENT) */}
      <div
        style={{
          position: "absolute",
          left: PAGE_INSET,
          right: PAGE_INSET,
          top: PAGE_INSET + 4,
          display: "flex",
          justifyContent: "space-between",
          alignItems: "baseline",
        }}
      >
        <span
          style={{
            opacity: bracketSweep,
            fontFamily: videoConfig.type.monoFamily,
            fontSize: 13,
            fontWeight: 600,
            letterSpacing: videoConfig.type.tracking.eyebrow,
            textTransform: "uppercase",
            color: videoConfig.palette.inkMutedSoft,
          }}
        >
          Pre-launch · I
        </span>
        <span
          style={{
            opacity: printMarkOp,
            fontFamily: videoConfig.type.monoFamily,
            fontSize: 11,
            fontWeight: 600,
            letterSpacing: videoConfig.type.tracking.eyebrow,
            textTransform: "uppercase",
            color: videoConfig.palette.accent,
          }}
        >
          02 / 06
        </span>
      </div>

      {/* Center hairline rule */}
      <div
        aria-hidden
        style={{
          position: "absolute",
          left: PAGE_INSET,
          right: PAGE_INSET,
          top: "50%",
          height: 1,
          background: videoConfig.palette.line,
          transform: `scaleX(${rule}) translateY(-32px)`,
          transformOrigin: "center center",
        }}
      />

      {/* Display headline — two explicit lines so layout is deterministic */}
      <div
        style={{
          position: "absolute",
          left: PAGE_INSET,
          right: PAGE_INSET,
          top: "calc(50% - 180px)",
          transform: `translateY(${headlineY}px)`,
          opacity: headlineOp,
        }}
      >
        <Display
          style={{
            fontSize: 132,
            lineHeight: 1,
            maxWidth: 1620,
          }}
        >
          An operating system
          <br />
          for thought.
        </Display>
      </div>

      {/* Azurite square accent — snapped at peak (the "click") */}
      <div
        aria-hidden
        style={{
          position: "absolute",
          right: PAGE_INSET,
          top: PAGE_INSET + 200,
          width: 12,
          height: 12,
          background: videoConfig.palette.accent,
          opacity: accentSquare,
        }}
      />

      {/* Body subline — pushed below the headline's bottom line */}
      <div
        style={{
          position: "absolute",
          left: PAGE_INSET,
          right: PAGE_INSET,
          top: "calc(50% + 110px)",
          opacity: bodyOp * exit,
          transform: `translateY(${(1 - bodyOp) * 8}px)`,
        }}
      >
        <Body style={{ maxWidth: 760, fontSize: 22 }}>
          Condura. A command layer above the tools you already use —
          locally, by default.
        </Body>
      </div>

      {/* Mark, bottom-right */}
      <div
        style={{
          position: "absolute",
          right: PAGE_INSET,
          bottom: PAGE_INSET + 80,
          opacity: markOp,
          transform: `scale(${markScale})`,
        }}
      >
        <Mark size="quiet" />
      </div>

      <FilmGrain />
    </AbsoluteFill>
  );
};
