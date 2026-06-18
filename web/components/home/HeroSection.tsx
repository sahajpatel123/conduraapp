"use client";

import { useState, useEffect, MouseEvent } from "react";
import { motion, AnimatePresence, useMotionValue, useSpring, useTransform } from "motion/react";
import DownloadDropdown from "./DownloadDropdown";
import NeuralHandshake from "./NeuralHandshake";

/**
 * DESIGN PHILOSOPHY — FINAL VERSION
 * 
 * The right side must feel like a photograph of a real product, not a collage of UI toys.
 * 
 * RULE: ONE element. One single, impeccably crafted terminal window floating over the
 * wallpaper. That's it. No fake docks, no fake menu bars, no fake notification banners,
 * no colored circles pretending to be app icons. Just ONE window. Confidence through restraint.
 * 
 * This is how Apple, Linear, and Raycast present their products: a single, perfectly lit
 * screenshot of the actual application. Nothing else competes for attention.
 */

const STEPS = [
  { input: "condura boot --local",          output: "Gatekeeper mounted. SQLite bus online." },
  { input: "condura spawn --agent=react",   output: "Agent spawned. Analyzing workspace AST..." },
  { input: "react-agent: patch Hero.tsx",   output: "Diff applied cleanly. +12 lines, -4 lines." },
  { input: "condura verify --strict",       output: "All deterministic safety rules passed." },
];

export default function HeroSection() {
  const [introFinished, setIntroFinished] = useState(false);
  const [step, setStep] = useState(0);

  // Subtle 3D tilt on hover
  const mx = useMotionValue(0);
  const my = useMotionValue(0);
  const sX = useSpring(mx, { stiffness: 80, damping: 30 });
  const sY = useSpring(my, { stiffness: 80, damping: 30 });
  const rotateX = useTransform(sY, [-0.5, 0.5], ["3deg", "-3deg"]);
  const rotateY = useTransform(sX, [-0.5, 0.5], ["-3deg", "3deg"]);

  const handleMouseMove = (e: MouseEvent<HTMLDivElement>) => {
    const r = e.currentTarget.getBoundingClientRect();
    mx.set((e.clientX - r.left) / r.width - 0.5);
    my.set((e.clientY - r.top) / r.height - 0.5);
  };
  const handleMouseLeave = () => { mx.set(0); my.set(0); };

  useEffect(() => {
    const t = setTimeout(() => setIntroFinished(true), 1700);
    return () => clearTimeout(t);
  }, []);

  useEffect(() => {
    if (!introFinished) return;
    const interval = setInterval(() => setStep((p) => (p + 1) % STEPS.length), 3500);
    return () => clearInterval(interval);
  }, [introFinished]);

  return (
    <>
      {/* Neural Handshake — opening sequence */}
      <NeuralHandshake />

      <section className="relative w-full h-screen min-h-[800px] bg-[#000] flex flex-col lg:flex-row overflow-hidden">

        {/* ── LEFT: Copy ── */}
        <div className="w-full lg:w-1/2 h-full flex flex-col justify-between px-8 lg:pl-28 lg:pr-16 pt-32 pb-12 relative z-20">
          <div className="flex-1 flex flex-col justify-center">
            <motion.div
              initial={{ opacity: 0, y: 30 }}
              animate={{ opacity: introFinished ? 1 : 0, y: introFinished ? 0 : 30 }}
              transition={{ duration: 1.2, delay: 0.1, ease: [0.16, 1, 0.3, 1] }}
            >
              <div className="font-mono text-[11px] text-[#a1a1aa] tracking-widest uppercase mb-8 flex items-center gap-3">
                <span className="w-8 h-[1px] bg-white/20" />
                V0.1.0 Open Alpha
              </div>

              <h1 className="text-[56px] lg:text-[72px] font-medium leading-[0.95] tracking-[-0.03em] text-[#fff] mb-6">
                Your OS, now <br />
                <span className="text-transparent bg-clip-text bg-gradient-to-r from-white to-[#71717a]">
                  autonomous.
                </span>
              </h1>

              <p className="font-body-mature text-[#a1a1aa] text-[16px] leading-[1.6] mb-12 max-w-md">
                Stop pasting code into a browser tab. Condura orchestrates massive parallel AI workflows directly on your machine. Secure, local, and incredibly fast.
              </p>

              <div className="flex flex-col sm:flex-row items-center gap-4 w-full sm:w-auto">
                <DownloadDropdown />
              </div>
            </motion.div>
          </div>

        </div>

        {/* ── RIGHT: One Perfect Window ── */}
        <div
          className="hidden lg:flex w-1/2 h-full relative items-center justify-center overflow-hidden"
          style={{ perspective: "1200px" }}
          onMouseMove={handleMouseMove}
          onMouseLeave={handleMouseLeave}
        >
          {/* Wallpaper — subtle, warm, ambient */}
          <div
            className="absolute inset-0 bg-cover bg-center opacity-40"
            style={{ backgroundImage: "url('/images/condura-desktop-light.jpg')" }}
          />
          {/* Smooth blended transition from the left panel — no sharp edge */}
          <div className="absolute inset-0 bg-gradient-to-r from-black via-black/70 to-black/30" />
          <div className="absolute inset-0 bg-gradient-to-b from-black/30 via-black/60 to-black/80" />

          {/* The Single Window */}
          <motion.div
            initial={{ scale: 1.06, opacity: 0 }}
            animate={{ scale: introFinished ? 1 : 1.06, opacity: introFinished ? 1 : 0 }}
            transition={{ duration: 2, delay: 0.4, ease: "easeOut" }}
            className="relative z-10 w-[88%] max-w-[580px]"
          >
            <motion.div
              style={{ rotateX, rotateY, transformStyle: "preserve-3d" }}
              className="w-full rounded-xl overflow-hidden shadow-[0_50px_100px_rgba(0,0,0,0.6),0_0_0_1px_rgba(255,255,255,0.06)]"
            >
              {/* Title Bar */}
              <div className="h-[40px] bg-[#1e1e1e] border-b border-white/[0.06] flex items-center px-4 relative">
                <div className="flex items-center gap-2">
                  <div className="w-3 h-3 rounded-full bg-[#ff5f57]" />
                  <div className="w-3 h-3 rounded-full bg-[#febc2e]" />
                  <div className="w-3 h-3 rounded-full bg-[#28c840]" />
                </div>
                <span className="absolute left-1/2 -translate-x-1/2 text-[12px] text-white/30 font-medium">
                  Condura
                </span>
              </div>

              {/* Toolbar */}
              <div className="h-[44px] bg-[#181818] border-b border-white/[0.04] flex items-center justify-between px-5">
                <div className="flex items-center gap-3">
                  <div className="w-6 h-6 rounded-md bg-white/10 flex items-center justify-center">
                    <span className="text-white/70 text-[11px] font-semibold">C</span>
                  </div>
                  <div>
                    <p className="text-[11px] font-medium text-white/70 leading-none">Orchestrator Engine</p>
                    <p className="text-[9px] text-white/25 font-mono mt-[2px] leading-none">condura-core v0.1.0</p>
                  </div>
                </div>
                <div className="flex items-center gap-3">
                  <span className="text-[9px] text-white/20 font-mono border border-white/[0.06] rounded px-1.5 py-0.5">⌘K</span>
                </div>
              </div>

              {/* Terminal Body */}
              <div className="bg-[#0e0e0e] p-6 min-h-[280px] flex flex-col justify-between">
                <div className="space-y-5">
                  <AnimatePresence mode="wait">
                    {STEPS.slice(0, step + 1).map((s, i) => (
                      <motion.div
                        key={`step-${i}`}
                        initial={{ opacity: 0, y: 8 }}
                        animate={{ opacity: i === step ? 1 : 0.4, y: 0 }}
                        transition={{ duration: 0.4 }}
                        className="font-mono text-[12px] leading-relaxed"
                      >
                        <p className="text-white/80">
                          <span className="text-white/50 mr-2">❯</span>
                          {s.input}
                        </p>
                        <p className="text-white/30 mt-1 pl-5">{s.output}</p>
                      </motion.div>
                    ))}
                  </AnimatePresence>
                </div>

                {/* Blinking cursor at bottom */}
                <div className="mt-6 flex items-center gap-2 font-mono text-[12px]">
                  <span className="text-white/50">❯</span>
                  <motion.span
                    animate={{ opacity: [1, 0] }}
                    transition={{ repeat: Infinity, duration: 0.9 }}
                    className="inline-block w-[7px] h-[14px] bg-white/40"
                  />
                </div>
              </div>

              {/* Status Bar */}
              <div className="h-[28px] bg-[#151515] border-t border-white/[0.04] flex items-center justify-between px-5">
                <div className="flex items-center gap-4">
                  <span className="flex items-center gap-1.5 text-[10px] text-white/25 font-mono">
                    <span className="w-1.5 h-1.5 rounded-full bg-white/40" />
                    3 agents active
                  </span>
                  <span className="text-[10px] text-white/15 font-mono">sqlite: locked</span>
                </div>
                <span className="text-[10px] text-white/15 font-mono">gatekeeper: sealed</span>
              </div>
            </motion.div>
          </motion.div>
        </div>

      </section>
    </>
  );
}
