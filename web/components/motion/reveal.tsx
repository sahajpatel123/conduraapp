"use client";

/*
  Typographic reveal primitives. Display type enters through a mask
  (lines slide up out of an overflow-hidden clip); body content rises.
*/
import { m } from "motion/react";
import type { ReactNode } from "react";
import { lineReveal, rise, stagger, VIEWPORT } from "@/lib/motion";

/** Each child string renders as one masked line of display type. */
export function Lines({
  lines,
  className = "",
  lineClassName = "",
  as: Tag = "h2",
  delay = 0,
}: {
  lines: ReactNode[];
  className?: string;
  lineClassName?: string;
  as?: "h1" | "h2" | "h3" | "p";
  delay?: number;
}) {
  return (
    <m.div
      initial="hidden"
      whileInView="visible"
      viewport={VIEWPORT}
      variants={stagger(delay, 0.12)}
    >
      <Tag className={className}>
        {lines.map((line, i) => (
          <span key={i} className="block overflow-hidden pb-[0.08em] -mb-[0.08em]">
            <m.span variants={lineReveal} className={`block ${lineClassName}`}>
              {line}
            </m.span>
          </span>
        ))}
      </Tag>
    </m.div>
  );
}

/** Generic fade-and-rise block. */
export function Rise({
  children,
  className = "",
  delay = 0,
}: {
  children: ReactNode;
  className?: string;
  delay?: number;
}) {
  return (
    <m.div
      initial="hidden"
      whileInView="visible"
      viewport={VIEWPORT}
      variants={stagger(delay)}
      className={className}
    >
      <m.div variants={rise}>{children}</m.div>
    </m.div>
  );
}

/** Parent that staggers multiple Rise-like children. */
export function Cascade({
  children,
  className = "",
  delay = 0,
  step = 0.09,
}: {
  children: ReactNode;
  className?: string;
  delay?: number;
  step?: number;
}) {
  return (
    <m.div
      initial="hidden"
      whileInView="visible"
      viewport={VIEWPORT}
      variants={stagger(delay, step)}
      className={className}
    >
      {children}
    </m.div>
  );
}

/** A child item for Cascade. */
export function Item({
  children,
  className = "",
}: {
  children: ReactNode;
  className?: string;
}) {
  return (
    <m.div variants={rise} className={className}>
      {children}
    </m.div>
  );
}
