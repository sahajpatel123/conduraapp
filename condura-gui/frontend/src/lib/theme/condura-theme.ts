/**
 * Condura theme — single source of truth for light / dark / system.
 *
 * Only document.documentElement carries data-mode (resolved light|dark).
 * Descendant wrappers must NOT set their own data-mode or tokens desync.
 */

export type ThemePreference = 'light' | 'dark' | 'system'
export type ResolvedTheme = 'light' | 'dark'

const STORAGE_KEY = 'condura-theme'
const VALID: readonly ThemePreference[] = ['light', 'dark', 'system']

type ThemeListener = (resolved: ResolvedTheme, preference: ThemePreference) => void

const listeners = new Set<ThemeListener>()
let systemListenerInstalled = false

function systemPrefersDark(): boolean {
  try {
    return matchMedia('(prefers-color-scheme: dark)').matches
  } catch {
    return false
  }
}

export function readPreference(): ThemePreference {
  try {
    const raw = localStorage.getItem(STORAGE_KEY)
    if (raw && (VALID as readonly string[]).includes(raw)) {
      return raw as ThemePreference
    }
  } catch {
    /* private mode / file:// */
  }
  return 'light'
}

export function resolvePreference(preference: ThemePreference): ResolvedTheme {
  if (preference === 'dark') return 'dark'
  if (preference === 'light') return 'light'
  return systemPrefersDark() ? 'dark' : 'light'
}

export function getResolvedTheme(): ResolvedTheme {
  if (typeof document === 'undefined') return 'light'
  const fromDom = document.documentElement.dataset.mode
  if (fromDom === 'light' || fromDom === 'dark') return fromDom
  return resolvePreference(readPreference())
}

export function applyTheme(preference: ThemePreference = readPreference()): ResolvedTheme {
  const resolved = resolvePreference(preference)
  if (typeof document !== 'undefined') {
    document.documentElement.dataset.mode = resolved
  }
  for (const fn of listeners) fn(resolved, preference)
  return resolved
}

export function setThemePreference(preference: ThemePreference): ResolvedTheme {
  try {
    localStorage.setItem(STORAGE_KEY, preference)
  } catch {
    /* ignore */
  }
  return applyTheme(preference)
}

export function setResolvedTheme(resolved: ResolvedTheme): ResolvedTheme {
  return setThemePreference(resolved)
}

export function toggleLightDark(): ResolvedTheme {
  const next: ResolvedTheme = getResolvedTheme() === 'dark' ? 'light' : 'dark'
  return setThemePreference(next)
}

export function onThemeChange(fn: ThemeListener): () => void {
  listeners.add(fn)
  return () => listeners.delete(fn)
}

/** Call once at boot (main.ts) before first paint. */
export function initTheme(): ResolvedTheme {
  const resolved = applyTheme(readPreference())

  if (typeof window === 'undefined' || systemListenerInstalled) return resolved
  systemListenerInstalled = true

  try {
    const mql = matchMedia('(prefers-color-scheme: dark)')
    mql.addEventListener('change', () => {
      if (readPreference() !== 'system') return
      applyTheme('system')
    })
  } catch {
    /* ignore */
  }

  return resolved
}
