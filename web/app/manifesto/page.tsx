"use client";

import { motion, useScroll, useTransform } from "motion/react";
import { useRef, useState, useEffect } from "react";
import AnimatedBadge from "@/components/motion/AnimatedBadge";
import { Icon } from "@/components/motion/Icon";
import { INVARIANTS } from "@/lib/site";
import { EASE_OUT } from "@/lib/motion";

/* ────────────────────────────────────────────────────────────
   MANIFESTO — Why Condura Exists
   A scroll-driven meditation on the seven non-negotiable
   invariants. The "eye" of Condura rotates slowly in the
   background as you move through each principle.
   ──────────────────────────────────────────────────────────── */

export default function ManifestoPage() {
  return (
    <main className="relative w-full bg-black text-white">
      <ManifestoHero />
      <TheProblemSection />
      <InvariantsScrollSection />
      <ThePromiseSection />
      <ClosingCTA />
    </main>
  );
}

/* ════════════════════════════════════════════════════════════
   1. HERO
   ════════════════════════════════════════════════════════════ */

function ManifestoHero() {
  const [mounted, setMounted] = useState(false);
  useEffect(() => { const t = setTimeout(() => setMounted(true), 200); return () => clearTimeout(t); }, []);

  return (
    <section className="relative min-h-screen flex flex-col items-center justify-center px-6 overflow-hidden">
      {/* The "Eye" of Condura — rotating background */}
      <div className="fixed inset-0 flex items-center justify-center opacity-30 pointer-events-none">
        <motion.div
          animate={{ rotate: 360 }}
          transition={{ duration: 100, repeat: Infinity, ease: "linear" }}
          className="w-[800px] h-[800px] rounded-full border border-white/[0.06] relative flex items-center justify-center"
        >
          <motion.div
            animate={{ rotate: -360 }}
            transition={{ duration: 60, repeat: Infinity, ease: "linear" }}
            className="w-[600px] h-[600px] rounded-full border border-dashed border-white/[0.08]"
          />
          <div className="absolute w-[400px] h-[400px] rounded-full bg-white/[0.02] blur-3xl" />
        </motion.div>
      </div>

      <div className="relative z-10 max-w-4xl text-center">
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          animate={{ opacity: mounted ? 1 : 0, y: mounted ? 0 : 30 }}
          transition={{ duration: 1, ease: EASE_OUT }}
        >
          <div className="mb-8 flex justify-center">
            <AnimatedBadge tone="neutral" pulse>Manifesto</AnimatedBadge>
          </div>

          <h1 className="font-display text-[clamp(2.5rem,7vw,5.5rem)] font-semibold leading-[1.02] tracking-[-0.04em]">
            Your computer should work
            <br />
            <span className="text-transparent bg-clip-text bg-gradient-to-r from-white via-white to-white/30">
              for you alone.
            </span>
          </h1>

          <p className="mt-10 mx-auto max-w-2xl font-lead-airy">
            Artificial intelligence is becoming how we use our machines. That shift is too
            important to hand to systems that watch everything you do, route every thought through
            a cloud, and charge you for the privilege of your own data.
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: mounted ? 1 : 0 }}
          transition={{ delay: 1, duration: 1 }}
          className="absolute bottom-12 left-1/2 -translate-x-1/2 flex flex-col items-center gap-2"
        >
          <span className="font-mono text-[10px] uppercase tracking-widest text-white/25">Scroll to explore</span>
          <div className="w-[1px] h-12 bg-gradient-to-b from-white/30 to-transparent" />
        </motion.div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   2. THE PROBLEM — Why This Matters
   ════════════════════════════════════════════════════════════ */

function TheProblemSection() {
  const problems = [
    {
      title: "They watch everything.",
      desc: "Cloud AI agents see your screen, your files, your keystrokes. Your data trains their models. You are the product, again.",
    },
    {
      title: "They cost you forever.",
      desc: "Subscriptions that never end. Per-token charges that scale with your ambition. The more you use it, the more you pay.",
    },
    {
      title: "They don't talk to each other.",
      desc: "Claude can't see what Codex did. ChatGPT can't call your local model. Every tool is an island, and you are the ferry.",
    },
    {
      title: "They can't be stopped.",
      desc: "Once an autonomous agent starts, you're a passenger. No kill switch. No audit. No proof of what it did or why.",
    },
  ];

  return (
    <section className="relative w-full py-[160px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-20 max-w-3xl">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">The Problem</span>
          <h2 className="mt-4 font-display text-[clamp(2rem,5vw,3.5rem)] font-semibold tracking-[-0.03em] leading-[1.1]">
            The tools are amazing.
            <br />
            <span className="text-white/40">The deal is broken.</span>
          </h2>
        </div>

        <div className="grid md:grid-cols-2 gap-6">
          {problems.map((p, i) => (
            <motion.div
              key={p.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.1 }}
              className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-8"
            >
              <div className="flex items-start gap-4">
                <span className="font-mono text-[14px] text-white/30 mt-1">0{i + 1}</span>
                <div>
                  <h3 className="font-body-mature text-[18px] font-semibold text-white">{p.title}</h3>
                  <p className="mt-3 font-body-mature text-[15px] text-white/45 leading-relaxed">{p.desc}</p>
                </div>
              </div>
            </motion.div>
          ))}
        </div>

        {/* Transition statement */}
        <motion.div
          initial={{ opacity: 0 }}
          whileInView={{ opacity: 1 }}
          viewport={{ once: true }}
          className="mt-20 text-center"
        >
          <p className="font-display text-[clamp(1.5rem,4vw,2.5rem)] font-medium tracking-[-0.02em] text-white/70 leading-snug max-w-3xl mx-auto">
            Condura is the opposite deal.
            <br />
            <span className="text-white/30">Free. Local. Yours.</span>
          </p>
        </motion.div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   3. INVARIANTS — Horizontal Scroll Gallery
   Vertical scroll drives a horizontal track of 7 full-screen
   invariant cards. Each card is a self-contained "panel"
   with its numeral, title, and body.
   ════════════════════════════════════════════════════════════ */

function InvariantsScrollSection() {
  const containerRef = useRef<HTMLDivElement>(null);
  const n = INVARIANTS.length;
  const { scrollYProgress } = useScroll({
    target: containerRef,
    offset: ["start start", "end end"],
  });

  // The sticky element (h-screen) pins from progress 0 until the container's
  // bottom passes the sticky's bottom — that happens at progress (n-1)/n.
  // We must complete the horizontal travel by that exact point, otherwise the
  // last panels slide up into blank space while the sticky unpins.
  const pinRange = (n - 1) / n;
  const x = useTransform(scrollYProgress, [0, pinRange, 1], ["0vw", `-${(n - 1) * 100}vw`, `-${(n - 1) * 100}vw`]);

  // Active index for the progress dots
  const [activeIndex, setActiveIndex] = useState(0);
  useEffect(() => {
    return scrollYProgress.on("change", (v) => {
      const norm = Math.min(1, Math.max(0, v / pinRange));
      setActiveIndex(Math.min(n - 1, Math.max(0, Math.round(norm * (n - 1)))));
    });
  }, [scrollYProgress, pinRange, n]);

  return (
    <section ref={containerRef} className="relative" style={{ height: `${n * 100}vh` }}>
      {/* Sticky horizontal viewport */}
      <div className="sticky top-0 h-screen overflow-hidden flex flex-col">
        {/* Section label */}
        <div className="pt-24 pb-4 text-center">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">
            The Seven Invariants · scroll to advance →
          </span>
        </div>

        {/* Horizontal track */}
        <div className="flex-1 flex items-center">
          <motion.div style={{ x }} className="flex">
            {INVARIANTS.map((inv, idx) => (
              <div
                key={inv.numeral}
                className="w-screen h-full flex items-center justify-center shrink-0 px-6"
              >
                <div className="flex flex-col md:flex-row items-center gap-8 md:gap-20 max-w-5xl w-full">
                  {/* Left: Giant numeral */}
                  <div className="relative flex items-center justify-center">
                    {/* Pulse ring */}
                    <motion.div
                      animate={{ scale: [1, 1.4], opacity: [0.3, 0] }}
                      transition={{ duration: 3, repeat: Infinity, delay: idx * 0.2 }}
                      className="absolute w-56 h-56 md:w-80 md:h-80 rounded-full border border-white/15"
                    />
                    {/* Glow */}
                    <div className="absolute w-40 h-40 md:w-60 md:h-60 rounded-full bg-white/[0.03] blur-3xl" />
                    {/* Numeral circle */}
                    <div className="relative w-48 h-48 md:w-72 md:h-72 rounded-full border border-white/20 bg-white/[0.02] backdrop-blur-md flex items-center justify-center overflow-hidden">
                      <div className="absolute inset-0 bg-gradient-to-br from-white/10 to-transparent opacity-50" />
                      <span className="font-mono text-6xl md:text-8xl font-light text-white/80 tracking-tighter">
                        {inv.numeral}
                      </span>
                    </div>
                  </div>

                  {/* Right: Text */}
                  <div className="flex-1 max-w-lg">
                    <span className="font-mono text-[11px] uppercase tracking-widest text-white/30 mb-4 block">
                      Invariant {inv.numeral} of {n}
                    </span>
                    <h3 className="font-display text-[clamp(1.5rem,4vw,3rem)] font-semibold tracking-[-0.02em] text-white leading-[1.15] mb-6">
                      {inv.title}
                    </h3>
                    <p className="font-lead-airy text-white/50">
                      {inv.body}
                    </p>
                  </div>
                </div>
              </div>
            ))}
          </motion.div>
        </div>

        {/* Progress dots */}
        <div className="pb-10 flex justify-center gap-2">
          {INVARIANTS.map((_, idx) => (
            <div
              key={idx}
              className="h-1 rounded-full transition-all duration-500"
              style={{
                width: activeIndex === idx ? "32px" : "8px",
                background: activeIndex === idx ? "rgba(255,255,255,0.5)" : "rgba(255,255,255,0.1)",
              }}
            />
          ))}
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   4. THE PROMISE — What Condura Commits To
   ════════════════════════════════════════════════════════════ */

function ThePromiseSection() {
  const promises = [
    { title: "Free forever", desc: "No feature gates. No premium tier. No nags. A donate button in the menu bar — that's it." },
    { title: "Local-first", desc: "Memory, skills, audit log, embeddings — all on disk, encrypted. The only network calls are to your LLM provider." },
    { title: "Your keys, your models", desc: "API keys live encrypted on your device. Never logged, never sent anywhere except the provider you configured." },
    { title: "Open ecosystem", desc: "12+ LLM providers. 8 CLI sub-agents. Any local model. No vendor lock-in, ever." },
    { title: "Auditable by design", desc: "Every action logged in an HMAC-chained, append-only trail. You can prove what happened." },
    { title: "Yours to leave", desc: "Uninstall auto-backs-up your data. No cloud account to cancel. No data to delete from someone else's server." },
  ];

  return (
    <section className="relative w-full py-[160px] px-6 border-t border-white/[0.08]">
      <div className="mx-auto max-w-5xl">
        <div className="mb-20 max-w-3xl">
          <span className="font-mono text-[11px] uppercase tracking-widest text-white/30">The Promise</span>
          <h2 className="mt-4 font-display text-[clamp(2rem,5vw,3.5rem)] font-semibold tracking-[-0.03em] leading-[1.1]">
            What we commit to.
          </h2>
          <p className="mt-6 font-lead-airy">
            These aren&apos;t features. They&apos;re constraints. If we ever break one, the product
            is wrong — not the constraint.
          </p>
        </div>

        <div className="grid md:grid-cols-2 lg:grid-cols-3 gap-6">
          {promises.map((p, i) => (
            <motion.div
              key={p.title}
              initial={{ opacity: 0, y: 20 }}
              whileInView={{ opacity: 1, y: 0 }}
              viewport={{ once: true }}
              transition={{ delay: i * 0.08 }}
              className="rounded-2xl border border-white/[0.08] bg-white/[0.02] p-6"
            >
              <div className="flex h-8 w-8 items-center justify-center rounded-full border border-white/15 bg-white/[0.04] mb-4">
                <Icon name="check" size={16} className="text-white/60" />
              </div>
              <h3 className="font-body-mature text-[16px] font-semibold text-white">{p.title}</h3>
              <p className="mt-2 font-body-mature text-[14px] text-white/45 leading-relaxed">{p.desc}</p>
            </motion.div>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ════════════════════════════════════════════════════════════
   5. CLOSING CTA
   ════════════════════════════════════════════════════════════ */

function ClosingCTA() {
  return (
    <section className="relative w-full py-[200px] px-6 border-t border-white/[0.08] overflow-hidden">
      <div className="absolute inset-0 flex items-center justify-center pointer-events-none">
        <motion.div
          animate={{ scale: [1, 1.15, 1], opacity: [0.05, 0.1, 0.05] }}
          transition={{ duration: 5, repeat: Infinity }}
          className="w-[600px] h-[300px] rounded-full bg-white blur-[150px]"
        />
      </div>

      <div className="relative z-10 mx-auto max-w-3xl text-center">
        <motion.div
          initial={{ opacity: 0, y: 30 }}
          whileInView={{ opacity: 1, y: 0 }}
          viewport={{ once: true }}
          transition={{ duration: 1, ease: EASE_OUT }}
        >
          <p className="font-display text-[clamp(1.75rem,5vw,3rem)] font-medium tracking-[-0.03em] leading-[1.15] text-white/80">
            Make AI useful to every ordinary person,
            <br />
            <span className="text-white/40">on every computer, for free.</span>
          </p>
          <p className="mt-8 font-lead-airy mx-auto max-w-lg">
            No lock-in. No tracking. No compromise on speed or safety. That&apos;s the whole mission.
          </p>
          <div className="mt-12 flex flex-col sm:flex-row items-center justify-center gap-4">
            <a
              href="/download"
              className="mature-button inline-flex items-center gap-2 px-8 py-4 font-body-mature text-[15px] font-semibold"
            >
              Download v0.1.0 →
            </a>
            <a
              href="/security"
              className="mature-button-secondary inline-flex items-center gap-2 px-6 py-4 font-body-mature text-[14px]"
            >
              How it stays safe
            </a>
          </div>
        </motion.div>
      </div>
    </section>
  );
}
