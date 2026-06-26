"use client";

import { useEffect, useState } from "react";
import { motion } from "motion/react";
import PageHeader from "@/components/shell/PageHeader";
import OrchestrationScrollStage from "@/components/orchestration/OrchestrationScrollStage";
import Reveal from "@/components/motion/Reveal";

export default function OrchestrationPage() {
  const [logs, setLogs] = useState<{ text: string; tone: string }[]>([]);

  useEffect(() => {
    const messages: { text: string; tone: string }[] = [
      { text: "[SYS] initializing SQLite state tables…", tone: "sys" },
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
    const id = window.setInterval(() => {
      if (i >= messages.length) {
        window.clearInterval(id);
        return;
      }
      const entry = messages[i];
      i += 1;
      setLogs((prev) => [...prev, entry]);
      if (i >= messages.length) window.clearInterval(id);
    }, 780);
    return () => window.clearInterval(id);
  }, []);

  return (
    <PageHeader
      eyebrow="Engine"
      title="Sub-agent delegation with a safety gate."
      titleAccent=""
      description="Condura can spawn the AI CLIs you have installed and gate each action through the deterministic Gatekeeper. Orchestrated parallel workflows are on the v0.2.0 roadmap."
    >
      <OrchestrationScrollStage />

      <Reveal>
        <div className="mt-32">
          <div className="mb-12 text-center">
            <p className="text-eyebrow mb-3">— Shared state</p>
            <h2 className="text-display mx-auto max-w-[18ch] text-balance text-[var(--color-ink)]">
              One SQLite. Every agent.
            </h2>
            <p className="text-lead mx-auto mt-5 max-w-2xl text-pretty text-[var(--color-ink-soft)]">
              Agents don&apos;t just talk to each other in a vacuum. Every thought, state change, and
              action is written to a highly-concurrent local SQLite database. This creates a
              permanent, auditable state.
            </p>
          </div>

          <div className="overflow-hidden rounded-2xl border border-[rgba(20,17,11,0.14)] bg-[var(--color-ink)] shadow-[var(--shadow-float)]">
            <div className="flex items-center gap-2 border-b border-[rgba(244,239,228,0.08)] px-5 py-3">
              <span className="h-2 w-2 rounded-full bg-[var(--color-pollen)]" />
              <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-[rgba(244,239,228,0.5)]">
                condura-bus-monitor.db
              </span>
            </div>
            <div className="relative h-[400px] overflow-y-auto p-6 font-mono text-[13px] leading-relaxed">
              <div className="absolute right-4 top-3 z-10 rounded-full border border-[rgba(244,239,228,0.15)] bg-[rgba(244,239,228,0.08)] px-3 py-1 font-mono text-[10px] uppercase tracking-wider text-[rgba(244,239,228,0.65)]">
                Illustrative sequence — full orchestration is v0.2.0
              </div>
              {logs.map((log, i) => {
                if (!log?.text) return null;
                return (
                <motion.div
                  key={i}
                  initial={{ opacity: 0, x: -10 }}
                  animate={{ opacity: 1, x: 0 }}
                  className={`mb-2 ${busTone(log.tone)}`}
                >
                  <span className="mr-3 opacity-50">
                    {new Date().toISOString().split("T")[1].split(".")[0]}
                  </span>
                  {log.text}
                </motion.div>
                );
              })}
              {logs.length < 14 && (
                <motion.div
                  animate={{ opacity: [1, 0] }}
                  transition={{ repeat: Infinity, duration: 0.8 }}
                  className="mt-1 inline-block h-4 w-2 bg-[var(--color-paper)]"
                />
              )}
            </div>
          </div>
        </div>
      </Reveal>

      <div className="mt-32">
        <Reveal>
          <div className="mb-12 text-center">
            <p className="text-eyebrow mb-3">— Performance</p>
            <h2 className="text-display mx-auto max-w-[14ch] text-balance text-[var(--color-ink)]">
              Built for speed.
            </h2>
            <p className="text-lead mx-auto mt-5 max-w-2xl text-pretty text-[var(--color-ink-soft)]">
              No Python dependency hell. Condura is a single standalone binary. The core engine leans
              on native OS primitives.
            </p>
          </div>
        </Reveal>
        <div className="grid gap-6 md:grid-cols-3">
          {[
            {
              stat: "< 50ms",
              label: "Agent spawn time (design target)",
              desc: "Lightweight goroutines instead of heavy system processes. The sub-agent spawn path is in-process; no formal benchmark has been published yet.",
            },
            {
              stat: "High write throughput",
              label: "SQLite WAL concurrency",
              desc: "SQLite in WAL mode supports high concurrent write rates. Specific event/sec numbers vary by workload; no formal benchmark has been published yet.",
            },
            {
              stat: "0 bytes",
              label: "Cloud storage",
              desc: "Every byte of your code stays on your local filesystem.",
            },
          ].map((item, i) => (
            <Reveal key={item.label} delay={i * 0.1}>
              <div className="surface-card flex h-full flex-col items-center p-8 text-center transition-colors hover:bg-[var(--color-paper-deep)]">
                <div className="font-display text-[44px] leading-none text-[var(--color-ink)]">
                  {item.stat}
                </div>
                <div className="text-mono-label mb-4 mt-3">{item.label}</div>
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
    case "sys":
      return "text-[rgba(244,239,228,0.5)]";
    case "agent":
      return "text-[rgba(244,239,228,0.85)]";
    case "gate":
      return "text-[var(--color-pollen-light)]";
    case "user":
      return "text-[var(--color-synapse-light)]";
    default:
      return "text-[rgba(244,239,228,0.65)]";
  }
}
