"use client";

import { useState, useEffect } from "react";
import { motion, AnimatePresence } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/**
 * RoutingDemo — the live value proposition inside the hero terminal.
 *
 * Minimal. No glow, no aurora, no fancy effects. Just information made
 * beautiful: the conductor routing real work to sub-agents, one turn at
 * a time. Each turn teaches the entire product in five seconds.
 *
 *   1. A user request appears (one line, typed)
 *   2. A row of agent pills fades in, dim
 *   3. A thin scan-line sweeps across them — the router deciding
 *   4. One pill brightens — the router's choice
 *   5. A thin progress line fills — the agent working
 *   6. The result returns — one line, subtle check
 *
 * Then the next turn. The cycle is the content.
 */

type AgentKey = "claude" | "codex" | "ollama" | "gemini";

const AGENTS: { key: AgentKey; label: string }[] = [
  { key: "claude", label: "Claude Code" },
  { key: "codex", label: "Codex" },
  { key: "ollama", label: "Ollama" },
  { key: "gemini", label: "Gemini" },
];

const TURNS: { request: string; picks: AgentKey; result: string }[] = [
  {
    request: "review the auth middleware for timing attacks",
    picks: "claude",
    result: "no constant-time violation found · 4 files",
  },
  {
    request: "summarize yesterday's meeting notes",
    picks: "gemini",
    result: "8 decisions, 3 open questions, 2 owners",
  },
  {
    request: "scaffold a rate-limiter from the openapi spec",
    picks: "codex",
    result: "token_bucket.go + 3 tests · all green",
  },
  {
    request: "explain this stack trace offline",
    picks: "ollama",
    result: "nil map write at handler.go:42 · local model",
  },
];

const PHASES = ["request", "routing", "working", "result"] as const;
type Phase = (typeof PHASES)[number];

// Phase timings (ms). Total ~5.6s per turn.
const T = {
  request: 900,   // request types in
  routing: 1100,  // pills appear + scan + selection
  working: 1400,  // progress fills
  result: 800,    // result shown
  rest: 500,      // beat before next turn
} as const;

export default function RoutingDemo({ active }: { active: boolean }) {
  const [turn, setTurn] = useState(0);
  const [phase, setPhase] = useState<Phase>("request");

  useEffect(() => {
    if (!active) return;
    let mounted = true;
    let timer: ReturnType<typeof setTimeout>;

    const run = () => {
      // request → routing → working → result → rest → next turn
      const sequence: [Phase, number][] = [
        ["request", T.request],
        ["routing", T.routing],
        ["working", T.working],
        ["result", T.result],
      ];
      let i = 0;
      const step = () => {
        if (!mounted) return;
        if (i < sequence.length) {
          const [ph, dur] = sequence[i];
          setPhase(ph);
          timer = setTimeout(() => {
            i += 1;
            step();
          }, dur);
        } else {
          // rest beat, then advance turn
          timer = setTimeout(() => {
            setTurn((p) => (p + 1) % TURNS.length);
            setPhase("request");
            timer = setTimeout(run, T.request);
          }, T.rest);
        }
      };
      step();
    };

    // small delay so the terminal intro finishes first
    timer = setTimeout(run, 600);
    return () => {
      mounted = false;
      clearTimeout(timer);
    };
  }, [active]);

  const current = TURNS[turn];
  const pickedIndex = AGENTS.findIndex((a) => a.key === current.picks);

  return (
    <div className="font-mono text-[12px] leading-relaxed text-white/80 min-h-[180px] flex flex-col justify-between">
      <div className="space-y-4">
        {/* ── Request ── */}
        <AnimatePresence mode="wait">
          <motion.div
            key={`req-${turn}`}
            initial={{ opacity: 0, y: 6 }}
            animate={{ opacity: phase === "request" ? 1 : 0.55, y: 0 }}
            transition={{ duration: 0.35, ease: EASE_OUT }}
          >
            <p className="text-white/80">
              <span className="text-white/45 mr-2">❯</span>
              <TypeLine text={current.request} active={phase === "request"} />
            </p>
          </motion.div>
        </AnimatePresence>

        {/* ── Routing: agent pills + scan + selection ── */}
        <AnimatePresence>
          {(phase === "routing" || phase === "working" || phase === "result") && (
            <motion.div
              key={`route-${turn}`}
              initial={{ opacity: 0 }}
              animate={{ opacity: 1 }}
              exit={{ opacity: 0 }}
              transition={{ duration: 0.3 }}
              className="pl-5"
            >
              {/* Eyebrow */}
              <p className="text-[10px] text-white/30 mb-2.5">routing to agent</p>

              {/* Pills row */}
              <div className="flex flex-wrap gap-1.5 relative">
                {AGENTS.map((a, i) => {
                  const isPicked = a.key === current.picks;
                  const pickedVisible = phase !== "routing" || i <= pickedIndex;
                  return (
                    <motion.span
                      key={a.key}
                      initial={{ opacity: 0 }}
                      animate={{
                        opacity: isPicked && pickedVisible ? 1 : 0.32,
                        color: isPicked && pickedVisible ? "rgba(255,255,255,0.92)" : "rgba(255,255,255,0.4)",
                        borderColor:
                          isPicked && pickedVisible ? "rgba(255,255,255,0.28)" : "rgba(255,255,255,0.07)",
                      }}
                      transition={{ duration: 0.35, ease: EASE_OUT }}
                      className="px-2 py-0.5 rounded-md border text-[10.5px] whitespace-nowrap"
                      style={{ borderColor: "rgba(255,255,255,0.07)" }}
                    >
                      {a.label}
                    </motion.span>
                  );
                })}

                {/* Scan line — sweeps across during routing, fades when done */}
                {phase === "routing" && (
                  <motion.div
                    initial={{ scaleX: 0, opacity: 0.6 }}
                    animate={{ scaleX: 1, opacity: 0 }}
                    transition={{ duration: T.routing / 1000, ease: "easeInOut" }}
                    className="absolute left-0 right-0 top-1/2 -translate-y-1/2 h-px bg-gradient-to-r from-transparent via-white/30 to-transparent origin-left pointer-events-none"
                  />
                )}
              </div>

              {/* ── Working: progress fill ── */}
              <AnimatePresence>
                {phase === "working" && (
                  <motion.div
                    key="work"
                    initial={{ opacity: 0 }}
                    animate={{ opacity: 1 }}
                    exit={{ opacity: 0 }}
                    transition={{ duration: 0.25 }}
                    className="mt-3 flex items-center gap-2.5"
                  >
                    <div className="relative h-px w-32 bg-white/[0.08] overflow-hidden rounded-full">
                      <motion.div
                        initial={{ scaleX: 0 }}
                        animate={{ scaleX: 1 }}
                        transition={{ duration: T.working / 1000, ease: "easeInOut" }}
                        className="absolute inset-0 bg-white/55 origin-left rounded-full"
                      />
                    </div>
                    <span className="text-[10px] text-white/30">running</span>
                  </motion.div>
                )}
              </AnimatePresence>

              {/* ── Result ── */}
              <AnimatePresence>
                {phase === "result" && (
                  <motion.div
                    key={`res-${turn}`}
                    initial={{ opacity: 0, y: 4 }}
                    animate={{ opacity: 1, y: 0 }}
                    transition={{ duration: 0.3, ease: EASE_OUT }}
                    className="mt-3 flex items-center gap-2"
                  >
                    <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" className="text-white/55 shrink-0">
                      <path d="M5 12l5 5 9-10" />
                    </svg>
                    <span className="text-[11px] text-white/55">{current.result}</span>
                  </motion.div>
                )}
              </AnimatePresence>
            </motion.div>
          )}
        </AnimatePresence>
      </div>

      {/* Blinking cursor — only when waiting for next request */}
      <div className="mt-5 flex items-center gap-2 font-mono text-[12px]">
        <span className="text-white/45">❯</span>
        <motion.span
          animate={{ opacity: [1, 0] }}
          transition={{ repeat: Infinity, duration: 0.9 }}
          className="inline-block w-[7px] h-[14px] bg-white/40"
        />
      </div>
    </div>
  );
}

/* ── Typewriter for the request line ── */
function TypeLine({ text, active }: { text: string; active: boolean }) {
  const [shown, setShown] = useState(active ? 0 : text.length);

  useEffect(() => {
    if (!active) {
      setShown(text.length);
      return;
    }
    setShown(0);
    let i = 0;
    const tick = () => {
      i += 1;
      setShown(i);
      if (i < text.length) timer = setTimeout(tick, 28);
    };
    let timer = setTimeout(tick, 28);
    return () => clearTimeout(timer);
  }, [active, text]);

  return <span>{text.slice(0, shown)}</span>;
}