"use client";

import PageChrome from "@/components/shell/PageChrome";
import { motion, useScroll, useTransform, useSpring } from "motion/react";
import { useRef } from "react";
import AnimatedBadge from "@/components/motion/AnimatedBadge";

export default function OrchestrationPage() {
  const containerRef = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ["start start", "end end"],
  });

  const smoothProgress = useSpring(scrollYProgress, { damping: 20, stiffness: 100 });
  
  const y1 = useTransform(smoothProgress, [0, 1], [0, -200]);
  const y2 = useTransform(smoothProgress, [0, 1], [0, -400]);
  const y3 = useTransform(smoothProgress, [0, 1], [0, -600]);

  const opacity1 = useTransform(smoothProgress, [0, 0.3, 0.4], [0, 1, 0.2]);
  const opacity2 = useTransform(smoothProgress, [0.3, 0.6, 0.7], [0, 1, 0.2]);
  const opacity3 = useTransform(smoothProgress, [0.6, 0.9, 1], [0, 1, 1]);

  return (
    <PageChrome
      eyebrow="Engine"
      title="Massive parallel workflows."
      description="Condura doesn't just run agents sequentially. It spins up highly concurrent, local swarms that communicate through a fast SQLite event bus."
      badge="Orchestration"
    >
      <div ref={containerRef} className="relative mt-24 h-[300vh]">
        {/* Sticky viewport for the animations */}
        <div className="sticky top-32 flex h-[80vh] items-center justify-center overflow-hidden rounded-[32px] border border-white/10 bg-[#050505] shadow-[0_0_80px_rgba(255,255,255,0.03)]">
          
          {/* Abstract Grid Background */}
          <div className="absolute inset-0 bg-grid-dark opacity-30" />
          
          {/* Phase 1: Planning */}
          <motion.div 
            style={{ y: y1, opacity: opacity1 }}
            className="absolute inset-0 flex flex-col items-center justify-center p-8"
          >
            <div className="mb-8 rounded-2xl border border-white/10 bg-white/[0.02] p-6 backdrop-blur-xl">
              <h3 className="font-mono text-[11px] uppercase tracking-widest text-white/40 mb-2">Phase 1</h3>
              <p className="text-2xl font-medium text-white">Decomposition</p>
            </div>
            <div className="flex gap-4">
              {[...Array(4)].map((_, i) => (
                <motion.div
                  key={i}
                  initial={{ opacity: 0, scale: 0.8 }}
                  whileInView={{ opacity: 1, scale: 1 }}
                  transition={{ delay: i * 0.1, type: "spring" }}
                  className="w-16 h-16 rounded-xl border border-white/20 bg-white/5 flex items-center justify-center shadow-[0_0_20px_rgba(255,255,255,0.05)]"
                >
                  <span className="font-mono text-white/50 text-[10px]">T-{i+1}</span>
                </motion.div>
              ))}
            </div>
          </motion.div>

          {/* Phase 2: Fan Out */}
          <motion.div 
            style={{ y: y2, opacity: opacity2 }}
            className="absolute inset-0 flex flex-col items-center justify-center p-8"
          >
            <div className="mb-8 rounded-2xl border border-white/10 bg-white/[0.02] p-6 backdrop-blur-xl">
              <h3 className="font-mono text-[11px] uppercase tracking-widest text-white/40 mb-2">Phase 2</h3>
              <p className="text-2xl font-medium text-white">Parallel Fan-Out</p>
            </div>
            <div className="relative w-full max-w-lg h-64 border border-white/10 rounded-3xl bg-black/50 p-6 overflow-hidden">
              <div className="absolute inset-0 flex items-center justify-center">
                <div className="w-16 h-16 rounded-full bg-white/10 blur-xl animate-pulse" />
              </div>
              <div className="flex flex-col gap-3 relative z-10">
                {[
                  "Spawning react-agent...",
                  "Spawning rust-agent...",
                  "Mounting DOM analyzer...",
                  "Starting headless browser..."
                ].map((text, i) => (
                  <motion.div
                    key={i}
                    initial={{ x: -20, opacity: 0 }}
                    whileInView={{ x: 0, opacity: 1 }}
                    transition={{ delay: i * 0.15 }}
                    className="font-mono text-[12px] text-white/60 flex items-center gap-3"
                  >
                    <span className="w-2 h-2 rounded-full bg-white/30" />
                    {text}
                  </motion.div>
                ))}
              </div>
            </div>
          </motion.div>

          {/* Phase 3: Resolution */}
          <motion.div 
            style={{ y: y3, opacity: opacity3 }}
            className="absolute inset-0 flex flex-col items-center justify-center p-8"
          >
             <div className="mb-8 rounded-2xl border border-white/10 bg-white/[0.02] p-6 backdrop-blur-xl text-center">
              <h3 className="font-mono text-[11px] uppercase tracking-widest text-white/40 mb-2">Phase 3</h3>
              <p className="text-2xl font-medium text-white">Deterministic Resolution</p>
            </div>
            
            <motion.div 
              className="w-32 h-32 rounded-[2rem] border-2 border-white/20 bg-white/10 flex items-center justify-center shadow-[0_0_60px_rgba(255,255,255,0.1)]"
              animate={{ rotate: 360 }}
              transition={{ duration: 10, repeat: Infinity, ease: "linear" }}
            >
              <div className="w-16 h-16 rounded-full border border-white/30 bg-white/5" />
            </motion.div>
            
            <p className="mt-8 font-mono text-[12px] text-white/40">Diffs merged. AST verified. Lockfile updated.</p>
          </motion.div>

        </div>
      </div>
    </PageChrome>
  );
}
