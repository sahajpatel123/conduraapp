// Update store. Background auto-update check (per locked-in
// decision: force auto-update by default, user can disable in
// settings). When an update is found, applies it silently.

import { ipc } from '../ipc/client'
import type { DaemonUpdateResult } from '../ipc/types'
import { notifications } from './notifications.svelte'

class UpdateStore {
  enabled = $state<boolean>(true)
  current = $state<string>('')
  latest = $state<string>('')
  downloading = $state<boolean>(false)
  lastCheck = $state<number>(0)
  private interval: ReturnType<typeof setInterval> | null = null

  setEnabled(v: boolean): void {
    this.enabled = v
  }

  async checkNow(): Promise<DaemonUpdateResult> {
    const r = await ipc.updateCheck()
    this.current = r.current_version
    this.latest = r.latest_version || r.current_version
    this.lastCheck = Date.now()
    return r
  }

  /**
   * Check for updates and apply if forced. Called periodically by
   * the GUI. The daemon handles the actual download + install;
   * the GUI just surfaces the result.
   */
  async autoUpdateCycle(): Promise<void> {
    if (!this.enabled) {
      return
    }
    try {
      const r = await this.checkNow()
      if (r.update_available) {
        this.downloading = true
        const applied = await ipc.updateApply()
        if (applied.update_available) {
          notifications.push({
            kind: 'info',
            title: 'Updated',
            message: `Synaptic ${applied.latest_version} is being installed. Restart when convenient.`,
            sticky: true
          })
        }
      }
    } catch {
      // ignore — update check is best-effort
    } finally {
      this.downloading = false
    }
  }

  startPolling(): void {
    if (this.interval) {
      return
    }
    void this.autoUpdateCycle()
    this.interval = setInterval(() => {
      void this.autoUpdateCycle()
    }, 6 * 60 * 60 * 1000) // every 6h
  }

  stopPolling(): void {
    if (this.interval) {
      clearInterval(this.interval)
      this.interval = null
    }
  }
}

export const updateStore = new UpdateStore()
