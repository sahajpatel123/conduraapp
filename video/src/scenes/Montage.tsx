/*
  30.6–38.4s — The work. Summarize a PDF, draft an email, schedule a meeting —
  one continuous flow, each step stamped with a green check and a timestamp.
*/
import React from "react";
import { AbsoluteFill, useCurrentFrame, interpolate, spring, useVideoConfig, Sequence } from "remotion";
import { DARK, FONT, EASE } from "../theme";
import { StaffLines } from "../components/Background";
import { Window, Check, LightSweep } from "../components/Chrome";
import { Caption, CountUp, Annotation } from "../components/Primitives";

// A line that "writes itself" — width grows as if being typed.
const FillLine: React.FC<{ at: number; w: number; accent?: boolean; tall?: boolean }> = ({ at, w, accent, tall }) => {
  const frame = useCurrentFrame();
  const grow = interpolate(frame - at, [0, 14], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: EASE });
  return (
    <div
      style={{
        height: tall ? 14 : 10,
        width: `${w * grow}%`,
        borderRadius: 4,
        background: accent ? DARK.accent : DARK.lineStrong,
        marginBottom: 12,
        opacity: accent ? 0.9 : 0.6,
      }}
    />
  );
};

const Panel: React.FC<{ title: string; at: number; checkAt: number; stamp: string; children: React.ReactNode }> = ({ title, at, checkAt, stamp, children }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const s = spring({ frame: frame - at, fps, config: { damping: 16, mass: 0.8 } });
  return (
    <div style={{ opacity: s, transform: `translateY(${(1 - s) * 40}px) scale(${0.92 + s * 0.08})`, position: "relative" }}>
      <Window title={title} theme={DARK} width={500} accentLight>
        <div style={{ padding: "22px 24px", minHeight: 210 }}>{children}</div>
      </Window>
      {/* completion stamp */}
      <div style={{ position: "absolute", right: 16, bottom: 16, display: "flex", alignItems: "center", gap: 10 }}>
        <span style={{ fontFamily: FONT.mono, fontSize: 12, color: DARK.fgFaint, opacity: frame > checkAt ? 1 : 0 }}>{stamp}</span>
        <Check at={checkAt} theme={DARK} size={26} />
      </div>
    </div>
  );
};

export const Montage: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <AbsoluteFill style={{ background: DARK.bg }}>
      <StaffLines theme={DARK} opacity={0.5} />
      <LightSweep start={0} />

      <AbsoluteFill style={{ alignItems: "center", justifyContent: "center", gap: 34, flexDirection: "row" }}>
        {/* PDF → summary */}
        <Panel title="report.pdf — viewer" at={6} checkAt={66} stamp="14:02:07">
          <Annotation color={DARK.accent}>summary</Annotation>
          <div style={{ height: 14 }} />
          <FillLine at={20} w={92} />
          <FillLine at={28} w={80} />
          <FillLine at={36} w={88} />
          <FillLine at={44} w={64} />
          <FillLine at={52} w={74} />
        </Panel>

        {/* Email composer */}
        <Panel title="Mail — new message" at={70} checkAt={140} stamp="14:02:11">
          <div style={{ display: "flex", gap: 10, marginBottom: 14, fontFamily: FONT.mono, fontSize: 13, color: DARK.fgDim }}>
            <span style={{ color: DARK.fgFaint }}>To:</span> the team
          </div>
          <FillLine at={86} w={70} accent tall />
          <div style={{ height: 8 }} />
          <FillLine at={98} w={94} />
          <FillLine at={106} w={86} />
          <FillLine at={114} w={90} />
          <FillLine at={122} w={58} />
        </Panel>

        {/* Calendar */}
        <Panel title="Calendar — June" at={134} checkAt={206} stamp="14:02:15">
          <div style={{ display: "grid", gridTemplateColumns: "repeat(5, 1fr)", gap: 6 }}>
            {new Array(20).fill(0).map((_, i) => {
              const isSlot = i === 12;
              const fill = interpolate(frame - 150, [0, 18], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp" });
              return (
                <div
                  key={i}
                  style={{
                    height: 34,
                    borderRadius: 6,
                    border: `1px solid ${DARK.line}`,
                    background: isSlot ? `rgba(255,161,46,${0.18 * fill})` : "transparent",
                    borderColor: isSlot ? DARK.accent : DARK.line,
                  }}
                />
              );
            })}
          </div>
          <div style={{ marginTop: 14, fontFamily: FONT.mono, fontSize: 13, color: DARK.accent, opacity: interpolate(frame - 168, [0, 14], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp" }) }}>
            ▸ Team sync · Thu 3:00 PM
          </div>
        </Panel>
      </AbsoluteFill>

      {/* counters */}
      <Sequence from={176}>
        <div style={{ position: "absolute", left: 0, right: 0, bottom: 176, display: "flex", justifyContent: "center", gap: 80 }}>
          <CounterStat label="actions completed" to={3} prefix="" unit="" />
          <CounterStat label="time saved" to={45} prefix="~" unit="s" />
        </div>
      </Sequence>

      <Caption start={6} end={228}>
        From documents to emails to meetings — Synaptic orchestrates every step.
      </Caption>
    </AbsoluteFill>
  );
};

const CounterStat: React.FC<{ label: string; to: number; prefix: string; unit: string }> = ({ label, to, prefix, unit }) => (
  <div style={{ textAlign: "center" }}>
    <CountUp to={to} unit={unit} prefix={prefix} start={0} duration={34} theme={DARK} fontSize={54} />
    <div style={{ marginTop: 4 }}>
      <Annotation>{label}</Annotation>
    </div>
  </div>
);
