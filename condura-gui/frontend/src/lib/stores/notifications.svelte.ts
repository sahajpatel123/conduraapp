// Toast / banner notifications. Auto-dismisses after a few seconds.

export type NotificationKind = 'info' | 'warn' | 'error' | 'success'

export interface Notification {
  id: number
  kind: NotificationKind
  title: string
  message: string
  createdAt: number
  // For errors that should stick until acknowledged.
  sticky: boolean
}

class NotificationStore {
  list = $state<Notification[]>([])
  private nextId = 1
  /**
   * Per-kind TTL (ms). Success is shorter (acknowledged signal),
   * warn/error stay longer so users can actually read them. Sticky
   * notifications bypass this entirely.
   */
  private ttlByKind: Record<NotificationKind, number> = {
    info: 5000,
    success: 4000,
    warn: 8000,
    error: 10000,
  }

  push(opts: { kind: NotificationKind; title: string; message: string; sticky?: boolean }): number {
    const n: Notification = {
      id: this.nextId++,
      kind: opts.kind,
      title: opts.title,
      message: opts.message,
      createdAt: Date.now(),
      sticky: !!opts.sticky
    }
    this.list = [...this.list, n]
    if (!n.sticky) {
      setTimeout(() => this.dismiss(n.id), this.ttlByKind[n.kind])
    }
    return n.id
  }

  dismiss(id: number): void {
    this.list = this.list.filter((n) => n.id !== id)
  }

  clear(): void {
    this.list = []
  }
}

export const notifications = new NotificationStore()
