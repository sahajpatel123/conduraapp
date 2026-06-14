package daemon

import (
	"context"
	"encoding/json"
	"fmt"
	"sort"
	"strings"

	"github.com/sahajpatel123/synapticapp/internal/backup"
	"github.com/sahajpatel123/synapticapp/internal/ipc"
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
//nolint:gocognit
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
	srv.Register("backup.create", func(ctx context.Context, _ json.RawMessage) (any, error) {
		path, err := subs.Backup.Create(ctx)
		if err != nil {
			return nil, fmt.Errorf("backup: create: %w", err)
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
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "denied by safety policy"}
		}
		mk, err := subs.MasterKey()
		if err != nil || len(mk) != 32 {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "master key unavailable"}
		}
		err = backup.Restore(ctx, backup.RestoreOptions{
			ArchivePath:          p.Path,
			DataDir:              subs.GeneralDataDir(),
			MasterKey:            mk,
			CurrentSchemaVersion: currentSchemaVersion(subs),
		})
		if err != nil {
			return nil, fmt.Errorf("backup: restore: %w", err)
		}
		_ = subs.Audit.Append(ctx, buildAuditEvent("backup.restore", appSynapticd, auditResultAllow, "path="+p.Path))
		return auditOK(), nil
	})

	// backup.rollback — revert the last session's writes via
	// the rollback checkpoint. GATED through the Gatekeeper.
	srv.Register("backup.rollback", func(ctx context.Context, _ json.RawMessage) (any, error) {
		if !subs.GatekeeperAllow(ctx, "backup.rollback", "Revert last session") {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: "denied by safety policy"}
		}
		rb := backup.NewRollback(subs.Storage.SQL())
		n, err := rb.RevertLastSession(ctx)
		if err != nil {
			return nil, fmt.Errorf("backup: rollback: %w", err)
		}
		_ = subs.Audit.Append(ctx, buildAuditEvent("backup.rollback", appSynapticd, auditResultAllow, fmt.Sprintf("reverted %d rows", n)))
		return map[string]any{"reverted_rows": n, "honest_scope": rb.HonestScope()}, nil
	})
}

// backupDir is the on-disk location the backup Manager writes
// archives to. We put it next to the data dir so the user can
// find it. Tests can override by setting SYNAPTIC_BACKUP_DIR.
func backupDir(subs *Subsystems) string {
	if dir := subs.GeneralDataDir(); dir != "" {
		return dir + "/backups"
	}
	return ""
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
		full := dir + "/" + name
		size, _ := fileSize(full)
		out = append(out, backupEntry{Name: name, Path: full, Size: size})
	}
	return out, nil
}
