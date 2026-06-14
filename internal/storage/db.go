// Package storage provides SQLite-backed persistence for Synaptic.
//
// We use modernc.org/sqlite (pure Go, no CGO) for portability.
//
// Security model:
//   - The SQLite database file itself is unencrypted on disk.
//   - Sensitive columns (API key ciphertext, OAuth tokens, anything
//     containing user PII) are encrypted at the application layer using
//     AES-256-GCM before being written, and decrypted on read.
//   - The 32-byte master key is stored in the OS keyring (internal/secrets)
//     under "master_key" as base64. On first run, the daemon generates a
//     key and stores it. On subsequent runs, it loads the key from the
//     keyring.
//   - Each encrypted value has a per-value random nonce, prefixed to the
//     ciphertext. The AAD (additional authenticated data) is the row ID + the
//     column name, so a value moved between rows fails authentication.
//
// Schema migrations: this package embeds SQL files and applies them in order
// at Open() time. See migrations.go.
package storage

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"database/sql"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"

	// Pure-Go sqlite driver.
	_ "modernc.org/sqlite"

	"github.com/sahajpatel123/synapticapp/internal/secrets"
)

// DB is the storage handle.
type DB struct {
	sql       *sql.DB
	gcm       cipher.AEAD
	masterKey []byte     // raw 32-byte master key, exposed via MasterKey()
	mu        sync.Mutex // guards nonce generation & schema_version writes
	path      string
	closing   chan struct{}
	openedAt  time.Time
}

// Config is the storage configuration.
type Config struct {
	// Path is the on-disk SQLite file path.
	Path string
	// MasterKey is the base64-encoded 32-byte master key.
	// If empty, storage will load it from secrets.Manager under
	// secrets.MasterKey, or generate+store a new one on first run.
	MasterKey string
	// Secrets is used to load/store the master key when MasterKey is empty.
	// If nil and MasterKey is empty, the master key is generated ephemerally
	// (useful for tests; not safe for production).
	Secrets secrets.Manager
	// OnMigrate is called after each migration is applied. Optional.
	OnMigrate func(version int) error
}

// File mode for the SQLite database file. Owner-only because the
// database contains API keys (encrypted), audit logs, and memory.
const dbDirPerm = 0o700

// Open opens (or creates) the SQLite database, applies migrations, and
// returns a handle. The caller must call Close.
func Open(ctx context.Context, cfg Config) (*DB, error) {
	if cfg.Path == "" {
		return nil, errors.New("storage: Path is required")
	}
	if err := os.MkdirAll(filepath.Dir(cfg.Path), dbDirPerm); err != nil {
		return nil, fmt.Errorf("storage: create data dir: %w", err)
	}

	// Load or generate the master key.
	mk, err := loadOrCreateMasterKey(cfg)
	if err != nil {
		return nil, fmt.Errorf("storage: load master key: %w", err)
	}
	key, err := base64.StdEncoding.DecodeString(mk)
	if err != nil {
		return nil, fmt.Errorf("storage: master key base64: %w", err)
	}
	if len(key) != 32 {
		return nil, fmt.Errorf("storage: master key must be 32 bytes (got %d)", len(key))
	}

	block, err := aes.NewCipher(key)
	if err != nil {
		return nil, fmt.Errorf("storage: aes.NewCipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("storage: cipher.NewGCM: %w", err)
	}

	// Open SQLite. WAL + foreign keys; reasonable busy timeout.
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)", cfg.Path)
	sdb, err := sql.Open("sqlite", dsn)
	if err != nil {
		return nil, fmt.Errorf("storage: open sqlite: %w", err)
	}
	// SQLite is single-writer; we serialize through this connection to keep
	// the WAL happy. Long-running reads should use a separate connection.
	sdb.SetMaxOpenConns(1)

	if err := sdb.PingContext(ctx); err != nil {
		_ = sdb.Close()
		return nil, fmt.Errorf("storage: ping: %w", err)
	}

	db := &DB{
		sql:       sdb,
		gcm:       gcm,
		masterKey: key,
		path:      cfg.Path,
		closing:   make(chan struct{}),
		openedAt:  time.Now(),
	}
	if err := db.migrate(ctx, cfg.OnMigrate); err != nil {
		_ = sdb.Close()
		return nil, fmt.Errorf("storage: migrate: %w", err)
	}
	return db, nil
}

// Close releases the database. Safe to call multiple times
// and after Reload.
func (d *DB) Close() error {
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.sql == nil {
		return nil
	}
	close(d.closing)
	err := d.sql.Close()
	d.sql = nil
	return err
}

// Reload closes the current SQLite handle and reopens a fresh
// one against the same on-disk file. Use this after a backup
// restore (or any other operation that replaces the file
// underneath us) so the in-memory connection pool re-reads
// the new file's contents. The master key, encryption
// parameters, and migration history are preserved — only
// the *sql.DB connection is rebuilt.
//
// Safe to call from any goroutine; the underlying SQLite
// driver handles the close/open sequencing.
func (d *DB) Reload(ctx context.Context) error {
	if d == nil {
		return errors.New("storage: nil receiver")
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	if d.sql != nil {
		_ = d.sql.Close()
	}
	dsn := fmt.Sprintf("file:%s?_pragma=journal_mode(WAL)&_pragma=foreign_keys(1)&_pragma=busy_timeout(5000)&_pragma=synchronous(NORMAL)", d.path)
	sdb, err := sql.Open("sqlite", dsn)
	if err != nil {
		return fmt.Errorf("storage: reopen sqlite: %w", err)
	}
	sdb.SetMaxOpenConns(1)
	if err := sdb.PingContext(ctx); err != nil {
		_ = sdb.Close()
		return fmt.Errorf("storage: ping after reload: %w", err)
	}
	d.sql = sdb
	// Recreate the closing channel so Close can be called again
	// after Reload (e.g. during backup restore + cleanup).
	d.closing = make(chan struct{})
	return nil
}

// Path returns the on-disk path of the database.
func (d *DB) Path() string { return d.path }

// OpenedAt returns when the database was opened.
func (d *DB) OpenedAt() time.Time { return d.openedAt }

// MasterKey returns the raw 32-byte master key. Callers that need
// to derive additional secrets (e.g. the audit log HMAC key) should
// call this. Returning the key is safe because it is only ever in
// memory after Open; it is never written to disk in plaintext.
func (d *DB) MasterKey() []byte { return d.masterKey }

// SQL returns the underlying *sql.DB for advanced use.
// Most callers should use the typed methods on this package.
func (d *DB) SQL() *sql.DB { return d.sql }

// Encrypt encrypts plaintext for storage in the given row/column.
// The returned ciphertext is safe to store in a TEXT column; it includes
// the 12-byte nonce prepended.
func (d *DB) Encrypt(plaintext []byte, rowID int64, column string) ([]byte, error) {
	nonce, err := d.newNonce()
	if err != nil {
		return nil, err
	}
	aad := []byte(fmt.Sprintf("%d:%s", rowID, column))
	sealed := d.gcm.Seal(nonce, nonce, plaintext, aad)
	return sealed, nil
}

// Decrypt decrypts ciphertext produced by Encrypt.
func (d *DB) Decrypt(ciphertext []byte, rowID int64, column string) ([]byte, error) {
	if len(ciphertext) < d.gcm.NonceSize() {
		return nil, errors.New("storage: ciphertext too short")
	}
	nonce := ciphertext[:d.gcm.NonceSize()]
	body := ciphertext[d.gcm.NonceSize():]
	aad := []byte(fmt.Sprintf("%d:%s", rowID, column))
	plain, err := d.gcm.Open(nil, nonce, body, aad)
	if err != nil {
		return nil, fmt.Errorf("storage: decrypt: %w", err)
	}
	return plain, nil
}

// EncryptString is a convenience wrapper.
func (d *DB) EncryptString(plaintext string, rowID int64, column string) (string, error) {
	ct, err := d.Encrypt([]byte(plaintext), rowID, column)
	if err != nil {
		return "", err
	}
	return base64.StdEncoding.EncodeToString(ct), nil
}

// DecryptString is a convenience wrapper. Returns the empty string on a
// nil/empty input (so nullable columns work).
func (d *DB) DecryptString(s string, rowID int64, column string) (string, error) {
	if s == "" {
		return "", nil
	}
	ct, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", fmt.Errorf("storage: base64 decode: %w", err)
	}
	pt, err := d.Decrypt(ct, rowID, column)
	if err != nil {
		return "", err
	}
	return string(pt), nil
}

func (d *DB) newNonce() ([]byte, error) {
	// Reserve a slice and backfill random bytes. We hold d.mu briefly.
	d.mu.Lock()
	defer d.mu.Unlock()
	nonce := make([]byte, d.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return nil, fmt.Errorf("storage: nonce: %w", err)
	}
	return nonce, nil
}

// -----------------------------------------------------------------------------
// Master key handling
// -----------------------------------------------------------------------------

func loadOrCreateMasterKey(cfg Config) (string, error) {
	if cfg.MasterKey != "" {
		return cfg.MasterKey, nil
	}
	if cfg.Secrets == nil {
		// Ephemeral mode (tests). Generate but don't persist.
		return generateMasterKey()
	}
	// Try to load.
	existing, err := cfg.Secrets.Get(secrets.MasterKey)
	if err == nil {
		return existing, nil
	}
	if !errors.Is(err, secrets.ErrNotFound) {
		return "", err
	}
	// Not found: generate and store.
	mk, err := generateMasterKey()
	if err != nil {
		return "", err
	}
	if err := cfg.Secrets.Set(secrets.MasterKey, mk); err != nil {
		return "", fmt.Errorf("store master key: %w", err)
	}
	return mk, nil
}

func generateMasterKey() (string, error) {
	key := make([]byte, 32)
	if _, err := io.ReadFull(rand.Reader, key); err != nil {
		return "", fmt.Errorf("generate master key: %w", err)
	}
	return base64.StdEncoding.EncodeToString(key), nil
}
