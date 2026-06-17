"use client";

import { motion } from "motion/react";

const PARTICLE_CONFIGS = Array.from({ length: 8 }, (_, i) => ({
  id: i,
  left: `${10 + Math.random() * 80}%`,
  delay: Math.random() * 8,
  duration: 8 + Math.random() * 6,
  size: 1 + Math.random() * 2,
  xOffset: Math.sin(i) * 30,
}));

export default function Stats() {
  return (
    <section className="relative overflow-hidden bg-[#050505] py-32">
      <div className="bg-circuit-fine pointer-events-none absolute inset-0 opacity-30" />
      <div className="absolute top-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/[0.06] to-transparent" />

      {PARTICLE_CONFIGS.map((p) => (
        <motion.div
          key={p.id}
          className="pointer-events-none absolute rounded-full bg-white/[0.15]"
          style={{
            left: p.left,
            width: p.size,
            height: p.size,
            bottom: "-10%",
          }}
          animate={{
            y: [0, -600],
            x: [0, p.xOffset],
            opacity: [0, 0.6, 0.6, 0],
          }}
          transition={{
            duration: p.duration,
            repeat: Infinity,
            delay: p.delay,
            ease: "linear",
          }}
        />
      ))}

      <div className="relative z-10 mx-auto max-w-4xl px-6 text-center">
        <motion.div
          initial={{ opacity: 0, y: 24 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true, amount: 0.4 }}
          transition={{ duration: 0.8, ease: [0.22, 1, 0.36, 1] as [number, number, number, number] }}
        >
          <blockquote className="text-balance text-[28px] font-semibold leading-snug tracking-tighter text-white sm:text-[36px] md:text-[44px]">
            <span className="text-white/20">&ldquo;</span>
            Every action recorded. Every decision auditable. Nothing hidden. Nothing lost.
            <span className="text-white/20">&rdquo;</span>
          </blockquote>

          <div className="mt-16 grid gap-8 sm:grid-cols-3">
            {[
              { value: "100%", label: "Local-first", desc: "Your data never leaves your machine" },
              { value: "100%", label: "Free forever", desc: "No subscription. No tier gates. No limits." },
              { value: "100%", label: "Yours", desc: "Open source core. Auditable. Tamper-resistant." },
            ].map((stat, i) => (
              <motion.div
                key={stat.label}
                initial={{ opacity: 0, y: 16 }}
                whileInView={{ opacity: 1, y: 0 }}
                viewport={{ once: true, amount: 0.5 }}
                transition={{ duration: 0.5, delay: i * 0.12 }}
                className="text-center"
              >
                <div className="text-glow text-[48px] font-bold tracking-tighter text-white sm:text-[56px]">
                  {stat.value}
                </div>
                <div className="mt-1 text-[15px] font-medium text-white/60">
                  {stat.label}
                </div>
                <div className="mt-1 text-[13px] text-white/30">{stat.desc}</div>
              </motion.div>
            ))}
          </div>
        </motion.div>
      </div>

      <div className="absolute bottom-0 left-0 right-0 h-px bg-gradient-to-r from-transparent via-white/[0.06] to-transparent" />
    </section>
  );
}
