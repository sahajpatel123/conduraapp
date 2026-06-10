import { ImageResponse } from "next/og";
import { SITE } from "@/lib/site";

export const size = { width: 1200, height: 630 };
export const contentType = "image/png";
export const alt = `${SITE.name} — ${SITE.tagline}`;

/* The link card: a lit bulb in the dark, one sentence. */
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
          background: "#08080c",
          padding: 72,
          position: "relative",
        }}
      >
        {/* glow */}
        <div
          style={{
            position: "absolute",
            right: 60,
            top: 40,
            width: 360,
            height: 360,
            borderRadius: 360,
            background:
              "radial-gradient(circle, rgba(255,226,174,0.9) 0%, rgba(255,196,107,0.35) 45%, rgba(255,196,107,0) 70%)",
          }}
        />
        {/* bulb */}
        <svg
          width="170"
          height="240"
          viewBox="0 0 240 360"
          style={{ position: "absolute", right: 155, top: 60 }}
        >
          <line x1="120" y1="0" x2="120" y2="78" stroke="#97928a" strokeWidth="5" />
          <rect x="96" y="78" width="48" height="34" rx="7" fill="#16161d" stroke="#97928a" strokeWidth="5" />
          <path
            d="M100,112 C100,140 82,150 75,176 A62,62 0 1,0 165,176 C158,150 140,140 140,112 Z"
            fill="rgba(255,196,107,0.55)"
            stroke="#efeae0"
            strokeWidth="5"
          />
          <path
            d="M106,118 V162 M134,118 V162 M106,162 Q113,180 120,162 Q127,144 134,162"
            fill="none"
            stroke="#ffd9a0"
            strokeWidth="5"
            strokeLinecap="round"
          />
        </svg>
        <div style={{ display: "flex", alignItems: "center", gap: 16, color: "#efeae0", fontSize: 32 }}>
          Synaptic
        </div>
        <div
          style={{
            display: "flex",
            flexDirection: "column",
            color: "#efeae0",
            fontSize: 78,
            fontWeight: 700,
            lineHeight: 1.06,
            letterSpacing: -2,
            maxWidth: 760,
          }}
        >
          <div>One touch wakes</div>
          <div style={{ display: "flex" }}>
            every AI on&nbsp;<span style={{ color: "#ffa12e" }}>your machine.</span>
          </div>
        </div>
        <div
          style={{
            color: "#97928a",
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
