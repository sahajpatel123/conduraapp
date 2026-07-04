"use client";

import { useRef, useState, type ReactNode } from "react";
import { motion } from "motion/react";

/**
 * MagneticButton — a wrapper that pulls its child gently toward the
 * pointer when the pointer is within `radius` px of the element center.
 * The pull is strongest at the center and fades to zero at the edge.
 *
 * The child should be the interactive element (a Next <Link> or <button>
 * with .btn classes). This wrapper only provides the magnetic motion.
 *
 * Respects reduced motion (no pull, just the child's own hover styles).
 */
export default function MagneticButton({
  children,
  radius = 80,
  strength = 0.35,
  className = "",
}: {
  children: ReactNode;
  radius?: number;
  strength?: number;
  className?: string;
}) {
  const ref = useRef<HTMLDivElement | null>(null);
  const [offset, setOffset] = useState({ x: 0, y: 0 });

  const onMove = (e: React.PointerEvent) => {
    const el = ref.current;
    if (!el) return;
    const r = el.getBoundingClientRect();
    const cx = r.left + r.width / 2;
    const cy = r.top + r.height / 2;
    const dx = e.clientX - cx;
    const dy = e.clientY - cy;
    const dist = Math.hypot(dx, dy);
    if (dist > radius) {
      setOffset({ x: 0, y: 0 });
      return;
    }
    const pull = (1 - dist / radius) * strength;
    setOffset({ x: dx * pull, y: dy * pull });
  };

  const onLeave = () => setOffset({ x: 0, y: 0 });

  return (
    <motion.div
      ref={ref}
      onPointerMove={onMove}
      onPointerLeave={onLeave}
      animate={{ x: offset.x, y: offset.y }}
      transition={{ type: "spring", stiffness: 260, damping: 22, mass: 0.6 }}
      className={`inline-block ${className}`}
    >
      {children}
    </motion.div>
  );
}
