"use client";

/*
  Latency numbers that count up when they enter view — speed you watch
  arrive. Reduced motion renders the final number outright.
*/
import { animate, m, useInView, useMotionValue, useTransform } from "motion/react";
import { useEffect, useRef } from "react";
import { usePrefersReducedMotion } from "@/lib/use-reduced-motion";

export function Counter({
  to,
  unit,
  decimals = 0,
  prefix = "<",
}: {
  to: number;
  unit: string;
  decimals?: number;
  prefix?: string;
}) {
  const ref = useRef<HTMLSpanElement>(null);
  const inView = useInView(ref, { once: true, amount: 0.6 });
  const reduced = usePrefersReducedMotion();
  const value = useMotionValue(0);
  const text = useTransform(value, (v) => v.toFixed(decimals));

  useEffect(() => {
    if (!inView) return;
    if (reduced) {
      value.set(to);
      return;
    }
    const controls = animate(value, to, { duration: 1.4, ease: [0.16, 1, 0.3, 1] });
    return () => controls.stop();
  }, [inView, reduced, to, value]);

  return (
    <span ref={ref} className="font-display text-4xl font-bold tabular-nums md:text-5xl">
      {prefix}
      <m.span>{text}</m.span>
      <span className="ml-1 text-2xl text-ivory-dim md:text-3xl">{unit}</span>
    </span>
  );
}
