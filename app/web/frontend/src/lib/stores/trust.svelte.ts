// Trust & Recovery store — backup list and OS permission probes.

import { ipc } from '../ipc/client'
import type { BackupEntry, PermissionGuide, PermissionStatus } from '../ipc/types'

class TrustStore {
  backups = $state<BackupEntry[]>([])
  permissions = $state<PermissionStatus[]>([])
  loadingBackups = $state(false)
  loadingPermissions = $state(false)
  lastError = $state('')

  async refreshBackups(): Promise<void> {
    this.loadingBackups = true
    this.lastError = ''
    try {
      this.backups = await ipc.backupList()
    } catch (err) {
      this.lastError = String(err)
      this.backups = []
    } finally {
      this.loadingBackups = false
    }
  }

  async createBackup(): Promise<string> {
    const r = await ipc.backupCreate()
    await this.refreshBackups()
    return r.path
  }

  async refreshPermissions(): Promise<void> {
    this.loadingPermissions = true
    this.lastError = ''
    try {
      this.permissions = await ipc.permissionsStatus()
    } catch (err) {
      this.lastError = String(err)
      this.permissions = []
    } finally {
      this.loadingPermissions = false
    }
  }

  async loadGuide(kind: string): Promise<PermissionGuide> {
    return ipc.permissionsGuide(kind)
  }
}

export const trust = new TrustStore()
