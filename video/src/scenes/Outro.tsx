/*
  57.8–61.0s — The mark lingers, the filament dims, and the room returns to the
  dark it came from. synaptic.app.
*/
import React from "react";
import { AbsoluteFill, useCurrentFrame, interpolate } from "remotion";
import { DARK, FONT, EASE } from "../theme";
import { Bulb } from "../components/Bulb";
import { Motes } from "../components/Background";

export const Outro: React.FC = () => {
  const frame = useCurrentFrame();
  // filament dims out
  const glow = interpolate(frame, [0, 30, 70], [1, 0.6, 0], { extrapolateRight: "clamp", easing: EASE });
  const filament = interpolate(frame, [20, 70], [1, 0], { extrapolateLeft: "clamp", extrapolateRight: "clamp" });
  const fade = interpolate(frame, [54, 90], [1, 0], { extrapolateLeft: "clamp", extrapolateRight: "clamp" });
  const urlT = interpolate(frame, [6, 30], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: EASE });

  return (
    <AbsoluteFill style={{ background: DARK.bg }}>
      <Motes glow={DARK.glow} count={12} />
      <AbsoluteFill style={{ alignItems: "center", justifyContent: "center", opacity: fade, flexDirection: "column" }}>
        <div style={{ height: 130, marginBottom: 4 }}>
          <Bulb glow={glow} filament={filament} sway={Math.sin(frame / 30) * 0.6} fgDim={DARK.fgDim} fgFaint={DARK.fgFaint} bg3={DARK.bg3} width={130} />
        </div>
        <div style={{ opacity: urlT, textAlign: "center" }}>
          <div style={{ fontFamily: FONT.display, fontWeight: 800, fontSize: 40, color: DARK.fg, letterSpacing: "-0.02em" }}>synaptic.app</div>
          <div style={{ marginTop: 10, fontFamily: FONT.mono, fontSize: 13, letterSpacing: "0.16em", textTransform: "uppercase", color: DARK.fgFaint }}>
            © 2026 Synaptic — Proprietary source, free binary
          </div>
        </div>
      </AbsoluteFill>
    </AbsoluteFill>
  );
};
