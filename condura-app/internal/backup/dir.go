package backup

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ResolveBackupDir returns the directory where encrypted backup
// archives are stored. Priority:
//  1. CONDURA_BACKUP_DIR environment variable (absolute path)
//  2. ~/Documents/condura-backups (MISSION §24.1 / decision #17)
//  3. <data-dir>/backups (daemon-local default)
func ResolveBackupDir(dataDir string) string {
	if dir := os.Getenv("CONDURA_BACKUP_DIR"); dir != "" {
		return dir
	}
	if home := userHomeDir(); home != "" {
		return filepath.Join(home, "Documents", "condura-backups")
	}
	return filepath.Join(dataDir, "backups")
}

func userHomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	if h := os.Getenv("USERPROFILE"); h != "" {
		return h
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}

// archiveExt is the on-disk extension for backup archives.
// Centralized here so the 'list archives' filter stays in sync
// with whatever ArchivePathFor produces.
const archiveExt = ".zip"

// ListBackupArchives returns absolute paths to backup archives
// in dir, sorted newest-first by modification time. Non-archive
// files (manifests, sidecars, log files accidentally dropped
// here) are filtered out.
//
// Newest-first matches the operator's mental model: "what's the
// most recent backup I can restore?" sorts to the top.
//
// Errors:
//   - dir does not exist OR is not a directory → ErrBackupDirNotFound
//   - permission denied on the dir             → wrapped OS error
//   - individual unreadable entries are SKIPPED (a single bad
//     file must not abort the whole list)
//
// This is the LOCAL-ONLY equivalent of the (currently daemon-only)
// list view; it does NOT require condurad to be running.
func ListBackupArchives(dir string) ([]string, error) {
	//nolint:gosec // G304: dir is the function parameter (operator-supplied via CONDURA_BACKUP_DIR / --data-dir); no untrusted-taint concern at this layer.
	f, err := os.Open(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, &backupDirError{kind: "not-found", path: dir}
		}
		return nil, &backupDirError{kind: "open", path: dir, err: err}
	}
	defer func() { _ = f.Close() }()

	// Stat the open fd to confirm it's a directory; using the
	// fd (rather than os.Stat(dir)) avoids a TOCTOU window
	// between existence-check and directory-read.
	fi, err := f.Stat()
	if err != nil {
		return nil, &backupDirError{kind: "stat", path: dir, err: err}
	}
	if !fi.IsDir() {
		return nil, &backupDirError{kind: "not-a-directory", path: dir}
	}

	type entry struct {
		path    string
		modTime int64 // unix nanos; larger = newer
	}
	var archives []entry

	for {
		names, err := f.Readdirnames(64)
		for _, name := range names {
			// Cheap extension filter first — avoids a Stat()
			// for every non-archive file (which is the
			// common case in the data dir where many
			// sidecar files live alongside the archives).
			if !strings.HasSuffix(name, archiveExt) {
				continue
			}
			full := filepath.Join(dir, name)
			stfi, err := os.Stat(full)
			if err != nil {
				// Skip unreadable entries; the list must
				// not abort on a single bad file.
				continue
			}
			if !stfi.Mode().IsRegular() {
				continue
			}
			archives = append(archives, entry{
				path:    full,
				modTime: stfi.ModTime().UnixNano(),
			})
		}
		if err == io.EOF {
			break
		}
		if err != nil {
			return nil, &backupDirError{kind: "readdir", path: dir, err: err}
		}
	}

	// Newest-first: sort by modTime descending. Stable for
	// ties (lexical by path) so the output is deterministic
	// across calls.
	sort.SliceStable(archives, func(i, j int) bool {
		if archives[i].modTime != archives[j].modTime {
			return archives[i].modTime > archives[j].modTime
		}
		return archives[i].path < archives[j].path
	})

	out := make([]string, len(archives))
	for i, a := range archives {
		out[i] = a.path
	}
	return out, nil
}

// backupDirError is the structured error type returned by
// ListBackupArchives. Callers can errors.As on it to distinguish
// 'directory not found' (an empty result is the right answer)
// from 'permission denied' (the caller should surface this to
// the operator).
type backupDirError struct {
	kind string // "not-found" | "open" | "stat" | "not-a-directory" | "readdir"
	path string
	err  error
}

func (e *backupDirError) Error() string {
	if e.err != nil {
		return "backup: " + e.kind + " " + e.path + ": " + e.err.Error()
	}
	return "backup: " + e.kind + " " + e.path
}

func (e *backupDirError) Unwrap() error { return e.err }

// IsBackupDirNotFound reports whether err is a
// ListBackupArchives "directory does not exist" error. Use
// this to distinguish "no backups yet" (return empty result
// to the operator) from real I/O errors (which should be
// surfaced).
func IsBackupDirNotFound(err error) bool {
	var bde *backupDirError
	return errors.As(err, &bde) && (bde.kind == "not-found" || bde.kind == "not-a-directory")
}
