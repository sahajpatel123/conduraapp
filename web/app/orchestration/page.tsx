"use client";

import { motion, AnimatePresence, useScroll, useTransform, useSpring } from "motion/react";
import { useRef, useState, useEffect } from "react";
import AnimatedBadge from "@/components/motion/AnimatedBadge";
import TiltCard from "@/components/motion/TiltCard";
import { Icon, type IconKey } from "@/components/motion/Icon";
import { EASE_OUT } from "@/lib/motion";

/* ────────────────────────────────────────────────────────────
   ORCHESTRATION — The Engine Story
   A conductor, not a wrapper. This page walks through how
   Condura decomposes a task into a DAG of parallel waves,
   coordinates agents through a SQLite event bus, and resolves
   deterministically.
   ──────────────────────────────────────────────────────────── */

const TERMINAL_STEPS = [
  { input: "condura run \"refactor auth module + add tests\"", output: "Decomposing task into DAG… 4 nodes, 2 waves." },
  { input: "", output: "Wave 1 → 3 parallel agents spawned." },
  { input: "", output: "  ◇ claude-code: analyzing auth.go (340ms)" },
  { input: "", output: "  ◇ codex: writing auth_test.go (1.2s)" },
  { input: "", output: "  ◇ ollama: reviewing imports (90ms)" },
  { input: "", output: "Wave 1 resolved. 0 conflicts. Lockfile +3." },
  { input: "", output: "Wave 2 → 1 agent spawned." },
  { input: "", output: "  ◇ claude-code: updating README (280ms)" },
  { input: "", output: "Wave 2 resolved. Task complete. 6.4s total." },
];

const DAG_NODES = [
  { id: "root", label: "Refactor auth", x: 50, y: 10, wave: 0 },
  { id: "analyze", label: "Analyze AST", x: 20, y: 45, wave: 1 },
  { id: "write", label: "Write tests", x: 50, y: 45, wave: 1 },
  { id: "review", label: "Review imports", x: 80, y: 45, wave: 1 },
  { id: "docs", label: "Update docs", x: 50, y: 80, wave: 2 },
];

const DAG_EDGES = [
  { from: "root", to: "analyze" },
  { from: "root", to: "write" },
  { from: "root", to: "review" },
  { from: "analyze", to: "docs" },
  { from: "write", to: "docs" },
  { from: "review", to: "docs" },
];

const AGENTS = [
  { name: "Claude Code", role: "AST analysis + patch", color: "#d97757", latency: "340ms" },
  { name: "Codex", role: "Test generation", color: "#10a37f", latency: "1.2s" },
  { name: "Ollama", role: "Local import review", color: "#6b7280", latency: "90ms" },
  { name: "Antigravity", role: "Refactor core", color: "#8b5cf6", latency: "running" },
];

const BUS_EVENTS = [
  { ts: "00:00.00", src: "planner", msg: "task.decompose → 4 nodes", type: "plan" },
  { ts: "00:00.12", src: "scheduler", msg: "wave.1.start → 3 agents", type: "wave" },
  { ts: "00:00.34", src: "claude-code", msg: "patch.applied → auth.go +12 -4", type: "patch" },
  { ts: "00:00.46", src: "codex", msg: "file.created → auth_test.go", type: "patch" },
  { ts: "00:01.20", src: "codex", msg: "test.complete → 12 pass, 0 fail", type: "test" },
  { ts: "00:00.09", src: "ollama", msg: "review.done → 0 banned imports", type: "review" },
  { ts: "00:01.34", src: "scheduler", msg: "wave.1.resolve → 0 conflicts", type: "wave" },
  { ts: "00:01.35", src: "scheduler", msg: "wave.2.start → 1 agent", type: "wave" },
  { ts: "00:01.63", src: "claude-code", msg: "patch.applied → README.md +8", type: "patch" },
  { ts: "00:01.64", src: "scheduler", msg: "wave.2.resolve → task.complete", type: "wave" },
];

const EVENT_COLORS: Record<string, string> = {
  plan: "#a1a1aa",
  wave: "#ffffff",
  patch: "#10a37f",
  test: "#3b82f6",
  review: "#8b5cf6",
};

export default function OrchestrationPage() {
  return (
    <main className="relative w-full bg-black text-white overflow-hidden">
      <OrchestrationHero />
      <DecompositionSection />
      <FanOutSection />
      <EventBusSection />
      <WavesTimeline />
      <LiveTerminalDemo />
      <ClosingCTA />
    </main>
  );
}

/* ════════════════════════════════════════════════════════════
   1. HERO — The Conductor
   ════════════════════════════════════════════════════════════ */

function OrchestrationHero() {
  const [mounted, setMounted] = useState(false);
  useEffect(() => { const t = setTimeout(() => setMounted(true), 200); return () => clearTimeout(t); }, []);

  return (
    <section className="relative min-h-screen flex flex-col items-center justify-center px-6 overflow-hidden">
      {/* Ambient grid */}
      <div className="absolute inset-0 bg-grid-dark opacity-20" />
      {/* Radial glow */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: mounted ? 0.15 : 0 }}
        transition={{ duration: 2 }}
        className="absolute inset-0 bg-[radial-gradient(circle_at_50%_40%,rgba(255,255,255,0.12),transparent_60%)]"
      />

      {/* Orbiting rings — the "conductor" metaphor */}
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 40, repeat: Infinity, ease: "linear" }}
          className="w-[600px] h-[600px] rounded-full border border-white/[0.06]"
        >
          {[0, 90, 180, 270].map((deg) => (
            <div
              key={deg}
              style={{ transform: `rotate(${deg}deg)` }}
              className="absolute top-0 left-1/2 w-2 h-2 -translate-x-1/2 rounded-full bg-white/20 shadow-[0_0_12px_rgba(255,255,255,0.3)]"
            />
          ))}
        </motion.div>
        <motion.div
          animate={{ rotate: -360 }}
          transition={{ duration: 25, repeat: Infinity, ease: "linear" }}
          className="absolute w-[400px] h-[400px] rounded-full border border-dashed border-white/[0.08]"
        />
        <div className="absolute w-[200px] h-[200px] rounded-full bg-white/[0.02] blur-3xl" />
      </div>

      <div className="relative z-10 max-w-4xl text-center">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: mounted ? 1 : 0, y: mounted ? 0 : 20 }}
          transition={{ duration: 1, ease: EASE_OUT }}
        >
          <div className="mb-8 flex justify-center">
            <AnimatedBadge tone="neutral" pulse>Orchestration Engine</AnimatedBadge>
          </div>

          <h1 className="font-display text-[clamp(2.5rem,7vw,5rem)] font-semibold leading-[1.05] tracking-[-0.04em]">
            A conductor.
            <br />
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-white via-white to-white/30">
              Not a wrapper.
            </span>
          </h1>

          <p className="mt-8 mx-auto max-w-2xl font-lead-airy">
            Condura doesn&apos;t just call agents one at a time. It decomposes your task into a directed
            acyclic graph, fans out parallel waves of sub-agents, coordinates them through a
            microsecond SQLite event bus, and resolves every conflict deterministically.
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: mounted ? 1 : 0 }}
          transition={{ delay: 0.8, duration: 1 }}
          className="mt-12 flex flex-wrap items-center justify-center gap-6"
        >
          {[
            { label: "Max parallel agents", value: "5" },
            { label: "Bus latency", value: "<1ms" },
            { label: "Conflict resolution", value: "deterministic" },
          ].map((stat) => (
            <div key={stat.label} className="flex flex-col items-center">
              <span className="font-mono text-[22px] font-medium text-white">{stat.value}</span>
              <span className="mt-1 font-mono text-[10px] uppercase tracking-widest text-white/30">
                {stat.label}
              </span>
            </div>
          ))}
        </motion.div>
      </div>

      {/* Scroll indicator */}
      <motion.div
        initial={{ opacity: 0 }}
        animate={{ opacity: mounted ? 1 : 0 }}
        transition={{ delay: 1.5, duration: 1 }}
        className="absolute bottom-10 left-1/2 -translate-x-1/2 flex flex-col items-center gap-2"
      >
        <span className="font-mono text-[10px] uppercase tracking-widest text-white/25">Scroll</span>
        <div className="w-[1px] h-10 bg-gradient-to-b from-white/25 to-transparent" />
      </motion.div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   2. DECOMPOSITION — Task → DAG
   ════════════════════════════════════════════════════════════ */

function DecompositionSection() {
  const ref = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({ target: ref, offset: ["start end", "end start"] });
  const drawProgress = useTransform(scrollYProgress, [0.1, 0.5], [0, 1]);
  useSpring(drawProgress, { damping: 30, stiffness: 80 });

  return (
    <section ref={ref} className="relative w-full py-[160px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        {/* Section header */}
        <div className="mb-20 max-w-3xl">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Phase 01</span>
          <h2 className="mt-4 font-display text-[clamp(2rem,5vw,3.5rem)] font-semibold tracking-[-0.03em] leading-[1.1]">
            Every task becomes a graph.
          </h2>
          <p className="mt-6 font-lead-airy">
            You give Condura a sentence. It returns a directed acyclic graph of atomic operations.
            Each node is a unit of work assigned to exactly one agent. Each edge is a dependency
            the scheduler will not violate.
          </p>
        </div>

        {/* DAG Visualization */}
        <div className="mature-panel relative aspect-[16/10] w-full overflow-hidden rounded-3xl p-8">
          <div className="absolute inset-0 bg-grid-dark opacity-15" />

          <svg className="absolute inset-0 w-full h-full" viewBox="0 0 100 100" preserveAspectRatio="none">
            {/* Edges */}
            {DAG_EDGES.map((edge, i) => {
              const from = DAG_NODES.find((n) => n.id === edge.from)!;
              const to = DAG_NODES.find((n) => n.id === edge.to)!;
              return (
                <motion.line
                  key={i}
                  x1={from.x}
                  y1={from.y}
                  x2={to.x}
                  y2={to.y}
                  stroke="rgba(255,255,255,0.25)"
                  strokeWidth="0.3"
                  initial={{ pathLength: 0, opacity: 0 }}
                  whileInView={{ pathLength: 1, opacity: 1 }}
                  viewport={{ once: true }}
                  transition={{ delay: i * 0.2, duration: 0.8 }}
                />
              );
            })}
          </svg>

          {/* Nodes */}
          {DAG_NODES.map((node, i) => (
            <motion.div
              key={node.id}
              initial={{ opacity: 0, scale: 0.6 }}
              whileInView={{ opacity: 1, scale: 1 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.15, type: "spring", stiffness: 200, damping: 20 }}
              style={{
                left: `${node.x}%`,
                top: `${node.y}%`,
                transform: "translate(-50%, -50%)",
              }}
              className="absolute z-10"
            >
              <div className={`flex flex-col items-center gap-2 ${node.wave === 0 ? "scale-110" : ""}`}>
                <div
                  className={`flex items-center justify-center rounded-2xl border backdrop-blur-md ${
                    node.wave === 0
                      ? "w-20 h-20 border-white/25 bg-white/[0.08] shadow-[0_0_30px_rgba(255,255,255,0.08)]"
                      : "w-16 h-16 border-white/15 bg-white/[0.04]"
                  }`}
                >
                  <span className="font-mono text-[10px] text-white/60">{node.label}</span>
                </div>
                {node.wave > 0 && (
                  <span className="font-mono text-[9px] uppercase tracking-wider text-white/20">
                    W{node.wave}
                  </span>
                )}
              </div>
            </motion.div>
          ))}

          {/* Wave labels */}
          <div className="absolute left-4 top-[10%] font-mono text-[9px] uppercase tracking-widest text-white/20">
            Root
          </div>
          <div className="absolute left-4 top-[45%] font-mono text-[9px] uppercase tracking-widest text-white/20">
            Wave 1 · parallel
          </div>
          <div className="absolute left-4 bottom-[8%] font-mono text-[9px] uppercase tracking-widest text-white/20">
            Wave 2 · sequential
          </div>
        </div>

        {/* Explanation grid */}
        <div className="mt-12 grid md:grid-cols-3 gap-6">
          {[
            { num: "01", title: "Parse intent", desc: "The planner model reads your natural-language task and identifies atomic operations." },
            { num: "02", title: "Build DAG", desc: "Operations are arranged into a dependency graph. Cycles are rejected at construction time." },
            { num: "03", title: "Assign agents", desc: "Each node is matched to the best available backend — local, cloud, or CLI — by the router." },
          ].map((step, i) => (
            <motion.div
              key={step.num}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.1 }}
              className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6"
            >
              <span className="font-mono text-[13px] text-white/40">{step.num}</span>
              <h3 className="mt-3 font-body-mature text-[16px] font-semibold text-white">{step.title}</h3>
              <p className="mt-2 font-body-mature text-[14px] text-white/45 leading-relaxed">{step.desc}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   3. FAN-OUT — Parallel Agent Swarm
   ════════════════════════════════════════════════════════════ */

function FanOutSection() {
  return (
    <section className="relative w-full py-[160px] px-6 border-t border-white/[0.08] overflow-hidden">
      {/* Ambient pulse */}
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <motion.div
          animate={{ scale: [1, 1.3, 1], opacity: [0.05, 0.12, 0.05] }}
          transition={{ duration: 4, repeat: Infinity, ease: "easeInOut" }}
          className="w-[500px] h-[500px] rounded-full bg-white blur-[120px]"
        />
      </div>

      <div className="relative z-10 mx-auto max-w-5xl">
        <div className="mb-20 max-w-3xl">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Phase 02</span>
          <h2 className="mt-4 font-display text-[clamp(2rem,5vw,3.5rem)] font-semibold tracking-[-0.03em] leading-[1.1]">
            Waves run in parallel.
          </h2>
          <p className="mt-6 font-lead-airy">
            Within a wave, every node fires simultaneously. Condura holds a per-backend semaphore
            (default 2, max 5) so your machine never thrashes. File coordination is enforced by
            SQLite row-level locks, not filesystem races.
          </p>
        </div>

        {/* Agent swarm grid */}
        <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-6">
          {AGENTS.map((agent, i) => {
            const isRunning = agent.latency === "running";
            return (
              <TiltCard key={agent.name} maxRotate={8} className="h-full">
                <motion.div
                  initial={{ opacity: 0, y: 30 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true }}
                  transition={{ delay: i * 0.1, type: "spring", stiffness: 100, damping: 15 }}
                  className="relative h-full overflow-hidden rounded-2xl border border-white/[0.10] bg-white/[0.02] p-6 backdrop-blur-md"
                >
                  {/* Agent color accent */}
                  <div
                    className="absolute top-0 left-0 right-0 h-[2px]"
                    style={{ background: agent.color, opacity: 0.6 }}
                  />

                  <div className="flex items-start justify-between mb-6">
                    <div
                      className="flex h-10 w-10 items-center justify-center rounded-xl border border-white/15"
                      style={{ background: `${agent.color}15` }}
                    >
                      <span className="font-mono text-[13px]" style={{ color: agent.color }}>
                        {agent.name[0]}
                      </span>
                    </div>
                    {isRunning ? (
                      <span className="flex items-center gap-1.5 font-mono text-[10px] text-white/50">
                        <motion.span
                          animate={{ opacity: [1, 0.3, 1] }}
                          transition={{ duration: 1, repeat: Infinity }}
                          className="w-1.5 h-1.5 rounded-full bg-white/60"
                        />
                        running
                      </span>
                    ) : (
                      <span className="font-mono text-[10px] text-white/30">{agent.latency}</span>
                    )}
                  </div>

                  <h3 className="font-body-mature text-[15px] font-semibold text-white">{agent.name}</h3>
                  <p className="mt-1 font-body-mature text-[13px] text-white/40">{agent.role}</p>

                  {/* Progress bar for running agents */}
                  {isRunning && (
                    <div className="mt-4 h-[2px] w-full overflow-hidden rounded-full bg-white/10">
                      <motion.div
                        animate={{ x: ["-100%", "200%"] }}
                        transition={{ duration: 1.5, repeat: Infinity, ease: "easeInOut" }}
                        className="h-full w-1/2 rounded-full bg-white/40"
                      />
                    </div>
                  )}
                </motion.div>
              </TiltCard>
            );
          })}
        </div>

        {/* Concurrency controls */}
        <div className="mt-16 mature-panel rounded-2xl p-8">
          <div className="grid md:grid-cols-2 gap-8 items-center">
            <div>
              <h3 className="font-body-mature text-[18px] font-semibold text-white">
                Conservative by default.
              </h3>
              <p className="mt-3 font-body-mature text-[14px] text-white/45 leading-relaxed">
                Two parallel agents out of the box. Five if you push it. The semaphore is
                per-backend, so spawning three Ollama tasks won&apos;t starve your Claude Code slot.
                You can tune it in <code className="rounded bg-white/[0.06] px-1.5 py-0.5 font-mono text-[12px] text-white/60">~/.condura/config.yaml</code>.
              </p>
            </div>
            <div className="space-y-3">
              {[
                { label: "Default parallel", value: 2, max: 5 },
                { label: "Max parallel", value: 5, max: 5 },
                { label: "Per-backend semaphore", value: 2, max: 3 },
              ].map((ctrl) => (
                <div key={ctrl.label}>
                  <div className="flex items-center justify-between mb-1.5">
                    <span className="font-mono text-[11px] text-white/40">{ctrl.label}</span>
                    <span className="font-mono text-[11px] text-white/60">{ctrl.value} / {ctrl.max}</span>
                  </div>
                  <div className="h-[3px] w-full rounded-full bg-white/10">
                    <motion.div
                      initial={{ width: 0 }}
                      whileInView={{ width: `${(ctrl.value / ctrl.max) * 100}%` }}
                      viewport={{ once: true }}
                      transition={{ duration: 1, ease: EASE_OUT }}
                      className="h-full rounded-full bg-white/40"
                    />
                  </div>
                </div>
              ))}
            </div>
          </div>
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   4. EVENT BUS — SQLite Communication Layer
   ════════════════════════════════════════════════════════════ */

function EventBusSection() {
  const [visibleEvents, setVisibleEvents] = useState(0);

  useEffect(() => {
    const interval = setInterval(() => {
      setVisibleEvents((prev) => {
        if (prev >= BUS_EVENTS.length) return 0;
        return prev + 1;
      });
    }, 800);
    return () => clearInterval(interval);
  }, []);

  return (
    <section className="relative w-full py-[160px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-20 max-w-3xl">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Phase 03</span>
          <h2 className="mt-4 font-display text-[clamp(2rem,5vw,3.5rem)] font-semibold tracking-[-0.03em] leading-[1.1]">
            One bus. Zero races.
          </h2>
          <p className="mt-6 font-lead-airy">
            Agents don&apos;t talk to each other directly. They publish events to a local SQLite
            bus. The scheduler subscribes, resolves conflicts, and writes the merged state. No
            filesystem locks, no race conditions, no lost messages.
          </p>
        </div>

        {/* Live event bus terminal */}
        <div className="mature-panel overflow-hidden rounded-2xl">
          {/* Title bar */}
          <div className="flex items-center justify-between border-b border-white/[0.06] bg-white/[0.02] px-5 py-3">
            <div className="flex items-center gap-2">
              <div className="w-2.5 h-2.5 rounded-full bg-[#ff5f57]" />
              <div className="w-2.5 h-2.5 rounded-full bg-[#febc2e]" />
              <div className="w-2.5 h-2.5 rounded-full bg-[#28c840]" />
            </div>
            <span className="font-mono text-[11px] text-white/30">condura-bus — live event stream</span>
            <span className="flex items-center gap-1.5 font-mono text-[10px] text-white/30">
              <motion.span
                animate={{ opacity: [1, 0.3, 1] }}
                transition={{ duration: 1.5, repeat: Infinity }}
                className="w-1.5 h-1.5 rounded-full bg-green-400/60"
              />
              streaming
            </span>
          </div>

          {/* Event log */}
          <div className="bg-[#0a0a0a] p-6 h-[420px] overflow-hidden font-mono text-[12px]">
            <AnimatePresence>
              {BUS_EVENTS.slice(0, visibleEvents).map((event, i) => (
                <motion.div
                  key={`${event.ts}-${i}`}
                  initial={{ opacity: 0, x: -10 }}
                  animate={{ opacity: 1, x: 0 }}
                  className="flex items-start gap-4 py-1.5"
                >
                  <span className="text-white/25 shrink-0">{event.ts}</span>
                  <span
                    className="shrink-0 rounded px-1.5 py-0.5 text-[10px] uppercase"
                    style={{
                      color: EVENT_COLORS[event.type],
                      background: `${EVENT_COLORS[event.type]}15`,
                    }}
                  >
                    {event.src}
                  </span>
                  <span className="text-white/55">{event.msg}</span>
                </motion.div>
              ))}
            </AnimatePresence>

            {/* Blinking cursor */}
            <div className="mt-2 flex items-center gap-2">
              <span className="text-white/30">❯</span>
              <motion.span
                animate={{ opacity: [1, 0] }}
                transition={{ repeat: Infinity, duration: 0.9 }}
                className="inline-block w-[7px] h-[13px] bg-white/30"
              />
            </div>
          </div>
        </div>

        {/* Bus properties */}
        <div className="mt-12 grid md:grid-cols-3 gap-6">
          {([
            { icon: "bolt" as IconKey, title: "Sub-millisecond", desc: "SQLite WAL mode delivers events in under 1ms on local disk. No network hop." },
            { icon: "lock" as IconKey, title: "ACID guarantees", desc: "Every event is a transaction. Partial writes are rolled back. State is always consistent." },
            { icon: "list" as IconKey, title: "Replayable", desc: "The full event log is your audit trail. Replay any session, inspect any decision." },
          ]).map((prop, i) => (
            <motion.div
              key={prop.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.1 }}
              className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6"
            >
              <div className="flex h-10 w-10 items-center justify-center rounded-xl border border-white/10 bg-white/[0.03]">
                <Icon name={prop.icon} size={20} className="text-white/60" />
              </div>
              <h3 className="mt-3 font-body-mature text-[16px] font-semibold text-white">{prop.title}</h3>
              <p className="mt-2 font-body-mature text-[14px] text-white/45 leading-relaxed">{prop.desc}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   5. WAVES TIMELINE — Execution Flow
   ════════════════════════════════════════════════════════════ */

function WavesTimeline() {
  const waves = [
    {
      num: "01",
      title: "Wave 1 — Research",
      status: "resolved",
      agents: ["Ollama (local)", "Claude Code", "Codex"],
      duration: "1.3s",
      detail: "Three agents analyzed the codebase in parallel. AST parsed, imports reviewed, test stubs generated.",
    },
    {
      num: "02",
      title: "Wave 2 — Modify",
      status: "resolved",
      agents: ["Claude Code", "Antigravity"],
      duration: "2.1s",
      detail: "Patches applied to source files. SQLite file-locks prevented write conflicts. Diffs merged deterministically.",
    },
    {
      num: "03",
      title: "Wave 3 — Verify",
      status: "resolved",
      agents: ["Ollama (local)"],
      duration: "0.8s",
      detail: "Strict verification ran all deterministic safety rules. AST re-checked. Lockfile updated.",
    },
    {
      num: "04",
      title: "Wave 4 — Document",
      status: "resolved",
      agents: ["Claude Code"],
      duration: "0.6s",
      detail: "README updated with new API surface. Changelog entry drafted. Task marked complete.",
    },
  ];

  return (
    <section className="relative w-full py-[160px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-20 max-w-3xl">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Phase 04</span>
          <h2 className="mt-4 font-display text-[clamp(2rem,5vw,3.5rem)] font-semibold tracking-[-0.03em] leading-[1.1]">
            Waves resolve in order.
          </h2>
          <p className="mt-6 font-lead-airy">
            Later waves wait for earlier ones. Within a wave, everything fires at once. The
            scheduler tracks completion, detects stalls via heartbeat, and retries with
            fingerprinting to avoid infinite loops.
          </p>
        </div>

        {/* Timeline */}
        <div className="relative">
          {/* Vertical line */}
          <div className="absolute left-[20px] md:left-[28px] top-0 bottom-0 w-[1px] bg-gradient-to-b from-white/20 via-white/10 to-transparent" />

          <div className="space-y-8">
            {waves.map((wave, i) => (
              <motion.div
                key={wave.num}
                initial={{ opacity: 0, x: -20 }}
                whileInView={{ opacity: 1, x: 0 }}
                viewport={{ once: true, margin: "-50px" }}
                transition={{ delay: i * 0.1, duration: 0.6 }}
                className="relative flex gap-6 md:gap-8"
              >
                {/* Node dot */}
                <div className="relative z-10 flex h-10 w-10 md:h-14 md:w-14 shrink-0 items-center justify-center rounded-full border border-white/15 bg-black">
                  <span className="font-mono text-[11px] md:text-[13px] text-white/50">{wave.num}</span>
                  {/* Pulse ring */}
                  <motion.div
                    animate={{ scale: [1, 1.4], opacity: [0.4, 0] }}
                    transition={{ duration: 2, repeat: Infinity, delay: i * 0.3 }}
                    className="absolute inset-0 rounded-full border border-white/20"
                  />
                </div>

                {/* Content */}
                <div className="flex-1 mature-panel rounded-2xl p-6">
                  <div className="flex items-center justify-between mb-4">
                    <h3 className="font-body-mature text-[17px] font-semibold text-white">{wave.title}</h3>
                    <div className="flex items-center gap-3">
                      <span className="font-mono text-[11px] text-white/30">{wave.duration}</span>
                      <span className="flex items-center gap-1.5 rounded-full border border-green-400/20 bg-green-400/10 px-2.5 py-0.5 font-mono text-[10px] text-green-400/70">
                        <span className="w-1.5 h-1.5 rounded-full bg-green-400/60" />
                        {wave.status}
                      </span>
                    </div>
                  </div>
                  <p className="font-body-mature text-[14px] text-white/45 leading-relaxed mb-4">
                    {wave.detail}
                  </p>
                  <div className="flex flex-wrap gap-2">
                    {wave.agents.map((agent) => (
                      <span
                        key={agent}
                        className="rounded-full border border-white/[0.08] bg-white/[0.03] px-3 py-1 font-mono text-[11px] text-white/50"
                      >
                        {agent}
                      </span>
                    ))}
                  </div>
                </div>
              </motion.div>
            ))}
          </div>
        </div>

        {/* Total summary */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          className="mt-12 flex items-center justify-center gap-8 rounded-2xl border border-white/[0.08] bg-white/[0.02] py-6"
        >
          {[
            { label: "Total time", value: "4.8s" },
            { label: "Agents used", value: "4" },
            { label: "Waves", value: "4" },
            { label: "Conflicts", value: "0" },
          ].map((stat) => (
            <div key={stat.label} className="flex flex-col items-center">
              <span className="font-mono text-[24px] font-medium text-white">{stat.value}</span>
              <span className="mt-1 font-mono text-[10px] uppercase tracking-widest text-white/25">
                {stat.label}
              </span>
            </div>
          ))}
        </motion.div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   6. LIVE TERMINAL DEMO — Interactive Command Sequence
   ════════════════════════════════════════════════════════════ */

function LiveTerminalDemo() {
  const [visibleSteps, setVisibleSteps] = useState(0);

  useEffect(() => {
    if (visibleSteps >= TERMINAL_STEPS.length) {
      const reset = setTimeout(() => setVisibleSteps(0), 4000);
      return () => clearTimeout(reset);
    }
    const t = setTimeout(() => setVisibleSteps((p) => p + 1), visibleSteps === 0 ? 600 : 700);
    return () => clearTimeout(t);
  }, [visibleSteps]);

  return (
    <section className="relative w-full py-[160px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-20 max-w-3xl">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">Live Demo</span>
          <h2 className="mt-4 font-display text-[clamp(2rem,5vw,3.5rem)] font-semibold tracking-[-0.03em] leading-[1.1]">
            Watch it run.
          </h2>
          <p className="mt-6 font-lead-airy">
            A real task, decomposed and executed. This is the exact output format Condura
            produces in its terminal UI — streaming, colored, and timestamped.
          </p>
        </div>

        {/* Terminal window */}
        <div className="overflow-hidden rounded-2xl border border-white/[0.10] bg-[#0e0e0e] shadow-[0_40px_80px_rgba(0,0,0,0.5)]">
          {/* Title bar */}
          <div className="h-[40px] border-b border-white/[0.06] bg-[#1a1a1a] flex items-center px-4 relative">
            <div className="flex items-center gap-2">
              <div className="w-3 h-3 rounded-full bg-[#ff5f57]" />
              <div className="w-3 h-3 rounded-full bg-[#febc2e]" />
              <div className="w-3 h-3 rounded-full bg-[#28c840]" />
            </div>
            <span className="absolute left-1/2 -translate-x-1/2 font-mono text-[12px] text-white/25">
              condura-tui
            </span>
          </div>

          {/* Terminal body */}
          <div className="p-6 min-h-[400px] font-mono text-[13px] space-y-3">
            <AnimatePresence>
              {TERMINAL_STEPS.slice(0, visibleSteps).map((step, i) => (
                <motion.div
                  key={i}
                  initial={{ opacity: 0, y: 6 }}
                  animate={{ opacity: 1, y: 0 }}
                  transition={{ duration: 0.3 }}
                >
                  {step.input && (
                    <p className="text-white/80">
                      <span className="text-white/50 mr-2">❯</span>
                      {step.input}
                    </p>
                  )}
                  <p className={`mt-1 ${step.input ? "pl-5" : ""} text-white/35`}>
                    {step.output}
                  </p>
                </motion.div>
              ))}
            </AnimatePresence>

            {/* Blinking cursor */}
            {visibleSteps < TERMINAL_STEPS.length && (
              <div className="flex items-center gap-2 pt-2">
                <span className="text-white/40">❯</span>
                <motion.span
                  animate={{ opacity: [1, 0] }}
                  transition={{ repeat: Infinity, duration: 0.9 }}
                  className="inline-block w-[7px] h-[14px] bg-white/30"
                />
              </div>
            )}
          </div>

          {/* Status bar */}
          <div className="h-[32px] border-t border-white/[0.06] bg-[#151515] flex items-center justify-between px-5">
            <div className="flex items-center gap-4">
              <span className="flex items-center gap-1.5 font-mono text-[10px] text-white/25">
                <span className="w-1.5 h-1.5 rounded-full bg-white/40" />
                {visibleSteps >= 6 ? "wave 2" : visibleSteps >= 2 ? "wave 1" : "planning"}
              </span>
              <span className="font-mono text-[10px] text-white/15">sqlite: WAL</span>
            </div>
            <span className="font-mono text-[10px] text-white/15">gatekeeper: sealed</span>
          </div>
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   7. CLOSING CTA
   ════════════════════════════════════════════════════════════ */

function ClosingCTA() {
  return (
    <section className="relative w-full py-[200px] px-6 border-t border-white/[0.08] overflow-hidden">
      {/* Ambient glow */}
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <motion.div
          animate={{ scale: [1, 1.15, 1], opacity: [0.05, 0.1, 0.05] }}
          transition={{ duration: 5, repeat: Infinity }}
          className="w-[600px] h-[300px] rounded-full bg-white blur-[150px]"
        />
      </div>

      <div className="relative z-10 mx-auto max-w-3xl text-center">
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 1, ease: EASE_OUT }}
        >
          <h2 className="font-display text-[clamp(2rem,6vw,4rem)] font-semibold tracking-[-0.04em] leading-[1.05]">
            Stop pasting code into a browser tab.
          </h2>
          <p className="mt-8 font-lead-airy mx-auto max-w-xl">
            Download Condura, set your hotkey, and let the conductor orchestrate every AI tool
            you already have installed.
          </p>
          <div className="mt-12 flex flex-col sm:flex-row items-center justify-center gap-4">
            <a
              href="/download"
              className="mature-button inline-flex items-center gap-2 px-8 py-4 font-body-mature text-[15px] font-semibold"
            >
              Download v0.1.0
              <span aria-hidden>→</span>
            </a>
            <a
              href="/security"
              className="mature-button-secondary inline-flex items-center gap-2 px-6 py-4 font-body-mature text-[14px]"
            >
              How it stays safe
            </a>
          </div>
        </motion.div>
      </div>
    </section>
  );
}
