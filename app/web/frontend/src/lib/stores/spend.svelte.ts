// Spend monitoring store. Polls the daemon every 30s; pushes toasts
// at 80% / 100% of the daily cap.

import { ipc } from '../ipc/client'
import type { SpendSummary } from '../ipc/types'
import { notifications } from './notifications.svelte'

class SpendStore {
  summary = $state<SpendSummary | null>(null)
  private interval: ReturnType<typeof setInterval> | null = null
  private warned80 = false
  private warned100 = false

  async refresh(): Promise<void> {
    this.summary = await ipc.spendToday()
    this.checkWarnings()
  }

  startPolling(): void {
    if (this.interval) {
      return
    }
    void this.refresh()
    this.interval = setInterval(() => {
      void this.refresh()
    }, 30_000)
  }

  stopPolling(): void {
    if (this.interval) {
      clearInterval(this.interval)
      this.interval = null
    }
  }

  private checkWarnings(): void {
    if (!this.summary || this.summary.cap <= 0) {
      return
    }
    const pct = (this.summary.spent / this.summary.cap) * 100
    if (pct >= 100 && !this.warned100) {
      this.warned100 = true
      notifications.push({
        kind: 'error',
        title: 'Daily spend cap reached',
        message: `$${this.summary.spent.toFixed(2)} of $${this.summary.cap.toFixed(2)} used. Further calls will be blocked.`
      })
    } else if (pct >= 80 && !this.warned80) {
      this.warned80 = true
      notifications.push({
        kind: 'warn',
        title: 'Approaching daily spend cap',
        message: `$${this.summary.spent.toFixed(2)} of $${this.summary.cap.toFixed(2)} used (${pct.toFixed(0)}%).`
      })
    }
    // Reset warning flags if the user raises the cap.
    if (pct < 80) {
      this.warned80 = false
    }
    if (pct < 100) {
      this.warned100 = false
    }
  }
}

export const spend = new SpendStore()
