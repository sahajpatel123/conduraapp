"use client";

import { type ReactNode } from "react";
import SiteDock from "@/components/shell/SiteDock";
import CommandPalette from "@/components/shell/CommandPalette";
import DynamicIsland from "@/components/motion/DynamicIsland";
import ToastStack from "@/components/motion/ToastStack";

export default function SiteShell({ children }: { children: ReactNode }) {
  return (
    <>
      <a
        href="#main"
        className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-[300] focus:rounded-md focus:bg-white focus:px-3 focus:py-2 focus:text-black"
      >
        Skip to content
      </a>
      <DynamicIsland />
      <CommandPalette />
      <ToastStack />
      {children}
      <SiteDock />
    </>
  );
}
