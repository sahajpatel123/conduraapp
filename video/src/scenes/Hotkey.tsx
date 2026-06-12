/*
  12.6–18.4s — Press your hotkey, the overlay appears instantly. Beside it, the
  daemon answers a ping with a pong. Speed is the product:
  cold start < 500 ms, hotkey → overlay < 100 ms.
*/
import React from "react";
import { AbsoluteFill, useCurrentFrame, interpolate, spring, useVideoConfig } from "remotion";
import { DARK, FONT } from "../theme";
import { StaffLines, Motes } from "../components/Background";
import { Window, Overlay, LightSweep } from "../components/Chrome";
import { Caption, CountUp, Annotation } from "../components/Primitives";

const CMD = "synaptic --data-dir ./build/data ping";
const OUT = [
  { tag: "▸", text: "overlay raised", detail: "87ms", brass: true },
  { tag: "▸", text: "orchestra detected", detail: "claude code · codex · ollama +5", brass: false },
  { tag: "▸", text: "gatekeeper", detail: "armed", brass: false },
  { tag: "↳", text: "pong", detail: "", brass: true },
];

const KeyCap: React.FC<{ ch: string; pressed: number }> = ({ ch, pressed }) => (
  <div
    style={{
      minWidth: 58,
      height: 58,
      padding: "0 16px",
      display: "grid",
      placeItems: "center",
      borderRadius: 10,
      background: DARK.bg3,
      border: `1px solid ${DARK.lineStrong}`,
      fontFamily: FONT.mono,
      fontSize: 22,
      color: pressed > 0.5 ? DARK.accent : DARK.fg,
      transform: `translateY(${pressed * 3}px)`,
      boxShadow: pressed > 0.5 ? `0 0 22px -4px ${DARK.glow}, inset 0 -3px 0 ${DARK.accentDeep}` : `inset 0 -4px 0 ${DARK.bg}`,
    }}
  >
    {ch}
  </div>
);

export const Hotkey: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();

  const press = interpolate(frame % 60, [0, 8, 20], [0, 1, 0], { extrapolateRight: "clamp" });
  const overlayS = spring({ frame: frame - 14, fps, config: { damping: 16, mass: 0.8 } });

  const typed = Math.floor(interpolate(frame, [30, 30 + CMD.length * 1.4], [0, CMD.length], { extrapolateLeft: "clamp", extrapolateRight: "clamp" }));
  const doneTyping = typed >= CMD.length;
  const linesShown = doneTyping ? Math.floor(interpolate(frame, [70, 130], [0, OUT.length], { extrapolateLeft: "clamp", extrapolateRight: "clamp" })) : 0;

  return (
    <AbsoluteFill style={{ background: DARK.bg }}>
      <StaffLines theme={DARK} opacity={0.5} />
      <Motes glow={DARK.glow} count={16} />
      <LightSweep start={0} />

      {/* hotkey chip up top */}
      <AbsoluteFill style={{ alignItems: "center", justifyContent: "flex-start", paddingTop: 96 }}>
        <div style={{ display: "flex", flexDirection: "column", alignItems: "center", gap: 14 }}>
          <Annotation>press to summon</Annotation>
          <div style={{ display: "flex", gap: 10 }}>
            <KeyCap ch="⌘" pressed={press} />
            <KeyCap ch="⇧" pressed={press} />
            <KeyCap ch="Space" pressed={press} />
          </div>
        </div>
      </AbsoluteFill>

      {/* split: overlay + terminal */}
      <AbsoluteFill style={{ alignItems: "center", justifyContent: "center", gap: 56, marginTop: 30 }}>
        <div style={{ width: 640, opacity: overlayS, transform: `translateY(${(1 - overlayS) * 40}px)` }}>
          <Overlay theme={DARK} state="Listening…" width={640} pulse={interpolate(Math.sin(frame / 6), [-1, 1], [0.2, 0.8])} />
        </div>

        <Window title="synapticd — local" theme={DARK} width={680} accentLight>
          <div style={{ padding: "22px 26px", fontFamily: FONT.mono, fontSize: 18, lineHeight: 1.9, minHeight: 230 }}>
            <div style={{ color: DARK.fg }}>
              <span style={{ color: DARK.fgFaint }}>$ </span>
              {CMD.slice(0, typed)}
              {!doneTyping && <Caret />}
            </div>
            {OUT.slice(0, linesShown).map((l, i) => (
              <div key={i} style={{ color: DARK.fgDim }}>
                <span style={{ color: DARK.accent }}>{l.tag} </span>
                {l.text}
                {l.detail && (
                  <>
                    <span style={{ color: DARK.fgFaint }}> ·· </span>
                    <span style={{ color: l.brass ? DARK.accent : DARK.fg }}>{l.detail}</span>
                  </>
                )}
                {i === linesShown - 1 && doneTyping && <Caret />}
              </div>
            ))}
          </div>
        </Window>
      </AbsoluteFill>

      {/* latency badges */}
      <div style={{ position: "absolute", left: 0, right: 0, bottom: 180, display: "flex", justifyContent: "center", gap: 60, alignItems: "flex-end" }}>
        <Metric label="cold start" to={500} start={96} />
        <Metric label="hotkey → overlay" to={100} start={108} />
        <Metric label="first token" to={1.5} unit="s" decimals={1} start={120} />
      </div>

      <Caption start={6} end={168}>
        Press your hotkey — the overlay appears instantly, ready to listen.
      </Caption>
    </AbsoluteFill>
  );
};

const Metric: React.FC<{ label: string; to: number; unit?: string; decimals?: number; start: number }> = ({ label, to, unit = "ms", decimals = 0, start }) => (
  <div style={{ textAlign: "center" }}>
    <CountUp to={to} unit={unit} decimals={decimals} start={start} theme={DARK} fontSize={56} />
    <div style={{ marginTop: 4 }}>
      <Annotation>{label}</Annotation>
    </div>
  </div>
);

const Caret: React.FC = () => {
  const frame = useCurrentFrame();
  return (
    <span
      style={{
        display: "inline-block",
        width: 10,
        height: 20,
        marginLeft: 3,
        verticalAlign: "text-bottom",
        background: DARK.accent,
        opacity: Math.floor(frame / 16) % 2 ? 0 : 1,
      }}
    />
  );
};
