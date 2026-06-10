// Overlay store. Tracks whether the window is in overlay mode
// (frameless, always-on-top, transparent) or normal mode.

import { ipc } from '../ipc/client'

class OverlayStore {
  active = $state<boolean>(false)

  private cleanups: Array<() => void> = []

  start(): void {
    // Listen for overlay show/hide events from the daemon.
    // The conductor's hotkey triggers these via the IPC surface.
    this.cleanups.push(
      ipc.on('connected', () => {
        // Query initial state on connect.
        void ipc.call('presence.state', {}).then((state: unknown) => {
          this.active = state === 'active'
        }).catch(() => {})
      })
    )
  }

  stop(): void {
    this.cleanups.forEach((c) => c())
    this.cleanups = []
    this.active = false
  }

  show(): void {
    this.active = true
    void ipc.overlayShow().catch(() => {})
  }

  hide(): void {
    this.active = false
    void ipc.overlayHide().catch(() => {})
  }

  toggle(): void {
    if (this.active) {
      this.hide()
    } else {
      this.show()
    }
  }
}

export const overlay = new OverlayStore()
