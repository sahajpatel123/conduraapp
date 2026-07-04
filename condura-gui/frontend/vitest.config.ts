import { defineConfig } from 'vitest/config';
import { svelte } from '@sveltejs/vite-plugin-svelte';

// Vitest config for the Condura Svelte frontend.
//
// Why this config exists (audit SB-09): the project declared vitest
// in package.json but never created a config file, so `pnpm test`
// found zero tests and exited 0 — a false positive. This config
// adds jsdom for DOM testing, the Svelte plugin so .svelte files
// compile, and isolates the test glob to .test.ts files.
//
// Svelte 5 ships separate browser + server builds. To make vitest
// resolve the browser build (so `mount(...)` works), we set
// `resolve.conditions = ['browser']`. Without this, Svelte 5
// loads `svelte/src/index-server.js` and `mount()` throws
// "lifecycle_function_unavailable".
export default defineConfig({
  plugins: [svelte({ hot: false })],
  resolve: {
    conditions: ['browser'],
  },
  test: {
    environment: 'jsdom',
    globals: false,
    include: ['src/**/*.test.ts'],
    setupFiles: ['./vitest.setup.ts'],
  },
});