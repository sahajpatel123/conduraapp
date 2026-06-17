"use client";

import { useEffect, useRef, useState, type ReactNode, type MouseEvent } from "react";
import { motion, useMotionValue, useSpring } from "motion/react";
import { useReducedMotion } from "@/hooks/useReducedMotion";
import { springSnappy } from "@/lib/motion";

interface MagneticButtonProps {
  children: ReactNode;
  className?: string;
  onClick?: () => void;
  href?: string;
  type?: "button" | "submit";
  disabled?: boolean;
  "aria-label"?: string;
}

const MAGNET = 18;

export default function MagneticButton({
  children,
  className = "",
  onClick,
  href,
  type = "button",
  disabled,
  "aria-label": ariaLabel,
}: MagneticButtonProps) {
  const reduced = useReducedMotion();
  const ref = useRef<HTMLDivElement>(null);
  const x = useMotionValue(0);
  const y = useMotionValue(0);
  const sx = useSpring(x, springSnappy);
  const sy = useSpring(y, springSnappy);
  const [pressed, setPressed] = useState(false);

  const onMove = (e: MouseEvent) => {
    if (reduced || disabled || !ref.current) return;
    const rect = ref.current.getBoundingClientRect();
    const dx = e.clientX - (rect.left + rect.width / 2);
    const dy = e.clientY - (rect.top + rect.height / 2);
    x.set((dx / rect.width) * MAGNET);
    y.set((dy / rect.height) * MAGNET);
  };

  const reset = () => {
    x.set(0);
    y.set(0);
    setPressed(false);
  };

  const body = (
    <motion.span
      style={reduced ? undefined : { x: sx, y: sy }}
      animate={{ scale: pressed ? 0.96 : 1 }}
      transition={{ duration: 0.12 }}
      className={`relative inline-flex items-center justify-center gap-2 ${className}`}
    >
      {children}
    </motion.span>
  );

  if (href) {
    return (
      <div
        ref={ref}
        className="inline-block"
        onMouseMove={onMove}
        onMouseLeave={reset}
        onMouseDown={() => setPressed(true)}
        onMouseUp={() => setPressed(false)}
      >
        <a href={href} aria-label={ariaLabel} className="block">
          {body}
        </a>
      </div>
    );
  }

  return (
    <div
      ref={ref}
      className="inline-block"
      onMouseMove={onMove}
      onMouseLeave={reset}
      onMouseDown={() => setPressed(true)}
      onMouseUp={() => setPressed(false)}
    >
      <button
        type={type}
        disabled={disabled}
        aria-label={ariaLabel}
        onClick={onClick}
        className="block disabled:cursor-not-allowed disabled:opacity-50"
      >
        {body}
      </button>
    </div>
  );
}
