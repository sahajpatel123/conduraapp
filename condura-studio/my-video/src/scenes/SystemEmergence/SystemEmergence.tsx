// src/scenes/SystemEmergence/SystemEmergence.tsx
//
// Scene 2 — System Emergence. 10 seconds.
//
// Beat sheet:
//   •  0–48f   72 hairline gray dots materialize (staggered)
//   • 48–156f  Wave sweeps left→right; columns light briefly in azurite
//   • 168–232f 12 PROVIDERS stay lit (in a 6×2 subset of the grid)
//   • 200–256f Display headline resolves (124px hairline)
//   • 232–272f Body subline resolves
//   • 256–285f BREATH BEAT
//   • 285–300f Exit (aligned to transition)
//
// HUMAN MOMENT: one dot in the grid — column 7, row 2 — is TWICE the
// size of every other dot, has a soft azurite halo, and stays lit
// from the wave onward. It reads as a marked cell, an editorial
// asterisk — a hand's intention in an otherwise perfect grid.
//
// PERFORMANCE: dots are rendered as 72 stable-keyed <circle> children
// inside one <g>. React reconciles in place (no unmount), so per-frame
// work is one style patch per dot. Browsers paint 72 tiny SVG circles
// in a single layer pass.

import { AbsoluteFill, interpolate, useCurrentFrame } from "remotion";
import type { CSSProperties } from "react";

import { videoConfig } from "../../config/video.config";
import { CornerBrackets } from "../../components/CornerBrackets";
import { FilmGrain } from "../../components/FilmGrain";
import { Mark } from "../../components/Mark";
import { Body, Display } from "../../components/Typography";
import type { SceneProps } from "../types";

// ----------------------------------------------------------------------
// Geometry — generated once at module scope.
// ----------------------------------------------------------------------

const COLS = 12;
const ROWS = 6;
const CELL = 64;
const GRID_W = COLS * CELL;
const GRID_H = ROWS * CELL;
const GRID_LEFT = (videoConfig.composition.width - GRID_W) / 2;
const GRID_TOP = (videoConfig.composition.height - GRID_H) / 2 - 24;

// 12 provider indexes (row-major) — the lit subset.
const PROVIDER_INDEXES = new Set([
  4, 9, 16, 19, 22, 27, 32, 35, 40, 43, 50, 53,
]);
// HUMAN-MOMENT marked cell: column 7, row 2.
const HUMAN_MOMENT_ID = 2 * COLS + 7; // = 31

const WAVE_START = 48;
const WAVE_STEP = 96 / COLS;

type Dot = {
  id: number;
  col: number;
  row: number;
  x: number;
  y: number;
  r: number;
};

const dots: Dot[] = (() => {
  const out: Dot[] = [];
  let id = 0;
  for (let row = 0; row < ROWS; row++) {
    for (let col = 0; col < COLS; col++) {
      out.push({
        id: id++,
        col,
        row,
        x: GRID_LEFT + col * CELL + CELL / 2,
        y: GRID_TOP + row * CELL + CELL / 2,
        r: id - 1 === HUMAN_MOMENT_ID ? 5 : 1.8,
      });
    }
  }
  return out;
})();

// Precompute human-moment halo and a stable style.
const HALO_RADIUS = 11;

// ----------------------------------------------------------------------
// Component
// ----------------------------------------------------------------------

export const SystemEmergence = ({ durationInFrames }: SceneProps) => {
  const frame = useCurrentFrame();
  const TRANSITION = 15;
  const exit = interpolate(
    frame,
    [durationInFrames - TRANSITION, durationInFrames],
    [1, 0],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" },
  );

  const gridAppear = interpolate(frame, [0, 48], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });

  const headlineRise = interpolate(frame, [200, 256], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const headlineY = (1 - headlineRise) * 12;
  const headlineOp = headlineRise * exit;

  const bodyOp = interpolate(frame, [232, 272], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });

  const markAppear = interpolate(frame, [248, 276], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const markOp = markAppear * exit;

  const bracketIn = interpolate(frame, [0, 32], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const bracketOp = bracketIn * exit;

  const accentSquare = interpolate(frame, [212, 244], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });

  // After wave passes, default-gray dots become visible.
  const showGray = frame > WAVE_START + (COLS - 1) * WAVE_STEP + 24;

  return (
    <AbsoluteFill style={{ background: videoConfig.palette.bg }}>
      <CornerBrackets inset={96} style={{ opacity: bracketOp }} />

      {/* Eyebrow row */}
      <div
        style={{
          position: "absolute",
          left: 96,
          right: 96,
          top: 96,
          opacity: bracketIn,
          display: "flex",
          justifyContent: "space-between",
        }}
      >
        <span style={eyebrowStyle}>Section II · System</span>
        <span style={eyebrowStyle}>12 / 8 / ∞</span>
      </div>

      {/* SVG grid — 72 stable-keyed circles, React reconciles in place */}
      <svg
        viewBox={`0 0 ${videoConfig.composition.width} ${videoConfig.composition.height}`}
        style={{ position: "absolute", inset: 0, width: "100%", height: "100%" }}
      >
        <g style={{ opacity: gridAppear }}>
          {dots.map((d) => {
            const waveAt = WAVE_START + d.col * WAVE_STEP;
            const isInWaveWindow =
              frame >= waveAt - 12 && frame <= waveAt + 12;
            const env = isInWaveWindow
              ? Math.max(0, 1 - Math.abs((frame - waveAt) / 12))
              : 0;

            const isProvider = PROVIDER_INDEXES.has(d.id);
            const isMarked = d.id === HUMAN_MOMENT_ID;
            const stayLit = isProvider && frame > waveAt + 8;

            // Wave-active takes priority (visual on top of base)
            if (env > 0.2) {
              return (
                <circle
                  key={d.id}
                  cx={d.x}
                  cy={d.y}
                  r={d.r}
                  fill={videoConfig.palette.accent}
                  fillOpacity={env * 0.92}
                />
              );
            }
            // Provider lit
            if (stayLit) {
              return (
                <circle
                  key={d.id}
                  cx={d.x}
                  cy={d.y}
                  r={d.r}
                  fill={videoConfig.palette.accent}
                  fillOpacity={0.95}
                />
              );
            }
            // HUMAN-MOMENT marked cell: azurite with soft halo
            if (isMarked && frame > WAVE_START + 24) {
              return (
                <g key={d.id}>
                  <circle
                    cx={d.x}
                    cy={d.y}
                    r={HALO_RADIUS}
                    fill={videoConfig.palette.accent}
                    fillOpacity={0.12}
                  />
                  <circle
                    cx={d.x}
                    cy={d.y}
                    r={d.r}
                    fill={videoConfig.palette.accent}
                    fillOpacity={0.92}
                  />
                </g>
              );
            }
            // Default gray
            if (showGray) {
              return (
                <circle
                  key={d.id}
                  cx={d.x}
                  cy={d.y}
                  r={d.r}
                  fill={videoConfig.palette.ink}
                  fillOpacity={0.18}
                />
              );
            }
            return null;
          })}
        </g>
      </svg>

      {/* Azurite accent */}
      <div
        aria-hidden
        style={{
          position: "absolute",
          right: 96,
          top: 220,
          width: 12,
          height: 12,
          background: videoConfig.palette.accent,
          opacity: accentSquare,
        }}
      />

      {/* Display headline + body */}
      <div
        style={{
          position: "absolute",
          left: 96,
          right: 96,
          bottom: 240,
          opacity: headlineOp,
          transform: `translateY(${headlineY}px)`,
        }}
      >
        <Display style={{ fontSize: 132, lineHeight: 1.0 }}>
          Twelve providers. Eight agents.
        </Display>
        <Body
          style={{
            marginTop: 28,
            maxWidth: 720,
            fontSize: 22,
            opacity: bodyOp,
          }}
        >
          One conductor above the rest.
        </Body>
      </div>

      {/* Mark (quiet) */}
      <div
        style={{
          position: "absolute",
          left: "50%",
          bottom: 96,
          transform: "translateX(-50%)",
          opacity: markOp,
        }}
      >
        <Mark size="quiet" />
      </div>

      <FilmGrain />
    </AbsoluteFill>
  );
};

const eyebrowStyle: CSSProperties = {
  fontFamily: videoConfig.type.monoFamily,
  fontSize: 13,
  fontWeight: 600,
  lineHeight: 1,
  letterSpacing: videoConfig.type.tracking.eyebrow,
  textTransform: "uppercase",
  color: videoConfig.palette.inkMutedSoft,
};
