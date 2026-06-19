"use client";

import { motion } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/* ────────────────────────────────────────────────────────────
   SafetyTile — "Actions that cannot be undone."

   The old layout was a shield icon, three lines, a decorative
   Unsplash image, and three cards with abstract feature names
   ("Twin-Snapshot Verification", etc.) that meant nothing to a
   normal user.

   This replacement is the actual safety system, shown as the thing
   it is: a four-tier blast-radius ladder. Every action Condura takes
   is classified before it runs — READ, WRITE, NETWORK, or
   DESTRUCTIVE — and each tier gets a progressively stricter gate.
   That is the most concrete, most understandable way to explain
   safety: show the real decision, with real examples.

   The layout is a vertical ladder where risk ascends and the gate
   tightens. No abstract image. The content is the visual.
   ──────────────────────────────────────────────────────────── */

type Tier = {
  level: string;
  name: string;
  meaning: string;
  examples: string[];
  verdict: "allow" | "verify" | "ask" | "human";
  verdictLabel: string;
  verdictNote: string;
  tone: "green" | "sky" | "amber" | "red";
};

const TIERS: Tier[] = [
  {
    level: "01",
    name: "READ",
    meaning: "Look, don't touch.",
    examples: ["screenshot", "read a file", "inspect the AX tree", "copy text"],
    verdict: "allow",
    verdictLabel: "ALLOWED",
    verdictNote: "no prompt, runs freely",
    tone: "green",
  },
  {
    level: "02",
    name: "WRITE",
    meaning: "Change something on your machine.",
    examples: ["edit a file", "type text", "paste content", "save a draft"],
    verdict: "verify",
    verdictLabel: "VERIFIED",
    verdictNote: "twin-snapshot check, then run",
    tone: "sky",
  },
  {
    level: "03",
    name: "NETWORK",
    meaning: "Reach out to the world.",
    examples: ["click a link", "submit a form", "send an email", "post a message"],
    verdict: "ask",
    verdictLabel: "ASKS FIRST",
    verdictNote: "native dialog — you approve or deny",
    tone: "amber",
  },
  {
    level: "04",
    name: "DESTRUCTIVE",
    meaning: "Cannot be undone.",
    examples: ["delete files", "format a disk", "transfer money", "authorize a purchase"],
    verdict: "human",
    verdictLabel: "BLOCKS WITHOUT YOU",
    verdictNote: "real human at the keyboard — no exceptions, no overrides",
    tone: "red",
  },
];

const TONE: Record<Tier["tone"], { bar: string; text: string; border: string; bg: string; dot: string }> = {
  green: {
    bar: "bg-green-400/70",
    text: "text-green-400/80",
    border: "border-green-400/20",
    bg: "bg-green-400/[0.06]",
    dot: "bg-green-400/70",
  },
  sky: {
    bar: "bg-sky-400/70",
    text: "text-sky-400/80",
    border: "border-sky-400/20",
    bg: "bg-sky-400/[0.06]",
    dot: "bg-sky-400/70",
  },
  amber: {
    bar: "bg-amber-400/70",
    text: "text-amber-400/80",
    border: "border-amber-400/20",
    bg: "bg-amber-400/[0.06]",
    dot: "bg-amber-400/70",
  },
  red: {
    bar: "bg-red-400/70",
    text: "text-red-400/80",
    border: "border-red-400/20",
    bg: "bg-red-400/[0.06]",
    dot: "bg-red-400/70",
  },
};

export default function SafetyTile() {
  return (
    <section
      id="safety-tile"
      className="relative w-full bg-[#000000] py-[140px] px-6 text-white overflow-hidden border-t border-white/[0.08]"
    >
      <div className="mx-auto w-full max-w-5xl">
        {/* ── Header ── */}
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-60px" }}
          transition={{ duration: 0.7, ease: EASE_OUT }}
          className="max-w-2xl mb-16"
        >
          <div className="flex items-center gap-2.5 mb-5">
            <span className="h-px w-8 bg-white/20" />
            <span className="font-mono text-[11px] uppercase tracking-[0.25em] text-white/35">
              The safety system
            </span>
          </div>
          <h2 className="font-display text-[clamp(1.75rem,3.6vw,2.75rem)] font-semibold leading-[1.1] tracking-[-0.03em]">
            Actions that <span className="text-white/40">cannot be undone.</span>
          </h2>
          <p className="mt-5 font-body-mature text-[15px] text-white/45 leading-relaxed max-w-xl">
            Every action is classified before it runs — by how much damage it
            could do. The higher the risk, the stricter the gate. No model
            decides whether it's safe; deterministic code does. A model can
            never override it.
          </p>
        </motion.div>

        {/* ── The blast-radius ladder ── */}
        <div className="flex flex-col gap-3">
          {TIERS.map((tier, i) => {
            const tone = TONE[tier.tone];
            return (
              <motion.div
                key={tier.level}
                initial={{ opacity: 0, x: -16 }}
                whileInView={{ opacity: 1, x: 0 }}
                viewport={{ once: true, margin: "-40px" }}
                transition={{ delay: i * 0.1, duration: 0.6, ease: EASE_OUT }}
                className="relative"
              >
                <div
                  className={`group relative grid grid-cols-1 md:grid-cols-[auto_1fr_auto] gap-4 md:gap-6 items-center rounded-2xl border bg-white/[0.015] px-5 py-5 md:py-6 transition-colors hover:bg-white/[0.03] ${
                    i === TIERS.length - 1 ? tone.border : "border-white/[0.08] hover:border-white/[0.14]"
                  }`}
                >
                  {/* Left rail — level + severity bar */}
                  <div className="flex items-center gap-4">
                    <span className="font-mono text-[12px] text-white/30 tabular-nums">
                      {tier.level}
                    </span>
                    <span className={`w-1 h-10 rounded-full ${tone.bar}`} />
                  </div>

                  {/* Middle — name + meaning + examples */}
                  <div className="min-w-0">
                    <div className="flex items-baseline gap-3 mb-1.5">
                      <span className="font-display text-[clamp(18px,2vw,22px)] font-bold tracking-[-0.01em] text-white">
                        {tier.name}
                      </span>
                      <span className="font-body-mature text-[13px] text-white/40">
                        {tier.meaning}
                      </span>
                    </div>
                    {/* Example chips */}
                    <div className="flex flex-wrap gap-1.5">
                      {tier.examples.map((ex) => (
                        <span
                          key={ex}
                          className="font-mono text-[10.5px] text-white/45 rounded-md border border-white/[0.07] bg-white/[0.02] px-2 py-0.5"
                        >
                          {ex}
                        </span>
                      ))}
                    </div>
                  </div>

                  {/* Right — verdict badge */}
                  <div className={`flex items-center gap-2.5 rounded-xl border ${tone.border} ${tone.bg} px-3.5 py-2.5 md:self-stretch`}>
                    <span className={`w-2 h-2 rounded-full ${tone.dot} shrink-0`} />
                    <div className="flex flex-col">
                      <span className={`font-mono text-[11px] font-semibold tracking-wider ${tone.text}`}>
                        {tier.verdictLabel}
                      </span>
                      <span className="font-body-mature text-[10.5px] text-white/45 leading-tight mt-0.5">
                        {tier.verdictNote}
                      </span>
                    </div>
                  </div>
                </div>
              </motion.div>
            );
          })}
        </div>

        {/* ── Footer — the guarantee in plain words ── */}
        <motion.div
          initial={{ opacity: 0, y: 12 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ delay: 0.3, duration: 0.6, ease: EASE_OUT }}
          className="mt-12 flex flex-wrap items-center justify-center gap-x-6 gap-y-2 font-mono text-[11px] text-white/30"
        >
          <span className="flex items-center gap-1.5">
            <span className="w-1.5 h-1.5 rounded-full bg-white/40" />
            gatekeeper is pure Go, not a model
          </span>
          <span className="text-white/15">·</span>
          <span>policy lives in a YAML file you can read and edit</span>
          <span className="text-white/15">·</span>
          <span>every decision written to a tamper-proof audit log</span>
          <span className="text-white/15">·</span>
          <span>you can always stop it — four independent kill switches</span>
        </motion.div>
      </div>
    </section>
  );
}