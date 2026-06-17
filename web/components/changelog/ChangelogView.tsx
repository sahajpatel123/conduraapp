"use client";

import { motion } from "motion/react";
import { useEffect, useRef, useState } from "react";
import AnimatedBadge from "@/components/motion/AnimatedBadge";
import { EASE_OUT } from "@/lib/motion";

/* ────────────────────────────────────────────────────────────
   CHANGELOG — What Shipped
   Release notes rendered as a living timeline. Each h2
   becomes a version node on a vertical spine; entries
   fade in as you scroll.
   ──────────────────────────────────────────────────────────── */

interface ChangelogViewProps {
  html: string | null;
}

export default function ChangelogView({ html }: ChangelogViewProps) {
  const [mounted, setMounted] = useState(false);
  useEffect(() => { const t = setTimeout(() => setMounted(true), 100); return () => clearTimeout(t); }, []);

  if (!html) {
    return (
      <div className="mx-auto max-w-2xl py-24 text-center">
        <p className="font-body-mature text-white/45">
          The changelog is not available right now. Check{" "}
          <a className="underline hover:text-white" href="https://github.com/sahajpatel123/conduraapp/releases">
            GitHub releases
          </a>{" "}
          for the latest changes.
        </p>
      </div>
    );
  }

  // Split the rendered HTML by <h2> tags so we can build a timeline.
  const sections = html.split(/<(h2)[^>]*>/).filter(Boolean);

  return (
    <main className="relative w-full bg-black text-white overflow-hidden">
      {/* Hero */}
      <section className="relative min-h-[70vh] flex flex-col items-center justify-center px-6 overflow-hidden">
        <div className="absolute inset-0 bg-grid-dark opacity-20" />
        <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
          <motion.div
            animate={{ rotate: 360 }}
            transition={{ duration: 60, repeat: Infinity, ease: "linear" }}
            className="w-[500px] h-[500px] rounded-full border border-white/[0.06]"
          />
          <motion.div
            animate={{ rotate: -360 }}
            transition={{ duration: 40, repeat: Infinity, ease: "linear" }}
            className="absolute w-[350px] h-[350px] rounded-full border border-dashed border-white/[0.08]"
          />
        </div>

        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: mounted ? 1 : 0, y: mounted ? 0 : 20 }}
          transition={{ duration: 1, ease: EASE_OUT }}
          className="relative z-10 max-w-3xl text-center"
        >
          <div className="mb-8 flex justify-center">
            <AnimatedBadge tone="neutral" pulse>Changelog</AnimatedBadge>
          </div>
          <h1 className="font-display text-[clamp(2.5rem,6vw,4.5rem)] font-semibold leading-[1.05] tracking-[-0.04em]">
            What shipped.
          </h1>
          <p className="mt-8 mx-auto max-w-xl font-lead-airy">
            Every notable change to Condura, release by release. Pulled live from the repository
            changelog. No marketing — just the diff.
          </p>
        </motion.div>
      </section>

      {/* Timeline */}
      <section className="relative w-full py-[120px] px-6 border-t border-white/[0.08]">
        <div className="mx-auto max-w-3xl">
          {/* Vertical spine */}
          <div className="absolute left-6 md:left-1/2 top-[120px] bottom-0 w-[1px] bg-gradient-to-b from-white/20 via-white/10 to-transparent md:-translate-x-1/2" />

          <div className="space-y-24">
            {sections.map((section, i) => {
              // The first chunk before any h2 is the intro text; re-insert h2 tag for parsing
              const full = i === 0 ? section : `<h2>${section}`;
              return (
                <TimelineEntry key={i} html={full} index={i} />
              );
            })}
          </div>
        </div>
      </section>
    </main>
  );
}

function TimelineEntry({ html, index }: { html: string; index: number }) {
  const ref = useRef<HTMLDivElement>(null);
  const [visible, setVisible] = useState(false);

  useEffect(() => {
    const obs = new IntersectionObserver(
      ([entry]) => entry.isIntersecting && setVisible(true),
      { threshold: 0.15 }
    );
    if (ref.current) obs.observe(ref.current);
    return () => obs.disconnect();
  }, []);

  const isLeft = index % 2 === 0;

  return (
    <motion.div
      ref={ref}
      initial={{ opacity: 0, y: 40 }}
      animate={visible ? { opacity: 1, y: 0 } : {}}
      transition={{ duration: 0.7, ease: EASE_OUT }}
      className={`relative flex ${isLeft ? "md:justify-start" : "md:justify-end"}`}
    >
      {/* Node dot on the spine */}
      <div className="absolute left-6 md:left-1/2 top-2 -translate-x-1/2 z-10">
        <motion.div
          animate={{ scale: visible ? [1, 1.3, 1] : 1 }}
          transition={{ duration: 0.6 }}
          className="flex h-3 w-3 items-center justify-center rounded-full bg-white shadow-[0_0_12px_rgba(255,255,255,0.4)]"
        />
        <motion.div
          animate={{ scale: visible ? [1, 2] : 1, opacity: visible ? [0.4, 0] : 0 }}
          transition={{ duration: 1.5, repeat: Infinity, delay: 0.5 }}
          className="absolute inset-0 rounded-full border border-white/30"
        />
      </div>

      {/* Content card */}
      <div className={`ml-12 md:ml-0 md:w-[46%] ${isLeft ? "" : "md:ml-auto"}`}>
        <div className="mature-panel rounded-2xl p-8">
          <article
            className="prose-condura"
            dangerouslySetInnerHTML={{ __html: html }}
          />
        </div>
      </div>
    </motion.div>
  );
}
