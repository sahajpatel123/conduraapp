/*
  The room. Faint ruled "staff" lines, drifting dust motes, optional light
  shafts, a vignette, and a generated film grain — all recreated from the
  site's globals.css so the film breathes like the website.
*/
import React from "react";
import { AbsoluteFill, useCurrentFrame, interpolate, random } from "remotion";

type Theme = {
  bg: string;
  bg2: string;
  line: string;
  glow: string;
};

export const StaffLines: React.FC<{ theme: Theme; opacity?: number }> = ({
  theme,
  opacity = 1,
}) => (
  <AbsoluteFill
    style={{
      opacity,
      backgroundImage: `repeating-linear-gradient(to bottom, transparent 0, transparent 26px, ${theme.line} 26px, ${theme.line} 28px)`,
      backgroundSize: "100% 140px",
      maskImage:
        "linear-gradient(to bottom, transparent, black 16%, black 84%, transparent)",
      WebkitMaskImage:
        "linear-gradient(to bottom, transparent, black 16%, black 84%, transparent)",
    }}
  />
);

export const Motes: React.FC<{ count?: number; glow: string }> = ({
  count = 26,
  glow,
}) => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{ overflow: "hidden" }}>
      {new Array(count).fill(0).map((_, i) => {
        const size = 2 + random(`s${i}`) * 4;
        const baseX = 4 + random(`x${i}`) * 92;
        const baseY = 8 + random(`y${i}`) * 80;
        const dur = 280 + random(`d${i}`) * 360;
        const phase = random(`p${i}`) * dur;
        const t = ((frame + phase) % dur) / dur;
        const driftX = (random(`dx${i}`) - 0.5) * 140 * t;
        const driftY = -(40 + random(`dy${i}`) * 120) * t;
        const op = interpolate(t, [0, 0.35, 1], [0.1, 0.5, 0.12]);
        return (
          <div
            key={i}
            style={{
              position: "absolute",
              left: `${baseX}%`,
              top: `${baseY}%`,
              width: size,
              height: size,
              borderRadius: 9999,
              background: glow,
              opacity: op,
              transform: `translate3d(${driftX}px, ${driftY}px, 0)`,
              filter: "blur(0.3px)",
            }}
          />
        );
      })}
    </AbsoluteFill>
  );
};

export const Shafts: React.FC<{ glow: string; opacity?: number }> = ({
  glow,
  opacity = 1,
}) => {
  const frame = useCurrentFrame();
  const breathe = interpolate(
    Math.sin(frame / 60),
    [-1, 1],
    [0.04, 0.11],
  );
  return (
    <AbsoluteFill style={{ opacity }}>
      {[18, 55].map((left, i) => (
        <div
          key={i}
          style={{
            position: "absolute",
            top: "-12%",
            left: `${left}%`,
            height: "124%",
            width: i === 0 ? 160 : 280,
            background: `linear-gradient(105deg, transparent 20%, ${glow} 50%, transparent 80%)`,
            opacity: breathe,
            filter: "blur(26px)",
            transform: "skewX(-18deg)",
          }}
        />
      ))}
    </AbsoluteFill>
  );
};

export const Vignette: React.FC<{ strength?: number }> = ({
  strength = 0.55,
}) => (
  <AbsoluteFill
    style={{
      background: `radial-gradient(120% 100% at 50% 45%, transparent 40%, rgba(0,0,0,${strength}) 100%)`,
      pointerEvents: "none",
    }}
  />
);

// Generated film grain — same data-URI SVG technique as the site, animated.
export const Grain: React.FC<{ opacity?: number }> = ({ opacity = 0.05 }) => {
  const frame = useCurrentFrame();
  const steps = 6;
  const idx = Math.floor((frame / 3) % steps);
  const shifts = [
    [0, 0],
    [-4, 2],
    [3, -3],
    [-2, 4],
    [4, 1],
    [-3, -2],
  ];
  const [tx, ty] = shifts[idx];
  return (
    <AbsoluteFill
      style={{
        inset: "-15%",
        width: "130%",
        height: "130%",
        pointerEvents: "none",
        opacity,
        transform: `translate(${tx}%, ${ty}%)`,
        backgroundImage:
          "url(\"data:image/svg+xml,%3Csvg xmlns='http://www.w3.org/2000/svg' width='256' height='256'%3E%3Cfilter id='n'%3E%3CfeTurbulence type='fractalNoise' baseFrequency='0.8' numOctaves='3' stitchTiles='stitch'/%3E%3C/filter%3E%3Crect width='100%25' height='100%25' filter='url(%23n)'/%3E%3C/svg%3E\")",
      }}
    />
  );
};
