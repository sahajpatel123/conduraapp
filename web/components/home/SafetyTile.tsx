"use client";

import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/* ────────────────────────────────────────────────────────────
   SafetyTile — "Actions that cannot be undone."

   Premium, mature, subtle. An editorial split:

     Left  — the thesis, set like a magazine pull-quote: a short
             headline, one paragraph, and three quiet guarantees
             as hairline-separated rows. No icons, no badges, no
             chips. Just type and space.

     Right — a single glass panel showing one real decision being
             made, live. A proposed action, the gatekeeper's
             verdict, and the audit line it writes. It cycles
             through three real examples (ALLOW / ASK / DENY) so
             the abstract becomes concrete — but never more than
             one panel on screen. Restraint is the point.

   The whole section breathes. Generous padding, muted whites,
   a hairline divider between the columns on desktop. Nothing
   competes for attention.
   ──────────────────────────────────────────────────────────── */

type Decision = {
  action: string;
  context: string;
  verdict: "ALLOW" | "ASK" | "DENY";
  verdictNote: string;
  audit: string;
};

const DECISIONS: Decision[] = [
  {
    action: "edit auth.ts in VS Code",
    context: "WRITE · approved app",
    verdict: "ALLOW",
    verdictNote: "twin-snapshot verified",
    audit: "gate.allow · write · vscode · 14:03:22",
  },
  {
    action: "click ‘Send’ in Gmail",
    context: "NETWORK · messaging",
    verdict: "ASK",
    verdictNote: "waiting for your approval",
    audit: "gate.ask · network · gmail · 14:03:25",
  },
  {
    action: "rm -rf ~/Documents",
    context: "DESTRUCTIVE",
    verdict: "DENY",
    verdictNote: "blocked — no human at the keyboard",
    audit: "gate.deny · destructive · 14:03:27",
  },
];

const VERDICT: Record<Decision["verdict"], { text: string; dot: string; ring: string; label: string }> = {
  ALLOW: { text: "text-green-400/80", dot: "bg-green-400/80", ring: "ring-green-400/20", label: "ALLOW" },
  ASK: { text: "text-amber-400/80", dot: "bg-amber-400/80", ring: "ring-amber-400/20", label: "ASK" },
  DENY: { text: "text-red-400/80", dot: "bg-red-400/80", ring: "ring-red-400/20", label: "DENY" },
};

const GUARANTEES = [
  { label: "The gatekeeper is pure Go", note: "not a model — cannot be prompt-injected" },
  { label: "Policy lives in a YAML file", note: "you can read it, edit it, version it" },
  { label: "Every decision is audited", note: "HMAC-chained, tamper-proof, never deleted" },
];

export default function SafetyTile() {
  const [i, setI] = useState(0);

  useEffect(() => {
    const timer = setInterval(() => setI((p) => (p + 1) % DECISIONS.length), 4200);
    return () => clearInterval(timer);
  }, []);

  const d = DECISIONS[i];
  const v = VERDICT[d.verdict];

  return (
    <section
      id="safety-tile"
      className="relative w-full bg-black border-t border-white/[0.08] py-[140px] px-6 overflow-hidden"
    >
      <div className="mx-auto w-full max-w-6xl grid lg:grid-cols-[1fr_1.1fr] gap-16 lg:gap-24 items-start">
        {/* ── Left: the thesis ── */}
        <motion.div
          initial={{ opacity: 0, y: 16 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-80px" }}
          transition={{ duration: 0.8, ease: EASE_OUT }}
        >
          <div className="flex items-center gap-2.5 mb-8">
            <span className="h-px w-8 bg-white/20" />
            <span className="font-mono text-[11px] uppercase tracking-[0.25em] text-white/35">
              The safety system
            </span>
          </div>

          <h2 className="font-display text-[clamp(1.9rem,4vw,3.1rem)] font-semibold leading-[1.08] tracking-[-0.035em] text-white">
            Actions that
            <br />
            <span className="text-white/40">cannot be undone.</span>
          </h2>

          <p className="mt-7 font-body-mature text-[16px] text-white/45 leading-[1.65] max-w-md">
            Automating your machine is a survival problem, not an accuracy
            test. A model decides <span className="text-white/75">what</span> to do — deterministic
            code decides <span className="text-white/75">whether</span> it's safe. The two are
            never the same system, and no model can override the gate.
          </p>

          {/* Guarantees — hairline-separated rows, no cards */}
          <div className="mt-12 max-w-md">
            {GUARANTEES.map((g, idx) => (
              <motion.div
                key={g.label}
                initial={{ opacity: 0, y: 10 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true }}
                transition={{ delay: 0.15 * idx + 0.2, duration: 0.5, ease: EASE_OUT }}
                className={`flex items-baseline justify-between gap-4 py-4 ${
                  idx === 0 ? "border-t" : ""
                } border-white/[0.07]`}
              >
                <span className="font-body-mature text-[14px] font-medium text-white/80">
                  {g.label}
                </span>
                <span className="font-mono text-[11px] text-white/30 text-right">
                  {g.note}
                </span>
              </motion.div>
            ))}
            <div className="border-t border-white/[0.07]" />
          </div>
        </motion.div>

        {/* ── Right: one live decision panel ── */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-80px" }}
          transition={{ duration: 0.9, delay: 0.1, ease: EASE_OUT }}
          className="lg:sticky lg:top-24"
        >
          <div className="mature-panel rounded-2xl p-7">
            {/* Panel header */}
            <div className="flex items-center justify-between mb-6">
              <span className="font-mono text-[10px] uppercase tracking-[0.2em] text-white/35">
                Gatekeeper · live
              </span>
              <span className="flex items-center gap-1.5">
                <span className="w-1.5 h-1.5 rounded-full bg-white/30 animate-pulse" />
                <span className="font-mono text-[10px] text-white/30">watching</span>
              </span>
            </div>

            {/* The decision */}
            <AnimatePresence mode="wait">
              <motion.div
                key={`dec-${i}`}
                initial={{ opacity: 0, y: 10 }}
                animate={{ opacity: 1, y: 0 }}
                exit={{ opacity: 0, y: -10 }}
                transition={{ duration: 0.4, ease: EASE_OUT }}
              >
                {/* Proposed action */}
                <p className="font-mono text-[10px] uppercase tracking-wider text-white/30 mb-2">
                  proposed action
                </p>
                <div className="rounded-xl border border-white/[0.07] bg-black/40 px-4 py-4 mb-5">
                  <p className="font-body-mature text-[15px] text-white/85">
                    {d.action}
                  </p>
                  <p className="font-mono text-[11px] text-white/30 mt-1.5">
                    {d.context}
                  </p>
                </div>

                {/* Verdict */}
                <p className="font-mono text-[10px] uppercase tracking-wider text-white/30 mb-2">
                  verdict
                </p>
                <div className={`flex items-center gap-3 rounded-xl border bg-black/40 px-4 py-3.5 mb-5 ring-1 ${v.ring}`}
                  style={{ borderColor: "rgba(255,255,255,0.07)" }}
                >
                  <span className={`flex h-7 w-7 items-center justify-center rounded-lg ${v.dot}/10 ring-1 ${v.ring}`}>
                    <span className={`w-2 h-2 rounded-full ${v.dot}`} />
                  </span>
                  <div className="flex flex-col">
                    <span className={`font-display text-[15px] font-bold tracking-[-0.01em] ${v.text}`}>
                      {v.label}
                    </span>
                    <span className="font-mono text-[11px] text-white/35 mt-0.5">
                      {d.verdictNote}
                    </span>
                  </div>
                </div>

                {/* Audit line */}
                <p className="font-mono text-[10px] uppercase tracking-wider text-white/30 mb-2">
                  audit log
                </p>
                <div className="rounded-lg border border-white/[0.05] bg-black/50 px-3.5 py-2.5">
                  <p className="font-mono text-[11.5px] text-white/45 leading-relaxed break-all">
                    {d.audit}
                  </p>
                </div>
              </motion.div>
            </AnimatePresence>

            {/* Panel footer — the principle */}
            <p className="mt-6 pt-5 border-t border-white/[0.06] font-body-mature text-[12.5px] text-white/35 leading-relaxed">
              Every action — every click, keystroke, and shell command —
              passes this gate. The gate is deterministic. The log is
              append-only. You can always stop it.
            </p>
          </div>
        </motion.div>
      </div>
    </section>
  );
}