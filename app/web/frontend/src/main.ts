import './style.css'
import { mount } from 'svelte'
import { initTheme } from '$tokens/themes'
import App from './App.svelte'

initTheme()

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
    mount(App, { target })
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
