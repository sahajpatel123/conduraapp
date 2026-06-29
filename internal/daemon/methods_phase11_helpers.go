package daemon

import (
	"database/sql"
	"encoding/base64"
	"encoding/json"
	"io"
	"os"
	"path/filepath"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/audit"

	_ "modernc.org/sqlite"
)

// zeroTime returns a zero time.Time. The replay.Timeline
// treats zero as "use time.Now()". We centralize this so the
// helper is documented and easy to find.
func zeroTime() time.Time { return time.Time{} }

// base64Encode returns the standard base64 encoding of b. Used
// for embedding screenshot bytes in JSON RPC responses.
func base64Encode(b []byte) string {
	return base64.StdEncoding.EncodeToString(b)
}

// readDirNames returns the names of files in dir.
// Returns an empty slice (not an error) if dir does not exist;
// the caller can decide whether "no backups yet" is a real
// error or a normal state.
func readDirNames(dir string) ([]string, error) {
	f, err := os.Open(dir) //nolint:gosec // path is from trusted backup dir config
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()
	names, err := f.Readdirnames(0)
	if err != nil {
		return nil, err
	}
	return names, nil
}

// fileSize returns the size of path in bytes, or 0 if the file
// is missing. Best-effort: callers use this for display only.
func fileSize(path string) (int64, error) {
	fi, err := os.Stat(path) //nolint:gosec // path is from trusted backup dir config
	if err != nil {
		return 0, err
	}
	return fi.Size(), nil
}

// buildAuditEvent is a Phase-11-specific audit event builder
// that returns a fully-populated audit.Event the caller can
// tweak and Append. We avoid clashing with the existing
// auditEvent(ctx, subs, action, msg) helper in methods_more.go.
//
//nolint:unparam // app is passed through; future callers may vary it
func buildAuditEvent(action, app, result, message string) audit.Event {
	return audit.Event{
		Actor:   actorDaemon,
		Action:  action,
		App:     app,
		Level:   auditLevelInfo,
		Result:  result,
		Message: message,
		TS:      time.Now().UTC(),
	}
}

// jsonRaw returns params unchanged.
func jsonRaw(params json.RawMessage) json.RawMessage { return params } //nolint:unused

// OpenRollbackDB opens a SQLite database for rollback if the file exists.
// The caller (or Rollback.TrackOpened) is responsible for closing it.
func OpenRollbackDB(path string) *sql.DB {
	if path == "" {
		return nil
	}
	if _, err := os.Stat(path); err != nil {
		return nil
	}
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil
	}
	return db
}

// copyFile copies src to dst, creating parent directories as needed.
func copyFile(src, dst string) error {
	in, err := os.Open(src) //nolint:gosec // trusted backup path from daemon
	if err != nil {
		return err
	}
	defer func() { _ = in.Close() }()
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return err
	}
	out, err := os.Create(dst) //nolint:gosec // user-chosen export destination
	if err != nil {
		return err
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, in); err != nil {
		return err
	}
	return out.Close()
}
