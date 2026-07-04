// Pending sub-agent actions store.
//
// Phase 18 (v0.2.0): surfaces the persistent queue of
// ActionRequests emitted by spawned sub-agents. The user
// approves or denies each one via the GUI; approved rows
// flow into the executor (shell.exec or computeruse.*).
//
// v0.2.0 transports updates via polling. SSE wiring for the
// namespaced `pending_action.*` events is a Phase 18.1
// follow-on (the daemon already publishes them; the IPC
// client's typed event list is the blocker).
import { writable, derived, get } from 'svelte/store'
import { ipc } from '../ipc/client'

export type PendingStatus =
  | 'pending'
  | 'approved'
  | 'denied'
  | 'executed'
  | 'failed'
  | 'expired'
  | 'superseded'

export interface PendingAction {
  id: string
  spawn_id: string
  agent_name: string
  kind: string
  payload: {
    command?: string
    path?: string
    body?: string
    target?: string
    key?: string
  }
  gate_decision: string
  gate_reason: string
  status: PendingStatus
  created_at: string
  expires_at: string
  decided_at?: string
  decided_by?: string
  decision_note?: string
  executed_at?: string
  exit_code: number
  result: string
  execution_error?: string
  duration_ms: number
}

export const pendingActions = writable<PendingAction[]>([])

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

/**
 * Refresh the entire pending-action list from the daemon.
 * Called on mount, after every user action, and on every
 * poll tick.
 */
export async function refreshPendingActions(status?: PendingStatus): Promise<void> {
  try {
    const resp = await ipc.delegatePendingList(status)
    pendingActions.set(resp?.actions ?? [])
  } catch (e) {
    console.error('refresh pending actions failed', e)
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
    await refreshPendingActions()
    return updated ?? null
  } catch (e) {
    console.error('approve failed', e)
    return null
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
    await refreshPendingActions()
    return updated ?? null
  } catch (e) {
    console.error('deny failed', e)
    return null
  }
}

/**
 * Execute a previously-approved pending action. Used by the
 * GUI's "Run now" button on an already-approved row.
 */
export async function executePending(id: string): Promise<PendingAction | null> {
  try {
    const updated = await ipc.delegatePendingExecute({ id })
    await refreshPendingActions()
    return updated ?? null
  } catch (e) {
    console.error('execute failed', e)
    return null
  }
}

let pollTimer: ReturnType<typeof setInterval> | null = null

/**
 * Start polling the daemon's pending list every `intervalMs`.
 * The caller is responsible for calling stopPolling()
 * when the panel unmounts. Idempotent.
 */
export function startPolling(intervalMs = 5000): void {
  if (pollTimer != null) return
  // Fire one refresh immediately so the panel doesn't sit empty.
  void refreshPendingActions()
  pollTimer = setInterval(() => {
    void refreshPendingActions()
  }, intervalMs)
}

/** Stop polling. Safe to call when not polling. */
export function stopPolling(): void {
  if (pollTimer == null) return
  clearInterval(pollTimer)
  pollTimer = null
}

/** Convenience for tests: current snapshot. */
export function snapshot(): PendingAction[] {
  return get(pendingActions)
}
