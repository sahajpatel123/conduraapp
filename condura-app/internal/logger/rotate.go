package logger

import (
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"sync"
	"time"
)

// rotatingConfig holds size/age limits for a single log file sink.
type rotatingConfig struct {
	Filename   string
	MaxSize    int64 // bytes; must be > 0
	MaxBackups int   // number of rotated siblings to keep; <0 = unlimited
	MaxAgeDays int   // days to keep rotated files; <0 = unlimited
}

// rotatingWriter is a size-based rotating file writer.
//
// Layout (for Filename = /path/condura.log):
//
//	/path/condura.log      — active file
//	/path/condura.log.1    — most recent rotation
//	/path/condura.log.2    — older
//	…
//
// On Write, if the next write would exceed MaxSize, the active file is
// closed, siblings are shifted (.N → .N+1), the active file is renamed
// to .1, and a fresh active file is opened. Age-based pruning runs after
// each rotation.
//
// Thread-safe. Safe to use as an io.Writer under slog.
type rotatingWriter struct {
	mu   sync.Mutex
	cfg  rotatingConfig
	file *os.File
	size int64
}

func newRotatingWriter(cfg rotatingConfig) (*rotatingWriter, error) {
	if cfg.Filename == "" {
		return nil, fmt.Errorf("logger: empty log filename")
	}
	if cfg.MaxSize <= 0 {
		cfg.MaxSize = int64(DefaultMaxSizeMB) * 1024 * 1024
	}
	w := &rotatingWriter{cfg: cfg}
	if err := w.openExistingOrNew(); err != nil {
		return nil, err
	}
	return w, nil
}

func (w *rotatingWriter) Write(p []byte) (int, error) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.file == nil {
		if err := w.openExistingOrNew(); err != nil {
			return 0, err
		}
	}

	// Rotate before write if the next chunk would push us over the cap.
	// A single write larger than MaxSize is still accepted (we never
	// split a log line) but triggers rotation on the subsequent write.
	if w.size > 0 && w.size+int64(len(p)) > w.cfg.MaxSize {
		if err := w.rotateLocked(); err != nil {
			return 0, err
		}
	}

	n, err := w.file.Write(p)
	w.size += int64(n)
	return n, err
}

// Close closes the underlying file. Safe to call multiple times.
func (w *rotatingWriter) Close() error {
	w.mu.Lock()
	defer w.mu.Unlock()
	return w.closeFileLocked()
}

func (w *rotatingWriter) openExistingOrNew() error {
	if err := os.MkdirAll(filepath.Dir(w.cfg.Filename), logDirPerm); err != nil {
		return fmt.Errorf("logger: create log dir: %w", err)
	}
	f, err := os.OpenFile(w.cfg.Filename, os.O_CREATE|os.O_WRONLY|os.O_APPEND, logFilePerm) //nolint:gosec // G304: path is the user-configured log file
	if err != nil {
		return fmt.Errorf("logger: open log file: %w", err)
	}
	info, err := f.Stat()
	if err != nil {
		_ = f.Close()
		return fmt.Errorf("logger: stat log file: %w", err)
	}
	w.file = f
	w.size = info.Size()
	return nil
}

func (w *rotatingWriter) closeFileLocked() error {
	if w.file == nil {
		return nil
	}
	err := w.file.Close()
	w.file = nil
	w.size = 0
	return err
}

func (w *rotatingWriter) rotateLocked() error {
	if err := w.closeFileLocked(); err != nil {
		return err
	}

	// Shift existing backups upward: .N → .N+1 (highest first so we
	// don't clobber). MaxBackups < 0 means keep everything (age prune
	// still applies); MaxBackups == 0 is treated as "no backups" and
	// the active file is truncated in place.
	if w.cfg.MaxBackups == 0 {
		// Truncate by removing then reopening.
		_ = os.Remove(w.cfg.Filename)
		return w.openExistingOrNew()
	}

	backups, err := listBackups(w.cfg.Filename)
	if err != nil {
		return err
	}

	// Drop the oldest backup if we are at capacity before shifting.
	if w.cfg.MaxBackups > 0 && len(backups) >= w.cfg.MaxBackups {
		// Remove from the end (highest index = oldest).
		for i := len(backups) - 1; i >= w.cfg.MaxBackups-1; i-- {
			_ = os.Remove(backups[i].path)
		}
		// Re-list after pruning so the shift loop sees a clean set.
		backups, err = listBackups(w.cfg.Filename)
		if err != nil {
			return err
		}
	}

	// Shift .N → .N+1 starting from the highest index.
	for i := len(backups) - 1; i >= 0; i-- {
		b := backups[i]
		next := fmt.Sprintf("%s.%d", w.cfg.Filename, b.index+1)
		if err := os.Rename(b.path, next); err != nil && !os.IsNotExist(err) {
			return fmt.Errorf("logger: shift backup %s: %w", b.path, err)
		}
	}

	// Active → .1
	if _, err := os.Stat(w.cfg.Filename); err == nil {
		if err := os.Rename(w.cfg.Filename, w.cfg.Filename+".1"); err != nil {
			return fmt.Errorf("logger: rotate active log: %w", err)
		}
	}

	if err := w.pruneByAgeLocked(); err != nil {
		// Age prune failure is non-fatal for the write path; surface via
		// open error only if we cannot reopen the active file.
		_ = err
	}

	return w.openExistingOrNew()
}

func (w *rotatingWriter) pruneByAgeLocked() error {
	if w.cfg.MaxAgeDays < 0 {
		return nil
	}
	if w.cfg.MaxAgeDays == 0 {
		// Caller should have normalized 0 → DefaultMaxAgeDays; treat
		// residual 0 as "use default" for safety.
		w.cfg.MaxAgeDays = DefaultMaxAgeDays
	}
	cutoff := time.Now().AddDate(0, 0, -w.cfg.MaxAgeDays)
	backups, err := listBackups(w.cfg.Filename)
	if err != nil {
		return err
	}
	for _, b := range backups {
		info, err := os.Stat(b.path)
		if err != nil {
			continue
		}
		if info.ModTime().Before(cutoff) {
			_ = os.Remove(b.path)
		}
	}
	return nil
}

type backupFile struct {
	path  string
	index int
}

// listBackups returns rotated siblings of filename sorted by ascending
// index (filename.1, filename.2, …). Non-numeric suffixes are ignored.
func listBackups(filename string) ([]backupFile, error) {
	dir := filepath.Dir(filename)
	base := filepath.Base(filename)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	prefix := base + "."
	var out []backupFile
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, prefix) {
			continue
		}
		suffix := name[len(prefix):]
		// Only pure integer suffixes count as rotations.
		idx, ok := parsePositiveInt(suffix)
		if !ok {
			continue
		}
		out = append(out, backupFile{
			path:  filepath.Join(dir, name),
			index: idx,
		})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].index < out[j].index })
	return out, nil
}

// maxBackupIndex rejects absurd rotation suffixes (defense in depth).
const maxBackupIndex = 1_000_000

func parsePositiveInt(s string) (int, bool) {
	if s == "" {
		return 0, false
	}
	n := 0
	for _, c := range s {
		if c < '0' || c > '9' {
			return 0, false
		}
		n = n*10 + int(c-'0')
		if n > maxBackupIndex {
			return 0, false
		}
	}
	if n <= 0 {
		return 0, false
	}
	return n, true
}
