"use client";

import { useEffect, useRef, useMemo } from "react";

/**
 * SynapseGarden — the generative hero scene.
 *
 * Not a photograph. A living illustration rendered in SVG:
 *  • rolling paper hills (3 layers, parallax via scroll-y of the section)
 *  • a single tree on the right crest (hand-drawn Bézier trunk + canopy)
 *  • a breathing sun upper-left
 *  • 2 light-trail threads that draw themselves on mount and re-draw on
 *    pointer move, ending in a glowing pollen node
 *  • drifting pollen motes (CSS-animated)
 *
 * Everything is vector. It scales to any viewport. It is cheap to render
 * and uses no images. It is *the brand*, not a stock asset.
 */
export default function SynapseGarden({
  className = "",
}: {
  className?: string;
}) {
  const wrapRef = useRef<HTMLDivElement | null>(null);
  const trail1Ref = useRef<SVGPathElement | null>(null);
  const trail2Ref = useRef<SVGPathElement | null>(null);
  const trail1GlowRef = useRef<SVGPathElement | null>(null);
  const trail2GlowRef = useRef<SVGPathElement | null>(null);
  const sunRef = useRef<SVGGElement | null>(null);
  const hillsRef = useRef<SVGGElement | null>(null);
  const treeRef = useRef<SVGGElement | null>(null);

  // Stable pollen positions (deterministic — no Math.random during render)
  const pollen = useMemo(
    () => {
      const rand = (seed: number) => {
        const x = Math.sin(seed * 12.9898) * 43758.5453;
        return x - Math.floor(x);
      };
      return Array.from({ length: 14 }, (_, i) => ({
        id: i,
        left: 4 + rand(i * 4 + 1) * 92,
        top: 50 + rand(i * 4 + 2) * 40,
        delay: rand(i * 4 + 3) * 8,
        dur: 9 + rand(i * 4 + 4) * 7,
        size: 2 + rand(i * 4 + 5) * 3,
        dx: (rand(i * 4 + 6) - 0.5) * 60,
        dy: -80 - rand(i * 4 + 7) * 80,
      }));
    },
    []
  );

  useEffect(() => {
    const prefersReduced = window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;
    if (prefersReduced) return;

    const trail1 = trail1Ref.current;
    const trail2 = trail2Ref.current;
    const trail1Glow = trail1GlowRef.current;
    const trail2Glow = trail2GlowRef.current;
    if (!trail1 || !trail2) return;

    // Initial draw-in
    [trail1, trail2, trail1Glow, trail2Glow].filter(Boolean).forEach((p) => {
      const len = p!.getTotalLength();
      p!.style.strokeDasharray = `${len}`;
      p!.style.strokeDashoffset = `${len}`;
      p!.getBoundingClientRect(); // reflow
      p!.style.transition = "stroke-dashoffset 2.4s cubic-bezier(0.22,1,0.36,1)";
      p!.style.strokeDashoffset = "0";
    });

    let raf = 0;
    let t = 0;
    const onPointer = (e: PointerEvent) => {
      const w = window.innerWidth;
      const h = window.innerHeight;
      const px = e.clientX / w; // 0..1
      const py = e.clientY / h;
      // Re-shape the trails subtly toward the pointer
      const cx = 50 + (px - 0.5) * 12;
      const cy = 60 + (py - 0.5) * 8;
      trail1.setAttribute(
        "d",
        `M -2 78 C 20 ${cy - 6}, 35 ${72 - (py * 8)}, ${cx} ${cy}`
      );
      trail2.setAttribute(
        "d",
        `M ${cx} ${cy} C 70 ${cy + 8}, 82 ${58 - (py * 6)}, 102 42`
      );
      trail1Glow?.setAttribute(
        "d",
        `M -2 78 C 20 ${cy - 6}, 35 ${72 - (py * 8)}, ${cx} ${cy}`
      );
      trail2Glow?.setAttribute(
        "d",
        `M ${cx} ${cy} C 70 ${cy + 8}, 82 ${58 - (py * 6)}, 102 42`
      );
    };

    const sway = () => {
      t += 0.01;
      if (treeRef.current) {
        const r = Math.sin(t) * 0.6;
        treeRef.current.style.transform = `rotate(${r}deg)`;
      }
      if (sunRef.current) {
        const s = 1 + Math.sin(t * 0.6) * 0.03;
        sunRef.current.style.transform = `scale(${s})`;
      }
      raf = requestAnimationFrame(sway);
    };
    raf = requestAnimationFrame(sway);

    window.addEventListener("pointermove", onPointer, { passive: true });
    return () => {
      cancelAnimationFrame(raf);
      window.removeEventListener("pointermove", onPointer);
    };
  }, []);

  // Parallax on scroll
  useEffect(() => {
    const prefersReduced = window.matchMedia(
      "(prefers-reduced-motion: reduce)"
    ).matches;
    if (prefersReduced) return;
    const onScroll = () => {
      const y = window.scrollY;
      if (hillsRef.current) {
        hillsRef.current.style.transform = `translateY(${y * 0.06}px)`;
      }
      if (sunRef.current) {
        sunRef.current.style.translate = `0 ${y * 0.04}px`;
      }
    };
    window.addEventListener("scroll", onScroll, { passive: true });
    return () => window.removeEventListener("scroll", onScroll);
  }, []);

  return (
    <div
      ref={wrapRef}
      className={`pointer-events-none absolute inset-0 overflow-hidden ${className}`}
      aria-hidden
    >
      {/* sky wash — the breath behind everything */}
      <div
        className="absolute inset-0"
        style={{
          background:
            "linear-gradient(180deg, #c8dee8 0%, #d8e2d0 38%, #ece5d4 62%, #f4efe4 100%)",
        }}
      />
      {/* warm sun bloom upper-left */}
      <div
        className="absolute -left-[10%] -top-[12%] h-[60vh] w-[60vh] rounded-full"
        style={{
          background:
            "radial-gradient(circle, rgba(240,192,130,0.55) 0%, rgba(240,192,130,0.18) 40%, transparent 70%)",
          filter: "blur(8px)",
        }}
      />

      <svg
        className="absolute inset-0 h-full w-full"
        viewBox="0 0 100 100"
        preserveAspectRatio="xMidYMid slice"
      >
        {/* ── Breathing sun ── */}
        <g ref={sunRef} style={{ transformOrigin: "20% 18%" }}>
          <circle cx="20" cy="18" r="6.5" fill="#f0c082" opacity="0.9" />
          <circle cx="20" cy="18" r="9" fill="none" stroke="#f0c082" strokeWidth="0.25" opacity="0.5" />
          <circle cx="20" cy="18" r="12" fill="none" stroke="#f0c082" strokeWidth="0.18" opacity="0.3" />
        </g>

        {/* ── Distant mountains ── */}
        <path
          d="M 0 70 L 14 56 L 22 62 L 34 50 L 46 60 L 58 52 L 70 60 L 82 54 L 100 64 L 100 100 L 0 100 Z"
          fill="#b8c8c0"
          opacity="0.45"
        />

        {/* ── Hills (3 layers) ── */}
        <g ref={hillsRef}>
          <path
            d="M 0 78 C 18 70, 32 74, 50 72 C 68 70, 82 76, 100 74 L 100 100 L 0 100 Z"
            fill="#9cb89a"
            opacity="0.7"
          />
          <path
            d="M 0 84 C 22 78, 38 82, 56 80 C 72 78, 86 84, 100 82 L 100 100 L 0 100 Z"
            fill="#7aa078"
            opacity="0.85"
          />
          <path
            d="M 0 90 C 24 86, 42 88, 60 86 C 78 84, 90 90, 100 88 L 100 100 L 0 100 Z"
            fill="#5e8a62"
          />
        </g>

        {/* ── The tree on the right crest ── */}
        <g ref={treeRef} style={{ transformOrigin: "78% 78%" }}>
          {/* trunk */}
          <path
            d="M 78 78 C 78.4 70, 78 62, 79 58"
            stroke="#3a2a18"
            strokeWidth="0.6"
            fill="none"
            strokeLinecap="round"
          />
          {/* canopy — a few overlapping Bézier blobs, hand-drawn feel */}
          <g fill="#2e6a3e" opacity="0.92">
            <path d="M 79 58 C 75 52, 77 46, 82 45 C 86 44, 89 48, 88 53 C 87 56, 83 58, 79 58 Z" />
            <path d="M 79 56 C 82 50, 88 49, 91 53 C 93 57, 90 60, 85 60 C 82 60, 79 58, 79 56 Z" />
            <path d="M 78 56 C 76 52, 78 48, 82 48 C 84 48, 84 52, 82 55 C 81 57, 79 57, 78 56 Z" />
          </g>
          {/* canopy highlight */}
          <path
            d="M 80 50 C 82 48, 85 48, 86 50"
            stroke="#5e9a6a"
            strokeWidth="0.4"
            fill="none"
            opacity="0.7"
          />
        </g>

        {/* ── A few small distant trees ── */}
        <g fill="#3a6a48" opacity="0.7">
          <path d="M 22 78 l 0 -3" stroke="#3a2a18" strokeWidth="0.25" />
          <circle cx="22" cy="74" r="1.2" />
          <path d="M 40 80 l 0 -2.4" stroke="#3a2a18" strokeWidth="0.22" />
          <circle cx="40" cy="77" r="0.9" />
          <path d="M 64 78 l 0 -3" stroke="#3a2a18" strokeWidth="0.25" />
          <circle cx="64" cy="74.5" r="1" />
        </g>

        {/* ── Light trails (the synapse threads) ── */}
        <path
          ref={trail1GlowRef}
          className="synapse-thread-glow"
          d="M -2 78 C 20 54, 35 64, 50 60"
        />
        <path
          ref={trail2GlowRef}
          className="synapse-thread-glow"
          d="M 50 60 C 70 68, 82 52, 102 42"
        />
        <path
          ref={trail1Ref}
          className="synapse-thread"
          d="M -2 78 C 20 54, 35 64, 50 60"
          strokeWidth="0.5"
        />
        <path
          ref={trail2Ref}
          className="synapse-thread"
          d="M 50 60 C 70 68, 82 52, 102 42"
          strokeWidth="0.5"
        />

        {/* ── Trail nodes ── */}
        <g>
          <circle cx="50" cy="60" r="0.9" className="synapse-node" />
          <circle cx="50" cy="60" r="2" className="synapse-node-ring">
            <animate attributeName="r" values="1.4;3;1.4" dur="3.2s" repeatCount="indefinite" />
            <animate attributeName="opacity" values="0.6;0;0.6" dur="3.2s" repeatCount="indefinite" />
          </circle>
        </g>

        {/* ── Foreground grass tufts ── */}
        <g stroke="#3a6a48" strokeWidth="0.18" opacity="0.6" strokeLinecap="round">
          {Array.from({ length: 26 }).map((_, i) => {
            const x = 2 + i * 3.8;
            const h = 1 + ((Math.sin(i * 2.7) + 1) / 2) * 1.4;
            return (
              <path key={i} d={`M ${x} 95 q 0.4 -${h} 0.8 -${h + 0.4}`} fill="none" />
            );
          })}
        </g>
      </svg>

      {/* ── Floating pollen (CSS-animated motes) ── */}
      {pollen.map((p) => (
        <span
          key={p.id}
          className="absolute rounded-full"
          style={{
            left: `${p.left}%`,
            top: `${p.top}%`,
            width: p.size,
            height: p.size,
            background: "rgba(201,123,46,0.8)",
            boxShadow: "0 0 6px rgba(201,123,46,0.6)",
            // @ts-expect-error custom props
            "--drift-x": `${p.dx}px`,
            "--drift-y": `${p.dy}px`,
            animation: `pollen-float ${p.dur}s ease-in-out ${p.delay}s infinite`,
          }}
        />
      ))}

      {/* paper grain over the scene so it never looks like a flat gradient */}
      <div className="paper-grain absolute inset-0" />
    </div>
  );
}
