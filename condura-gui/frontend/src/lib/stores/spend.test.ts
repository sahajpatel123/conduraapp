import { beforeEach, describe, expect, it, vi } from 'vitest'

const { spendToday, on, push } = vi.hoisted(() => ({
  spendToday: vi.fn(),
  on: vi.fn(() => () => {}),
  push: vi.fn(),
}))

vi.mock('../ipc/client', () => ({
  ipc: {
    spendToday,
    on,
  },
}))

vi.mock('./notifications.svelte', () => ({
  notifications: { push },
}))

import { spend } from './spend.svelte'

describe('SpendStore', () => {
  beforeEach(() => {
    spendToday.mockReset()
    on.mockClear()
    push.mockClear()
    spend.summary = null
    spend.live = false
    spend.stopPolling()
  })

  it('applySummary updates spent/cap/pct and warns at 80%', () => {
    spend.applySummary({ spent: 8, cap: 10, remaining: 2 })
    expect(spend.spent).toBe(8)
    expect(spend.cap).toBe(10)
    expect(spend.pct).toBe(80)
    expect(push).toHaveBeenCalledWith(
      expect.objectContaining({ kind: 'warn', title: expect.stringMatching(/approaching/i) })
    )
  })

  it('sticky error toast at 100%', () => {
    spend.applySummary({ spent: 12, cap: 10, remaining: 0 })
    expect(spend.pct).toBe(100)
    expect(push).toHaveBeenCalledWith(
      expect.objectContaining({ kind: 'error', sticky: true })
    )
  })

  it('startLive subscribes to spend_warning', () => {
    spend.startLive()
    expect(on).toHaveBeenCalledWith('spend_warning', expect.any(Function))
    expect(spend.live).toBe(true)
    spend.stopLive()
    expect(spend.live).toBe(false)
  })

  it('refresh loads spend.today', async () => {
    spendToday.mockResolvedValue({ spent: 1.5, cap: 20, remaining: 18.5 })
    await spend.refresh()
    expect(spend.spent).toBe(1.5)
    expect(spend.cap).toBe(20)
  })
})
