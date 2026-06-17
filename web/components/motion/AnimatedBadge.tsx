"use client";

import { motion } from "motion/react";
import { useReducedMotion } from "@/hooks/useReducedMotion";

interface AnimatedBadgeProps {
  children: React.ReactNode;
  tone?: "mint" | "violet" | "neutral";
  pulse?: boolean;
  className?: string;
}

const tones = {
  mint: "border-white/20 bg-white/[0.06] text-white/60",
  violet: "border-[#8b7dff]/25 bg-[#8b7dff]/10 text-[#c4bbff]",
  neutral: "border-white/10 bg-white/[0.04] text-white/60",
};

export default function AnimatedBadge({
  children,
  tone = "neutral",
  pulse = false,
  className = "",
}: AnimatedBadgeProps) {
  const reduced = useReducedMotion();

  return (
    <motion.span
      initial={reduced ? false : { opacity: 0, scale: 0.9 }}
      animate={{ opacity: 1, scale: 1 }}
      transition={{ type: "spring", stiffness: 400, damping: 24 }}
      className={`inline-flex items-center gap-1.5 rounded-full border px-2.5 py-1 text-[11px] font-medium uppercase tracking-wide ${tones[tone]} ${className}`}
    >
      {pulse && !reduced && (
        <motion.span
          className="h-1.5 w-1.5 rounded-full bg-current"
          animate={{ scale: [1, 1.35, 1], opacity: [1, 0.55, 1] }}
          transition={{ duration: 1.8, repeat: Infinity }}
        />
      )}
      {children}
    </motion.span>
  );
}
