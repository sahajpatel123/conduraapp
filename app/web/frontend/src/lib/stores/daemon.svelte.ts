// Daemon connection store. Tracks whether the IPC client is
// connected, the last error, and the configured address.

import { ipc } from '../ipc/client'

class DaemonStore {
  baseURL = $state<string>('')
  authToken = $state<string>('')
  connected = $state<boolean>(false)
  lastError = $state<string>('')
  reconnectAttempt = $state<number>(0)
  reconnectDelayMs = $state<number>(0)

  private cleanups: Array<() => void> = []

  configure(opts: { baseURL: string; authToken: string }): void {
    this.baseURL = opts.baseURL
    this.authToken = opts.authToken
  }

  start(): void {
    if (!this.baseURL) {
      this.lastError = 'no daemon address configured'
      return
    }
    this.cleanups.push(
      ipc.on('connected', () => {
        this.connected = true
        this.lastError = ''
        this.reconnectAttempt = 0
      }),
      ipc.on('disconnected', (reason) => {
        this.connected = false
        this.lastError = reason
      }),
      ipc.on('reconnecting', (attempt, delayMs) => {
        this.reconnectAttempt = attempt
        this.reconnectDelayMs = delayMs
      })
    )
    void ipc.start({ baseURL: this.baseURL, authToken: this.authToken })
  }

  stop(): void {
    this.cleanups.forEach((c) => c())
    this.cleanups = []
    ipc.stop()
    this.connected = false
  }
}

export const daemon = new DaemonStore()
