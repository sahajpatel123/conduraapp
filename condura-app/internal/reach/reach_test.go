package reach

import (
	"context"
	"database/sql"
	"errors"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"
)

func newTestReachStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	db, err := sql.Open("sqlite", filepath.Join(dir, "reach.db"))
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	s, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore: %v", err)
	}
	return s
}

func TestStore_SaveAndList(t *testing.T) {
	s := newTestReachStore(t)
	ctx := context.Background()
	// Save now expects (ctx, name, tokenCiphertext, chatID, enabled).
	// Pass an empty ciphertext for tests that don't exercise encryption.
	if err := s.Save(ctx, "telegram", "", "12345", true); err != nil {
		t.Fatalf("Save: %v", err)
	}
	statuses, err := s.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("list length: got %d, want 1", len(statuses))
	}
	if statuses[0].Name != "telegram" {
		t.Fatalf("name: got %q", statuses[0].Name)
	}
	if statuses[0].ChatID != "12345" {
		t.Fatalf("chatID: got %q", statuses[0].ChatID)
	}
	if !statuses[0].Connected {
		t.Fatal("should be connected")
	}
}

func TestStore_ListEmpty(t *testing.T) {
	s := newTestReachStore(t)
	statuses, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("List empty: %v", err)
	}
	if len(statuses) != 0 {
		t.Fatalf("empty list: got %d", len(statuses))
	}
}

func TestStore_Delete(t *testing.T) {
	s := newTestReachStore(t)
	ctx := context.Background()
	_ = s.Save(ctx, "telegram", "", "12345", true)
	if err := s.Delete(ctx, "telegram"); err != nil {
		t.Fatalf("Delete: %v", err)
	}
	statuses, _ := s.List(ctx)
	if len(statuses) != 0 {
		t.Fatal("Delete should remove the channel")
	}
}

func TestStore_UpdateExisting(t *testing.T) {
	s := newTestReachStore(t)
	ctx := context.Background()
	_ = s.Save(ctx, "telegram", "ct1", "111", true)
	_ = s.Save(ctx, "telegram", "ct2", "222", false)
	statuses, _ := s.List(ctx)
	if len(statuses) != 1 {
		t.Fatalf("should still have 1 entry: got %d", len(statuses))
	}
	if statuses[0].ChatID != "222" {
		t.Fatalf("chatID not updated: got %q", statuses[0].ChatID)
	}
	if statuses[0].Connected {
		t.Fatal("should be disconnected after update")
	}
}

func TestStore_UpdateChatID(t *testing.T) {
	s := newTestReachStore(t)
	ctx := context.Background()
	_ = s.Save(ctx, "telegram", "ct", "", true)
	if err := s.UpdateChatID(ctx, "telegram", "99999"); err != nil {
		t.Fatalf("UpdateChatID: %v", err)
	}
	statuses, _ := s.List(ctx)
	if len(statuses) != 1 || statuses[0].ChatID != "99999" {
		t.Fatalf("chatID not updated: %+v", statuses)
	}
}

func TestStore_GetTokenCiphertext_Empty(t *testing.T) {
	s := newTestReachStore(t)
	ctx := context.Background()
	ct, err := s.GetTokenCiphertext(ctx, "telegram")
	if err != nil {
		t.Fatalf("GetTokenCiphertext missing: %v", err)
	}
	if ct != "" {
		t.Fatalf("expected empty ciphertext for missing channel, got %q", ct)
	}
}

func TestStore_GetTokenCiphertext_Roundtrip(t *testing.T) {
	s := newTestReachStore(t)
	ctx := context.Background()
	const want = "ciphertext-blob"
	_ = s.Save(ctx, "telegram", want, "12345", true)
	got, err := s.GetTokenCiphertext(ctx, "telegram")
	if err != nil {
		t.Fatalf("GetTokenCiphertext: %v", err)
	}
	if got != want {
		t.Fatalf("ciphertext: got %q, want %q", got, want)
	}
}

func TestManager_CreateChannel(t *testing.T) {
	s := newTestReachStore(t)
	m := NewManager(s, nil, nil)
	ch, err := m.getOrCreateChannel("telegram")
	if err != nil {
		t.Fatalf("getOrCreateChannel: %v", err)
	}
	if ch == nil {
		t.Fatal("channel should not be nil")
	}
}

func TestManager_UnknownChannel(t *testing.T) {
	s := newTestReachStore(t)
	m := NewManager(s, nil, nil)
	_, err := m.getOrCreateChannel("discord")
	if err == nil {
		t.Fatal("should reject unknown channel")
	}
	var ue *UnsupportedError
	if !errors.As(err, &ue) {
		t.Fatalf("wrong error type: %T", err)
	}
	if ue.Name != "discord" {
		t.Fatalf("error name: got %q", ue.Name)
	}
}

func TestManager_WhatsAppConnectComingSoon(t *testing.T) {
	s := newTestReachStore(t)
	m := NewManager(s, nil, nil)
	_, err := m.Connect(context.Background(), "whatsapp", "token")
	if err == nil {
		t.Fatal("whatsapp Connect should fail")
	}
	var ue *UnsupportedError
	if !errors.As(err, &ue) {
		t.Fatalf("wrong error type: %T", err)
	}
	if ue.Message != whatsAppComingSoon {
		t.Fatalf("message: got %q", ue.Message)
	}
}

func TestManager_SignalConnectComingSoon(t *testing.T) {
	s := newTestReachStore(t)
	m := NewManager(s, nil, nil)
	_, err := m.Connect(context.Background(), "signal", "token")
	if err == nil {
		t.Fatal("signal Connect should fail")
	}
	var ue *UnsupportedError
	if !errors.As(err, &ue) {
		t.Fatalf("wrong error type: %T", err)
	}
	if ue.Message != signalComingSoon {
		t.Fatalf("message: got %q", ue.Message)
	}
}

func TestManager_CreateStubChannels(t *testing.T) {
	s := newTestReachStore(t)
	m := NewManager(s, nil, nil)
	for _, name := range []string{"whatsapp", "signal", "imessage"} {
		ch, err := m.getOrCreateChannel(name)
		if err != nil {
			t.Fatalf("getOrCreateChannel(%q): %v", name, err)
		}
		if ch == nil {
			t.Fatalf("channel %q is nil", name)
		}
	}
}

func TestManager_List(t *testing.T) {
	s := newTestReachStore(t)
	ctx := context.Background()
	_ = s.Save(ctx, "telegram", "", "12345", true)
	m := NewManager(s, nil, nil)
	statuses, err := m.List(ctx)
	if err != nil {
		t.Fatalf("List: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("list: got %d", len(statuses))
	}
}

func TestManager_Status(t *testing.T) {
	s := newTestReachStore(t)
	m := NewManager(s, nil, nil)
	ctx := context.Background()
	status, err := m.Status(ctx, "telegram")
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.Name != "telegram" {
		t.Fatalf("name: got %q", status.Name)
	}
}

func TestManager_Disconnect(t *testing.T) {
	s := newTestReachStore(t)
	ctx := context.Background()
	_ = s.Save(ctx, "telegram", "", "12345", true)
	m := NewManager(s, nil, nil)
	if err := m.Disconnect(ctx, "telegram"); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	statuses, _ := m.List(ctx)
	if len(statuses) != 0 {
		t.Fatal("disconnect should clear the entry")
	}
}

// TestManager_Connect_PersistsToken covers the encryption path:
// Manager.Connect with a real cipher must encrypt the plaintext
// token via the Cipher and persist the ciphertext (NOT the
// plaintext) to reach_channels.
func TestManager_Connect_PersistsToken(t *testing.T) {
	s := newTestReachStore(t)
	ctx := context.Background()
	// Use a fake cipher that records what it encrypted and
	// returns a recognizable sentinel.
	const sentinelCT = "ENC<<token-plaintext-was-secret-123>>"
	fake := &fakeCipher{
		encrypt: func(plaintext string, aad []byte) (string, error) {
			return sentinelCT, nil
		},
		decrypt: func(envelope string) (string, error) {
			return "secret-123", nil
		},
	}
	m := NewManager(s, fake, nil)

	// Build a Telegram channel with a mocked HTTP client that
	// returns getMe=ok for the Connect validation.
	tc := &telegramChannel{
		httpClient: newFakeTelegramServer(t, &fakeTelegramHandlers{getMeOK: true}),
	}
	m.channels["telegram"] = tc

	if _, err := m.Connect(ctx, "telegram", "secret-123"); err != nil {
		t.Fatalf("Connect: %v", err)
	}

	// Verify the persisted row carries the ciphertext, not the
	// plaintext. This is the bug-fix assertion: if Save stored
	// the plaintext, this would fail.
	ct, err := s.GetTokenCiphertext(ctx, "telegram")
	if err != nil {
		t.Fatalf("GetTokenCiphertext: %v", err)
	}
	if ct != sentinelCT {
		t.Fatalf("persisted ciphertext: got %q, want %q (plaintext leakage)", ct, sentinelCT)
	}
}

// TestManager_Restore_RebuildsConnectedChannel covers the boot
// path: a previously-connected channel with a stored ciphertext
// must come back online after Restore() with the same in-memory
// connected=true state.
func TestManager_Restore_RebuildsConnectedChannel(t *testing.T) {
	s := newTestReachStore(t)
	ctx := context.Background()

	// Pre-populate a connected channel with a known ciphertext
	// and a known chat_id.
	const sentinelCT = "ENC<<restore-test>>"
	_ = s.Save(ctx, "telegram", sentinelCT, "99999", true)

	fake := &fakeCipher{
		encrypt: func(string, []byte) (string, error) { return "", nil },
		decrypt: func(envelope string) (string, error) {
			if envelope != sentinelCT {
				t.Fatalf("decrypt got unexpected envelope %q", envelope)
			}
			return "restored-token", nil
		},
	}
	m := NewManager(s, fake, nil)
	// Pre-install a Telegram channel that will accept the token.
	tc := &telegramChannel{
		httpClient: newFakeTelegramServer(t, &fakeTelegramHandlers{getMeOK: true}),
	}
	m.channels["telegram"] = tc

	if err := m.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}

	// The channel should be connected now.
	status, err := m.Status(ctx, "telegram")
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if !status.Connected {
		t.Fatal("Restore should have brought the channel back online")
	}
	if status.ChatID != "99999" {
		t.Fatalf("chatID not restored: got %q, want 99999", status.ChatID)
	}
}

// TestManager_Restore_NoCipher covers the no-cipher path: if
// Manager is constructed without a cipher (e.g., tests), Restore
// is a no-op even when there are connected rows.
func TestManager_Restore_NoCipher(t *testing.T) {
	s := newTestReachStore(t)
	ctx := context.Background()
	_ = s.Save(ctx, "telegram", "ct", "12345", true)
	m := NewManager(s, nil, nil) // cipher is nil
	if err := m.Restore(ctx); err != nil {
		t.Fatalf("Restore: %v", err)
	}
	status, err := m.Status(ctx, "telegram")
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.Connected {
		t.Fatal("without a cipher, Restore should leave channel disconnected")
	}
}

func TestUnsupportedError(t *testing.T) {
	e := &UnsupportedError{Name: "test"}
	if e.Error() != "reach: unsupported channel: test" {
		t.Fatalf("Error(): got %q", e.Error())
	}
	e2 := &UnsupportedError{Name: "whatsapp", Message: whatsAppComingSoon}
	if e2.Error() != whatsAppComingSoon {
		t.Fatalf("custom Error(): got %q", e2.Error())
	}
}

func TestTelegramChannel_EmptyToken(t *testing.T) {
	tc := newTelegramChannel()
	err := tc.Connect(context.Background(), "")
	if err == nil {
		t.Fatal("should reject empty token")
	}
}

func TestChannelStatus_Defaults(t *testing.T) {
	tc := newTelegramChannel()
	status, err := tc.Status(context.Background())
	if err != nil {
		t.Fatalf("Status: %v", err)
	}
	if status.Connected {
		t.Fatal("fresh channel should not be connected")
	}
	if status.Name != "telegram" {
		t.Fatalf("name: got %q", status.Name)
	}
}

func TestPersistenceAcrossInstances(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "reach.db")
	db1, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open1: %v", err)
	}
	s1, err := NewStore(db1)
	if err != nil {
		t.Fatalf("NewStore1: %v", err)
	}
	ctx := context.Background()
	_ = s1.Save(ctx, "telegram", "ct", "chat-123", true)
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
	statuses, _ := s2.List(ctx)
	if len(statuses) != 1 {
		t.Fatal("channel lost after DB reopen")
	}
	if statuses[0].ChatID != "chat-123" {
		t.Fatalf("chatID: got %q", statuses[0].ChatID)
	}
}

// TestPersistence_MigrationV8 covers the migration path: an
// existing database created BEFORE migration v8 must come back
// with the token_ciphertext column present after v8 applies.
// Without the v8 migration, this test would fail (column missing).
func TestPersistence_MigrationV8(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "reach.db")
	db, err := sql.Open("sqlite", path)
	if err != nil {
		t.Fatalf("open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// First, simulate the pre-v8 schema by creating the table
	// WITHOUT token_ciphertext (the legacy CREATE TABLE).
	if _, err := db.ExecContext(context.Background(), `
CREATE TABLE reach_channels (
    name         TEXT PRIMARY KEY,
    token        TEXT DEFAULT '',
    chat_id      TEXT DEFAULT '',
    enabled      INTEGER DEFAULT 0,
    connected_at TEXT DEFAULT ''
);
`); err != nil {
		t.Fatalf("create legacy schema: %v", err)
	}
	if _, err := db.ExecContext(context.Background(),
		`INSERT INTO reach_channels (name, chat_id, enabled) VALUES ('telegram', 'pre-v8', 1)`,
	); err != nil {
		t.Fatalf("insert pre-v8 row: %v", err)
	}

	// Now apply migration v8.
	if _, err := db.ExecContext(context.Background(),
		`ALTER TABLE reach_channels ADD COLUMN token_ciphertext TEXT NOT NULL DEFAULT ''`,
	); err != nil {
		t.Fatalf("apply v8: %v", err)
	}

	// Open via NewStore (idempotent — won't recreate an existing table).
	s, err := NewStore(db)
	if err != nil {
		t.Fatalf("NewStore post-v8: %v", err)
	}

	// The pre-existing row must still be there with the legacy
	// chat_id, and the new column must be empty (default '').
	statuses, err := s.List(context.Background())
	if err != nil {
		t.Fatalf("List post-v8: %v", err)
	}
	if len(statuses) != 1 {
		t.Fatalf("expected 1 row post-v8, got %d", len(statuses))
	}
	if statuses[0].ChatID != "pre-v8" {
		t.Fatalf("chatID lost: got %q", statuses[0].ChatID)
	}
	if !statuses[0].Connected {
		t.Fatal("connected flag lost post-v8")
	}
	ct, err := s.GetTokenCiphertext(context.Background(), "telegram")
	if err != nil {
		t.Fatalf("GetTokenCiphertext post-v8: %v", err)
	}
	if ct != "" {
		t.Fatalf("token_ciphertext should default to '' pre-v8, got %q", ct)
	}
}

// fakeCipher implements Cipher for tests. Encryption and decryption
// behaviors are supplied via function fields so each test can
// specify what it wants to verify (e.g., "always return a sentinel"
// or "fail with a particular error").
type fakeCipher struct {
	encrypt func(plaintext string, aad []byte) (string, error)
	decrypt func(envelope string) (string, error)
}

func (f *fakeCipher) EncryptStringWithAAD(plaintext string, aad []byte) (string, error) {
	if f.encrypt == nil {
		return "", errors.New("fakeCipher: encrypt not configured")
	}
	return f.encrypt(plaintext, aad)
}

func (f *fakeCipher) DecryptStringWithAAD(envelope string) (string, error) {
	if f.decrypt == nil {
		return "", errors.New("fakeCipher: decrypt not configured")
	}
	return f.decrypt(envelope)
}
