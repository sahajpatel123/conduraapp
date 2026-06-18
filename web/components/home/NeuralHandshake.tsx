"use client";

import { useEffect, useState } from "react";
import { motion, AnimatePresence } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/**
 * NeuralHandshake — the 1.6s opening sequence.
 *
 * A single point of light appears at the center of a black stage and
 * rapidly draws six radiating synaptic arcs. Each arc terminates in a
 * node that pulses alive. The nodes link to each other, forming a
 * constellation. The constellation contracts into a single luminous
 * mark — the "synapse" — which holds for a beat, then dissolves into
 * particles as the hero fades up beneath it.
 *
 * The sequence is driven by a single progress clock (0 → 1) so every
 * layer stays in sync. No loose timers, no jank. Reduced-motion users
 * see a calm one-step fade instead.
 */

const ARCS = Array.from({ length: 6 }, (_, i) => {
  const angle = (i / 6) * Math.PI * 2 - Math.PI / 2;
  const radius = 120;
  return {
    id: i,
    x: Math.cos(angle) * radius,
    y: Math.sin(angle) * radius,
  };
});

export default function NeuralHandshake() {
  const [done, setDone] = useState(false);
  const [reduced, setReduced] = useState(false);

  useEffect(() => {
    const mq = window.matchMedia("(prefers-reduced-motion: reduce)");
    setReduced(mq.matches);
    const duration = mq.matches ? 600 : 1600;
    const t = setTimeout(() => setDone(true), duration);
    return () => clearTimeout(t);
  }, []);

  return (
    <AnimatePresence>
      {!done && (
        <motion.div
          key="neural-handshake"
          initial={{ opacity: 1 }}
          exit={{ opacity: 0 }}
          transition={{ duration: reduced ? 0.4 : 0.6, ease: "easeInOut" }}
          className="fixed inset-0 z-[100] bg-black flex items-center justify-center overflow-hidden"
        >
          {reduced ? (
            <motion.div
              initial={{ opacity: 0, scale: 0.9 }}
              animate={{ opacity: 1, scale: 1 }}
              transition={{ duration: 0.4, ease: EASE_OUT }}
              className="relative"
            >
              <SynapseMark className="w-16 h-16 text-white/80" />
            </motion.div>
          ) : (
            <FullSequence />
          )}
        </motion.div>
      )}
    </AnimatePresence>
  );
}

/* ── The full motion sequence ── */

function FullSequence() {
  return (
    <div className="relative w-[300px] h-[300px] flex items-center justify-center">
      {/* Ambient bloom behind the constellation */}
      <motion.div
        initial={{ opacity: 0, scale: 0.4 }}
        animate={{ opacity: [0, 0.25, 0.18, 0.3, 0], scale: [0.4, 1, 0.9, 1.1, 1.4] }}
        transition={{ duration: 1.6, ease: "easeInOut", times: [0, 0.3, 0.5, 0.75, 1] }}
        className="absolute w-[260px] h-[260px] rounded-full bg-white blur-[90px]"
      />

      {/* SVG stage — arcs + nodes */}
      <motion.svg
        viewBox="-150 -150 300 300"
        className="absolute inset-0 w-full h-full"
        initial={{ opacity: 1 }}
        animate={{ opacity: [0, 1, 1, 0] }}
        transition={{ duration: 1.6, ease: "easeInOut", times: [0, 0.15, 0.75, 1] }}
      >
        {/* Six radiating arcs drawn from center outward */}
        {ARCS.map((arc, i) => (
          <motion.line
            key={`arc-${i}`}
            x1={0}
            y1={0}
            x2={arc.x}
            y2={arc.y}
            stroke="url(#arcGradient)"
            strokeWidth={1}
            strokeLinecap="round"
            initial={{ pathLength: 0, opacity: 0 }}
            animate={{ pathLength: [0, 1], opacity: [0, 0.7, 0.4] }}
            transition={{ duration: 0.5, delay: 0.05 * i, ease: EASE_OUT }}
          />
        ))}

        {/* Cross-links between adjacent nodes — the constellation mesh */}
        {ARCS.map((arc, i) => {
          const next = ARCS[(i + 1) % ARCS.length];
          return (
            <motion.line
              key={`link-${i}`}
              x1={arc.x}
              y1={arc.y}
              x2={next.x}
              y2={next.y}
              stroke="rgba(255,255,255,0.18)"
              strokeWidth={0.75}
              initial={{ pathLength: 0, opacity: 0 }}
              animate={{ pathLength: [0, 1], opacity: [0, 0.5] }}
              transition={{ duration: 0.4, delay: 0.4 + 0.04 * i, ease: EASE_OUT }}
            />
          );
        })}

        {/* Node pulses at each arc terminus */}
        {ARCS.map((arc, i) => (
          <motion.circle
            key={`node-${i}`}
            cx={arc.x}
            cy={arc.y}
            r={3}
            fill="white"
            initial={{ scale: 0, opacity: 0 }}
            animate={{ scale: [0, 1.4, 1], opacity: [0, 1, 0.85] }}
            transition={{ duration: 0.5, delay: 0.35 + 0.04 * i, ease: EASE_OUT }}
          />
        ))}

        {/* Center origin point */}
        <motion.circle
          cx={0}
          cy={0}
          r={2.5}
          fill="white"
          initial={{ scale: 0, opacity: 0 }}
          animate={{ scale: [0, 1, 1], opacity: [0, 1, 1] }}
          transition={{ duration: 0.3, ease: EASE_OUT }}
        />

        <defs>
          <linearGradient id="arcGradient" x1="0" y1="0" x2="1" y2="0">
            <stop offset="0%" stopColor="rgba(255,255,255,0.9)" />
            <stop offset="100%" stopColor="rgba(255,255,255,0.2)" />
          </linearGradient>
        </defs>
      </motion.svg>

      {/* The resolved synapse mark — appears as the constellation fades */}
      <motion.div
        initial={{ opacity: 0, scale: 0.6 }}
        animate={{ opacity: [0, 0, 1, 1, 0], scale: [0.6, 0.6, 1, 1, 1.25] }}
        transition={{ duration: 1.6, ease: "easeInOut", times: [0, 0.55, 0.7, 0.85, 1] }}
        className="relative"
      >
        <SynapseMark className="w-12 h-12 text-white" />
      </motion.div>

      {/* Wordmark reveal — brief, then gone with the rest */}
      <motion.div
        initial={{ opacity: 0, y: 6 }}
        animate={{ opacity: [0, 0, 1, 1, 0], y: [6, 6, 0, 0, -4] }}
        transition={{ duration: 1.6, ease: "easeInOut", times: [0, 0.6, 0.72, 0.85, 1] }}
        className="absolute bottom-[-48px] font-mono text-[10px] tracking-[0.35em] uppercase text-white/40"
      >
        Condura
      </motion.div>
    </div>
  );
}

/* ── The synapse mark — two nodes joined by an arc, the product glyph ── */

function SynapseMark({ className = "" }: { className?: string }) {
  return (
    <svg viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth={1.5} strokeLinecap="round" strokeLinejoin="round" className={className}>
      <circle cx="6" cy="12" r="2.2" fill="currentColor" stroke="none" />
      <circle cx="18" cy="12" r="2.2" fill="currentColor" stroke="none" />
      <path d="M8.2 12c2.8-3.2 4.8-3.2 7.6 0" opacity={0.9} />
      <path d="M8.2 12c2.8 3.2 4.8 3.2 7.6 0" opacity={0.5} />
    </svg>
  );
}