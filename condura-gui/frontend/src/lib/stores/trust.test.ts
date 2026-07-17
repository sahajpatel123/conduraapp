import { beforeEach, describe, expect, it, vi } from 'vitest'

const { backupList, backupCreate, permissionsStatus } = vi.hoisted(() => ({
  backupList: vi.fn(),
  backupCreate: vi.fn(),
  permissionsStatus: vi.fn(),
}))

vi.mock('../ipc/client', () => ({
  ipc: {
    backupList,
    backupCreate,
    permissionsStatus,
  },
}))

import { TrustStore } from './trust.svelte'

describe('TrustStore', () => {
  beforeEach(() => {
    backupList.mockReset()
    backupCreate.mockReset()
    permissionsStatus.mockReset()
  })

  it('starts with empty backups and permissions', () => {
    const store = new TrustStore()
    expect(store.backups).toEqual([])
    expect(store.permissions).toEqual([])
    expect(store.loadingBackups).toBe(false)
    expect(store.loadingPermissions).toBe(false)
  })

  it('refreshBackups populates backups from daemon', async () => {
    backupList.mockResolvedValue([
      { name: '2026-07-01', path: '/backups/2026-07-01.tar', size: 1024 },
      { name: '2026-07-15', path: '/backups/2026-07-15.tar', size: 2048 },
    ])
    const store = new TrustStore()
    await store.refreshBackups()
    expect(store.backups).toHaveLength(2)
    expect(store.backups[0]?.name).toBe('2026-07-01')
    expect(store.loadingBackups).toBe(false)
  })

  it('refreshBackups records error and clears backups on daemon failure', async () => {
    backupList.mockRejectedValue(new Error('IPC client not started'))
    const store = new TrustStore()
    await store.refreshBackups()
    expect(store.lastError).toMatch(/IPC client not started/i)
    expect(store.backups).toEqual([])
  })

  it('createBackup calls daemon and returns the path, then refreshes', async () => {
    backupCreate.mockResolvedValue({ path: '/backups/2026-07-17.tar' })
    backupList.mockResolvedValue([])
    const store = new TrustStore()
    const path = await store.createBackup()
    expect(path).toBe('/backups/2026-07-17.tar')
    expect(backupList).toHaveBeenCalled()
    expect(store.backups).toEqual([])
  })

  it('refreshPermissions populates from daemon', async () => {
    permissionsStatus.mockResolvedValue([
      { kind: 'accessibility', status: 'granted' },
      { kind: 'screen_recording', status: 'denied' },
    ])
    const store = new TrustStore()
    await store.refreshPermissions()
    expect(store.permissions).toHaveLength(2)
    expect(store.permissions[0]?.status).toBe('granted')
  })

  it('refreshPermissions quiet option does not flash loading state', async () => {
    permissionsStatus.mockResolvedValue([])
    const store = new TrustStore()
    store.loadingPermissions = false
    await store.refreshPermissions({ quiet: true })
    expect(store.loadingPermissions).toBe(false)
  })

  it('refreshPermissions records error and clears on daemon failure', async () => {
    permissionsStatus.mockRejectedValue(new Error('IPC client not started'))
    const store = new TrustStore()
    await store.refreshPermissions()
    expect(store.lastError).toMatch(/IPC client not started/i)
    expect(store.permissions).toEqual([])
  })
})