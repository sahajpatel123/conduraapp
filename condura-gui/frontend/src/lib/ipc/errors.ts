/** Shared transport/offline error helpers for Meridian stores + panes. */

const OFFLINE_RE =
  /Load failed|Failed to fetch|NetworkError|TypeError:\s*Failed to fetch|ECONNREFUSED|not connected|IPC client not started|daemon/i

export function isOfflineError(err: unknown): boolean {
  return OFFLINE_RE.test(String(err ?? ''))
}

/** Strip `Error: IPC -32603: ` noise from ipc client throws. */
export function stripIpcErrorPrefix(raw: string): string {
  return raw
    .replace(/^Error:\s*/i, '')
    .replace(/^IPC\s*-?\d+:\s*/i, '')
    .trim()
}

/**
 * Turn raw fetch/TypeError noise into a short human line.
 * Strips JSON-RPC code prefixes; never surfaces bare `internal error`.
 */
export function humanizeIpcError(
  err: unknown,
  offlineMessage = 'Daemon offline — start Condura and try again'
): string {
  const raw = String(err ?? '').trim()
  if (!raw) return offlineMessage
  if (isOfflineError(raw)) return offlineMessage
  const cleaned = stripIpcErrorPrefix(raw)
  if (!cleaned) return offlineMessage
  if (isOfflineError(cleaned)) return offlineMessage
  if (cleaned.toLowerCase() === 'internal error') {
    return 'Something went wrong on this machine — try again in a moment.'
  }
  return cleaned
}

const HUB_OFFLINE = 'Daemon offline — the shelf will load when Condura reconnects'

/** Calm shelf copy for hub.search / hub.install failures. */
export function humanizeHubError(err: unknown): string {
  const raw = String(err ?? '').trim()
  if (!raw) return HUB_OFFLINE
  if (isOfflineError(raw)) return HUB_OFFLINE

  const msg = stripIpcErrorPrefix(raw).toLowerCase()
  if (msg === 'hub not configured' || msg.includes('hub not configured')) {
    return "Community hub isn't enabled — turn it on in Settings to browse the shelf."
  }
  if (msg.includes('sign-in') || msg.includes('authentication required')) {
    return 'Community hub needs a sign-in token — add one in Settings.'
  }
  if (
    msg === 'internal error' ||
    msg.includes('community hub unreachable') ||
    msg.includes('hub unreachable') ||
    msg.startsWith('hub search:') ||
    msg.startsWith('hub get:')
  ) {
    return 'Community hub is unreachable — check your connection and try again.'
  }
  return humanizeIpcError(raw, HUB_OFFLINE)
}
