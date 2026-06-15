// Action Replay store — scrubbable 24h timeline with screenshots.

import { ipc } from '../ipc/client'
import type { ReplayFrame, ReplayIntegrityReport } from '../ipc/types'

class ReplayStore {
  frames = $state<ReplayFrame[]>([])
  selectedIndex = $state(0)
  loading = $state(false)
  exporting = $state(false)
  integrity = $state<ReplayIntegrityReport | null>(null)
  lastError = $state('')

  get selected(): ReplayFrame | null {
    return this.frames[this.selectedIndex] ?? null
  }

  async refresh(): Promise<void> {
    this.loading = true
    this.lastError = ''
    try {
      this.frames = await ipc.replayTimeline()
      if (this.selectedIndex >= this.frames.length) {
        this.selectedIndex = Math.max(0, this.frames.length - 1)
      }
    } catch (err) {
      this.lastError = String(err)
      this.frames = []
    } finally {
      this.loading = false
    }
  }

  async verifyIntegrity(): Promise<void> {
    try {
      this.integrity = await ipc.replayVerifyIntegrity()
    } catch (err) {
      this.lastError = String(err)
    }
  }

  async exportMP4(): Promise<string> {
    this.exporting = true
    this.lastError = ''
    try {
      const r = await ipc.replayExport()
      return r.path
    } catch (err) {
      this.lastError = String(err)
      throw err
    } finally {
      this.exporting = false
    }
  }

  selectIndex(i: number): void {
    if (i >= 0 && i < this.frames.length) {
      this.selectedIndex = i
    }
  }
}

export const replay = new ReplayStore()
