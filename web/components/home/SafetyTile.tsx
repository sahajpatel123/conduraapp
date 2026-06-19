"use client";

import { motion } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/* ────────────────────────────────────────────────────────────
   SafetyTile — "Actions that cannot be undone."

   Minimal. The safety system is a flow: one line, four stops,
   each stricter. That's it. No cards, no badges, no chips, no
   abstract image. Just the escalation, shown as a journey.

   Each stop carries three things and nothing else:
     · the tier name  (READ → WRITE → NETWORK → DESTRUCTIVE)
     · one real example
     · one verdict word

   The connecting line shifts color as risk rises. A single
   sentence under the flow states the rule. That's the whole
   section — the user reads the direction of travel and
   understands safety in one glance.
   ──────────────────────────────────────────────────────────── */

const FLOW = [
  { name: "READ", example: "screenshot", verdict: "allowed", tone: "green" },
  { name: "WRITE", example: "edit a file", verdict: "verified", tone: "sky" },
  { name: "NETWORK", example: "send an email", verdict: "asks you", tone: "amber" },
  { name: "DESTRUCTIVE", example: "delete files", verdict: "blocks", tone: "red" },
] as const;

const COLOR: Record<string, { dot: string; line: string; text: string }> = {
  green: { dot: "bg-green-400/70", line: "bg-green-400/30", text: "text-green-400/80" },
  sky: { dot: "bg-sky-400/70", line: "bg-sky-400/30", text: "text-sky-400/80" },
  amber: { dot: "bg-amber-400/70", line: "bg-amber-400/30", text: "text-amber-400/80" },
  red: { dot: "bg-red-400/70", line: "bg-red-400/30", text: "text-red-400/80" },
};

export default function SafetyTile() {
  return (
    <section
      id="safety-tile"
      className="relative w-full bg-[#000000] py-[120px] px-6 text-white overflow-hidden border-t border-white/[0.08]"
    >
      <div className="mx-auto w-full max-w-4xl">
        {/* ── Header — one short line, no icon ── */}
        <motion.div
          initial={{ opacity: 0, y: 14 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-60px" }}
          transition={{ duration: 0.7, ease: EASE_OUT }}
          className="mb-20"
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
        </motion.div>

        {/* ── The flow: one line, four stops ── */}
        <motion.div
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true, margin: "-40px" }}
          transition={{ duration: 0.8, ease: EASE_OUT }}
          className="relative"
        >
          {/* The connecting line — sits behind the dots, shifts color left→right */}
          <div className="absolute left-0 right-0 top-[22px] h-px bg-white/[0.08]" aria-hidden />
          <div className="absolute left-0 top-[22px] h-px bg-gradient-to-r from-green-400/40 via-amber-400/40 to-red-400/50" aria-hidden />

          {/* Stops */}
          <div className="relative grid grid-cols-4 gap-3">
            {FLOW.map((stop, i) => {
              const c = COLOR[stop.tone];
              return (
                <motion.div
                  key={stop.name}
                  initial={{ opacity: 0, y: 12 }}
                  whileInView={{ opacity: 1, y: 0 }}
                  viewport={{ once: true }}
                  transition={{ delay: 0.15 * i + 0.2, duration: 0.5, ease: EASE_OUT }}
                  className="flex flex-col items-center text-center"
                >
                  {/* Dot on the line */}
                  <div className={`relative z-10 w-[14px] h-[14px] rounded-full ${c.dot} mb-7 ring-4 ring-black`} />

                  {/* Tier name */}
                  <span className="font-display text-[clamp(13px,1.4vw,15px)] font-bold tracking-[-0.01em] text-white mb-1.5">
                    {stop.name}
                  </span>

                  {/* One example */}
                  <span className="font-body-mature text-[12px] text-white/35 mb-3">
                    {stop.example}
                  </span>

                  {/* Verdict — one word */}
                  <span className={`font-mono text-[11px] font-medium ${c.text}`}>
                    {stop.verdict}
                  </span>
                </motion.div>
              );
            })}
          </div>
        </motion.div>

        {/* ── One sentence — the rule, in plain words ── */}
        <motion.p
          initial={{ opacity: 0, y: 10 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ delay: 0.4, duration: 0.6, ease: EASE_OUT }}
          className="mt-20 text-center font-body-mature text-[14px] text-white/40 leading-relaxed max-w-xl mx-auto"
        >
          Every action is classified before it runs. The higher the risk, the
          stricter the gate. A model decides <span className="text-white/70">what</span> to do —
          deterministic code decides <span className="text-white/70">whether</span> it's safe.
          No model can override it.
        </motion.p>
      </div>
    </section>
  );
}