import { beforeEach, describe, expect, it, vi } from 'vitest'

const { hubSearch, skillsList } = vi.hoisted(() => ({
  hubSearch: vi.fn(),
  skillsList: vi.fn(),
}))

vi.mock('../ipc/client', () => ({
  ipc: {
    hubSearch,
    skillsList,
    hubInstall: vi.fn(),
    hubPublishTyped: vi.fn(),
  },
}))

import { HubStore } from './hub.svelte'

describe('HubStore.search', () => {
  beforeEach(() => {
    hubSearch.mockReset()
    skillsList.mockReset()
  })

  it('browses the shelf on empty query (does not short-circuit)', async () => {
    hubSearch.mockResolvedValue({
      skills: [{ id: 'weather', name: 'Weather', description: '', author: '', trust: 'community' }],
      total: 1,
      query: '',
    })
    const store = new HubStore()
    await store.search('', 24)
    expect(hubSearch).toHaveBeenCalledWith('', 24)
    expect(store.results).toHaveLength(1)
    expect(store.results[0]?.id).toBe('weather')
    expect(store.lastQuery).toBe('')
    expect(store.loading).toBe(false)
  })

  it('drops stale responses when a newer query is in flight', async () => {
    let resolveSlow: (v: unknown) => void = () => {}
    hubSearch.mockImplementationOnce(
      () =>
        new Promise((resolve) => {
          resolveSlow = resolve
        })
    )
    hubSearch.mockResolvedValueOnce({
      skills: [{ id: 'fast', name: 'Fast', description: '', author: '', trust: 'community' }],
      total: 1,
      query: 'fast',
    })

    const store = new HubStore()
    const slow = store.search('slow', 10)
    const fast = store.search('fast', 10)
    await fast
    resolveSlow({
      skills: [{ id: 'slow', name: 'Slow', description: '', author: '', trust: 'community' }],
      total: 1,
      query: 'slow',
    })
    await slow
    expect(store.results[0]?.id).toBe('fast')
    expect(store.lastQuery).toBe('fast')
  })

  it('refreshInstalled hydrates id and hub_id', async () => {
    skillsList.mockResolvedValue([
      { id: 'local-1', name: 'A', description: '', version: '', author: '', license: '', trust: '', hub_id: 'hub-1' },
      { id: 'local-2', name: 'B', description: '', version: '', author: '', license: '', trust: '' },
    ])
    const store = new HubStore()
    await store.refreshInstalled()
    expect(store.installed.has('local-1')).toBe(true)
    expect(store.installed.has('hub-1')).toBe(true)
    expect(store.installed.has('local-2')).toBe(true)
  })

  it('humanizes hub.search IPC failures', async () => {
    hubSearch.mockRejectedValue(new Error('IPC -32603: internal error'))
    const store = new HubStore()
    await store.search('', 24)
    expect(store.error).toBe(
      'Community hub is unreachable — check your connection and try again.'
    )
    expect(store.results).toEqual([])
  })

  it('humanizes hub not configured', async () => {
    hubSearch.mockRejectedValue(new Error('IPC -32603: hub not configured'))
    const store = new HubStore()
    await store.search('test', 10)
    expect(store.error).toBe(
      "Community hub isn't enabled — turn it on in Settings to browse the shelf."
    )
  })
})
