"use client";

import { useRef } from "react";
import { motion, useScroll, useTransform } from "motion/react";

export default function HowItFeels() {
  const containerRef = useRef<HTMLDivElement>(null);
  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ["start end", "end start"],
  });

  const keyOpacity = useTransform(scrollYProgress, [0.0, 0.12, 0.20, 0.28], [0, 1, 1, 0]);
  const keyScale = useTransform(scrollYProgress, [0.0, 0.12], [0.8, 1]);
  const rippleScale = useTransform(scrollYProgress, [0.08, 0.20], [0, 1.5]);
  const rippleOpacity = useTransform(scrollYProgress, [0.08, 0.20], [0.6, 0]);

  const overlayY = useTransform(scrollYProgress, [0.15, 0.35], ["120%", "0%"]);
  const overlayOpacity = useTransform(scrollYProgress, [0.15, 0.25, 0.45, 0.55], [0, 1, 1, 0]);
  const overlayScale = useTransform(scrollYProgress, [0.15, 0.30], [0.85, 1]);

  const textReveal1 = useTransform(scrollYProgress, [0.30, 0.38, 0.45, 0.52], [0, 1, 1, 0]);
  const textReveal2 = useTransform(scrollYProgress, [0.42, 0.50, 0.58, 0.65], [0, 1, 1, 0]);

  return (
    <section ref={containerRef} className="relative overflow-hidden bg-[#050505] py-32">
      <div className="bg-circuit-fine pointer-events-none absolute inset-0 opacity-30" />
      <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/[0.06] to-transparent" />

      <div className="relative z-10 mx-auto max-w-4xl px-6">
        <motion.div
          initial={{ opacity: 0, y: 20 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.3 }}
          transition={{ duration: 0.6 }}
          className="mb-24 text-center"
        >
          <p className="mb-3 text-[13px] font-medium uppercase tracking-widest text-white/30">
            How it feels
          </p>
          <h2 className="gradient-headline text-[32px] font-semibold tracking-tighter sm:text-[44px]">
            One hotkey. Zero friction.
          </h2>
        </motion.div>

        {/* Step 1: The keypress */}
        <div className="flex min-h-[60vh] items-center justify-center">
          <div className="relative flex flex-col items-center">
            <motion.div
              style={{ opacity: keyOpacity, scale: keyScale }}
              className="relative"
            >
              <div className="glass-overlay rounded-xl px-8 py-5 text-[15px] font-mono text-white/70">
                <kbd className="text-[20px] font-bold text-white">⌘ + ⇧ + Space</kbd>
              </div>
              <motion.div
                style={{ scale: rippleScale, opacity: rippleOpacity }}
                className="absolute inset-0 rounded-xl border-2 border-[#0066cc]/40"
              />
            </motion.div>
            <motion.p
              style={{ opacity: keyOpacity }}
              className="mt-6 text-[15px] text-white/40"
            >
              Press your hotkey
            </motion.p>
          </div>
        </div>

        {/* Step 2: Overlay appears */}
        <div className="flex min-h-[60vh] items-center justify-center">
          <div className="relative flex flex-col items-center">
            <motion.div
              style={{ y: overlayY, opacity: overlayOpacity, scale: overlayScale }}
              className="w-full max-w-[480px]"
            >
              <div className="glass-overlay rounded-2xl p-5">
                <div className="flex items-center gap-2 border-b border-white/[0.06] pb-3">
                  <div className="h-2 w-2 rounded-full bg-[#ff5f57]" />
                  <div className="h-2 w-2 rounded-full bg-[#febc2e]" />
                  <div className="h-2 w-2 rounded-full bg-[#28c840]" />
                </div>
                <div className="mt-4 flex items-start gap-2.5">
                  <div className="animate-orb h-3 w-3 shrink-0 rounded-full bg-[#0066cc]" />
                  <div className="flex-1 space-y-3">
                    <div className="rounded-xl bg-white/[0.04] px-3.5 py-2.5">
                      <span className="text-[13px] text-white/70">
                        Hey Condura.
                      </span>
                    </div>
                    <motion.div style={{ opacity: textReveal1 }}>
                      <div className="rounded-xl bg-[#0066cc]/10 px-3.5 py-2.5">
                        <span className="text-[13px] text-[#64c8ff]">
                          Listening...
                        </span>
                      </div>
                    </motion.div>
                  </div>
                </div>
              </div>
            </motion.div>
            <motion.p
              style={{ opacity: overlayOpacity }}
              className="mt-6 text-[15px] text-white/40"
            >
              The overlay appears — glass, glowing, alive
            </motion.p>
          </div>
        </div>

        {/* Step 3: Response streams in */}
        <div className="flex min-h-[60vh] items-center justify-center">
          <div className="relative flex flex-col items-center">
            <motion.div
              style={{ opacity: textReveal2 }}
              className="w-full max-w-[480px]"
            >
              <div className="glass-overlay rounded-2xl p-5">
                <div className="flex items-center gap-2 border-b border-white/[0.06] pb-3">
                  <div className="h-2 w-2 rounded-full bg-[#ff5f57]" />
                  <div className="h-2 w-2 rounded-full bg-[#febc2e]" />
                  <div className="h-2 w-2 rounded-full bg-[#28c840]" />
                </div>
                <div className="mt-4 space-y-2.5">
                  <div className="flex items-start gap-2.5">
                    <div className="animate-orb h-3 w-3 shrink-0 rounded-full bg-[#0066cc]" />
                    <div className="rounded-xl bg-white/[0.04] px-3.5 py-2.5 text-[13px] text-white/70">
                      Write a haiku about AI
                    </div>
                  </div>
                  <div className="flex items-start gap-2.5 pl-6">
                    <div className="rounded-xl bg-[#0066cc]/10 px-3.5 py-2.5 text-[13px] leading-relaxed text-[#64c8ff]">
                      Ghost in the machine wakes —<br />
                      One hotkey, one thought, one deed.<br />
                      Dawn breaks in silence.
                    </div>
                  </div>
                </div>

                <div className="mt-3 flex items-center gap-1.5 pl-6">
                  <span className="h-1 w-1 animate-[typing-dot_1.4s_infinite_-0.32s] rounded-full bg-white/30" />
                  <span className="h-1 w-1 animate-[typing-dot_1.4s_infinite_-0.16s] rounded-full bg-white/30" />
                  <span className="h-1 w-1 animate-[typing-dot_1.4s_infinite_0s] rounded-full bg-white/30" />
                </div>
              </div>
            </motion.div>
            <motion.p
              style={{ opacity: textReveal2 }}
              className="mt-6 text-[15px] text-white/40"
            >
              It thinks. It speaks. It acts.
            </motion.p>
          </div>
        </div>
      </div>

      <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/[0.06] to-transparent" />
    </section>
  );
}
