import './lib/condura/condura.css'
import './lib/components/living/living-paper.css'
import { initTheme } from './lib/theme/condura-theme'
import { mount } from 'svelte'
import { LivingPaperShell } from './lib/shell'

// Theme — light by default. Resolved mode is applied to <html data-mode>
// before paint via initTheme(). See lib/theme/condura-theme.ts.
initTheme()

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
