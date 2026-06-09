"use client";

/*
  The Gatekeeper schematic. One nine-second bar, repeated: an approved
  action passes through to the screen; a destructive one is halted at the
  gate. All choreography is keyframed against the same clock so the dots,
  stamps and captions can never drift apart. The loop only runs while the
  schematic is on screen, and the audience can pause it.
*/
import { m, useInView } from "motion/react";
import { useRef, useState } from "react";
import { usePrefersReducedMotion } from "@/lib/use-reduced-motion";
import { ruleDraw, VIEWPORT } from "@/lib/motion";

const T = 9; // seconds per full cycle

const dotPass = {
  left: ["12%", "46%", "46%", "82%", "82%"],
  opacity: [0, 1, 1, 1, 0],
  transition: {
    duration: T,
    times: [0.02, 0.16, 0.3, 0.42, 0.46],
    repeat: Infinity,
    ease: "linear" as const,
  },
};

const dotHalt = {
  left: ["12%", "46%", "46%", "46%"],
  opacity: [0, 1, 1, 0],
  scale: [1, 1, 1, 0.4],
  transition: {
    duration: T,
    times: [0.54, 0.68, 0.84, 0.9],
    repeat: Infinity,
    ease: "linear" as const,
  },
};

const stamp = (times: number[]) => ({
  opacity: [0, 1, 1, 0],
  y: [4, 0, 0, -2],
  transition: { duration: T, times, repeat: Infinity, ease: "linear" as const },
});

const still = { opacity: 0 };

function Node({
  label,
  sub,
  highlight = false,
}: {
  label: string;
  sub: string;
  highlight?: boolean;
}) {
  return (
    <div
      className={`relative z-10 border bg-ink-2 px-3 py-3 text-center md:px-5 ${
        highlight ? "border-line-strong" : "border-line"
      }`}
    >
      <p className="font-mono text-[11px] tracking-[0.14em] uppercase text-ivory md:text-xs">
        {label}
      </p>
      <p className="annotation mt-1 !text-ivory-faint !tracking-[0.1em] normal-case">{sub}</p>
    </div>
  );
}

export function Gatekeeper() {
  const reduced = usePrefersReducedMotion();
  const ref = useRef<HTMLDivElement>(null);
  const inView = useInView(ref, { amount: 0.3 });
  const [playing, setPlaying] = useState(true);
  const running = inView && playing && !reduced;

  return (
    <figure aria-label="Schematic: every model proposal must pass the deterministic Gatekeeper before it can touch the screen. Approved actions pass; destructive ones are halted.">
      <div
        ref={ref}
        className="relative border border-line bg-ink/60 px-4 pt-12 pb-16 md:px-10 md:pt-16 md:pb-20"
      >
        {/* the wire draws itself in */}
        <m.div
          initial="hidden"
          whileInView="visible"
          viewport={VIEWPORT}
          variants={ruleDraw}
          className="absolute inset-x-[8%] top-1/2 -mt-7 h-px origin-left bg-line-strong"
          aria-hidden
        />

        {/* travelling actions — hidden outright for reduced motion */}
        {!reduced && (
          <div aria-hidden className="motion-reduce:hidden">
            <m.span
              animate={running ? dotPass : still}
              className="absolute top-1/2 z-20 -mt-[31px] size-2 -translate-x-1 rounded-full bg-brass"
            />
            <m.span
              animate={running ? dotHalt : still}
              className="absolute top-1/2 z-20 -mt-[31px] size-2 -translate-x-1 rounded-full bg-halt"
            />
          </div>
        )}

        <div className="relative grid grid-cols-3 items-center gap-2 md:gap-6">
          <Node label="Strategist" sub="a model · proposes" />
          <div className="relative">
            <Node label="Gatekeeper" sub="deterministic code" highlight />
            {/* verdict stamps */}
            <div
              className="pointer-events-none absolute inset-x-0 -bottom-9 text-center font-mono text-[11px] tracking-[0.2em] uppercase"
              aria-hidden
            >
              {reduced || !running ? (
                <span>
                  <span className="text-brass">pass</span>
                  <span className="mx-2 text-ivory-faint">/</span>
                  <span className="text-halt">halt</span>
                </span>
              ) : (
                <>
                  <m.span
                    animate={stamp([0.18, 0.24, 0.4, 0.46])}
                    className="absolute inset-x-0 text-brass"
                  >
                    pass · click “export”
                  </m.span>
                  <m.span
                    animate={stamp([0.7, 0.76, 0.88, 0.94])}
                    className="absolute inset-x-0 text-halt"
                  >
                    halt · rm -rf ~
                  </m.span>
                </>
              )}
            </div>
          </div>
          <Node label="Your screen" sub="clicks · keys · shell" />
        </div>
      </div>
      <figcaption className="annotation mt-4 flex flex-wrap items-center gap-x-6 gap-y-1">
        <span>
          <span aria-hidden className="mr-2 inline-block size-1.5 rounded-full bg-brass align-middle" />
          permitted, logged, executed
        </span>
        <span>
          <span aria-hidden className="mr-2 inline-block size-1.5 rounded-full bg-halt align-middle" />
          destructive — a human must click
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
