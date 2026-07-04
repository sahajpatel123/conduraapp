// src/components/CornerBrackets.tsx
//
// Four L-shaped hairline brackets anchored to the four corners of a
// containing block. The classic "measurement frame" — engineering
// drawings have these; high-end UI mockups have them; Type foundry
// specimens have them. In the film they appear at scene boundaries
// to anchor the canvas against pure white.
//
// Frames use `box-sizing: content-box` so the inset math is simple:
//   stroke 1.5px sits *inside* the inset via border-box. We use
//   absolutely positioned sub-spans instead so a frame can scale
//   open from 0 length.

import type { CSSProperties } from "react";

import { videoConfig } from "../config/video.config";

const ARM = 48; // length of each bracket arm, px

export function CornerBrackets({
  inset = 60,
  color = videoConfig.palette.lineStrong,
  style,
}: {
  /** Distance from each canvas edge, px. */
  inset?: number;
  color?: string;
  style?: CSSProperties;
}) {
  const armStyle: CSSProperties = {
    position: "absolute",
    width: ARM,
    height: 1.5,
    background: color,
  };
  const armStyleV: CSSProperties = {
    position: "absolute",
    width: 1.5,
    height: ARM,
    background: color,
  };

  const wrap: CSSProperties = {
    position: "absolute",
    inset: 0,
    pointerEvents: "none",
    ...style,
  };

  return (
    <div style={wrap} aria-hidden>
      {/* TL */}
      <div style={{ ...armStyle, top: inset, left: inset }} />
      <div
        style={{ ...armStyleV, top: inset, left: inset, transform: `translateY(${ARM}px)` }}
      />
      {/* TR */}
      <div style={{ ...armStyle, top: inset, right: inset }} />
      <div
        style={{
          ...armStyleV,
          top: inset,
          right: inset,
          transform: `translateY(${ARM}px)`,
        }}
      />
      {/* BL */}
      <div style={{ ...armStyle, bottom: inset, left: inset }} />
      <div
        style={{
          ...armStyleV,
          bottom: inset,
          left: inset,
          transform: `translateY(-${ARM}px)`,
        }}
      />
      {/* BR */}
      <div style={{ ...armStyle, bottom: inset, right: inset }} />
      <div
        style={{
          ...armStyleV,
          bottom: inset,
          right: inset,
          transform: `translateY(-${ARM}px)`,
        }}
      />
    </div>
  );
}
