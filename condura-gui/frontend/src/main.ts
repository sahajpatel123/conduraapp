import './lib/condura/condura.css'
import './lib/components/living/living-paper.css'
import { mount } from 'svelte'
import { LivingPaperShell } from './lib/shell'

// Theme — light by default for every new user. Applied before paint so there
// is no flash. The only modes the user can pick are 'light' | 'dark' | 'system'.
//   - 'light' / 'dark'  : forced. Never auto-switches to the other.
//   - 'system'          : follows the OS via matchMedia, live.
//   - unset / garbage   : treated as 'light' (the default), NOT 'system'.
// The whole thing is wrapped in try/catch for SSR / file:// safety where
// localStorage and matchMedia can throw.
type ThemeMode = 'light' | 'dark' | 'system';
type ResolvedMode = 'light' | 'dark';

const STORAGE_KEY = 'condura-theme';
const VALID: ReadonlyArray<ThemeMode> = ['light', 'dark', 'system'];

function readStored(): ThemeMode {
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (raw && (VALID as ReadonlyArray<string>).includes(raw)) return raw as ThemeMode;
  } catch {
    /* SSR / file:// / private mode */
  }
  return 'light'; // DEFAULT — not 'system'. A new user sees light.
}

function systemPrefersDark(): boolean {
  try {
    return matchMedia('(prefers-color-scheme: dark)').matches;
  } catch {
    return false;
  }
}

function resolve(stored: ThemeMode): ResolvedMode {
  if (stored === 'dark') return 'dark';
  if (stored === 'light') return 'light';
  return systemPrefersDark() ? 'dark' : 'light';
}

(function applyTheme() {
  const stored = readStored();
  document.documentElement.dataset.mode = resolve(stored);

  // Live-follow OS changes — but ONLY when the user opted into 'system'.
  // If they picked 'light' or 'dark', we never auto-switch on them.
  try {
    const mql = matchMedia('(prefers-color-scheme: dark)');
    mql.addEventListener('change', (e) => {
      let cur: ThemeMode = 'light';
      try {
        const raw = localStorage.getItem(STORAGE_KEY);
        if (raw && (VALID as ReadonlyArray<string>).includes(raw)) cur = raw as ThemeMode;
      } catch {
        /* ignore */
      }
      if (cur !== 'system') return; // forced mode — respect the override
      document.documentElement.dataset.mode = e.matches ? 'dark' : 'light';
    });
  } catch {
    /* ignore */
  }
})();

// Living Paper shell — replaces the v1/Condura shells. The warm paper-and-ink
// aesthetic draws from the Synapse Garden brand (paper · ink · synapse green ·
// pollen amber · sky blue). The old Shell.svelte and App.svelte stay on disk,
// unmounted, as safety nets. The daemon contract (ipc + stores) is unchanged —
// this is purely the view layer.

// Svelte 5 uses mount() instead of the legacy `new Component()` API.
// The legacy constructor may not survive minification in the Wails
// bundled environment, causing `effect_orphan` errors.

function bootstrap(): void {
  const target = document.getElementById('app')
  if (!target) {
    document.body.innerHTML =
      '<pre style="color:#f87171;padding:24px;font-family:monospace">' +
      '#app element not found in DOM!</pre>'
    return
  }
  try {
    mount(LivingPaperShell, { target })
  } catch (e) {
    const name = e instanceof Error ? e.name : 'Error'
    const message = e instanceof Error ? e.message : String(e)
    const stack = e instanceof Error ? e.stack || '' : ''
    document.body.innerHTML =
      '<pre style="color:#f87171;padding:24px;font-family:monospace;white-space:pre-wrap;">' +
      name + ': ' + message + '\n\n' + stack + '</pre>'
  }
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', bootstrap, { once: true })
} else {
  bootstrap()
}