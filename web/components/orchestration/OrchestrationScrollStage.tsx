"use client";

import { useEffect, useRef, useState } from "react";
import {
  motion,
  useMotionValueEvent,
  useScroll,
  useSpring,
  useTransform,
  type MotionValue,
} from "motion/react";
import { useReducedMotion } from "@/hooks/useReducedMotion";

const PHASES = [
  {
    id: "decomposition",
    label: "Phase 1",
    title: "Decomposition",
    body: "A single complex prompt is torn into discrete sub-tasks. The Strategist plans the assault.",
  },
  {
    id: "fan-out",
    label: "Phase 2",
    title: "Parallel fan-out · v0.2.0",
    body: "Condura will spawn lightweight sub-agents for each task and run them concurrently. Single sub-agent spawns work today; orchestrated fan-out is on the roadmap.",
  },
  {
    id: "resolution",
    label: "Phase 3",
    title: "Deterministic resolution",
    body: "Results are collated, verified by strict logic (not a hallucinating LLM), and committed to your workspace.",
  },
] as const;

const SPAWN_LINES = [
  "Spawning react-agent…",
  "Spawning rust-agent…",
  "Mounting DOM analyzer…",
  "Starting headless browser…",
];

/** Scroll-scrubbed sticky stage — one phase at a time while the page scroll is “held” in the tall runway. */
export default function OrchestrationScrollStage() {
  const reduced = useReducedMotion();
  const containerRef = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ["start start", "end end"],
  });
  const smooth = useSpring(scrollYProgress, { damping: 28, stiffness: 120, mass: 0.35 });

  const phase1Opacity = useTransform(smooth, [0, 0.02, 0.28, 0.34], [1, 1, 1, 0]);
  const phase2Opacity = useTransform(smooth, [0.28, 0.34, 0.61, 0.67], [0, 1, 1, 0]);
  const phase3Opacity = useTransform(smooth, [0.61, 0.67, 0.98, 1], [0, 1, 1, 1]);

  const phase1Y = useTransform(smooth, [0, 0.34], [12, 0]);
  const phase2Y = useTransform(smooth, [0.28, 0.67], [16, 0]);
  const phase3Y = useTransform(smooth, [0.61, 1], [16, 0]);

  const phase2Inner = useTransform(smooth, [0.34, 0.67], [0, 1]);
  const phase1Tasks = useTransform(smooth, [0.04, 0.28], [0, 4]);

  const [activePhase, setActivePhase] = useState(0);
  const [spawnVisible, setSpawnVisible] = useState(0);
  const [tasksVisible, setTasksVisible] = useState(0);

  useMotionValueEvent(smooth, "change", (v) => {
    if (v < 0.34) setActivePhase(0);
    else if (v < 0.67) setActivePhase(1);
    else setActivePhase(2);
  });

  useMotionValueEvent(phase2Inner, "change", (v) => {
    setSpawnVisible(Math.min(4, Math.ceil(v * 4)));
  });

  useMotionValueEvent(phase1Tasks, "change", (v) => {
    setTasksVisible(Math.min(4, Math.ceil(v * 4)));
  });

  if (reduced) {
    return (
      <div className="mt-12 space-y-8">
        {PHASES.map((phase) => (
          <div
            key={phase.id}
            className="rounded-3xl border border-[rgba(20,17,11,0.14)] bg-[var(--color-paper-warm)] p-8"
          >
            <p className="text-mono-label mb-2">{phase.label}</p>
            <p className="font-display text-[26px] text-[var(--color-ink)]">{phase.title}</p>
            <p className="mt-2 max-w-prose text-[14px] text-[var(--color-ink-mute)]">{phase.body}</p>
          </div>
        ))}
      </div>
    );
  }

  return (
    <div ref={containerRef} className="relative mt-12 h-[360vh]">
      <div className="sticky top-28 flex h-[min(78vh,720px)] items-stretch justify-center overflow-hidden rounded-3xl border border-[rgba(20,17,11,0.14)] bg-[var(--color-paper-warm)] shadow-[var(--shadow-card)]">
        <div className="paper-grain absolute inset-0" aria-hidden />

        <ProgressRail progress={smooth} activePhase={activePhase} />

        <div className="relative z-10 flex w-full items-center justify-center p-6 sm:p-10">
          <motion.div
            style={{ opacity: phase1Opacity, y: phase1Y }}
            className={`absolute inset-0 flex flex-col items-center justify-center p-6 sm:p-8 ${activePhase === 0 ? "pointer-events-auto" : "pointer-events-none"}`}
            aria-hidden={activePhase !== 0}
          >
            <PhaseCard phase={PHASES[0]} />
            <div className="mt-8 flex gap-3 sm:gap-4">
              {[0, 1, 2, 3].map((i) => (
                <motion.div
                  key={i}
                  animate={{
                    opacity: i < tasksVisible ? 1 : 0.2,
                    scale: i < tasksVisible ? 1 : 0.92,
                    y: i < tasksVisible ? 0 : 8,
                  }}
                  transition={{ duration: 0.35, ease: [0.22, 1, 0.36, 1] }}
                  className="grid h-14 w-14 place-items-center rounded-2xl border border-[rgba(20,17,11,0.18)] bg-[var(--color-paper)] shadow-[var(--shadow-paper)] sm:h-16 sm:w-16"
                >
                  <span className="font-mono text-[11px] text-[var(--color-ink-mute)]">T-{i + 1}</span>
                </motion.div>
              ))}
            </div>
          </motion.div>

          <motion.div
            style={{ opacity: phase2Opacity, y: phase2Y }}
            className={`absolute inset-0 flex flex-col items-center justify-center p-6 sm:p-8 ${activePhase === 1 ? "pointer-events-auto" : "pointer-events-none"}`}
            aria-hidden={activePhase !== 1}
          >
            <PhaseCard phase={PHASES[1]} />
            <div className="relative mt-8 w-full max-w-lg overflow-hidden rounded-3xl border border-[rgba(20,17,11,0.14)] bg-[var(--color-ink)] p-5 sm:p-6">
              <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
                <div className="h-16 w-16 animate-pulse rounded-full bg-[var(--color-synapse-glow)] opacity-40 blur-2xl" />
              </div>
              <div className="relative z-10 flex max-h-[180px] flex-col gap-3 overflow-y-auto pr-1">
                {SPAWN_LINES.map((line, i) => (
                  <motion.div
                    key={line}
                    animate={{
                      opacity: i < spawnVisible ? 1 : 0,
                      x: i < spawnVisible ? 0 : -12,
                    }}
                    transition={{ duration: 0.28 }}
                    className="flex items-center gap-3 font-mono text-[12px] text-[rgba(244,239,228,0.72)]"
                  >
                    <span className="h-2 w-2 shrink-0 rounded-full bg-[var(--color-synapse-light)]" />
                    {line}
                  </motion.div>
                ))}
              </div>
            </div>
          </motion.div>

          <motion.div
            style={{ opacity: phase3Opacity, y: phase3Y }}
            className={`absolute inset-0 flex flex-col items-center justify-center p-6 text-center sm:p-8 ${activePhase === 2 ? "pointer-events-auto" : "pointer-events-none"}`}
            aria-hidden={activePhase !== 2}
          >
            <PhaseCard phase={PHASES[2]} />
            <motion.div
              className="mt-8 grid h-28 w-28 place-items-center rounded-[2rem] border-2 border-[var(--color-synapse)] bg-[rgba(11,61,46,0.08)] sm:h-32 sm:w-32"
              animate={{ rotate: 360 }}
              transition={{ duration: 12, repeat: Infinity, ease: "linear" }}
            >
              <div className="h-14 w-14 rounded-full border border-[var(--color-synapse)] bg-[var(--color-paper)] sm:h-16 sm:w-16" />
            </motion.div>
            <p className="mt-6 font-mono text-[12px] text-[var(--color-ink-mute)]">
              Diffs merged. AST verified. Lockfile updated.
            </p>
          </motion.div>
        </div>

        <ScrollHint progress={smooth} />
      </div>
    </div>
  );
}

function PhaseCard({ phase }: { phase: (typeof PHASES)[number] }) {
  return (
    <div className="max-w-md rounded-2xl border border-[rgba(20,17,11,0.12)] bg-[var(--color-paper)] p-6 text-center">
      <p className="text-mono-label mb-2">{phase.label}</p>
      <p className="font-display text-[26px] text-[var(--color-ink)]">{phase.title}</p>
      <p className="mt-2 text-[14px] leading-relaxed text-[var(--color-ink-mute)]">{phase.body}</p>
    </div>
  );
}

function ProgressRail({
  progress,
  activePhase,
}: {
  progress: MotionValue<number>;
  activePhase: number;
}) {
  const fill = useTransform(progress, [0, 1], ["0%", "100%"]);

  return (
    <div className="absolute bottom-8 left-6 top-8 z-20 hidden w-8 sm:block" aria-hidden>
      <div className="relative mx-auto h-full w-px bg-[rgba(20,17,11,0.1)]">
        <motion.div
          className="absolute left-0 top-0 w-px origin-top bg-[var(--color-synapse)]"
          style={{ height: fill }}
        />
        {[0, 1, 2].map((i) => (
          <span
            key={i}
            className={`absolute left-1/2 h-2 w-2 -translate-x-1/2 rounded-full border transition-colors duration-300 ${
              activePhase === i
                ? "border-[var(--color-pollen)] bg-[var(--color-pollen)] shadow-[0_0_0_4px_rgba(201,123,46,0.15)]"
                : activePhase > i
                  ? "border-[var(--color-synapse)] bg-[var(--color-synapse)]"
                  : "border-[rgba(20,17,11,0.2)] bg-[var(--color-paper)]"
            }`}
            style={{ top: `${12 + i * 38}%` }}
          />
        ))}
      </div>
    </div>
  );
}

function ScrollHint({ progress }: { progress: MotionValue<number> }) {
  const opacity = useTransform(progress, [0, 0.05, 0.92, 1], [1, 1, 1, 0]);

  return (
    <motion.p
      style={{ opacity }}
      className="absolute bottom-5 right-6 z-20 font-mono text-[10px] uppercase tracking-[0.16em] text-[var(--color-ink-faint)]"
    >
      Scroll to advance
    </motion.p>
  );
}
