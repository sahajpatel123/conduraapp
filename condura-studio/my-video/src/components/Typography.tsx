// src/components/Typography.tsx
//
// Three voices, three jobs. White-room treatment.
//
// All inline styles so we never load a custom webfont. Thin display at
// 184px / Inter 200 / Helvetica Neue Thin earns the white surface
// without ever needing a custom typeface.

import type { CSSProperties, ReactNode } from "react";

import { videoConfig } from "../config/video.config";

type Sx = Omit<CSSProperties, "fontFamily">;

export function Eyebrow({
  children,
  style,
  accent,
}: {
  children: ReactNode;
  style?: Sx;
  /** Render the label in azurite instead of ink — for the *one* eyebrow that earns it. */
  accent?: boolean;
}) {
  const { monoFamily, tracking } = videoConfig.type;
  return (
    <div
      style={{
        fontFamily: monoFamily,
        fontSize: 13,
        fontWeight: 600,
        lineHeight: 1,
        letterSpacing: tracking.eyebrow,
        textTransform: "uppercase",
        color: accent ? videoConfig.palette.accent : videoConfig.palette.inkMutedSoft,
        ...style,
      }}
    >
      {children}
    </div>
  );
}

export function Display({
  children,
  as: As = "h1",
  style,
}: {
  children: ReactNode;
  as?: "h1" | "h2" | "div";
  style?: Sx;
}) {
  const { displayFamily, tracking } = videoConfig.type;
  return (
    <As
      style={{
        margin: 0,
        fontFamily: displayFamily,
        fontSize: 184,
        fontWeight: 200,
        lineHeight: 0.94,
        letterSpacing: tracking.display,
        color: videoConfig.palette.ink,
        ...style,
      }}
    >
      {children}
    </As>
  );
}

export function Body({
  children,
  style,
}: {
  children: ReactNode;
  style?: Sx;
}) {
  const { bodyFamily, tracking } = videoConfig.type;
  return (
    <p
      style={{
        margin: 0,
        fontFamily: bodyFamily,
        fontSize: 22,
        fontWeight: 400,
        lineHeight: 1.5,
        letterSpacing: tracking.body,
        color: videoConfig.palette.inkMuted,
        ...style,
      }}
    >
      {children}
    </p>
  );
}

export function Label({
  children,
  style,
}: {
  children: ReactNode;
  style?: Sx;
}) {
  const { monoFamily, tracking } = videoConfig.type;
  return (
    <span
      style={{
        fontFamily: monoFamily,
        fontSize: 11,
        fontWeight: 600,
        lineHeight: 1,
        letterSpacing: tracking.eyebrow,
        textTransform: "uppercase",
        color: videoConfig.palette.inkMutedSoft,
        ...style,
      }}
    >
      {children}
    </span>
  );
}

/**
 * Numberly — for the 01 / 02 / 03 numerals used in the manifestos.
 * Inter ExtraLight @ 96px with negative tracking gives the numerals
 * a "stamped" feel — architecture, not editorial.
 */
export function Number({
  children,
  style,
}: {
  children: ReactNode;
  style?: Sx;
}) {
  const { displayFamily, tracking } = videoConfig.type;
  return (
    <span
      style={{
        fontFamily: displayFamily,
        fontSize: 96,
        fontWeight: 200,
        lineHeight: 1,
        letterSpacing: tracking.display,
        color: videoConfig.palette.inkMuted,
        fontVariantNumeric: "tabular-nums",
        ...style,
      }}
    >
      {children}
    </span>
  );
}
