"use client";

import { useCallback, useEffect, useRef, useState } from "react";
import { motion, AnimatePresence } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/**
 * HeroConductor — the hero's living centerpiece.
 *
 * A paper card that shows Condura doing what it does: one hotkey summons
 * the overlay, routes to a tool, acts on screen. Cycles through quiet
 * scenes on a timer; click or tap any tool node to jump. Mature, not loud.
 */
const SCENES = [
  {
    keys: ["⌥", "⌥"],
    active: 1,
    path: [1, 2],
    tool: "Condura",
    status: "Overlay open — waiting for your words…",
    radius: "READ" as const,
  },
  {
    keys: ["⌥", "⌥"],
    active: 2,
    path: [1, 2, 3],
    tool: "Claude Code",
    status: "Routing refactor to Claude Code…",
    radius: "WRITE" as const,
  },
  {
    keys: ["⌥", "⌥"],
    active: 3,
    path: [1, 2, 3],
    tool: "Codex",
    status: "Running tests in Terminal…",
    radius: "READ" as const,
  },
  {
    keys: ["⌥", "⌥"],
    active: 0,
    path: [1, 0],
    tool: "Ollama",
    status: "Summarizing offline — no API key needed…",
    radius: "READ" as const,
  },
  {
    keys: ["⌥", "⌥"],
    active: 1,
    path: [1, 2, 3],
    tool: "Condura",
    status: "Sending email — Gatekeeper requests your Allow…",
    radius: "NETWORK" as const,
  },
] as const;

const NODES = [
  { id: "ollama", label: "Ollama", x: 12 },
  { id: "condura", label: "Condura", x: 38 },
  { id: "claude", label: "Claude", x: 62 },
  { id: "codex", label: "Codex", x: 88 },
] as const;

const RADIUS_COLOR = {
  READ: "var(--color-ok)",
  WRITE: "var(--color-pollen)",
  NETWORK: "var(--color-danger)",
} as const;

export default function HeroConductor({
  inView,
  delay = 0.45,
}: {
  inView: boolean;
  delay?: number;
}) {
  const [scene, setScene] = useState(0);
  const [pressing, setPressing] = useState(false);
  const intervalRef = useRef<ReturnType<typeof setInterval> | null>(null);

  const current = SCENES[scene];

  const advance = useCallback((next?: number) => {
    setPressing(true);
    setTimeout(() => setPressing(false), 220);
    setScene((s) => (next !== undefined ? next : (s + 1) % SCENES.length));
  }, []);

  useEffect(() => {
    if (!inView) return;
    intervalRef.current = setInterval(() => advance(), 4200);
    return () => {
      if (intervalRef.current) clearInterval(intervalRef.current);
    };
  }, [inView, advance]);

  const onNodeClick = (idx: number) => {
    if (intervalRef.current) clearInterval(intervalRef.current);
    const match = SCENES.findIndex((s) => s.active === idx);
    advance(match >= 0 ? match : scene);
    intervalRef.current = setInterval(() => advance(), 4200);
  };

  // Build SVG path through active nodes
  const pathD = () => {
    const pts = current.path.map((i) => {
      const n = NODES[i];
      return { x: n.x, y: 36 };
    });
    if (pts.length < 2) return "";
    let d = `M ${pts[0].x} ${pts[0].y}`;
    for (let i = 1; i < pts.length; i++) {
      const prev = pts[i - 1];
      const cur = pts[i];
      const mx = (prev.x + cur.x) / 2;
      d += ` C ${mx} ${prev.y - 6}, ${mx} ${cur.y + 6}, ${cur.x} ${cur.y}`;
    }
    return d;
  };

  return (
    <motion.div
      initial={{ opacity: 0, y: 20, scale: 0.98 }}
      animate={
        inView
          ? { opacity: 1, y: 0, scale: 1 }
          : { opacity: 0, y: 20, scale: 0.98 }
      }
      transition={{ duration: 0.95, ease: EASE_OUT, delay }}
      className="hero-conductor group relative mx-auto w-full max-w-[620px]"
      onClick={() => advance()}
      role="presentation"
    >
      {/* soft bloom behind the card */}
      <div
        className="pointer-events-none absolute -inset-8 rounded-[40px] opacity-60 transition-opacity duration-700 group-hover:opacity-80"
        style={{
          background:
            "radial-gradient(ellipse at 50% 60%, rgba(26,138,106,0.14), transparent 68%)",
        }}
      />

      <div className="surface-card relative overflow-hidden rounded-[22px]">
        <div className="paper-grain absolute inset-0 opacity-40" />

        {/* top row — hotkey + live pip */}
        <div className="relative flex items-center justify-between border-b border-[rgba(20,17,11,0.08)] px-5 py-3.5 sm:px-6">
          <div className="flex items-center gap-2.5">
            <span className="text-mono-label text-[10px]">Hotkey</span>
            <div className="flex gap-1.5">
              {current.keys.map((k, i) => (
                <motion.kbd
                  key={i}
                  animate={{
                    y: pressing ? 2 : 0,
                    boxShadow: pressing
                      ? "0 0 0 rgba(20,17,11,0)"
                      : "0 2px 0 rgba(20,17,11,0.12), 0 1px 0 rgba(255,255,255,0.5) inset",
                  }}
                  transition={{ duration: 0.16, ease: EASE_OUT }}
                  className="flex h-8 min-w-8 items-center justify-center rounded-lg border border-[rgba(20,17,11,0.12)] bg-[var(--color-paper)] px-2 font-mono text-[13px] font-medium text-[var(--color-ink)]"
                >
                  {k}
                </motion.kbd>
              ))}
            </div>
          </div>
          <div className="flex items-center gap-2">
            <span className="relative flex h-2 w-2">
              <span className="absolute inset-0 animate-ping rounded-full bg-[var(--color-synapse-glow)] opacity-40" />
              <span className="relative h-2 w-2 rounded-full bg-[var(--color-synapse-glow)]" />
            </span>
            <span className="font-mono text-[10px] uppercase tracking-[0.18em] text-[var(--color-ink-mute)]">
              Live
            </span>
          </div>
        </div>

        {/* routing diagram */}
        <div className="relative px-4 py-5 sm:px-6 sm:py-6">
          <svg
            viewBox="0 0 100 44"
            className="h-[72px] w-full sm:h-[84px]"
            aria-hidden
          >
            <motion.path
              key={`path-${scene}`}
              d={pathD()}
              fill="none"
              stroke="var(--color-synapse-glow)"
              strokeWidth="0.85"
              strokeLinecap="round"
              initial={{ pathLength: 0, opacity: 0.5 }}
              animate={{ pathLength: 1, opacity: 1 }}
              transition={{ duration: 1.1, ease: EASE_OUT }}
            />
            {NODES.map((node, i) => {
              const lit =
                (current.path as readonly number[]).includes(i) ||
                current.active === i;
              const isCenter = i === 1;
              return (
                <g key={node.id}>
                  <motion.circle
                    cx={node.x}
                    cy={36}
                    r={isCenter ? 2.6 : 2}
                    fill={lit ? "var(--color-synapse)" : "var(--color-ink-faint)"}
                    animate={{
                      scale: current.active === i ? [1, 1.35, 1] : 1,
                    }}
                    transition={{ duration: 0.6, ease: EASE_OUT }}
                    style={{ transformOrigin: `${node.x}px 36px` }}
                  />
                  {current.active === i && (
                    <motion.circle
                      cx={node.x}
                      cy={36}
                      r={4}
                      fill="none"
                      stroke="var(--color-synapse-glow)"
                      strokeWidth="0.35"
                      initial={{ opacity: 0.7, scale: 0.8 }}
                      animate={{ opacity: 0, scale: 1.6 }}
                      transition={{ duration: 1.8, repeat: Infinity }}
                    />
                  )}
                </g>
              );
            })}
          </svg>

          {/* tool labels — clickable */}
          <div className="mt-1 grid grid-cols-4 gap-1">
            {NODES.map((node, i) => (
              <button
                key={node.id}
                type="button"
                onClick={(e) => {
                  e.stopPropagation();
                  onNodeClick(i);
                }}
                className={`truncate rounded-md px-1 py-1 font-mono text-[9px] uppercase tracking-[0.14em] transition-colors sm:text-[10px] ${
                  current.active === i
                    ? "text-[var(--color-synapse)]"
                    : "text-[var(--color-ink-faint)] hover:text-[var(--color-ink-mute)]"
                }`}
              >
                {node.label}
              </button>
            ))}
          </div>
        </div>

        {/* status strip */}
        <div className="relative flex items-center justify-between gap-4 border-t border-[rgba(20,17,11,0.08)] bg-[rgba(236,229,212,0.45)] px-5 py-3.5 sm:px-6">
          <div className="min-w-0 flex-1 overflow-hidden">
            <AnimatePresence mode="wait">
              <motion.p
                key={scene}
                initial={{ opacity: 0, y: 6, filter: "blur(4px)" }}
                animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
                exit={{ opacity: 0, y: -6, filter: "blur(4px)" }}
                transition={{ duration: 0.45, ease: EASE_OUT }}
                className="truncate font-mono text-[11px] tracking-[-0.01em] text-[var(--color-ink-soft)] sm:text-[12px]"
              >
                <span className="text-[var(--color-synapse)]">›</span>{" "}
                {current.status}
              </motion.p>
            </AnimatePresence>
          </div>
          <span
            className="shrink-0 rounded-full px-2 py-0.5 font-mono text-[9px] uppercase tracking-[0.14em]"
            style={{
              color: RADIUS_COLOR[current.radius],
              background: `color-mix(in srgb, ${RADIUS_COLOR[current.radius]} 12%, transparent)`,
              border: `1px solid color-mix(in srgb, ${RADIUS_COLOR[current.radius]} 28%, transparent)`,
            }}
          >
            {current.radius}
          </span>
        </div>
      </div>

      <p className="mt-2.5 text-center font-mono text-[9px] uppercase tracking-[0.18em] text-[var(--color-ink-ghost)] sm:text-[10px]">
        Tap to advance · click a node to route
      </p>
    </motion.div>
  );
}
