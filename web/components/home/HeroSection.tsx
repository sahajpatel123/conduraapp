"use client";

import { MouseEvent } from "react";
import { motion, useMotionValue, useSpring, useTransform } from "motion/react";
import HeroDownload from "./HeroDownload";
import OverlayPreview from "./OverlayPreview";

export default function HeroSection() {
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

  return (
    <>
      <section className="relative w-full h-screen min-h-[800px] bg-[#000] flex flex-col lg:flex-row overflow-hidden">

        {/* ── LEFT: Copy ── */}
        <div className="w-full lg:w-1/2 h-full flex flex-col justify-between px-8 lg:pl-20 lg:pr-16 pt-32 pb-12 relative z-20">
          <div className="flex-1 flex flex-col justify-center">
            <motion.div
              initial={{ opacity: 0, y: 30 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 1.2, delay: 0.1, ease: [0.16, 1, 0.3, 1] }}
            >
              <div className="font-mono text-[10px] text-[#71717a] tracking-[0.2em] uppercase mb-10 flex items-center gap-4">
                <span className="w-8 h-[2px] bg-[#D97757]" />
                RELEASE 0.1.0 · OPEN ALPHA
              </div>

              <h1 className="text-[64px] lg:text-[84px] font-semibold leading-[0.95] tracking-tight mb-8">
                <div className="text-white">Condura,</div>
                <div className="text-[#71717a]">on your</div>
                <div className="text-[#71717a]">machine.</div>
              </h1>

              <p className="font-body-mature text-[#a1a1aa] text-[18px] leading-[1.6] mb-12 max-w-md">
                A local-first intelligence layer for your OS. No account, no subscription, and no new workflow to learn.
              </p>

              <HeroDownload />
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
            animate={{ scale: 1, opacity: 1, y: 0 }}
            transition={{ duration: 1.6, delay: 0.4, ease: "easeOut" }}
            className="relative z-10 w-[88%] max-w-[520px]"
          >
            <motion.div
              style={{ rotateX, rotateY, transformStyle: "preserve-3d" }}
            >
              <OverlayPreview active={true} />
            </motion.div>
          </motion.div>
        </div>

      </section>
    </>
  );
}
