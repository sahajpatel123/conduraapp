/*
 * Synaptic — Theme Lifecycle
 *
 * Owns the `data-mode` attribute on <html>. Reads the user's stored
 * preference, falls back to 'system' (which follows OS via media query),
 * and persists changes to localStorage (or daemon config in production).
 *
 * Locked: see docs/design-v1-redesign.md §4 + implementation step 1.
 */

export type Mode = 'light' | 'dark' | 'hc' | 'system';

const STORAGE_KEY = 'synaptic:mode';
const VALID_MODES: Mode[] = ['light', 'dark', 'hc', 'system'];

let current: Mode = 'system';
let listeners: Array<(mode: Mode) => void> = [];

/** Initialize the theme system. Call once on app boot, before first paint. */
export function initTheme(): Mode {
  if (typeof document === 'undefined') return current;

  const stored = readStored();
  current = stored ?? 'system';
  applyMode(current);
  return current;
}

/** Get the currently active mode. */
export function getMode(): Mode {
  return current;
}

/** Set the mode and persist. Notifies all listeners. */
export function setMode(mode: Mode): void {
  if (!VALID_MODES.includes(mode)) {
    console.warn(`[synaptic] invalid mode "${mode}" — ignoring`);
    return;
  }
  current = mode;
  applyMode(mode);
  persist(mode);
  listeners.forEach((fn) => fn(mode));
}

/** Toggle between light and dark. Skips 'system' and 'hc'. */
export function toggleLightDark(): Mode {
  const next: Mode = current === 'dark' ? 'light' : 'dark';
  setMode(next);
  return next;
}

/** Subscribe to mode changes. Returns an unsubscribe function. */
export function onModeChange(fn: (mode: Mode) => void): () => void {
  listeners.push(fn);
  return () => {
    listeners = listeners.filter((l) => l !== fn);
  };
}

/** Returns the resolved mode (for 'system', resolves to light or dark). */
export function getResolvedMode(): 'light' | 'dark' | 'hc' {
  if (current === 'system') {
    if (typeof window === 'undefined') return 'light';
    return window.matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light';
  }
  return current;
}

function applyMode(mode: Mode): void {
  if (typeof document === 'undefined') return;
  document.documentElement.setAttribute('data-mode', mode);
}

function readStored(): Mode | null {
  if (typeof localStorage === 'undefined') return null;
  try {
    const v = localStorage.getItem(STORAGE_KEY);
    if (v && VALID_MODES.includes(v as Mode)) return v as Mode;
  } catch {
    /* ignore (privacy mode, etc.) */
  }
  return null;
}

function persist(mode: Mode): void {
  if (typeof localStorage === 'undefined') return;
  try {
    localStorage.setItem(STORAGE_KEY, mode);
  } catch {
    /* ignore */
  }
}

/* ---------------------------------------------------------------------- *
 * Listen to OS preference changes when in 'system' mode.
 * ---------------------------------------------------------------------- */

if (typeof window !== 'undefined') {
  const mq = window.matchMedia('(prefers-color-scheme: dark)');
  mq.addEventListener?.('change', () => {
    if (current === 'system') {
      listeners.forEach((fn) => fn('system'));
    }
  });
}