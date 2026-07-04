// src/components/FilmGrain.tsx
//
// Subtle film grain. On white, grain reads ~2× louder than on cream,
// so intensity is dialed to 0.05 (the config holds the exact number).
//
// Same deterministic seed (1729) so re-renders are byte-identical.

import { AbsoluteFill } from "remotion";

import { videoConfig } from "../config/video.config";

export function FilmGrain() {
  const { intensity, seed, baseFrequency } = videoConfig.grain;
  const filterId = "condura-grain";

  return (
    <AbsoluteFill
      aria-hidden
      style={{
        pointerEvents: "none",
        mixBlendMode: "multiply",
        opacity: intensity,
      }}
    >
      <svg
        width="100%"
        height="100%"
        xmlns="http://www.w3.org/2000/svg"
        style={{ position: "absolute", inset: 0 }}
      >
        <filter id={filterId}>
          <feTurbulence
            type="fractalNoise"
            baseFrequency={baseFrequency}
            numOctaves="2"
            seed={seed}
            stitchTiles="stitch"
          />
          <feColorMatrix type="saturate" values="0" />
        </filter>
        <rect width="100%" height="100%" filter={`url(#${filterId})`} />
      </svg>
    </AbsoluteFill>
  );
}
