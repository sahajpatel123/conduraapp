import { ImageResponse } from "next/og";
import { SITE } from "@/lib/site";

export const size = { width: 1200, height: 630 };
export const contentType = "image/png";
export const alt = `${SITE.name} — ${SITE.tagline}`;

/* The link card: ink field, staff lines, the brass diamond, one sentence. */
export default function OpenGraphImage() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "space-between",
          background: "#0b0b0e",
          padding: 72,
          position: "relative",
        }}
      >
        {[140, 165, 190, 215, 240].map((y) => (
          <div
            key={y}
            style={{
              position: "absolute",
              left: 0,
              right: 0,
              top: y,
              height: 1,
              background: "rgba(237,232,221,0.10)",
            }}
          />
        ))}
        <div style={{ display: "flex", alignItems: "center", gap: 18 }}>
          <div
            style={{
              width: 22,
              height: 22,
              background: "#e8a33d",
              transform: "rotate(45deg)",
            }}
          />
          <div style={{ color: "#ede8dd", fontSize: 34, letterSpacing: -0.5 }}>
            Synaptic
          </div>
        </div>
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            color: "#ede8dd",
            fontSize: 84,
            lineHeight: 1.05,
            letterSpacing: -2,
          }}
        >
          <div>Every AI on your</div>
          <div style={{ display: "flex" }}>
            machine.&nbsp;One&nbsp;
            <span style={{ color: "#e8a33d" }}>conductor.</span>
          </div>
        </div>
        <div
          style={{
            color: "#989284",
            fontSize: 24,
            letterSpacing: 5,
            textTransform: "uppercase",
          }}
        >
          free forever · no telemetry · local-first
        </div>
      </div>
    ),
    size,
  );
}
