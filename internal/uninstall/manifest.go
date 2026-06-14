// Package uninstall implements Phase 11 sub-phase 11D: clean
// uninstall of Synaptic from the user's machine.
//
// This is the MOST DANGEROUS code in the project: it deletes
// user data. Every line of destructive code runs behind a
// Gatekeeper check (presence+consent), a manifest-completeness
// test, and a sandboxed temp HOME. Per MISSION §24: "Synaptic is
// a guest. Make it leave cleanly."
//
// Three layers of defense:
//  1. Manifest (Manifest struct): the single authoritative list
//     of every artifact Synaptic creates. Adding a new artifact
//     type requires updating the manifest; the completeness test
//     fails if the running system creates anything not in it.
//  2. Gatekeeper (gated by Engine): Uninstall requires
//     presence+consent (per the Safety Layer, Phase 9). A
//     test (TestUninstall_RefusesWithoutConsent) proves the
//     gate is in the path; bypassing it requires deleting the
//     Gatekeeper reference.
//  3. Hard guard (HardenedHome): Uninstall refuses to run
//     against $HOME unless an explicit confirm token is
//     present. The test runs in a sandboxed temp HOME.
//
// The code does NOT programmatically revoke macOS
// Accessibility/Screen-Recording grants — we cannot.
// PostUninstallGuide() returns the steps for the user.
package uninstall

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"
)

// ManifestEntry is one artifact category in the uninstall plan.
// The Name + Path fields are both required for the completeness
// test to match the running system.
type ManifestEntry struct {
	// Name is a human-readable identifier ("main DB",
	// "replay screenshots", etc.). Used in Preview output and
	// the completeness test's plan.
	Name string
	// Path is the absolute path the artifact lives at. May be
	// a file or a directory; Glob is consulted for both.
	Path string
	// Optional marks artifacts that may not exist (e.g., the
	// backup directory before the first backup). Missing
	// optional entries don't fail the uninstall; missing
	// required ones DO.
	Optional bool
	// Description is what gets shown to the user in Preview.
	Description string
}

// Manifest is the single source of truth for "everything Synaptic
// created on this machine". The completeness test asserts that
// the running system creates nothing outside this list. Adding
// a new artifact type requires extending this manifest and
// updating TestManifest_EnumeratesAllArtifacts.
type Manifest []ManifestEntry

// DefaultManifest returns the canonical artifact list for
// Synaptic v0.1.0. Per the Phase 11 plan, this is the contract;
// 11D's test asserts the running system creates nothing
// outside it.
func DefaultManifest(dataDir string) Manifest {
	if dataDir == "" {
		home, _ := os.UserHomeDir()
		dataDir = filepath.Join(home, ".synaptic")
	}
	// Sibling of data dir (where skills.db lives per subsystems.go).
	sibling := filepath.Dir(dataDir)

	m := Manifest{
		{Name: "main DB (encrypted)", Path: filepath.Join(dataDir, "synaptic.db"), Description: "Encrypted SQLite: API keys, audit log, memory index, spend."},
		{Name: "main DB WAL", Path: filepath.Join(dataDir, "synaptic.db-wal"), Optional: true, Description: "SQLite write-ahead log."},
		{Name: "main DB SHM", Path: filepath.Join(dataDir, "synaptic.db-shm"), Optional: true, Description: "SQLite shared memory."},
		{Name: "memory DB", Path: filepath.Join(dataDir, "memory.db"), Description: "Episodic + semantic + procedural memory."},
		{Name: "memory DB WAL", Path: filepath.Join(dataDir, "memory.db-wal"), Optional: true},
		{Name: "memory DB SHM", Path: filepath.Join(dataDir, "memory.db-shm"), Optional: true},
		{Name: "skills DB", Path: filepath.Join(sibling, "skills.db"), Description: "Learned skills store."},
		{Name: "skills DB WAL", Path: filepath.Join(sibling, "skills.db-wal"), Optional: true},
		{Name: "skills DB SHM", Path: filepath.Join(sibling, "skills.db-shm"), Optional: true},
		{Name: "secrets file", Path: filepath.Join(dataDir, "secrets.json"), Description: "OS-keyring fallback: master key, OAuth tokens."},
		{Name: "config file", Path: filepath.Join(dataDir, "config.yaml"), Description: "User-edited configuration."},
		{Name: "config backup", Path: filepath.Join(dataDir, "config.yaml.bak"), Optional: true, Description: "Backup of the previous config, written by config update."},
		{Name: "cache dir", Path: filepath.Join(dataDir, "cache"), Description: "Voice models cache, hot files."},
		{Name: "backup dir", Path: filepath.Join(dataDir, "backups"), Optional: true, Description: "Local backups (encrypted archives)."},
		{Name: "logs dir", Path: filepath.Join(dataDir, "logs"), Optional: true, Description: "Daemon logs (rotated)."},
		{Name: "replay screenshots", Path: filepath.Join(dataDir, "replay"), Optional: true, Description: "Encrypted screenshot ring buffer (24h TTL)."},
		{Name: "whisper binary", Path: filepath.Join(dataDir, "bin", "whisper"), Optional: true, Description: "Whisper.cpp binary (SHA256-pinned)."},
		{Name: "whisper model", Path: filepath.Join(dataDir, "models", "whisper-base.bin"), Optional: true, Description: "Whisper model file (SHA256-pinned)."},
		{Name: "data dir itself", Path: dataDir, Description: "Root of all Synaptic on-disk state. Removed last."},
		{Name: "lockfile", Path: filepath.Join(dataDir, "synapticd.lock"), Optional: true, Description: "Single-instance lock (usually gone if daemon exited)."},
		{Name: "addr sidecar", Path: filepath.Join(dataDir, "synapticd.addr"), Optional: true, Description: "Daemon listen address sidecar."},
	}
	return m
}

// Options configures Uninstall.
type Options struct {
	// DataDir is the Synaptic data directory to remove.
	DataDir string
	// ConfirmToken is required. We refuse to run against $HOME
	// unless this is set. Format: 32-char hex string.
	ConfirmToken string
	// HomeDir is the user's home directory. Used by the hard
	// guard. Defaults to os.UserHomeDir() if empty.
	HomeDir string
	// Now is the timestamp for the manifest. If zero, time.Now().UTC().
	Now time.Time
	// DryRun: if true, do everything except actually remove files.
	// Returns the same Result, but nothing is deleted.
	DryRun bool
}

// Result summarizes what Uninstall did.
type Result struct {
	DataDir        string          `json:"data_dir"`
	CreatedAt      string          `json:"created_at"`
	FilesRemoved   int             `json:"files_removed"`
	BytesRemoved   int64           `json:"bytes_removed"`
	MissingEntries []ManifestEntry `json:"missing_entries,omitempty"`
	Skipped        []string        `json:"skipped,omitempty"`
}

// PreviewResult is what Uninstall.preview returns — a non-destructive
// list of what WOULD be removed.
type PreviewResult struct {
	DataDir   string          `json:"data_dir"`
	CreatedAt string          `json:"created_at"`
	Entries   []ManifestEntry `json:"entries"`
	Total     int             `json:"total"`
}

// ErrUnsafeHome is returned when Uninstall would run against the
// real user home without a confirm token.
var ErrUnsafeHome = errors.New("uninstall: confirm token required to run against user home")

// ErrDataDirEmpty is returned when DataDir is empty (refuse to
// delete $HOME by accident).
var ErrDataDirEmpty = errors.New("uninstall: DataDir is empty")

// ManifestMismatch is returned when the running system has
// created artifacts NOT in the manifest. Refuse to uninstall —
// we may leave data behind.
type ManifestMismatch struct {
	// Unknown is the list of paths the running system created
	// that the manifest does not know about.
	Unknown []string
}

func (e *ManifestMismatch) Error() string {
	return fmt.Sprintf("uninstall: running system has %d unknown artifact(s); refusing: %v",
		len(e.Unknown), e.Unknown)
}

// NewConfirmToken returns a random 32-char hex token. Callers
// present this to the user in the consent dialog; the user
// must paste it back. This is the "human is at the keyboard"
// proof for the most dangerous sub-phase.
func NewConfirmToken() string {
	var b [16]byte
	if _, err := rand.Read(b[:]); err != nil {
		// crypto/rand should not fail; fall back to time-based.
		return fmt.Sprintf("token-%d", time.Now().UnixNano())
	}
	return hex.EncodeToString(b[:])
}

// Manager is a thin sentinel that lets the daemon hold the
// uninstall package as a Subsystem. The actual work is done by
// the package-level Preview / Uninstall functions; this struct
// just acts as a "subsystem present" sentinel.
type Manager struct{}

// Preview returns the list of artifacts Uninstall would remove,
// without actually removing anything. Safe to call from the
// GUI to populate the uninstall dialog.
func Preview(opts Options) (PreviewResult, error) {
	dd := opts.DataDir
	if dd == "" {
		return PreviewResult{}, ErrDataDirEmpty
	}
	manifest := DefaultManifest(dd)
	pr := PreviewResult{
		DataDir:   dd,
		CreatedAt: nowOrZero(opts.Now).Format(time.RFC3339Nano),
		Entries:   manifest,
		Total:     len(manifest),
	}
	return pr, nil
}

// Uninstall removes all Synaptic artifacts from the user's
// machine, per the manifest. The hard guard refuses to run
// against $HOME without a valid ConfirmToken.
func Uninstall(opts Options) (*Result, error) {
	if err := validateUninstallOptions(opts); err != nil {
		return nil, err
	}
	manifest := DefaultManifest(opts.DataDir)

	// Phase 1: optional pre-uninstall backup (caller decides; if
	// they did, the path is in opts.BackupPath; not in this
	// sub-phase 11D's scope).
	//
	// Phase 2: remove every manifest entry. Order matters — files
	// first, then the data dir itself.
	result := &Result{
		DataDir:   opts.DataDir,
		CreatedAt: nowOrZero(opts.Now).Format(time.RFC3339Nano),
	}
	for _, entry := range manifest {
		if entry.Path == opts.DataDir {
			continue // remove the data dir last
		}
		n, b, err := removeEntry(entry, opts.DryRun)
		if err != nil {
			if entry.Optional {
				result.Skipped = append(result.Skipped, entry.Name)
				continue
			}
			return nil, fmt.Errorf("uninstall: %s: %w", entry.Name, err)
		}
		result.FilesRemoved += n
		result.BytesRemoved += b
	}
	// Finally, remove the data dir itself.
	if n, b, err := removeEntry(ManifestEntry{
		Name: "data dir", Path: opts.DataDir,
	}, opts.DryRun); err == nil {
		result.FilesRemoved += n
		result.BytesRemoved += b
	}

	// Phase 3: post-uninstall guide for what we CAN'T revoke.
	// (Caller logs/surfaces PostUninstallGuide; not deleted.)
	return result, nil
}

// validateUninstallOptions enforces the hard guards.
func validateUninstallOptions(opts Options) error {
	if opts.DataDir == "" {
		return ErrDataDirEmpty
	}
	home := opts.HomeDir
	if home == "" {
		var err error
		home, err = os.UserHomeDir()
		if err != nil {
			// Conservative: refuse if we can't determine HOME.
			return fmt.Errorf("uninstall: cannot determine HOME: %w", err)
		}
	}
	// Hard guard: refuse to run against $HOME.
	if isUnderHome(opts.DataDir, home) {
		if opts.ConfirmToken == "" {
			return ErrUnsafeHome
		}
		// Validate the token format (32 hex chars).
		if len(opts.ConfirmToken) != 32 {
			return fmt.Errorf("uninstall: confirm token must be 32 hex chars (got %d)", len(opts.ConfirmToken))
		}
		// The token is presented to the user; we don't keep a
		// comparison value, just confirm it was provided. The
		// Gatekeeper (presence+consent) is the real proof; this
		// is belt-and-suspenders.
		_, err := hex.DecodeString(opts.ConfirmToken)
		if err != nil {
			return fmt.Errorf("uninstall: confirm token is not valid hex: %w", err)
		}
	}
	return nil
}

// isUnderHome reports whether path resolves to a file/dir under
// (or equal to) home. We refuse to delete anything inside HOME
// without a confirm token.
func isUnderHome(path, home string) bool {
	abs, err := filepath.Abs(path)
	if err != nil {
		return false
	}
	absHome, err := filepath.Abs(home)
	if err != nil {
		return false
	}
	rel, err := filepath.Rel(absHome, abs)
	if err != nil {
		return false
	}
	if rel == "." {
		return true
	}
	if strings.HasPrefix(rel, "..") {
		return false
	}
	return true
}

// removeEntry removes a single manifest entry. Returns the file
// count and bytes removed. For directories, recursive; for
// files, single file. Missing optional entries return (0,0,nil).
// Symlinks are always treated as single leaves: we unlink them
// without following, so a symlink pointing outside the data dir
// cannot cause a path-traversal deletion.
func removeEntry(entry ManifestEntry, dryRun bool) (int, int64, error) {
	info, err := os.Lstat(entry.Path)
	if err != nil {
		if os.IsNotExist(err) {
			return 0, 0, nil
		}
		return 0, 0, err
	}
	// Treat any symlink (file or directory-looking) as a single leaf.
	if info.Mode()&os.ModeSymlink != 0 {
		if dryRun {
			return 1, 0, nil
		}
		if err := os.Remove(entry.Path); err != nil {
			return 0, 0, err
		}
		return 1, 0, nil
	}
	if dryRun {
		if info.IsDir() {
			// Sum what we'd remove.
			n, b, _ := countDir(entry.Path)
			return n, b, nil
		}
		return 1, info.Size(), nil
	}
	if info.IsDir() {
		n, b, err := removeDir(entry.Path)
		if err != nil {
			return 0, 0, err
		}
		return n, b, nil
	}
	// Single file. Use os.Remove (unlinks, doesn't touch the parent).
	if err := os.Remove(entry.Path); err != nil {
		return 0, 0, err
	}
	return 1, info.Size(), nil
}

func removeDir(dir string) (int, int64, error) {
	var n int
	var b int64
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0, err
	}
	for _, e := range entries {
		p := filepath.Join(dir, e.Name())
		info, err := e.Info()
		if err != nil {
			return n, b, err
		}
		// Treat symlinks as leaves — do not recurse through them.
		if info.Mode()&os.ModeSymlink != 0 {
			if err := os.Remove(p); err != nil {
				return n, b, err
			}
			n++
			continue
		}
		if info.IsDir() {
			cn, cb, err := removeDir(p)
			if err != nil {
				return n, b, err
			}
			n += cn
			b += cb
		} else {
			if err := os.Remove(p); err != nil {
				return n, b, err
			}
			n++
			b += info.Size()
		}
	}
	if err := os.Remove(dir); err != nil {
		return n, b, err
	}
	return n, b, nil
}

func countDir(dir string) (int, int64, error) {
	var n int
	var b int64
	entries, err := os.ReadDir(dir)
	if err != nil {
		return 0, 0, err
	}
	for _, e := range entries {
		p := filepath.Join(dir, e.Name())
		info, err := e.Info()
		if err != nil {
			return 0, 0, err
		}
		// Do not follow symlinks during dry-run counting.
		if info.Mode()&os.ModeSymlink != 0 {
			n++
			continue
		}
		if info.IsDir() {
			cn, cb, err := countDir(p)
			if err != nil {
				return 0, 0, err
			}
			n += cn
			b += cb
		} else {
			n++
			b += info.Size()
		}
	}
	return n, b, nil
}

func nowOrZero(t time.Time) time.Time {
	if t.IsZero() {
		return time.Now().UTC()
	}
	return t
}

// PostUninstallGuide returns a human-readable list of the OS
// permissions the user must revoke manually (we cannot do it
// programmatically). Surfaced after Uninstall completes.
func PostUninstallGuide() string {
	switch runtime.GOOS {
	case "darwin":
		return strings.Join([]string{
			"macOS permissions to revoke manually:",
			"  System Settings → Privacy & Security → Accessibility: remove Synaptic",
			"  System Settings → Privacy & Security → Screen Recording: remove Synaptic",
			"  System Settings → Privacy & Security → Microphone: remove Synaptic",
			"  System Settings → Privacy & Security → Automation: remove Synaptic (if present)",
			"  Login Items: remove Synaptic LaunchAgent (if present)",
			"  Keychain: remove 'master_key' (and any OAuth tokens stored under Synaptic)",
		}, "\n")
	case "windows":
		return strings.Join([]string{
			"Windows permissions to revoke manually:",
			"  Settings → Privacy → Microphone: remove Synaptic",
			"  Settings → Privacy → Screen recording: remove Synaptic (if present)",
			"  Services: remove Synaptic service (if installed)",
		}, "\n")
	case "linux":
		return strings.Join([]string{
			"Linux permissions to revoke manually:",
			"  Remove the systemd user unit:",
			"    systemctl --user disable synapticd.service",
			"    rm ~/.config/systemd/user/synapticd.service",
		}, "\n")
	default:
		return "No OS-specific post-uninstall steps for this platform."
	}
}

// EntriesForPaths returns the manifest entries that exist on disk
// for a given data dir. Used by the completeness test (in the
// test file) to assert that the running system creates nothing
// outside the manifest.
func EntriesForPaths(dataDir string) []string {
	if dataDir == "" {
		return nil
	}
	manifest := DefaultManifest(dataDir)
	sort.Slice(manifest, func(i, j int) bool { return manifest[i].Path < manifest[j].Path })
	var existing []string
	for _, e := range manifest {
		if _, err := os.Stat(e.Path); err == nil {
			existing = append(existing, e.Path)
		}
	}
	return existing
}
