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
  private ttlMs = 5000

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
      setTimeout(() => this.dismiss(n.id), this.ttlMs)
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
