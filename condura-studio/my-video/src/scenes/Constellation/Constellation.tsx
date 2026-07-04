// src/scenes/Constellation/Constellation.tsx
//
// Scene 5 — The Constellation. 10 seconds.
//
// Beat sheet:
//   •  0–24f   Brackets sweep in
//   •  24–96f  Eyebrow numerals (azurite) + headline resolve
//   • 112–240f 4×3 provider cells fade in (staggered 8f)
//   • 220–264f Footer rule + attribute keys fade in
//   • 256–285f BREATH BEAT
//   • 285–300f Exit aligned to transition
//
// HUMAN MOMENT: one specific cell — index 4 (second row, first column:
// "Mistral") — has azurite ticks at *all four corners*, not just one.
// Its borders are also azurite on top and left. Reads as "the active
// cell" — a hand has marked it.

import { AbsoluteFill, interpolate, useCurrentFrame } from "remotion";
import type { CSSProperties } from "react";

import { videoConfig } from "../../config/video.config";
import { CornerBrackets } from "../../components/CornerBrackets";
import { FilmGrain } from "../../components/FilmGrain";
import { Body, Display } from "../../components/Typography";
import type { SceneProps } from "../types";

const PROVIDERS = [
  "Anthropic",
  "OpenAI",
  "Google",
  "xAI",
  "Mistral",
  "DeepSeek",
  "OpenRouter",
  "Together",
  "Groq",
  "Fireworks",
  "Custom",
  "Ollama",
] as const;

const ACTIVE_CELL = 4;

const COLS = 4;
const PAGE_INSET = 96;

export const Constellation = ({ durationInFrames }: SceneProps) => {
  const frame = useCurrentFrame();
  const TRANSITION = 15;
  const exit = interpolate(
    frame,
    [durationInFrames - TRANSITION, durationInFrames],
    [1, 0],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" },
  );

  const headlineRise = interpolate(frame, [24, 96], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const headlineY = (1 - headlineRise) * 12;
  const headlineOp = headlineRise * exit;

  const bodyOp = interpolate(frame, [80, 116], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });

  const cellRise = (i: number) =>
    interpolate(frame, [112 + i * 8, 144 + i * 8], [0, 1], {
      extrapolateLeft: "clamp",
      extrapolateRight: "clamp",
    });

  const footer = interpolate(frame, [220, 264], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const footerOp = footer * exit;

  const bracketIn = interpolate(frame, [0, 24], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
  });
  const bracketOp = bracketIn * exit;

  // Cells layout
  const gridWidth = 720;
  const cellW = gridWidth / COLS;
  const cellH = 120;
  const cellsLeft = (videoConfig.composition.width - gridWidth) / 2;

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
        <span style={{ ...eyebrowStyle, color: videoConfig.palette.accent }}>
          Section V · Constellation
        </span>
        <span style={eyebrowStyle}>12 providers</span>
      </div>

      <div
        style={{
          position: "absolute",
          left: PAGE_INSET,
          right: PAGE_INSET,
          top: 180,
          opacity: headlineOp,
          transform: `translateY(${headlineY}px)`,
        }}
      >
        <Display style={{ fontSize: 132, lineHeight: 1.0 }}>
          Twelve. No lock-in.
        </Display>
        <Body style={{ marginTop: 28, fontSize: 22, maxWidth: 720, opacity: bodyOp }}>
          Use the subscription you already pay for. Bring your own key. Run
          it locally. We don't take a cut on the way through.
        </Body>
      </div>

      {/* Grid of cells */}
      <div
        style={{
          position: "absolute",
          left: cellsLeft,
          right: cellsLeft,
          top: 540,
        }}
      >
        {PROVIDERS.map((p, i) => {
          const op = cellRise(i) * exit;
          const isActive = i === ACTIVE_CELL;
          const col = i % COLS;
          const row = Math.floor(i / COLS);
          return (
            <div
              key={p}
              style={{
                position: "absolute",
                left: col * cellW,
                top: row * cellH,
                width: cellW - 1,
                height: cellH - 1,
                border: isActive
                  ? `1px solid ${videoConfig.palette.accent}`
                  : `1px solid ${videoConfig.palette.line}`,
                display: "grid",
                placeItems: "center",
                opacity: op,
                background: isActive
                  ? `rgba(27, 77, 255, 0.04)`
                  : "transparent",
              }}
            >
              {/* HUMAN-MOMENT active cell: ticks in all four corners */}
              {isActive ? (
                <>
                  <TickAt top={10} left={10} size={6} opacity={op} />
                  <TickAt top={10} right={10} size={6} opacity={op} />
                  <TickAt bottom={10} left={10} size={6} opacity={op} />
                  <TickAt bottom={10} right={10} size={6} opacity={op} />
                </>
              ) : (
                /* Default: single azurite tick at upper-right */
                <TickAt
                  top={12}
                  right={12}
                  size={6}
                  opacity={op * 0.9}
                />
              )}
              <span
                style={{
                  fontFamily: videoConfig.type.monoFamily,
                  fontSize: 14,
                  fontWeight: 600,
                  letterSpacing: videoConfig.type.tracking.eyebrow,
                  textTransform: "uppercase",
                  color: isActive
                    ? videoConfig.palette.accent
                    : videoConfig.palette.ink,
                }}
              >
                {p}
              </span>
            </div>
          );
        })}
      </div>

      {/* Footer attribution */}
      <div
        style={{
          position: "absolute",
          left: PAGE_INSET,
          right: PAGE_INSET,
          bottom: 96,
          opacity: footerOp,
        }}
      >
        <div
          aria-hidden
          style={{
            height: 1,
            background: videoConfig.palette.line,
            marginBottom: 24,
            transform: `scaleX(${footer})`,
            transformOrigin: "left",
          }}
        />
        <div
          style={{
            display: "flex",
            justifyContent: "space-between",
            color: videoConfig.palette.inkMutedSoft,
            fontFamily: videoConfig.type.monoFamily,
            fontSize: 11,
            fontWeight: 600,
            letterSpacing: videoConfig.type.tracking.eyebrow,
            textTransform: "uppercase",
          }}
        >
          <span>BYOK · Local · OAuth</span>
          <span>0% commission</span>
        </div>
      </div>

      <FilmGrain />
    </AbsoluteFill>
  );
};

/** A 6×6 azurite tick at a corner of a cell */
function TickAt({
  top,
  left,
  right,
  bottom,
  size,
  opacity,
}: {
  top?: number;
  left?: number;
  right?: number;
  bottom?: number;
  size: number;
  opacity: number;
}) {
  return (
    <span
      aria-hidden
      style={{
        position: "absolute",
        top,
        left,
        right,
        bottom,
        width: size,
        height: size,
        background: videoConfig.palette.accent,
        opacity,
      }}
    />
  );
}

const eyebrowStyle: CSSProperties = {
  fontFamily: videoConfig.type.monoFamily,
  fontSize: 13,
  fontWeight: 600,
  lineHeight: 1,
  letterSpacing: videoConfig.type.tracking.eyebrow,
  textTransform: "uppercase",
  color: videoConfig.palette.inkMutedSoft,
};
