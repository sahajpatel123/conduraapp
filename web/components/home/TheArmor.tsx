"use client";

import { useState } from "react";
import { motion } from "motion/react";
import Reveal from "@/components/motion/Reveal";
import WordReveal from "@/components/motion/WordReveal";
import { INVARIANTS } from "@/lib/site";
import { EASE_OUT } from "@/lib/motion";

/**
 * TheArmor — the seven non-negotiable invariants, presented as an
 * editorial ledger. Hovering a row draws a protective synapse thread
 * around its numeral and reveals its body. The section closes with a
 * single ink panel that states the survival rule.
 */
export default function TheArmor() {
  const [open, setOpen] = useState<number | null>(0);

  return (
    <section data-section="armor" className="relative mx-auto max-w-[1180px] px-6 py-28 sm:py-36">
      <Reveal>
        <p className="text-eyebrow mb-4">— The armor</p>
      </Reveal>
      <WordReveal as="h2" text="A feature without armor under it is the wrong feature." className="text-display text-[var(--color-ink)] max-w-[16ch] text-balance" />

      <div className="mt-12 grid gap-10 md:grid-cols-[1.3fr_1fr] md:gap-16">
        {/* ── Invariants ledger ── */}
        <div className="divide-y divide-[rgba(20,17,11,0.12)] border-y border-[rgba(20,17,11,0.12)]">
          {INVARIANTS.map((inv, i) => {
            const isOpen = open === i;
            return (
              <button
                key={inv.numeral}
                onMouseEnter={() => setOpen(i)}
                onClick={() => setOpen(isOpen ? null : i)}
                className="group flex w-full items-start gap-5 py-5 text-left"
              >
                {/* numeral with protective thread */}
                <div className="relative grid h-12 w-12 shrink-0 place-items-center">
                  <svg viewBox="0 0 48 48" className="absolute inset-0" aria-hidden>
                    <motion.rect
                      x="3"
                      y="3"
                      width="42"
                      height="42"
                      rx="12"
                      fill="none"
                      stroke="var(--color-synapse)"
                      strokeWidth="1.2"
                      animate={{
                        pathLength: isOpen ? 1 : 0,
                        opacity: isOpen ? 1 : 0,
                      }}
                      transition={{ duration: 0.6, ease: EASE_OUT }}
                      style={{ pathLength: 1 }}
                    />
                  </svg>
                  <span
                    className={`font-display text-[20px] leading-none transition-colors ${
                      isOpen ? "text-[var(--color-ink)]" : "text-[var(--color-ink-mute)]"
                    }`}
                  >
                    {inv.numeral}
                  </span>
                </div>
                <div className="flex-1">
                  <h3
                    className={`font-display text-[20px] leading-tight transition-colors ${
                      isOpen ? "text-[var(--color-ink)]" : "text-[var(--color-ink-soft)]"
                    }`}
                  >
                    {inv.title}
                  </h3>
                  <motion.div
                    initial={false}
                    animate={{ height: isOpen ? "auto" : 0, opacity: isOpen ? 1 : 0 }}
                    transition={{ duration: 0.4, ease: EASE_OUT }}
                    style={{ overflow: "hidden" }}
                  >
                    <p className="text-body mt-2 max-w-[52ch] text-[var(--color-ink-mute)]">
                      {inv.body}
                    </p>
                  </motion.div>
                </div>
                <span
                  className={`mt-1 font-mono text-[12px] transition-all ${
                    isOpen ? "text-[var(--color-synapse)] rotate-0" : "text-[var(--color-ink-faint)] -rotate-45"
                  }`}
                >
                  →
                </span>
              </button>
            );
          })}
        </div>

        {/* ── The survival rule (ink panel) ── */}
        <Reveal delay={0.1}>
          <div className="surface-ink sticky top-28 p-8 sm:p-10">
            <p className="text-mono-label text-on-ink-eyebrow">The survival rule</p>
            <p className="font-display mt-4 text-[28px] leading-tight text-on-ink-headline text-balance">
              This is not an optimization problem.{" "}
              <span className="italic text-on-ink-emphasis">It is a survival problem.</span>
            </p>
            <p className="text-on-ink-body mt-5">
              Condura performs physical, often irreversible actions on your
              operating system. A fallible multi-model system, async-supervised,
              operating with stale screen state. Every design decision bends to
              that fact.
            </p>
            <div className="rule-ink-on-ink my-7" />
            <ul className="space-y-3">
              {[
                "HMAC-chained audit log of every action",
                "Twin-snapshot verification before every click",
                "4-layer kill switch you control",
                "Deterministic Gatekeeper — never a model",
              ].map((x) => (
                <li key={x} className="flex items-start gap-3 text-on-ink-list">
                  <span className="text-on-ink-bullet mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full" />
                  <span>{x}</span>
                </li>
              ))}
            </ul>
            <a
              href="/security"
              className="text-on-ink-link mt-7 inline-flex items-center gap-2 text-[14px] font-medium"
            >
              Read the full security model
              <span aria-hidden>→</span>
            </a>
          </div>
        </Reveal>
      </div>
    </section>
  );
}
