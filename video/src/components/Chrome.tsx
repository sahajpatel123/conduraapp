/*
  Reusable window chrome for the app mockups (terminal, overlay, PDF viewer,
  email composer, calendar) plus the Synaptic overlay and the brass
  light-sweep transition that wipes between major sections (echoing the
  bulb-touch bloom).
*/
import React from "react";
import { AbsoluteFill, interpolate, useCurrentFrame, spring, useVideoConfig } from "remotion";
import { EASE, FONT, DARK } from "../theme";

// A macOS-ish frameless window with a title bar and traffic lights.
export const Window: React.FC<{
  title: string;
  children: React.ReactNode;
  width?: number;
  theme?: typeof DARK;
  accentLight?: boolean;
  style?: React.CSSProperties;
}> = ({ title, children, width = 720, theme = DARK, accentLight, style }) => (
  <div
    style={{
      width,
      border: `1px solid ${theme.line}`,
      background: theme.bg2,
      boxShadow: "0 40px 120px -30px rgba(0,0,0,0.7)",
      overflow: "hidden",
      ...style,
    }}
  >
    <div
      style={{
        display: "flex",
        alignItems: "center",
        justifyContent: "space-between",
        borderBottom: `1px solid ${theme.line}`,
        padding: "12px 18px",
      }}
    >
      <span style={{ fontFamily: FONT.mono, fontSize: 13, letterSpacing: "0.18em", textTransform: "uppercase", color: theme.fgFaint }}>
        {title}
      </span>
      <span style={{ display: "flex", gap: 7 }}>
        <Dot theme={theme} />
        <Dot theme={theme} />
        <Dot theme={theme} fill={accentLight ? theme.accent : undefined} />
      </span>
    </div>
    <div>{children}</div>
  </div>
);

const Dot: React.FC<{ theme: typeof DARK; fill?: string }> = ({ theme, fill }) => (
  <span
    style={{
      width: 11,
      height: 11,
      borderRadius: 9999,
      border: fill ? "none" : `1px solid ${theme.lineStrong}`,
      background: fill ?? "transparent",
    }}
  />
);

// The frameless, always-on-top Synaptic overlay: a transparent pill with a
// mic and the listening state.
export const Overlay: React.FC<{
  theme?: typeof DARK;
  state: string;
  width?: number;
  pulse?: number;
  children?: React.ReactNode;
}> = ({ theme = DARK, state, width = 640, pulse = 0, children }) => (
  <div
    style={{
      width,
      borderRadius: 22,
      border: `1px solid ${theme.lineStrong}`,
      background: "rgba(14,14,20,0.72)",
      backdropFilter: "blur(18px)",
      WebkitBackdropFilter: "blur(18px)",
      boxShadow: `0 30px 90px -20px rgba(0,0,0,0.8), 0 0 ${30 + pulse * 40}px -10px ${theme.glow}`,
      overflow: "hidden",
    }}
  >
    <div style={{ display: "flex", alignItems: "center", gap: 18, padding: "20px 26px" }}>
      <Mic theme={theme} pulse={pulse} />
      <div style={{ flex: 1 }}>{children ?? <span style={{ fontFamily: FONT.sans, fontSize: 22, color: theme.fgDim }}>{state}</span>}</div>
    </div>
  </div>
);

export const Mic: React.FC<{ theme: typeof DARK; pulse?: number }> = ({ theme, pulse = 0 }) => (
  <div
    style={{
      width: 46,
      height: 46,
      borderRadius: 9999,
      display: "grid",
      placeItems: "center",
      background: theme.bg3,
      border: `1px solid ${theme.lineStrong}`,
      boxShadow: `0 0 ${10 + pulse * 26}px ${pulse * 3}px ${theme.accent}`,
    }}
  >
    <svg viewBox="0 0 24 24" width={22} height={22} fill="none" stroke={theme.accent} strokeWidth={1.8} strokeLinecap="round">
      <rect x="9" y="3" width="6" height="11" rx="3" />
      <path d="M6 11a6 6 0 0 0 12 0" />
      <path d="M12 17v3" />
    </svg>
  </div>
);

// Live voice waveform (deterministic, frame-driven).
export const Waveform: React.FC<{ bars?: number; color: string; height?: number; energy?: number }> = ({
  bars = 38,
  color,
  height = 44,
  energy = 1,
}) => {
  const frame = useCurrentFrame();
  return (
    <div style={{ display: "flex", alignItems: "center", gap: 4, height }}>
      {new Array(bars).fill(0).map((_, i) => {
        const wob =
          Math.sin(frame / 4 + i * 0.7) * 0.5 +
          Math.sin(frame / 7 + i * 1.3) * 0.3 +
          Math.sin(frame / 11 + i) * 0.2;
        const env = Math.sin((i / bars) * Math.PI); // taper edges
        const h = Math.max(3, (0.2 + Math.abs(wob) * env) * height * energy);
        return (
          <div
            key={i}
            style={{ width: 4, height: h, borderRadius: 4, background: color, opacity: 0.55 + env * 0.45 }}
          />
        );
      })}
    </div>
  );
};

// The brass light-sweep that wipes the screen between sections.
export const LightSweep: React.FC<{ start: number; duration?: number; from?: string }> = ({
  start,
  duration = 26,
  from = "#ffc46b",
}) => {
  const frame = useCurrentFrame();
  const p = interpolate(frame - start, [0, duration], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
    easing: EASE,
  });
  if (p <= 0 || p >= 1) return null;
  const x = interpolate(p, [0, 1], [-40, 140]);
  const op = interpolate(p, [0, 0.3, 0.7, 1], [0, 0.9, 0.9, 0]);
  return (
    <AbsoluteFill style={{ pointerEvents: "none", overflow: "hidden" }}>
      <div
        style={{
          position: "absolute",
          top: 0,
          bottom: 0,
          left: `${x}%`,
          width: "55%",
          transform: "skewX(-14deg)",
          background: `linear-gradient(100deg, transparent, ${from} 50%, transparent)`,
          opacity: op,
          filter: "blur(8px)",
        }}
      />
    </AbsoluteFill>
  );
};

// A spring-scaled green check that stamps on completion.
export const Check: React.FC<{ at: number; theme?: typeof DARK; size?: number }> = ({
  at,
  theme = DARK,
  size = 28,
}) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const s = spring({ frame: frame - at, fps, config: { damping: 12, mass: 0.5 } });
  if (frame < at) return null;
  return (
    <div
      style={{
        width: size,
        height: size,
        borderRadius: 9999,
        background: theme.success,
        display: "grid",
        placeItems: "center",
        transform: `scale(${s})`,
      }}
    >
      <svg viewBox="0 0 24 24" width={size * 0.62} height={size * 0.62} fill="none" stroke="#08080c" strokeWidth={3} strokeLinecap="round" strokeLinejoin="round">
        <path d="M5 13l4 4L19 7" />
      </svg>
    </div>
  );
};
