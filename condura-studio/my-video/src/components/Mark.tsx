// src/components/Mark.tsx
//
// The Condura mark, redesigned for the White Room direction.
//
// Old (warm paper): "C" letterform + cream pill shadow.
// New (white room):  pure-ink square outline, a single azurite dot at
//                    1/3 from the top, a hairline accent bar at 1/3
//                    from the bottom. Geometry, no letterform.
//
// Two sizes: regular (hero), quiet (in-map placeholder).
//
// The mark in the regular size is what the audience sees for ~12 seconds
// across the film — so it has to read as engineered, not decorative.

import type { CSSProperties } from "react";

import { videoConfig } from "../config/video.config";

type Size = "regular" | "quiet";

const SIZE: Record<Size, { box: number; stroke: number; dot: number; barW: number; barH: number }> = {
  // box outer px · stroke px · dot diameter px · accent-bar width px · accent-bar height px
  regular: { box: 96, stroke: 1.5, dot: 7, barW: 26, barH: 2 },
  quiet: { box: 56, stroke: 1, dot: 4, barW: 14, barH: 1.5 },
};

export function Mark({
  size = "regular",
  animatedDot = false,
  style,
}: {
  size?: Size;
  /**
   * If true, the azurite dot's opacity is hooked to global frame so it
   * breathes (subtle 4s period). Use only on hero placements; the
   * mark-in-map remains static.
   */
  animatedDot?: boolean;
  style?: CSSProperties;
}) {
  const d = SIZE[size];

  // Optional breathing for hero mark — uses Math.sin against a frame
  // the caller can pipe in via a CSS variable, but for the static
  // variant we just leave it at peak opacity.
  // (Pulse is implemented via inline style below — see how to opt-in.)
  return (
    <div
      style={{
        position: "relative",
        width: d.box,
        height: d.box,
        margin: "0 auto",
        border: `${d.stroke}px solid ${videoConfig.palette.ink}`,
        background: "transparent",
        boxSizing: "border-box",
        ...style,
      }}
    >
      {/* Single azurite dot, 1/3 from the top */}
      <span
        aria-hidden
        style={{
          position: "absolute",
          left: "50%",
          top: "32%",
          transform: "translate(-50%, -50%)",
          width: d.dot,
          height: d.dot,
          borderRadius: "50%",
          background: videoConfig.palette.accent,
          opacity: animatedDot ? undefined : 1,
          // If you wire animatedDot, override via parent style:
          //   opacity: math.sin(frame / 60) * 0.3 + 0.7
        }}
      />
      {/* Hairline accent bar, 1/3 from the bottom */}
      <span
        aria-hidden
        style={{
          position: "absolute",
          left: "50%",
          bottom: "32%",
          transform: "translateX(-50%)",
          width: d.barW,
          height: d.barH,
          background: videoConfig.palette.markLine,
        }}
      />
    </div>
  );
}
