"use client";

import { useState, useEffect, useCallback } from "react";
import { motion, AnimatePresence } from "motion/react";
import { EASE_OUT } from "@/lib/motion";

/**
 * ProductTour — the hero's right side, repurposed to actually explain the product.
 *
 * A self-running panel that cycles through four scenes. Each scene is one
 * pillar of Condura, shown as: a headline, one line of description, and a
 * minimal live visual. A thin progress rail marks the active scene.
 *
 *   1. Summon — press your hotkey, it appears (overlay reveal)
 *   2. Route  — it picks the right agent (routing demo)
 *   3. Guard  — it checks every action (gatekeeper verdict)
 *   4. Yours  — everything stays on your machine (local + locked + no telemetry)
 *
 * A first-time visitor understands the whole product in ~24 seconds without
 * reading a single paragraph. Minimal, no glow, no aurora — just clear
 * information made beautiful. Clicking a scene dot jumps to it.
 */

const SCENE_DURATIONS = [6200, 7000, 6200, 6200] as const;

type SceneKey = "summon" | "route" | "guard" | "yours";

const SCENES: { key: SceneKey; index: number; title: string; desc: string }[] = [
  { key: "summon", index: 0, title: "Summon", desc: "Press your hotkey. It appears from your OS — a floating overlay, anywhere." },
  { key: "route", index: 1, title: "Route", desc: "It hands the work to the right agent from the ones you already have." },
  { key: "guard", index: 2, title: "Guard", desc: "Every action passes a deterministic gatekeeper before it runs." },
  { key: "yours", index: 3, title: "Yours", desc: "Memory, keys, and logs stay on your machine. No telemetry. Ever." },
];

export default function ProductTour({ active }: { active: boolean }) {
  const [scene, setScene] = useState<SceneKey>("summon");

  const goTo = useCallback((s: SceneKey) => setScene(s), []);
  const next = useCallback(() => {
    setScene((s) => {
      const i = SCENES.findIndex((x) => x.key === s);
      return SCENES[(i + 1) % SCENES.length].key;
    });
  }, []);

  useEffect(() => {
    if (!active) return;
    let timer: ReturnType<typeof setTimeout>;
    const cycle = () => {
      const idx = SCENES.findIndex((x) => x.key === scene);
      timer = setTimeout(() => next(), SCENE_DURATIONS[idx]);
    };
    cycle();
    return () => clearTimeout(timer);
  }, [scene, active, next]);

  const current = SCENES.find((s) => s.key === scene)!;

  return (
    <div className="font-mono text-[12px] text-white/80 min-h-[230px] flex flex-col">
      {/* ── Scene header ── */}
      <div className="mb-5">
        <AnimatePresence mode="wait">
          <motion.div
            key={`hdr-${scene}`}
            initial={{ opacity: 0, y: 8 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -8 }}
            transition={{ duration: 0.4, ease: EASE_OUT }}
          >
            <div className="flex items-baseline gap-2.5 mb-1.5">
              <span className="font-mono text-[10px] text-white/30 tracking-widest uppercase">
                {String(current.index + 1).padStart(2, "0")}
              </span>
              <h3 className="font-body-mature text-[18px] font-semibold text-white tracking-tight">
                {current.title}
              </h3>
            </div>
            <p className="text-[12px] text-white/45 leading-relaxed max-w-[420px]">
              {current.desc}
            </p>
          </motion.div>
        </AnimatePresence>
      </div>

      {/* ── Scene visual ── */}
      <div className="flex-1 flex items-center min-h-[140px]">
        <AnimatePresence mode="wait">
          <motion.div
            key={`vis-${scene}`}
            initial={{ opacity: 0, y: 10 }}
            animate={{ opacity: 1, y: 0 }}
            exit={{ opacity: 0, y: -10 }}
            transition={{ duration: 0.4, ease: EASE_OUT }}
            className="w-full"
          >
            {scene === "summon" && <SummonVisual />}
            {scene === "route" && <RouteVisual active={active} />}
            {scene === "guard" && <GuardVisual />}
            {scene === "yours" && <YoursVisual />}
          </motion.div>
        </AnimatePresence>
      </div>

      {/* ── Scene rail ── */}
      <div className="mt-5 flex items-center gap-2">
        {SCENES.map((s) => {
          const isActive = s.key === scene;
          return (
            <button
              key={s.key}
              type="button"
              onClick={() => goTo(s.key)}
              aria-label={`${s.title} scene`}
              className="group relative h-1.5 flex-1 max-w-[64px] rounded-full overflow-hidden bg-white/[0.08]"
            >
              <motion.span
                animate={{ width: isActive ? "100%" : "0%" }}
                transition={{ duration: isActive ? SCENE_DURATIONS[s.index] / 1000 : 0.3, ease: "linear" }}
                className="absolute inset-y-0 left-0 bg-white/40 rounded-full"
              />
            </button>
          );
        })}
      </div>
    </div>
  );
}

/* ════════════════════════════════════════════════════════════
   Scene 1 — Summon: keystroke + overlay reveal
   ════════════════════════════════════════════════════════════ */
function SummonVisual() {
  return (
    <div className="relative h-[140px] flex items-center justify-center">
      {/* Desktop line — a faint "desktop" strip */}
      <div className="absolute inset-x-0 bottom-8 h-px bg-white/[0.06]" />

      {/* Keystroke chip on the left */}
      <motion.div
        initial={{ opacity: 0, y: 8 }}
        animate={{ opacity: [0, 1, 1, 0.7], y: 0 }}
        transition={{ duration: 6, times: [0, 0.1, 0.85, 1], ease: EASE_OUT }}
        className="absolute left-2 bottom-6 flex items-center gap-1.5"
      >
        <span className="px-2 py-1 rounded-md border border-white/15 bg-white/[0.04] text-[10px] text-white/55">
          ⌘⇧Space
        </span>
      </motion.div>

      {/* Overlay window fading in from center, then out */}
      <motion.div
        initial={{ opacity: 0, scale: 0.92, y: 10 }}
        animate={{ opacity: [0, 0, 1, 1, 0], scale: [0.92, 0.92, 1, 1, 0.96], y: [10, 10, 0, 0, -4] }}
        transition={{ duration: 6, times: [0, 0.15, 0.3, 0.85, 1], ease: EASE_OUT }}
        className="relative w-[240px] rounded-xl border border-white/12 bg-[#151515]/90 backdrop-blur-md shadow-[0_24px_60px_rgba(0,0,0,0.5)] overflow-hidden"
      >
        <div className="h-6 bg-[#1c1c1c] border-b border-white/[0.06] flex items-center px-3">
          <div className="flex gap-1">
            <span className="w-1.5 h-1.5 rounded-full bg-white/20" />
            <span className="w-1.5 h-1.5 rounded-full bg-white/20" />
            <span className="w-1.5 h-1.5 rounded-full bg-white/20" />
          </div>
        </div>
        <div className="p-3.5">
          <div className="flex items-center gap-2 mb-2.5">
            <span className="w-1.5 h-1.5 rounded-full bg-green-400/60" />
            <span className="text-[9px] text-white/40">listening</span>
          </div>
          <motion.div
            initial={{ opacity: 0 }}
            animate={{ opacity: 1 }}
            transition={{ delay: 0.4, duration: 0.4 }}
            className="text-[11px] text-white/70"
          >
            <span className="text-white/45 mr-1.5">❯</span>
            review the auth middleware
            <motion.span
              animate={{ opacity: [1, 0] }}
              transition={{ repeat: Infinity, duration: 0.8 }}
              className="inline-block w-[6px] h-[11px] bg-white/50 ml-0.5 align-middle"
            />
          </motion.div>
        </div>
      </motion.div>
    </div>
  );
}

/* ════════════════════════════════════════════════════════════
   Scene 2 — Route: the routing demo (request → pills → selection → result)
   ════════════════════════════════════════════════════════════ */

const AGENTS = ["Claude Code", "Codex", "Ollama", "Gemini"];
const ROUTE_TURNS = [
  { request: "review auth middleware for timing attacks", pick: 0, result: "no constant-time violation · 4 files" },
  { request: "scaffold a rate-limiter from the openapi spec", pick: 1, result: "token_bucket.go + 3 tests" },
  { request: "explain this stack trace offline", pick: 2, result: "nil map write at handler.go:42" },
];

function RouteVisual({ active }: { active: boolean }) {
  const [turn, setTurn] = useState(0);
  const [phase, setPhase] = useState<"req" | "route" | "work" | "res">("req");

  useEffect(() => {
    if (!active) return;
    let mounted = true;
    let timer: ReturnType<typeof setTimeout>;
    const seq: ["req" | "route" | "work" | "res", number][] = [["req", 800], ["route", 1300], ["work", 1300], ["res", 900]];
    const run = () => {
      let i = 0;
      const step = () => {
        if (!mounted) return;
        if (i < seq.length) {
          const [ph, dur] = seq[i];
          setPhase(ph);
          timer = setTimeout(() => { i += 1; step(); }, dur);
        } else {
          timer = setTimeout(() => {
            setTurn((p) => (p + 1) % ROUTE_TURNS.length);
            setPhase("req");
            timer = setTimeout(run, 400);
          }, 400);
        }
      };
      step();
    };
    timer = setTimeout(run, 300);
    return () => { mounted = false; clearTimeout(timer); };
  }, [active]);

  const t = ROUTE_TURNS[turn];
  const showPick = phase === "work" || phase === "res";

  return (
    <div className="space-y-3">
      {/* request */}
      <div className="text-[12px] text-white/80">
        <span className="text-white/45 mr-2">❯</span>
        {phase === "req" ? <TypeLine text={t.request} active={phase === "req"} /> : t.request}
      </div>

      {(phase !== "req") && (
        <div className="pl-5 space-y-2.5">
          <p className="text-[10px] text-white/30">routing to agent</p>
          <div className="flex flex-wrap gap-1.5 relative">
            {AGENTS.map((a, i) => {
              const picked = showPick && i === t.pick;
              return (
                <motion.span
                  key={a}
                  initial={{ opacity: 0 }}
                  animate={{ opacity: picked ? 1 : 0.3 }}
                  transition={{ duration: 0.3, ease: EASE_OUT }}
                  className="px-2 py-0.5 rounded-md border text-[10.5px] whitespace-nowrap"
                  style={{ borderColor: picked ? "rgba(255,255,255,0.28)" : "rgba(255,255,255,0.07)" }}
                >
                  {a}
                </motion.span>
              );
            })}
            {phase === "route" && (
              <motion.div
                initial={{ scaleX: 0, opacity: 0.5 }}
                animate={{ scaleX: 1, opacity: 0 }}
                transition={{ duration: 1.3, ease: "easeInOut" }}
                className="absolute left-0 right-0 top-1/2 -translate-y-1/2 h-px bg-gradient-to-r from-transparent via-white/30 to-transparent origin-left"
              />
            )}
          </div>

          {/* progress */}
          {phase === "work" && (
            <div className="flex items-center gap-2.5 pt-0.5">
              <div className="relative h-px w-28 bg-white/[0.08] overflow-hidden rounded-full">
                <motion.div
                  initial={{ scaleX: 0 }}
                  animate={{ scaleX: 1 }}
                  transition={{ duration: 1.3, ease: "easeInOut" }}
                  className="absolute inset-0 bg-white/55 origin-left rounded-full"
                />
              </div>
              <span className="text-[10px] text-white/30">running</span>
            </div>
          )}

          {/* result */}
          {phase === "res" && (
            <motion.div
              initial={{ opacity: 0, y: 4 }}
              animate={{ opacity: 1, y: 0 }}
              transition={{ duration: 0.3, ease: EASE_OUT }}
              className="flex items-center gap-2 pt-0.5"
            >
              <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="3" strokeLinecap="round" strokeLinejoin="round" className="text-white/55 shrink-0">
                <path d="M5 12l5 5 9-10" />
              </svg>
              <span className="text-[11px] text-white/55">{t.result}</span>
            </motion.div>
          )}
        </div>
      )}
    </div>
  );
}

/* ════════════════════════════════════════════════════════════
   Scene 3 — Guard: a gatekeeper verdict on an action
   ════════════════════════════════════════════════════════════ */

const GUARD_ACTIONS = [
  { action: "click ‘Send Email’ in Gmail", verdict: "ask", label: "Network · requires consent" },
  { action: "edit file in VS Code", verdict: "allow", label: "Write · approved app" },
  { action: "rm -rf ~/Documents", verdict: "deny", label: "Destructive · blocked" },
];

const VERDICT_STYLE = {
  allow: { dot: "bg-green-400/70", text: "text-green-400/80", label: "ALLOW" },
  ask: { dot: "bg-amber-400/70", text: "text-amber-400/80", label: "ASK" },
  deny: { dot: "bg-red-400/70", text: "text-red-400/80", label: "DENY" },
} as const;

function GuardVisual() {
  const [i, setI] = useState(0);
  useEffect(() => {
    const timer = setInterval(() => setI((p) => (p + 1) % GUARD_ACTIONS.length), 2000);
    return () => clearInterval(timer);
  }, []);
  const a = GUARD_ACTIONS[i];
  const v = VERDICT_STYLE[a.verdict as keyof typeof VERDICT_STYLE];

  return (
    <div className="space-y-2.5">
      <p className="text-[10px] text-white/30">proposed action</p>
      <AnimatePresence mode="wait">
        <motion.div
          key={`act-${i}`}
          initial={{ opacity: 0, x: -8 }}
          animate={{ opacity: 1, x: 0 }}
          exit={{ opacity: 0, x: 8 }}
          transition={{ duration: 0.35, ease: EASE_OUT }}
          className="rounded-lg border border-white/[0.08] bg-white/[0.02] px-3.5 py-3"
        >
          <div className="flex items-center justify-between gap-3">
            <span className="text-[12px] text-white/75 truncate">{a.action}</span>
            <motion.span
              initial={{ scale: 0.8, opacity: 0 }}
              animate={{ scale: 1, opacity: 1 }}
              transition={{ delay: 0.15, duration: 0.3, ease: EASE_OUT }}
              className={`flex items-center gap-1.5 shrink-0 ${v.text}`}
            >
              <span className={`w-1.5 h-1.5 rounded-full ${v.dot}`} />
              <span className="text-[10px] font-semibold tracking-wider">{v.label}</span>
            </motion.span>
          </div>
          <p className="mt-2 text-[10px] text-white/30">{a.label}</p>
        </motion.div>
      </AnimatePresence>

      {/* Gatekeeper footer */}
      <div className="flex items-center justify-between pt-1">
        <span className="text-[10px] text-white/30">gatekeeper</span>
        <span className="text-[10px] text-white/25">deterministic · no model</span>
      </div>
    </div>
  );
}

/* ════════════════════════════════════════════════════════════
   Scene 4 — Yours: local + locked + no telemetry
   ════════════════════════════════════════════════════════════ */

const YOURS_ROWS = [
  { label: "Memory", value: "~/.condura/memory.db", state: "local" },
  { label: "API keys", value: "encrypted, on disk", state: "locked" },
  { label: "Audit log", value: "HMAC-chained, 24h", state: "local" },
  { label: "Telemetry", value: "off by default", state: "none" },
];

const STATE_BADGE = {
  local: { text: "text-white/45", icon: "local" },
  locked: { text: "text-white/45", icon: "locked" },
  none: { text: "text-white/35", icon: "none" },
} as const;

function YoursVisual() {
  return (
    <div className="space-y-1.5">
      {YOURS_ROWS.map((r, idx) => {
        const badge = STATE_BADGE[r.state as keyof typeof STATE_BADGE];
        return (
          <motion.div
            key={r.label}
            initial={{ opacity: 0, y: 6 }}
            animate={{ opacity: 1, y: 0 }}
            transition={{ delay: 0.1 * idx + 0.1, duration: 0.35, ease: EASE_OUT }}
            className="flex items-center justify-between gap-3 rounded-md px-3 py-2 border border-white/[0.05] bg-white/[0.015]"
          >
            <div className="flex items-center gap-2.5 min-w-0">
              <StateGlyph kind={badge.icon} />
              <span className="text-[11px] text-white/70 shrink-0">{r.label}</span>
            </div>
            <span className="text-[10.5px] text-white/35 truncate">{r.value}</span>
          </motion.div>
        );
      })}
    </div>
  );
}

function StateGlyph({ kind }: { kind: "local" | "locked" | "none" }) {
  const cls = "text-white/45 shrink-0";
  if (kind === "local") {
    return (
      <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={cls}>
        <path d="M3 7l9-4 9 4-9 4-9-4z" />
        <path d="M3 7v6l9 4 9-4V7" opacity={0.5} />
      </svg>
    );
  }
  if (kind === "locked") {
    return (
      <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className={cls}>
        <rect x="5" y="11" width="14" height="9" rx="2" />
        <path d="M8 11V8a4 4 0 0 1 8 0v3" />
      </svg>
    );
  }
  return (
    <svg width="11" height="11" viewBox="0 0 24 24" fill="none" stroke="currentColor" strokeWidth="2" strokeLinecap="round" strokeLinejoin="round" className="text-white/30 shrink-0">
      <circle cx="12" cy="12" r="9" />
      <path d="M5 5l14 14" />
    </svg>
  );
}

/* ── Typewriter ── */
function TypeLine({ text, active }: { text: string; active: boolean }) {
  const [shown, setShown] = useState(active ? 0 : text.length);
  useEffect(() => {
    if (!active) { setShown(text.length); return; }
    setShown(0);
    let i = 0;
    const tick = () => { i += 1; setShown(i); if (i < text.length) timer = setTimeout(tick, 26); };
    let timer = setTimeout(tick, 26);
    return () => clearTimeout(timer);
  }, [active, text]);
  return <span>{text.slice(0, shown)}</span>;
}