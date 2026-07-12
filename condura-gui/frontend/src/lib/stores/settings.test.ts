import { describe, it, expect, vi, beforeEach } from 'vitest'

const ipcMock = {
  configGet: vi.fn(),
  configUpdate: vi.fn().mockResolvedValue(undefined),
}

vi.mock('../ipc/client', () => ({
  ipc: {
    configGet: (...args: unknown[]) => ipcMock.configGet(...args),
    configUpdate: (...args: unknown[]) => ipcMock.configUpdate(...args),
  },
}))

const { settings, deepMergeConfig } = await import('./settings.svelte')

describe('deepMergeConfig', () => {
  it('merges nested objects without wiping siblings', () => {
    const base = {
      voice: { enabled: true, binary_path: '/bin/w', wake: { enabled: false, hotword: 'hey' } },
      telemetry: { enabled: false, endpoint: 'x' },
    }
    const next = deepMergeConfig(base, {
      voice: { wake: { enabled: true } },
    })
    expect(next.voice.enabled).toBe(true)
    expect(next.voice.binary_path).toBe('/bin/w')
    expect(next.voice.wake.enabled).toBe(true)
    expect(next.voice.wake.hotword).toBe('hey')
    expect(next.telemetry.endpoint).toBe('x')
  })

  it('replaces arrays rather than merging by index', () => {
    const base = { tags: ['a', 'b'] }
    const next = deepMergeConfig(base, { tags: ['c'] })
    expect(next.tags).toEqual(['c'])
  })
})

describe('settings.save', () => {
  beforeEach(() => {
    vi.clearAllMocks()
    ipcMock.configUpdate.mockResolvedValue(undefined)
    settings.config = {
      version: 1,
      general: { data_dir: '', language: 'en' },
      daemon: { idle_timeout_minutes: 0, default_autonomy: 'warn' },
      logging: { level: 'info', format: 'text', file: '', add_source: false },
      storage: { path: '', backup: { dir: '', retention_days: 7 } },
      security: { audit_log_path: '', audit_retention_days: 90, spend_limit_usd_per_day: 5 },
      api_server: {
        host: '127.0.0.1',
        port: 7666,
        auth_token: '',
        tls_enabled: false,
        allowed_origins: [],
      },
      llm: { providers: {}, oauth_providers: {} },
      autonomy: { default_level: 'warn', per_app: {}, per_task: {} },
      telemetry: { enabled: false, endpoint: '' },
      hotkey: { overlay: '' },
      window: { width: 0, height: 0, x: 0, y: 0, last_conversation_id: 0 },
      voice: { enabled: false, wake: { enabled: false } },
    } as never
  })

  it('deep-merges then refreshes from config.get', async () => {
    ipcMock.configGet.mockResolvedValueOnce({
      ...settings.config,
      voice: { enabled: false, wake: { enabled: true } },
      autonomy: { default_level: 'supervised', per_app: {}, per_task: {} },
    })
    await settings.save({
      autonomy: { default_level: 'supervised' },
      voice: { wake: { enabled: true } },
    })
    expect(ipcMock.configUpdate).toHaveBeenCalled()
    expect(ipcMock.configGet).toHaveBeenCalled()
    expect(settings.config?.voice?.wake?.enabled).toBe(true)
    expect(settings.config?.autonomy?.default_level).toBe('supervised')
  })

  it('throws when config is not loaded instead of silent no-op', async () => {
    settings.config = null
    await expect(settings.save({ telemetry: { enabled: true } })).rejects.toThrow(/not loaded/i)
    expect(ipcMock.configUpdate).not.toHaveBeenCalled()
    expect(settings.lastSaveError).toMatch(/not loaded/i)
  })
})
