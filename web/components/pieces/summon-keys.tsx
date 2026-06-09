"use client";

/*
  Movement I set-piece: the option key taps twice, the overlay answers.
  Choreographed against a single looping clock; runs only while on
  screen, and the audience can pause it.
*/
import { m, useInView } from "motion/react";
import { useRef, useState } from "react";
import { usePrefersReducedMotion } from "@/lib/use-reduced-motion";

const T = 4.5;

const keyPress = {
  y: [0, 3, 0, 3, 0],
  transition: {
    duration: T,
    times: [0.1, 0.16, 0.22, 0.28, 0.34],
    repeat: Infinity,
    ease: "easeOut" as const,
  },
};

const overlayPop = {
  opacity: [0, 1, 1, 0],
  y: [8, 0, 0, 4],
  transition: {
    duration: T,
    times: [0.36, 0.42, 0.86, 0.95],
    repeat: Infinity,
    ease: "easeOut" as const,
  },
};

export function SummonKeys() {
  const reduced = usePrefersReducedMotion();
  const ref = useRef<HTMLDivElement>(null);
  const inView = useInView(ref, { amount: 0.4 });
  const [playing, setPlaying] = useState(true);
  const running = inView && playing && !reduced;

  return (
    <div
      ref={ref}
      role="img"
      aria-label="Double-tap the option key and the Synaptic overlay appears."
      className="relative flex flex-col items-center gap-8 border border-line bg-ink-2/60 px-8 py-12"
    >
      <div className="flex items-end gap-3" aria-hidden>
        <m.div
          animate={running ? keyPress : { y: 0 }}
          className="flex h-16 w-20 flex-col items-center justify-center border border-line-strong bg-ink-3 shadow-[0_4px_0_rgba(237,232,221,0.18)]"
        >
          <span className="font-mono text-lg text-ivory">⌥</span>
          <span className="annotation !text-ivory-faint">option</span>
        </m.div>
        <span className="annotation pb-1 !text-ivory-faint">× 2</span>
      </div>

      <m.div
        aria-hidden
        animate={running ? overlayPop : { opacity: 1, y: 0 }}
        className="w-full max-w-xs border border-line bg-ink px-4 py-3 font-mono text-xs text-ivory-dim shadow-[0_20px_60px_rgba(0,0,0,0.5)]"
      >
        <span className="mr-2 inline-block size-1.5 rotate-45 bg-brass align-middle" />
        Synaptic — at your service
        <span className="caret" />
      </m.div>

      <p className="annotation" aria-hidden>
        anywhere · any app · 87ms
      </p>

      {!reduced && (
        <button
          type="button"
          onClick={() => setPlaying((p) => !p)}
          aria-pressed={!playing}
          aria-label={playing ? "Pause animation" : "Play animation"}
          className="annotation absolute right-3 bottom-3 border border-line px-2 py-1 transition-colors duration-200 hover:border-line-strong hover:!text-ivory"
        >
          {playing ? "pause" : "play"}
        </button>
      )}
    </div>
  );
}
