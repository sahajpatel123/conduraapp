// Settings store. Mirrors the daemon's AppConfig but the GUI keeps
// a working copy in memory; we push changes via config.update.

import { ipc } from '../ipc/client'
import type { AppConfig } from '../ipc/types'

/**
 * Deep-merge plain objects for local config mirrors.
 * Arrays and non-objects replace (no array element merge).
 * Prevents shallow save from wiping nested keys (e.g. voice.wake
 * patch dropping voice.binary_path in the local mirror).
 */
export function deepMergeConfig<T>(base: T, patch: unknown): T {
  if (patch === null || patch === undefined) return base
  if (typeof patch !== 'object' || Array.isArray(patch)) {
    return patch as T
  }
  if (typeof base !== 'object' || base === null || Array.isArray(base)) {
    return { ...(patch as object) } as T
  }
  const out: Record<string, unknown> = {
    ...(base as Record<string, unknown>),
  }
  for (const [k, v] of Object.entries(patch as Record<string, unknown>)) {
    if (v === undefined) continue
    const prev = out[k]
    if (
      v !== null &&
      typeof v === 'object' &&
      !Array.isArray(v) &&
      prev !== null &&
      typeof prev === 'object' &&
      !Array.isArray(prev)
    ) {
      out[k] = deepMergeConfig(prev, v)
    } else {
      out[k] = v
    }
  }
  return out as T
}

class SettingsStore {
  config = $state<AppConfig | null>(null)
  loaded = $state<boolean>(false)
  saving = $state<boolean>(false)
  lastSaveError = $state<string>('')

  async refresh(): Promise<void> {
    this.config = await ipc.configGet()
    this.loaded = true
  }

  async save(patch: Partial<AppConfig> | Record<string, unknown>): Promise<void> {
    if (!this.config) {
      // Never look like a successful save when config was never loaded.
      this.lastSaveError = 'Settings not loaded yet — wait for the daemon, then try again.'
      throw new Error(this.lastSaveError)
    }
    this.saving = true
    this.lastSaveError = ''
    try {
      await ipc.configUpdate(patch as Partial<AppConfig>)
      // Optimistic deep merge so nested patches don't clobber siblings.
      this.config = deepMergeConfig(this.config, patch)
      // Authoritative re-read (snake_case publicConfigView) after daemon apply.
      try {
        this.config = await ipc.configGet()
      } catch {
        // Keep optimistic merge if refresh fails mid-reconnect.
      }
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
      [k1]: { ...(this.config[k1] as Record<string, unknown>), [k2]: value },
    } as AppConfig
    this.config = next
    void this.save({ [k1]: next[k1] } as Partial<AppConfig>)
  }
}

export const settings = new SettingsStore()
