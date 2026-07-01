package daemon

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/backup"
	"github.com/sahajpatel123/conduraapp/internal/ipc"
	"github.com/sahajpatel123/conduraapp/internal/sanitize"
)

// registerBackupMethods wires the backup.* and uninstall.* RPC
// methods (Phase 11, sub-phases 11B, 11C, 11D). Backup creation
// is a routine operation. Restore and rollback are gated —
// they always go through the Gatekeeper before executing.
//
// Method names:
//   - backup.list          — list local backup archives
//   - backup.preview       — describe a backup without decrypting
//   - backup.derive_key    — return the on-first-backup key (base64)
//   - backup.create        — create a new encrypted backup
//   - backup.restore       — restore a backup (gated)
//   - backup.rollback      — revert last session (gated)
//
//nolint:gocognit,gocyclo
func registerBackupMethods(srv *ipc.Server, subs *Subsystems) {
	if subs.Backup == nil {
		// Register stubs that return "not available" so the
		// GUI can probe the subsystem before showing the
		// backup panel.
		notAvailable := func(_ context.Context, _ json.RawMessage) (any, error) {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "backup subsystem not available"}
		}
		srv.Register("backup.list", notAvailable)
		srv.Register("backup.preview", notAvailable)
		srv.Register("backup.derive_key", notAvailable)
		srv.Register("backup.create", notAvailable)
		srv.Register("backup.restore", notAvailable)
		srv.Register("backup.rollback", notAvailable)
		return
	}

	// backup.list — list backup archives in the standard
	// backup directory.
	srv.Register("backup.list", func(_ context.Context, _ json.RawMessage) (any, error) {
		dir := backupDir(subs)
		entries, err := listBackupArchives(dir)
		if err != nil {
			return nil, err
		}
		return entries, nil
	})

	// backup.preview — describe a backup archive (manifest
	// fields, file list) without decrypting.
	srv.Register("backup.preview", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if p.Path == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "path is required"}
		}
		m, err := backup.LoadManifest(p.Path)
		if err != nil {
			return nil, fmt.Errorf("backup: load manifest: %w", err)
		}
		return m, nil
	})

	// backup.derive_key — return the derived backup key as
	// base64. Surfaced to the GUI on the first backup so the
	// user can save it. We DO NOT log this key; the GUI shows
	// it directly to the user.
	srv.Register("backup.derive_key", func(_ context.Context, _ json.RawMessage) (any, error) {
		mk, err := subs.MasterKey()
		if err != nil || len(mk) != 32 {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "master key unavailable"}
		}
		k, err := backup.DeriveKeyBase64(mk)
		if err != nil {
			return nil, err
		}
		return map[string]any{"key": k}, nil
	})

	// backup.create — create a new encrypted backup.
	srv.Register("backup.create", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Destination string `json:"destination"`
		}
		_ = json.Unmarshal(params, &p)
		path, err := subs.Backup.Create(ctx)
		if err != nil {
			return nil, fmt.Errorf("backup: create: %w", err)
		}
		if p.Destination != "" {
			if err := copyFile(path, p.Destination); err != nil {
				return nil, fmt.Errorf("backup: copy to destination: %w", err)
			}
			path = p.Destination
		}
		return map[string]any{"path": path}, nil
	})

	// backup.restore — restore a backup archive. GATED through
	// the Gatekeeper. The GUI should prompt the user first;
	// the Gatekeeper is the second line of defense.
	srv.Register("backup.restore", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Path string `json:"path"`
		}
		if err := json.Unmarshal(params, &p); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: err.Error()}
		}
		if p.Path == "" {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: "path is required"}
		}
		if !subs.GatekeeperAllow(ctx, "backup.restore", "Restore backup from "+p.Path) {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: msgDeniedBySafetyPolicy}
		}
		mk, err := subs.MasterKey()
		if err != nil || len(mk) != 32 {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "master key unavailable"}
		}
		// Close all database connections before the atomic swap so
		// Windows file locks are released. The subsequent
		// Storage.Reload reopens the main DB on the restored files.
		if subs.Storage != nil {
			// Force a WAL checkpoint so data is flushed from
			// the WAL into the main DB file.
			_, _ = subs.Storage.SQL().ExecContext(ctx, "PRAGMA wal_checkpoint(TRUNCATE)")
		}
		subs.CloseDatabases()
		// Remove WAL/SHM sidecar files if they still exist.
		// On Windows these can hold file locks even after Close.
		dbPath := subs.Storage.Path()
		os.Remove(dbPath + "-wal") //nolint:errcheck
		os.Remove(dbPath + "-shm") //nolint:errcheck
		// Also clean up WAL/SHM for memory.db and skills.db.
		dataDir := filepath.Dir(dbPath)
		os.Remove(filepath.Join(dataDir, "memory.db-wal")) //nolint:errcheck
		os.Remove(filepath.Join(dataDir, "memory.db-shm")) //nolint:errcheck
		os.Remove(filepath.Join(dataDir, "skills.db-wal")) //nolint:errcheck
		os.Remove(filepath.Join(dataDir, "skills.db-shm")) //nolint:errcheck
		// Create a pre-restore safety snapshot so the user can
		// recover from a bad restore. The snapshot is written
		// into the backup directory (same one backup.create uses).
		preRestorePath := filepath.Join(
			backup.ResolveBackupDir(subs.GeneralDataDir()),
			fmt.Sprintf("pre-restore-%s.zip", time.Now().UTC().Format("20060102-150405Z")),
		)
		err = backup.Restore(ctx, backup.RestoreOptions{
			ArchivePath:          p.Path,
			DataDir:              subs.GeneralDataDir(),
			MasterKey:            mk,
			CurrentSchemaVersion: currentSchemaVersion(subs),
			PreRestoreBackupPath: preRestorePath,
		})
		if err != nil {
			return nil, fmt.Errorf("backup: restore: %w", err)
		}
		// The atomic swap inside Restore moved the restored
		// files into the data dir, but the daemon's open
		// SQLite handle is still bound to the old (unlinked)
		// inode. Reload the storage handle so subsequent
		// queries see the restored data immediately, without
		// requiring the user to restart the daemon.
		if subs.Storage != nil {
			if rerr := subs.Storage.Reload(ctx); rerr != nil {
				_ = subs.Audit.Append(ctx, buildAuditEvent("backup.restore.reload_failed", appCondurad, auditResultError, sanitize.RedactSecrets(rerr.Error())))
				return nil, fmt.Errorf("backup: restore succeeded on disk but storage reload failed: %w (daemon restart required to see restored data)", rerr)
			}
		}
		// Reload memory + skills databases so subsequent RPC
		// calls (memory.recall, skills.list) see the restored
		// data, not the stale handles from before the swap.
		if rerr := subs.ReloadAuxiliaryDatabases(); rerr != nil {
			_ = subs.Audit.Append(ctx, buildAuditEvent("backup.restore.aux_reload_failed", appCondurad, auditResultError, sanitize.RedactSecrets(rerr.Error())))
			return nil, fmt.Errorf("backup: restore succeeded but auxiliary db reload failed: %w", rerr)
		}
		// Run integrity checks on all three restored databases.
		// If any database is corrupt, report it (best-effort;
		// we don't abort since the main data is already swapped in).
		if ierr := runPostRestoreIntegrityChecks(ctx, subs); ierr != nil {
			_ = subs.Audit.Append(ctx, buildAuditEvent("backup.restore.integrity_warning", appCondurad, auditResultError, sanitize.RedactSecrets(ierr.Error())))
		}
		_ = subs.Audit.Append(ctx, buildAuditEvent("backup.restore", appCondurad, auditResultAllow, sanitize.RedactSecrets("path="+p.Path)))
		return auditOK(), nil
	})

	// backup.rollback — revert the last session's writes via
	// the rollback checkpoint. GATED through the Gatekeeper.
	srv.Register("backup.rollback", func(ctx context.Context, _ json.RawMessage) (any, error) {
		if !subs.GatekeeperAllow(ctx, "backup.rollback", "Revert last session") {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: msgDeniedBySafetyPolicy}
		}
		memDB := OpenRollbackDB(subs.MemoryDBPath())
		skillDB := OpenRollbackDB(subs.SkillDBPath())
		rb := backup.NewRollbackMulti(
			subs.Storage.SQL(),
			memDB,
			skillDB,
		)
		rb.TrackOpened(memDB, skillDB)
		defer func() { _ = rb.Close() }()
		if subs.cfg != nil && subs.cfg.Storage.Backup.RollbackWindow > 0 {
			rb.SetWindow(subs.cfg.Storage.Backup.RollbackWindow)
		}
		n, err := rb.RevertLastSession(ctx)
		if err != nil {
			return nil, fmt.Errorf("backup: rollback: %w", err)
		}
		_ = subs.Audit.Append(ctx, buildAuditEvent("backup.rollback", appCondurad, auditResultAllow, fmt.Sprintf("reverted %d rows", n)))
		return map[string]any{"reverted_rows": n, "honest_scope": rb.HonestScope()}, nil
	})
}

// runPostRestoreIntegrityChecks runs PRAGMA integrity_check on
// all three databases after a restore. Returns a combined error
// if any database fails the check. This is best-effort: we report
// but don't abort, since the data is already on disk.
func runPostRestoreIntegrityChecks(ctx context.Context, subs *Subsystems) error {
	var errs []error
	if subs.Storage != nil {
		if err := checkSQLiteIntegrity(ctx, subs.Storage.SQL(), "synaptic.db"); err != nil {
			errs = append(errs, err)
		}
	}
	if subs.Memory != nil {
		memPath := subs.MemoryDBPath()
		if memPath != "" {
			db, err := openIntegrityDB(memPath)
			if err == nil {
				defer func() { _ = db.Close() }()
				if err := checkSQLiteIntegrity(ctx, db, "memory.db"); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}
	if subs.Phase12 != nil && subs.Phase12.SkillStore != nil {
		skillPath := subs.SkillDBPath()
		if skillPath != "" {
			db, err := openIntegrityDB(skillPath)
			if err == nil {
				defer func() { _ = db.Close() }()
				if err := checkSQLiteIntegrity(ctx, db, "skills.db"); err != nil {
					errs = append(errs, err)
				}
			}
		}
	}
	if len(errs) > 0 {
		return fmt.Errorf("integrity check: %v", errs)
	}
	return nil
}

// checkSQLiteIntegrity runs PRAGMA integrity_check on a single db.
func checkSQLiteIntegrity(ctx context.Context, db *sql.DB, name string) error {
	var result string
	if err := db.QueryRowContext(ctx, "PRAGMA integrity_check").Scan(&result); err != nil {
		return fmt.Errorf("%s: integrity_check query failed: %w", name, err)
	}
	if result != "ok" {
		return fmt.Errorf("%s: integrity_check returned %q", name, result)
	}
	return nil
}

// openIntegrityDB opens a read-only SQLite connection for integrity checking.
func openIntegrityDB(path string) (*sql.DB, error) {
	return sql.Open("sqlite3", "file:"+path+"?mode=ro")
}

// backupDir is the on-disk location backup archives are written to.
// Uses backup.ResolveBackupDir (Documents/condura-backups by default).
func backupDir(subs *Subsystems) string {
	if subs == nil {
		return ""
	}
	return backup.ResolveBackupDir(subs.GeneralDataDir())
}

// listBackupArchives returns a sorted list of .zip files in
// the backup dir along with their size in bytes.
type backupEntry struct {
	Name string `json:"name"`
	Path string `json:"path"`
	Size int64  `json:"size"`
}

func listBackupArchives(dir string) ([]backupEntry, error) {
	if dir == "" {
		return []backupEntry{}, nil
	}
	entries, err := readDirNames(dir)
	if err != nil {
		return nil, fmt.Errorf("backup: list: %w", err)
	}
	sort.Strings(entries)
	out := make([]backupEntry, 0, len(entries))
	for _, name := range entries {
		if !strings.HasSuffix(name, ".zip") {
			continue
		}
		full := filepath.Join(dir, name)
		size, _ := fileSize(full)
		out = append(out, backupEntry{Name: name, Path: full, Size: size})
	}
	return out, nil
}
