// Pending sub-agent actions store.
//
// Surfaces the persistent queue of ActionRequests emitted by spawned
// sub-agents. The user approves or denies each one via the GUI; approved
// rows flow into the executor (shell.exec or computeruse.*).
//
// Transport: SSE `pending_action.<status>` (live) + light polling fallback
// (reconnect gaps / missed events). Daemon publishes on insert/decide/execute.
import { writable, derived, get } from 'svelte/store'
import { ipc } from '../ipc/client'
import type { PendingActionRow, PendingActionStatus } from '../ipc/types'
import { humanizeIpcError } from '../ipc/errors'
import { notifications } from './notifications.svelte'

export type PendingStatus = PendingActionStatus

export type PendingAction = PendingActionRow

export const pendingActions = writable<PendingAction[]>([])

/** Last refresh failure — empty string when the last list pull succeeded. */
export const pendingRefreshError = writable<string>('')

/** Currently-decided-by identifier sent with every decide call. */
let currentActor = 'user:anonymous'

export function setPendingActor(actor: string) {
  currentActor = actor
}

/**
 * Pending count (status === 'pending') — the badge the
 * tray shows in the top bar.
 */
export const pendingCount = derived(pendingActions, ($rows) =>
  $rows.filter((r) => r.status === 'pending').length,
)

function normalizeRow(row: PendingActionRow): PendingAction {
  return {
    ...row,
    payload: row.payload ?? {},
    exit_code: row.exit_code ?? 0,
    result: row.result ?? '',
    duration_ms: row.duration_ms ?? 0,
    agent_name: row.agent_name ?? '',
    kind: row.kind ?? '',
    gate_decision: row.gate_decision ?? '',
    gate_reason: row.gate_reason ?? '',
    status: row.status ?? 'pending',
    created_at: row.created_at ?? '',
    expires_at: row.expires_at ?? '',
    spawn_id: row.spawn_id ?? '',
  }
}

/** Upsert one row from SSE or an RPC response. Newest-first for new ids. */
export function applyPendingRow(raw: PendingActionRow): void {
  if (!raw?.id) return
  const row = normalizeRow(raw)
  pendingActions.update((list) => {
    const i = list.findIndex((r) => r.id === row.id)
    if (i < 0) return [row, ...list]
    const next = list.slice()
    next[i] = { ...list[i], ...row }
    return next
  })
}

/**
 * Refresh the entire pending-action list from the daemon.
 * Called on mount, after every user action, and on every
 * poll tick.
 */
export async function refreshPendingActions(status?: PendingStatus): Promise<void> {
  try {
    const resp = await ipc.delegatePendingList(status)
    const rows = (resp?.actions ?? []).map(normalizeRow)
    pendingActions.set(rows)
    pendingRefreshError.set('')
  } catch (e) {
    console.error('refresh pending actions failed', e)
    pendingRefreshError.set(
      humanizeIpcError(e, 'Daemon offline — pending queue will refresh when Condura reconnects')
    )
  }
}

/**
 * Approve a pending action. If autoRun is true, the daemon
 * will also fire the executor right after approval.
 */
export async function approvePending(
  id: string,
  note = '',
  autoRun = true,
): Promise<PendingAction | null> {
  try {
    const updated = await ipc.delegatePendingDecide({
      id,
      decision: 'approve',
      decided_by: currentActor,
      note,
      auto_run: autoRun,
    })
    if (updated?.id) applyPendingRow(updated)
    else await refreshPendingActions()
    return updated ? normalizeRow(updated) : null
  } catch (e) {
    // Gatekeeper decisions must never look successful when the RPC failed.
    console.error('approve failed', e)
    notifications.push({
      kind: 'error',
      title: 'Could not approve action',
      message: humanizeIpcError(e),
      sticky: true,
    })
    throw e instanceof Error ? e : new Error(String(e))
  }
}

/** Deny a pending action. No executor side-effect. */
export async function denyPending(
  id: string,
  note = '',
): Promise<PendingAction | null> {
  try {
    const updated = await ipc.delegatePendingDecide({
      id,
      decision: 'deny',
      decided_by: currentActor,
      note,
      auto_run: false,
    })
    if (updated?.id) applyPendingRow(updated)
    else await refreshPendingActions()
    return updated ? normalizeRow(updated) : null
  } catch (e) {
    console.error('deny failed', e)
    notifications.push({
      kind: 'error',
      title: 'Could not deny action',
      message: humanizeIpcError(e),
      sticky: true,
    })
    throw e instanceof Error ? e : new Error(String(e))
  }
}

/**
 * Execute a previously-approved pending action. Used by the
 * GUI's "Run now" button on an already-approved row.
 */
export async function executePending(id: string): Promise<PendingAction | null> {
  try {
    const updated = await ipc.delegatePendingExecute({ id })
    if (updated?.id) applyPendingRow(updated)
    else await refreshPendingActions()
    return updated ? normalizeRow(updated) : null
  } catch (e) {
    console.error('execute failed', e)
    notifications.push({
      kind: 'error',
      title: 'Could not run action',
      message: humanizeIpcError(e),
      sticky: true,
    })
    throw e instanceof Error ? e : new Error(String(e))
  }
}

let pollTimer: ReturnType<typeof setInterval> | null = null
let pollRefCount = 0
let unsubSSE: (() => void) | null = null

/**
 * Subscribe to live pending_action SSE. Idempotent. Safe to leave
 * running for the app lifetime (dock badge).
 */
export function startListening(): void {
  if (unsubSSE) return
  unsubSSE = ipc.on('pending_action', (row) => {
    applyPendingRow(row)
  })
}

/** Drop the SSE subscription. Prefer leaving it on for the dock badge. */
export function stopListening(): void {
  if (!unsubSSE) return
  unsubSSE()
  unsubSSE = null
}

/**
 * Start polling the daemon's pending list every `intervalMs`.
 * Ref-counted so Agents page unmount does not kill the global dock poll.
 * Also attaches SSE for live updates.
 */
export function startPolling(intervalMs = 5000): void {
  pollRefCount++
  startListening()
  if (pollTimer != null) return
  void refreshPendingActions()
  pollTimer = setInterval(() => {
    void refreshPendingActions()
  }, intervalMs)
}

/**
 * Release one polling interest. Timer + SSE stay up while any
 * interest remains (initStores + Agents route).
 */
export function stopPolling(): void {
  pollRefCount = Math.max(0, pollRefCount - 1)
  if (pollRefCount > 0) return
  if (pollTimer != null) {
    clearInterval(pollTimer)
    pollTimer = null
  }
  // Keep SSE so a late-arriving sub-agent still lights the dock.
}

/** Convenience for tests: current snapshot. */
export function snapshot(): PendingAction[] {
  return get(pendingActions)
}
