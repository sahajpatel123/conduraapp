import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'

// Vite config for the Synaptic GUI frontend.
// Wails auto-generates frontend/wailsjs/ on each `wails build`,
// so we don't need a custom output dir.
//
// base: './' produces relative asset URLs (./assets/...) instead of
// absolute (/assets/...). The Wails asset server serves the embed
// FS at a non-standard origin, and absolute paths can fail to
// resolve depending on the Wails version and macOS WebView config.
// Relative paths work everywhere.
export default defineConfig({
  base: './',
  plugins: [svelte()],
  build: {
    target: 'esnext',
    sourcemap: true
  }
})
