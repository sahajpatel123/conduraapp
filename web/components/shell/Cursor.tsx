"use client";

import { useEffect, useRef } from "react";

/**
 * Cursor — a small ink dot that morphs into a synapse ring with a pollen
 * spark when hovering interactive elements. Disabled on touch devices and
 * when prefers-reduced-motion.
 *
 * Implementation note: we drive a single rAF loop and lerp toward the
 * pointer for a subtle trailing feel. The dot itself is direct-driven
 * (no lag) so clicks never feel off.
 */
export default function Cursor() {
  const rootRef = useRef<HTMLDivElement | null>(null);
  const dotRef = useRef<HTMLDivElement | null>(null);

  useEffect(() => {
    const prefersReduced = window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;
    const fine = window.matchMedia("(pointer: fine)").matches;
    if (prefersReduced || !fine) return;

    const root = rootRef.current!;
    const dot = dotRef.current!;
    let visible = false;

    let tx = window.innerWidth / 2;
    let ty = window.innerHeight / 2;
    let cx = tx;
    let cy = ty;
    let raf = 0;

    const onMove = (e: PointerEvent) => {
      tx = e.clientX;
      ty = e.clientY;
      if (!visible) {
        visible = true;
        root.style.opacity = "1";
      }
      // direct-drive the dot for zero latency
      dot.style.transform = `translate(${tx}px, ${ty}px) translate(-50%, -50%)`;
    };

    const onOver = (e: PointerEvent) => {
      const target = e.target as Element | null;
      if (!target) return;
      const interactive = target.closest(
        'a, button, [role="button"], input, textarea, select, .clickable, [data-cursor="hover"]'
      );
      root.classList.toggle("hovering", !!interactive);
    };

    const onLeave = () => {
      visible = false;
      root.style.opacity = "0";
    };

    const onDown = () => root.classList.add("pressing");
    const onUp = () => root.classList.remove("pressing");

    const loop = () => {
      // lerp the outer ring toward the pointer for a soft trailing feel
      cx += (tx - cx) * 0.18;
      cy += (ty - cy) * 0.18;
      root.style.transform = `translate(${cx}px, ${cy}px) translate(-50%, -50%)`;
      raf = requestAnimationFrame(loop);
    };
    raf = requestAnimationFrame(loop);

    window.addEventListener("pointermove", onMove, { passive: true });
    window.addEventListener("pointerover", onOver, { passive: true });
    document.addEventListener("pointerleave", onLeave);
    window.addEventListener("pointerdown", onDown, { passive: true });
    window.addEventListener("pointerup", onUp, { passive: true });

    return () => {
      cancelAnimationFrame(raf);
      window.removeEventListener("pointermove", onMove);
      window.removeEventListener("pointerover", onOver);
      document.removeEventListener("pointerleave", onLeave);
      window.removeEventListener("pointerdown", onDown);
      window.removeEventListener("pointerup", onUp);
    };
  }, []);

  return (
    <div ref={rootRef} className="condura-cursor" style={{ opacity: 0 }}>
      <div ref={dotRef} className="condura-cursor-dot" />
      <div className="condura-cursor-spark" />
    </div>
  );
}
