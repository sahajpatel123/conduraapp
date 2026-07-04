"use client";

import { type ReactNode } from "react";
import SiteDock from "@/components/shell/SiteDock";
import PageTransition from "@/components/shell/PageTransition";
import RouteProgress from "@/components/shell/RouteProgress";
import CommandPalette from "@/components/shell/CommandPalette";
import DynamicIsland from "@/components/motion/DynamicIsland";
import ToastStack from "@/components/motion/ToastStack";
import Cursor from "@/components/shell/Cursor";
import ScrollThread from "@/components/shell/ScrollThread";
import KbdHint from "@/components/shell/KbdHint";

export default function SiteShell({ children }: { children: ReactNode }) {
  return (
    <>
      <a
        href="#main"
        className="sr-only focus:not-sr-only focus:fixed focus:left-4 focus:top-4 focus:z-[300] focus:rounded-md focus:bg-[var(--color-ink)] focus:px-3 focus:py-2 focus:text-[var(--color-paper)]"
      >
        Skip to content
      </a>
      <Cursor />
      <ScrollThread />
      <KbdHint />
      <RouteProgress />
      <DynamicIsland />
      <CommandPalette />
      <ToastStack />
      <PageTransition>{children}</PageTransition>
      <SiteDock />
    </>
  );
}
