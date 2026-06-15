// Initialize all runes-based stores in dependency order.
// Called from main.ts before mounting the root component.

import { daemon } from './daemon.svelte'
import { conversation } from './conversation.svelte'
import { settings } from './settings.svelte'
import { spend } from './spend.svelte'
import { audit } from './audit.svelte'
import { halt } from './halt.svelte'
import { apiKeys } from './apikeys.svelte'
import { updateStore } from './update.svelte'
import { overlay } from './overlay.svelte'
import { trust } from './trust.svelte'
import { wailsBindings, ipc } from '../ipc/client'
import { mergeDaemonCatalog } from '../i18n'

export async function initStores(): Promise<void> {
  // Step 1: ask the Wails-side App for the in-process daemon status.
  // Wails is the source of truth for the loopback URL + auth token
  // since the daemon is embedded in the same binary. (Falls back to
  // localhost:7666 when the Wails bindings aren't available — e.g.
  // during pure Vite dev or in a non-Wails browser preview.)
  let baseURL = ''
  let authToken = ''
  try {
    const status = await wailsBindings.DaemonStatus()
    if (status.ready && status.addr) {
      baseURL = `http://${status.addr}`
    }
  } catch {
    // Wails bindings not available in the browser; fall through.
  }

  // If Wails didn't give us a URL, fall back to localhost:7666
  // (the default the standalone daemon uses).
  if (!baseURL) {
    baseURL = 'http://127.0.0.1:7666'
  }

  // Auth token is read from the daemon's config on first request.
  // The config.get response includes the (possibly empty) token;
  // we use it for subsequent calls. (Sub-phase 2.6: a dedicated
  // auth.get method may replace this.)
  try {
    const cfg = await fetch(`${baseURL}/api`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ jsonrpc: '2.0', method: 'config.get', id: 1 })
    })
    if (cfg.ok) {
      const r = await cfg.json()
      if (r?.result?.api_server?.auth_token) {
        authToken = r.result.api_server.auth_token
      }
    }
  } catch {
    // ignore — config.get will fail if daemon isn't up yet
  }

  // Step 2: configure + start the IPC client.
  daemon.configure({ baseURL, authToken })
  daemon.start()

  // Step 3: kick off background stores.
  spend.startPolling()
  halt.startPolling()
  updateStore.startPolling()
  conversation.startListening()
  overlay.start()

  // Step 4: load initial state from the daemon. Tolerate failures
  // (the daemon may be mid-startup); stores will refresh when the
  // SSE connection comes up.
  try {
    await Promise.allSettled([
      settings.refresh(),
      conversation.refreshList(),
      apiKeys.refresh(),
      audit.refresh(),
      trust.refreshBackups(),
      trust.refreshPermissions(),
      ipc.i18nLocale('en').then((r) => mergeDaemonCatalog('en', r.translations))
    ])
  } catch {
    // ignore
  }
}
