import './lib/shell/meridian/meridian.css'
import { initTheme } from './lib/theme/condura-theme'
import { mount } from 'svelte'
import { MeridianShell } from './lib/shell/meridian'

// Meridian only — Inkboard / Living Paper / Lumen are not mounted.
initTheme()

function showBootFail(err: unknown): void {
  const panel = document.getElementById('boot-fail')
  const detail = document.getElementById('boot-fail-detail')
  const name = err instanceof Error ? err.name : 'Error'
  const message = err instanceof Error ? err.message : String(err)
  const stack = err instanceof Error ? err.stack || '' : ''
  if (panel) panel.classList.add('show')
  if (detail) detail.textContent = `${name}: ${message}\n\n${stack}`
  else {
    // Fallback when neither the boot-fail panel nor its detail node
    // are present in the DOM. Build the element via the safe DOM API
    // (createElement + textContent) — never innerHTML with string
    // concatenation, since err.message can flow from Wails IPC and
    // could include attacker-controlled bytes in a future caller.
    const pre = document.createElement('pre')
    pre.style.color = '#2F5BFF'
    pre.style.padding = '24px'
    pre.style.fontFamily = 'monospace'
    pre.style.whiteSpace = 'pre-wrap'
    pre.textContent = `${name}: ${message}\n\n${stack}`
    document.body.appendChild(pre)
  }
}

function bootstrap(): void {
  const target = document.getElementById('app')
  if (!target) {
    showBootFail(new Error('#app missing'))
    return
  }
  try {
    mount(MeridianShell, { target })
    document.getElementById('boot-fail')?.remove()
  } catch (e) {
    showBootFail(e)
  }
}

if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', bootstrap, { once: true })
} else {
  bootstrap()
}
