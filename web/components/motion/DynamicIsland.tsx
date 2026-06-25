"use client";

import { motion } from "motion/react";
import { useIsland } from "@/context/IslandContext";
import { useReducedMotion } from "@/hooks/useReducedMotion";
import { springSoft } from "@/lib/motion";

const phaseColor = {
  idle: "bg-[var(--color-ink-faint)]",
  listening: "bg-[var(--color-synapse)]",
  routing: "bg-[var(--color-synapse-glow)]",
  download: "bg-[var(--color-pollen)]",
};

export default function DynamicIsland() {
  const { state } = useIsland();
  const reduced = useReducedMotion();
  const expanded = state.phase !== "idle";
  if (!expanded) return null;

  return (
    <div className="pointer-events-none fixed left-1/2 top-[88px] z-[180] -translate-x-1/2">
      <motion.div
        layout
        animate={{ width: 280, height: 52, borderRadius: 22 }}
        transition={reduced ? { duration: 0 } : springSoft}
        className="flex items-center justify-center gap-2 border border-[rgba(20,17,11,0.12)] bg-[var(--color-paper)] px-4 shadow-[var(--shadow-float)] backdrop-blur-xl"
        aria-live="polite"
        aria-atomic="true"
      >
        <span
          className={`h-2 w-2 shrink-0 rounded-full ${phaseColor[state.phase]} ${
            state.phase === "listening" && !reduced ? "animate-pulse" : ""
          }`}
        />
        <motion.div layout className="min-w-0 text-center">
          <p className="truncate text-[12px] font-medium text-[var(--color-ink)]">{state.label}</p>
          {state.detail && expanded && (
            <p className="truncate text-[10px] text-[var(--color-ink-mute)]">{state.detail}</p>
          )}
        </motion.div>
      </motion.div>
    </div>
  );
}
