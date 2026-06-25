"use client";

import { AnimatePresence, motion } from "motion/react";
import { useToast } from "@/context/ToastContext";
import { springSnappy } from "@/lib/motion";

const toneBorder = {
  default: "border-[rgba(20,17,11,0.12)]",
  success: "border-[rgba(11,61,46,0.35)]",
  error: "border-[rgba(163,49,42,0.4)]",
};

const toneDot = {
  default: "bg-[var(--color-ink-faint)]",
  success: "bg-[var(--color-synapse)]",
  error: "bg-[var(--color-danger)]",
};

export default function ToastStack() {
  const { toasts, dismiss } = useToast();

  return (
    <div
      className="pointer-events-none fixed right-4 top-20 z-[250] flex w-[min(100%,320px)] flex-col gap-2"
      aria-live="polite"
      aria-relevant="additions"
    >
      <AnimatePresence initial={false}>
        {toasts.map((toast) => (
          <motion.div
            key={toast.id}
            layout
            initial={{ opacity: 0, x: 24, scale: 0.96 }}
            animate={{ opacity: 1, x: 0, scale: 1 }}
            exit={{ opacity: 0, x: 24, scale: 0.95 }}
            transition={springSnappy}
            className={`pointer-events-auto rounded-xl border bg-[var(--color-paper-warm)] p-3 shadow-[var(--shadow-card)] backdrop-blur-xl ${toneBorder[toast.tone ?? "default"]}`}
          >
            <div className="flex items-start justify-between gap-3">
              <div className="flex items-start gap-2.5">
                <span className={`mt-1.5 h-1.5 w-1.5 shrink-0 rounded-full ${toneDot[toast.tone ?? "default"]}`} />
                <div>
                  <p className="text-sm font-medium text-[var(--color-ink)]">{toast.title}</p>
                  {toast.description && (
                    <p className="mt-0.5 text-xs text-[var(--color-ink-mute)]">{toast.description}</p>
                  )}
                </div>
              </div>
              <button
                type="button"
                onClick={() => dismiss(toast.id)}
                className="text-xs text-[var(--color-ink-faint)] hover:text-[var(--color-ink)]"
                aria-label="Dismiss notification"
              >
                ✕
              </button>
            </div>
          </motion.div>
        ))}
      </AnimatePresence>
    </div>
  );
}
