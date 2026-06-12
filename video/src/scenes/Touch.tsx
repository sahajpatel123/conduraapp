/*
  7.6–12.6s — The touch. A hand reaches in, one finger meets the glass, the
  filament catches, a bloom swallows the screen, and the room lights up.
  "One hotkey. Unlimited AI."
*/
import React from "react";
import { AbsoluteFill, useCurrentFrame, interpolate } from "remotion";
import { DARK, FONT, EASE } from "../theme";
import { Bulb, Hand } from "../components/Bulb";
import { StaffLines, Motes, Shafts } from "../components/Background";
import { Caption } from "../components/Primitives";

const TOUCH = 56; // frame of contact

export const Touch: React.FC = () => {
  const frame = useCurrentFrame();

  const handX = interpolate(frame, [0, TOUCH, 90, 130], [120, 8, 4, 120], {
    extrapolateLeft: "clamp",
    extrapolateRight: "clamp",
    easing: EASE,
  });

  const filament = interpolate(frame, [TOUCH, TOUCH + 8], [0.4, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp" });
  const glow = interpolate(frame, [TOUCH, TOUCH + 26], [0.4, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: EASE });

  // The bloom wash that masks the ignition.
  const washScale = interpolate(frame, [TOUCH, TOUCH + 40], [0.1, 7], { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: EASE });
  const washOp = interpolate(frame, [TOUCH, TOUCH + 18, TOUCH + 40, TOUCH + 64], [0, 0.95, 0.6, 0], { extrapolateLeft: "clamp", extrapolateRight: "clamp" });

  // Contact spark.
  const spark = interpolate(frame, [TOUCH - 3, TOUCH + 2, TOUCH + 12], [0, 1, 0], { extrapolateLeft: "clamp", extrapolateRight: "clamp" });

  const sway = Math.sin(frame / 30) * 1.1;
  const litWorld = interpolate(frame, [TOUCH + 30, TOUCH + 60], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: EASE });
  const titleT = interpolate(frame, [TOUCH + 40, TOUCH + 78], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: EASE });

  return (
    <AbsoluteFill style={{ background: DARK.bg }}>
      <StaffLines theme={DARK} opacity={0.6} />
      <Motes glow={DARK.glow} count={22} />
      <Shafts glow={DARK.glow} opacity={litWorld} />

      {/* bulb, up and to the left of center */}
      <AbsoluteFill style={{ alignItems: "center", justifyContent: "center", marginTop: -160 }}>
        <Bulb glow={glow} filament={filament} sway={sway} fgDim={DARK.fgDim} fgFaint={DARK.fgFaint} bg3={DARK.bg3} width={300} />
      </AbsoluteFill>

      {/* contact spark near the glass */}
      <AbsoluteFill style={{ alignItems: "center", justifyContent: "center", marginTop: -120 }}>
        <svg viewBox="0 0 60 60" width={120} height={120} style={{ opacity: spark, transform: `scale(${0.5 + spark})` }}>
          {[0, 45, 90, 135, 180, 225, 270, 315].map((a) => (
            <line
              key={a}
              x1="30"
              y1="30"
              x2={30 + 26 * Math.cos((a * Math.PI) / 180)}
              y2={30 + 26 * Math.sin((a * Math.PI) / 180)}
              stroke="#ffe2ae"
              strokeWidth="2.5"
              strokeLinecap="round"
            />
          ))}
        </svg>
      </AbsoluteFill>

      {/* the reaching hand */}
      <div style={{ position: "absolute", top: "16%", right: 0, transform: `translateX(${handX}%)` }}>
        <Hand fg={DARK.fg} fgDim={DARK.fgDim} bg2={DARK.bg2} />
      </div>

      {/* the bloom */}
      <AbsoluteFill style={{ alignItems: "center", justifyContent: "center", marginTop: -120 }}>
        <div
          style={{
            width: 700,
            height: 700,
            borderRadius: 9999,
            opacity: washOp,
            transform: `scale(${washScale})`,
            background: "radial-gradient(circle, #fff0d2 0%, #ffc46b 38%, rgba(255,196,107,0) 72%)",
          }}
        />
      </AbsoluteFill>

      {/* tagline revealed by the light */}
      <AbsoluteFill style={{ alignItems: "center", justifyContent: "flex-end", paddingBottom: 150 }}>
        <div style={{ textAlign: "center", opacity: titleT, transform: `translateY(${(1 - titleT) * 20}px)` }}>
          <div style={{ fontFamily: FONT.display, fontWeight: 800, fontSize: 92, letterSpacing: "-0.03em", color: DARK.fg, lineHeight: 1 }}>
            One hotkey.
          </div>
          <div style={{ fontFamily: FONT.serif, fontStyle: "italic", fontSize: 92, color: DARK.accent, lineHeight: 1.05 }}>
            Unlimited AI.
          </div>
        </div>
      </AbsoluteFill>

      <Caption start={TOUCH + 6} end={144} accent>
        Synaptic changes that. One agent unlocks every AI you already have.
      </Caption>
    </AbsoluteFill>
  );
};
