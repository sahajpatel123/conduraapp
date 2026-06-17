"use client";

import PageChrome from "@/components/shell/PageChrome";
import { motion, AnimatePresence } from "motion/react";
import { useState } from "react";
import AnimatedBadge from "@/components/motion/AnimatedBadge";

export default function SecurityPage() {
  const [isHovered, setIsHovered] = useState(false);

  return (
    <PageChrome
      eyebrow="Security Gatekeeper"
      title="Capability without the blank check."
      description="Models hallucinate, prompts get injected, and some actions cannot be undone. Condura draws a hard line between thinking and acting with a deterministic permission gatekeeper."
      badge="Zero Trust"
    >
      <div className="mt-24 h-[600px] w-full rounded-[32px] border border-white/10 bg-[#050505] relative overflow-hidden flex items-center justify-center p-8">
        
        {/* Background glow that pulses based on state */}
        <motion.div 
          animate={{ 
            opacity: isHovered ? 0.4 : 0.1,
            scale: isHovered ? 1.2 : 1
          }}
          transition={{ duration: 1 }}
          className="absolute inset-0 bg-[radial-gradient(circle_at_center,rgba(255,255,255,0.2),transparent_50%)]"
        />

        {/* The Vault Mechanism */}
        <div 
          className="relative z-10 w-[400px] h-[400px] rounded-full border-4 border-white/5 flex items-center justify-center cursor-pointer"
          onMouseEnter={() => setIsHovered(true)}
          onMouseLeave={() => setIsHovered(false)}
        >
          {/* Outer rotating ring */}
          <motion.div 
            animate={{ rotate: isHovered ? 180 : 0 }}
            transition={{ type: "spring", stiffness: 50, damping: 20 }}
            className="absolute inset-0 rounded-full border-2 border-dashed border-white/20"
          />
          
          {/* Middle rotating ring */}
          <motion.div 
            animate={{ rotate: isHovered ? -90 : 0 }}
            transition={{ type: "spring", stiffness: 60, damping: 25 }}
            className="absolute inset-8 rounded-full border border-white/10 flex items-center justify-center"
          >
            {/* Locking pins */}
            {[0, 90, 180, 270].map((deg) => (
              <motion.div 
                key={deg}
                style={{ rotate: deg }}
                className="absolute w-full h-[2px] flex justify-between"
              >
                <motion.div 
                  animate={{ scaleX: isHovered ? 0 : 1 }}
                  transition={{ duration: 0.3 }}
                  className="w-4 h-full bg-white/40 origin-left"
                />
                <motion.div 
                  animate={{ scaleX: isHovered ? 0 : 1 }}
                  transition={{ duration: 0.3 }}
                  className="w-4 h-full bg-white/40 origin-right"
                />
              </motion.div>
            ))}
          </motion.div>

          {/* Core Vault */}
          <motion.div 
            animate={{ 
              scale: isHovered ? 1.05 : 1,
              boxShadow: isHovered 
                ? "0 0 60px rgba(255,255,255,0.1), inset 0 0 20px rgba(255,255,255,0.05)" 
                : "0 0 0px rgba(255,255,255,0), inset 0 0 0px rgba(255,255,255,0)"
            }}
            transition={{ duration: 0.5 }}
            className="relative w-48 h-48 rounded-full bg-black border border-white/20 flex flex-col items-center justify-center z-20"
          >
            <motion.div
              animate={{ y: isHovered ? -5 : 0 }}
              className="font-mono text-[11px] text-white/40 uppercase tracking-widest mb-2"
            >
              Status
            </motion.div>
            <motion.div
              animate={{ 
                color: isHovered ? "rgba(255,255,255,1)" : "rgba(255,255,255,0.4)" 
              }}
              className="text-xl font-medium tracking-tight"
            >
              {isHovered ? "UNLOCKED" : "LOCKED"}
            </motion.div>
            
            {/* Permission badges that appear when unlocked */}
            <AnimatePresence>
              {isHovered && (
                <motion.div 
                  initial={{ opacity: 0, y: 10 }}
                  animate={{ opacity: 1, y: 0 }}
                  exit={{ opacity: 0, y: -10 }}
                  className="absolute -bottom-16 flex gap-2"
                >
                  <AnimatedBadge tone="neutral">Read FS</AnimatedBadge>
                  <AnimatedBadge tone="neutral">Port 3000</AnimatedBadge>
                </motion.div>
              )}
            </AnimatePresence>
          </motion.div>
        </div>

      </div>

      <div className="mt-16 grid md:grid-cols-3 gap-6">
        {[
          { title: "Deterministic Rules", desc: "Hard-coded boundaries that cannot be overridden by prompt injection." },
          { title: "Audit Trail", desc: "Every API call, file edit, and terminal command logged locally in SQLite." },
          { title: "Air-gapped Memory", desc: "Vector store lives entirely on your device. Never sent to the cloud." }
        ].map((feat, i) => (
          <motion.div 
            key={i}
            initial={{ opacity: 0, y: 20 }}
            whileInView={{ opacity: 1, y: 0 }}
            transition={{ delay: i * 0.1 }}
            className="p-6 rounded-2xl border border-white/10 bg-white/[0.02]"
          >
            <h4 className="text-white font-medium mb-3">{feat.title}</h4>
            <p className="text-white/50 text-[14px] leading-relaxed">{feat.desc}</p>
          </motion.div>
        ))}
      </div>
    </PageChrome>
  );
}
