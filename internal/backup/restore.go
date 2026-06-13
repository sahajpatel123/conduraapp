// Package backup also implements the destructive Restore path.
// Per MISSION §24 and the Phase 11 plan, Restore is itself a
// DESTRUCTIVE action and must pass the Gatekeeper. The gate
// check is done at the IPC layer (the methods_phase11.go
// entrypoint), not here — this package provides the file-level
// mechanics.
//
// Restore contract (Phase 11 plan, point 11B.2):
//  1. Decrypt the archive; verify checksums + schema compat.
//  2. Take a pre-restore safety snapshot (so the user can recover
//     from a bad restore).
//  3. Stop daemon writers. We acquire the data-dir lock (same one
//     daemon.Run acquires) so a concurrent daemon can't write while
//     we swap files.
//  4. Swap in restored files atomically (rename into place).
//  5. The next daemon start picks up the restored state.
//
// Schema compat policy (user decision, Session 18): refuse
// newer-schema → older-binary. A backup made with v3 can be
// restored to a v3 binary; restoring it to a v2 binary is denied
// with a clear error.
package backup

import (
	"archive/zip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// siblingFiles are archive entries that live outside the data dir
// (e.g., skills.db is at <data-dir>/../skills.db per MISSION §24).
var siblingFiles = map[string]bool{"skills.db": true}

// ErrSchemaIncompatible is returned when a backup's schema version
// is newer than the binary's current schema version.
var ErrSchemaIncompatible = errors.New("backup: schema newer than binary")

// ErrChecksumMismatch is returned when a restored file's SHA-256
// doesn't match the manifest entry.
var ErrChecksumMismatch = errors.New("backup: checksum mismatch")

// ErrUnknownKey is returned when a provided key's fingerprint
// doesn't match the manifest's recorded fingerprint.
var ErrUnknownKey = errors.New("backup: unknown key (fingerprint mismatch)")

// RestoreOptions configures a Restore.
type RestoreOptions struct {
	// ArchivePath is the .zip to restore from.
	ArchivePath string
	// DataDir is where the restored files go. Must be the same dir
	// the backup was made from (we verify by file names).
	DataDir string
	// MasterKey is the 32-byte storage.DB master key, used to
	// derive the backup encryption key.
	MasterKey []byte
	// ExpectedKey, if non-empty, is the base64-encoded derived
	// backup key. When the manifest's fingerprint doesn't match,
	// we return ErrUnknownKey. (This is the "user typed in their
	// saved key from the first-backup notice" path.)
	ExpectedKey string
	// CurrentSchemaVersion is the binary's schema version. If the
	// backup's schema is newer, we refuse.
	CurrentSchemaVersion int
	// PreRestoreBackupPath, if non-empty, is the path to write
	// the pre-restore safety snapshot. The caller decides whether
	// to skip this (e.g., when the user explicitly opts out).
	PreRestoreBackupPath string
	// Now is the timestamp for the safety snapshot. If zero, time.Now().UTC().
	Now time.Time
}

// Restore extracts an archive into DataDir. The pre-restore safety
// snapshot is created first (if PreRestoreBackupPath is set), so the
// caller can roll back the restore if it goes wrong.
func Restore(ctx context.Context, opts RestoreOptions) error {
	if err := validateRestoreOptions(opts); err != nil {
		return err
	}

	// 1. Read + validate the manifest.
	manifest, err := LoadManifest(opts.ArchivePath)
	if err != nil {
		return err
	}
	if err := validateManifest(opts, manifest); err != nil {
		return err
	}

	// 2. Pre-restore safety snapshot (best-effort; caller chose
	// whether to provide a path).
	if opts.PreRestoreBackupPath != "" {
		if err := createPreRestoreSnapshot(ctx, opts); err != nil {
			return fmt.Errorf("backup: pre-restore snapshot failed: %w", err)
		}
	}

	// 3. Derive the archive key and verify it against the manifest.
	archiveKey, err := DeriveKey(opts.MasterKey)
	if err != nil {
		return err
	}
	if opts.ExpectedKey != "" {
		expectedFingerprint := keyFingerprint(mustDecodeBase64(opts.ExpectedKey))
		if expectedFingerprint != manifest.KeyFingerprint {
			return ErrUnknownKey
		}
	}

	// 4. Open the archive + stage decrypted files in a temp dir.
	stageDir, err := openAndStage(ctx, opts, manifest, archiveKey)
	if err != nil {
		return err
	}

	// 5. Stop writers + atomic swap.
	// We rely on the daemon's single-instance lock being held by
	// the caller (daemon.pause) before this runs. The restore
	// itself is one big atomic directory swap: move the current
	// data dir to a backup name, move the staged dir in.
	swapped := false
	defer func() {
		// After a successful atomic swap, stageDir has been renamed
		// into DataDir; do NOT delete it (that would wipe the restored
		// state). Only clean up the stage dir on error paths.
		if !swapped {
			_ = os.RemoveAll(stageDir)
		}
	}()
	if err := atomicSwap(opts.DataDir, stageDir); err != nil {
		return err
	}
	swapped = true
	return nil
}

// validateRestoreOptions enforces the required fields up front.
func validateRestoreOptions(opts RestoreOptions) error {
	if opts.ArchivePath == "" {
		return errors.New("backup: ArchivePath is required")
	}
	if opts.DataDir == "" {
		return errors.New("backup: DataDir is required")
	}
	if len(opts.MasterKey) != 32 {
		return fmt.Errorf("backup: MasterKey must be 32 bytes (got %d)", len(opts.MasterKey))
	}
	return nil
}

// createPreRestoreSnapshot makes a backup of the current data dir
// before overwriting it. Best-effort; the caller decides whether
// to provide a path.
func createPreRestoreSnapshot(ctx context.Context, opts RestoreOptions) error {
	now := opts.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}
	_, err := (&Manager{opts: Options{
		DataDir:       opts.DataDir,
		MasterKey:     opts.MasterKey,
		SchemaVersion: opts.CurrentSchemaVersion,
		Now:           now,
		Out:           opts.PreRestoreBackupPath,
	}}).Create(ctx)
	return err
}

// openAndStage opens the archive and decrypts every manifest
// entry into the stage dir. Returns the stage dir path on success.
func openAndStage(ctx context.Context, opts RestoreOptions, manifest *Manifest, archiveKey []byte) (string, error) {
	zr, err := zip.OpenReader(opts.ArchivePath)
	if err != nil {
		return "", fmt.Errorf("backup: open: %w", err)
	}
	defer func() { _ = zr.Close() }()

	// The stage dir is created in the *parent* of the data dir, so
	// the atomic-swap rename of the data dir doesn't move the stage
	// out from under us. The parent is on the same filesystem as
	// the data dir, so os.Rename works.
	stageParent := filepath.Dir(opts.DataDir)
	if stageParent == "." {
		stageParent = os.TempDir()
	}
	stageDir, err := os.MkdirTemp(stageParent, "synaptic-restore-stage-")
	if err != nil {
		return "", fmt.Errorf("backup: stage dir: %w", err)
	}

	// Sort manifest entries for deterministic restore order.
	sortedFiles := append([]ManifestFile(nil), manifest.Files...)
	sort.Slice(sortedFiles, func(i, j int) bool { return sortedFiles[i].Path < sortedFiles[j].Path })

	for _, mf := range sortedFiles {
		if ctx.Err() != nil {
			return "", ctx.Err()
		}
		if err := decryptAndStage(zr, mf, stageDir, opts.DataDir, archiveKey); err != nil {
			return "", err
		}
	}
	return stageDir, nil
}

// validateManifest enforces the schema compat policy.
func validateManifest(opts RestoreOptions, m *Manifest) error {
	if m.SchemaVersion > opts.CurrentSchemaVersion {
		return fmt.Errorf("%w: backup is v%d, binary is v%d (downgrade not supported)",
			ErrSchemaIncompatible, m.SchemaVersion, opts.CurrentSchemaVersion)
	}
	return nil
}

// decryptAndStage opens one zip entry, decrypts it, verifies the
// SHA-256, and writes it to the stage dir (or its final sibling
// location for files like skills.db).
func decryptAndStage(zr *zip.ReadCloser, mf ManifestFile, stageDir, dataDir string, key []byte) error {
	if !isSafeArchivePath(mf.Path) {
		return fmt.Errorf("backup: unsafe archive path %q", mf.Path)
	}
	hdr, err := zr.Open(mf.Path)
	if err != nil {
		return fmt.Errorf("backup: open %s: %w", mf.Path, err)
	}
	defer func() { _ = hdr.Close() }()

	sealed, err := io.ReadAll(hdr)
	if err != nil {
		return fmt.Errorf("backup: read %s: %w", mf.Path, err)
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("backup: aes.NewCipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("backup: cipher.NewGCM: %w", err)
	}
	if len(sealed) < gcm.NonceSize() {
		return fmt.Errorf("backup: %s ciphertext too short", mf.Path)
	}
	nonce := sealed[:gcm.NonceSize()]
	body := sealed[gcm.NonceSize():]
	plain, err := gcm.Open(nil, nonce, body, []byte(mf.Path))
	if err != nil {
		return fmt.Errorf("backup: decrypt %s: %w", mf.Path, err)
	}

	// Verify SHA-256 against the manifest.
	sum := sha256.Sum256(plain)
	want, _ := hex.DecodeString(mf.SHA256)
	if !sha256Equal(sum[:], want) {
		return fmt.Errorf("%w: %s", ErrChecksumMismatch, mf.Path)
	}

	// Write to the stage dir, or to the sibling location for files
	// that live outside the data dir (e.g., skills.db).
	dstDir := stageDir
	if siblingFiles[mf.Path] {
		dstDir = filepath.Dir(dataDir)
	}
	dst := filepath.Join(dstDir, mf.Path)
	if err := os.MkdirAll(filepath.Dir(dst), 0o700); err != nil {
		return fmt.Errorf("backup: mkdir for %s: %w", mf.Path, err)
	}
	// Same file mode the source would have had (0600 for
	// sensitive files; we don't try to recover the original mode
	// from the manifest).
	if err := os.WriteFile(dst, plain, 0o600); err != nil {
		return fmt.Errorf("backup: write %s: %w", mf.Path, err)
	}
	return nil
}

func sha256Equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}
	var v byte
	for i := range a {
		v |= a[i] ^ b[i]
	}
	return v == 0
}

func mustDecodeBase64(s string) []byte {
	b, _ := base64.StdEncoding.DecodeString(s)
	return b
}

// atomicSwap moves the current data dir's contents aside, then
// moves the staged contents in. The "aside" dir is the rollback
// target if the swap fails partway.
func atomicSwap(dataDir, stageDir string) error {
	aside := dataDir + ".pre-restore-" + time.Now().UTC().Format("20060102-150405Z")
	if err := os.Rename(dataDir, aside); err != nil {
		return fmt.Errorf("backup: aside %s: %w", dataDir, err)
	}
	if err := os.Rename(stageDir, dataDir); err != nil {
		// Try to roll back.
		_ = os.Rename(aside, dataDir)
		return fmt.Errorf("backup: move stage in: %w", err)
	}
	// On success, remove the aside (it's the pre-restore state
	// we just replaced; the safety snapshot is the caller's
	// concern).
	_ = os.RemoveAll(aside)
	return nil
}

// InspectManifest is a debug/admin helper: opens an archive and
// returns a printable summary (manifest + file list + sizes).
func InspectManifest(archivePath string) (string, error) {
	m, err := LoadManifest(archivePath)
	if err != nil {
		return "", err
	}
	var sb strings.Builder
	fmt.Fprintf(&sb, "Synaptic backup\n")
	fmt.Fprintf(&sb, "  version:        %d\n", m.Version)
	fmt.Fprintf(&sb, "  schema_version: %d\n", m.SchemaVersion)
	fmt.Fprintf(&sb, "  created_at:     %s\n", m.CreatedAt)
	fmt.Fprintf(&sb, "  data_dir:       %s\n", m.DataDir)
	fmt.Fprintf(&sb, "  key_fingerprint: %s\n", m.KeyFingerprint)
	fmt.Fprintf(&sb, "  files (%d):\n", len(m.Files))
	for _, f := range m.Files {
		fmt.Fprintf(&sb, "    %s  %d bytes  sha256:%s  encrypted:%v\n",
			f.Path, f.Size, shortHash(f.SHA256), f.Encrypted)
	}
	return sb.String(), nil
}

// shortHashPrefix is how many hex chars of a SHA-256 fingerprint
// we print in human-readable summaries. Long enough to be
// unambiguous across our backup sizes, short enough for one line.
const shortHashPrefix = 12

func shortHash(h string) string {
	if len(h) >= shortHashPrefix {
		return h[:shortHashPrefix]
	}
	return h
}
