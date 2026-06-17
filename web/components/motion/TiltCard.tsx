"use client";

import { useRef, useState, type MouseEvent, type ReactNode } from "react";
import { motion, useMotionValue, useSpring, useTransform } from "motion/react";
import { useReducedMotion } from "@/hooks/useReducedMotion";
import { springSnappy } from "@/lib/motion";

interface TiltCardProps {
  children: ReactNode;
  className?: string;
  maxRotate?: number;
  glareOpacity?: number;
}

export default function TiltCard({
  children,
  className = "",
  maxRotate = 10,
  glareOpacity = 0.15,
}: TiltCardProps) {
  const reduced = useReducedMotion();
  const cardRef = useRef<HTMLDivElement>(null);
  const [hovering, setHovering] = useState(false);

  // Motion Values for rotation
  const rotateXVal = useMotionValue(0);
  const rotateYVal = useMotionValue(0);

  // Springs for smooth movement
  const rx = useSpring(rotateXVal, springSnappy);
  const ry = useSpring(rotateYVal, springSnappy);

  // Glare position
  const glareX = useMotionValue(50);
  const glareY = useMotionValue(50);
  const glareXSpring = useSpring(glareX, springSnappy);
  const glareYSpring = useSpring(glareY, springSnappy);

  const handleMouseMove = (e: MouseEvent<HTMLDivElement>) => {
    if (reduced || !cardRef.current) return;

    const rect = cardRef.current.getBoundingClientRect();
    const width = rect.width;
    const height = rect.height;

    // Mouse coordinates relative to card center (ranged -0.5 to 0.5)
    const mouseX = (e.clientX - rect.left) / width - 0.5;
    const mouseY = (e.clientY - rect.top) / height - 0.5;

    // Calculate rotation: rotateY corresponds to horizontal movement, rotateX to vertical
    rotateXVal.set(-mouseY * maxRotate * 2);
    rotateYVal.set(mouseX * maxRotate * 2);

    // Glare position percentage
    glareX.set(((e.clientX - rect.left) / width) * 100);
    glareY.set(((e.clientY - rect.top) / height) * 100);
  };

  const handleMouseEnter = () => {
    setHovering(true);
  };

  const handleMouseLeave = () => {
    setHovering(false);
    rotateXVal.set(0);
    rotateYVal.set(0);
    glareX.set(50);
    glareY.set(50);
  };

  const glareBackground = useTransform(
    [glareXSpring, glareYSpring],
    ([x, y]) =>
      `radial-gradient(circle 250px at ${x}% ${y}%, rgba(255, 255, 255, ${glareOpacity}), transparent)`
  );

  return (
    <div
      ref={cardRef}
      onMouseMove={handleMouseMove}
      onMouseEnter={handleMouseEnter}
      onMouseLeave={handleMouseLeave}
      className={`relative select-none ${className}`}
      style={{ perspective: 1200 }}
    >
      <motion.div
        style={
          reduced
            ? undefined
            : {
                rotateX: rx,
                rotateY: ry,
                transformStyle: "preserve-3d",
              }
        }
        className="w-full h-full relative"
      >
        {children}

        {/* Glare Overlay */}
        {!reduced && hovering && (
          <motion.div
            style={{
              background: glareBackground,
              transform: "translateZ(1px)",
            }}
            className="absolute inset-0 pointer-events-none rounded-[inherit] mix-blend-overlay z-30"
          />
        )}
      </motion.div>
    </div>
  );
}
