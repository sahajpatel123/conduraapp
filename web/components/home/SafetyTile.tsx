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
     · Four rules laid out as a legal document — Roman numerals,
       a title, and a single sentence, separated by hairlines
     · A quiet footer line stating where the gate lives

   This is how Apple's privacy page and the best editorial sites
   handle trust: the words are the design.
   ──────────────────────────────────────────────────────────── */

const RULES = [
  {
    numeral: "I",
    title: "Separation",
    text: "The model decides what to do. Deterministic code decides whether it's safe. They are never the same system.",
  },
  {
    numeral: "II",
    title: "Classification",
    text: "Every action is sorted by blast radius — read, write, network, destructive — before it is allowed to run.",
  },
  {
    numeral: "III",
    title: "Consent",
    text: "Network and destructive actions require a real human at the keyboard. No exceptions. No overrides. No 'trust me.'",
  },
  {
    numeral: "IV",
    title: "Record",
    text: "Every decision — allowed or denied — is written to a tamper-proof, HMAC-chained audit log that never deletes.",
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
              key={rule.numeral}
              initial={{ opacity: 0, y: 16 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true, margin: "-40px" }}
              transition={{ duration: 0.7, delay: i * 0.08, ease: EASE_OUT }}
              className="border-t border-white/[0.08]"
            >
              <div className="flex items-start gap-6 py-8 group">
                {/* Roman numeral — large, muted, tabular */}
                <span className="font-display text-[clamp(28px,3vw,36px)] font-light text-white/20 tabular-nums leading-none mt-1 select-none w-12 shrink-0">
                  {rule.numeral}
                </span>

                {/* Title + text */}
                <div className="flex-1 min-w-0">
                  <h3 className="font-body-mature text-[15px] font-semibold text-white/85 mb-2 tracking-[-0.01em]">
                    {rule.title}
                  </h3>
                  <p className="font-body-mature text-[15px] text-white/45 leading-[1.6]">
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