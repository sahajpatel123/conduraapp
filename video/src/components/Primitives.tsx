/*
  Small shared building blocks: kinetic captions (the on-screen narration),
  annotation labels, metric badges, count-up numerals, and the one button
  shape from the site. Everything speaks the site's single ease dialect.
*/
import React from "react";
import { interpolate, useCurrentFrame, spring, useVideoConfig } from "remotion";
import { EASE, FONT, DARK } from "../theme";

// A line of display type that mask-reveals up from below.
export const LineReveal: React.FC<{
  children: React.ReactNode;
  delay?: number;
  duration?: number;
  style?: React.CSSProperties;
}> = ({ children, delay = 0, duration = 26, style }) => {
  const frame = useCurrentFrame();
  const t = interpolate(frame - delay, [0, duration], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
    easing: EASE,
  });
  return (
    <span style={{ display: "inline-block", overflow: "hidden", paddingBottom: "0.12em" }}>
      <span
        style={{
          display: "inline-block",
          transform: `translateY(${(1 - t) * 110}%)`,
          ...style,
        }}
      >
        {children}
      </span>
    </span>
  );
};

// Fade + rise for blocks.
export const Rise: React.FC<{
  children: React.ReactNode;
  delay?: number;
  y?: number;
  duration?: number;
  style?: React.CSSProperties;
}> = ({ children, delay = 0, y = 30, duration = 30, style }) => {
  const frame = useCurrentFrame();
  const t = interpolate(frame - delay, [0, duration], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
    easing: EASE,
  });
  return (
    <div style={{ opacity: t, transform: `translateY(${(1 - t) * y}px)`, ...style }}>
      {children}
    </div>
  );
};

// Subtitle-style narration line, pinned low. Fades in and out on its window.
export const Caption: React.FC<{
  children: React.ReactNode;
  start: number;
  end: number;
  fade?: number;
  color?: string;
  accent?: boolean;
  theme?: typeof DARK;
}> = ({ children, start, end, fade = 12, color, accent, theme = DARK }) => {
  const frame = useCurrentFrame();
  const opacity = interpolate(
    frame,
    [start, start + fade, end - fade, end],
    [0, 1, 1, 0],
    { extrapolateLeft: "clamp", extrapolateRight: "clamp" },
  );
  const y = interpolate(frame, [start, start + fade], [10, 0], {
    extrapolateRight: "clamp",
    easing: EASE,
  });
  return (
    <div
      style={{
        position: "absolute",
        bottom: 92,
        left: 0,
        right: 0,
        textAlign: "center",
        opacity,
        transform: `translateY(${y}px)`,
        pointerEvents: "none",
      }}
    >
      <span
        style={{
          fontFamily: FONT.sans,
          fontWeight: 500,
          fontSize: 32,
          letterSpacing: "0.01em",
          color: color ?? (accent ? theme.accent : theme.fg),
          textShadow: "0 2px 30px rgba(0,0,0,0.6)",
        }}
      >
        {children}
      </span>
    </div>
  );
};

export const Annotation: React.FC<{
  children: React.ReactNode;
  color?: string;
  style?: React.CSSProperties;
}> = ({ children, color = DARK.fgDim, style }) => (
  <span
    style={{
      fontFamily: FONT.mono,
      fontSize: 14,
      letterSpacing: "0.2em",
      textTransform: "uppercase",
      color,
      ...style,
    }}
  >
    {children}
  </span>
);

// Latency / metric badge, glass-on-ink with a brass dot.
export const Badge: React.FC<{
  label: string;
  value: string;
  theme?: typeof DARK;
  delay?: number;
  tone?: "brass" | "success" | "halt";
}> = ({ label, value, theme = DARK, delay = 0, tone = "brass" }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const s = spring({ frame: frame - delay, fps, config: { damping: 16, mass: 0.7 } });
  const dot = tone === "halt" ? theme.halt : tone === "success" ? theme.success : theme.accent;
  return (
    <div
      style={{
        display: "inline-flex",
        alignItems: "center",
        gap: 12,
        padding: "12px 18px",
        border: `1px solid ${theme.lineStrong}`,
        background: theme.bg2,
        opacity: s,
        transform: `translateY(${(1 - s) * 14}px)`,
        boxShadow: `0 18px 50px -28px ${dot}`,
      }}
    >
      <span style={{ width: 8, height: 8, borderRadius: 9999, background: dot, boxShadow: `0 0 12px 2px ${dot}` }} />
      <span style={{ fontFamily: FONT.mono, fontSize: 15, letterSpacing: "0.08em", color: theme.fgDim }}>
        {label}
      </span>
      <span style={{ fontFamily: FONT.mono, fontSize: 16, fontWeight: 500, color: theme.fg }}>{value}</span>
    </div>
  );
};

// Count-up numeral, like the site's <Counter/>.
export const CountUp: React.FC<{
  to: number;
  unit: string;
  prefix?: string;
  decimals?: number;
  start?: number;
  duration?: number;
  theme?: typeof DARK;
  fontSize?: number;
}> = ({ to, unit, prefix = "<", decimals = 0, start = 0, duration = 42, theme = DARK, fontSize = 72 }) => {
  const frame = useCurrentFrame();
  const t = interpolate(frame - start, [0, duration], [0, 1], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
    easing: EASE,
  });
  const value = (to * t).toFixed(decimals);
  return (
    <span style={{ fontFamily: FONT.display, fontWeight: 800, fontSize, color: theme.fg, fontVariantNumeric: "tabular-nums" }}>
      {prefix}
      {value}
      <span style={{ fontSize: fontSize * 0.5, color: theme.fgDim, marginLeft: 6 }}>{unit}</span>
    </span>
  );
};

// The single CTA button shape, with filament glow.
export const Cta: React.FC<{
  children: React.ReactNode;
  theme?: typeof DARK;
  primary?: boolean;
  glow?: number;
}> = ({ children, theme = DARK, primary, glow = 0 }) => (
  <div
    style={{
      display: "inline-flex",
      alignItems: "center",
      gap: 10,
      border: `1px solid ${primary ? theme.accent : theme.lineStrong}`,
      background: theme.bg2,
      padding: "16px 30px",
      fontFamily: FONT.mono,
      fontSize: 18,
      letterSpacing: "0.025em",
      color: primary ? theme.accent : theme.fg,
      boxShadow: primary
        ? `0 0 ${24 + glow * 24}px ${-4 + glow * 4}px ${theme.glow}, inset 0 0 18px -12px ${theme.glow}`
        : "none",
    }}
  >
    {children}
  </div>
);
