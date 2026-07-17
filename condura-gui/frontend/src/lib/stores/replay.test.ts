import { beforeEach, describe, expect, it, vi } from 'vitest'

const { replayTimeline, replayVerifyIntegrity, replayExport } = vi.hoisted(() => ({
  replayTimeline: vi.fn(),
  replayVerifyIntegrity: vi.fn(),
  replayExport: vi.fn(),
}))

vi.mock('../ipc/client', () => ({
  ipc: {
    replayTimeline,
    replayVerifyIntegrity,
    replayExport,
  },
}))

import { ReplayStore } from './replay.svelte'

describe('ReplayStore', () => {
  beforeEach(() => {
    replayTimeline.mockReset()
    replayVerifyIntegrity.mockReset()
    replayExport.mockReset()
  })

  it('starts with empty frames and selectedIndex 0', () => {
    const store = new ReplayStore()
    expect(store.frames).toEqual([])
    expect(store.selectedIndex).toBe(0)
    expect(store.selected).toBeNull()
    expect(store.loading).toBe(false)
    expect(store.exporting).toBe(false)
  })

  it('refresh populates frames from daemon', async () => {
    replayTimeline.mockResolvedValue([
      { id: 'f1', timestamp: '2026-07-17T10:00:00Z', action: 'click' },
      { id: 'f2', timestamp: '2026-07-17T10:01:00Z', action: 'type' },
    ])
    const store = new ReplayStore()
    await store.refresh()
    expect(store.frames).toHaveLength(2)
    expect(store.loading).toBe(false)
    expect(store.lastError).toBe('')
  })

  it('refresh records error and clears frames on daemon failure', async () => {
    replayTimeline.mockRejectedValue(new Error('IPC client not started'))
    const store = new ReplayStore()
    await store.refresh()
    expect(store.lastError).toMatch(/IPC client not started/i)
    expect(store.frames).toEqual([])
  })

  it('clamps selectedIndex when frames shrink after refresh', async () => {
    replayTimeline.mockResolvedValueOnce([
      { id: 'a', timestamp: 't', action: 'click' },
      { id: 'b', timestamp: 't', action: 'click' },
      { id: 'c', timestamp: 't', action: 'click' },
    ])
    const store = new ReplayStore()
    await store.refresh()
    store.selectIndex(2)
    expect(store.selectedIndex).toBe(2)
    // Second refresh returns fewer frames — selectedIndex must clamp.
    replayTimeline.mockResolvedValueOnce([
      { id: 'x', timestamp: 't', action: 'click' },
    ])
    await store.refresh()
    expect(store.selectedIndex).toBe(0)
  })

  it('selected getter returns the frame at selectedIndex', async () => {
    replayTimeline.mockResolvedValue([
      { id: 'a', timestamp: 't', action: 'click' },
      { id: 'b', timestamp: 't', action: 'type' },
    ])
    const store = new ReplayStore()
    await store.refresh()
    store.selectIndex(1)
    expect(store.selected?.id).toBe('b')
  })

  it('selectIndex ignores out-of-range indices', async () => {
    replayTimeline.mockResolvedValue([{ id: 'a', timestamp: 't', action: 'click' }])
    const store = new ReplayStore()
    await store.refresh()
    store.selectIndex(0)
    store.selectIndex(99) // beyond length
    store.selectIndex(-1) // negative
    expect(store.selectedIndex).toBe(0)
  })

  it('verifyIntegrity populates the integrity report', async () => {
    replayVerifyIntegrity.mockResolvedValue({
      ok: true,
      chain_length: 42,
      verified_at: '2026-07-17T11:00:00Z',
    })
    const store = new ReplayStore()
    await store.verifyIntegrity()
    expect(store.integrity?.chain_length).toBe(42)
    expect(store.lastError).toBe('')
  })

  it('verifyIntegrity records error on failure', async () => {
    replayVerifyIntegrity.mockRejectedValue(new Error('hash mismatch on frame 12'))
    const store = new ReplayStore()
    await store.verifyIntegrity()
    expect(store.lastError).toMatch(/hash mismatch/i)
    expect(store.integrity).toBeNull()
  })

  it('exportMP4 returns the path and clears exporting on success', async () => {
    replayExport.mockResolvedValue({ path: '/tmp/replay.mp4' })
    const store = new ReplayStore()
    const p = await store.exportMP4()
    expect(p).toBe('/tmp/replay.mp4')
    expect(store.exporting).toBe(false)
  })

  it('exportMP4 re-throws on failure but clears exporting', async () => {
    replayExport.mockRejectedValue(new Error('ffmpeg exited 1'))
    const store = new ReplayStore()
    await expect(store.exportMP4()).rejects.toThrow(/ffmpeg/i)
    expect(store.exporting).toBe(false)
  })
})