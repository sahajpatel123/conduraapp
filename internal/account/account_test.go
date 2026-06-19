package account

import (
	"context"
	"database/sql"
	"path/filepath"
	"testing"
	"time"

	_ "modernc.org/sqlite"
)

func newTestDB(t *testing.T) *sql.DB {
	t.Helper()
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "account.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return db
}

func newTestStore(t *testing.T) *Store {
	t.Helper()
	s, err := NewStore(newTestDB(t))
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

type fakeTokenManager struct {
	data map[string]string
}

func (f *fakeTokenManager) Get(k string) (string, error) {
	v, ok := f.data[k]
	if !ok {
		return "", sql.ErrNoRows
	}
	return v, nil
}
func (f *fakeTokenManager) Set(k, v string) error {
	f.data[k] = v
	return nil
}
func (f *fakeTokenManager) Delete(k string) error {
	delete(f.data, k)
	return nil
}

func newTestManager(t *testing.T) (*Manager, *Store) {
	t.Helper()
	s := newTestStore(t)
	tm := &fakeTokenManager{data: make(map[string]string)}
	m, err := NewManager(s, tm, []byte("test-master-key-0123456789abcdef"), 1*time.Hour)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	return m, s
}

// --- Store tests ---

func TestStore_SaveAndLoad(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	sess := &Session{
		Email:     "user@example.com",
		Provider:  "google",
		AvatarURL: "https://example.com/avatar.png",
		ExpiresAt: time.Now().Add(1 * time.Hour),
	}
	if err := s.Save(ctx, sess); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, err := s.Load(ctx)
	if err != nil {
		t.Fatalf("Load: %v", err)
	}
	if loaded == nil {
		t.Fatal("Load returned nil session")
	}
	if loaded.Email != "user@example.com" {
		t.Fatalf("Email: got %q", loaded.Email)
	}
	if loaded.Provider != "google" {
		t.Fatalf("Provider: got %q", loaded.Provider)
	}
	if loaded.AvatarURL != "https://example.com/avatar.png" {
		t.Fatalf("AvatarURL: got %q", loaded.AvatarURL)
	}
}

func TestStore_UpsertReplacesOld(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	s1 := &Session{Email: "old@example.com", Provider: "google", ExpiresAt: time.Now().Add(1 * time.Hour)}
	_ = s.Save(ctx, s1)
	s2 := &Session{Email: "new@example.com", Provider: "github", ExpiresAt: time.Now().Add(2 * time.Hour)}
	if err := s.Save(ctx, s2); err != nil {
		t.Fatalf("Save: %v", err)
	}
	loaded, _ := s.Load(ctx)
	if loaded.Email != "new@example.com" {
		t.Fatalf("Upsert failed: got %q", loaded.Email)
	}
	if loaded.Provider != "github" {
		t.Fatalf("Upsert provider: got %q", loaded.Provider)
	}
}

func TestStore_LoadEmptyDB(t *testing.T) {
	s := newTestStore(t)
	loaded, err := s.Load(context.Background())
	if err != nil {
		t.Fatalf("Load empty: %v", err)
	}
	if loaded != nil {
		t.Fatal("Load on empty DB should return nil")
	}
}

func TestStore_Clear(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	_ = s.Save(ctx, &Session{Email: "x@x.com", Provider: "google", ExpiresAt: time.Now().Add(1 * time.Hour)})
	if err := s.Clear(ctx); err != nil {
		t.Fatalf("Clear: %v", err)
	}
	loaded, _ := s.Load(ctx)
	if loaded != nil {
		t.Fatal("Clear should remove the session")
	}
}

func TestStore_ClearIdempotent(t *testing.T) {
	s := newTestStore(t)
	ctx := context.Background()
	if err := s.Clear(ctx); err != nil {
		t.Fatalf("Clear on empty: %v", err)
	}
	if err := s.Clear(ctx); err != nil {
		t.Fatalf("Clear twice: %v", err)
	}
}

// --- Manager tests ---

func TestNewManager_RejectsNilStore(t *testing.T) {
	tm := &fakeTokenManager{data: make(map[string]string)}
	_, err := NewManager(nil, tm, []byte("key"), 1*time.Hour)
	if err == nil {
		t.Fatal("should reject nil store")
	}
}

func TestNewManager_RejectsNilTokenManager(t *testing.T) {
	s := newTestStore(t)
	_, err := NewManager(s, nil, []byte("key"), 1*time.Hour)
	if err == nil {
		t.Fatal("should reject nil token manager")
	}
}

func TestNewManager_ZeroTTLDefaults(t *testing.T) {
	s := newTestStore(t)
	tm := &fakeTokenManager{data: make(map[string]string)}
	m, err := NewManager(s, tm, []byte("key"), 0)
	if err != nil {
		t.Fatalf("NewManager: %v", err)
	}
	if m.sessionTTL != DefaultSessionTTL {
		t.Fatalf("TTL: got %v, want %v", m.sessionTTL, DefaultSessionTTL)
	}
}

func TestNewSession_CreatesSession(t *testing.T) {
	m, _ := newTestManager(t)
	ctx := context.Background()
	sess, err := m.NewSession(ctx, "hello@test.com", "google")
	if err != nil {
		t.Fatalf("NewSession: %v", err)
	}
	if sess.Email != "hello@test.com" {
		t.Fatalf("Email: got %q", sess.Email)
	}
	if sess.Provider != "google" {
		t.Fatalf("Provider: got %q", sess.Provider)
	}
	if sess.Expired() {
		t.Fatal("new session should not be expired")
	}
}

func TestNewSession_RejectsEmptyEmail(t *testing.T) {
	m, _ := newTestManager(t)
	_, err := m.NewSession(context.Background(), "", "google")
	if err == nil {
		t.Fatal("should reject empty email")
	}
}

func TestNewSession_PersistsToStore(t *testing.T) {
	m, s := newTestManager(t)
	ctx := context.Background()
	_, _ = m.NewSession(ctx, "test@test.com", "github")
	loaded, _ := s.Load(ctx)
	if loaded == nil {
		t.Fatal("session not persisted to store")
	}
	if loaded.Email != "test@test.com" {
		t.Fatalf("persisted email: got %q", loaded.Email)
	}
}

func TestStatus_ReturnsNilWhenNoSession(t *testing.T) {
	m, _ := newTestManager(t)
	sess, err := m.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if sess != nil {
		t.Fatal("Status should return nil when no session")
	}
}

func TestStatus_ReturnsSession(t *testing.T) {
	m, _ := newTestManager(t)
	ctx := context.Background()
	_, _ = m.NewSession(ctx, "me@test.com", "google")
	sess, err := m.Status(ctx)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if sess == nil {
		t.Fatal("Status should return the session")
	}
	if sess.Email != "me@test.com" {
		t.Fatalf("Email: got %q", sess.Email)
	}
}

func TestStatus_ClearsExpiredSession(t *testing.T) {
	m, s := newTestManager(t)
	ctx := context.Background()
	// Create an already-expired session directly.
	sess := &Session{
		Email:     "old@test.com",
		Provider:  "google",
		ExpiresAt: time.Now().Add(-1 * time.Hour),
	}
	_ = s.Save(ctx, sess)
	got, err := m.Status(ctx)
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if got != nil {
		t.Fatal("expired session should return nil")
	}
	// Store should be cleared.
	loaded, _ := s.Load(ctx)
	if loaded != nil {
		t.Fatal("store should be cleared after expired session")
	}
}

func TestSignOut_ClearsSessionAndTokens(t *testing.T) {
	tm := &fakeTokenManager{data: make(map[string]string)}
	s := newTestStore(t)
	m, _ := NewManager(s, tm, []byte("key"), 1*time.Hour)
	ctx := context.Background()
	_ = tm.Set("oauth-google", "access-token-123")
	_, _ = m.NewSession(ctx, "user@test.com", "google")
	if err := m.SignOut(ctx); err != nil {
		t.Fatalf("SignOut: %v", err)
	}
	loaded, _ := s.Load(ctx)
	if loaded != nil {
		t.Fatal("session should be cleared after sign out")
	}
	if _, err := tm.Get("oauth-google"); err == nil {
		t.Fatal("OAuth tokens should be deleted on sign out")
	}
}

// --- Session tests ---

func TestSession_Expired(t *testing.T) {
	tests := []struct {
		name    string
		expires time.Time
		want    bool
	}{
		{"future", time.Now().Add(1 * time.Hour), false},
		{"past", time.Now().Add(-1 * time.Hour), true},
	}
	for _, tt := range tests {
		s := &Session{ExpiresAt: tt.expires}
		if got := s.Expired(); got != tt.want {
			t.Fatalf("%s: Expired() = %v, want %v", tt.name, got, tt.want)
		}
	}
}

func TestSession_Expired_NilSession(t *testing.T) {
	var s *Session
	if s.Expired() {
		t.Fatal("nil session should not be expired")
	}
}

func TestSession_Expired_ZeroTime(t *testing.T) {
	s := &Session{ExpiresAt: time.Time{}}
	if s.Expired() {
		t.Fatal("zero-time session should not be expired")
	}
}

// --- Valid email tests ---

func TestValidEmail(t *testing.T) {
	tests := []struct {
		email string
		valid bool
	}{
		{"user@example.com", true},
		{"a@b.co", true},
		{"not-an-email", false},
		{"", false},
	}
	for _, tt := range tests {
		if got := validEmail(tt.email); got != tt.valid {
			t.Fatalf("validEmail(%q) = %v, want %v", tt.email, got, tt.valid)
		}
	}
}

// --- OAuth tests ---

func TestGenerateAuthURL_UnknownProvider(t *testing.T) {
	m, _ := newTestManager(t)
	_, _, err := m.GenerateAuthURL("unknown", "condura://auth/callback")
	if err == nil {
		t.Fatal("should reject unknown provider")
	}
}

func TestGenerateAuthURL_EmptyClientID(t *testing.T) {
	m, _ := newTestManager(t)
	_, _, err := m.GenerateAuthURL("google", "condura://auth/callback")
	if err == nil {
		t.Fatal("should reject provider with empty ClientID")
	}
}

func TestCleanupExpiredStates(t *testing.T) {
	m, _ := newTestManager(t)
	// Inject an expired state.
	m.oauthStates.Store("expired-state", oauthStateEntry{
		verifier:  "v",
		provider:  "google",
		expiresAt: time.Now().Add(-1 * time.Minute),
	})
	m.CleanupExpiredStates()
	if _, ok := m.oauthStates.Load("expired-state"); ok {
		t.Fatal("expired state should be cleaned up")
	}
}

// --- Magic link tests ---

func TestValidEmail_RejectsInvalid(t *testing.T) {
	m, _ := newTestManager(t)
	err := m.RequestMagicLink(context.Background(), "not-an-email")
	if err == nil {
		t.Fatal("should reject invalid email for magic link")
	}
}

func TestVerifyMagicToken_RejectsEmpty(t *testing.T) {
	m, _ := newTestManager(t)
	_, err := m.VerifyMagicToken(context.Background(), "")
	if err == nil {
		t.Fatal("should reject empty token")
	}
}

// --- Persistence across instances ---

func TestStore_PersistenceAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "account.db")
	db1, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open1: %v", err)
	}
	s1, err := NewStore(db1)
	if err != nil {
		t.Fatalf("NewStore1: %v", err)
	}
	ctx := context.Background()
	_ = s1.Save(ctx, &Session{Email: "persist@test.com", Provider: "google", ExpiresAt: time.Now().Add(1 * time.Hour)})
	_ = db1.Close()

	db2, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open2: %v", err)
	}
	defer func() { _ = db2.Close() }()
	s2, err := NewStore(db2)
	if err != nil {
		t.Fatalf("NewStore2: %v", err)
	}
	loaded, _ := s2.Load(ctx)
	if loaded == nil {
		t.Fatal("session lost after DB reopen")
	}
	if loaded.Email != "persist@test.com" {
		t.Fatalf("email: got %q", loaded.Email)
	}
}

// --- Keychain fallback ---

func TestFileTokenManager_RoundTrip(t *testing.T) {
	dir := t.TempDir()
	tm := NewFallbackTokenManager(dir, []byte("master-key-0123456789abcdef01234567"))
	if err := tm.Set("test-key", "test-value"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := tm.Get("test-key")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "test-value" {
		t.Fatalf("value: got %q", got)
	}
}

func TestFileTokenManager_Delete(t *testing.T) {
	dir := t.TempDir()
	tm := NewFallbackTokenManager(dir, []byte("master-key-0123456789abcdef01234567"))
	_ = tm.Set("key", "val")
	if err := tm.Delete("key"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	_, err := tm.Get("key")
	if err == nil {
		t.Fatal("Get after delete should fail")
	}
}

func TestFileTokenManager_DeleteMissingKey(t *testing.T) {
	dir := t.TempDir()
	tm := NewFallbackTokenManager(dir, []byte("master-key-0123456789abcdef01234567"))
	if err := tm.Delete("nonexistent"); err != nil {
		t.Fatalf("Delete missing key should not error: %v", err)
	}
}

func TestFileTokenManager_UpdateExisting(t *testing.T) {
	dir := t.TempDir()
	tm := NewFallbackTokenManager(dir, []byte("master-key-0123456789abcdef01234567"))
	_ = tm.Set("key", "v1")
	_ = tm.Set("key", "v2")
	got, _ := tm.Get("key")
	if got != "v2" {
		t.Fatalf("update: got %q", got)
	}
}

func TestFileTokenManager_GetMissing(t *testing.T) {
	dir := t.TempDir()
	tm := NewFallbackTokenManager(dir, []byte("master-key-0123456789abcdef01234567"))
	_, err := tm.Get("missing")
	if err == nil {
		t.Fatal("Get missing key should fail")
	}
}

func TestKeyringAdapter_RoundTrip(t *testing.T) {
	data := make(map[string]string)
	km := NewKeychainTokenManager(
		func(k string) (string, error) {
			v, ok := data[k]
			if !ok {
				return "", sql.ErrNoRows
			}
			return v, nil
		},
		func(k, v string) error { data[k] = v; return nil },
		func(k string) error { delete(data, k); return nil },
	)
	if err := km.Set("k", "v"); err != nil {
		t.Fatalf("Set: %v", err)
	}
	got, err := km.Get("k")
	if err != nil {
		t.Fatalf("Get: %v", err)
	}
	if got != "v" {
		t.Fatalf("value: got %q", got)
	}
	if err := km.Delete("k"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
}
