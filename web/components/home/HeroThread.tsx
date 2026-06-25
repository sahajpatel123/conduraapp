"use client";

import { useEffect, useRef } from "react";

/**
 * HeroThread — a thin synapse hairline that drifts across the hero
 * and bends gently toward the pointer. Subtle and mature: it never
 * grabs attention, it rewards a slow cursor sweep with a quiet flex.
 *
 * One rAF loop, one SVG path, pointer-lerped control points. Disabled
 * under reduced motion (the path is drawn static).
 */
export default function HeroThread() {
  const pathRef = useRef<SVGPathElement | null>(null);
  const pointer = useRef({ x: 0.5, y: 0.4, tx: 0.5, ty: 0.4 });

  useEffect(() => {
    const prefersReduced = window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;
    const path = pathRef.current;
    if (!path) return;
    if (prefersReduced) {
      path.setAttribute("d", base(0.5, 0.4));
      return;
    }

    let raf = 0;
    let t = 0;

    const onMove = (e: PointerEvent) => {
      pointer.current.tx = e.clientX / window.innerWidth;
      pointer.current.ty = e.clientY / window.innerHeight;
    };

    const loop = () => {
      t += 0.006;
      const p = pointer.current;
      // lerp pointer for a soft, lagging bend
      p.x += (p.tx - p.x) * 0.06;
      p.y += (p.ty - p.y) * 0.06;
      // base wave + pointer bias
      const bend = (p.x - 0.5) * 18;
      const lift = (p.y - 0.4) * 10;
      path.setAttribute("d", base(0.5 + Math.sin(t) * 0.02, 0.42, bend, lift, t));
      raf = requestAnimationFrame(loop);
    };
    raf = requestAnimationFrame(loop);

    window.addEventListener("pointermove", onMove, { passive: true });
    return () => {
      cancelAnimationFrame(raf);
      window.removeEventListener("pointermove", onMove);
    };
  }, []);

  return (
    <svg
      className="pointer-events-none absolute inset-0 z-[5] h-full w-full opacity-50"
      viewBox="0 0 100 100"
      preserveAspectRatio="none"
      aria-hidden
    >
      <path
        ref={pathRef}
        className="synapse-thread"
        d={base(0.5, 0.42)}
        strokeWidth="0.35"
      />
      <circle cx="50" cy="42" r="0.5" className="synapse-node" opacity="0.7" />
    </svg>
  );
}

/** Build a gentle horizontal wave path with an optional pointer bend. */
function base(
  midX: number,
  midY: number,
  bend = 0,
  lift = 0,
  t = 0
): string {
  const y = midY * 100 + lift;
  const cx1 = 22 + Math.sin(t) * 2 + bend * 0.4;
  const cy1 = y - 8 + Math.cos(t * 0.8) * 1.5;
  const cx2 = 78 + Math.cos(t * 0.9) * 2 - bend * 0.4;
  const cy2 = y + 8 + Math.sin(t * 0.7) * 1.5;
  return `M -2 ${y + 4} C ${cx1} ${cy1}, ${cx2} ${cy2}, 102 ${y - 4}`;
}
