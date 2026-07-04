"use client";

import { useEffect, useRef } from "react";

/**
 * BrandSurface — the living canvas behind every page.
 *
 * Three layers, all GPU-friendly:
 *  1. Paper grain + warm light blooms (CSS, in globals.css `.surface-paper`)
 *  2. A slow drift of soft "pollen" specks (canvas, rAF, very cheap)
 *  3. A global synapse thread — an SVG path that gently sways and whose
 *     control points are nudged by pointer proximity.
 *
 * The intent is a *quiet* ambience. Not a screensaver. Not a vibe-coded
 * particle storm. A breathing paper that earns its keep.
 */
export default function BrandSurface() {
  const pollenRef = useRef<HTMLCanvasElement | null>(null);
  const threadRef = useRef<SVGSVGElement | null>(null);
  const pointer = useRef({ x: 0.5, y: 0.5, active: false });

  useEffect(() => {
    const prefersReduced = window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;
    if (prefersReduced) return;

    // ── Pollen canvas ──
    const canvas = pollenRef.current;
    if (!canvas) return;
    const ctx = canvas.getContext("2d", { alpha: true });
    if (!ctx) return;

    let raf = 0;
    let w = 0;
    let h = 0;
    const dpr = Math.min(window.devicePixelRatio || 1, 2);

    type Mote = {
      x: number;
      y: number;
      r: number;
      vx: number;
      vy: number;
      hue: number;
      life: number;
      max: number;
    };
    let motes: Mote[] = [];

    const resize = () => {
      w = canvas.clientWidth;
      h = canvas.clientHeight;
      canvas.width = Math.floor(w * dpr);
      canvas.height = Math.floor(h * dpr);
      ctx.setTransform(dpr, 0, 0, dpr, 0, 0);
      const target = Math.min(28, Math.floor((w * h) / 48000));
      motes = Array.from({ length: target }, () => spawn());
    };

    const spawn = (): Mote => {
      const max = 600 + Math.random() * 800;
      return {
        x: Math.random() * w,
        y: Math.random() * h,
        r: 0.6 + Math.random() * 1.8,
        vx: (Math.random() - 0.5) * 0.06,
        vy: -0.04 - Math.random() * 0.09,
        hue: Math.random() < 0.7 ? 38 : 160, // pollen or synapse-green
        life: Math.random() * max,
        max,
      };
    };

    const tick = () => {
      ctx.clearRect(0, 0, w, h);
      for (let i = 0; i < motes.length; i++) {
        const m = motes[i];
        m.life += 1;
        m.x += m.vx;
        m.y += m.vy;
        if (m.life > m.max || m.y < -20 || m.x < -20 || m.x > w + 20) {
          motes[i] = spawn();
          motes[i].y = h + 10;
          continue;
        }
        const t = m.life / m.max;
        const fade = Math.sin(Math.PI * t); // 0 → 1 → 0
        const alpha = fade * 0.5;
        ctx.beginPath();
        ctx.arc(m.x, m.y, m.r, 0, Math.PI * 2);
        if (m.hue === 38) {
          ctx.fillStyle = `rgba(201,123,46,${alpha})`;
        } else {
          ctx.fillStyle = `rgba(26,138,106,${alpha * 0.7})`;
        }
        ctx.fill();
      }
      raf = requestAnimationFrame(tick);
    };

    resize();
    window.addEventListener("resize", resize);
    raf = requestAnimationFrame(tick);

    // ── Thread pointer tracking ──
    const onPointer = (e: PointerEvent) => {
      pointer.current.x = e.clientX / window.innerWidth;
      pointer.current.y = e.clientY / window.innerHeight;
      pointer.current.active = true;
    };
    window.addEventListener("pointermove", onPointer, { passive: true });

    // ── Thread sway ──
    const svg = threadRef.current;
    const path = svg?.querySelector<SVGPathElement>("#global-thread");
    let threadRaf = 0;
    let t = 0;
    const sway = () => {
      t += 0.008;
      if (path) {
        const W = window.innerWidth;
        const H = window.innerHeight;
        const px = pointer.current.x;
        const py = pointer.current.y;
        // Base wave across the viewport, gently nudged by pointer
        const amp = 60 + py * 80;
        const midY = H * (0.32 + py * 0.12);
        const cx1 = W * 0.2 + Math.sin(t) * 40 + (px - 0.5) * 80;
        const cy1 = midY - amp + Math.cos(t * 0.8) * 20;
        const cx2 = W * 0.7 + Math.cos(t * 0.9) * 40 - (px - 0.5) * 80;
        const cy2 = midY + amp + Math.sin(t * 0.7) * 20;
        const d = `M -40 ${midY} C ${cx1} ${cy1}, ${cx2} ${cy2}, ${W + 40} ${midY}`;
        path.setAttribute("d", d);
      }
      threadRaf = requestAnimationFrame(sway);
    };
    threadRaf = requestAnimationFrame(sway);

    return () => {
      cancelAnimationFrame(raf);
      cancelAnimationFrame(threadRaf);
      window.removeEventListener("resize", resize);
      window.removeEventListener("pointermove", onPointer);
    };
  }, []);

  return (
    <div
      aria-hidden
      className="pointer-events-none fixed inset-0 z-0 overflow-hidden"
    >
      {/* pollen */}
      <canvas
        ref={pollenRef}
        className="absolute inset-0 h-full w-full"
        style={{ width: "100%", height: "100%" }}
      />
      <svg
        ref={threadRef}
        className="absolute inset-0 h-full w-full"
        preserveAspectRatio="none"
        viewBox="0 0 100 100"
        style={{ opacity: 0.18 }}
      >
        <path
          id="global-thread"
          className="synapse-thread"
          d="M -5 50 C 25 30, 75 70, 105 50"
          pathLength={1}
        />
      </svg>
    </div>
  );
}
