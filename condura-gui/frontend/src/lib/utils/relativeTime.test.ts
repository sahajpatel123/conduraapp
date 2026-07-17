import { describe, it, expect } from 'vitest'
import { formatRelativeTime } from './relativeTime'

// formatRelativeTime — used by Sync and Channels "Last refreshed Xm ago".
// Pins the contract: short form ("5s" / "2m" / "1h"), clamps
// negative deltas, and the now() injection lets tests pin specific
// clock readings.

describe('formatRelativeTime', () => {
  const NOW = 1_000_000_000_000 // fixed clock

  it('returns seconds when under a minute', () => {
    expect(formatRelativeTime(NOW - 5_000, NOW)).toBe('5s ago')
    expect(formatRelativeTime(NOW - 59_000, NOW)).toBe('59s ago')
  })

  it('returns minutes when under an hour', () => {
    expect(formatRelativeTime(NOW - 60_000, NOW)).toBe('1m ago')
    expect(formatRelativeTime(NOW - 5 * 60_000, NOW)).toBe('5m ago')
    expect(formatRelativeTime(NOW - 59 * 60_000, NOW)).toBe('59m ago')
  })

  it('returns hours when over an hour', () => {
    expect(formatRelativeTime(NOW - 60 * 60_000, NOW)).toBe('1h ago')
    expect(formatRelativeTime(NOW - 25 * 60 * 60_000, NOW)).toBe('25h ago')
  })

  it('clamps negative deltas (clock drift, future timestamps)', () => {
    expect(formatRelativeTime(NOW + 5_000, NOW)).toBe('0s ago')
  })

  it('uses Date.now() when no now is provided', () => {
    const before = Date.now()
    const result = formatRelativeTime(before - 1_000)
    const after = Date.now()
    // Should be a string ending in "s ago"
    expect(result).toMatch(/^\d+s ago$/)
    // Should be 1s or 2s depending on timing
    expect(result === '1s ago' || result === '2s ago').toBe(true)
  })
})