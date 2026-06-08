import './style.css'
import App from './App.svelte'
import { initStores } from './lib/stores/init'

// Initialize runes-based stores before mounting so the components
// see populated state on first render.
initStores()

const app = new App({
  target: document.getElementById('app')!
})

export default app
