/*
  3.2–7.6s — The problem. A cluttered desktop: AI tools scattered everywhere,
  a cursor darting between them. Subscription lock-in, isolated tools, no
  universal hotkey. Each pain gets a red ✕.
*/
import React from "react";
import { AbsoluteFill, useCurrentFrame, interpolate, spring, useVideoConfig, random } from "remotion";
import { DARK, FONT, EASE } from "../theme";
import { StaffLines } from "../components/Background";
import { Caption } from "../components/Primitives";

const TOOLS = ["ChatGPT", "Claude", "Codex", "Ollama", "Gemini", "Antigravity", "OpenCode", "Kilo", "Hermes"];

const Tile: React.FC<{ label: string; x: number; y: number; i: number }> = ({ label, x, y, i }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const s = spring({ frame: frame - i * 3, fps, config: { damping: 14, mass: 0.6 } });
  const jit = Math.sin(frame / 8 + i) * 3;
  const initial = (label.match(/[A-Z]/g) || ["A"]).slice(0, 2).join("");
  return (
    <div
      style={{
        position: "absolute",
        left: `${x}%`,
        top: `${y}%`,
        transform: `translate(-50%, -50%) translateY(${jit}px) scale(${0.7 + s * 0.3})`,
        opacity: 0.4 + s * 0.45,
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        gap: 8,
        filter: "grayscale(0.5)",
      }}
    >
      <div
        style={{
          width: 76,
          height: 76,
          borderRadius: 18,
          background: DARK.bg3,
          border: `1px solid ${DARK.lineStrong}`,
          display: "grid",
          placeItems: "center",
          fontFamily: FONT.display,
          fontWeight: 700,
          fontSize: 26,
          color: DARK.fgDim,
        }}
      >
        {initial}
      </div>
      <span style={{ fontFamily: FONT.mono, fontSize: 12, color: DARK.fgFaint }}>{label}</span>
    </div>
  );
};

const PainLabel: React.FC<{ text: string; x: number; y: number; at: number }> = ({ text, x, y, at }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const s = spring({ frame: frame - at, fps, config: { damping: 13, mass: 0.7 } });
  if (frame < at) return null;
  return (
    <div
      style={{
        position: "absolute",
        left: `${x}%`,
        top: `${y}%`,
        transform: `translate(-50%, -50%) scale(${s})`,
        display: "flex",
        alignItems: "center",
        gap: 12,
        padding: "12px 20px",
        background: "rgba(20,12,12,0.85)",
        border: `1px solid ${DARK.halt}`,
        borderRadius: 10,
        boxShadow: `0 14px 40px -16px ${DARK.halt}`,
      }}
    >
      <svg viewBox="0 0 24 24" width={22} height={22} stroke={DARK.halt} strokeWidth={2.6} strokeLinecap="round">
        <path d="M6 6l12 12M18 6L6 18" />
      </svg>
      <span style={{ fontFamily: FONT.sans, fontSize: 22, fontWeight: 600, color: DARK.fg }}>{text}</span>
    </div>
  );
};

export const Problem: React.FC = () => {
  const frame = useCurrentFrame();
  const positions = TOOLS.map((t, i) => ({
    label: t,
    x: 14 + random(`px${i}`) * 72,
    y: 20 + random(`py${i}`) * 56,
    i,
  }));

  // A frantic cursor pinging between tiles.
  const target = positions[Math.floor((frame / 14) % positions.length)];
  const cx = interpolate(frame % 14, [0, 14], [target.x, positions[Math.floor((frame / 14 + 1) % positions.length)].x]);
  const cy = interpolate(frame % 14, [0, 14], [target.y, positions[Math.floor((frame / 14 + 1) % positions.length)].y]);

  const dim = interpolate(frame, [70, 110], [1, 0.35], { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: EASE });

  return (
    <AbsoluteFill style={{ background: DARK.bg }}>
      <StaffLines theme={DARK} opacity={0.5} />
      <AbsoluteFill style={{ opacity: dim }}>
        {positions.map((p) => (
          <Tile key={p.label} {...p} />
        ))}
        {/* frantic cursor */}
        <svg
          viewBox="0 0 24 24"
          width={26}
          height={26}
          style={{ position: "absolute", left: `${cx}%`, top: `${cy}%`, filter: `drop-shadow(0 2px 4px rgba(0,0,0,0.6))` }}
          fill={DARK.fg}
        >
          <path d="M3 2l7 18 2.5-7L20 11 3 2z" stroke={DARK.bg} strokeWidth={1} />
        </svg>
      </AbsoluteFill>

      <PainLabel text="Subscription lock-in" x={28} y={32} at={44} />
      <PainLabel text="Isolated tools" x={72} y={48} at={62} />
      <PainLabel text="No universal hotkey" x={42} y={70} at={80} />

      <Caption start={8} end={126}>
        Today, AI power is locked behind subscriptions, separate apps, endless switching.
      </Caption>
    </AbsoluteFill>
  );
};
