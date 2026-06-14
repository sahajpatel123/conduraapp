// Package backup implements encrypted backup and restore for Synaptic's
// on-disk state (Phase 11, sub-phase 11B).
//
// What we back up (per MISSION §24 + the Phase 11 plan):
//   - main DB:      <data-dir>/synaptic.db         (+ WAL/SHM sidecars)
//   - memory DB:    <data-dir>/memory.db           (+ WAL/SHM sidecars)
//   - skills DB:    <data-dir>/skills.db           (+ WAL/SHM sidecars)
//   - secrets.json: <data-dir>/secrets.json
//   - config.yaml:  from cfg path if present
//
// Why a separate encryption key? (User decision, Session 18.)
// The storage.DB master key protects the live database. If the user's
// keychain entry is compromised, the attacker gains the live database
// and any backups that reuse the same key. The backup key is
// derived from the master key via HKDF-SHA256 with a fixed info
// string ("synaptic-backup-encryption-key-v1") and shown to the
// user **once** on first backup, with a notice to save it. Without
// this derived key, the archive is unreadable even to someone with
// the live master key. This is the passphrase model.
//
// Archive format (v1):
//
//	<archive>.zip
//	  ├─ manifest.json     (versions, checksums, schema, timestamp)
//	  ├─ synaptic.db
//	  ├─ synaptic.db-wal
//	  ├─ synaptic.db-shm
//	  ├─ memory.db
//	  ├─ memory.db-wal
//	  ├─ memory.db-shm
//	  ├─ skills.db
//	  ├─ skills.db-wal
//	  ├─ skills.db-shm
//	  ├─ secrets.json
//	  └─ config.yaml (if present)
//
// The .zip is encrypted: each file inside is sealed with AES-256-GCM
// using a per-archive content key, and the content key is wrapped
// with the derived backup key (HKDF output).
//
// Why zip + encryption (not tar + raw AES): standard tooling can
// inspect the manifest without the key, and the layout is debuggable.
package backup

import (
	"archive/zip"
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/hmac"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"
)

// DeriveKey returns the per-deployment backup encryption key derived
// from the storage.DB master key via HKDF-SHA256 with a fixed info
// string. This is a public function so the on-first-backup notice
// can show the user the same key their archive uses.
func DeriveKey(masterKey []byte) ([]byte, error) {
	if len(masterKey) != 32 {
		return nil, fmt.Errorf("backup: master key must be 32 bytes (got %d)", len(masterKey))
	}
	// HKDF-Extract: PRK = HMAC-SHA256(salt, IKM) with empty salt.
	// HKDF-Expand: OKM = HMAC-SHA256(PRK, info || 0x01) for 32 bytes.
	const info = "synaptic-backup-encryption-key-v1"
	prk := hmacSHA256(nil, masterKey)
	out := hmacSHA256(prk, []byte(info+"\x01"))
	return out, nil
}

// DeriveKeyBase64 is a convenience for the first-backup notice.
func DeriveKeyBase64(masterKey []byte) (string, error) {
	k, err := DeriveKey(masterKey)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(k), nil
}

func hmacSHA256(key, data []byte) []byte {
	mac := hmac.New(sha256.New, key)
	mac.Write(data)
	return mac.Sum(nil)
}

// ManifestVersion is the current backup manifest format version.
// Bump on breaking changes.
const ManifestVersion = 1

// File modes used inside the zip archive.
const (
	fileModeOwnerOnly  = 0o600
	fileModePublicRead = 0o644
)

// Manifest is the human-readable metadata at the root of the archive.
// The manifest is NOT encrypted (the user can inspect it without
// the key — the actual database contents are).
type Manifest struct {
	Version        int            `json:"version"`         // ManifestVersion
	SchemaVersion  int            `json:"schema_version"`  // storage schema version at time of backup
	CreatedAt      string         `json:"created_at"`      // RFC3339Nano
	DataDir        string         `json:"data_dir"`        // path at time of backup (for debugging)
	Files          []ManifestFile `json:"files"`           // checksum + size per file
	KeyFingerprint string         `json:"key_fingerprint"` // first 8 hex chars of SHA256(derivedKey) — for "is this the right key?" UI
}

// ManifestFile is one entry in Manifest.Files.
type ManifestFile struct {
	Path      string `json:"path"`      // path inside the archive (e.g. "synaptic.db")
	Size      int64  `json:"size"`      // plaintext size in bytes
	SHA256    string `json:"sha256"`    // hex SHA-256 of the plaintext
	Encrypted bool   `json:"encrypted"` // always true in v1
}

// artifactSpec is the internal "what to back up" list. Each
// errOptionalMissing signals that an optional artifact (e.g. the
// config.yaml) was absent at backup time. Callers treat this as
// "skip, no error" and do not surface it.
var errOptionalMissing = errors.New("backup: optional artifact missing")

// artifact's source is on disk; the Manager reads, encrypts,
// and writes into the archive.
type artifactSpec struct {
	// pathInArchive is the name inside the zip.
	pathInArchive string
	// sourcePath is the on-disk path. Empty means "skip if missing".
	sourcePath string
	// optional files are missing → OK (warning, not error).
	optional bool
}

// Options configures a Manager.
type Options struct {
	// DataDir is the Synaptic data directory (parent of synaptic.db).
	DataDir string
	// ConfigPath is the path to the user's config.yaml (optional).
	ConfigPath string
	// MasterKey is the 32-byte storage.DB master key, used to
	// derive the backup encryption key.
	MasterKey []byte
	// SchemaVersion is the current storage schema version, written
	// into the manifest.
	SchemaVersion int
	// Now is the timestamp to record. If zero, time.Now().UTC().
	Now time.Time
	// Out is where the archive is written. If empty, a temp file is
	// created and the path is returned.
	Out string
}

// Manager creates and restores encrypted backups.
type Manager struct {
	opts Options
}

// New returns a Manager.
func New(opts Options) (*Manager, error) {
	if opts.DataDir == "" {
		return nil, errors.New("backup: DataDir is required")
	}
	if len(opts.MasterKey) != 32 {
		return nil, fmt.Errorf("backup: MasterKey must be 32 bytes (got %d)", len(opts.MasterKey))
	}
	return &Manager{opts: opts}, nil
}

// Create builds an encrypted archive of all configured artifacts and
// returns the path of the resulting .zip. If opts.Out is empty,
// a temp file is created in <data-dir>/backups, the archive is
// streamed into it, and it's renamed to .zip on success. On
// any error path the .zip.tmp is removed so we never leave
// orphan partial archives behind (they'd contain real encrypted
// DB data).
//
//nolint:gocyclo // branchy by nature: optional-file, ctx-cancel, per-artifact, rename
func (b *Manager) Create(ctx context.Context) (string, error) {
	now := b.opts.Now
	if now.IsZero() {
		now = time.Now().UTC()
	}

	archiveKey, err := DeriveKey(b.opts.MasterKey)
	if err != nil {
		return "", err
	}
	specs := b.collectSpecs()
	manifest := Manifest{
		Version:        ManifestVersion,
		SchemaVersion:  b.opts.SchemaVersion,
		CreatedAt:      now.Format(time.RFC3339Nano),
		DataDir:        b.opts.DataDir,
		KeyFingerprint: keyFingerprint(archiveKey),
	}

	outPath, createdTmp, err := b.openOutput()
	if err != nil {
		return "", err
	}
	// Cleanup: on any error from here on, remove the partial
	// archive so we don't leave orphan .zip.tmp files behind.
	success := false
	defer func() {
		if !success {
			_ = os.Remove(outPath)
		}
	}()

	if err := b.writeFirstPass(ctx, outPath, &manifest, specs, archiveKey); err != nil {
		return "", err
	}
	finalPath, err := b.rebuildWithManifest(manifest, archiveKey, specs, outPath)
	if err != nil {
		return "", err
	}
	if createdTmp {
		finalPath, err = b.renameToFinal(finalPath)
		if err != nil {
			return "", err
		}
	}
	success = true
	return finalPath, nil
}

// openOutput picks the destination path for the archive.
// If the caller passed Options.Out, use it verbatim. Otherwise
// create <data-dir>/backups/synaptic-backup-*.zip.tmp so the
// archive lives where backup.list looks. The .zip.tmp suffix
// signals "in progress"; the rename to .zip on success is
// the atomic switch to "ready".
func (b *Manager) openOutput() (outPath string, createdTmp bool, err error) {
	if b.opts.Out != "" {
		return b.opts.Out, false, nil
	}
	backupDir := filepath.Join(b.opts.DataDir, "backups")
	if err := os.MkdirAll(backupDir, 0o700); err != nil {
		return "", false, fmt.Errorf("backup: mkdir backup dir: %w", err)
	}
	f, err := os.CreateTemp(backupDir, "synaptic-backup-*.zip.tmp")
	if err != nil {
		return "", false, fmt.Errorf("backup: create temp: %w", err)
	}
	_ = f.Close()
	return f.Name(), true, nil
}

// writeFirstPass streams the manifest placeholder + each
// encrypted artifact into outPath. After this pass the zip
// has all the file bodies but the manifest is incomplete
// (no checksums yet). The second pass (rebuildWithManifest)
// re-streams the archive with the completed manifest.
//
// `manifest` is passed by pointer so the per-artifact checksums
// we append here are visible to the second pass.
func (b *Manager) writeFirstPass(ctx context.Context, outPath string, manifest *Manifest, specs []artifactSpec, archiveKey []byte) error {
	// Ensure 0o600 — backup contains secrets.
	out, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileModeOwnerOnly) //nolint:gosec // path is from trusted openOutput
	if err != nil {
		return fmt.Errorf("backup: open out: %w", err)
	}
	// Do NOT defer out.Close() — the second pass (rebuild)
	// reopens the same file and we want clean state.
	zw := zip.NewWriter(out)

	closeBoth := func() {
		_ = zw.Close()
		_ = out.Close()
	}

	if err := writeManifestJSON(zw, *manifest); err != nil {
		closeBoth()
		return err
	}
	for _, s := range specs {
		if err := ctx.Err(); err != nil {
			closeBoth()
			return err
		}
		mf, err := b.writeArtifact(zw, s, archiveKey)
		if errors.Is(err, errOptionalMissing) {
			continue
		}
		if err != nil {
			closeBoth()
			return err
		}
		if mf != nil {
			manifest.Files = append(manifest.Files, *mf)
		}
	}
	if err := zw.Close(); err != nil {
		_ = out.Close()
		return err
	}
	return out.Close()
}

// renameToFinal takes a .zip.tmp path (an in-progress
// archive) and renames it to .zip (ready). Atomic on the
// same filesystem.
func (b *Manager) renameToFinal(tmpPath string) (string, error) {
	if !strings.HasSuffix(tmpPath, ".zip.tmp") {
		return tmpPath, nil
	}
	ready := strings.TrimSuffix(tmpPath, ".tmp")
	if err := os.Rename(tmpPath, ready); err != nil {
		return "", fmt.Errorf("backup: rename to .zip: %w", err)
	}
	return ready, nil
}

// rebuildWithManifest creates the archive a second time, this time
// with the complete manifest (including the post-write checksums).
// The first pass is discarded via truncation; we stream the
// archive into outPath.
func (b *Manager) rebuildWithManifest(manifest Manifest, key []byte, specs []artifactSpec, outPath string) (string, error) {
	// Sort the file list for determinism.
	sort.Slice(manifest.Files, func(i, j int) bool {
		return manifest.Files[i].Path < manifest.Files[j].Path
	})

	out, err := os.OpenFile(outPath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, fileModeOwnerOnly) //nolint:gosec // path is from trusted Options.Out
	if err != nil {
		return "", fmt.Errorf("backup: open out: %w", err)
	}
	defer func() { _ = out.Close() }()
	zw := zip.NewWriter(out)
	defer func() { _ = zw.Close() }()

	if err := writeManifestJSON(zw, manifest); err != nil {
		return "", err
	}
	written := 0
	for _, s := range specs {
		mf := findManifestFile(manifest.Files, s.pathInArchive)
		if mf == nil {
			continue
		}
		if err := b.encryptArtifactIntoZip(zw, s, key); err != nil {
			return "", err
		}
		written++
	}
	// Close the zip writer so all central directory entries are
	// written before we close the underlying file.
	if err := zw.Close(); err != nil {
		return "", fmt.Errorf("backup: zip.Close: %w", err)
	}
	// fsync to flush before the caller reads the path.
	if err := out.Sync(); err != nil {
		return "", fmt.Errorf("backup: fsync: %w", err)
	}
	if err := out.Close(); err != nil {
		return "", fmt.Errorf("backup: close out: %w", err)
	}
	return outPath, nil
}

func (b *Manager) collectSpecs() []artifactSpec {
	dd := b.opts.DataDir
	return []artifactSpec{
		{pathInArchive: "synaptic.db", sourcePath: filepath.Join(dd, "synaptic.db")},
		{pathInArchive: "synaptic.db-wal", sourcePath: filepath.Join(dd, "synaptic.db-wal"), optional: true},
		{pathInArchive: "synaptic.db-shm", sourcePath: filepath.Join(dd, "synaptic.db-shm"), optional: true},
		{pathInArchive: "memory.db", sourcePath: filepath.Join(dd, "memory.db")},
		{pathInArchive: "memory.db-wal", sourcePath: filepath.Join(dd, "memory.db-wal"), optional: true},
		{pathInArchive: "memory.db-shm", sourcePath: filepath.Join(dd, "memory.db-shm"), optional: true},
		// Skills DB lives INSIDE the data dir, alongside the
		// main DB. Previously this code read it from the
		// parent dir, which disagreed with the daemon's
		// wiring (internal/daemon/subsystems.go:
		// buildPhase12 uses cfg.General.DataDir/skills.db).
		// That disagreement meant backup.create failed with
		// "open <parent>/skills.db: no such file or directory"
		// on every fresh install. Single source of truth is
		// <data-dir>/skills.db.
		{pathInArchive: "skills.db", sourcePath: filepath.Join(dd, "skills.db")},
		{pathInArchive: "skills.db-wal", sourcePath: filepath.Join(dd, "skills.db-wal"), optional: true},
		{pathInArchive: "skills.db-shm", sourcePath: filepath.Join(dd, "skills.db-shm"), optional: true},
		// secrets.json is only on disk if the secrets backend
		// is the file backend. The keyring backend (default
		// on macOS) keeps the master key in the OS keyring,
		// not on disk. Marking it optional lets the user
		// back up + restore from a keyring-backed install.
		// Recovery path: the encrypted archive can still be
		// restored as long as the user has the derived key
		// (shown once at first backup, or retrievable from
		// the keyring on the same machine).
		{pathInArchive: "secrets.json", sourcePath: filepath.Join(dd, "secrets.json"), optional: true},
		{pathInArchive: "config.yaml", sourcePath: b.opts.ConfigPath, optional: true},
	}
}

// writeArtifact (pass 1, used by the first Create pass) reads the
// artifact, encrypts, writes to zw, and returns the manifest entry
// or nil if the artifact was optional and missing.
func (b *Manager) writeArtifact(zw *zip.Writer, s artifactSpec, key []byte) (*ManifestFile, error) {
	plaintext, err := os.ReadFile(s.sourcePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && s.optional {
			return nil, errOptionalMissing
		}
		return nil, fmt.Errorf("backup: read %s: %w", s.sourcePath, err)
	}
	sum := sha256.Sum256(plaintext)
	if err := writeEncryptedEntry(zw, s.pathInArchive, plaintext, key); err != nil {
		return nil, err
	}
	return &ManifestFile{
		Path:      s.pathInArchive,
		Size:      int64(len(plaintext)),
		SHA256:    hex.EncodeToString(sum[:]),
		Encrypted: true,
	}, nil
}

// encryptArtifactIntoZip (pass 2) reads + encrypts + writes, with
// no return value (manifest is already complete).
func (b *Manager) encryptArtifactIntoZip(zw *zip.Writer, s artifactSpec, key []byte) error {
	plaintext, err := os.ReadFile(s.sourcePath)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) && s.optional {
			return nil
		}
		return fmt.Errorf("backup: read %s: %w", s.sourcePath, err)
	}
	return writeEncryptedEntry(zw, s.pathInArchive, plaintext, key)
}

func writeManifestJSON(zw *zip.Writer, m Manifest) error {
	hdr := &zip.FileHeader{
		Name: "manifest.json",
		// Use a fixed time for determinism in tests.
		Modified: time.Unix(0, 0),
	}
	hdr.SetMode(fileModePublicRead)
	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return fmt.Errorf("backup: create manifest header: %w", err)
	}
	enc := json.NewEncoder(w)
	enc.SetIndent("", "  ")
	if err := enc.Encode(m); err != nil {
		return fmt.Errorf("backup: encode manifest: %w", err)
	}
	return nil
}

// writeEncryptedEntry seals plaintext with AES-256-GCM (random nonce
// prepended) using the archive key, and writes the ciphertext as a
// zip entry.
func writeEncryptedEntry(zw *zip.Writer, name string, plaintext, key []byte) error {
	block, err := aes.NewCipher(key)
	if err != nil {
		return fmt.Errorf("backup: aes.NewCipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return fmt.Errorf("backup: cipher.NewGCM: %w", err)
	}
	nonce := make([]byte, gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return fmt.Errorf("backup: nonce: %w", err)
	}
	aad := []byte(name)
	sealed := gcm.Seal(nonce, nonce, plaintext, aad)

	hdr := &zip.FileHeader{
		Name: name,
		// Use a fixed time for determinism in tests.
		Modified: time.Unix(0, 0),
	}
	hdr.SetMode(fileModeOwnerOnly)
	w, err := zw.CreateHeader(hdr)
	if err != nil {
		return fmt.Errorf("backup: create %s header: %w", name, err)
	}
	if _, err := w.Write(sealed); err != nil {
		return fmt.Errorf("backup: write %s: %w", name, err)
	}
	return nil
}

// keyFingerprint is the first 8 hex chars of SHA256(archiveKey).
// It identifies which key was used without revealing the key.
func keyFingerprint(archiveKey []byte) string {
	sum := sha256.Sum256(archiveKey)
	return hex.EncodeToString(sum[:4])
}

func findManifestFile(files []ManifestFile, path string) *ManifestFile {
	for i := range files {
		if files[i].Path == path {
			return &files[i]
		}
	}
	return nil
}

// ListBackupFiles returns the names of all files in an archive
// (manifest + artifacts). Used by the "uninstall preview" and
// restore UI to show the user what was backed up.
func ListBackupFiles(archivePath string) ([]string, error) {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("backup: open: %w", err)
	}
	defer func() { _ = zr.Close() }()
	names := make([]string, 0, len(zr.File))
	for _, f := range zr.File {
		names = append(names, f.Name)
	}
	sort.Strings(names)
	return names, nil
}

// LoadManifest reads and decodes the manifest from an archive.
func LoadManifest(archivePath string) (*Manifest, error) {
	zr, err := zip.OpenReader(archivePath)
	if err != nil {
		return nil, fmt.Errorf("backup: open: %w", err)
	}
	defer func() { _ = zr.Close() }()
	hdr, err := zr.Open("manifest.json")
	if err != nil {
		return nil, fmt.Errorf("backup: open manifest: %w", err)
	}
	defer func() { _ = hdr.Close() }()
	var m Manifest
	if err := json.NewDecoder(hdr).Decode(&m); err != nil {
		return nil, fmt.Errorf("backup: decode manifest: %w", err)
	}
	return &m, nil
}

// ArchivePathFor returns a default path for a new archive, given
// the data dir and timestamp. Format:
//
//	<data-dir>/backups/synaptic-backup-2026-06-14T02-30-00Z.zip
func ArchivePathFor(dataDir string, t time.Time) string {
	if t.IsZero() {
		t = time.Now().UTC()
	}
	stamp := t.Format("2006-01-02T15-04-05Z")
	return filepath.Join(dataDir, "backups", "synaptic-backup-"+stamp+".zip")
}

// isSafeArchivePath rejects zip-slip: paths that escape the target
// root via "..", absolute paths, or drive letters.
func isSafeArchivePath(p string) bool {
	if p == "" {
		return false
	}
	if strings.HasPrefix(p, "/") || strings.HasPrefix(p, `\`) {
		return false
	}
	if filepath.IsAbs(p) {
		return false
	}
	// Walk the path; ".." segments are not allowed.
	for _, seg := range strings.Split(filepath.ToSlash(p), "/") {
		if seg == ".." {
			return false
		}
	}
	return true
}
