// Overlay store. Tracks whether the window is in overlay mode
// (frameless, always-on-top, transparent) or normal mode.
// Handles Esc key and ~5s inactivity auto-dismiss.

import { EventsOn } from '../../../wailsjs/runtime/runtime'
import {
  CloseQuickPrompt,
  OpenQuickPrompt,
  ToggleQuickPrompt,
} from '../../../wailsjs/go/main/App'
import { ipc } from '../ipc/client'

class OverlayStore {
  active = $state<boolean>(false)

  private cleanups: Array<() => void> = []
  private inactivityTimer: ReturnType<typeof setTimeout> | null = null
  private lastActivity = 0

  // Inactivity timeout before auto-dismiss (5 seconds).
  private static readonly INACTIVITY_MS = 5000

  start(): void {
    // Go → JS sync when menu bar, tray, or global hotkey opens the prompt.
    try {
      const off = EventsOn('condura:overlay', (data: { active?: boolean }) => {
        if (typeof data?.active === 'boolean') {
          this.setFromHost(data.active)
        }
      })
      this.cleanups.push(off)
    } catch {
      // Not running inside Wails (unit tests / static preview).
    }

    // Listen for overlay show/hide events from the daemon.
    this.cleanups.push(
      ipc.on('connected', () => {
        void ipc.call('presence.state', {}).then((state: unknown) => {
          this.setFromHost(state === 'active')
        }).catch(() => {})
      }),
      ipc.on('disconnected', () => {
        // Daemon gone — dismiss overlay to avoid stale state.
        if (this.active) {
          this.hide()
        }
      })
    )

    // Global keyboard handler: Esc dismisses overlay.
    const onKeyDown = (e: KeyboardEvent) => {
      if (e.key === 'Escape' && this.active) {
        this.hide()
        return
      }
      // Linux has no global hotkey yet — Ctrl+S toggles when the app
      // is focused (matches the default overlay hotkey).
      if (this.isQuickPromptShortcut(e)) {
        const target = e.target as HTMLElement | null
        const tag = target?.tagName
        if (tag === 'INPUT' || tag === 'TEXTAREA' || target?.isContentEditable) {
          return
        }
        e.preventDefault()
        void this.toggleFromBackend()
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
    this.setFromHost(true)
    void this.openFromBackend()
  }

  hide(): void {
    this.setFromHost(false)
    void this.closeFromBackend()
  }

  toggle(): void {
    void this.toggleFromBackend()
  }

  setFromHost(active: boolean): void {
    this.active = active
    if (active) {
      this.lastActivity = Date.now()
      this.resetInactivityTimer()
    } else {
      this.clearInactivityTimer()
    }
  }

  private isQuickPromptShortcut(e: KeyboardEvent): boolean {
    return e.ctrlKey && !e.metaKey && !e.altKey && !e.shiftKey && e.key.toLowerCase() === 's'
  }

  private async openFromBackend(): Promise<void> {
    try {
      await OpenQuickPrompt()
    } catch {
      await ipc.overlayShow().catch(() => {})
    }
  }

  private async closeFromBackend(): Promise<void> {
    try {
      await CloseQuickPrompt()
    } catch {
      await ipc.overlayHide().catch(() => {})
    }
  }

  private async toggleFromBackend(): Promise<void> {
    try {
      await ToggleQuickPrompt()
    } catch {
      if (this.active) {
        await this.closeFromBackend()
      } else {
        await this.openFromBackend()
      }
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
