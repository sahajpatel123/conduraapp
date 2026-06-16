// Package account implements optional user sign-in (Phase 14B).
//
// Synaptic is local-first and fully usable signed-out; an account
// is purely additive (cloud sync, Hub publishing identity, and
// cross-device settings). Sign-in supports OAuth (Google, GitHub —
// see oauth.go) and email magic links (see magic.go). OAuth refresh
// tokens are held by a TokenManager (OS keychain, or an encrypted
// file fallback — see keychain.go). The current session is persisted
// in the main SQLite DB so it survives restarts.
package account

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net/mail"
	"sync"
	"time"
)

// validEmail reports whether email is a syntactically valid address.
func validEmail(email string) bool {
	_, err := mail.ParseAddress(email)
	return err == nil
}

// DefaultSessionTTL is used when the configured TTL is non-positive.
const DefaultSessionTTL = 720 * time.Hour // 30 days

// Session is an authenticated local session.
type Session struct {
	Email     string    `json:"email"`
	Provider  string    `json:"provider"`
	AvatarURL string    `json:"avatar_url"`
	ExpiresAt time.Time `json:"expires_at"`
}

// Expired reports whether the session has passed its TTL.
func (s *Session) Expired() bool {
	return s != nil && !s.ExpiresAt.IsZero() && time.Now().After(s.ExpiresAt)
}

// TokenManager persists OAuth tokens. Implemented by the OS
// keychain adapter and the encrypted-file fallback (keychain.go).
type TokenManager interface {
	Get(key string) (string, error)
	Set(key, value string) error
	Delete(key string) error
}

// oauthStateEntry tracks an in-flight OAuth authorization (PKCE
// verifier + provider) keyed by the random state value.
type oauthStateEntry struct {
	verifier  string
	provider  string
	expiresAt time.Time
}

// Manager owns sessions, OAuth state, and token storage.
type Manager struct {
	store        *Store
	tokenManager TokenManager
	masterKey    []byte
	sessionTTL   time.Duration

	// oauthStates maps an OAuth state value → oauthStateEntry.
	// Entries expire after 5 minutes (see oauth.go).
	oauthStates sync.Map
}

// NewManager constructs a Manager. store and tm are required.
func NewManager(store *Store, tm TokenManager, masterKey []byte, ttl time.Duration) (*Manager, error) {
	if store == nil {
		return nil, errors.New("account: nil store")
	}
	if tm == nil {
		return nil, errors.New("account: nil token manager")
	}
	if ttl <= 0 {
		ttl = DefaultSessionTTL
	}
	return &Manager{
		store:        store,
		tokenManager: tm,
		masterKey:    masterKey,
		sessionTTL:   ttl,
	}, nil
}

// NewSession creates, persists, and returns a session for the given
// identity. Every successful sign-in path (OAuth, magic link)
// funnels through here.
func (m *Manager) NewSession(ctx context.Context, email, provider string) (*Session, error) {
	if email == "" {
		return nil, errors.New("account: empty email")
	}
	sess := &Session{
		Email:     email,
		Provider:  provider,
		ExpiresAt: time.Now().Add(m.sessionTTL),
	}
	if err := m.store.Save(ctx, sess); err != nil {
		return nil, fmt.Errorf("account: persist session: %w", err)
	}
	return sess, nil
}

// Status returns the current session, or nil when signed out or the
// session has expired (expired sessions are cleared as a side
// effect).
func (m *Manager) Status(ctx context.Context) (*Session, error) {
	sess, err := m.store.Load(ctx)
	if err != nil {
		return nil, err
	}
	if sess == nil {
		return nil, nil //nolint:nilnil // nil Session means not signed in
	}
	if sess.Expired() {
		_ = m.store.Clear(ctx)
		return nil, nil //nolint:nilnil // expired session means not signed in
	}
	return sess, nil
}

// SignOut clears the local session and best-effort removes stored
// OAuth tokens.
func (m *Manager) SignOut(ctx context.Context) error {
	if sess, _ := m.store.Load(ctx); sess != nil && sess.Provider != "" {
		_ = m.tokenManager.Delete("oauth-" + sess.Provider)
		_ = m.tokenManager.Delete("oauth-" + sess.Provider + "-refresh")
	}
	return m.store.Clear(ctx)
}

// Store persists the current session in the main SQLite database.
// Exactly one session is stored at a time (single-user desktop app),
// keyed by the constant row id 1.
type Store struct {
	db *sql.DB
}

// NewStore creates the session table if needed and returns a Store.
func NewStore(db *sql.DB) (*Store, error) {
	if db == nil {
		return nil, errors.New("account: nil db")
	}
	const ddl = `
CREATE TABLE IF NOT EXISTS account_session (
	id         INTEGER PRIMARY KEY CHECK (id = 1),
	email      TEXT NOT NULL,
	provider   TEXT NOT NULL,
	avatar_url TEXT NOT NULL DEFAULT '',
	expires_at TEXT NOT NULL
);`
	if _, err := db.ExecContext(context.Background(), ddl); err != nil {
		return nil, fmt.Errorf("account: create table: %w", err)
	}
	return &Store{db: db}, nil
}

// Save upserts the single session row.
func (s *Store) Save(ctx context.Context, sess *Session) error {
	if sess == nil {
		return errors.New("account: nil session")
	}
	const q = `
INSERT INTO account_session (id, email, provider, avatar_url, expires_at)
VALUES (1, ?, ?, ?, ?)
ON CONFLICT(id) DO UPDATE SET
	email = excluded.email,
	provider = excluded.provider,
	avatar_url = excluded.avatar_url,
	expires_at = excluded.expires_at;`
	_, err := s.db.ExecContext(ctx, q,
		sess.Email, sess.Provider, sess.AvatarURL,
		sess.ExpiresAt.UTC().Format(time.RFC3339))
	return err
}

// Load returns the stored session, or nil when none exists.
func (s *Store) Load(ctx context.Context) (*Session, error) {
	const q = `SELECT email, provider, avatar_url, expires_at FROM account_session WHERE id = 1;`
	row := s.db.QueryRowContext(ctx, q)
	var (
		sess      Session
		expiresAt string
	)
	switch err := row.Scan(&sess.Email, &sess.Provider, &sess.AvatarURL, &expiresAt); {
	case errors.Is(err, sql.ErrNoRows):
		return nil, nil //nolint:nilnil // no active session
	case err != nil:
		return nil, err
	}
	if t, err := time.Parse(time.RFC3339, expiresAt); err == nil {
		sess.ExpiresAt = t
	}
	return &sess, nil
}

// Clear removes the stored session (idempotent).
func (s *Store) Clear(ctx context.Context) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM account_session WHERE id = 1;`)
	return err
}
