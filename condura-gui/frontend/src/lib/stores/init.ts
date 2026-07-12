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
import { replay } from './replay.svelte'
import { consent } from './consent.svelte'
import {
  startListening as startPendingListening,
  startPolling as startPendingPolling,
  refreshPendingActions,
} from './pending.svelte'
import { wailsBindings, ipc } from '../ipc/client'
import { mergeDaemonCatalog } from '../i18n'

/**
 * After IPC/SSE comes back: resync kill-switch, consent, pending queue,
 * and the open Ask thread. Prevents "Ready" while daemon is halted, and
 * recovers assistant text finished offline. Safe on first connect too.
 */
export async function resyncAfterReconnect(): Promise<void> {
  await Promise.allSettled([
    halt.refresh(),
    consent.poll(),
    refreshPendingActions(),
    conversation.resyncFromDaemon(),
  ])
}

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

  // If Wails didn't give us a URL:
  //   - Vite / Cursor Simple Browser (DEV): use same-origin "" so
  //     /api + /events go through the Vite proxy → :7666. Chromium
  //     blocks cross-origin private-network fetches (:5173 → :7666).
  //   - Otherwise fall back to the standalone daemon address.
  if (!baseURL) {
    if (import.meta.env.DEV) {
      baseURL = ''
    } else {
      baseURL = 'http://127.0.0.1:7666'
    }
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
  // Ensure SSE open has begun (started=true) before store RPCs race.
  // daemon.start is sync until ipc.start's first await; give the microtask
  // a tick so openSse is in flight before refresh calls.
  await Promise.resolve()
  // Register after daemon.start so connected handlers run with connected=true.
  // First open + every SSE reconnect.
  ipc.on('connected', () => {
    void resyncAfterReconnect()
  })

  // Step 3: kick off background stores.
  spend.startPolling()
  halt.startPolling()
  updateStore.startPolling()
  // Ask / MeridianChat: SSE stream deltas + disconnect cleanup.
  // Legacy routes/Chat.svelte mounted this; Meridian never did — without
  // it isStreaming sticks and tokens never render.
  conversation.startListening()
  // Sub-agent queue — SSE live + poll fallback; dock badge without visiting Agents.
  startPendingListening()
  startPendingPolling(8000)
  overlay.start()

  // Step 4: load initial state from the daemon. Tolerate failures
  // (the daemon may be mid-startup); stores will refresh when the
  // SSE connection comes up. resyncAfterReconnect also runs on connected.
  try {
    await Promise.allSettled([
      settings.refresh(),
      conversation.refreshList(),
      apiKeys.refresh(),
      audit.refresh(),
      trust.refreshBackups(),
      trust.refreshPermissions(),
      replay.refresh(),
      halt.refresh(),
      ipc.i18nLocale('en').then((r) => mergeDaemonCatalog('en', r.translations))
    ])
  } catch {
    // ignore
  }
}
