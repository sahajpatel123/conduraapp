"use client";

import { useEffect, useState } from "react";
import { motion } from "motion/react";
import PageHeader from "@/components/shell/PageHeader";
import Reveal from "@/components/motion/Reveal";

export default function SecurityPageClient() {
  const [unlocked, setUnlocked] = useState(false);
  const [auditLogs, setAuditLogs] = useState<{ text: string; tone: string }[]>([]);

  useEffect(() => {
    const logs: { text: string; tone: string }[] = [
      { text: "actor=condura  action=READ     resource=/src/index.ts", tone: "read" },
      { text: "actor=strategist action=PROPOSE  resource=npm install express", tone: "propose" },
      { text: "actor=gatekeeper action=BLOCK    resource=npm install express", tone: "block" },
      { text: "actor=gatekeeper action=PROMPT   resource=npm install express", tone: "prompt" },
      { text: "actor=user       action=GRANT    resource=npm install express", tone: "grant" },
      { text: "actor=condura    action=EXEC     resource=npm install express", tone: "exec" },
      { text: "actor=condura    action=WRITE    resource=package.json", tone: "write" },
    ];
    let i = 0;
    const id = window.setInterval(() => {
      if (i >= logs.length) {
        window.clearInterval(id);
        return;
      }
      const entry = logs[i];
      i += 1;
      setAuditLogs((prev) => [...prev, entry]);
      if (i >= logs.length) window.clearInterval(id);
    }, 1100);
    return () => window.clearInterval(id);
  }, []);

  return (
    <PageHeader
      eyebrow="Security"
      title="Capability without"
      titleAccent="the blank check."
      description="Models hallucinate, prompts get injected, and some actions cannot be undone. Condura draws a hard line between thinking and acting with a deterministic permission gatekeeper — not a model, ever."
    >
      {/* ── The Gatekeeper vault ── */}
      <Reveal>
        <div className="surface-ink relative flex h-[440px] w-full items-center justify-center overflow-hidden p-8">
          <motion.div
            animate={{ opacity: unlocked ? 0.25 : 0.08, scale: unlocked ? 1.3 : 1 }}
            transition={{ duration: 1 }}
            className="absolute inset-0"
            style={{ background: "radial-gradient(circle at center, rgba(156,232,200,0.25), transparent 55%)" }}
          />
          <div
            className="relative z-10 grid h-72 w-72 place-items-center"
            onMouseEnter={() => setUnlocked(true)}
            onMouseLeave={() => setUnlocked(false)}
          >
            <motion.div
              animate={{ rotate: unlocked ? 180 : 0 }}
              transition={{ type: "spring", stiffness: 50, damping: 20 }}
              className="absolute inset-0 rounded-full border-2 border-dashed border-[rgba(156,232,200,0.25)]"
            />
            <motion.div
              animate={{ rotate: unlocked ? -90 : 0 }}
              transition={{ type: "spring", stiffness: 60, damping: 25 }}
              className="absolute inset-8 flex items-center justify-center rounded-full border border-[rgba(244,239,228,0.12)]"
            >
              {[0, 90, 180, 270].map((deg) => (
                <motion.div key={deg} style={{ rotate: deg }} className="absolute flex h-full w-full justify-between">
                  <motion.div animate={{ scaleX: unlocked ? 0 : 1 }} transition={{ duration: 0.3 }} className="h-[2px] w-4 origin-left bg-[rgba(244,239,228,0.4)]" />
                  <motion.div animate={{ scaleX: unlocked ? 0 : 1 }} transition={{ duration: 0.3 }} className="h-[2px] w-4 origin-right bg-[rgba(244,239,228,0.4)]" />
                </motion.div>
              ))}
            </motion.div>
            <motion.div
              animate={{ scale: unlocked ? 1.05 : 1, boxShadow: unlocked ? "0 0 60px rgba(156,232,200,0.25), inset 0 0 20px rgba(156,232,200,0.08)" : "0 0 0px rgba(0,0,0,0)" }}
              transition={{ duration: 0.5 }}
              className="relative z-20 flex h-44 w-44 flex-col items-center justify-center rounded-full border border-[rgba(244,239,228,0.25)] bg-[var(--color-ink)]"
            >
              <div className="mb-2 font-mono text-[10px] uppercase tracking-[0.2em] text-[rgba(244,239,228,0.5)]">
                {unlocked ? "Unlocked" : "Hover to unlock"}
              </div>
              <div className="font-display text-[22px] tracking-tight text-[var(--color-paper)]">
                {unlocked ? "ALLOW" : "LOCKED"}
              </div>
              {unlocked && (
                <motion.div initial={{ opacity: 0, y: 10 }} animate={{ opacity: 1, y: 0 }} className="absolute -bottom-14 flex gap-2">
                  {["READ FS", "PORT 3000"].map((b) => (
                    <span key={b} className="rounded-full border border-[rgba(156,232,200,0.4)] px-2.5 py-1 font-mono text-[10px] uppercase tracking-wider text-[var(--color-synapse-light)]">
                      {b}
                    </span>
                  ))}
                </motion.div>
              )}
            </motion.div>
          </div>
        </div>
      </Reveal>

      {/* ── Core principles ── */}
      <div className="mt-20 grid gap-5 md:grid-cols-3">
        {[
          { title: "Deterministic rules", desc: "Security rules are hard-coded, not written in a prompt. A model cannot talk the Gatekeeper into dropping its guard — it isn't a model." },
          { title: "Air-gapped memory", desc: "Your workspace memory and skills live on your local SSD. No proprietary code is ever indexed in the cloud. (Vector embeddings are on the v0.2.0 roadmap.)" },
          { title: "Three kill switches", desc: "Three kill switches: a hard hotkey, a watchdog timer, and a network guard. The guard is in-process in v0.1.x; a hard separate-process guard is planned for v0.2.0. The agent cannot disable the hotkey or watchdog." },
        ].map((feat, i) => (
          <Reveal key={feat.title} delay={i * 0.1}>
            <div className="surface-card h-full p-7">
              <h3 className="font-display text-[20px] leading-tight text-[var(--color-ink)]">{feat.title}</h3>
              <p className="mt-3 text-body text-[var(--color-ink-mute)]">{feat.desc}</p>
            </div>
          </Reveal>
        ))}
      </div>

      {/* ── The immutable audit trail ── */}
      <Reveal>
        <div className="mt-32 flex flex-col items-center gap-12 md:flex-row md:items-center md:gap-16">
          <div className="w-full flex-1">
            <div className="overflow-hidden rounded-2xl border border-[rgba(20,17,11,0.14)] bg-[var(--color-ink)] shadow-[var(--shadow-float)]">
              <div className="flex items-center gap-2 border-b border-[rgba(244,239,228,0.08)] px-4 py-3">
                <span className="h-2 w-2 rounded-full bg-[var(--color-pollen)]" />
                <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-[rgba(244,239,228,0.5)]">condura_audit.sqlite</span>
              </div>
              <div className="h-[340px] overflow-y-auto p-5 font-mono text-[12.5px] leading-[1.85]">
                {auditLogs.map((log, idx) => {
                  if (!log?.text) return null;
                  return (
                  <motion.div key={idx} initial={{ opacity: 0, x: -10 }} animate={{ opacity: 1, x: 0 }} className="mb-1.5">
                    <span className="mr-3 text-[rgba(244,239,228,0.35)]">{String(idx + 1).padStart(4, "0")}</span>
                    <span className={toneClass(log.tone)}>{log.text}</span>
                  </motion.div>
                  );
                })}
                {auditLogs.length < 7 && (
                  <motion.span animate={{ opacity: [1, 0] }} transition={{ repeat: Infinity, duration: 0.8 }} className="inline-block h-4 w-2 bg-[var(--color-paper)]" />
                )}
              </div>
            </div>
          </div>
          <div className="flex-1">
            <p className="text-eyebrow mb-3">— The audit trail</p>
            <h2 className="text-display text-[var(--color-ink)] max-w-[14ch] text-balance">Immutable proof of every move.</h2>
            <p className="text-lead mt-5 max-w-[44ch] text-[var(--color-ink-soft)] text-pretty">
              Trust requires verification. Every action Condura takes — from reading a file to spawning a subprocess to hitting an API — is logged locally in an SQLite database.
            </p>
            <p className="text-body mt-4 max-w-[48ch] text-[var(--color-ink-mute)]">
              These logs are HMAC-chained and append-only. If the key on your machine stays secret, the log is tamper-evident.
            </p>
          </div>
        </div>
      </Reveal>
    </PageHeader>
  );
}

function toneClass(tone: string) {
  switch (tone) {
    case "block": return "text-[#e88a7e]";
    case "grant": return "text-[var(--color-synapse-light)]";
    case "prompt": return "text-[var(--color-pollen-light)]";
    case "exec": return "text-[var(--color-paper)]";
    case "write": return "text-[var(--color-paper)]";
    default: return "text-[rgba(244,239,228,0.7)]";
  }
}
