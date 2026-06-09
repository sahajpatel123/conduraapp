"use client";

/*
  Hydration-safe reduced-motion preference. The server (and first client
  render) assume motion is allowed; the real preference takes over right
  after hydration without a markup mismatch — something motion/react's
  own useReducedMotion cannot guarantee when the preference changes the
  rendered tree.
*/
import { useSyncExternalStore } from "react";

const QUERY = "(prefers-reduced-motion: reduce)";

let mql: MediaQueryList | undefined;

function getMql(): MediaQueryList {
  mql ??= window.matchMedia(QUERY);
  return mql;
}

function subscribe(callback: () => void) {
  const list = getMql();
  list.addEventListener("change", callback);
  return () => list.removeEventListener("change", callback);
}

export function usePrefersReducedMotion(): boolean {
  return useSyncExternalStore(
    subscribe,
    () => getMql().matches,
    () => false,
  );
}
