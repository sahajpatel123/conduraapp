/*
  Synaptic — "The Free AI Conductor". A ~61s cinematic demo.

  One continuous world, dark and warm, lit by a single filament. The story:
  genius in the dark → the touch → the overlay → the voice → how it perceives,
  verifies and gates every action → the work it does → the proof and the
  kill-switch → the wider orchestra → the call to conduct.
*/
import React from "react";
import { AbsoluteFill, Sequence, Audio, staticFile, useCurrentFrame, interpolate } from "remotion";
import { SCENES, TOTAL } from "./timeline";
import { DARK, FONT } from "./theme";
import { fontVars, useSynapticFonts } from "./fonts";
import { Grain, Vignette } from "./components/Background";

import { Ignition } from "./scenes/Ignition";
import { Problem } from "./scenes/Problem";
import { Touch } from "./scenes/Touch";
import { Hotkey } from "./scenes/Hotkey";
import { Voice } from "./scenes/Voice";
import { Perception } from "./scenes/Perception";
import { Montage } from "./scenes/Montage";
import { Audit } from "./scenes/Audit";
import { Features } from "./scenes/Features";
import { Cta } from "./scenes/Cta";
import { Outro } from "./scenes/Outro";

// The brass reading-progress baton across the very top (from web's <Baton/>).
const Baton: React.FC = () => {
  const frame = useCurrentFrame();
  const scaleX = interpolate(frame, [0, TOTAL], [0, 1]);
  return (
    <div
      style={{
        position: "absolute",
        top: 0,
        left: 0,
        right: 0,
        height: 2,
        transformOrigin: "left",
        transform: `scaleX(${scaleX})`,
        background: DARK.accent,
        zIndex: 100,
        boxShadow: `0 0 10px ${DARK.glow}`,
      }}
    />
  );
};

export const DemoVideo: React.FC = () => {
  useSynapticFonts();
  return (
    <AbsoluteFill style={{ background: DARK.bg, fontFamily: FONT.sans, ...fontVars() }}>
      {/* Optional narration + score. Drop files in public/ to enable. */}
      <OptionalAudio src="voiceover.mp3" volume={1} />
      <OptionalAudio src="music.mp3" volume={0.32} />

      <Sequence from={SCENES.ignition.from} durationInFrames={SCENES.ignition.duration}>
        <Ignition />
      </Sequence>
      <Sequence from={SCENES.problem.from} durationInFrames={SCENES.problem.duration}>
        <Problem />
      </Sequence>
      <Sequence from={SCENES.touch.from} durationInFrames={SCENES.touch.duration}>
        <Touch />
      </Sequence>
      <Sequence from={SCENES.hotkey.from} durationInFrames={SCENES.hotkey.duration}>
        <Hotkey />
      </Sequence>
      <Sequence from={SCENES.voice.from} durationInFrames={SCENES.voice.duration}>
        <Voice />
      </Sequence>
      <Sequence from={SCENES.perception.from} durationInFrames={SCENES.perception.duration}>
        <Perception />
      </Sequence>
      <Sequence from={SCENES.montage.from} durationInFrames={SCENES.montage.duration}>
        <Montage />
      </Sequence>
      <Sequence from={SCENES.audit.from} durationInFrames={SCENES.audit.duration}>
        <Audit />
      </Sequence>
      <Sequence from={SCENES.features.from} durationInFrames={SCENES.features.duration}>
        <Features />
      </Sequence>
      <Sequence from={SCENES.cta.from} durationInFrames={SCENES.cta.duration}>
        <Cta />
      </Sequence>
      <Sequence from={SCENES.outro.from} durationInFrames={SCENES.outro.duration}>
        <Outro />
      </Sequence>

      <Vignette strength={0.5} />
      <Grain opacity={0.045} />
      <Baton />
    </AbsoluteFill>
  );
};

// Renders an <Audio> only if the asset is present, so the video works with or
// without the separately-produced VO and score.
const OptionalAudio: React.FC<{ src: string; volume: number }> = ({ src, volume }) => {
  const [ok, setOk] = React.useState(false);
  React.useEffect(() => {
    let alive = true;
    fetch(staticFile(src), { method: "HEAD" })
      .then((r) => alive && setOk(r.ok))
      .catch(() => {});
    return () => {
      alive = false;
    };
  }, [src]);
  if (!ok) return null;
  return <Audio src={staticFile(src)} volume={volume} />;
};
