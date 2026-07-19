// Package backup provides the encrypted backup and restore
// functionality for Condura.
//
// Architecture:
//   - backup.go: BackupManager — the high-level API for creating
//     and inspecting backups.
//   - dir.go: ResolveBackupDir — resolves the standard backup
//     location (~/Documents/condura-backups per MISSION §24.1,
//     overridable via CONDURA_BACKUP_DIR).
//   - restore.go: Restore — reads an archive and atomically
//     swaps the data dir contents.
//   - rollback.go: NewRollbackMulti — action-replay rollback
//     support for in-flight chat transactions.
//
// All file I/O goes through the stdlib (os, archive/zip). The
// encryption layer is the AES-256-GCM key in the data dir's
// secrets.json. The package is local-only — no daemon IPC
// required.
package backup
