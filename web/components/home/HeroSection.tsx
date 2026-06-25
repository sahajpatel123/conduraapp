"use client";

import { motion } from "motion/react";
import HeroDownload from "./HeroDownload";
import OverlayPreview from "./OverlayPreview";

export default function HeroSection() {
  return (
    <section className="relative w-full h-full flex flex-col items-center justify-between overflow-hidden pt-24 pb-12">
      {/* Background Image with Cinematic Overlay */}
      <div 
        className="absolute inset-0 bg-cover bg-center bg-no-repeat opacity-50 mix-blend-screen scale-105"
        style={{ backgroundImage: "url('/hero-bg.png')" }}
      />
      <div className="absolute inset-0 bg-gradient-to-b from-[#000000]/90 via-[#000000]/60 to-[#000000]/90" />
      <div className="absolute inset-0 bg-[radial-gradient(ellipse_at_center,rgba(0,223,216,0.06)_0%,rgba(0,0,0,0.8)_100%)]" />

      {/* Main Content Container - Text Block */}
      <div className="relative z-20 w-full max-w-[1200px] mx-auto px-4 flex flex-col items-center text-center mt-2 sm:mt-6 shrink-0">
        
        {/* Massive Headline */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 1, delay: 0.1, ease: [0.16, 1, 0.3, 1] }}
          className="w-full max-w-[900px] mb-3"
        >
          <h1 className="font-hero-display text-[42px] sm:text-[60px] md:text-[76px] lg:text-[88px] font-extrabold leading-[1.05] tracking-tighter">
            <span className="text-white block">One hotkey.</span>
            <span className="bg-clip-text text-transparent bg-gradient-to-r from-[#00DFD8] to-[#007CF0]">Every AI you own.</span>
          </h1>
        </motion.div>

        {/* Subtitle */}
        <motion.p
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 1, delay: 0.2, ease: [0.16, 1, 0.3, 1] }}
          className="font-lead-airy text-[15px] sm:text-[18px] md:text-[20px] text-white/60 max-w-[600px] mb-6"
        >
          The ultimate orchestration layer for your local environment. A free desktop app that summons every AI tool instantly.
        </motion.p>

        {/* Actions */}
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          animate={{ opacity: 1, y: 0 }}
          transition={{ duration: 1, delay: 0.3, ease: [0.16, 1, 0.3, 1] }}
          className="flex flex-col items-center w-full"
        >
          <div className="scale-[0.80] sm:scale-90 origin-top">
            <HeroDownload />
          </div>
        </motion.div>
      </div>

      {/* Product Preview in Center - Bottom section */}
      <motion.div
        initial={{ opacity: 0, y: 40, scale: 0.95 }}
        animate={{ opacity: 1, y: 0, scale: 1 }}
        transition={{ duration: 1.2, delay: 0.4, ease: [0.16, 1, 0.3, 1] }}
        className="w-full max-w-[600px] lg:max-w-[700px] relative z-10 flex-1 flex flex-col justify-end min-h-0"
      >
        {/* Subtle glow behind the preview */}
        <div className="absolute inset-0 top-1/4 bg-[#00DFD8]/10 blur-[80px] rounded-full pointer-events-none" />
        
        <div className="relative transform-gpu hover:scale-[1.02] transition-transform duration-700 ease-out shadow-[0_0_0_1px_rgba(255,255,255,0.08),0_20px_60px_rgba(0,0,0,0.8)] rounded-2xl bg-black/40 backdrop-blur-3xl p-1 mb-2">
          <OverlayPreview active={true} />
        </div>
      </motion.div>

    </section>
  );
}
