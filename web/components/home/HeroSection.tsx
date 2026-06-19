"use client";

import { useState, useEffect, MouseEvent } from "react";
import { motion, useMotionValue, useSpring, useTransform } from "motion/react";
import DownloadDropdown from "./DownloadDropdown";
import NeuralHandshake from "./NeuralHandshake";
import OverlayPreview from "./OverlayPreview";

/**
 * DESIGN PHILOSOPHY — FINAL VERSION
 *
 * The right side shows the actual product UI a user will see — a floating
 * chat overlay with a real conversation cycling through relatable examples —
 * so a first-time visitor instantly understands: press a hotkey, this window
 * appears, ask anything, it uses the AI you already have, and it's safe.
 * Below the chat, four plain-word badges state what makes it different.
 * Minimal. Honest. No glow, no aurora — just the product, shown clearly.
 */

export default function HeroSection() {
  const [introFinished, setIntroFinished] = useState(false);

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

          {/* The overlay preview — the actual product UI */}
          <motion.div
            initial={{ scale: 1.04, opacity: 0, y: 16 }}
            animate={{ scale: introFinished ? 1 : 1.04, opacity: introFinished ? 1 : 0, y: introFinished ? 0 : 16 }}
            transition={{ duration: 1.6, delay: 0.4, ease: "easeOut" }}
            className="relative z-10 w-[88%] max-w-[520px]"
          >
            <motion.div
              style={{ rotateX, rotateY, transformStyle: "preserve-3d" }}
            >
              <OverlayPreview active={introFinished} />
            </motion.div>
          </motion.div>
        </div>

      </section>
    </>
  );
}
