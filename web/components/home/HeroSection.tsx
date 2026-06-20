"use client";

import { MouseEvent } from "react";
import { motion, useMotionValue, useSpring, useTransform } from "motion/react";
import HeroDownload from "./HeroDownload";
import OverlayPreview from "./OverlayPreview";
import { SITE } from "@/lib/site";

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
        <div className="w-full lg:w-1/2 h-full flex flex-col justify-between px-8 lg:pl-32 xl:pl-48 lg:pr-16 pt-32 pb-12 relative z-20">
          <div className="flex-1 flex flex-col justify-center max-w-[540px]">
            <motion.div
              initial={{ opacity: 0, y: 30 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 1.2, delay: 0.1, ease: [0.16, 1, 0.3, 1] }}
            >
              <div className="font-mono text-[10px] text-[#71717a] tracking-[0.2em] uppercase mb-10 flex items-center gap-4">
                <span className="w-8 h-[2px] bg-[#D97757]" />
                RELEASE 0.1.0 · PUBLIC ALPHA
              </div>

              <h1 className="text-[64px] lg:text-[84px] font-semibold leading-[0.95] tracking-tight mb-8">
                <div className="text-white">One hotkey.</div>
                <div className="text-[#71717a]">Every AI</div>
                <div className="text-[#71717a]">you own.</div>
              </h1>

              <p className="font-body-mature text-[#a1a1aa] text-[18px] leading-[1.6] mb-12 max-w-md">
                A free desktop app that summons every AI tool on your machine with one hotkey. No account, no subscription, no data leaves your computer.
              </p>

              <HeroDownload />

              <div className="mt-8 flex items-center gap-6 font-mono text-[11px] text-[#a1a1aa]">
                <a
                  href={SITE.github}
                  target="_blank"
                  rel="noopener noreferrer"
                  className="inline-flex items-center gap-1.5 hover:text-white transition-colors"
                >
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="currentColor"><path d="M12 0C5.37 0 0 5.37 0 12c0 5.31 3.435 9.795 8.205 11.385.6.105.825-.255.825-.57 0-.285-.015-1.23-.015-2.235-3.015.555-3.795-.735-4.035-1.41-.135-.345-.72-1.41-1.23-1.695-.42-.225-1.02-.78-.015-.795.945-.015 1.62.87 1.845 1.23 1.08 1.815 2.805 1.305 3.495.99.105-.78.42-1.305.765-1.605-2.67-.3-5.46-1.335-5.46-5.925 0-1.305.465-2.385 1.23-3.225-.12-.3-.54-1.53.12-3.18 0 0 1.005-.315 3.3 1.23.96-.27 1.98-.405 3-.405s2.04.135 3 .405c2.295-1.56 3.3-1.23 3.3-1.23.66 1.65.24 2.88.12 3.18.765.84 1.23 1.905 1.23 3.225 0 4.605-2.805 5.625-5.475 5.925.435.375.81 1.095.81 2.22 0 1.605-.015 2.895-.015 3.3 0 .315.225.69.825.57A12.02 12.02 0 0024 12c0-6.63-5.37-12-12-12z"/></svg>
                  {SITE.github.replace("https://", "")}
                </a>
                <span className="w-px h-3 bg-white/[0.12]" />
                <span>Free & open source</span>
                <span className="w-px h-3 bg-white/[0.12]" />
                <span>No account</span>
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
