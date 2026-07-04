// Halt / kill-switch store. Polls the daemon's halt state and exposes
// a one-click "stop everything" method.

import { ipc } from '../ipc/client'
import type { DaemonResumeRequestResult, HaltState } from '../ipc/types'
import { notifications } from './notifications.svelte'

class HaltStore {
  state = $state<HaltState>({ halted: false })
  private interval: ReturnType<typeof setInterval> | null = null

  async refresh(): Promise<void> {
    this.state = await ipc.haltState()
  }

  startPolling(): void {
    if (this.interval) {
      return
    }
    void this.refresh()
    this.interval = setInterval(() => {
      void this.refresh()
    }, 5000)
  }

  stopPolling(): void {
    if (this.interval) {
      clearInterval(this.interval)
      this.interval = null
    }
  }

  async halt(reason: string = 'user requested'): Promise<void> {
    const res = await ipc.daemonHalt(reason)
    notifications.push({
      kind: 'warn',
      title: 'All activity halted',
      message: `Halted at ${res.timestamp}. ${res.active_streams_canceled} stream(s) canceled.`,
      sticky: true
    })
    await this.refresh()
  }

  /**
   * resume mints a sticky-resume ticket (T3b P0-1 core). The actual
   * un-halt requires a human-confirmed action via the CLI (`condura
   * resume confirm --ticket T`) — out of the in-process trust
   * boundary. We surface the ticket + the CLI hint to the user and
   * refresh; the polling tick will reflect the un-halt once the human
   * confirms.
   */
  async resume(): Promise<void> {
    const res: DaemonResumeRequestResult = await ipc.daemonResumeRequest()
    await this.refresh()
    if (!res.halted || !res.ticket) {
      notifications.push({
        kind: 'info',
        title: 'Already running',
        message: 'The daemon is not halted; nothing to resume.'
      })
      return
    }
    notifications.push({
      kind: 'info',
      title: 'Resume ticket minted',
      message:
        `Ticket: ${res.ticket.slice(0, 8)}…\n` +
        `Confirm in a terminal with: ${res.confirm_via ?? 'condura resume confirm --ticket ' + res.ticket}`,
      sticky: true
    })
  }
}

export const halt = new HaltStore()
