import { describe, it, expect, vi, beforeEach, afterEach } from 'vitest'
import { notifications } from './notifications.svelte'

// notifications store — used by MeridianToasts.
//
// Pins the per-kind TTL contract and the sticky behavior. Auto-dismiss
// runs via setTimeout inside push(); we use vitest fake timers to drive
// the timer without slowing the suite.

describe('notifications store', () => {
  beforeEach(() => {
    notifications.clear()
    vi.useFakeTimers()
  })

  afterEach(() => {
    notifications.clear()
    vi.useRealTimers()
  })

  it('starts empty', () => {
    expect(notifications.list).toEqual([])
  })

  it('push assigns a numeric id and adds to the list', () => {
    const id = notifications.push({ kind: 'info', title: 'hi', message: '' })
    expect(typeof id).toBe('number')
    expect(notifications.list).toHaveLength(1)
    expect(notifications.list[0]?.title).toBe('hi')
  })

  it('info notifications auto-dismiss after ~5s', () => {
    notifications.push({ kind: 'info', title: 'flash', message: '' })
    expect(notifications.list).toHaveLength(1)
    vi.advanceTimersByTime(4999)
    expect(notifications.list).toHaveLength(1)
    vi.advanceTimersByTime(2)
    expect(notifications.list).toHaveLength(0)
  })

  it('success notifications auto-dismiss after ~4s (shorter than info)', () => {
    notifications.push({ kind: 'success', title: 'done', message: '' })
    vi.advanceTimersByTime(3999)
    expect(notifications.list).toHaveLength(1)
    vi.advanceTimersByTime(2)
    expect(notifications.list).toHaveLength(0)
  })

  it('warn notifications stay for ~8s', () => {
    notifications.push({ kind: 'warn', title: 'careful', message: '' })
    vi.advanceTimersByTime(7999)
    expect(notifications.list).toHaveLength(1)
    vi.advanceTimersByTime(2)
    expect(notifications.list).toHaveLength(0)
  })

  it('error notifications stay for ~10s (longest non-sticky)', () => {
    notifications.push({ kind: 'error', title: 'broken', message: '' })
    vi.advanceTimersByTime(9999)
    expect(notifications.list).toHaveLength(1)
    vi.advanceTimersByTime(2)
    expect(notifications.list).toHaveLength(0)
  })

  it('sticky notifications never auto-dismiss', () => {
    notifications.push({ kind: 'error', title: 'permanent', message: '', sticky: true })
    vi.advanceTimersByTime(60_000)
    expect(notifications.list).toHaveLength(1)
  })

  it('dismiss removes a specific id without affecting others', () => {
    const a = notifications.push({ kind: 'info', title: 'a', message: '' })
    notifications.push({ kind: 'info', title: 'b', message: '' })
    notifications.dismiss(a)
    expect(notifications.list).toHaveLength(1)
    expect(notifications.list[0]?.title).toBe('b')
  })

  it('clear empties the list', () => {
    notifications.push({ kind: 'info', title: 'x', message: '' })
    notifications.push({ kind: 'warn', title: 'y', message: '' })
    notifications.clear()
    expect(notifications.list).toEqual([])
  })

  it('ids increment monotonically across pushes', () => {
    const a = notifications.push({ kind: 'info', title: 'a', message: '' })
    const b = notifications.push({ kind: 'info', title: 'b', message: '' })
    expect(b).toBeGreaterThan(a)
  })
})