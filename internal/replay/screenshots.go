// Package replay contains the on-disk screenshot buffer for Action
// Replay (Phase 11, sub-phase 11A).
//
// Screenshots are the most sensitive data in the app — they show the
// user's screen. We store them on disk inside the Synaptic data
// directory, encrypted with the storage.DB master key, with a 24h
// TTL. The auto-pruner runs on every Put and on a background ticker.
//
// Storage layout:
//
//	<data-dir>/replay/<YYYY-MM-DD>/<id>.bin
//
// The .bin file is the encrypted screenshot bytes (AES-256-GCM,
// per-blob random nonce, AAD = id + position). The metadata index
// (id, position, dimensions, capture time) lives in the replay_screenshots
// table so the audit log can join to it by ref.
package replay

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
	"regexp"
	"strings"
	"sync"
	"time"
)

var screenshotIDRegex = regexp.MustCompile("^[0-9a-f]{16}$")

// ScreenshotStore is the on-disk screenshot buffer for replay.
type ScreenshotStore struct {
	mu      sync.Mutex
	db      *sql.DB
	root    string // <data-dir>/replay
	gcm     cipher.AEAD
	ttl     time.Duration
	stop    chan struct{}
	stopped bool
}

// NewScreenshotStore creates a ScreenshotStore rooted at
// <data-dir>/replay. The master key is the same one the storage.DB
// uses (so a single keyring entry protects everything).
func NewScreenshotStore(db *sql.DB, dataDir string, masterKey []byte) (*ScreenshotStore, error) {
	if db == nil {
		return nil, errors.New("replay: db is required")
	}
	if dataDir == "" {
		return nil, errors.New("replay: dataDir is required")
	}
	if len(masterKey) != 32 {
		return nil, errors.New("replay: masterKey must be 32 bytes")
	}
	block, err := aes.NewCipher(masterKey)
	if err != nil {
		return nil, fmt.Errorf("replay: aes.NewCipher: %w", err)
	}
	gcm, err := cipher.NewGCM(block)
	if err != nil {
		return nil, fmt.Errorf("replay: cipher.NewGCM: %w", err)
	}
	root := filepath.Join(dataDir, "replay")
	if err := os.MkdirAll(root, 0o700); err != nil {
		return nil, fmt.Errorf("replay: mkdir: %w", err)
	}
	s := &ScreenshotStore{
		db:   db,
		root: root,
		gcm:  gcm,
		ttl:  24 * time.Hour,
		stop: make(chan struct{}),
	}
	s.startBackgroundPruner()
	return s, nil
}

// startBackgroundPruner runs TTL cleanup on a ticker so expired
// screenshots are removed even when no new Put occurs.
func (s *ScreenshotStore) startBackgroundPruner() {
	go func() {
		ticker := time.NewTicker(time.Hour)
		defer ticker.Stop()
		for {
			select {
			case <-s.stop:
				return
			case <-ticker.C:
				_ = s.Prune(context.Background(), time.Time{})
			}
		}
	}()
}

// SetTTL overrides the default 24h TTL. Mostly for tests.
func (s *ScreenshotStore) SetTTL(ttl time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if ttl > 0 {
		s.ttl = ttl
	}
}

// Put stores an encrypted screenshot and returns its id (the
// reference that the audit log stores in screenshot_before_ref /
// screenshot_after_ref). The position is "before" or "after" so the
// audit log can record which twin-snapshot is which.
//
// Returns the id (a ULID-like timestamp+random string) on success.
func (s *ScreenshotStore) Put(ctx context.Context, position string, width, height int, png []byte) (string, error) {
	if position != "before" && position != "after" {
		return "", fmt.Errorf("replay: invalid position %q", position)
	}
	if len(png) == 0 {
		return "", errors.New("replay: empty screenshot")
	}
	id := newScreenshotID()
	capturedAt := time.Now().UTC()

	// Encrypt.
	nonce := make([]byte, s.gcm.NonceSize())
	if _, err := io.ReadFull(rand.Reader, nonce); err != nil {
		return "", fmt.Errorf("replay: nonce: %w", err)
	}
	aad := []byte(id + ":" + position)
	sealed := s.gcm.Seal(nonce, nonce, png, aad)

	// Write to disk.
	day := capturedAt.Format("2006-01-02")
	dir := filepath.Join(s.root, day)
	if err := os.MkdirAll(dir, 0o700); err != nil {
		return "", fmt.Errorf("replay: mkdir day: %w", err)
	}
	path := filepath.Join(dir, id+".bin")
	if err := os.WriteFile(path, sealed, 0o600); err != nil {
		return "", fmt.Errorf("replay: write: %w", err)
	}

	// Index in the metadata table.
	if _, err := s.db.ExecContext(ctx,
		`INSERT INTO replay_screenshots (id, captured_at, audit_event_id, position, width, height, byte_size)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		id, capturedAt.Format(time.RFC3339Nano), 0, position, width, height, len(png),
	); err != nil {
		_ = os.Remove(path)
		return "", fmt.Errorf("replay: insert index: %w", err)
	}
	// Best-effort prune of expired entries.
	_ = s.Prune(ctx, capturedAt)
	return id, nil
}

// Get retrieves and decrypts a screenshot by id. Returns
// (nil, nil) if the id is unknown.
func (s *ScreenshotStore) Get(ctx context.Context, id string) ([]byte, error) {
	var capturedAtStr, position string
	err := s.db.QueryRowContext(ctx,
		`SELECT captured_at, position FROM replay_screenshots WHERE id = ?`, id,
	).Scan(&capturedAtStr, &position)
	if errors.Is(err, sql.ErrNoRows) {
		return nil, nil
	}
	if err != nil {
		return nil, fmt.Errorf("replay: lookup: %w", err)
	}
	capturedAt, _ := time.Parse(time.RFC3339Nano, capturedAtStr)
	day := capturedAt.UTC().Format("2006-01-02")
	if !screenshotIDRegex.MatchString(id) {
		return nil, fmt.Errorf("replay: invalid screenshot id %q", id)
	}
	path := filepath.Join(s.root, day, id+".bin")
	if !strings.HasPrefix(filepath.Clean(path), s.root) {
		return nil, fmt.Errorf("replay: screenshot path escapes root")
	}
	sealed, err := os.ReadFile(path) //nolint:gosec // id validated and path root-checked above
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, fmt.Errorf("replay: read: %w", err)
	}
	if len(sealed) < s.gcm.NonceSize() {
		return nil, fmt.Errorf("replay: ciphertext truncated")
	}
	nonce := sealed[:s.gcm.NonceSize()]
	body := sealed[s.gcm.NonceSize():]
	aad := []byte(id + ":" + position)
	plain, err := s.gcm.Open(nil, nonce, body, aad)
	if err != nil {
		return nil, fmt.Errorf("replay: decrypt: %w", err)
	}
	return plain, nil
}

// Prune removes screenshots older than the TTL. Safe to call
// concurrently; takes s.mu to avoid stepping on Put's own prune.
func (s *ScreenshotStore) Prune(ctx context.Context, now time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if now.IsZero() {
		now = time.Now().UTC()
	}
	cutoff := now.Add(-s.ttl)
	cutoffStr := cutoff.Format(time.RFC3339Nano)
	// Find expired ids.
	rows, err := s.db.QueryContext(ctx,
		`SELECT id, captured_at FROM replay_screenshots WHERE captured_at < ?`,
		cutoffStr,
	)
	if err != nil {
		return fmt.Errorf("replay: prune query: %w", err)
	}
	type entry struct {
		id  string
		day string
	}
	var expired []entry
	for rows.Next() {
		var id, ts string
		if err := rows.Scan(&id, &ts); err != nil {
			_ = rows.Close()
			return err
		}
		day, _ := time.Parse(time.RFC3339Nano, ts)
		expired = append(expired, entry{id: id, day: day.UTC().Format("2006-01-02")})
	}
	_ = rows.Close()
	for _, e := range expired {
		_ = os.Remove(filepath.Join(s.root, e.day, e.id+".bin"))
		_, _ = s.db.ExecContext(ctx, `DELETE FROM replay_screenshots WHERE id = ?`, e.id)
	}
	return nil
}

// Close stops the background pruner.
func (s *ScreenshotStore) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.stopped {
		return nil
	}
	s.stopped = true
	close(s.stop)
	return nil
}

// newScreenshotID returns a base32-like ID encoded as a short
// string: 8 bytes of random data, hex-encoded. Total length 16.
// The audit log's metadata table uses this as the primary key.
func newScreenshotID() string {
	var b [8]byte
	if _, err := io.ReadFull(rand.Reader, b[:]); err != nil {
		// crypto/rand should never fail; if it does, fall back to time.
		return base64.RawURLEncoding.EncodeToString([]byte(time.Now().Format(time.RFC3339Nano)))
	}
	return fmt.Sprintf("%x", b[:])
}
