// Overlay store. Tracks whether the window is in overlay mode
// (frameless, always-on-top, transparent) or normal mode.
// Handles Esc key and ~5s inactivity auto-dismiss.

import { ipc } from '../ipc/client'

class OverlayStore {
  active = $state<boolean>(false)

  private cleanups: Array<() => void> = []
  private inactivityTimer: ReturnType<typeof setTimeout> | null = null
  private lastActivity = 0

  // Inactivity timeout before auto-dismiss (5 seconds).
  private static readonly INACTIVITY_MS = 5000

  start(): void {
    // Listen for overlay show/hide events from the daemon.
    this.cleanups.push(
      ipc.on('connected', () => {
        void ipc.call('presence.state', {}).then((state: unknown) => {
          this.active = state === 'active'
        }).catch(() => {})
      })
    )

    // Global keyboard handler: Esc dismisses overlay.
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && this.active) {
        this.hide()
      }
    }
    window.addEventListener('keydown', onKeyDown)
    this.cleanups.push(() => window.removeEventListener('keydown', onKeyDown))

    // Track user activity for inactivity timer.
    const onActivity = () => {
      if (this.active) {
        this.lastActivity = Date.now()
        this.resetInactivityTimer()
      }
    }
    window.addEventListener('mousemove', onActivity)
    window.addEventListener('keydown', onActivity)
    window.addEventListener('click', onActivity)
    this.cleanups.push(() => {
      window.removeEventListener('mousemove', onActivity)
      window.removeEventListener('keydown', onActivity)
      window.removeEventListener('click', onActivity)
    })
  }

  stop(): void {
    this.cleanups.forEach((c) => c())
    this.cleanups = []
    this.clearInactivityTimer()
    this.active = false
  }

  show(): void {
    this.active = true
    this.lastActivity = Date.now()
    this.resetInactivityTimer()
    void ipc.overlayShow().catch(() => {})
  }

  hide(): void {
    this.active = false
    this.clearInactivityTimer()
    void ipc.overlayHide().catch(() => {})
  }

  toggle(): void {
    if (this.active) {
      this.hide()
    } else {
      this.show()
    }
  }

  private resetInactivityTimer(): void {
    this.clearInactivityTimer()
    this.inactivityTimer = setTimeout(() => {
      // Auto-dismiss if no activity for INACTIVITY_MS.
      if (this.active && Date.now() - this.lastActivity >= OverlayStore.INACTIVITY_MS) {
        this.hide()
      }
    }, OverlayStore.INACTIVITY_MS)
  }

  private clearInactivityTimer(): void {
    if (this.inactivityTimer) {
      clearTimeout(this.inactivityTimer)
      this.inactivityTimer = null
    }
  }
}

export const overlay = new OverlayStore()
