// Relative time formatter — turns a unix-ms timestamp into
// "5s ago" / "2m ago" / "1h ago" for "Last refreshed" indicators.
// Used by Sync, Channels, and any other surface that needs to
// surface how stale its data is without claiming real-time
// precision.

/** Format the time elapsed since `t` (unix ms) as a short
 *  human-readable string. Negative deltas clamp to "0s ago"
 *  (clocks can drift, futures shouldn't render minus-time). */
export function formatRelativeTime(t: number, now: number = Date.now()): string {
  const sec = Math.max(0, Math.floor((now - t) / 1000))
  if (sec < 60) return `${sec}s ago`
  const min = Math.floor(sec / 60)
  if (min < 60) return `${min}m ago`
  const hr = Math.floor(min / 60)
  return `${hr}h ago`
}