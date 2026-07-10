/**
 * Open the OS System Settings / Privacy pane for a permission kind.
 *
 * Prefer daemon-side open (permissions.open_settings) so deep links like
 * x-apple.systempreferences: and ms-settings: work from the Wails WebView.
 * Falls back to Wails BrowserOpenURL, then help_url.
 */
export async function openPermissionSettings(
  kind: string,
  ipc: {
    permissionsOpenSettings: (k: string) => Promise<{
      guide: { deep_link?: string; help_url?: string; steps?: string[] }
      opened: boolean
      error?: string
    }>
    permissionsGuide?: (k: string) => Promise<{ deep_link?: string; help_url?: string }>
  },
): Promise<{ opened: boolean; error?: string }> {
  try {
    if (typeof ipc.permissionsOpenSettings === 'function') {
      const res = await ipc.permissionsOpenSettings(kind)
      if (res.opened) return { opened: true }
      // Daemon could not open; try Wails / browser with the guide URL.
      const url = res.guide?.deep_link || res.guide?.help_url
      if (url && tryBrowserOpen(url)) return { opened: true }
      return { opened: false, error: res.error || 'could not open System Settings' }
    }
  } catch (e) {
    // Fall through to guide-only path.
    const msg = String(e)
    let guideErr: string | undefined
    try {
      if (ipc.permissionsGuide) {
        const g = await ipc.permissionsGuide(kind)
        const url = g.deep_link || g.help_url
        if (url && tryBrowserOpen(url)) return { opened: true }
      }
    } catch (inner) {
      // Preserve both errors so UI debugging and telemetry don't
      // lose the real failure when the daemon RPC AND the guide
      // fallback both fail.
      guideErr = String(inner)
    }
    return { opened: false, error: guideErr ? `${msg}; guide fallback: ${guideErr}` : msg }
  }
  return { opened: false, error: 'permissions open not available' }
}

function tryBrowserOpen(url: string): boolean {
  try {
    const w = window as unknown as { runtime?: { BrowserOpenURL?: (u: string) => void } }
    if (w.runtime?.BrowserOpenURL) {
      w.runtime.BrowserOpenURL(url)
      return true
    }
  } catch {
    /* ignore */
  }
  // Custom URL schemes often fail in window.open; still try as last resort.
  try {
    if (url.startsWith('http://') || url.startsWith('https://')) {
      window.open(url, '_blank', 'noopener,noreferrer')
      return true
    }
  } catch {
    /* ignore */
  }
  return false
}
