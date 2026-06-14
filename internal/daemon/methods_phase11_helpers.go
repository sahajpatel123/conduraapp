package daemon

import (
	"encoding/base64"
	"encoding/json"
	"os"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/audit"
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
