/*
  0.0–3.2s — Black, then a single filament catches and brightens. Staff lines
  bleed in. The title settles. "Imagine every AI you own, working together."
*/
import React from "react";
import { AbsoluteFill, useCurrentFrame, interpolate } from "remotion";
import { DARK, FONT, EASE } from "../theme";
import { Bulb } from "../components/Bulb";
import { StaffLines, Motes } from "../components/Background";
import { Caption } from "../components/Primitives";

export const Ignition: React.FC = () => {
  const frame = useCurrentFrame();

  // Filament catches with a flicker, then holds.
  const flicker = [0, 0.15, 0.05, 0.4, 0.2, 0.7, 0.5, 0.9, 0.8, 1];
  const fi = Math.min(flicker.length - 1, Math.floor(interpolate(frame, [10, 50], [0, flicker.length - 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp" })));
  const filament = flicker[fi];
  const glow = interpolate(frame, [30, 90], [0, 0.55], { extrapolateRight: "clamp", easing: EASE });
  const sway = Math.sin(frame / 34) * 1.2;

  const staff = interpolate(frame, [24, 70], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp" });
  const titleT = interpolate(frame, [40, 72], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: EASE });

  return (
    <AbsoluteFill style={{ background: DARK.bg }}>
      <StaffLines theme={DARK} opacity={staff * 0.8} />
      <Motes glow={DARK.glow} count={18} />

      <AbsoluteFill style={{ alignItems: "center", justifyContent: "center", marginTop: -120 }}>
        <Bulb glow={glow} filament={filament} sway={sway} fgDim={DARK.fgDim} fgFaint={DARK.fgFaint} bg3={DARK.bg3} width={300} />
      </AbsoluteFill>

      <AbsoluteFill style={{ alignItems: "center", justifyContent: "flex-end", paddingBottom: 200 }}>
        <div style={{ textAlign: "center", opacity: titleT, transform: `translateY(${(1 - titleT) * 16}px)` }}>
          <div style={{ fontFamily: FONT.display, fontWeight: 800, fontSize: 70, letterSpacing: "-0.02em", color: DARK.fg }}>
            Synaptic
          </div>
          <div style={{ fontFamily: FONT.serif, fontStyle: "italic", fontSize: 34, color: DARK.accent, marginTop: 6 }}>
            The Free AI Conductor
          </div>
        </div>
      </AbsoluteFill>

      <Caption start={6} end={90}>
        Imagine every AI you own — working together. On your computer. For free.
      </Caption>
    </AbsoluteFill>
  );
};
