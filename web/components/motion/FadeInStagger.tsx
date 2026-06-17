"use client";

import { motion } from "motion/react";
import { type ReactNode } from "react";

interface FadeInStaggerProps {
  children: ReactNode;
}

export default function FadeInStagger({ children }: FadeInStaggerProps) {
  return (
    <motion.div
      initial={{ opacity: 0, y: 20 }}
      animate={{ opacity: 1, y: 0 }}
      transition={{ duration: 0.8, ease: [0.16, 1, 0.3, 1] }}
      className="relative z-10"
    >
      {/* Subtle glowing accent line on the left */}
      <div className="absolute -left-8 top-0 bottom-0 w-[1px] bg-gradient-to-b from-white/0 via-white/20 to-white/0 hidden md:block" />
      {children}
    </motion.div>
  );
}
