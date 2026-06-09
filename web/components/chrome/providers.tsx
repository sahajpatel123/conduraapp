"use client";

/*
  Client boundary for the app shell: motion config (respects the user's
  reduced-motion preference globally) and the ⌘K palette state.
*/
import { domAnimation, LazyMotion, MotionConfig } from "motion/react";
import type { ReactNode } from "react";
import { PaletteProvider } from "./palette";

export function Providers({ children }: { children: ReactNode }) {
  return (
    <LazyMotion features={domAnimation} strict>
      <MotionConfig reducedMotion="user">
        <PaletteProvider>{children}</PaletteProvider>
      </MotionConfig>
    </LazyMotion>
  );
}
