/*
  18.4–24.4s — Speak naturally. The overlay listens; a live waveform tracks the
  voice. Local Whisper STT — no cloud cost, zero latency, privacy-first.
*/
import React from "react";
import { AbsoluteFill, useCurrentFrame, interpolate, spring, useVideoConfig } from "remotion";
import { DARK, FONT, EASE } from "../theme";
import { StaffLines, Motes } from "../components/Background";
import { Overlay, Waveform, LightSweep } from "../components/Chrome";
import { Caption } from "../components/Primitives";

const SPOKEN = "Hey Synaptic — summarize this PDF and draft an email to the team.";

export const Voice: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const enter = spring({ frame, fps, config: { damping: 18, mass: 0.8 } });

  const chars = Math.floor(interpolate(frame, [24, 150], [0, SPOKEN.length], { extrapolateLeft: "clamp", extrapolateRight: "clamp" }));
  const energy = interpolate(Math.sin(frame / 5), [-1, 1], [0.7, 1.1]);

  return (
    <AbsoluteFill style={{ background: DARK.bg }}>
      <StaffLines theme={DARK} opacity={0.5} />
      <Motes glow={DARK.glow} count={16} />
      <LightSweep start={0} />

      <AbsoluteFill style={{ alignItems: "center", justifyContent: "center" }}>
        <div style={{ width: 1040, opacity: enter, transform: `translateY(${(1 - enter) * 30}px)` }}>
          <Overlay theme={DARK} state="" width={1040} pulse={0.7}>
            <div style={{ display: "flex", alignItems: "center", gap: 26 }}>
              <Waveform color={DARK.accent} bars={40} height={56} energy={energy} />
              <div style={{ flex: 1, fontFamily: FONT.sans, fontSize: 26, color: DARK.fg, minHeight: 36 }}>
                {SPOKEN.slice(0, chars)}
                <span style={{ color: DARK.accent }}>{chars < SPOKEN.length ? "▍" : ""}</span>
              </div>
            </div>
          </Overlay>

          {/* privacy / local tag row */}
          <div style={{ display: "flex", gap: 16, marginTop: 26, justifyContent: "center", opacity: interpolate(frame, [40, 70], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: EASE }) }}>
            <Pill icon="mic">whisper.cpp · on-device STT</Pill>
            <Pill icon="lock">privacy-first · zero runtime cost</Pill>
            <Pill icon="bolt">push-to-talk</Pill>
          </div>
        </div>
      </AbsoluteFill>

      <Caption start={6} end={174}>
        Speak naturally. Local Whisper STT — no cloud cost, zero latency.
      </Caption>
    </AbsoluteFill>
  );
};

const Pill: React.FC<{ children: React.ReactNode; icon: "mic" | "lock" | "bolt" }> = ({ children, icon }) => (
  <div
    style={{
      display: "inline-flex",
      alignItems: "center",
      gap: 10,
      padding: "10px 18px",
      borderRadius: 9999,
      border: `1px solid ${DARK.line}`,
      background: DARK.bg2,
    }}
  >
    <Glyph kind={icon} />
    <span style={{ fontFamily: FONT.mono, fontSize: 14, letterSpacing: "0.04em", color: DARK.fgDim }}>{children}</span>
  </div>
);

const Glyph: React.FC<{ kind: "mic" | "lock" | "bolt" }> = ({ kind }) => {
  const c = DARK.accent;
  if (kind === "lock")
    return (
      <svg viewBox="0 0 24 24" width={16} height={16} fill="none" stroke={c} strokeWidth={1.8}>
        <rect x="5" y="11" width="14" height="9" rx="2" />
        <path d="M8 11V8a4 4 0 0 1 8 0v3" />
      </svg>
    );
  if (kind === "bolt")
    return (
      <svg viewBox="0 0 24 24" width={16} height={16} fill={c} stroke="none">
        <path d="M13 2L4 14h6l-1 8 9-12h-6z" />
      </svg>
    );
  return (
    <svg viewBox="0 0 24 24" width={16} height={16} fill="none" stroke={c} strokeWidth={1.8} strokeLinecap="round">
      <rect x="9" y="3" width="6" height="11" rx="3" />
      <path d="M6 11a6 6 0 0 0 12 0M12 17v3" />
    </svg>
  );
};
