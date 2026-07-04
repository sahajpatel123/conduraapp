// src/components/ArchitecturalRule.tsx
//
// A horizontal hairline rule with optional numbered tick marks.
// The workhorse component for the White Room: timelines, dividers,
// provider-row separators. Use `<Ticks>` inline to place marks.
//
// Pure presentational — animation timing lives in the scene.

import type { CSSProperties, ReactNode } from "react";

import { Number } from "./Typography";
import { videoConfig } from "../config/video.config";

export function Rule({
  children,
  style,
  color = videoConfig.palette.line,
  weight = 1,
}: {
  /** Optional content to drop *on* the rule — usually `<Tick>` elements. */
  children?: ReactNode;
  style?: CSSProperties;
  color?: string;
  weight?: number;
}) {
  return (
    <div
      style={{
        position: "relative",
        width: "100%",
        height: weight,
        background: color,
        ...style,
      }}
    >
      {children}
    </div>
  );
}

export function Tick({
  /** Position along the rule, 0..1. */
  at,
  height = 16,
  weight = 1.5,
  color = videoConfig.palette.ink,
  style,
}: {
  at: number;
  height?: number;
  weight?: number;
  color?: string;
  style?: CSSProperties;
}) {
  return (
    <span
      aria-hidden
      style={{
        position: "absolute",
        left: `${at * 100}%`,
        top: "50%",
        width: weight,
        height,
        background: color,
        transform: "translate(-50%, -50%)",
        ...style,
      }}
    />
  );
}

/**
 * A header+number combo used to introduce a manifesto row.
 *   <RowHeader n="01">Local when it can.</RowHeader>
 * Renders the numeral in azurite (one chromatic element per row).
 */
export function RowHeader({ n, children }: { n: string; children: ReactNode }) {
  return (
    <div
      style={{
        display: "grid",
        gridTemplateColumns: "120px 1fr",
        alignItems: "baseline",
        columnGap: 48,
        padding: "44px 0",
        borderBottom: `1px solid ${videoConfig.palette.line}`,
      }}
    >
      <span
        style={{
          fontFamily: videoConfig.type.monoFamily,
          fontSize: 12,
          fontWeight: 600,
          letterSpacing: videoConfig.type.tracking.eyebrow,
          color: videoConfig.palette.accent,
        }}
      >
        {n}
      </span>
      <div
        style={{
          fontFamily: videoConfig.type.displayFamily,
          fontSize: 56,
          fontWeight: 200,
          lineHeight: 1.1,
          letterSpacing: videoConfig.type.tracking.display,
          color: videoConfig.palette.ink,
        }}
      >
        {children}
      </div>
    </div>
  );
}

// Suppress unused warning for Number — it's re-exported via the
// Typography barrel; keeping the dependency explicit for future use.
void Number;
