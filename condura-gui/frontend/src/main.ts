import './lib/shell/meridian/meridian.css'
import { initTheme } from './lib/theme/condura-theme'
import { mount } from 'svelte'
import { MeridianShell } from './lib/shell/meridian'

// Meridian only — Inkboard / Living Paper / Lumen are not mounted.
initTheme()

function bootstrap(): void {
  const target = document.getElementById('app')
  if (!target) {
    document.body.innerHTML =
      '<pre style="color:#2F5BFF;padding:24px;font-family:monospace">#app missing</pre>'
    return
  }
  try {
    mount(MeridianShell, { target })
  } catch (e) {
    const name = e instanceof Error ? e.name : 'Error'
    const message = e instanceof Error ? e.message : String(e)
    const stack = e instanceof Error ? e.stack || '' : ''
    document.body.innerHTML =
      '<pre style="color:#2F5BFF;padding:24px;font-family:monospace;white-space:pre-wrap;">' +
      name + ': ' + message + '\n\n' + stack + '</pre>'
  }
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', bootstrap, { once: true })
} else {
  bootstrap()
}
