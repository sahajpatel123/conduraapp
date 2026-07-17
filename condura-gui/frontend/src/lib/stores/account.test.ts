import { beforeEach, describe, expect, it, vi } from 'vitest'

const { accountStatus, accountLogout } = vi.hoisted(() => ({
  accountStatus: vi.fn(),
  accountLogout: vi.fn(),
}))

vi.mock('../ipc/client', () => ({
  ipc: {
    accountStatus,
    accountLogout,
  },
}))

import { AccountStore } from './account.svelte'

describe('AccountStore', () => {
  beforeEach(() => {
    accountStatus.mockReset()
    accountLogout.mockReset()
  })

  it('starts in the unknown state', () => {
    const store = new AccountStore()
    expect(store.status).toBeNull()
    expect(store.isSignedIn).toBe(false)
    expect(store.email).toBe('')
    expect(store.provider).toBe('')
    expect(store.loading).toBe(false)
  })

  it('checkStatus populates status on daemon success', async () => {
    accountStatus.mockResolvedValue({
      signed_in: true,
      email: 'sahaj@condura.app',
      provider: 'github',
      avatar_url: 'https://example.com/avatar.png',
      display_name: 'Sahaj Patel',
      tier: 'pro',
      created_at: '2026-01-01T00:00:00Z',
      providers: ['google', 'github'],
    })
    const store = new AccountStore()
    await store.checkStatus()
    expect(store.isSignedIn).toBe(true)
    expect(store.email).toBe('sahaj@condura.app')
    expect(store.provider).toBe('github')
    expect(store.tier).toBe('pro')
    expect(store.avatarURL).toBe('https://example.com/avatar.png')
    expect(store.displayName).toBe('Sahaj Patel')
    expect(store.configuredProviders).toEqual(['google', 'github'])
  })

  it('checkStatus falls back to signed-out on daemon error', async () => {
    accountStatus.mockRejectedValue(new Error('IPC client not started'))
    const store = new AccountStore()
    await store.checkStatus()
    expect(store.isSignedIn).toBe(false)
    expect(store.error).toMatch(/IPC client not started/i)
  })

  it('signOut calls daemon and clears local status', async () => {
    accountStatus.mockResolvedValue({
      signed_in: true,
      email: 'sahaj@condura.app',
      provider: 'github',
      avatar_url: '',
      display_name: 'Sahaj',
      tier: 'pro',
      created_at: '2026-01-01T00:00:00Z',
      providers: [],
    })
    accountLogout.mockResolvedValue({ ok: true })
    const store = new AccountStore()
    await store.checkStatus()
    expect(store.isSignedIn).toBe(true)
    await store.signOut()
    expect(accountLogout).toHaveBeenCalled()
    expect(store.isSignedIn).toBe(false)
    expect(store.email).toBe('')
  })
})