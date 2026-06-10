"use client";

/*
  3D tilt with a light-source shine that follows the pointer across the
  surface. Pure transforms; disabled for touch and reduced motion.
*/
import { m, useSpring, useTransform } from "motion/react";
import { useRef, type PointerEvent, type ReactNode } from "react";
import { usePrefersReducedMotion } from "@/lib/use-reduced-motion";

export function Tilt({
  children,
  max = 7,
  className = "",
}: {
  children: ReactNode;
  max?: number;
  className?: string;
}) {
  const ref = useRef<HTMLDivElement>(null);
  const reduced = usePrefersReducedMotion();
  const px = useSpring(0.5, { stiffness: 220, damping: 24 });
  const py = useSpring(0.5, { stiffness: 220, damping: 24 });
  const rotateX = useTransform(py, [0, 1], [max, -max]);
  const rotateY = useTransform(px, [0, 1], [-max, max]);

  function onPointerMove(e: PointerEvent<HTMLDivElement>) {
    if (reduced || e.pointerType !== "mouse" || !ref.current) return;
    const r = ref.current.getBoundingClientRect();
    const x = (e.clientX - r.left) / r.width;
    const y = (e.clientY - r.top) / r.height;
    px.set(x);
    py.set(y);
    ref.current.style.setProperty("--mx", `${x * 100}%`);
    ref.current.style.setProperty("--my", `${y * 100}%`);
  }

  function onPointerLeave() {
    px.set(0.5);
    py.set(0.5);
  }

  return (
    <m.div
      ref={ref}
      onPointerMove={onPointerMove}
      onPointerLeave={onPointerLeave}
      style={reduced ? undefined : { rotateX, rotateY, transformPerspective: 900 }}
      className={`shine relative ${className}`}
    >
      {children}
    </m.div>
  );
}
