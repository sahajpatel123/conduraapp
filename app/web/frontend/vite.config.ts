import { defineConfig } from 'vite'
import { svelte } from '@sveltejs/vite-plugin-svelte'
import path from 'node:path'

// Vite config for the Condura GUI frontend.
// Wails auto-generates frontend/wailsjs/ on each `wails build`,
// so we don't need a custom output dir.
//
// base: './' produces relative asset URLs (./assets/...) instead of
// absolute (/assets/...). The Wails asset server serves the embed
// FS at a non-standard origin, and absolute paths can fail to
// resolve depending on the Wails version and macOS WebView config.
// Relative paths work everywhere.
//
// Path aliases: $tokens and $components. Both resolve under src/lib/*
// so the literal reading of "$tokens/primitives.css" is "the primitives
// file in the tokens folder under src/lib".
export default defineConfig({
  base: './',
  plugins: [svelte()],
  resolve: {
    alias: {
      $tokens: path.resolve(__dirname, 'src/lib/tokens'),
      $components: path.resolve(__dirname, 'src/lib/components'),
    },
  },
  build: {
    target: 'esnext',
    sourcemap: true
  }
})