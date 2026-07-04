"use client";

import { useSyncExternalStore } from "react";
import type { PlatformKey } from "@/lib/site";

function detectPlatform(): PlatformKey {
  if (typeof navigator === "undefined") return "mac";
  const ua = navigator.userAgent.toLowerCase();
  const platform = navigator.platform?.toLowerCase() ?? "";
  if (platform.includes("win") || ua.includes("windows")) return "windows";
  if (platform.includes("linux") || ua.includes("linux")) return "linux";
  return "mac";
}

export function usePlatform(): PlatformKey {
  return useSyncExternalStore(() => () => {}, detectPlatform, () => "mac");
}
