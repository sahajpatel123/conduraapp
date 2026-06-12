/*
  44.4–50.8s — The wider orchestra. Peer-to-peer encrypted sync, action replay,
  a public skills hub, and an adaptive engine that learns your style — all
  local, all free.
*/
import React from "react";
import { AbsoluteFill, useCurrentFrame, interpolate, spring, useVideoConfig } from "remotion";
import { DARK, FONT, EASE } from "../theme";
import { StaffLines } from "../components/Background";
import { LightSweep } from "../components/Chrome";
import { Caption, Annotation } from "../components/Primitives";

export const Features: React.FC = () => {
  return (
    <AbsoluteFill style={{ background: DARK.bg }}>
      <StaffLines theme={DARK} opacity={0.5} />
      <LightSweep start={0} />

      <AbsoluteFill style={{ alignItems: "center", justifyContent: "center" }}>
        <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: 26, width: 1180 }}>
          <Card title="P2P Encrypted Sync" sub="device-to-device · E2E · no central server" at={6}>
            <P2PVisual />
          </Card>
          <Card title="Action Replay" sub="scrubbable 24h timeline · screenshots + decisions" at={22}>
            <ReplayVisual />
          </Card>
          <Card title="Skills Hub" sub="public · curated · safety-scanned" at={38}>
            <SkillsVisual />
          </Card>
          <Card title="User-Adaptive Engine" sub="observer → dialectic → predictor" at={54}>
            <AdaptiveVisual />
          </Card>
        </div>
      </AbsoluteFill>

      <Caption start={6} end={186}>
        Peer-to-peer sync, action replay, a public skills hub, an engine that learns you.
      </Caption>
    </AbsoluteFill>
  );
};

const Card: React.FC<{ title: string; sub: string; at: number; children: React.ReactNode }> = ({ title, sub, at, children }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const s = spring({ frame: frame - at, fps, config: { damping: 16, mass: 0.8 } });
  // pointer-tracked shine, faked with a slow sweep
  const shine = interpolate((frame + at * 4) % 120, [0, 120], [-20, 120]);
  return (
    <div
      style={{
        position: "relative",
        overflow: "hidden",
        display: "flex",
        alignItems: "center",
        gap: 22,
        padding: "26px 30px",
        background: DARK.bg2,
        border: `1px solid ${DARK.line}`,
        opacity: s,
        transform: `translateY(${(1 - s) * 30}px)`,
      }}
    >
      <div style={{ position: "absolute", inset: 0, pointerEvents: "none", background: `radial-gradient(420px circle at ${shine}% 30%, ${DARK.glow}, transparent 55%)`, opacity: 0.08 }} />
      <div style={{ width: 150, flexShrink: 0 }}>{children}</div>
      <div>
        <div style={{ fontFamily: FONT.display, fontWeight: 700, fontSize: 26, color: DARK.fg }}>{title}</div>
        <div style={{ marginTop: 6 }}>
          <Annotation>{sub}</Annotation>
        </div>
      </div>
    </div>
  );
};

const P2PVisual: React.FC = () => {
  const frame = useCurrentFrame();
  const t = (frame % 60) / 60;
  const x = interpolate(t, [0, 1], [22, 118]);
  return (
    <svg viewBox="0 0 150 90" width={150} height={90}>
      <Laptop x={6} />
      <Laptop x={108} />
      <line x1={40} y1={45} x2={120} y2={45} stroke={DARK.line} strokeWidth={2} />
      <g transform={`translate(${x}, 38)`}>
        <rect x={-6} y={-5} width={12} height={11} rx={2} fill="none" stroke={DARK.accent} strokeWidth={1.6} />
        <path d="M-3 -5 V-8 a3 3 0 0 1 6 0 V-5" fill="none" stroke={DARK.accent} strokeWidth={1.6} />
      </g>
    </svg>
  );
};

const Laptop: React.FC<{ x: number }> = ({ x }) => (
  <g transform={`translate(${x}, 30)`} stroke={DARK.fgDim} strokeWidth={1.6} fill="none">
    <rect x={0} y={0} width={36} height={24} rx={2} />
    <path d="M-4 28 H40" />
  </g>
);

const ReplayVisual: React.FC = () => {
  const frame = useCurrentFrame();
  const play = interpolate(frame % 70, [0, 70], [0, 100], { extrapolateRight: "clamp" });
  return (
    <div>
      <div style={{ display: "flex", gap: 5, marginBottom: 10 }}>
        {new Array(5).fill(0).map((_, i) => (
          <div key={i} style={{ width: 26, height: 20, borderRadius: 3, border: `1px solid ${DARK.line}`, background: DARK.bg3 }} />
        ))}
      </div>
      <div style={{ position: "relative", height: 6, borderRadius: 4, background: DARK.bg3 }}>
        <div style={{ position: "absolute", left: 0, top: 0, bottom: 0, width: `${play}%`, background: DARK.accent, borderRadius: 4, opacity: 0.7 }} />
        <div style={{ position: "absolute", left: `${play}%`, top: -4, width: 3, height: 14, background: DARK.accent, boxShadow: `0 0 8px ${DARK.glow}` }} />
      </div>
    </div>
  );
};

const SkillsVisual: React.FC = () => {
  const frame = useCurrentFrame();
  const dl = interpolate(frame % 80, [40, 60], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp" });
  return (
    <div style={{ border: `1px solid ${DARK.lineStrong}`, borderRadius: 8, padding: 12, background: DARK.bg3 }}>
      <div style={{ fontFamily: FONT.mono, fontSize: 11, color: DARK.fg, marginBottom: 8 }}>Summarize-and-Email</div>
      <div style={{ display: "flex", alignItems: "center", gap: 8 }}>
        <div style={{ flex: 1, height: 5, borderRadius: 4, background: DARK.bg }}>
          <div style={{ width: `${dl * 100}%`, height: "100%", background: DARK.accent, borderRadius: 4 }} />
        </div>
        <svg viewBox="0 0 24 24" width={16} height={16} fill="none" stroke={DARK.accent} strokeWidth={2} strokeLinecap="round">
          <path d="M12 3v12M7 11l5 5 5-5M5 21h14" />
        </svg>
      </div>
    </div>
  );
};

const AdaptiveVisual: React.FC = () => {
  const frame = useCurrentFrame();
  const bars = [0.4, 0.7, 0.5, 0.85, 0.65];
  return (
    <svg viewBox="0 0 150 90" width={150} height={90}>
      {bars.map((b, i) => {
        const g = interpolate(frame - i * 6, [0, 24], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: EASE });
        const h = b * 70 * g;
        return <rect key={i} x={10 + i * 28} y={80 - h} width={18} height={h} fill={DARK.accent} opacity={0.5 + b * 0.5} rx={2} />;
      })}
      <line x1={4} y1={80} x2={146} y2={80} stroke={DARK.line} strokeWidth={1.5} />
    </svg>
  );
};
