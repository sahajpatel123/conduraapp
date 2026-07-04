"use client";

import { type ReactNode } from "react";
import { ToastProvider } from "@/context/ToastContext";
import { IslandProvider } from "@/context/IslandContext";

export default function Providers({ children }: { children: ReactNode }) {
  return (
    <ToastProvider>
      <IslandProvider>
        {children}
      </IslandProvider>
    </ToastProvider>
  );
}
