/*
  24.4–30.6s — How a request becomes a safe action. The pipeline lights up
  stage by stage: classify the blast radius → pick the lightest capture →
  verify the screen hasn't changed → pass the deterministic Gatekeeper →
  route to the right tool.
*/
import React from "react";
import { AbsoluteFill, useCurrentFrame, interpolate, spring, useVideoConfig } from "remotion";
import { DARK, FONT, EASE } from "../theme";
import { StaffLines } from "../components/Background";
import { LightSweep } from "../components/Chrome";
import { Caption, Annotation } from "../components/Primitives";

const STAGES = [
  { n: "1", title: "Blast-Radius", sub: "READ · WRITE · NETWORK · DESTRUCTIVE", icon: "radius" },
  { n: "2", title: "Capture", sub: "AX-only → window → diff → full → CUA", icon: "eye" },
  { n: "3", title: "Twin-Snapshot", sub: "pre / post AX-tree diff", icon: "diff" },
  { n: "4", title: "Gatekeeper", sub: "deterministic rules engine", icon: "shield" },
  { n: "5", title: "Delegate", sub: "route to the best tool", icon: "bus" },
] as const;

const TOOLS = ["Claude Code", "Codex", "Ollama", "Gemini", "OpenCode"];

export const Perception: React.FC = () => {
  const frame = useCurrentFrame();
  const active = Math.floor(interpolate(frame, [16, 150], [0, STAGES.length], { extrapolateLeft: "clamp", extrapolateRight: "clamp" }));

  return (
    <AbsoluteFill style={{ background: DARK.bg }}>
      <StaffLines theme={DARK} opacity={0.5} />
      <LightSweep start={0} />

      <AbsoluteFill style={{ alignItems: "center", justifyContent: "center", flexDirection: "column" }}>
        <div style={{ marginBottom: 30 }}>
          <Annotation color={DARK.fgFaint}>selective perception → safe action</Annotation>
        </div>

        <div style={{ display: "flex", alignItems: "stretch", gap: 0 }}>
          {STAGES.map((s, i) => (
            <React.Fragment key={s.n}>
              <Stage stage={s} index={i} active={i <= active} current={i === active} />
              {i < STAGES.length - 1 && <Wire on={i < active} index={i} />}
            </React.Fragment>
          ))}
        </div>

        {/* delegation fan-out under stage 5 */}
        <div style={{ display: "flex", gap: 12, marginTop: 40, opacity: interpolate(frame, [120, 150], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: EASE }) }}>
          {TOOLS.map((t, i) => (
            <ToolChip key={t} label={t} delay={120 + i * 5} />
          ))}
        </div>
      </AbsoluteFill>

      <Caption start={6} end={180}>
        It perceives only what's needed, verifies, checks the Gatekeeper, then routes the task.
      </Caption>
    </AbsoluteFill>
  );
};

const Stage: React.FC<{ stage: (typeof STAGES)[number]; index: number; active: boolean; current: boolean }> = ({ stage, active, current }) => {
  const frame = useCurrentFrame();
  const pulse = current ? interpolate(Math.sin(frame / 5), [-1, 1], [0.4, 1]) : active ? 0.6 : 0;
  const color = active ? DARK.accent : DARK.fgFaint;
  return (
    <div
      style={{
        width: 196,
        padding: "22px 18px",
        background: DARK.bg2,
        border: `1px solid ${active ? DARK.accent : DARK.line}`,
        boxShadow: active ? `0 0 ${20 + pulse * 26}px -10px ${DARK.glow}` : "none",
        display: "flex",
        flexDirection: "column",
        alignItems: "center",
        gap: 12,
        textAlign: "center",
        transition: "none",
      }}
    >
      <div
        style={{
          width: 30,
          height: 30,
          borderRadius: 9999,
          border: `1px solid ${color}`,
          display: "grid",
          placeItems: "center",
          fontFamily: FONT.mono,
          fontSize: 14,
          color,
        }}
      >
        {stage.n}
      </div>
      <StageIcon kind={stage.icon} color={color} />
      <div style={{ fontFamily: FONT.display, fontWeight: 700, fontSize: 20, color: active ? DARK.fg : DARK.fgDim }}>{stage.title}</div>
      <div style={{ fontFamily: FONT.mono, fontSize: 11, letterSpacing: "0.04em", color: DARK.fgFaint, lineHeight: 1.5 }}>{stage.sub}</div>
    </div>
  );
};

const Wire: React.FC<{ on: boolean; index: number }> = ({ on, index }) => {
  const frame = useCurrentFrame();
  const dash = -(frame * 1.4) % 24;
  return (
    <div style={{ width: 46, alignSelf: "center", position: "relative", height: 4 }}>
      <svg width={46} height={4} style={{ overflow: "visible" }}>
        <line x1={0} y1={2} x2={46} y2={2} stroke={DARK.line} strokeWidth={2} />
        {on && (
          <line x1={0} y1={2} x2={46} y2={2} stroke={DARK.glow} strokeWidth={2.4} strokeDasharray="6 6" strokeDashoffset={dash} opacity={0.9} />
        )}
      </svg>
    </div>
  );
};

const ToolChip: React.FC<{ label: string; delay: number }> = ({ label, delay }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const s = spring({ frame: frame - delay, fps, config: { damping: 14 } });
  return (
    <div
      style={{
        padding: "10px 18px",
        borderRadius: 9999,
        border: `1px solid ${DARK.lineStrong}`,
        background: DARK.bg3,
        fontFamily: FONT.mono,
        fontSize: 14,
        color: DARK.fg,
        opacity: s,
        transform: `translateY(${(1 - s) * 12}px)`,
      }}
    >
      {label}
    </div>
  );
};

const StageIcon: React.FC<{ kind: string; color: string }> = ({ kind, color }) => {
  const p = { width: 30, height: 30, fill: "none", stroke: color, strokeWidth: 1.6, strokeLinecap: "round" as const, strokeLinejoin: "round" as const };
  switch (kind) {
    case "radius":
      return (
        <svg viewBox="0 0 24 24" {...p}>
          <circle cx="12" cy="12" r="3" />
          <circle cx="12" cy="12" r="8" opacity={0.5} />
        </svg>
      );
    case "eye":
      return (
        <svg viewBox="0 0 24 24" {...p}>
          <path d="M2 12s4-7 10-7 10 7 10 7-4 7-10 7-10-7-10-7z" />
          <circle cx="12" cy="12" r="3" />
        </svg>
      );
    case "diff":
      return (
        <svg viewBox="0 0 24 24" {...p}>
          <rect x="3" y="4" width="7" height="16" rx="1" />
          <rect x="14" y="4" width="7" height="16" rx="1" />
        </svg>
      );
    case "shield":
      return (
        <svg viewBox="0 0 24 24" {...p}>
          <path d="M12 3l8 3v6c0 5-3.5 8-8 9-4.5-1-8-4-8-9V6z" />
          <path d="M9 12l2 2 4-4" />
        </svg>
      );
    default:
      return (
        <svg viewBox="0 0 24 24" {...p}>
          <path d="M4 12h16M12 4v16M7 7l10 10M17 7L7 17" opacity={0.8} />
        </svg>
      );
  }
};
