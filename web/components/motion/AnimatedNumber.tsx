"use client";

import { useEffect, useRef, useState } from "react";
import { motion, useSpring } from "motion/react";
import { useReducedMotion } from "@/hooks/useReducedMotion";
import { springSoft } from "@/lib/motion";

interface AnimatedNumberProps {
  value: number;
  suffix?: string;
  className?: string;
}

export default function AnimatedNumber({ value, suffix = "", className = "" }: AnimatedNumberProps) {
  const reduced = useReducedMotion();
  const [display, setDisplay] = useState(0);
  const spring = useSpring(0, springSoft);
  const started = useRef(false);

  useEffect(() => {
    if (reduced) {
      setDisplay(value);
      return;
    }
    const unsub = spring.on("change", (v) => setDisplay(Math.round(v)));
    if (!started.current) {
      started.current = true;
      spring.set(value);
    } else {
      spring.set(value);
    }
    return unsub;
  }, [value, spring, reduced]);

  return (
    <span className={`tabular-nums ${className}`}>
      {display}
      {suffix}
    </span>
  );
}
