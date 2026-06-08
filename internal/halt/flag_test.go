package halt

import (
	"context"
	"path/filepath"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/storage"
)

func setupFlag(t *testing.T) *Flag {
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

func TestFlag_DefaultNotHalted(t *testing.T) {
	f := setupFlag(t)
	if err := f.Refresh(context.Background()); err != nil {
		t.Fatal(err)
	}
	if f.IsHalted() {
		t.Fatal("flag should not be halted by default")
	}
}

func TestFlag_HaltAndResume(t *testing.T) {
	f := setupFlag(t)
	ctx := context.Background()
	if err := f.Refresh(ctx); err != nil {
		t.Fatal(err)
	}

	_, err := f.Halt(ctx, "test")
	if err != nil {
		t.Fatal(err)
	}
	if !f.IsHalted() {
		t.Fatal("flag should be halted after Halt()")
	}
	s := f.Halted()
	if s.Reason != "test" {
		t.Fatalf("reason = %q, want test", s.Reason)
	}
	if s.Since.IsZero() {
		t.Fatal("since should be set")
	}

	_, err = f.Resume(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if f.IsHalted() {
		t.Fatal("flag should be resumed")
	}
}

func TestFlag_PersistsAcrossRefresh(t *testing.T) {
	f := setupFlag(t)
	ctx := context.Background()
	if err := f.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	_, _ = f.Halt(ctx, "first")
	// Simulate restart: refresh from DB.
	if err := f.Refresh(ctx); err != nil {
		t.Fatal(err)
	}
	if !f.IsHalted() {
		t.Fatal("flag should still be halted after Refresh()")
	}
}
