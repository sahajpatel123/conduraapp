"use client";

import { motion } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/* ────────────────────────────────────────────────────────────
   SafetyTile — "The Rule."

   Completely different from the previous three attempts (flow line,
   ladder, editorial split). None of those felt premium because they
   all used containers — cards, panels, badges, chips. Premium
   editorial design doesn't box content; it lets type and space do
   the work.

   This section has zero containers. No cards, no panels, no glass,
   no borders around content. Just typography on black, hairlines as
   the only structural element, and generous space. The safety
   principle is presented as a constitution — a dramatic centered
   statement, followed by four numbered rules, each one line.

   The layout is deliberately asymmetric and spacious:
     · A massive centered headline that fills the viewport width
     · Four rules laid out as plain sentences — a small dot+index
       ordinal, a one-line statement as the lead, then the
       explanation — separated only by hairlines
     · A quiet footer line stating where the gate lives

   The titles are the content: each rule is a single plain sentence
   a first-time visitor understands at a glance, not jargon. This is
   how Apple's privacy page and the best editorial sites handle
   trust: the words are the design.
   ──────────────────────────────────────────────────────────── */

const RULES = [
  {
    title: "A model can't grade its own homework.",
    text: "The AI proposes the action. A separate, deterministic gatekeeper approves or blocks it. They are never the same system — a model can never decide it's safe.",
  },
  {
    title: "Every action is sorted before it runs.",
    text: "By how much damage it could do: read, write, network, or destructive. The higher the blast radius, the stricter the gate. Nothing skips this step.",
  },
  {
    title: "The dangerous stuff needs you present.",
    text: "Sending an email, deleting files, transferring money — these open a native dialog on your screen. If you're away from the keyboard, they queue. They don't run.",
  },
  {
    title: "Every decision is written down.",
    text: "Allowed or denied, every verdict lands in a tamper-proof, HMAC-chained audit log that never deletes. If something goes wrong, you can prove exactly what happened.",
  },
];

export default function SafetyTile() {
  return (
    <section
      id="safety-tile"
      className="relative w-full bg-black border-t border-white/[0.08] py-[160px] px-6 overflow-hidden"
    >
      <div className="mx-auto w-full max-w-4xl">
        {/* ── Eyebrow ── */}
        <motion.div
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true, margin: "-80px" }}
          transition={{ duration: 0.8, ease: EASE_OUT }}
          className="flex justify-center mb-12"
        >
          <span className="font-mono text-[11px] uppercase tracking-[0.3em] text-white/30">
            The Safety System
          </span>
        </motion.div>

        {/* ── The statement — massive, centered, the whole point ── */}
        <motion.h2
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-80px" }}
          transition={{ duration: 1, ease: EASE_OUT }}
          className="text-center font-display text-[clamp(2rem,5.5vw,4rem)] font-medium leading-[1.1] tracking-[-0.04em] text-white mb-6"
        >
          A model can't decide
          <br />
          <span className="text-white/35">if it's safe.</span>
        </motion.h2>

        <motion.p
          initial={{ opacity: 0, y: 16 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-80px" }}
          transition={{ duration: 0.9, delay: 0.15, ease: EASE_OUT }}
          className="text-center font-body-mature text-[clamp(15px,1.6vw,17px)] text-white/40 leading-[1.6] max-w-lg mx-auto mb-24"
        >
          So Condura doesn't let it. Four rules, written in code, that no
          prompt can bend.
        </motion.p>

        {/* ── The rules — constitutional entries, hairline-separated ── */}
        <div className="max-w-2xl mx-auto">
          {RULES.map((rule, i) => (
            <motion.div
              key={rule.title}
              initial={{ opacity: 0, y: 16 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-40px" }}
              transition={{ duration: 0.7, delay: i * 0.08, ease: EASE_OUT }}
              className="border-t border-white/[0.08]"
            >
              <div className="flex items-start gap-5 py-9">
                {/* Quiet ordinal — a small dot + index, nothing loud */}
                <span className="flex items-center gap-2 mt-1 shrink-0 select-none">
                  <span className="w-1.5 h-1.5 rounded-full bg-white/30" />
                  <span className="font-mono text-[12px] text-white/30 tabular-nums">
                    {String(i + 1).padStart(2, "0")}
                  </span>
                </span>

                {/* Title as the lead, then the explanation */}
                <div className="flex-1 min-w-0">
                  <h3 className="font-display text-[clamp(17px,2vw,21px)] font-semibold text-white mb-2.5 tracking-[-0.02em] leading-snug">
                    {rule.title}
                  </h3>
                  <p className="font-body-mature text-[15px] text-white/45 leading-[1.65]">
                    {rule.text}
                  </p>
                </div>
              </div>
            </motion.div>
          ))}
          {/* Closing hairline */}
          <div className="border-t border-white/[0.08]" />
        </div>

        {/* ── Footer — where the gate lives ── */}
        <motion.div
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.3, ease: EASE_OUT }}
          className="mt-16 flex items-center justify-center gap-3"
        >
          <span className="font-mono text-[11px] text-white/25">
            The gatekeeper is pure Go. The policy is a YAML file in your home directory. You own both.
          </span>
        </motion.div>
      </div>
    </section>
  );
}