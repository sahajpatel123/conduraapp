/*
  50.8–57.8s — The call to conduct. The mark settles, three doors open:
  download (coming soon), Discord, GitHub. Free AI, yours to command.
*/
import React from "react";
import { AbsoluteFill, useCurrentFrame, interpolate, spring, useVideoConfig } from "remotion";
import { DARK, FONT, EASE } from "../theme";
import { StaffLines, Motes } from "../components/Background";
import { Bulb } from "../components/Bulb";
import { Cta as CtaButton, Annotation } from "../components/Primitives";
import { Waveform } from "../components/Chrome";

export const Cta: React.FC = () => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const logoS = spring({ frame, fps, config: { damping: 18, mass: 0.9 } });
  const glow = interpolate(frame, [0, 40], [0.3, 1], { extrapolateRight: "clamp", easing: EASE });
  const headT = interpolate(frame, [24, 56], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp", easing: EASE });
  const sway = Math.sin(frame / 32) * 0.8;

  return (
    <AbsoluteFill style={{ background: DARK.bg }}>
      <StaffLines theme={DARK} opacity={0.45} />
      <Motes glow={DARK.glow} count={18} />

      <AbsoluteFill style={{ alignItems: "center", justifyContent: "center", flexDirection: "column" }}>
        {/* logo mark: small lit bulb + wordmark + waveform underline */}
        <div style={{ transform: `scale(${0.8 + logoS * 0.2})`, opacity: logoS, display: "flex", flexDirection: "column", alignItems: "center" }}>
          <div style={{ height: 150, marginBottom: -10 }}>
            <Bulb glow={glow} filament={1} sway={sway} fgDim={DARK.fgDim} fgFaint={DARK.fgFaint} bg3={DARK.bg3} width={150} />
          </div>
          <div style={{ fontFamily: FONT.display, fontWeight: 800, fontSize: 78, letterSpacing: "-0.03em", color: DARK.fg }}>Synaptic</div>
          <div style={{ width: 320, marginTop: 8, opacity: 0.8 }}>
            <Waveform color={DARK.accent} bars={36} height={26} energy={0.7} />
          </div>
        </div>

        {/* headline */}
        <div style={{ marginTop: 34, textAlign: "center", opacity: headT, transform: `translateY(${(1 - headT) * 16}px)` }}>
          <span style={{ fontFamily: FONT.serif, fontStyle: "italic", fontSize: 38, color: DARK.fgDim }}>
            Ready to conduct your own AI orchestra?
          </span>
        </div>

        {/* buttons */}
        <div style={{ marginTop: 40, display: "flex", gap: 22 }}>
          <Btn at={60} primary>
            <Dot /> Download — Coming Soon
          </Btn>
          <Btn at={72}>
            <DiscordGlyph /> Join Discord
          </Btn>
          <Btn at={84}>
            <StarGlyph /> Star on GitHub
          </Btn>
        </div>

        {/* url */}
        <div style={{ marginTop: 44, opacity: interpolate(frame, [96, 120], [0, 1], { extrapolateLeft: "clamp", extrapolateRight: "clamp" }) }}>
          <Annotation color={DARK.fgFaint}>synaptic.app · free forever · no tracking</Annotation>
        </div>
      </AbsoluteFill>
    </AbsoluteFill>
  );
};

const Btn: React.FC<{ children: React.ReactNode; at: number; primary?: boolean }> = ({ children, at, primary }) => {
  const frame = useCurrentFrame();
  const { fps } = useVideoConfig();
  const s = spring({ frame: frame - at, fps, config: { damping: 15 } });
  const g = interpolate(Math.sin(frame / 12), [-1, 1], [0.2, 1]);
  return (
    <div style={{ opacity: s, transform: `translateY(${(1 - s) * 16}px)` }}>
      <CtaButton theme={DARK} primary={primary} glow={primary ? g : 0}>
        <span style={{ display: "inline-flex", alignItems: "center", gap: 10 }}>{children}</span>
      </CtaButton>
    </div>
  );
};

const Dot: React.FC = () => <span style={{ width: 8, height: 8, borderRadius: 9999, background: DARK.accent, boxShadow: `0 0 10px ${DARK.glow}` }} />;
const StarGlyph: React.FC = () => (
  <svg viewBox="0 0 24 24" width={16} height={16} fill={DARK.fg} stroke="none">
    <path d="M12 2l2.9 6.3 6.9.8-5.1 4.7 1.4 6.8L12 17.8 5.9 21l1.4-6.8L2.2 9.1l6.9-.8z" />
  </svg>
);
const DiscordGlyph: React.FC = () => (
  <svg viewBox="0 0 24 24" width={17} height={17} fill={DARK.fg}>
    <path d="M19.5 5.5A16 16 0 0 0 15.5 4l-.2.4a14 14 0 0 1 3.3 1.6 13 13 0 0 0-11.2 0A14 14 0 0 1 10.7 4.4L10.5 4A16 16 0 0 0 6.5 5.5C3.7 9.6 3 13.6 3.3 17.5a16 16 0 0 0 4.9 2.5l.6-1a10 10 0 0 1-1.6-.8l.4-.3a11 11 0 0 0 9.6 0l.4.3a10 10 0 0 1-1.6.8l.6 1a16 16 0 0 0 4.9-2.5c.4-4.6-.7-8.5-3-12zM9.3 15.1c-.8 0-1.4-.7-1.4-1.6s.6-1.6 1.4-1.6 1.4.7 1.4 1.6-.6 1.6-1.4 1.6zm5.4 0c-.8 0-1.4-.7-1.4-1.6s.6-1.6 1.4-1.6 1.4.7 1.4 1.6-.6 1.6-1.4 1.6z" />
  </svg>
);
