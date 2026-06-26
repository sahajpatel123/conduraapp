"use client";

import { useEffect, useRef, useState } from "react";
import { motion, useInView } from "motion/react";
import Reveal from "@/components/motion/Reveal";
import WordReveal from "@/components/motion/WordReveal";
import { EASE_OUT } from "@/lib/motion";

/**
 * TheConductor — "how it works" in three acts.
 * Each act has a living micro-illustration rendered in SVG + motion:
 *   Act 1 — The Hotkey: a key that presses itself, ripples, summons the overlay
 *   Act 2 — The Overlay: a prompt that types itself and streams a reply
 *   Act 3 — The Gatekeeper: a destructive action hitting a wall, then a human "allow"
 * The acts auto-advance on a timer and on scroll, and are also clickable.
 */
export default function TheConductor() {
  const [active, setActive] = useState(0);
  const ref = useRef<HTMLDivElement | null>(null);
  const inView = useInView(ref, { margin: "-20%" });

  useEffect(() => {
    if (!inView) return;
    const id = setInterval(() => setActive((a) => (a + 1) % ACTS.length), 5200);
    return () => clearInterval(id);
  }, [inView]);

  return (
    <section ref={ref} data-section="conductor" className="relative mx-auto max-w-[1180px] px-6 py-28 sm:py-36">
      <header className="mb-16 sm:mb-20">
        <Reveal>
          <p className="text-eyebrow mb-6 sm:mb-8">— How it works</p>
        </Reveal>
        <WordReveal
          as="h2"
          text="Three moves. Then it gets out of your way."
          className="text-display max-w-[18ch] text-balance leading-[1.12] text-[var(--color-ink)]"
        />
      </header>

      <div className="grid gap-10 md:grid-cols-[1.1fr_1fr] md:gap-16">
        {/* ── Stage ── */}
        <div className="surface-card relative aspect-[4/3] overflow-hidden md:aspect-auto md:min-h-[440px]">
          <div className="paper-grain absolute inset-0" />
          <div className="absolute inset-0 p-6 sm:p-10">
            {ACTS.map((act, i) => (
              <ActScene
                key={act.id}
                act={act}
                active={active === i}
                index={i}
              />
            ))}
          </div>

          {/* act pips */}
          <div className="absolute bottom-5 left-1/2 z-20 flex -translate-x-1/2 gap-2">
            {ACTS.map((_, i) => (
              <button
                key={i}
                onClick={() => setActive(i)}
                aria-label={`Go to act ${i + 1}`}
                className={`h-1.5 rounded-full transition-all duration-300 ${
                  active === i
                    ? "w-7 bg-[var(--color-ink)]"
                    : "w-1.5 bg-[var(--color-ink-faint)] hover:bg-[var(--color-ink-mute)]"
                }`}
              />
            ))}
          </div>
        </div>

        {/* ── Act copy ── */}
        <div className="flex flex-col gap-3 md:pt-6">
          {ACTS.map((act, i) => (
            <button
              key={act.id}
              onClick={() => setActive(i)}
              className={`group rounded-2xl p-5 text-left transition-all duration-300 ${
                active === i
                  ? "bg-[var(--color-paper-warm)] border border-[rgba(20,17,11,0.10)]"
                  : "border border-transparent hover:bg-[rgba(20,17,11,0.03)]"
              }`}
            >
              <div className="flex items-baseline gap-3">
                <span
                  className={`font-mono text-[12px] transition-colors ${
                    active === i ? "text-[var(--color-synapse)]" : "text-[var(--color-ink-faint)]"
                  }`}
                >
                  {act.numeral}
                </span>
                <h3
                  className={`font-display text-[24px] leading-tight transition-colors ${
                    active === i ? "text-[var(--color-ink)]" : "text-[var(--color-ink-mute)]"
                  }`}
                >
                  {act.title}
                </h3>
              </div>
              <AnimateHeight open={active === i}>
                <p className="text-body mt-3 max-w-[46ch] text-[var(--color-ink-soft)]">
                  {act.body}
                </p>
              </AnimateHeight>
            </button>
          ))}
        </div>
      </div>
    </section>
  );
}

function AnimateHeight({ open, children }: { open: boolean; children: React.ReactNode }) {
  return (
    <motion.div
      initial={false}
      animate={{ height: open ? "auto" : 0, opacity: open ? 1 : 0 }}
      transition={{ duration: 0.4, ease: EASE_OUT }}
      style={{ overflow: "hidden" }}
    >
      <div>{children}</div>
    </motion.div>
  );
}

function ActScene({
  act,
  active,
  index,
}: {
  act: (typeof ACTS)[number];
  active: boolean;
  index: number;
}) {
  return (
    <motion.div
      initial={false}
      animate={{ opacity: active ? 1 : 0, scale: active ? 1 : 0.98 }}
      transition={{ duration: 0.5, ease: EASE_OUT }}
      style={{ pointerEvents: active ? "auto" : "none" }}
      className="absolute inset-6 sm:inset-10"
    >
      {act.id === "hotkey" && <HotkeyScene />}
      {act.id === "overlay" && <OverlayScene />}
      {act.id === "gatekeeper" && <GatekeeperScene />}
    </motion.div>
  );
}

/* ────────────────────────────────────────────────────────────────
   Act 1 — The Hotkey
   A key that presses itself, emits a ripple, and summons the overlay chip.
   ──────────────────────────────────────────────────────────────── */
function HotkeyScene() {
  return (
    <div className="relative flex h-full flex-col items-center justify-center">
      <div className="relative">
        {/* key */}
        <motion.div
          initial={{ y: 0 }}
          animate={{ y: [0, 0, 6, 0] }}
          transition={{ duration: 2.4, times: [0, 0.4, 0.5, 0.6], repeat: Infinity, repeatDelay: 2.6 }}
          className="relative grid h-24 w-24 place-items-center rounded-2xl border border-[rgba(20,17,11,0.18)] bg-[var(--color-paper)] shadow-[0_18px_40px_-18px_rgba(20,17,11,0.4)]"
        >
          <span className="font-mono text-[11px] uppercase tracking-[0.18em] text-[var(--color-ink-mute)]">
            Press
          </span>
          <span className="font-display text-[26px] text-[var(--color-ink)]">⌘⇧␣</span>
          {/* keytop sheen */}
          <span className="pointer-events-none absolute inset-x-3 top-1 h-px bg-[rgba(255,255,255,0.5)]" />
        </motion.div>
        {/* ripple */}
        <motion.span
          className="absolute left-1/2 top-1/2 h-24 w-24 -translate-x-1/2 -translate-y-1/2 rounded-2xl border border-[var(--color-synapse)]"
          animate={{ scale: [1, 2.2], opacity: [0.5, 0] }}
          transition={{ duration: 1.4, repeat: Infinity, repeatDelay: 3.6, ease: EASE_OUT }}
        />
      </div>

      {/* overlay chip rising */}
      <motion.div
        className="mt-10 flex items-center gap-3 rounded-full border border-[rgba(20,17,11,0.12)] bg-[var(--color-paper)] px-4 py-2.5 shadow-[0_20px_50px_-22px_rgba(20,17,11,0.4)]"
        animate={{ y: [12, 0], opacity: [0, 1] }}
        transition={{ duration: 0.7, delay: 0.7, repeat: Infinity, repeatDelay: 4.3, ease: EASE_OUT }}
      >
        <span className="relative h-2 w-2">
          <span className="absolute inset-0 rounded-full bg-[var(--color-synapse-light)]" />
          <span className="absolute inset-0 animate-[breathe_2s_ease-in-out_infinite] rounded-full bg-[var(--color-synapse-glow)]" />
        </span>
        <span className="text-[13px] font-medium text-[var(--color-ink)]">Condura is listening…</span>
      </motion.div>
    </div>
  );
}

/* ────────────────────────────────────────────────────────────────
   Act 2 — The Overlay
   A prompt that types itself, then streams a reply with a caret.
   ──────────────────────────────────────────────────────────────── */
function OverlayScene() {
  const typed = useTypewriter(
    "Open the Figma file and add a 12px grid.",
    38,
  );
  return (
    <div className="flex h-full flex-col justify-center">
      <div className="rounded-2xl border border-[rgba(20,17,11,0.14)] bg-[var(--color-paper)] p-4 shadow-[0_24px_60px_-28px_rgba(20,17,11,0.4)]">
        <div className="mb-3 flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-[var(--color-pollen)]" />
          <span className="text-mono-label">You · overlay</span>
        </div>
        <p className="font-mono text-[14px] text-[var(--color-ink)] min-h-[1.4em]">
          {typed}
          <span className="ml-0.5 inline-block h-[1.05em] w-[2px] translate-y-[2px] bg-[var(--color-ink)] animate-[breathe_1s_steps(2)_infinite]" />
        </p>
        <div className="rule-ink my-4" />
        <div className="flex items-start gap-2.5">
          <span className="mt-1 h-2 w-2 rounded-full bg-[var(--color-synapse)]" />
          <div className="flex-1">
            <div className="text-mono-label mb-1.5">Condura</div>
            <StreamingLines />
          </div>
        </div>
      </div>
    </div>
  );
}

function StreamingLines() {
  const [n, setN] = useState(0);
  useEffect(() => {
    const id = setInterval(() => setN((x) => (x + 1) % 4), 700);
    return () => clearInterval(id);
  }, []);
  return (
    <div className="space-y-1.5">
      {[0, 1, 2, 3].map((i) => (
        <div
          key={i}
          className="h-2 rounded-full bg-[var(--color-ink-faint)] opacity-30 transition-all"
          style={{ width: `${[100, 92, 78, 60][i]}%`, opacity: i <= n ? 0.5 : 0.18 }}
        />
      ))}
    </div>
  );
}

/* ────────────────────────────────────────────────────────────────
   Act 3 — The Gatekeeper
   A destructive action travels along a thread, hits a wall, and waits
   for a human "allow" — which then lets it through with a pollen spark.
   ──────────────────────────────────────────────────────────────── */
function GatekeeperScene() {
  const phase = useCycle(4000);
  // phase 0: action travels. phase 1: hits wall, dialog appears. phase 2: allow, passes.
  const stage = phase < 1600 ? 0 : phase < 3200 ? 1 : 2;

  return (
    <div className="relative flex h-full flex-col items-center justify-center">
      <svg viewBox="0 0 300 160" className="w-full max-w-[360px]" aria-hidden>
        {/* the wall (gatekeeper) */}
        <rect x="148" y="20" width="4" height="120" rx="2" fill="var(--color-synapse)" opacity="0.85" />
        <rect x="142" y="18" width="16" height="6" rx="3" fill="var(--color-synapse-deep)" />
        <rect x="142" y="136" width="16" height="6" rx="3" fill="var(--color-synapse-deep)" />

        {/* action bead travelling */}
        <motion.circle
          r="5"
          className="synapse-node"
          animate={{ cx: stage === 0 ? [10, 140] : stage === 1 ? [140, 140] : [140, 290], cy: 80 }}
          transition={{ duration: stage === 0 ? 1.4 : stage === 2 ? 1.2 : 0, ease: EASE_OUT }}
        />
        {/* trail behind */}
        <motion.path
          d="M 10 80 L 140 80"
          className="synapse-thread"
          animate={{ pathLength: stage === 0 ? [0, 1] : 1, opacity: stage === 2 ? 0.3 : 1 }}
          transition={{ duration: 1.4, ease: EASE_OUT }}
        />
        <motion.path
          d="M 156 80 L 290 80"
          className="synapse-thread"
          animate={{ pathLength: stage === 2 ? [0, 1] : 0 }}
          transition={{ duration: 1.2, ease: EASE_OUT }}
        />
        {stage === 2 && (
          <circle cx="290" cy="80" r="4" className="synapse-node">
            <animate attributeName="r" values="3;7;3" dur="0.8s" repeatCount="2" />
          </circle>
        )}
      </svg>

      {/* dialog */}
      <motion.div
        className="mt-6 w-full max-w-[320px] rounded-2xl border border-[rgba(20,17,11,0.14)] bg-[var(--color-paper)] p-4 shadow-[0_24px_60px_-28px_rgba(20,17,11,0.4)]"
        animate={{ opacity: stage >= 1 ? 1 : 0, y: stage >= 1 ? 0 : 10 }}
        transition={{ duration: 0.4, ease: EASE_OUT }}
      >
        <div className="mb-2 flex items-center gap-2">
          <span className="h-2 w-2 rounded-full bg-[var(--color-danger)]" />
          <span className="text-mono-label !text-[var(--color-danger)]">Destructive · needs you</span>
        </div>
        <p className="text-[13px] text-[var(--color-ink-soft)]">
          Condura wants to <span className="font-mono">rm -rf dist/</span>. Allow?
        </p>
        <div className="mt-3 flex gap-2">
          <motion.div
            className="rounded-full bg-[var(--color-ink)] px-3 py-1.5 text-[12px] font-medium text-[var(--color-paper)]"
            animate={{ scale: stage === 2 ? [1, 0.96, 1] : 1 }}
            transition={{ duration: 0.3 }}
          >
            Allow
          </motion.div>
          <div className="rounded-full border border-[rgba(20,17,11,0.18)] px-3 py-1.5 text-[12px] text-[var(--color-ink-mute)]">
            Deny
          </div>
        </div>
      </motion.div>
    </div>
  );
}

/* ── hooks ── */
function useTypewriter(text: string, speed = 30) {
  const [out, setOut] = useState("");
  useEffect(() => {
    let i = 0;
    const id = window.setInterval(() => {
      i += 1;
      setOut(text.slice(0, i));
      if (i >= text.length) window.clearInterval(id);
    }, speed);
    return () => window.clearInterval(id);
  }, [text, speed]);
  return out;
}

function useCycle(period: number) {
  const [t, setT] = useState(0);
  useEffect(() => {
    const start = performance.now();
    let raf = 0;
    const loop = () => {
      const now = performance.now();
      setT((now - start) % period);
      raf = requestAnimationFrame(loop);
    };
    raf = requestAnimationFrame(loop);
    return () => cancelAnimationFrame(raf);
  }, [period]);
  return t;
}

const ACTS = [
  {
    id: "hotkey" as const,
    numeral: "01",
    title: "Summon it with one hotkey.",
    body:
      "On macOS and Windows, pick a combo on first run. Tap it anywhere and the overlay appears. Screen-aware context is on the v0.2.0 roadmap.",
  },
  {
    id: "overlay" as const,
    numeral: "02",
    title: "It conducts the tools you have.",
    body:
      "Condura can spawn individual sub-agents you have installed. Orchestrated parallel waves are coming in v0.2.0.",
  },
  {
    id: "gatekeeper" as const,
    numeral: "03",
    title: "Nothing dangerous happens without you.",
    body:
      "A deterministic Gatekeeper vets every computer-use action. Destructive moves stop at an in-app consent dialog until you click Allow. A native OS dialog is planned for v0.2.0.",
  },
] as const;
