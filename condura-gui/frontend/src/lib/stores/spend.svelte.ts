// Spend monitoring store. Polls the daemon every 30s and listens for
// SSE spend_warning so Meridian Settings / toasts stay live under 24/7 use.

import { ipc } from '../ipc/client'
import type { SpendSummary } from '../ipc/types'
import { notifications } from './notifications.svelte'

class SpendStore {
  summary = $state<SpendSummary | null>(null)
  /** True while SSE spend_warning is subscribed. */
  live = $state(false)
  private interval: ReturnType<typeof setInterval> | null = null
  private sseOff: (() => void) | null = null
  private warned80 = false
  private warned100 = false

  get spent(): number {
    return this.summary?.spent ?? 0
  }

  get cap(): number {
    return this.summary?.cap ?? 0
  }

  get remaining(): number {
    return this.summary?.remaining ?? 0
  }

  /** 0–100 fill for gauges. */
  get pct(): number {
    if (!this.summary || this.summary.cap <= 0) return 0
    return Math.min(100, Math.round((this.summary.spent / this.summary.cap) * 100))
  }

  applySummary(s: SpendSummary | null | undefined): void {
    if (!s || typeof s.spent !== 'number') return
    this.summary = {
      spent: s.spent ?? 0,
      cap: s.cap ?? 0,
      remaining: s.remaining ?? Math.max(0, (s.cap ?? 0) - (s.spent ?? 0)),
    }
    this.checkWarnings()
  }

  async refresh(): Promise<void> {
    try {
      const s = await ipc.spendToday()
      this.applySummary(s)
    } catch {
      // Keep last summary; offline is soft for the gauge.
    }
  }

  startPolling(): void {
    if (this.interval) return
    void this.refresh()
    this.interval = setInterval(() => {
      void this.refresh()
    }, 30_000)
    this.startLive()
  }

  stopPolling(): void {
    if (this.interval) {
      clearInterval(this.interval)
      this.interval = null
    }
    this.stopLive()
  }

  /** Subscribe to daemon spend_warning (published at ≥80% of cap after stream spend). */
  startLive(): void {
    if (this.sseOff) return
    try {
      this.sseOff = ipc.on('spend_warning', (s) => {
        this.applySummary(s as SpendSummary)
      })
      this.live = true
    } catch {
      this.sseOff = null
      this.live = false
    }
  }

  stopLive(): void {
    if (this.sseOff) {
      try {
        this.sseOff()
      } catch {
        /* ignore */
      }
      this.sseOff = null
    }
    this.live = false
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
        message: `$${this.summary.spent.toFixed(2)} of $${this.summary.cap.toFixed(2)} used. Further cloud calls are blocked until tomorrow or you raise the cap.`,
        sticky: true,
      })
    } else if (pct >= 80 && !this.warned80) {
      this.warned80 = true
      notifications.push({
        kind: 'warn',
        title: 'Approaching daily spend cap',
        message: `$${this.summary.spent.toFixed(2)} of $${this.summary.cap.toFixed(2)} used (${pct.toFixed(0)}%).`,
      })
    }
    // Reset warning flags if the user raises the cap or a new day resets spend.
    if (pct < 80) {
      this.warned80 = false
    }
    if (pct < 100) {
      this.warned100 = false
    }
  }
}

export const spend = new SpendStore()
