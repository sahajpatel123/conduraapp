package conversation

import (
	"context"
	"encoding/json"
	"errors"
	"path/filepath"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/storage"
)

func setupStore(t *testing.T) *Store {
	t.Helper()
	dir := t.TempDir()
	db, err := storage.Open(context.Background(), storage.Config{
		Path: filepath.Join(dir, "test.db"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return New(db.SQL())
}

func TestStore_CreateAndGet(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()
	m, err := s.Create(ctx, "test conv")
	if err != nil {
		t.Fatal(err)
	}
	if m.ID == 0 {
		t.Fatal("ID should be non-zero")
	}
	if m.Title != "test conv" {
		t.Fatalf("title = %q, want test conv", m.Title)
	}
	if m.MessageCount != 0 {
		t.Fatalf("MessageCount = %d, want 0", m.MessageCount)
	}

	// Empty get.
	c, err := s.Get(ctx, m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Messages) != 0 {
		t.Fatalf("got %d messages, want 0", len(c.Messages))
	}
}

func TestStore_Create_DefaultTitle(t *testing.T) {
	s := setupStore(t)
	m, err := s.Create(context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if m.Title != "New conversation" {
		t.Fatalf("title = %q, want default", m.Title)
	}
}

func TestStore_AppendAndList(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()
	m, err := s.Create(ctx, "thread")
	if err != nil {
		t.Fatal(err)
	}
	for _, role := range []string{"user", "assistant", "user"} {
		if err := s.Append(ctx, m.ID, Message{Role: role, Content: "hi"}); err != nil {
			t.Fatal(err)
		}
	}
	c, err := s.Get(ctx, m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Messages) != 3 {
		t.Fatalf("got %d messages, want 3", len(c.Messages))
	}
	if c.Messages[0].Role != "user" {
		t.Fatalf("first role = %q, want user", c.Messages[0].Role)
	}
	if c.MessageCount != 3 {
		t.Fatalf("count = %d, want 3", c.MessageCount)
	}

	list, err := s.List(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(list) != 1 {
		t.Fatalf("got %d conversations, want 1", len(list))
	}
	if list[0].MessageCount != 3 {
		t.Fatalf("sidebar count = %d, want 3", list[0].MessageCount)
	}
}

func TestStore_Delete(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()
	m, err := s.Create(ctx, "doomed")
	if err != nil {
		t.Fatal(err)
	}
	if err := s.Delete(ctx, m.ID); err != nil {
		t.Fatal(err)
	}
	if _, err := s.Get(ctx, m.ID); !errors.Is(err, ErrNotFound) {
		t.Fatalf("expected ErrNotFound, got %v", err)
	}
}

func TestStore_Append_ToolCalls(t *testing.T) {
	s := setupStore(t)
	ctx := context.Background()
	m, err := s.Create(ctx, "tools")
	if err != nil {
		t.Fatal(err)
	}
	tcs, _ := json.Marshal([]map[string]any{
		{"id": "call_1", "type": "function", "function": map[string]string{"name": "get_weather"}},
	})
	if err := s.Append(ctx, m.ID, Message{
		Role:      "assistant",
		Content:   "",
		ToolCalls: tcs,
	}); err != nil {
		t.Fatal(err)
	}
	c, err := s.Get(ctx, m.ID)
	if err != nil {
		t.Fatal(err)
	}
	if len(c.Messages) != 1 {
		t.Fatalf("got %d messages", len(c.Messages))
	}
	if len(c.Messages[0].ToolCalls) == 0 {
		t.Fatal("tool_calls should be preserved")
	}
}
