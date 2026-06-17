"use client";

import { AnimatePresence, motion } from "motion/react";
import { useToast } from "@/context/ToastContext";
import { springSnappy } from "@/lib/motion";

const toneBorder = {
  default: "border-white/10",
  success: "border-white/20",
  error: "border-[#ff6b6b]/30",
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
            className={`pointer-events-auto rounded-xl border bg-[#111113]/95 p-3 shadow-xl backdrop-blur-xl ${toneBorder[toast.tone ?? "default"]}`}
          >
            <div className="flex items-start justify-between gap-3">
              <div>
                <p className="text-sm font-medium text-white">{toast.title}</p>
                {toast.description && (
                  <p className="mt-0.5 text-xs text-white/45">{toast.description}</p>
                )}
              </div>
              <button
                type="button"
                onClick={() => dismiss(toast.id)}
                className="text-xs text-white/35 hover:text-white/70"
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
