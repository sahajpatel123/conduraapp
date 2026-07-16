import { beforeEach, describe, expect, it, vi } from 'vitest'

const { syncPairBeginTyped, syncPeers, syncListPairs, syncWith } = vi.hoisted(() => ({
  syncPairBeginTyped: vi.fn(),
  syncPeers: vi.fn(),
  syncListPairs: vi.fn(),
  syncWith: vi.fn(),
}))

vi.mock('../ipc/client', () => ({
  ipc: {
    syncPairBeginTyped,
    syncPeers,
    syncListPairs,
    syncWith,
    syncPairConfirmTyped: vi.fn(),
    syncRevokeTyped: vi.fn(),
  },
}))

import { SyncStore } from './sync.svelte'

describe('SyncStore.pairWith', () => {
  beforeEach(() => {
    syncPairBeginTyped.mockReset()
    syncPeers.mockReset()
    syncListPairs.mockReset()
  })

  it('uses expires_in seconds for pendingExpiresAt', async () => {
    syncPairBeginTyped.mockResolvedValue({
      ok: true,
      pin: '123456',
      peer: 'peer-1',
      expires_in: 120,
    })
    const store = new SyncStore()
    const before = Date.now()
    await store.pairWith('peer-1')
    expect(store.pendingPin).toBe('123456')
    expect(store.pendingPeerId).toBe('peer-1')
    const exp = new Date(store.pendingExpiresAt).getTime()
    expect(exp).toBeGreaterThanOrEqual(before + 119_000)
    expect(exp).toBeLessThanOrEqual(before + 121_000)
  })

  it('falls back to 300s when expires_in is missing', async () => {
    syncPairBeginTyped.mockResolvedValue({
      ok: true,
      pin: '654321',
      peer: 'peer-2',
      // expires_in omitted — old daemon
    })
    const store = new SyncStore()
    const before = Date.now()
    await store.pairWith('peer-2')
    const exp = new Date(store.pendingExpiresAt).getTime()
    expect(exp).toBeGreaterThanOrEqual(before + 299_000)
    expect(exp).toBeLessThanOrEqual(before + 301_000)
  })

  it('quiet refresh does not leave loading stuck true', async () => {
    syncPeers.mockResolvedValue({ peers: [] })
    syncListPairs.mockResolvedValue({ devices: [] })
    const store = new SyncStore()
    store.loading = false
    await store.refresh({ quiet: true })
    expect(store.loading).toBe(false)
    expect(syncPeers).toHaveBeenCalled()
  })

  it('syncWith calls daemon and quiet-refreshes', async () => {
    syncWith.mockResolvedValue({ ok: true, merged: 3 })
    syncPeers.mockResolvedValue({ peers: [{ device_id: 'd1', name: 'Peer' }] })
    syncListPairs.mockResolvedValue({ devices: [{ device_id: 'd1', device_name: 'Peer' }] })
    const store = new SyncStore()
    const r = await store.syncWith('d1')
    expect(syncWith).toHaveBeenCalledWith('d1')
    expect(r).toEqual({ ok: true, merged: 3 })
    expect(store.pairs).toHaveLength(1)
    expect(store.isDiscoverable('d1')).toBe(true)
    expect(store.loading).toBe(false)
  })

  it('syncWith surfaces errors without throwing', async () => {
    syncWith.mockRejectedValue(new Error('device not currently discoverable on LAN'))
    const store = new SyncStore()
    const r = await store.syncWith('missing')
    expect(r).toBeNull()
    expect(store.error).toMatch(/discoverable/i)
  })

  it('refresh populates peers and pairs from daemon', async () => {
    syncPeers.mockResolvedValue({
      peers: [
        { device_id: 'p1', name: 'Laptop' },
        { device_id: 'p2', name: 'Phone' },
      ],
    })
    syncListPairs.mockResolvedValue({
      devices: [{ device_id: 'p2', device_name: 'Phone' }],
    })
    const store = new SyncStore()
    await store.refresh()
    expect(store.peers).toHaveLength(2)
    expect(store.pairs).toHaveLength(1)
    expect(store.loading).toBe(false)
  })

  it('refresh records error on daemon failure', async () => {
    syncPeers.mockRejectedValue(new Error('IPC client not started'))
    syncListPairs.mockRejectedValue(new Error('IPC client not started'))
    const store = new SyncStore()
    await store.refresh()
    expect(store.error).toMatch(/IPC client not started/i)
    expect(store.peers).toEqual([])
    expect(store.pairs).toEqual([])
  })

  it('peerById returns matching peer or null', async () => {
    syncPeers.mockResolvedValue({
      peers: [
        { device_id: 'p1', name: 'Laptop' },
        { device_id: 'p2', name: 'Phone' },
      ],
    })
    syncListPairs.mockResolvedValue({ devices: [] })
    const store = new SyncStore()
    await store.refresh()
    expect(store.peerById('p1')?.name).toBe('Laptop')
    expect(store.peerById('p2')?.name).toBe('Phone')
    expect(store.peerById('missing')).toBeNull()
  })

  it('isDiscoverable matches peer ids present in peers list', async () => {
    syncPeers.mockResolvedValue({
      peers: [{ device_id: 'p1', name: 'Laptop' }],
    })
    syncListPairs.mockResolvedValue({ devices: [] })
    const store = new SyncStore()
    await store.refresh()
    expect(store.isDiscoverable('p1')).toBe(true)
    expect(store.isDiscoverable('offline-peer')).toBe(false)
  })

  it('clearPending resets pendingPin + pendingPeerId + pendingExpiresAt', async () => {
    syncPairBeginTyped.mockResolvedValue({
      ok: true,
      pin: '999999',
      peer: 'peer-x',
      expires_in: 60,
    })
    const store = new SyncStore()
    await store.pairWith('peer-x')
    expect(store.pendingPin).toBe('999999')
    store.clearPending()
    expect(store.pendingPin).toBe('')
    expect(store.pendingPeerId).toBe('')
    expect(store.pendingExpiresAt).toBe('')
  })

  it('pairWith surfaces error and clears pending state on failure', async () => {
    syncPairBeginTyped.mockRejectedValue(new Error('peer already paired'))
    const store = new SyncStore()
    const r = await store.pairWith('peer-x')
    expect(r).toBeNull()
    expect(store.pendingPin).toBe('')
    expect(store.error).toMatch(/already paired/i)
  })
})
