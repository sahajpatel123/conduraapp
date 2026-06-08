import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// Vite config for the Synaptic GUI frontend.
// Wails auto-generates frontend/wailsjs/ on each `wails build`,
// so we don't need a custom output dir.
export default defineConfig({
  plugins: [svelte()],
  build: {
    target: 'esnext',
    sourcemap: true
  }
})
