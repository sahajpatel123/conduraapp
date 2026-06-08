package audit

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/storage"
)

func setupTestLog(t *testing.T) *Log {
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

func TestLog_AppendAndList(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		err := l.Append(ctx, Event{
			Actor:   "user",
			Action:  "test.append",
			App:     "synapticd",
			Level:   "info",
			Result:  "allow",
			Message: "hello",
		})
		if err != nil {
			t.Fatal(err)
		}
	}
	got, err := l.List(ctx, Query{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 3 {
		t.Fatalf("got %d events, want 3", len(got))
	}
}

func TestLog_FilterByAction(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()
	_ = l.Append(ctx, Event{Action: "apikey.set", Actor: "gui"})
	_ = l.Append(ctx, Event{Action: "apikey.delete", Actor: "gui"})
	_ = l.Append(ctx, Event{Action: "llm.chat", Actor: "gui"})

	got, err := l.List(ctx, Query{Action: "apikey"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 2 {
		t.Fatalf("got %d events, want 2 (apikey.*)", len(got))
	}
}

func TestLog_FilterByLevel(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()
	_ = l.Append(ctx, Event{Action: "a", Level: "info"})
	_ = l.Append(ctx, Event{Action: "b", Level: "warn"})
	_ = l.Append(ctx, Event{Action: "c", Level: "error"})

	got, err := l.List(ctx, Query{Level: "error"})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d events, want 1", len(got))
	}
}

func TestLog_FilterBySince(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()
	_ = l.Append(ctx, Event{Action: "old", TS: time.Now().Add(-1 * time.Hour)})
	_ = l.Append(ctx, Event{Action: "new", TS: time.Now()})

	got, err := l.List(ctx, Query{Since: time.Now().Add(-1 * time.Minute)})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d events, want 1 (only the recent one)", len(got))
	}
	if got[0].Action != "new" {
		t.Fatalf("got action %q, want new", got[0].Action)
	}
}

func TestLog_Count(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		_ = l.Append(ctx, Event{Action: "x"})
	}
	n, err := l.Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("count = %d, want 5", n)
	}
}
