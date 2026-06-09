import './style.css'
import App from './App.svelte'
import { initStores } from './lib/stores/init'

// Defer the Svelte mount to the next macrotask. This works around a
// Svelte 5 + Wails interaction where creating a component during
// module evaluation (which is what a <script type="module"> does)
// triggers `effect_orphan` because the reactive context isn't
// fully set up at top-level script time.
//
// requestAnimationFrame is the cleanest option — it runs after the
// browser has finished parsing and layout, and after any pending
// microtasks (including Svelte 5's internal setup). The window
// is also fully visible by then, so the mount is visible.
function mount(): void {
  const target = document.getElementById('app')
  if (!target) {
    document.body.innerHTML =
      '<pre style="color:#f87171;padding:24px;font-family:monospace">' +
      '#app element not found in DOM!</pre>'
    return
  }
  try {
    new App({ target })
  } catch (e) {
    const msg = e instanceof Error ? e.stack : String(e)
    document.body.innerHTML =
      '<pre style="color:#f87171;padding:24px;font-family:monospace;white-space:pre-wrap;">' +
      'Svelte mount failed:\n' + msg + '</pre>'
  }
}

function bootstrap(): void {
  // Initialize the rune-based stores. We don't await — the app
  // shell renders immediately and the stores populate as the
  // daemon comes online.
  void initStores()
  // Defer the actual mount to the next frame.
  requestAnimationFrame(mount)
}

if (document.readyState === 'loading') {
  // DOM not yet parsed (shouldn't happen for module scripts, but
  // be defensive).
  document.addEventListener('DOMContentLoaded', bootstrap, { once: true })
} else {
  // DOM already ready — boot immediately on next frame.
  requestAnimationFrame(bootstrap)
}
