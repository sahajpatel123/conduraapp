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
	if err := s.Save(ctx, "telegram", "12345", true); err != nil {
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
	_ = s.Save(ctx, "telegram", "12345", true)
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
	_ = s.Save(ctx, "telegram", "111", true)
	_ = s.Save(ctx, "telegram", "222", false)
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

func TestManager_CreateChannel(t *testing.T) {
	s := newTestReachStore(t)
	m := NewManager(s)
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
	m := NewManager(s)
	_, err := m.getOrCreateChannel("whatsapp")
	if err == nil {
		t.Fatal("should reject unknown channel")
	}
	var ue *UnsupportedError
	if !errors.As(err, &ue) {
		t.Fatalf("wrong error type: %T", err)
	}
	if ue.Name != "whatsapp" {
		t.Fatalf("error name: got %q", ue.Name)
	}
}

func TestManager_List(t *testing.T) {
	s := newTestReachStore(t)
	ctx := context.Background()
	_ = s.Save(ctx, "telegram", "12345", true)
	m := NewManager(s)
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
	m := NewManager(s)
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
	_ = s.Save(ctx, "telegram", "12345", true)
	m := NewManager(s)
	if err := m.Disconnect(ctx, "telegram"); err != nil {
		t.Fatalf("Disconnect: %v", err)
	}
	statuses, _ := m.List(ctx)
	if len(statuses) != 0 {
		t.Fatal("disconnect should clear the entry")
	}
}

func TestUnsupportedError(t *testing.T) {
	e := &UnsupportedError{Name: "test"}
	if e.Error() != "reach: unsupported channel: test" {
		t.Fatalf("Error(): got %q", e.Error())
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
	_ = s1.Save(ctx, "telegram", "chat-123", true)
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
