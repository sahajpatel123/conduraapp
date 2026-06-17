"use client";

import { AnimatePresence, motion } from "motion/react";
import { type ReactNode, useEffect } from "react";
import { springSoft } from "@/lib/motion";

interface MorphingModalProps {
  open: boolean;
  onClose: () => void;
  title: string;
  children: ReactNode;
  footer?: ReactNode;
}

export default function MorphingModal({
  open,
  onClose,
  title,
  children,
  footer,
}: MorphingModalProps) {
  useEffect(() => {
    if (!open) return;
    const onKey = (e: KeyboardEvent) => {
      if (e.key === "Escape") onClose();
    };
    document.body.style.overflow = "hidden";
    window.addEventListener("keydown", onKey);
    return () => {
      document.body.style.overflow = "";
      window.removeEventListener("keydown", onKey);
    };
  }, [open, onClose]);

  return (
    <AnimatePresence>
      {open && (
        <motion.div
          className="fixed inset-0 z-[240] flex items-end justify-center p-4 sm:items-center"
          initial={{ opacity: 0 }}
          animate={{ opacity: 1 }}
          exit={{ opacity: 0 }}
        >
          <button
            type="button"
            aria-label="Close dialog"
            className="absolute inset-0 bg-black/65 backdrop-blur-md"
            onClick={onClose}
          />
          <motion.div
            layoutId="condura-modal-surface"
            role="dialog"
            aria-modal="true"
            aria-labelledby="condura-modal-title"
            initial={{ opacity: 0, y: 24, scale: 0.96, borderRadius: 28 }}
            animate={{ opacity: 1, y: 0, scale: 1, borderRadius: 20 }}
            exit={{ opacity: 0, y: 16, scale: 0.98, borderRadius: 28 }}
            transition={springSoft}
            className="relative z-10 w-full max-w-lg overflow-hidden border border-white/[0.08] bg-[#111113]/95 shadow-[0_30px_120px_rgba(0,0,0,0.55)] backdrop-blur-xl"
          >
            <div className="border-b border-white/[0.07] px-5 py-4">
              <h2 id="condura-modal-title" className="text-base font-semibold text-white">
                {title}
              </h2>
            </div>
            <div className="px-5 py-4 text-sm leading-relaxed text-white/55">{children}</div>
            {footer && (
              <div className="flex items-center justify-end gap-2 border-t border-white/[0.07] px-5 py-4">
                {footer}
              </div>
            )}
          </motion.div>
        </motion.div>
      )}
    </AnimatePresence>
  );
}
