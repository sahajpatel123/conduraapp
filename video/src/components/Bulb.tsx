/*
  The bulb — Synaptic's signature motif, recreated from web's
  components/pieces/illumination.tsx. A hanging filament lamp that ignites:
  `glow` (0..1) blooms the halo + glass, `filament` (0..1) lights the wire.
  An optional flicker models the moment of ignition.
*/
import React from "react";

export const Bulb: React.FC<{
  glow: number;
  filament: number;
  sway?: number; // degrees
  fgDim: string;
  fgFaint: string;
  bg3: string;
  width?: number;
}> = ({ glow, filament, sway = 0, fgDim, fgFaint, bg3, width = 360 }) => {
  const h = (width / 240) * 400;
  return (
    <svg
      viewBox="0 0 240 400"
      width={width}
      height={h}
      style={{ transformOrigin: "120px 0px", transform: `rotate(${sway}deg)`, overflow: "visible" }}
    >
      <defs>
        <radialGradient id="bloom" cx="50%" cy="50%" r="50%">
          <stop offset="0%" stopColor="#ffe2ae" stopOpacity="0.95" />
          <stop offset="35%" stopColor="#ffc46b" stopOpacity="0.5" />
          <stop offset="100%" stopColor="#ffc46b" stopOpacity="0" />
        </radialGradient>
        <radialGradient id="glass" cx="50%" cy="42%" r="62%">
          <stop offset="0%" stopColor="#fff3d6" stopOpacity="0.95" />
          <stop offset="60%" stopColor="#ffc46b" stopOpacity="0.55" />
          <stop offset="100%" stopColor="#ffa12e" stopOpacity="0.18" />
        </radialGradient>
      </defs>

      {/* cord */}
      <line x1="120" y1="-200" x2="120" y2="118" stroke={fgDim} strokeWidth="2.5" />

      {/* halo */}
      <circle cx="120" cy="234" r="150" fill="url(#bloom)" opacity={glow} />

      {/* glass fill (lights up) */}
      <path
        d="M100,148 C100,176 82,186 75,212 A62,62 0 1,0 165,212 C158,186 140,176 140,148 Z"
        fill="url(#glass)"
        opacity={glow}
      />
      {/* glass outline */}
      <path
        d="M100,148 C100,176 82,186 75,212 A62,62 0 1,0 165,212 C158,186 140,176 140,148 Z"
        fill="none"
        stroke={fgDim}
        strokeWidth="2.5"
      />

      {/* filament — cold then hot */}
      <g fill="none" strokeWidth="2.5" strokeLinecap="round">
        <path
          d="M106,152 V198 M134,152 V198 M106,198 Q113,216 120,198 Q127,180 134,198"
          stroke={fgFaint}
        />
        <path
          d="M106,152 V198 M134,152 V198 M106,198 Q113,216 120,198 Q127,180 134,198"
          stroke="#ffd9a0"
          opacity={filament}
        />
      </g>

      {/* cap */}
      <rect x="98" y="118" width="44" height="32" rx="6" fill={bg3} stroke={fgDim} strokeWidth="2.5" />
      <line x1="100" y1="128" x2="140" y2="128" stroke={fgFaint} strokeWidth="1.5" />
      <line x1="100" y1="138" x2="140" y2="138" stroke={fgFaint} strokeWidth="1.5" />
    </svg>
  );
};

// The reaching hand silhouette, from illumination.tsx's <Hand/>.
export const Hand: React.FC<{ fg: string; fgDim: string; bg2: string }> = ({
  fg,
  fgDim,
  bg2,
}) => (
  <svg viewBox="0 0 560 360" width={620} height={399} fill="none" style={{ overflow: "visible" }}>
    <path
      d="M16,138 C7,142 7,158 16,162 L150,170
         C176,172 188,186 186,200 C184,216 166,220 154,210
         C178,222 184,242 170,252 C158,260 144,254 138,244
         C158,260 158,280 142,288 C130,294 116,288 110,278
         C124,300 170,312 230,312 L560,316 L560,84 L360,82
         C296,82 226,96 182,126 C174,116 158,114 150,128 L16,138 Z"
      fill={bg2}
      stroke={fg}
      strokeWidth="3"
      strokeLinejoin="round"
    />
    <path
      d="M192,152 C232,142 272,152 290,174 C300,188 294,206 276,210 C248,216 212,200 198,180"
      stroke={fg}
      strokeWidth="3"
      strokeLinecap="round"
    />
    <path
      d="M150,140 C156,148 156,160 150,168"
      stroke={fgDim}
      strokeWidth="2"
      strokeLinecap="round"
    />
  </svg>
);
