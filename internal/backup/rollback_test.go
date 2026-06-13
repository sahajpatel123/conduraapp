package backup

import (
	"context"
	"strings"
	"testing"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/storage"
)

func TestRollback_HonestScope(t *testing.T) {
	r := NewRollback(nil)
	s := r.HonestScope()
	if !strings.Contains(s, "conversation") {
		t.Error("honest scope should mention conversation")
	}
	if !strings.Contains(s, "memory") {
		t.Error("honest scope should mention memory")
	}
	if !strings.Contains(s, "Action Replay") {
		t.Error("honest scope should redirect to Action Replay for irreversible actions")
	}
}

func TestRollback_CreateCheckpoint(t *testing.T) {
	// v0.1.0 uses in-memory checkpoints; nil DB is OK.
	r := NewRollback(nil)
	cp, err := r.CreateCheckpoint(context.Background(), "test")
	if err != nil || cp == nil {
		t.Errorf("CreateCheckpoint: err=%v, cp=%v", err, cp)
	}
	if cp.Reason != "test" {
		t.Errorf("reason = %q, want %q", cp.Reason, "test")
	}
}

func TestRollback_RevertToCheckpoint_WithoutDB(t *testing.T) {
	r := NewRollback(nil)
	_, err := r.RevertToCheckpoint(context.Background(), Checkpoint{})
	if err == nil {
		t.Error("expected error when db is nil")
	}
}

func TestRollback_RevertToCheckpoint_EmptyDB(t *testing.T) {
	dir := t.TempDir()
	db, err := storage.Open(context.Background(), storage.Config{Path: dir + "/test.db"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	r := NewRollback(db.SQL())
	n, err := r.RevertToCheckpoint(context.Background(), Checkpoint{CreatedAt: zeroTime()})
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("got %d reverts, want 0 (no rows to revert)", n)
	}
}

func TestRollback_RevertLastSession(t *testing.T) {
	dir := t.TempDir()
	db, err := storage.Open(context.Background(), storage.Config{Path: dir + "/test.db"})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	r := NewRollback(db.SQL())
	n, err := r.RevertLastSession(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if n != 0 {
		t.Errorf("got %d reverts, want 0", n)
	}
}

func zeroTime() time.Time {
	return time.Unix(0, 0)
}
