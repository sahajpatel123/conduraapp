/*
  38.4–44.4s — Proof and control. Every action lands in a tamper-resistant,
  HMAC-chained audit log. And you can always stop the agent: hard hotkey,
  watchdog, network isolation, menu-bar kill.
*/
import React from "react";
import { AbsoluteFill, useCurrentFrame, interpolate, spring, useVideoConfig } from "remotion";
import { DARK, FONT } from "../theme";
import { StaffLines } from "../components/Background";
import { Window, LightSweep } from "../components/Chrome";
import { Caption, Annotation } from "../components/Primitives";

const ENTRIES = [
  { kind: "READ", text: "capture window · report.pdf", hash: "9f2a…c1" },
  { kind: "WRITE", text: "compose mail · the team", hash: "4be7…0d" },
  { kind: "READ", text: "ax-tree diff · verified", hash: "1c80…a9" },
  { kind: "WRITE", text: "calendar · team sync", hash: "77d1…3f" },
  { kind: "NETWORK", text: "provider · anthropic", hash: "a042…7e" },
  { kind: "READ", text: "spend check · $0.00", hash: "55fb…22" },
  { kind: "WRITE", text: "overlay · dismissed", hash: "e9c4…b8" },
];

const KILLS = [
  { label: "hard hotkey", sub: "⎋ ⎋", icon: "key" },
  { label: "watchdog", sub: "auto-halt", icon: "timer" },
  { label: "network isolation", sub: "offline", icon: "shield" },
  { label: "menu-bar kill", sub: "one click", icon: "power" },
];

export const Audit: React.FC = () => {
  const frame = useCurrentFrame();
  const scroll = interpolate(frame, [20, 170], [0, ENTRIES.length * 44], { extrapolateLeft: "clamp", extrapolateRight: "clamp" });

  return (
    <AbsoluteFill style={{ background: DARK.bg }}>
      <StaffLines theme={DARK} opacity={0.5} />
      <LightSweep start={0} />

      {/* menu-bar strip */}
      <TrayBar />

      <AbsoluteFill style={{ alignItems: "center", justifyContent: "center", gap: 48, marginTop: 30 }}>
        {/* audit log */}
        <Window title="audit.log — HMAC-chained · append-only" theme={DARK} width={620} accentLight>
          <div style={{ height: 330, overflow: "hidden", padding: "10px 0", position: "relative" }}>
            <div style={{ transform: `translateY(${-scroll}px)` }}>
              {[...ENTRIES, ...ENTRIES].map((e, i) => (
                <LogRow key={i} entry={e} index={i} />
              ))}
            </div>
            <div style={{ position: "absolute", inset: 0, pointerEvents: "none", background: `linear-gradient(${DARK.bg2}, transparent 14%, transparent 86%, ${DARK.bg2})` }} />
          </div>
        </Window>

        {/* kill switch */}
        <div style={{ display: "flex", flexDirection: "column", gap: 16 }}>
          <Annotation color={DARK.halt}>you can always stop the agent</Annotation>
          {KILLS.map((k, i) => (
            <KillCard key={k.label} kill={k} at={40 + i * 24} />
          ))}
        </div>
      </AbsoluteFill>

      <Caption start={6} end={174}>
        Every action is recorded in a tamper-proof log — and you keep full control.
      </Caption>
    </AbsoluteFill>
  );
};

const LogRow: React.FC<{ entry: (typeof ENTRIES)[number]; index: number }> = ({ entry }) => {
  const color = entry.kind === "WRITE" ? DARK.accent : entry.kind === "NETWORK" ? "#7aa2ff" : DARK.fgDim;
  return (
    <div style={{ display: "flex", alignItems: "center", gap: 14, padding: "0 22px", height: 44, fontFamily: FONT.mono, fontSize: 14 }}>
      <svg viewBox="0 0 24 24" width={14} height={14} fill="none" stroke={DARK.fgFaint} strokeWidth={1.8}>
        <rect x="5" y="11" width="14" height="9" rx="2" />
        <path d="M8 11V8a4 4 0 0 1 8 0v3" />
      </svg>
      <span style={{ width: 84, color, fontWeight: 500 }}>{entry.kind}</span>
      <span style={{ flex: 1, color: DARK.fgDim }}>{entry.text}</span>
      <span style={{ color: DARK.fgFaint }}>↳ {entry.hash}</span>
    </div>
  );
};

const KillCard: React.FC<{ kill: (typeof KILLS)[number]; at: number }> = ({ kill, at }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const s = spring({ frame: frame - at, fps, config: { damping: 14 } });
  const live = frame > at + 6;
  return (
    <div
      style={{
        display: "flex",
        alignItems: "center",
        gap: 16,
        width: 320,
        padding: "14px 18px",
        background: DARK.bg2,
        border: `1px solid ${live ? DARK.halt : DARK.line}`,
        opacity: s,
        transform: `translateX(${(1 - s) * 24}px)`,
        boxShadow: live ? `0 0 24px -12px ${DARK.halt}` : "none",
      }}
    >
      <KillIcon kind={kill.icon} color={live ? DARK.halt : DARK.fgDim} />
      <div>
        <div style={{ fontFamily: FONT.sans, fontSize: 18, fontWeight: 600, color: DARK.fg }}>{kill.label}</div>
        <div style={{ fontFamily: FONT.mono, fontSize: 12, color: DARK.fgFaint }}>{kill.sub}</div>
      </div>
    </div>
  );
};

const TrayBar: React.FC = () => {
  const frame = useCurrentFrame();
  const pulse = interpolate(Math.sin(frame / 7), [-1, 1], [0.4, 1]);
  return (
    <div style={{ position: "absolute", top: 0, left: 0, right: 0, height: 40, background: "rgba(8,8,12,0.9)", borderBottom: `1px solid ${DARK.line}`, display: "flex", alignItems: "center", justifyContent: "flex-end", gap: 22, paddingRight: 30, zIndex: 40 }}>
      <span style={{ fontFamily: FONT.mono, fontSize: 12, color: DARK.fgDim }}>active · 1 conversation</span>
      <span style={{ fontFamily: FONT.mono, fontSize: 12, color: DARK.fgDim }}>spend $0.00 today</span>
      <span style={{ display: "inline-flex", alignItems: "center", gap: 7 }}>
        <span style={{ width: 9, height: 9, borderRadius: 9999, background: DARK.accent, boxShadow: `0 0 ${6 + pulse * 8}px ${DARK.glow}` }} />
        <span style={{ fontFamily: FONT.display, fontWeight: 700, fontSize: 14, color: DARK.fg }}>Synaptic</span>
      </span>
    </div>
  );
};

const KillIcon: React.FC<{ kind: string; color: string }> = ({ kind, color }) => {
  const p = { width: 24, height: 24, fill: "none", stroke: color, strokeWidth: 1.8, strokeLinecap: "round" as const, strokeLinejoin: "round" as const };
  switch (kind) {
    case "key":
      return (
        <svg viewBox="0 0 24 24" {...p}>
          <circle cx="8" cy="8" r="4" />
          <path d="M11 11l8 8M16 16l3-3" />
        </svg>
      );
    case "timer":
      return (
        <svg viewBox="0 0 24 24" {...p}>
          <circle cx="12" cy="13" r="8" />
          <path d="M12 13V9M9 2h6" />
        </svg>
      );
    case "shield":
      return (
        <svg viewBox="0 0 24 24" {...p}>
          <path d="M12 3l8 3v6c0 5-3.5 8-8 9-4.5-1-8-4-8-9V6z" />
          <path d="M8 12h8" />
        </svg>
      );
    default:
      return (
        <svg viewBox="0 0 24 24" {...p}>
          <path d="M12 3v9" />
          <path d="M6.5 7a8 8 0 1 0 11 0" />
        </svg>
      );
  }
};
