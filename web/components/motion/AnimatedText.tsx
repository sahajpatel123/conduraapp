"use client";

import { motion } from "motion/react";
import { useReducedMotion } from "@/hooks/useReducedMotion";
import { EASE_OUT } from "@/lib/motion";

interface AnimatedTextProps {
  text: string;
  className?: string;
  as?: "h1" | "h2" | "p" | "span";
  delay?: number;
}

export default function AnimatedText({
  text,
  className = "",
  as: Tag = "span",
  delay = 0,
}: AnimatedTextProps) {
  const reduced = useReducedMotion();
  const words = text.split(" ");

  if (reduced) {
    return <Tag className={className}>{text}</Tag>;
  }

  return (
    <Tag className={className} aria-label={text}>
      {words.map((word, i) => (
        <motion.span
          key={`${word}-${i}`}
          initial={{ opacity: 0, y: 14, rotateX: -40 }}
          animate={{ opacity: 1, y: 0, rotateX: 0 }}
          transition={{ duration: 0.55, delay: delay + i * 0.06, ease: EASE_OUT }}
          className="mr-[0.28em] inline-block origin-bottom"
        >
          {word}
        </motion.span>
      ))}
    </Tag>
  );
}
