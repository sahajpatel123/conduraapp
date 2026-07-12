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
export default defineConfig(({ command }) => ({
  // Absolute base in `vite`/`serve` so Cursor Simple Browser / localhost
  // previews never resolve `./src/main.ts` against a bad path. Keep
  // relative `./` for Wails production embeds (asset server quirk).
  base: command === 'serve' ? '/' : './',
  plugins: [svelte()],
  resolve: {
    alias: {
      $tokens: path.resolve(__dirname, 'src/lib/tokens'),
      $components: path.resolve(__dirname, 'src/lib/components'),
      $lib: path.resolve(__dirname, 'src/lib'),
    },
  },
  server: {
    host: true, // 0.0.0.0 + localhost (::1) — Cursor browser often uses localhost
    port: 5173,
    strictPort: true,
    // Same-origin proxy so Chromium (Cursor Simple Browser / Playwright)
    // can reach the daemon. Direct :5173 → :7666 is blocked by Private
    // Network Access even when CORS reflects Origin.
    proxy: {
      '/api': { target: 'http://127.0.0.1:7666', changeOrigin: true },
      '/events': { target: 'http://127.0.0.1:7666', changeOrigin: true },
      '/sse-ticket': { target: 'http://127.0.0.1:7666', changeOrigin: true },
    },
  },
  build: {
    target: 'esnext',
    sourcemap: true,
    // outDir must be 'assets/dist' (not the default 'dist') because the
    // Wails embed package at condura-gui/frontend/assets/assets.go uses
    // //go:embed all:dist — Go embed resolves the pattern relative to the
    // package's own directory, so it expects dist/ INSIDE assets/. The
    // previous default ('dist' at the frontend root) silently broke
    // every test job and the GUI Build job with
    //   pattern all:dist: no matching files found
    // whenever CI ran from a fresh checkout (no pre-existing dist/).
    // 'npm run dev' is unaffected — Vite's dev server serves from memory
    // and ignores outDir.
    outDir: 'assets/dist',
    emptyOutDir: true,
  },
}))