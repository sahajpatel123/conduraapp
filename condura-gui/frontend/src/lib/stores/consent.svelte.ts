// Consent store. Live SSE for Gatekeeper tickets + light poll fallback.
// Daemon owns tickets, timeout, and audit; GUI only surfaces Allow/Deny.

import { ipc } from '../ipc/client'
import { notifications } from './notifications.svelte'
import type { ConsentTicket } from '../ipc/types'

/** Slow fallback when SSE drops; live path is safety.consent.request. */
const POLL_INTERVAL_MS = 4000
const CONSENT_TIMEOUT_MS = 300000 // 5 minutes, matches gatekeeper default.

class ConsentStore {
  ticket = $state<ConsentTicket | null>(null)
  error = $state<string>('')
  timer = $state<number>(CONSENT_TIMEOUT_MS)
  private intervalId: ReturnType<typeof setInterval> | null = null
  private countdownId: ReturnType<typeof setInterval> | null = null
  private unsubSSE: (() => void) | null = null

  start(): void {
    this.stop()
    this.startListening()
    this.intervalId = setInterval(() => {
      void this.poll()
    }, POLL_INTERVAL_MS)
    void this.poll()
  }

  /** Subscribe to live consent SSE. Idempotent. */
  startListening(): void {
    if (this.unsubSSE) return
    this.unsubSSE = ipc.on('consent', (t) => {
      if (!t?.nonce) {
        // Empty push → re-sync from daemon (ticket may have cleared).
        void this.poll()
        return
      }
      if (t.nonce !== this.ticket?.nonce) {
        this.ticket = {
          nonce: t.nonce,
          action_kind: t.action_kind ?? '',
          actor: t.actor ?? '',
          detail: t.detail ?? '',
          created_at: t.created_at ?? '',
          expires_at: t.expires_at ?? '',
          approved: !!t.approved,
        }
        this.error = ''
        this.resetCountdown()
      }
      // Always refresh full list so multi-ticket ordering stays correct.
      void this.poll()
    })
  }

  stop(): void {
    if (this.intervalId) {
      clearInterval(this.intervalId)
      this.intervalId = null
    }
    if (this.countdownId) {
      clearInterval(this.countdownId)
      this.countdownId = null
    }
    if (this.unsubSSE) {
      this.unsubSSE()
      this.unsubSSE = null
    }
  }

  async poll(): Promise<void> {
    if (!ipc.isConnected()) {
      return
    }
    try {
      const res = await ipc.gatekeeperPendingConsent()
      const tickets: ConsentTicket[] = res.tickets ?? []
      if (tickets.length === 0) {
        this.ticket = null
        return
      }
      // Prefer the first pending ticket. Reset countdown on nonce change.
      const next = tickets[0]
      if (next.nonce !== this.ticket?.nonce) {
        this.ticket = next
        this.error = ''
        this.resetCountdown()
      } else if (this.ticket) {
        // Refresh fields (actor/detail) without resetting timer.
        this.ticket = { ...this.ticket, ...next }
      }
    } catch (err) {
      // Don't surface every poll error as a toast; the daemon may
      // be temporarily unreachable. Keep the last ticket visible.
      this.error = String(err)
    }
  }

  resetCountdown(): void {
    if (this.countdownId) {
      clearInterval(this.countdownId)
    }
    this.timer = CONSENT_TIMEOUT_MS
    this.countdownId = setInterval(() => {
      this.timer -= 1000
      if (this.timer <= 0) {
        this.timer = 0
        // Timeout: the daemon queues the action automatically;
        // we just clear the local modal so the user isn't stuck.
        this.ticket = null
        clearInterval(this.countdownId ?? undefined)
        this.countdownId = null
      }
    }, 1000)
  }

  async approve(): Promise<void> {
    if (!this.ticket) return
    // Clear only on success. A failed RPC must leave the sheet up so the
    // user can retry — clearing here felt like Allow worked when it did not.
    const nonce = this.ticket.nonce
    try {
      await ipc.gatekeeperApprove(nonce)
      notifications.push({ kind: 'success', title: 'Action allowed', message: '' })
      if (this.ticket?.nonce === nonce) this.ticket = null
      this.error = ''
    } catch (err) {
      this.error = String(err)
      notifications.push({
        kind: 'error',
        title: 'Could not allow action',
        message: String(err),
        sticky: true,
      })
    }
  }

  async deny(): Promise<void> {
    if (!this.ticket) return
    const nonce = this.ticket.nonce
    try {
      await ipc.gatekeeperDeny(nonce)
      notifications.push({ kind: 'info', title: 'Action denied', message: '' })
      if (this.ticket?.nonce === nonce) this.ticket = null
      this.error = ''
    } catch (err) {
      this.error = String(err)
      notifications.push({
        kind: 'error',
        title: 'Could not deny action',
        message: String(err),
        sticky: true,
      })
    }
  }
}

export const consent = new ConsentStore()
