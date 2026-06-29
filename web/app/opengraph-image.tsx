import { ImageResponse } from "next/og";
import { SITE } from "@/lib/site";

export const runtime = "edge";
export const alt = `${SITE.name} — One hotkey. Your AI tools. Free.`;
export const size = { width: 1200, height: 630 };
export const contentType = "image/png";

/**
 * Static Open Graph image (1200x630) for social previews.
 * Kept static — no per-page variants — so every URL shares
 * a single 1200x630 PNG. Pairs with /twitter-image.tsx for
 * the Twitter card.
 */
export default async function Image() {
  return new ImageResponse(
    (
      <div
        style={{
          width: "100%",
          height: "100%",
          display: "flex",
          flexDirection: "column",
          justifyContent: "center",
          alignItems: "flex-start",
          padding: "72px 80px",
          background:
            "linear-gradient(135deg, #08080c 0%, #16161d 60%, #1f1f2a 100%)",
          color: "#efeae0",
          fontFamily: "system-ui, -apple-system, sans-serif",
        }}
      >
        <div
          style={{
            display: "flex",
            alignItems: "center",
            fontSize: 28,
            color: "#ffc46b",
            letterSpacing: 4,
            marginBottom: 24,
            textTransform: "uppercase",
          }}
        >
          {SITE.name}
        </div>
        <div
          style={{
            display: "flex",
            fontSize: 84,
            fontWeight: 700,
            lineHeight: 1.05,
            marginBottom: 32,
            maxWidth: 960,
          }}
        >
          One hotkey. Your AI tools. Free.
        </div>
        <div
          style={{
            display: "flex",
            fontSize: 32,
            color: "#a09a8e",
            lineHeight: 1.4,
            maxWidth: 900,
          }}
        >
          The conductor for every AI tool on your machine.
        </div>
      </div>
    ),
    { ...size }
  );
}
