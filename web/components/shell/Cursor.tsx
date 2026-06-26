"use client";

import { useEffect, useRef, useState } from "react";

/**
 * Cursor — a custom pixel-accurate cursor.
 *
 * The actual pointer is driven by the OS via CSS `cursor: url(...)` so
 * there is zero JS lag and never a double-dot: the SVG quill is the
 * real cursor, at the real hot spot.
 *
 * This component adds ONE optional layer on top: a soft synapse halo
 * that trails the pointer with a gentle lerp and only becomes visible
 * over interactive elements (links, buttons, [data-cursor]). It is
 * decorative — a quiet "breathing" ring, not a second dot.
 *
 * Disabled on touch devices and when prefers-reduced-motion.
 */
export default function Cursor() {
  const ringRef = useRef<HTMLDivElement | null>(null);
  const [enabled, setEnabled] = useState(false);

  useEffect(() => {
    const prefersReduced = window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;
    const fine = window.matchMedia("(pointer: fine)").matches;
    if (prefersReduced || !fine) return;
    setEnabled(true);

    const ring = ringRef.current;
    if (!ring) return;

    let tx = window.innerWidth / 2;
    let ty = window.innerHeight / 2;
    let cx = tx;
    let cy = ty;
    let hovering = false;
    let raf = 0;

    const onMove = (e: PointerEvent) => {
      tx = e.clientX;
      ty = e.clientY;
    };

    const onOver = (e: PointerEvent) => {
      const target = e.target as Element | null;
      if (!target || !ring) return;
      const interactive = target.closest(
        'a, button, [role="button"], input, textarea, select, .clickable, [data-cursor="hover"]'
      );
      hovering = !!interactive;
      ring.classList.toggle("is-hovering", hovering);
    };

    const onLeave = () => {
      ring.style.opacity = "0";
    };
    const onEnter = () => {
      ring.style.opacity = "1";
    };

    const loop = () => {
      if (!ring) return;
      // Lerp the halo toward the pointer — a soft, lagging ring
      cx += (tx - cx) * 0.16;
      cy += (ty - cy) * 0.16;
      ring.style.transform = `translate(${cx}px, ${cy}px) translate(-50%, -50%)`;
      raf = requestAnimationFrame(loop);
    };
    raf = requestAnimationFrame(loop);

    window.addEventListener("pointermove", onMove, { passive: true });
    window.addEventListener("pointerover", onOver, { passive: true });
    document.addEventListener("pointerleave", onLeave);
    document.addEventListener("pointerenter", onEnter);

    return () => {
      cancelAnimationFrame(raf);
      window.removeEventListener("pointermove", onMove);
      window.removeEventListener("pointerover", onOver);
      document.removeEventListener("pointerleave", onLeave);
      document.removeEventListener("pointerenter", onEnter);
    };
  }, []);

  if (!enabled) return null;

  return (
    <div ref={ringRef} className="condura-halo" aria-hidden>
      <div className="condura-halo-ring" />
      <div className="condura-halo-core" />
    </div>
  );
}
