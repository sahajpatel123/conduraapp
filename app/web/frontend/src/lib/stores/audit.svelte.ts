// Audit log store. Loads the most recent events; supports
// pagination + filters.

import { ipc } from '../ipc/client'
import type { AuditEvent } from '../ipc/types'

class AuditStore {
  events = $state<AuditEvent[]>([])
  loading = $state<boolean>(false)
  filterAction = $state<string>('')
  filterLevel = $state<'' | 'info' | 'warn' | 'error'>('')
  offset = $state<number>(0)
  limit = $state<number>(100)
  total = $state<number>(0)

  async refresh(): Promise<void> {
    this.loading = true
    try {
      const events = await ipc.auditList({
        limit: this.limit,
        offset: this.offset,
        action: this.filterAction || undefined,
        level: this.filterLevel || undefined
      })
      this.events = events
      this.total = events.length
    } finally {
      this.loading = false
    }
  }

  setFilter(action: string, level: '' | 'info' | 'warn' | 'error'): void {
    this.filterAction = action
    this.filterLevel = level
    this.offset = 0
    void this.refresh()
  }

  nextPage(): void {
    this.offset += this.limit
    void this.refresh()
  }

  prevPage(): void {
    this.offset = Math.max(0, this.offset - this.limit)
    void this.refresh()
  }
}

export const audit = new AuditStore()
