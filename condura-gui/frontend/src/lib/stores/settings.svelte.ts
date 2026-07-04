// Settings store. Mirrors the daemon's AppConfig but the GUI keeps
// a working copy in memory; we push changes via config.update.

import { ipc } from '../ipc/client'
import type { AppConfig } from '../ipc/types'

class SettingsStore {
  config = $state<AppConfig | null>(null)
  loaded = $state<boolean>(false)
  saving = $state<boolean>(false)
  lastSaveError = $state<string>('')

  async refresh(): Promise<void> {
    this.config = await ipc.configGet()
    this.loaded = true
  }

  async save(patch: Partial<AppConfig>): Promise<void> {
    if (!this.config) {
      return
    }
    this.saving = true
    this.lastSaveError = ''
    try {
      await ipc.configUpdate(patch)
      this.config = { ...this.config, ...patch } as AppConfig
    } catch (err) {
      this.lastSaveError = String(err)
      throw err
    } finally {
      this.saving = false
    }
  }

  /**
   * Convenience setter for deeply-nested config keys.
   *   setIn('hotkey', 'overlay', 'Cmd+Shift+Space')
   */
  setIn<K1 extends keyof AppConfig, K2 extends keyof AppConfig[K1]>(
    k1: K1,
    k2: K2,
    value: AppConfig[K1][K2]
  ): void {
    if (!this.config) {
      return
    }
    const next: AppConfig = {
      ...this.config,
      [k1]: { ...(this.config[k1] as Record<string, unknown>), [k2]: value }
    } as AppConfig
    this.config = next
    void this.save({ [k1]: next[k1] } as Partial<AppConfig>)
  }
}

export const settings = new SettingsStore()
