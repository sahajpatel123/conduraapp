// Halt / kill-switch store. Live SSE + light poll fallback.
// Resume is sticky (T3b): GUI mints a ticket; human confirms via CLI.

import { ipc } from '../ipc/client'
import type { DaemonResumeRequestResult, HaltState } from '../ipc/types'
import { humanizeIpcError } from '../ipc/errors'
import { notifications } from './notifications.svelte'

class HaltStore {
  state = $state<HaltState>({ halted: false })
  /** One-time resume ticket from daemon.resume_request (empty when none). */
  ticket = $state<string>('')
  /** CLI hint for human confirm. */
  confirmVia = $state<string>('')
  ticketBusy = $state(false)
  copyNote = $state('')

  private interval: ReturnType<typeof setInterval> | null = null
  private unsubSSE: (() => void) | null = null

  async refresh(): Promise<void> {
    try {
      this.state = await ipc.haltState()
      if (!this.state.halted) {
        this.clearTicket()
      }
    } catch {
      // Daemon unreachable — leave last known state for the overlay.
    }
  }

  /** Subscribe to live "halt" SSE (RPC + watchdog paths publish it). */
  startListening(): void {
    if (this.unsubSSE) return
    this.unsubSSE = ipc.on('halt', (s: HaltState) => {
      this.state = {
        halted: !!s?.halted,
        since: s?.since,
        reason: s?.reason,
      }
      if (!this.state.halted) {
        this.clearTicket()
      }
    })
  }

  startPolling(): void {
    this.startListening()
    if (this.interval) {
      return
    }
    void this.refresh()
    // Fallback when SSE drops; live path is SSE.
    this.interval = setInterval(() => {
      void this.refresh()
    }, 8000)
  }

  stopPolling(): void {
    if (this.interval) {
      clearInterval(this.interval)
      this.interval = null
    }
  }

  async halt(reason: string = 'user requested'): Promise<void> {
    // Never claim halted until the daemon confirms — false quiet is worse
    // than a sticky error toast on the kill-switch.
    try {
      const res = await ipc.daemonHalt(reason)
      this.state = {
        halted: true,
        reason,
        since: res.timestamp,
      }
      notifications.push({
        kind: 'warn',
        title: 'All activity halted',
        message: `Halted at ${res.timestamp}. ${res.active_streams_canceled} stream(s) canceled.`,
        sticky: true,
      })
      await this.refresh()
    } catch (err) {
      notifications.push({
        kind: 'error',
        title: 'Halt failed',
        message: humanizeIpcError(
          err,
          'Daemon offline — Condura may still be running. Start the daemon, then try Halt again.'
        ),
        sticky: true,
      })
      throw err instanceof Error ? err : new Error(String(err))
    }
  }

  /**
   * resume mints a sticky-resume ticket (T3b). Un-halt requires CLI
   * `condura resume confirm --ticket <ticket>` with the human secret.
   * Surfaces the full ticket on MeridianHalt (not only a toast).
   */
  async resume(): Promise<void> {
    this.ticketBusy = true
    this.copyNote = ''
    try {
      const res: DaemonResumeRequestResult = await ipc.daemonResumeRequest()
      await this.refresh()
      if (!res.halted || !res.ticket) {
        this.clearTicket()
        notifications.push({
          kind: 'info',
          title: 'Already running',
          message: 'The daemon is not halted; nothing to resume.',
        })
        return
      }
      this.ticket = res.ticket
      // Prefer an interpolated CLI hint; never show a literal <ticket> placeholder
      // or multi-clause paste bait ("OR halt.confirm_resume…").
      // Must match the real CLI: `condura resume confirm --ticket …`
      const via = (res.confirm_via ?? '').replace(/<ticket>/g, res.ticket).trim()
      const cliOnly = via.split(/\s+OR\s+/i)[0]?.trim() ?? ''
      this.confirmVia =
        cliOnly && !cliOnly.includes('<ticket>')
          ? cliOnly
          : `condura resume confirm --ticket ${res.ticket}`
      notifications.push({
        kind: 'info',
        title: 'Resume ticket ready',
        message: 'Confirm in a terminal (out of process). Ticket shown on this sheet.',
        sticky: true,
      })
    } catch (err) {
      this.copyNote = humanizeIpcError(err, 'Daemon offline — cannot mint a resume ticket')
      notifications.push({
        kind: 'error',
        title: 'Could not mint resume ticket',
        message: this.copyNote,
        sticky: true,
      })
    } finally {
      this.ticketBusy = false
    }
  }

  async copyTicket(): Promise<void> {
    if (!this.ticket) return
    const cmd =
      this.confirmVia.includes(this.ticket)
        ? this.confirmVia
        : `condura resume confirm --ticket ${this.ticket}`
    try {
      await navigator.clipboard.writeText(cmd)
      this.copyNote = 'Copied command'
      setTimeout(() => {
        if (this.copyNote === 'Copied command') this.copyNote = ''
      }, 2000)
    } catch {
      this.copyNote = 'Copy failed — select the command manually'
    }
  }

  private clearTicket(): void {
    this.ticket = ''
    this.confirmVia = ''
    this.copyNote = ''
  }
}

export const halt = new HaltStore()
