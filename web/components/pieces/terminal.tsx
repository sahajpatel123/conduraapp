"use client";

/*
  The hero set-piece: a terminal that performs the summon, once, when it
  enters view. Under reduced motion it simply presents the finished state.
*/
import { useInView } from "motion/react";
import { useEffect, useRef, useState } from "react";
import { usePrefersReducedMotion } from "@/lib/use-reduced-motion";

const CMD = "synaptic --summon";

const OUTPUT = [
  { text: "overlay raised", detail: "87ms", tone: "brass" },
  { text: "orchestra detected", detail: "claude code · codex · ollama +5", tone: "dim" },
  { text: "gatekeeper", detail: "armed", tone: "dim" },
  { text: "listening for the wake word", detail: "", tone: "dim" },
] as const;

export function Terminal() {
  const ref = useRef<HTMLDivElement>(null);
  const inView = useInView(ref, { once: true, amount: 0.4 });
  const reduced = usePrefersReducedMotion();
  const [typedState, setTyped] = useState(0);
  const [shownState, setShown] = useState(0);

  // Under reduced motion the finished state is derived, never animated to.
  const typed = reduced ? CMD.length : typedState;
  const shown = reduced ? OUTPUT.length : shownState;
  const done = typed >= CMD.length;

  useEffect(() => {
    if (!inView || reduced) return;
    if (!done) {
      const t = setTimeout(() => setTyped((n) => n + 1), 45 + Math.sin(typedState) * 18);
      return () => clearTimeout(t);
    }
    if (shownState < OUTPUT.length) {
      const t = setTimeout(() => setShown((n) => n + 1), shownState === 0 ? 350 : 420);
      return () => clearTimeout(t);
    }
  }, [inView, reduced, typedState, shownState, done]);

  return (
    <div
      ref={ref}
      role="img"
      aria-label="Terminal demonstration: the synaptic summon command raises the overlay in 87 milliseconds, detects installed AI tools, arms the gatekeeper, and listens for the wake word."
      className="border border-line bg-ink-2/80 font-mono text-[13px] leading-7 shadow-[0_30px_80px_rgba(0,0,0,0.45)]"
    >
      <div className="flex items-center justify-between border-b border-line px-4 py-2.5">
        <span className="annotation !text-ivory-faint">synapticd — local</span>
        <span className="flex gap-1.5" aria-hidden>
          <span className="size-2 rounded-full border border-line-strong" />
          <span className="size-2 rounded-full border border-line-strong" />
          <span className="size-2 rounded-full bg-brass/80" />
        </span>
      </div>
      <div className="px-4 py-4 md:px-5" aria-hidden>
        <p className="text-ivory">
          <span className="text-ivory-faint">$ </span>
          {CMD.slice(0, typed)}
          {!done && <span className="caret" />}
        </p>
        {OUTPUT.slice(0, shown).map((line, i) => (
          <p key={i} className="text-ivory-dim">
            <span className="text-brass">▸ </span>
            {line.text}
            {line.detail && (
              <>
                <span className="text-ivory-faint"> ·· </span>
                <span className={line.tone === "brass" ? "text-brass" : "text-ivory"}>
                  {line.detail}
                </span>
              </>
            )}
            {i === OUTPUT.length - 1 && <span className="caret" />}
          </p>
        ))}
      </div>
    </div>
  );
}
