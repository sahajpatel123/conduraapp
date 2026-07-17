import { beforeEach, describe, expect, it, vi } from 'vitest'

const { apiKeysList, apiKeysSet, apiKeysDelete } = vi.hoisted(() => ({
  apiKeysList: vi.fn(),
  apiKeysSet: vi.fn(),
  apiKeysDelete: vi.fn(),
}))

vi.mock('../ipc/client', () => ({
  ipc: {
    apiKeysList,
    apiKeysSet,
    apiKeysDelete,
  },
}))

import { APIKeysStore } from './apikeys.svelte'

describe('APIKeysStore', () => {
  beforeEach(() => {
    apiKeysList.mockReset()
    apiKeysSet.mockReset()
    apiKeysDelete.mockReset()
  })

  it('starts with empty list and no error', () => {
    const store = new APIKeysStore()
    expect(store.list).toEqual([])
    expect(store.loading).toBe(false)
    expect(store.saving).toBe(false)
    expect(store.lastError).toBe('')
  })

  it('refresh populates the key list from daemon', async () => {
    apiKeysList.mockResolvedValue([
      { id: 1, provider: 'openai', label: 'default', has_token: true },
      { id: 2, provider: 'anthropic', label: 'default', has_token: true },
    ])
    const store = new APIKeysStore()
    await store.refresh()
    expect(store.list).toHaveLength(2)
    expect(store.list[0]?.provider).toBe('openai')
    expect(store.loading).toBe(false)
  })

  it('set() calls daemon + refreshes list + clears saving', async () => {
    apiKeysSet.mockResolvedValue({ id: 1 })
    apiKeysList.mockResolvedValue([{ id: 1, provider: 'openai', label: 'default', has_token: true }])
    const store = new APIKeysStore()
    await store.set('openai', 'default', 'sk-...')
    expect(apiKeysSet).toHaveBeenCalledWith({
      provider: 'openai',
      label: 'default',
      secret: 'sk-...',
    })
    expect(store.list).toHaveLength(1)
    expect(store.saving).toBe(false)
  })

  it('set() re-throws on daemon failure and surfaces error', async () => {
    apiKeysSet.mockRejectedValue(new Error('key rejected by daemon'))
    apiKeysList.mockResolvedValue([])
    const store = new APIKeysStore()
    await expect(store.set('openai', 'default', 'bad')).rejects.toThrow(/rejected/i)
    expect(store.lastError).toMatch(/rejected/i)
    expect(store.saving).toBe(false)
  })

  it('remove() deletes from list optimistically', async () => {
    apiKeysDelete.mockResolvedValue({ ok: true })
    const store = new APIKeysStore()
    store.list = [
      { id: 1, provider: 'openai', label: 'default', has_token: true },
      { id: 2, provider: 'anthropic', label: 'default', has_token: true },
    ]
    await store.remove(1)
    expect(apiKeysDelete).toHaveBeenCalledWith(1)
    expect(store.list).toHaveLength(1)
    expect(store.list[0]?.id).toBe(2)
  })
})