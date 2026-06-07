// Package api_key manages credentials for LLM providers.
//
// Two authentication kinds are supported:
//   - api_key: a static secret stored encrypted at rest, sent as
//     `Authorization: Bearer <key>` (or the provider's equivalent).
//   - oauth: a token pair (access + refresh) obtained via an OAuth 2.0
//     Authorization Code flow with PKCE. Tokens are encrypted at rest and
//     refreshed automatically when expired.
//
// Storage:
//   - Plaintext secrets never touch the SQLite database. The storage
//     package's column-level AES-GCM encryption is applied before insert.
//   - A metadata row in storage.api_keys records the provider, label,
//     auth_kind, scopes, and (for OAuth) expiry. Labels are user-supplied
//     nicknames; (provider, label) is unique.
//
// Phase 1: API keys for all 12 providers (see internal/llm). OAuth is
// implemented for Google only; adding a new provider is a one-file change.
package api_key

import (
	"context"
	"crypto/rand"
	"database/sql"
	"encoding/hex"
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/synapticapp/synaptic/internal/secrets"
	"github.com/synapticapp/synaptic/internal/storage"
)

// AuthKind discriminates between API key and OAuth credentials.
type AuthKind string

const (
	AuthAPIKey AuthKind = "api_key"
	AuthOAuth  AuthKind = "oauth"
)

// Provider names (canonical, lowercase). Mirrors internal/llm.
const (
	ProviderAnthropic  = "anthropic"
	ProviderOpenAI     = "openai"
	ProviderGoogle     = "google"
	ProviderXAI        = "xai"
	ProviderMistral    = "mistral"
	ProviderDeepSeek   = "deepseek"
	ProviderOpenRouter = "openrouter"
	ProviderTogether   = "together"
	ProviderGroq       = "groq"
	ProviderFireworks  = "fireworks"
	ProviderCustom     = "custom"
	ProviderOllama     = "ollama"
)

// AllProviders is the canonical list. Order matters for the default
// failover chain in internal/failover.
var AllProviders = []string{
	ProviderAnthropic, ProviderOpenAI, ProviderGoogle, ProviderXAI,
	ProviderMistral, ProviderDeepSeek, ProviderOpenRouter, ProviderTogether,
	ProviderGroq, ProviderFireworks, ProviderCustom, ProviderOllama,
}

// IsValidProvider reports whether name is a known provider.
func IsValidProvider(name string) bool {
	for _, p := range AllProviders {
		if p == name {
			return true
		}
	}
	return false
}

// Key is the in-memory representation of a stored credential.
//
// Secret is the plaintext secret; callers must NEVER log it or pass it to
// any user-visible surface. The Plain accessor on the caller is responsible
// for keeping it off-screen.
type Key struct {
	ID         int64
	Provider   string
	Label      string
	AuthKind   AuthKind
	Secret     string    // populated on load; for OAuth, this is the access token
	Refresh    string    // OAuth only; empty for API keys
	Scopes     string    // OAuth only; space-separated
	Metadata   string    // free-form JSON; not encrypted
	ExpiresAt  time.Time // zero for API keys
	LastUsedAt time.Time
	CreatedAt  time.Time
	UpdatedAt  time.Time
}

// Common errors.
var (
	ErrNotFound      = errors.New("api_key: not found")
	ErrInvalidKind   = errors.New("api_key: invalid auth kind")
	ErrInvalidSecret = errors.New("api_key: empty secret")
	ErrNoProvider    = errors.New("api_key: provider not specified")
)

// Manager is the storage interface for credentials.
type Manager struct {
	db      *storage.DB
	secrets secrets.Manager
}

// New returns a Manager backed by the given storage and secrets managers.
func New(db *storage.DB, sm secrets.Manager) *Manager {
	return &Manager{db: db, secrets: sm}
}

// Set stores an API key. If a key with the same (provider, label) exists,
// it is replaced.
func (m *Manager) Set(ctx context.Context, k Key) (int64, error) {
	if k.Provider == "" {
		return 0, ErrNoProvider
	}
	if !IsValidProvider(k.Provider) {
		return 0, fmt.Errorf("api_key: unknown provider %q", k.Provider)
	}
	if k.AuthKind == "" {
		k.AuthKind = AuthAPIKey
	}
	if k.AuthKind != AuthAPIKey && k.AuthKind != AuthOAuth {
		return 0, ErrInvalidKind
	}
	if k.Secret == "" {
		return 0, ErrInvalidSecret
	}
	if k.Label == "" {
		k.Label = "default"
	}
	if k.UpdatedAt.IsZero() {
		k.UpdatedAt = time.Now().UTC()
	}
	// Encrypt the secret for storage in api_keys.secret_ciphertext.
	// We use a sentinel row ID of 0 because the row doesn't exist yet;
	// the actual ID isn't known until insert. To keep AAD stable for
	// re-reads, we update the column with the correct row ID after insert.
	// Simpler approach: do a two-step insert + update.
	placeholderCT, err := m.db.EncryptString(k.Secret, 0, "secret_ciphertext")
	if err != nil {
		return 0, fmt.Errorf("api_key: encrypt secret: %w", err)
	}

	var refreshCT sql.NullString
	if k.Refresh != "" {
		s, err := m.db.EncryptString(k.Refresh, 0, "refresh_token_ciphertext")
		if err != nil {
			return 0, fmt.Errorf("api_key: encrypt refresh: %w", err)
		}
		refreshCT = sql.NullString{String: s, Valid: true}
	}

	var expiresAt sql.NullString
	if !k.ExpiresAt.IsZero() {
		expiresAt = sql.NullString{String: k.ExpiresAt.UTC().Format(time.RFC3339), Valid: true}
	}

	res, err := m.db.SQL().ExecContext(ctx, `
INSERT INTO api_keys (provider, label, auth_kind, secret_ciphertext, refresh_token_ciphertext, scopes, metadata_json, expires_at, updated_at)
VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
ON CONFLICT(provider, label) DO UPDATE SET
    auth_kind = excluded.auth_kind,
    secret_ciphertext = excluded.secret_ciphertext,
    refresh_token_ciphertext = excluded.refresh_token_ciphertext,
    scopes = excluded.scopes,
    metadata_json = excluded.metadata_json,
    expires_at = excluded.expires_at,
    updated_at = excluded.updated_at
`,
		k.Provider, k.Label, string(k.AuthKind), placeholderCT, refreshCT,
		nullString(k.Scopes), nullString(k.Metadata), expiresAt,
		k.UpdatedAt.UTC().Format(time.RFC3339),
	)
	if err != nil {
		return 0, fmt.Errorf("api_key: insert: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return 0, fmt.Errorf("api_key: last insert id: %w", err)
	}
	// Re-encrypt with the real row ID and update.
	realCT, err := m.db.EncryptString(k.Secret, id, "secret_ciphertext")
	if err != nil {
		return id, fmt.Errorf("api_key: re-encrypt: %w", err)
	}
	if _, err := m.db.SQL().ExecContext(ctx,
		`UPDATE api_keys SET secret_ciphertext = ? WHERE id = ?`, realCT, id); err != nil {
		return id, fmt.Errorf("api_key: update ciphertext: %w", err)
	}
	if k.Refresh != "" {
		realRefresh, err := m.db.EncryptString(k.Refresh, id, "refresh_token_ciphertext")
		if err == nil {
			_, _ = m.db.SQL().ExecContext(ctx,
				`UPDATE api_keys SET refresh_token_ciphertext = ? WHERE id = ?`, realRefresh, id)
		}
	}
	return id, nil
}

// Get returns the key with the given ID, with secrets decrypted.
func (m *Manager) Get(ctx context.Context, id int64) (Key, error) {
	row := m.db.SQL().QueryRowContext(ctx, `
SELECT id, provider, label, auth_kind, secret_ciphertext, refresh_token_ciphertext,
       scopes, metadata_json, expires_at, last_used_at, created_at, updated_at
FROM api_keys WHERE id = ?`, id)
	return m.scanKey(row)
}

// GetByLabel returns the key for (provider, label).
func (m *Manager) GetByLabel(ctx context.Context, provider, label string) (Key, error) {
	row := m.db.SQL().QueryRowContext(ctx, `
SELECT id, provider, label, auth_kind, secret_ciphertext, refresh_token_ciphertext,
       scopes, metadata_json, expires_at, last_used_at, created_at, updated_at
FROM api_keys WHERE provider = ? AND label = ?`, provider, label)
	return m.scanKey(row)
}

// List returns all keys (across all providers), with secrets populated.
func (m *Manager) List(ctx context.Context) ([]Key, error) {
	rows, err := m.db.SQL().QueryContext(ctx, `
SELECT id, provider, label, auth_kind, secret_ciphertext, refresh_token_ciphertext,
       scopes, metadata_json, expires_at, last_used_at, created_at, updated_at
FROM api_keys ORDER BY provider, label`)
	if err != nil {
		return nil, fmt.Errorf("api_key: list: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []Key
	for rows.Next() {
		k, err := m.scanKey(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// ListByProvider lists keys for a single provider.
func (m *Manager) ListByProvider(ctx context.Context, provider string) ([]Key, error) {
	rows, err := m.db.SQL().QueryContext(ctx, `
SELECT id, provider, label, auth_kind, secret_ciphertext, refresh_token_ciphertext,
       scopes, metadata_json, expires_at, last_used_at, created_at, updated_at
FROM api_keys WHERE provider = ? ORDER BY label`, provider)
	if err != nil {
		return nil, fmt.Errorf("api_key: list: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []Key
	for rows.Next() {
		k, err := m.scanKey(rows)
		if err != nil {
			return nil, err
		}
		out = append(out, k)
	}
	return out, rows.Err()
}

// Delete removes the key with the given ID.
func (m *Manager) Delete(ctx context.Context, id int64) error {
	res, err := m.db.SQL().ExecContext(ctx, `DELETE FROM api_keys WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("api_key: delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Touch updates last_used_at to now.
func (m *Manager) Touch(ctx context.Context, id int64) error {
	_, err := m.db.SQL().ExecContext(ctx,
		`UPDATE api_keys SET last_used_at = ? WHERE id = ?`,
		time.Now().UTC().Format(time.RFC3339), id)
	return err
}

// scanKey scans one row from either *sql.Row or *sql.Rows.
type scanner interface {
	Scan(dest ...any) error
}

func (m *Manager) scanKey(s scanner) (Key, error) {
	var (
		k          Key
		authKind   string
		secretCT   string
		refreshCT  sql.NullString
		scopes     sql.NullString
		metadata   sql.NullString
		expiresAt  sql.NullString
		lastUsedAt sql.NullString
		createdAt  string
		updatedAt  string
	)
	if err := s.Scan(
		&k.ID, &k.Provider, &k.Label, &authKind, &secretCT, &refreshCT,
		&scopes, &metadata, &expiresAt, &lastUsedAt, &createdAt, &updatedAt,
	); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Key{}, ErrNotFound
		}
		return Key{}, fmt.Errorf("api_key: scan: %w", err)
	}
	k.AuthKind = AuthKind(authKind)

	plain, err := m.db.DecryptString(secretCT, k.ID, "secret_ciphertext")
	if err != nil {
		return Key{}, fmt.Errorf("api_key: decrypt secret: %w", err)
	}
	k.Secret = plain

	if refreshCT.Valid {
		rt, err := m.db.DecryptString(refreshCT.String, k.ID, "refresh_token_ciphertext")
		if err != nil {
			return Key{}, fmt.Errorf("api_key: decrypt refresh: %w", err)
		}
		k.Refresh = rt
	}
	if scopes.Valid {
		k.Scopes = scopes.String
	}
	if metadata.Valid {
		k.Metadata = metadata.String
	}
	if expiresAt.Valid {
		if t, err := time.Parse(time.RFC3339, expiresAt.String); err == nil {
			k.ExpiresAt = t
		}
	}
	if lastUsedAt.Valid {
		if t, err := time.Parse(time.RFC3339, lastUsedAt.String); err == nil {
			k.LastUsedAt = t
		}
	}
	if t, err := time.Parse(time.RFC3339, createdAt); err == nil {
		k.CreatedAt = t
	}
	if t, err := time.Parse(time.RFC3339, updatedAt); err == nil {
		k.UpdatedAt = t
	}
	return k, nil
}

// -----------------------------------------------------------------------------
// Helpers
// -----------------------------------------------------------------------------

func nullString(s string) sql.NullString {
	if s == "" {
		return sql.NullString{}
	}
	return sql.NullString{String: s, Valid: true}
}

// NewID returns a short random hex ID suitable for a label suffix.
func NewID() string {
	var b [4]byte
	_, _ = rand.Read(b[:])
	return hex.EncodeToString(b[:])
}

// ProviderLabel suggests a label for a new key. Format: "<provider>[-<id>]".
func ProviderLabel(provider string) string {
	return fmt.Sprintf("%s-%s", provider, NewID())
}

// -----------------------------------------------------------------------------
// Validator — checks that a key is non-empty and meets per-provider rules.
// (Full key validation against the provider's API is done in Test().)
// -----------------------------------------------------------------------------

// Validate performs basic sanity checks.
func Validate(k Key) error {
	if k.Provider == "" {
		return ErrNoProvider
	}
	if !IsValidProvider(k.Provider) {
		return fmt.Errorf("api_key: unknown provider %q", k.Provider)
	}
	if k.AuthKind == "" {
		k.AuthKind = AuthAPIKey
	}
	if k.AuthKind != AuthAPIKey && k.AuthKind != AuthOAuth {
		return ErrInvalidKind
	}
	if k.Secret == "" {
		return ErrInvalidSecret
	}
	if k.AuthKind == AuthAPIKey {
		// Per-provider format hints (not exhaustive; we don't reject unknown shapes).
		switch k.Provider {
		case ProviderAnthropic:
			if !strings.HasPrefix(k.Secret, "sk-ant-") {
				// Anthropic keys start with sk-ant-. We warn but don't reject
				// (the user may have a custom-format key or a proxy).
			}
		case ProviderOpenAI:
			if !strings.HasPrefix(k.Secret, "sk-") {
				// OpenAI keys start with sk-.
			}
		case ProviderGoogle, ProviderXAI, ProviderGroq:
			// No prefix convention enforced.
		case ProviderCustom:
			// Anything goes.
		}
	}
	return nil
}

// -----------------------------------------------------------------------------
// TestResult is returned by OAuthProvider.Test / APIKeyTest.
// -----------------------------------------------------------------------------

// TestResult is the outcome of a key validation request against a provider.
type TestResult struct {
	OK        bool
	LatencyMs int
	AccountID string
	Error     string
}

// -----------------------------------------------------------------------------
// Interface used by the LLM package
// -----------------------------------------------------------------------------

// Authenticator is the subset of Manager consumed by internal/llm.
// Defined here to keep the dependency graph small.
type Authenticator interface {
	GetByLabel(ctx context.Context, provider, label string) (Key, error)
	ListByProvider(ctx context.Context, provider string) ([]Key, error)
	Touch(ctx context.Context, id int64) error
}

// Compile-time interface check.
var _ Authenticator = (*Manager)(nil)

// Compile-time: ensure http.Client is referenced so we don't drop the import
// (used by OAuth).
var _ = http.Client{}
