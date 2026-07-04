"use client";

import { motion, useScroll, useTransform } from "motion/react";
import { useEffect, useRef, useState } from "react";
import Link from "next/link";
import GlobalNav from "@/components/shell/GlobalNav";
import Reveal from "@/components/motion/Reveal";
import { INVARIANTS } from "@/lib/site";
import { EASE_OUT } from "@/lib/motion";

export default function ManifestoPageClient() {
  return (
    <main className="relative w-full">
      <GlobalNav />
      <ManifestoHero />
      <TheProblemSection />
      <InvariantsScrollSection />
      <ThePromiseSection />
      <ClosingCTA />
    </main>
  );
}

/* ════════════ 1. HERO ════════════ */
function ManifestoHero() {
  const eyeRef = useRef<SVGGElement | null>(null);

  useEffect(() => {
    const prefersReduced = window.matchMedia("(prefers-reduced-motion: reduce)").matches;
    if (prefersReduced) return;
    const g = eyeRef.current;
    if (!g) return;
    let raf = 0;
    let t = 0;
    const loop = () => {
      t += 0.003;
      g.style.transform = `rotate(${t}rad)`;
      raf = requestAnimationFrame(loop);
    };
    raf = requestAnimationFrame(loop);
    return () => cancelAnimationFrame(raf);
  }, []);

  return (
    <section className="relative flex min-h-[100svh] flex-col items-center justify-center overflow-hidden px-6">
      {/* the slowly rotating "eye" — a synapse knot */}
      <div className="pointer-events-none absolute inset-0 flex items-center justify-center opacity-50">
        <svg viewBox="0 0 800 800" className="h-[900px] w-[900px] max-w-[120vw]" aria-hidden>
          <g ref={eyeRef} style={{ transformOrigin: "400px 400px" }}>
            <circle cx="400" cy="400" r="380" fill="none" stroke="var(--color-synapse)" strokeWidth="0.5" opacity="0.25" />
            <circle cx="400" cy="400" r="300" fill="none" stroke="var(--color-synapse)" strokeWidth="0.5" opacity="0.2" strokeDasharray="4 8" />
            <circle cx="400" cy="400" r="220" fill="none" stroke="var(--color-synapse)" strokeWidth="0.5" opacity="0.18" />
            {[...Array(12)].map((_, i) => {
              const a = (i / 12) * Math.PI * 2;
              return (
                <line
                  key={i}
                  x1={400 + Math.cos(a) * 220}
                  y1={400 + Math.sin(a) * 220}
                  x2={400 + Math.cos(a) * 380}
                  y2={400 + Math.sin(a) * 380}
                  stroke="var(--color-synapse)"
                  strokeWidth="0.4"
                  opacity="0.3"
                />
              );
            })}
            <circle cx="400" cy="400" r="6" className="synapse-node" />
          </g>
        </svg>
      </div>

      <div className="relative z-10 max-w-4xl text-center">
        <motion.div initial={{ opacity: 0, y: 30 }} animate={{ opacity: 1, y: 0 }} transition={{ duration: 1, ease: EASE_OUT }}>
          <p className="text-eyebrow mb-8">— Manifesto</p>
          <h1 className="text-hero text-[var(--color-ink)] text-balance">
            Your computer should work
            <br />
            <span className="italic text-[var(--color-synapse)]">for you alone.</span>
          </h1>
          <p className="text-lead mt-10 mx-auto max-w-2xl text-[var(--color-ink-soft)] text-pretty">
            Artificial intelligence is becoming how we use our machines. That shift is too important to hand to systems that watch everything you do, route every thought through a cloud, and charge you for the privilege of your own data.
          </p>
        </motion.div>

        <motion.div
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          transition={{ delay: 1, duration: 1 }}
          className="absolute bottom-10 left-1/2 flex -translate-x-1/2 flex-col items-center gap-2"
        >
          <span className="text-mono-label">Scroll to explore</span>
          <span className="relative h-12 w-px overflow-hidden bg-[rgba(20,17,11,0.15)]">
            <span className="absolute inset-x-0 top-0 h-3 animate-[thread-draw_1.8s_var(--thread-ease)_infinite] bg-[var(--color-synapse)]" />
          </span>
        </motion.div>
      </div>
    </section>
  );
}

/* ════════════ 2. THE PROBLEM ════════════ */
function TheProblemSection() {
  const problems = [
    { title: "They watch everything.", desc: "Cloud AI agents see your screen, your files, your keystrokes. Your data trains their models. You are the product, again." },
    { title: "They cost you forever.", desc: "Subscriptions that never end. Per-token charges that scale with your ambition. The more you use it, the more you pay." },
    { title: "They don't talk to each other.", desc: "Claude can't see what Codex did. ChatGPT can't call your local model. Every tool is an island, and you are the ferry." },
    { title: "They can't be stopped.", desc: "Once an autonomous agent starts, you're a passenger. No kill switch. No audit. No proof of what it did or why." },
  ];
  return (
    <section className="relative w-full border-t border-[rgba(20,17,11,0.12)] px-6 py-32 sm:py-40">
      <div className="mx-auto max-w-5xl">
        <Reveal>
          <p className="text-eyebrow mb-4">— The problem</p>
          <h2 className="text-display text-[var(--color-ink)] max-w-[18ch] text-balance">
            The tools are amazing.
            <br />
            <span className="text-[var(--color-ink-mute)]">The deal is broken.</span>
          </h2>
        </Reveal>

        <div className="mt-14 grid gap-5 md:grid-cols-2">
          {problems.map((p, i) => (
            <Reveal key={p.title} delay={i * 0.08}>
              <div className="surface-card h-full p-7">
                <div className="flex items-start gap-4">
                  <span className="mt-1 font-mono text-[13px] text-[var(--color-synapse)]">0{i + 1}</span>
                  <div>
                    <h3 className="font-display text-[20px] leading-tight text-[var(--color-ink)]">{p.title}</h3>
                    <p className="mt-2.5 text-body text-[var(--color-ink-mute)]">{p.desc}</p>
                  </div>
                </div>
              </div>
            </Reveal>
          ))}
        </div>

        <Reveal delay={0.2}>
          <div className="mt-16 text-center">
            <p className="font-display text-[clamp(24px,4vw,40px)] leading-snug tracking-[-0.02em] text-[var(--color-ink-soft)] max-w-3xl mx-auto text-balance">
              Condura is the opposite deal.
              <br />
              <span className="italic text-[var(--color-synapse)]">Free. Local. Yours.</span>
            </p>
          </div>
        </Reveal>
      </div>
    </section>
  );
}

/* ════════════ 3. INVARIANTS — horizontal scroll gallery ════════════ */
function InvariantsScrollSection() {
  const containerRef = useRef<HTMLDivElement>(null);
  const n = INVARIANTS.length;
  const { scrollYProgress } = useScroll({ target: containerRef, offset: ["start start", "end end"] });
  const pinRange = (n - 1) / n;
  const x = useTransform(scrollYProgress, [0, pinRange, 1], ["0vw", `-${(n - 1) * 100}vw`, `-${(n - 1) * 100}vw`]);

  const [activeIndex, setActiveIndex] = useState(0);
  useEffect(() => {
    return scrollYProgress.on("change", (v) => {
      const norm = Math.min(1, Math.max(0, v / pinRange));
      setActiveIndex(Math.min(n - 1, Math.max(0, Math.round(norm * (n - 1)))));
    });
  }, [scrollYProgress, pinRange, n]);

  return (
    <section ref={containerRef} className="relative bg-[var(--color-paper-warm)]" style={{ height: `${n * 100}vh` }}>
      <div className="sticky top-0 flex h-screen flex-col overflow-hidden">
        <div className="pb-4 pt-24 text-center">
          <span className="text-mono-label">The Seven Invariants · scroll to advance →</span>
        </div>
        <div className="flex flex-1 items-center">
          <motion.div style={{ x }} className="flex">
            {INVARIANTS.map((inv, idx) => (
              <div key={inv.numeral} className="flex h-full w-screen shrink-0 items-center justify-center px-6">
                <div className="flex max-w-5xl w-full flex-col items-center gap-8 md:flex-row md:gap-20">
                  {/* numeral with protective thread draw */}
                  <div className="relative flex items-center justify-center">
                    <motion.div
                      animate={{ scale: [1, 1.25], opacity: [0.3, 0] }}
                      transition={{ duration: 3, repeat: Infinity, delay: idx * 0.2 }}
                      className="absolute h-56 w-56 rounded-full border border-[var(--color-synapse)] md:h-80 md:w-80"
                    />
                    <div className="absolute h-40 w-40 rounded-full bg-[rgba(11,61,46,0.05)] blur-3xl md:h-60 md:w-60" />
                    <div className="relative flex h-48 w-48 items-center justify-center overflow-hidden rounded-full border border-[rgba(20,17,11,0.18)] bg-[var(--color-paper)] md:h-72 md:w-72">
                      <svg className="absolute inset-0" viewBox="0 0 100 100" aria-hidden>
                        <motion.rect
                          x="6" y="6" width="88" height="88" rx="44"
                          fill="none" stroke="var(--color-synapse)" strokeWidth="0.8"
                          initial={{ pathLength: 0 }}
                          whileInView={{ pathLength: 1 }}
                          viewport={{ once: true }}
                          transition={{ duration: 1.4, ease: EASE_OUT, delay: idx * 0.1 }}
                        />
                      </svg>
                      <span className="font-display text-[64px] leading-none text-[var(--color-ink)] md:text-[104px]">
                        {inv.numeral}
                      </span>
                    </div>
                  </div>
                  {/* text */}
                  <div className="max-w-lg flex-1">
                    <span className="text-mono-label mb-4 block">Invariant {inv.numeral} of {n}</span>
                    <h3 className="font-display text-[clamp(26px,4vw,44px)] leading-[1.12] tracking-[-0.02em] text-[var(--color-ink)] mb-6 text-balance">
                      {inv.title}
                    </h3>
                    <p className="text-lead text-[var(--color-ink-soft)] text-pretty">{inv.body}</p>
                  </div>
                </div>
              </div>
            ))}
          </motion.div>
        </div>
        <div className="flex justify-center gap-2 pb-10">
          {INVARIANTS.map((_, idx) => (
            <div
              key={idx}
              className="h-1 rounded-full transition-all duration-500"
              style={{
                width: activeIndex === idx ? "32px" : "8px",
                background: activeIndex === idx ? "var(--color-ink)" : "rgba(20,17,11,0.18)",
              }}
            />
          ))}
        </div>
      </div>
    </section>
  );
}

/* ════════════ 4. THE PROMISE ════════════ */
function ThePromiseSection() {
  const promises = [
    { title: "Free forever", desc: "No feature gates. No premium tier. No nags. A donate button in the menu bar — that's it." },
    { title: "Local-first", desc: "Memory, skills, and the audit log — all on disk, encrypted. The only network calls are to your LLM provider. (Vector embeddings arrive in v0.2.0.)" },
    { title: "Your keys, your models", desc: "API keys live encrypted on your device. Never logged, never sent anywhere except the provider you configured." },
    { title: "Open ecosystem", desc: "12+ LLM providers. 8 CLI sub-agents. Any local model. No vendor lock-in, ever." },
    { title: "Auditable by design", desc: "Every action logged in an HMAC-chained, append-only trail. You can prove what happened." },
    { title: "Yours to leave", desc: "Uninstall auto-backs-up your data. No cloud account to cancel. No data to delete from someone else's server." },
  ];
  return (
    <section className="relative w-full border-t border-[rgba(20,17,11,0.12)] px-6 py-32 sm:py-40">
      <div className="mx-auto max-w-5xl">
        <Reveal>
          <p className="text-eyebrow mb-4">— The promise</p>
          <h2 className="text-display text-[var(--color-ink)] max-w-[14ch] text-balance">What we commit to.</h2>
          <p className="text-lead mt-5 max-w-[52ch] text-[var(--color-ink-soft)] text-pretty">
            These aren&apos;t features. They&apos;re constraints. If we ever break one, the product is wrong — not the constraint.
          </p>
        </Reveal>

        <div className="mt-14 grid gap-5 md:grid-cols-2 lg:grid-cols-3">
          {promises.map((p, i) => (
            <Reveal key={p.title} delay={i * 0.07}>
              <div className="surface-card h-full p-6">
                <div className="mb-4 flex h-9 w-9 items-center justify-center rounded-full border border-[rgba(11,61,46,0.3)] bg-[rgba(11,61,46,0.08)]">
                  <svg width="14" height="14" viewBox="0 0 24 24" fill="none" stroke="var(--color-synapse)" strokeWidth="2.4" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
                    <path d="M5 12l5 5L20 7" />
                  </svg>
                </div>
                <h3 className="font-display text-[18px] leading-tight text-[var(--color-ink)]">{p.title}</h3>
                <p className="mt-2 text-[14px] leading-relaxed text-[var(--color-ink-mute)]">{p.desc}</p>
              </div>
            </Reveal>
          ))}
        </div>
      </div>
    </section>
  );
}

/* ════════════ 5. CLOSING CTA ════════════ */
function ClosingCTA() {
  return (
    <section className="relative w-full overflow-hidden border-t border-[rgba(20,17,11,0.12)] px-6 py-40">
      <div className="pointer-events-none absolute inset-0 flex items-center justify-center">
        <motion.div
          animate={{ scale: [1, 1.12, 1], opacity: [0.06, 0.12, 0.06] }}
          transition={{ duration: 5, repeat: Infinity, ease: EASE_OUT }}
          className="h-[300px] w-[600px] rounded-full bg-[var(--color-synapse-glow)] blur-[120px]"
        />
      </div>
      <div className="relative z-10 mx-auto max-w-3xl text-center">
        <Reveal>
          <p className="font-display text-[clamp(28px,5vw,52px)] leading-[1.12] tracking-[-0.03em] text-[var(--color-ink)] text-balance">
            Make AI useful to every ordinary person,
            <br />
            <span className="italic text-[var(--color-synapse)]">on every computer, for free.</span>
          </p>
          <p className="text-lead mt-8 mx-auto max-w-lg text-[var(--color-ink-soft)] text-pretty">
            No lock-in. No tracking. No compromise on speed or safety. That&apos;s the whole mission.
          </p>
          <div className="mt-12 flex flex-col items-center justify-center gap-3 sm:flex-row">
            <Link href="/download" prefetch className="btn btn-primary">
              Download Condura
              <span aria-hidden>→</span>
            </Link>
            <Link href="/security" prefetch className="btn btn-ghost">
              How it stays safe
            </Link>
          </div>
        </Reveal>
      </div>
    </section>
  );
}
