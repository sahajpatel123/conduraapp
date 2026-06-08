// Halt / kill-switch store. Polls the daemon's halt state and exposes
// a one-click "stop everything" method.

import { ipc } from '../ipc/client'
import type { HaltState } from '../ipc/types'
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

  async resume(): Promise<void> {
    await ipc.daemonResume()
    await this.refresh()
    notifications.push({
      kind: 'info',
      title: 'Activity resumed',
      message: 'The daemon is accepting requests again.'
    })
  }
}

export const halt = new HaltStore()
