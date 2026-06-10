"use client";

/*
  The Gatekeeper as a circuit breaker — the literal device on the wire
  that powers the bulb. A safe pulse flows through; a destructive surge
  trips the switch and never reaches your screen. One shared clock, so
  pulses, the switch arm and the verdicts stay in sync. Runs only while
  visible; the audience can pause it.
*/
import { m, useInView } from "motion/react";
import { useRef, useState } from "react";
import { usePrefersReducedMotion } from "@/lib/use-reduced-motion";

const T = 9;

const pulseSafe = {
  left: ["6%", "44%", "44%", "88%", "88%"],
  opacity: [0, 1, 1, 1, 0],
  transition: {
    duration: T,
    times: [0.03, 0.18, 0.3, 0.44, 0.47],
    repeat: Infinity,
    ease: "linear" as const,
  },
};

const pulseSurge = {
  left: ["6%", "44%", "44%", "44%"],
  opacity: [0, 1, 1, 0],
  scale: [1, 1, 1.6, 0.2],
  transition: {
    duration: T,
    times: [0.53, 0.68, 0.8, 0.86],
    repeat: Infinity,
    ease: "linear" as const,
  },
};

/* the breaker arm: closed (0deg) → tripped open (-38deg) on the surge */
const arm = {
  rotate: [0, 0, -38, -38, 0],
  transition: {
    duration: T,
    times: [0, 0.68, 0.72, 0.94, 0.99],
    repeat: Infinity,
    ease: "easeOut" as const,
  },
};

const verdict = (times: number[]) => ({
  opacity: [0, 1, 1, 0],
  y: [4, 0, 0, -2],
  transition: { duration: T, times, repeat: Infinity, ease: "linear" as const },
});

const still = { opacity: 0 };

export function Breaker() {
  const reduced = usePrefersReducedMotion();
  const ref = useRef<HTMLDivElement>(null);
  const inView = useInView(ref, { amount: 0.3 });
  const [playing, setPlaying] = useState(true);
  const running = inView && playing && !reduced;

  return (
    <figure aria-label="Schematic: every model proposal travels the wire through the Gatekeeper, a deterministic circuit breaker, before it can reach your screen. Safe actions pass; destructive surges trip the breaker.">
      <div
        ref={ref}
        className="relative border border-line bg-ink-2/70 px-4 pt-14 pb-16 shadow-[0_24px_80px_-24px_var(--t-shadow)] md:px-10 md:pt-16 md:pb-20"
      >
        {/* the wire */}
        <div className="absolute inset-x-[4%] top-1/2 -mt-8 h-px bg-line-strong" aria-hidden />
        {/* energized dashes once running */}
        {running && (
          <svg className="absolute inset-x-[4%] top-1/2 -mt-8 h-px w-[92%] overflow-visible" aria-hidden>
            <line x1="0" y1="0" x2="100%" y2="0" stroke="var(--t-glow)" strokeWidth="2" className="current" opacity="0.5" />
          </svg>
        )}

        {/* pulses */}
        {!reduced && (
          <div aria-hidden className="motion-reduce:hidden">
            <m.span
              animate={running ? pulseSafe : still}
              className="absolute top-1/2 z-20 -mt-[37px] size-2.5 -translate-x-1 rounded-full bg-brass shadow-[0_0_12px_2px_var(--t-glow)]"
            />
            <m.span
              animate={running ? pulseSurge : still}
              className="absolute top-1/2 z-20 -mt-[37px] size-2.5 -translate-x-1 rounded-full bg-halt shadow-[0_0_12px_2px_var(--t-halt)]"
            />
          </div>
        )}

        <div className="relative grid grid-cols-3 items-center gap-2 md:gap-6">
          {/* source */}
          <div className="relative z-10 border border-line bg-ink-3 px-3 py-3 text-center md:px-5">
            <p className="font-mono text-[11px] tracking-[0.14em] uppercase">Strategist</p>
            <p className="annotation mt-1 !text-ivory-faint !tracking-[0.1em] normal-case">
              a model · proposes
            </p>
          </div>

          {/* the breaker box */}
          <div className="relative z-10 mx-auto w-full max-w-[210px]">
            <div className="border border-line-strong bg-ink-3 px-3 pt-3 pb-5 text-center md:px-5">
              <p className="font-mono text-[11px] tracking-[0.14em] uppercase">Gatekeeper</p>
              <p className="annotation mt-1 !text-ivory-faint !tracking-[0.1em] normal-case">
                deterministic breaker
              </p>
              {/* the switch */}
              <svg viewBox="0 0 120 44" className="mx-auto mt-3 w-24" aria-hidden>
                <circle cx="14" cy="30" r="5" fill="none" stroke="var(--t-fg-dim)" strokeWidth="2.5" />
                <circle cx="106" cy="30" r="5" fill="none" stroke="var(--t-fg-dim)" strokeWidth="2.5" />
                <m.line
                  x1="19"
                  y1="30"
                  x2="101"
                  y2="30"
                  stroke="var(--t-accent)"
                  strokeWidth="3.5"
                  strokeLinecap="round"
                  style={{ originX: "19px", originY: "30px" }}
                  animate={running ? arm : { rotate: 0 }}
                />
              </svg>
            </div>
            {/* verdicts */}
            <div
              className="pointer-events-none absolute inset-x-0 -bottom-10 text-center font-mono text-[11px] tracking-[0.2em] uppercase"
              aria-hidden
            >
              {!running ? (
                <span>
                  <span className="text-brass">pass</span>
                  <span className="mx-2 text-ivory-faint">/</span>
                  <span className="text-halt">tripped</span>
                </span>
              ) : (
                <>
                  <m.span animate={verdict([0.2, 0.26, 0.42, 0.47])} className="absolute inset-x-0 text-brass">
                    pass · click “export”
                  </m.span>
                  <m.span animate={verdict([0.7, 0.75, 0.9, 0.96])} className="absolute inset-x-0 text-halt">
                    tripped · rm -rf ~
                  </m.span>
                </>
              )}
            </div>
          </div>

          {/* destination */}
          <div className="relative z-10 border border-line bg-ink-3 px-3 py-3 text-center md:px-5">
            <p className="font-mono text-[11px] tracking-[0.14em] uppercase">Your screen</p>
            <p className="annotation mt-1 !text-ivory-faint !tracking-[0.1em] normal-case">
              clicks · keys · shell
            </p>
          </div>
        </div>
      </div>
      <figcaption className="annotation mt-4 flex flex-wrap items-center gap-x-6 gap-y-1">
        <span>
          <span aria-hidden className="mr-2 inline-block size-1.5 rounded-full bg-brass align-middle" />
          permitted, logged, executed
        </span>
        <span>
          <span aria-hidden className="mr-2 inline-block size-1.5 rounded-full bg-halt align-middle" />
          destructive — the breaker trips, a human decides
        </span>
        {!reduced && (
          <button
            type="button"
            onClick={() => setPlaying((p) => !p)}
            aria-pressed={!playing}
            className="annotation ml-auto border border-line px-2.5 py-1 transition-colors duration-200 hover:border-line-strong hover:!text-ivory"
          >
            {playing ? "pause" : "play"}
          </button>
        )}
      </figcaption>
    </figure>
  );
}
