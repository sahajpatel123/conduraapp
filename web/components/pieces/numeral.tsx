"use client";

/*
  The movement numerals — outlined glyphs the height of the section,
  drifting slower than the text in front of them.
*/
import { m, useScroll, useTransform } from "motion/react";
import { useRef, type ReactNode } from "react";
import { usePrefersReducedMotion } from "@/lib/use-reduced-motion";

export function NumeralSection({
  numeral,
  children,
  className = "",
}: {
  numeral: string;
  children: ReactNode;
  className?: string;
}) {
  const ref = useRef<HTMLElement>(null);
  const reduced = usePrefersReducedMotion();
  const { scrollYProgress } = useScroll({
    target: ref,
    offset: ["start end", "end start"],
  });
  const y = useTransform(scrollYProgress, [0, 1], ["6%", "-14%"]);

  return (
    <section ref={ref} className={`relative overflow-hidden ${className}`}>
      <m.span
        aria-hidden
        style={reduced ? undefined : { y }}
        className="numeral-outline pointer-events-none absolute -top-8 right-[-2%] z-0 hidden text-[24vw] leading-none md:block"
      >
        {numeral}
      </m.span>
      <div className="relative z-10">{children}</div>
    </section>
  );
}
