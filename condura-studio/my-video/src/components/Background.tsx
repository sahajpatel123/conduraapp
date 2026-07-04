// src/components/Background.tsx
//
// White-room canvas. Just pure white — no washes, no gradient drift.
// A single, almost-invisible cool gradient at 1/8 opacity anchors the
// top-left corner; if you can't see it on a calibrated display, leave
// the file. Its job is only to break up the absolute zero of #FFFFFF
// for cameras that roll-off the top end.
//
// Scenes never paint their own background — they paint *over* this.

import { AbsoluteFill } from "remotion";

import { videoConfig } from "../config/video.config";

export function Background() {
  return (
    <AbsoluteFill
      style={{
        background: videoConfig.palette.bg,
      }}
    >
      {/* Imperceptible cool wash anchored to the top-left */}
      <div
        aria-hidden
        style={{
          position: "absolute",
          left: "-12%",
          top: "-22%",
          width: 1100,
          height: 1100,
          borderRadius: "50%",
          background: "rgba(27, 77, 255, 0.025)",
          filter: "blur(120px)",
        }}
      />
    </AbsoluteFill>
  );
}
