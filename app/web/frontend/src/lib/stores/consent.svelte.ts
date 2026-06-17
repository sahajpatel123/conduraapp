// Consent store. Polls the daemon for pending Gatekeeper consent
// tickets and surfaces them as a reactive modal state.
//
// The store intentionally keeps minimal local state: the daemon owns
// the tickets, timeout handling, and audit trail. The GUI only
// decides whether a modal is visible and which action string to show.

import { ipc } from '../ipc/client'
import { notifications } from './notifications.svelte'
import type { ConsentTicket } from '../ipc/types'

const POLL_INTERVAL_MS = 1200
const CONSENT_TIMEOUT_MS = 300000 // 5 minutes, matches gatekeeper default.

class ConsentStore {
  ticket = $state<ConsentTicket | null>(null)
  error = $state<string>('')
  timer = $state<number>(CONSENT_TIMEOUT_MS)
  private intervalId: ReturnType<typeof setInterval> | null = null
  private countdownId: ReturnType<typeof setInterval> | null = null

  start(): void {
    this.stop()
    this.intervalId = setInterval(() => {
      void this.poll()
    }, POLL_INTERVAL_MS)
    void this.poll()
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
      // Show the first pending ticket. If it differs from the one
      // already shown, reset the countdown.
      const next = tickets[0]
      if (next.nonce !== this.ticket?.nonce) {
        this.ticket = next
        this.error = ''
        this.resetCountdown()
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
    try {
      await ipc.gatekeeperApprove(this.ticket.nonce)
      notifications.push({ kind: 'success', title: 'Action allowed', message: '' })
    } catch (err) {
      notifications.push({ kind: 'error', title: 'Could not allow action', message: String(err) })
    } finally {
      this.ticket = null
    }
  }

  async deny(): Promise<void> {
    if (!this.ticket) return
    try {
      await ipc.gatekeeperDeny(this.ticket.nonce)
      notifications.push({ kind: 'info', title: 'Action denied', message: '' })
    } catch (err) {
      notifications.push({ kind: 'error', title: 'Could not deny action', message: String(err) })
    } finally {
      this.ticket = null
    }
  }
}

export const consent = new ConsentStore()
