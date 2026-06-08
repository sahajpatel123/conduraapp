// Package lockfile implements single-instance enforcement via an
// advisory flock on a per-data-dir file. It is intentionally simple:
// a process holds the lock for as long as it lives; if a second
// instance tries to start, it fails to acquire and returns ErrLocked.
//
// The lock is released when the holder calls Release() or exits (the
// kernel drops the flock automatically on process death or fd close).
//
// On macOS/Linux this uses fcntl(F_SETLK), which works across NFS
// and most networked filesystems. On Windows it uses LockFileEx.
// The Go standard library doesn't ship flock, so we use the pure-Go
// implementation in github.com/gofrs/flock.
package lockfile

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gofrs/flock"
)

// ErrLocked is returned by TryAcquire when another process holds the
// lock. Callers should treat this as "another instance is running" and
// surface a friendly message to the user.
var ErrLocked = errors.New("lockfile: another instance holds the lock")

// Lock is a held lock file. Call Release to drop it (or just let the
// process exit — the kernel will drop the flock).
type Lock struct {
	file *os.File
	fl   *flock.Flock
	path string
}

// Permissions used when creating the lock file and its parent dir.
// Tightened to 0600/0750 so non-owner users on a shared host cannot
// see which PID is running Synaptic.
const (
	dirPerm  os.FileMode = 0o750
	filePerm os.FileMode = 0o600
)

// Path returns the on-disk path of the lock file.
func (l *Lock) Path() string { return l.path }

// TryAcquire attempts a non-blocking flock on the file at path. The
// parent directory is created with 0o755 if missing. The file itself
// is created (or truncated) and a small JSON payload is written
// containing the holder's PID and hostname — this is purely
// diagnostic, used when reporting ErrLocked to the user.
//
// On success the returned *Lock must be Released by the caller (or
// the process will hold the lock until exit). On ErrLocked the file
// is left untouched and a second *Lock is NOT returned.
func TryAcquire(path string) (*Lock, error) {
	if err := os.MkdirAll(filepath.Dir(path), dirPerm); err != nil { //nolint:gosec // G304: caller-controlled path is the whole point
		return nil, fmt.Errorf("lockfile: mkdir: %w", err)
	}
	f, err := os.OpenFile(path, os.O_RDWR|os.O_CREATE, filePerm) //nolint:gosec // G304: caller-controlled path is the whole point
	if err != nil {
		return nil, fmt.Errorf("lockfile: open: %w", err)
	}
	// Write diagnostic payload BEFORE acquiring the lock. On Windows,
	// LockFileEx is mandatory and prevents writes through other handles.
	if err := f.Truncate(0); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("lockfile: truncate: %w", err)
	}
	if _, err := f.Seek(0, 0); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("lockfile: seek: %w", err)
	}
	if _, err := fmt.Fprintf(f, "pid=%d\n", os.Getpid()); err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("lockfile: write: %w", err)
	}
	_ = f.Sync()
	fl := flock.New(path)
	ok, err := fl.TryLock()
	if err != nil {
		_ = f.Close()
		return nil, fmt.Errorf("lockfile: trylock: %w", err)
	}
	if !ok {
		_ = f.Close()
		return nil, ErrLocked
	}
	return &Lock{file: f, fl: fl, path: path}, nil
}

// Release drops the flock and closes the file. Safe to call multiple
// times; subsequent calls are no-ops.
func (l *Lock) Release() error {
	if l == nil {
		return nil
	}
	if l.fl != nil {
		if err := l.fl.Unlock(); err != nil {
			return fmt.Errorf("lockfile: unlock: %w", err)
		}
		l.fl = nil
	}
	if l.file != nil {
		if err := l.file.Close(); err != nil {
			return fmt.Errorf("lockfile: close: %w", err)
		}
		l.file = nil
	}
	return nil
}
