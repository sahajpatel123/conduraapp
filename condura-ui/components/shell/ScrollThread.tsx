"use client";

import { useEffect, useRef, useState } from "react";
import { motion, useScroll, useSpring, useTransform } from "motion/react";

/**
 * ScrollThread — a refined vertical synapse thread on the left margin.
 *
 * Mature and subtle: a 1px track, a drawn synapse that fills as you
 * scroll, a single pollen node riding the current position, and a
 * small tick for each [data-section] on the page that lights when its
 * section enters the viewport. Hovering a tick jumps to that section.
 *
 * Hidden on narrow viewports and under reduced motion.
 */
type SectionTick = { id: string; top: number };

export default function ScrollThread() {
  const ref = useRef<HTMLDivElement | null>(null);
  const { scrollYProgress } = useScroll();
  const draw = useSpring(scrollYProgress, { stiffness: 120, damping: 30, mass: 0.4 });
  const nodeY = useTransform(draw, [0, 1], ["0%", "100%"]);
  const [ticks, setTicks] = useState<SectionTick[]>([]);
  const [activeId, setActiveId] = useState<string | null>(null);

  // Discover sections + their vertical position on the page.
  useEffect(() => {
    const measure = () => {
      const sections = Array.from(document.querySelectorAll<HTMLElement>("[data-section]"));
      const h = document.documentElement.scrollHeight - window.innerHeight;
      const found: SectionTick[] = sections
        .map((s) => {
          const top = s.offsetTop;
          const pct = h > 0 ? top / h : 0;
          return { id: s.id || s.dataset.section || "", top: Math.max(0, Math.min(1, pct)) };
        })
        .filter((t) => t.id);
      setTicks(found);
    };
    measure();
    window.addEventListener("resize", measure);
    // Re-measure after fonts/images settle
    const t = setTimeout(measure, 800);
    return () => {
      window.removeEventListener("resize", measure);
      clearTimeout(t);
    };
  }, []);

  // Track which section is currently in view.
  useEffect(() => {
    if (ticks.length === 0) return;
    const observer = new IntersectionObserver(
      (entries) => {
        for (const e of entries) {
          if (e.isIntersecting && e.target instanceof HTMLElement) {
            setActiveId(e.target.id || e.target.dataset.section || null);
          }
        }
      },
      { rootMargin: "-40% 0px -55% 0px" }
    );
    ticks.forEach((t) => {
      const el = document.getElementById(t.id) || document.querySelector(`[data-section="${t.id}"]`);
      if (el) observer.observe(el);
    });
    return () => observer.disconnect();
  }, [ticks]);

  useEffect(() => {
    const prefersReduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (prefersReduced && ref.current) ref.current.style.display = "none";
  }, []);

  const jumpTo = (id: string) => {
    const el = document.getElementById(id) || document.querySelector(`[data-section="${id}"]`);
    if (el instanceof HTMLElement) {
      el.scrollIntoView({ behavior: "smooth", block: "start" });
    }
  };

  return (
    <div
      ref={ref}
      aria-hidden
      className="pointer-events-none fixed left-5 top-0 z-30 hidden h-screen w-8 lg:block"
    >
      <svg viewBox="0 0 8 1000" className="h-full w-full" preserveAspectRatio="none">
        <line x1="4" y1="0" x2="4" y2="1000" stroke="rgba(20,17,11,0.08)" strokeWidth="1" />
        <motion.line
          x1="4" y1="0" x2="4" y2="1000"
          className="synapse-thread"
          strokeWidth="1.25"
          style={{ pathLength: draw, opacity: 0.85 }}
        />
      </svg>

      {/* section ticks */}
      {ticks.map((t) => (
        <button
          key={t.id}
          type="button"
          onClick={() => jumpTo(t.id)}
          className="pointer-events-auto group absolute left-1/2 -translate-x-1/2"
          style={{ top: `${t.top * 100}%`, transform: `translate(-50%, -50%)` }}
          aria-label={`Jump to section ${t.id}`}
        >
          <span
            className={`block h-1.5 w-1.5 rounded-full transition-all duration-300 ${
              activeId === t.id
                ? "scale-150 bg-[var(--color-pollen)] shadow-[0_0_8px_rgba(201,123,46,0.8)]"
                : "bg-[var(--color-ink-faint)] opacity-50 group-hover:opacity-100 group-hover:bg-[var(--color-synapse)]"
            }`}
          />
        </button>
      ))}

      {/* riding pollen node */}
      <motion.div
        className="absolute left-1/2 h-2 w-2 -translate-x-1/2 rounded-full bg-[var(--color-pollen)]"
        style={{ top: nodeY, y: "-50%", boxShadow: "0 0 10px rgba(201,123,46,0.7)" }}
      />
    </div>
  );
}
