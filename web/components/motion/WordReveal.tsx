"use client";

import { motion, useInView } from "motion/react";
import { useRef, type ReactNode } from "react";
import { EASE_OUT } from "@/lib/motion";

/**
 * WordReveal — a per-word ink reveal for headlines.
 *
 * Each word rises out of its own mask with a tiny stagger, so the
 * headline "dries" word by word as it enters the viewport. More
 * premium than a single line-reveal, and still subtle.
 *
 * Pass `text` for a plain string, OR `children` for mixed content
 * (e.g. an italic accent word — wrap it in <span class="accent">).
 * `delay` offsets the whole reveal. `stagger` is per-word (default 40ms).
 */
export default function WordReveal({
  text,
  children,
  delay = 0,
  stagger = 0.04,
  className = "",
  as: Tag = "h2",
}: {
  text?: string;
  children?: ReactNode;
  delay?: number;
  stagger?: number;
  className?: string;
  as?: "h1" | "h2" | "h3" | "p";
}) {
  const ref = useRef<HTMLDivElement | null>(null);
  const inView = useInView(ref, { once: true, margin: "-12%" });

  // If text is provided, split into words. Otherwise, split children by
  // extracting text runs — simpler: only support `text` for the per-word
  // mask, and fall back to a single line-reveal for `children`.
  if (text) {
    const words = text.split(" ");
    return (
      <Tag ref={ref as never} className={className}>
        {words.map((w, i) => (
          <span key={i} className="inline-block overflow-hidden align-bottom pb-[0.16em]">
            <motion.span
              className="inline-block"
              initial={{ y: "110%" }}
              animate={inView ? { y: "0%" } : { y: "110%" }}
              transition={{ duration: 0.8, ease: EASE_OUT, delay: delay + i * stagger }}
            >
              {w}
              {i < words.length - 1 ? "\u00A0" : ""}
            </motion.span>
          </span>
        ))}
      </Tag>
    );
  }

  // children fallback: single line reveal
  return (
    <Tag ref={ref as never} className={`overflow-hidden ${className}`}>
      <motion.span
        className="inline-block"
        initial={{ y: "110%" }}
        animate={inView ? { y: "0%" } : { y: "110%" }}
        transition={{ duration: 0.9, ease: EASE_OUT, delay }}
      >
        {children}
      </motion.span>
    </Tag>
  );
}
