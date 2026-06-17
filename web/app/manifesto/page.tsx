"use client";

import { motion, useScroll, useTransform } from "motion/react";
import { INVARIANTS, SITE } from "@/lib/site";
import { useRef, useState } from "react";
import AnimatedBadge from "@/components/motion/AnimatedBadge";

export default function ManifestoPage() {
  const containerRef = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ["start start", "end end"],
  });

  const [activeInvariant, setActiveInvariant] = useState(0);

  return (
    <main className="relative min-h-[300vh] bg-black text-white" ref={containerRef}>
      
      {/* ── Background Graphic: The "Eye" of Condura ── */}
      <div className="fixed inset-0 flex items-center justify-center opacity-30 pointer-events-none">
        <motion.div 
          animate={{ rotate: 360 }}
          transition={{ duration: 100, repeat: Infinity, ease: "linear" }}
          className="w-[800px] h-[800px] rounded-full border border-white/5 relative flex items-center justify-center"
        >
          <motion.div 
            animate={{ rotate: -360 }}
            transition={{ duration: 60, repeat: Infinity, ease: "linear" }}
            className="w-[600px] h-[600px] rounded-full border border-white/10 border-dashed"
          />
          <div className="absolute w-[400px] h-[400px] rounded-full bg-white/[0.02] blur-3xl" />
        </motion.div>
      </div>

      {/* ── Section 1: Intro ── */}
      <div className="h-screen flex items-center justify-center relative z-10 px-8">
        <div className="max-w-3xl text-center">
          <motion.div 
            initial={{ opacity: 0, y: 30 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ duration: 1, ease: "easeOut" }}
          >
            <div className="mb-6 flex justify-center">
              <AnimatedBadge tone="neutral">Manifesto</AnimatedBadge>
            </div>
            <h1 className="text-5xl md:text-7xl font-semibold tracking-tight mb-8">
              Your computer should work for <span className="text-white/40">you alone.</span>
            </h1>
            <p className="text-lg md:text-xl text-white/50 leading-relaxed max-w-2xl mx-auto">
              Artificial intelligence is becoming how we use our machines. That shift is too important to hand to systems that watch everything you do.
            </p>
          </motion.div>
          
          <motion.div 
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 1, duration: 1 }}
            className="absolute bottom-12 left-1/2 -translate-x-1/2 flex flex-col items-center gap-2"
          >
            <span className="text-[10px] uppercase tracking-widest text-white/30 font-mono">Scroll to explore</span>
            <div className="w-[1px] h-12 bg-gradient-to-b from-white/30 to-transparent" />
          </motion.div>
        </div>
      </div>

      {/* ── Section 2: The Invariants (Sticky Scroll) ── */}
      <div className="h-[200vh] relative z-10">
        <div className="sticky top-0 h-screen flex flex-col md:flex-row items-center justify-center gap-12 px-8 max-w-6xl mx-auto">
          
          {/* Left: The Visualizer */}
          <div className="w-full md:w-1/2 h-[400px] flex items-center justify-center relative">
            {INVARIANTS.map((inv, idx) => (
              <motion.div
                key={inv.numeral}
                initial={{ opacity: 0, scale: 0.8 }}
                animate={{ 
                  opacity: activeInvariant === idx ? 1 : 0,
                  scale: activeInvariant === idx ? 1 : 0.8,
                  zIndex: activeInvariant === idx ? 10 : 0
                }}
                transition={{ duration: 0.6 }}
                className="absolute inset-0 flex items-center justify-center"
              >
                <div className="w-64 h-64 rounded-full border border-white/20 bg-white/[0.02] backdrop-blur-md flex items-center justify-center relative overflow-hidden">
                  <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent opacity-50" />
                  <span className="font-mono text-6xl font-light text-white/80 tracking-tighter">
                    {inv.numeral}
                  </span>
                </div>
              </motion.div>
            ))}
          </div>

          {/* Right: The Text */}
          <div className="w-full md:w-1/2 relative h-[400px]">
            {INVARIANTS.map((inv, idx) => (
              <motion.div
                key={inv.title}
                onViewportEnter={() => setActiveInvariant(idx)}
                viewport={{ margin: "-40% 0px -40% 0px" }}
                className="absolute inset-0 flex flex-col justify-center"
                initial={{ opacity: 0, x: 20 }}
                animate={{ 
                  opacity: activeInvariant === idx ? 1 : 0,
                  x: activeInvariant === idx ? 0 : 20,
                  pointerEvents: activeInvariant === idx ? "auto" : "none"
                }}
                transition={{ duration: 0.5 }}
              >
                <h3 className="text-2xl md:text-4xl font-semibold mb-4 text-white">
                  {inv.title}
                </h3>
                <p className="text-white/50 text-lg leading-relaxed">
                  {inv.body}
                </p>
              </motion.div>
            ))}
            
            {/* Invisible scroll triggers */}
            <div className="absolute inset-0 overflow-y-auto hidden-scrollbar pointer-events-auto snap-y snap-mandatory h-[400px]">
              {INVARIANTS.map((inv, idx) => (
                <div 
                  key={`trigger-${idx}`} 
                  className="h-full w-full snap-center"
                  onMouseEnter={() => setActiveInvariant(idx)}
                />
              ))}
            </div>
          </div>

        </div>
      </div>

    </main>
  );
}
