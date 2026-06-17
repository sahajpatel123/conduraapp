"use client";

import { useState, type ReactNode } from "react";
import { AnimatePresence, motion } from "motion/react";
import MagneticButton from "@/components/motion/MagneticButton";
import { springSnappy } from "@/lib/motion";

type ButtonState = "idle" | "loading" | "success" | "error";

interface StatefulButtonProps {
  idleLabel: ReactNode;
  loadingLabel?: ReactNode;
  successLabel?: ReactNode;
  errorLabel?: ReactNode;
  onAction: () => Promise<boolean>;
  className?: string;
}

export default function StatefulButton({
  idleLabel,
  loadingLabel = "Working…",
  successLabel = "Done",
  errorLabel = "Try again",
  onAction,
  className = "",
}: StatefulButtonProps) {
  const [state, setState] = useState<ButtonState>("idle");

  const run = async () => {
    if (state === "loading") return;
    setState("loading");
    const ok = await onAction();
    setState(ok ? "success" : "error");
    window.setTimeout(() => setState("idle"), ok ? 1800 : 2200);
  };

  const label =
    state === "loading"
      ? loadingLabel
      : state === "success"
        ? successLabel
        : state === "error"
          ? errorLabel
          : idleLabel;

  return (
    <MagneticButton
      onClick={run}
      disabled={state === "loading"}
      className={`min-w-[168px] rounded-full px-6 py-3.5 text-sm font-semibold transition-colors ${className}`}
    >
      <AnimatePresence mode="wait" initial={false}>
        <motion.span
          key={String(state)}
          initial={{ opacity: 0, y: 8, filter: "blur(4px)" }}
          animate={{ opacity: 1, y: 0, filter: "blur(0px)" }}
          exit={{ opacity: 0, y: -8, filter: "blur(4px)" }}
          transition={springSnappy}
          className="inline-flex items-center gap-2"
        >
          {state === "loading" && (
            <span className="h-3.5 w-3.5 animate-spin rounded-full border-2 border-current border-t-transparent" />
          )}
          {state === "success" && <span aria-hidden>✓</span>}
          {label}
        </motion.span>
      </AnimatePresence>
    </MagneticButton>
  );
}
