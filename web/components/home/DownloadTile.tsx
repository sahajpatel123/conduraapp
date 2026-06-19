"use client";

import { motion } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/* ────────────────────────────────────────────────────────────
   DownloadTile — the closing CTA.

   The old version duplicated the dedicated /download page inline:
   an OS segmented control, an installer card, primary/secondary
   links, a modal with install steps. That's redundant when a real
   download page exists one click away.

   This replacement is a single quiet closing statement — the
   editorial register the rest of the site now uses. A short
   headline, one line, and a primary button that routes to
   /download where the real experience lives. Nothing else
   competes. The page ends the way it began: with restraint.
   ──────────────────────────────────────────────────────────── */

export default function DownloadTile() {
  return (
    <section
      id="download-tile"
      className="relative w-full bg-black border-t border-white/[0.08] py-[160px] px-6 overflow-hidden"
    >
      <div className="mx-auto w-full max-w-3xl flex flex-col items-center text-center">
        {/* Eyebrow */}
        <motion.div
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true, margin: "-80px" }}
          transition={{ duration: 0.8, ease: EASE_OUT }}
          className="flex items-center gap-2.5 mb-10"
        >
          <span className="h-px w-8 bg-white/20" />
          <span className="font-mono text-[11px] uppercase tracking-[0.25em] text-white/35">
            Ready when you are
          </span>
        </motion.div>

        {/* Headline */}
        <motion.h2
          initial={{ opacity: 0, y: 18 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-80px" }}
          transition={{ duration: 0.9, ease: EASE_OUT }}
          className="font-display text-[clamp(1.9rem,4vw,3.1rem)] font-semibold leading-[1.08] tracking-[-0.035em] text-white mb-6"
        >
          Your machine.
          <br />
          <span className="text-white/40">Your conductor.</span>
        </motion.h2>

        {/* One line */}
        <motion.p
          initial={{ opacity: 0, y: 14 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-80px" }}
          transition={{ duration: 0.8, delay: 0.12, ease: EASE_OUT }}
          className="font-body-mature text-[16px] text-white/45 leading-[1.6] max-w-md mb-12"
        >
          Free forever. Runs on your machine. Uses the AI you already have.
          One download, three platforms, no account.
        </motion.p>

        {/* Single CTA — routes to the real download page */}
        <motion.div
          initial={{ opacity: 0, y: 14 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, margin: "-80px" }}
          transition={{ duration: 0.8, delay: 0.22, ease: EASE_OUT }}
        >
          <a
            href="/download"
            className="glass-download inline-flex items-center justify-center gap-2.5 px-10 py-4 font-body-mature text-[15px] font-semibold"
          >
            <span>Get Condura</span>
            <svg width="15" height="15" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-white/70">
              <path d="M5 12h14M13 6l6 6-6 6" />
            </svg>
          </a>
        </motion.div>

        {/* Quiet footer — the honest detail */}
        <motion.p
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          transition={{ duration: 0.8, delay: 0.35, ease: EASE_OUT }}
          className="mt-10 font-mono text-[11px] text-white/25"
        >
          v0.1.0 · macOS · Windows · Linux · Free for personal &amp; commercial use
        </motion.p>
      </div>
    </section>
  );
}