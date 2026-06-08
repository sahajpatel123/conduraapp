// API keys store. Caches the list of stored API key metadata.

import { ipc } from '../ipc/client'
import type { APIKeyMeta } from '../ipc/types'

class APIKeysStore {
  list = $state<APIKeyMeta[]>([])
  loading = $state<boolean>(false)
  saving = $state<boolean>(false)
  lastError = $state<string>('')

  async refresh(): Promise<void> {
    this.loading = true
    try {
      this.list = await ipc.apiKeysList()
    } finally {
      this.loading = false
    }
  }

  async set(provider: string, label: string, secret: string): Promise<void> {
    this.saving = true
    this.lastError = ''
    try {
      await ipc.apiKeysSet({ provider, label, secret })
      await this.refresh()
    } catch (err) {
      this.lastError = String(err)
      throw err
    } finally {
      this.saving = false
    }
  }

  async remove(id: number): Promise<void> {
    await ipc.apiKeysDelete(id)
    this.list = this.list.filter((k) => k.id !== id)
  }
}

export const apiKeys = new APIKeysStore()
