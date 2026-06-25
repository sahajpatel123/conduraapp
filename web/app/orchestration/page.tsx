"use client";

import { useEffect, useRef, useState } from "react";
import { motion, useScroll, useSpring, useTransform } from "motion/react";
import PageHeader from "@/components/shell/PageHeader";
import Reveal from "@/components/motion/Reveal";
import { EASE_OUT } from "@/lib/motion";

export default function OrchestrationPage() {
  const containerRef = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({ target: containerRef, offset: ["start start", "end end"] });
  const smooth = useSpring(scrollYProgress, { damping: 20, stiffness: 100 });

  const y1 = useTransform(smooth, [0, 1], [0, -200]);
  const y2 = useTransform(smooth, [0, 1], [0, -400]);
  const y3 = useTransform(smooth, [0, 1], [0, -600]);
  const o1 = useTransform(smooth, [0, 0.3, 0.4], [0, 1, 0.2]);
  const o2 = useTransform(smooth, [0.3, 0.6, 0.7], [0, 1, 0.2]);
  const o3 = useTransform(smooth, [0.6, 0.9, 1], [0, 1, 1]);

  const [logs, setLogs] = useState<{ text: string; tone: string }[]>([]);
  useEffect(() => {
    const messages: { text: string; tone: string }[] = [
      { text: "[SYS] initializing SQLite event bus…", tone: "sys" },
      { text: "[SYS] memory mapped to ~/.condura/synaptic.db", tone: "sys" },
      { text: "[AGENT:strategist] received intent: 'Refactor auth module'", tone: "agent" },
      { text: "[AGENT:strategist] decomposing into 3 subtasks.", tone: "agent" },
      { text: "[BUS] spawned agent-01 (AST parser)", tone: "bus" },
      { text: "[BUS] spawned agent-02 (dependency graph)", tone: "bus" },
      { text: "[BUS] spawned agent-03 (state machine analyzer)", tone: "bus" },
      { text: "[AGENT-01] parsing /src/auth.ts…", tone: "agent" },
      { text: "[AGENT-02] traversing imports…", tone: "agent" },
      { text: "[AGENT-03] found 4 unhandled state transitions.", tone: "agent" },
      { text: "[GATEKEEPER] requesting FS write permission…", tone: "gate" },
      { text: "[USER] granted.", tone: "user" },
      { text: "[BUS] applying diffs to /src/auth.ts", tone: "bus" },
      { text: "[SYS] commit successful. hash: 8f4a2b1", tone: "sys" },
    ];
    let i = 0;
    const id = setInterval(() => {
      if (i < messages.length) {
        setLogs((prev) => [...prev, messages[i]]);
        i++;
      } else {
        clearInterval(id);
      }
    }, 780);
    return () => clearInterval(id);
  }, []);

  return (
    <PageHeader
      eyebrow="Engine"
      title="Massive parallel"
      titleAccent="workflows."
      description="Condura doesn't just run agents sequentially. It spins up highly concurrent, local swarms that communicate through a fast SQLite event bus. This is the story of how it thinks."
    >
      {/* ── 3-phase sticky scroll ── */}
      <div ref={containerRef} className="relative mt-12 h-[300vh]">
        <div className="sticky top-28 flex h-[78vh] items-center justify-center overflow-hidden rounded-3xl border border-[rgba(20,17,11,0.14)] bg-[var(--color-paper-warm)] shadow-[var(--shadow-card)]">
          <div className="paper-grain absolute inset-0" />

          {/* Phase 1 */}
          <motion.div style={{ y: y1, opacity: o1 }} className="absolute inset-0 flex flex-col items-center justify-center p-8">
            <div className="rounded-2xl border border-[rgba(20,17,11,0.12)] bg-[var(--color-paper)] p-6 text-center">
              <p className="text-mono-label mb-2">Phase 1</p>
              <p className="font-display text-[26px] text-[var(--color-ink)]">Decomposition</p>
              <p className="mt-2 max-w-sm text-[14px] text-[var(--color-ink-mute)]">A single complex prompt is torn into discrete sub-tasks. The Strategist plans the assault.</p>
            </div>
            <div className="mt-8 flex gap-4">
              {[0, 1, 2, 3].map((i) => (
                <motion.div
                  key={i}
                  initial={{ opacity: 0, scale: 0.8 }}
                  whileInView={{ opacity: 1, scale: 1 }}
                  transition={{ delay: i * 0.1, type: "spring" }}
                  className="grid h-16 w-16 place-items-center rounded-2xl border border-[rgba(20,17,11,0.18)] bg-[var(--color-paper)] shadow-[var(--shadow-paper)]"
                >
                  <span className="font-mono text-[11px] text-[var(--color-ink-mute)]">T-{i + 1}</span>
                </motion.div>
              ))}
            </div>
          </motion.div>

          {/* Phase 2 */}
          <motion.div style={{ y: y2, opacity: o2 }} className="absolute inset-0 flex flex-col items-center justify-center p-8">
            <div className="rounded-2xl border border-[rgba(20,17,11,0.12)] bg-[var(--color-paper)] p-6 text-center">
              <p className="text-mono-label mb-2">Phase 2</p>
              <p className="font-display text-[26px] text-[var(--color-ink)]">Parallel fan-out</p>
              <p className="mt-2 max-w-sm text-[14px] text-[var(--color-ink-mute)]">Condura spawns lightweight sub-agents for each task, running them concurrently to slash execution time.</p>
            </div>
            <div className="relative mt-8 w-full max-w-lg overflow-hidden rounded-3xl border border-[rgba(20,17,11,0.14)] bg-[var(--color-ink)] p-6">
              <div className="absolute inset-0 flex items-center justify-center">
                <div className="h-16 w-16 rounded-full bg-[var(--color-synapse-glow)] blur-2xl animate-pulse" />
              </div>
              <div className="relative z-10 flex flex-col gap-3">
                {["Spawning react-agent…", "Spawning rust-agent…", "Mounting DOM analyzer…", "Starting headless browser…"].map((t, i) => (
                  <motion.div key={i} initial={{ x: -20, opacity: 0 }} whileInView={{ x: 0, opacity: 1 }} transition={{ delay: i * 0.15 }} className="flex items-center gap-3 font-mono text-[12px] text-[rgba(244,239,228,0.7)]">
                    <span className="h-2 w-2 rounded-full bg-[var(--color-synapse-light)]" />
                    {t}
                  </motion.div>
                ))}
              </div>
            </div>
          </motion.div>

          {/* Phase 3 */}
          <motion.div style={{ y: y3, opacity: o3 }} className="absolute inset-0 flex flex-col items-center justify-center p-8 text-center">
            <div className="rounded-2xl border border-[rgba(20,17,11,0.12)] bg-[var(--color-paper)] p-6">
              <p className="text-mono-label mb-2">Phase 3</p>
              <p className="font-display text-[26px] text-[var(--color-ink)]">Deterministic resolution</p>
              <p className="mt-2 max-w-sm text-[14px] text-[var(--color-ink-mute)]">Results are collated, verified by strict logic (not a hallucinating LLM), and committed to your workspace.</p>
            </div>
            <motion.div
              className="mt-8 grid h-32 w-32 place-items-center rounded-[2rem] border-2 border-[var(--color-synapse)] bg-[rgba(11,61,46,0.08)]"
              animate={{ rotate: 360 }}
              transition={{ duration: 10, repeat: Infinity, ease: "linear" }}
            >
              <div className="h-16 w-16 rounded-full border border-[var(--color-synapse)] bg-[var(--color-paper)]" />
            </motion.div>
            <p className="mt-6 font-mono text-[12px] text-[var(--color-ink-mute)]">Diffs merged. AST verified. Lockfile updated.</p>
          </motion.div>
        </div>
      </div>

      {/* ── The SQLite event bus (live terminal) ── */}
      <Reveal>
        <div className="mt-32">
          <div className="mb-12 text-center">
            <p className="text-eyebrow mb-3">— The event bus</p>
            <h2 className="text-display text-[var(--color-ink)] max-w-[18ch] mx-auto text-balance">The SQLite event bus.</h2>
            <p className="text-lead mt-5 max-w-2xl mx-auto text-[var(--color-ink-soft)] text-pretty">
              Agents don&apos;t just talk to each other in a vacuum. Every thought, state change, and action is written to a highly-concurrent local SQLite database. This creates a permanent, auditable, replayable state.
            </p>
          </div>

          <div className="overflow-hidden rounded-2xl border border-[rgba(20,17,11,0.14)] bg-[var(--color-ink)] shadow-[var(--shadow-float)]">
            <div className="flex items-center gap-2 border-b border-[rgba(244,239,228,0.08)] px-5 py-3">
              <span className="h-2 w-2 rounded-full bg-[var(--color-pollen)]" />
              <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-[rgba(244,239,228,0.5)]">condura-bus-monitor.db</span>
            </div>
            <div className="relative h-[400px] overflow-y-auto p-6 font-mono text-[13px] leading-relaxed">
              {logs.map((log, i) => (
                <motion.div key={i} initial={{ opacity: 0, x: -10 }} animate={{ opacity: 1, x: 0 }} className={`mb-2 ${busTone(log.tone)}`}>
                  <span className="mr-3 opacity-50">{new Date().toISOString().split("T")[1].split(".")[0]}</span>
                  {log.text}
                </motion.div>
              ))}
              {logs.length < 14 && (
                <motion.div animate={{ opacity: [1, 0] }} transition={{ repeat: Infinity, duration: 0.8 }} className="mt-1 inline-block h-4 w-2 bg-[var(--color-paper)]" />
              )}
            </div>
          </div>
        </div>
      </Reveal>

      {/* ── Perf stats ── */}
      <div className="mt-32">
        <Reveal>
          <div className="mb-12 text-center">
            <p className="text-eyebrow mb-3">— Performance</p>
            <h2 className="text-display text-[var(--color-ink)] max-w-[14ch] mx-auto text-balance">Built for speed.</h2>
            <p className="text-lead mt-5 max-w-2xl mx-auto text-[var(--color-ink-soft)] text-pretty">
              No Python dependency hell. Condura is a single standalone binary. The core engine leans on native OS primitives.
            </p>
          </div>
        </Reveal>
        <div className="grid gap-6 md:grid-cols-3">
          {[
            { stat: "< 50ms", label: "Agent spawn time", desc: "Lightweight goroutines instead of heavy system processes." },
            { stat: "100k+", label: "Events per second", desc: "SQLite WAL mode handles massive concurrent write loads effortlessly." },
            { stat: "0 bytes", label: "Cloud storage", desc: "Every byte of your code stays on your local filesystem." },
          ].map((item, i) => (
            <Reveal key={item.label} delay={i * 0.1}>
              <div className="surface-card flex h-full flex-col items-center p-8 text-center transition-colors hover:bg-[var(--color-paper-deep)]">
                <div className="font-display text-[44px] leading-none text-[var(--color-ink)]">{item.stat}</div>
                <div className="text-mono-label mt-3 mb-4">{item.label}</div>
                <p className="text-body text-[var(--color-ink-mute)]">{item.desc}</p>
              </div>
            </Reveal>
          ))}
        </div>
      </div>
    </PageHeader>
  );
}

function busTone(tone: string) {
  switch (tone) {
    case "sys": return "text-[rgba(244,239,228,0.5)]";
    case "agent": return "text-[rgba(244,239,228,0.85)]";
    case "gate": return "text-[var(--color-pollen-light)]";
    case "user": return "text-[var(--color-synapse-light)]";
    default: return "text-[rgba(244,239,228,0.65)]";
  }
}
