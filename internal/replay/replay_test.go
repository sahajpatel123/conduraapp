package replay

import (
	"bytes"
	"context"
	"errors"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/audit"
	"github.com/sahajpatel123/synapticapp/internal/storage"
)

// seedEvents appends n events to the audit log with sequential
// timestamps in the recent past.
func seedEvents(t *testing.T, l *audit.Log, n int) {
	t.Helper()
	ctx := context.Background()
	base := time.Now().Add(-time.Duration(n) * time.Minute)
	for i := 0; i < n; i++ {
		ev := &audit.Event{
			Actor:   "gatekeeper",
			Action:  "consent.request",
			App:     "condura-gui",
			Level:   "info",
			Result:  "allow",
			Message: "test",
			Kind:    "consent.request",
			TS:      base.Add(time.Duration(i) * time.Minute),
		}
		if err := l.Append(ctx, *ev); err != nil {
			t.Fatal(err)
		}
	}
}

func setupReplay(t *testing.T, withShots bool) (*Replay, *audit.Log, *ScreenshotStore) {
	t.Helper()
	dir := t.TempDir()
	db, err := storage.Open(context.Background(), storage.Config{
		Path: filepath.Join(dir, "test.db"),
	})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })
	l := audit.New(db.SQL(), db.MasterKey())
	var shots *ScreenshotStore
	if withShots {
		shots, err = NewScreenshotStore(db.SQL(), dir, db.MasterKey())
		if err != nil {
			t.Fatal(err)
		}
		t.Cleanup(func() { _ = shots.Close() })
	}
	r, err := New(Options{Audit: l, Screenshots: shots})
	if err != nil {
		t.Fatal(err)
	}
	return r, l, shots
}

func TestReplay_TimelineChronological(t *testing.T) {
	r, l, _ := setupReplay(t, false)
	seedEvents(t, l, 5)
	got, err := r.Timeline(context.Background(), time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 5 {
		t.Fatalf("got %d frames, want 5", len(got))
	}
	// The audit log returns ts DESC; we re-order to chronological.
	for i := 1; i < len(got); i++ {
		if got[i].Event.TS.Before(got[i-1].Event.TS) {
			t.Errorf("frame %d ts %v before frame %d ts %v (not chronological)",
				i, got[i].Event.TS, i-1, got[i-1].Event.TS)
		}
	}
}

func TestReplay_TimelinePrunesExpired(t *testing.T) {
	r, l, _ := setupReplay(t, false)
	ctx := context.Background()
	// Insert one event 48h ago and one now.
	old := &audit.Event{
		Actor: "u", Action: "old", App: "condurad",
		TS:     time.Now().Add(-48 * time.Hour),
		Result: "allow",
	}
	if err := l.Append(ctx, *old); err != nil {
		t.Fatal(err)
	}
	old.ID = 1
	recent := &audit.Event{
		Actor: "u", Action: "new", App: "condurad",
		Result: "allow",
	}
	if err := l.Append(ctx, *recent); err != nil {
		t.Fatal(err)
	}
	recent.ID = 2
	got, err := r.Timeline(ctx, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d frames, want 1 (the 24h-pruned one)", len(got))
	}
	if got[0].Event.ID != 2 {
		t.Errorf("kept event id = %d, want 2", got[0].Event.ID)
	}
}

func TestReplay_OutcomeClassification(t *testing.T) {
	r, l, _ := setupReplay(t, false)
	ctx := context.Background()
	_ = l.Append(ctx, audit.Event{Actor: "u", Action: "a", Result: "allow"})
	_ = l.Append(ctx, audit.Event{Actor: "u", Action: "a", Result: "deny"})
	_ = l.Append(ctx, audit.Event{Actor: "u", Action: "a", Result: "error"})
	_ = l.Append(ctx, audit.Event{Actor: "u", Action: "a", Result: "weird"})

	got, err := r.Timeline(ctx, time.Time{})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 4 {
		t.Fatalf("got %d, want 4", len(got))
	}
	want := []Outcome{OutcomeAllowed, OutcomeDenied, OutcomeErrored, OutcomeUnknown}
	for i, f := range got {
		if f.Outcome != want[i] {
			t.Errorf("frame %d outcome = %q, want %q", i, f.Outcome, want[i])
		}
	}
}

func TestReplay_FrameByID(t *testing.T) {
	r, l, _ := setupReplay(t, false)
	seedEvents(t, l, 3)
	f, err := r.FrameByID(context.Background(), 2)
	if err != nil {
		t.Fatal(err)
	}
	if f == nil {
		t.Fatal("frame is nil")
	}
	if f.Event.ID != 2 {
		t.Errorf("got id %d, want 2", f.Event.ID)
	}
}

func TestReplay_FrameByID_Unknown(t *testing.T) {
	r, _, _ := setupReplay(t, false)
	_, err := r.FrameByID(context.Background(), 999)
	if !errors.Is(err, ErrFrameNotFound) {
		t.Errorf("expected ErrFrameNotFound, got %v", err)
	}
}

func TestReplay_VerifyIntegrity(t *testing.T) {
	r, l, _ := setupReplay(t, false)
	seedEvents(t, l, 4)
	rep, err := r.VerifyIntegrity(context.Background())
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Valid {
		t.Errorf("chain invalid: %s at row %d", rep.FirstBreakReason, rep.FirstBreakID)
	}
	if rep.RowsChecked != 4 {
		t.Errorf("rows checked = %d, want 4", rep.RowsChecked)
	}
}

func TestReplay_RejectsMissingAudit(t *testing.T) {
	_, err := New(Options{})
	if err == nil {
		t.Fatal("expected error when Audit is nil")
	}
}

// ---- ScreenshotStore tests ----

func TestScreenshotStore_PutGetRoundTrip(t *testing.T) {
	_, _, shots := setupReplay(t, true)
	ctx := context.Background()
	png := []byte{0x89, 0x50, 0x4E, 0x47, 0x0D, 0x0A, 0x1A, 0x0A, 'f', 'a', 'k', 'e'}
	id, err := shots.Put(ctx, "before", 1920, 1080, png)
	if err != nil {
		t.Fatal(err)
	}
	got, err := shots.Get(ctx, id)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, png) {
		t.Errorf("round-trip mismatch: got %v, want %v", got, png)
	}
}

func TestScreenshotStore_GetUnknown(t *testing.T) {
	_, _, shots := setupReplay(t, true)
	got, err := shots.Get(context.Background(), "nonexistent")
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("expected nil for unknown id, got %v", got)
	}
}

func TestScreenshotStore_TTLPrune(t *testing.T) {
	_, _, shots := setupReplay(t, true)
	shots.SetTTL(1 * time.Hour)
	ctx := context.Background()
	// Insert one expired, one fresh.
	expiredID, err := shots.Put(ctx, "before", 100, 100, []byte("expired"))
	if err != nil {
		t.Fatal(err)
	}
	// Manually backdate it.
	_, _ = shots.db.ExecContext(ctx,
		`UPDATE replay_screenshots SET captured_at = ? WHERE id = ?`,
		time.Now().Add(-2*time.Hour).UTC().Format(time.RFC3339Nano),
		expiredID,
	)
	// Re-Prune.
	if err := shots.Prune(ctx, time.Now()); err != nil {
		t.Fatal(err)
	}
	// Expired one should be gone.
	got, err := shots.Get(ctx, expiredID)
	if err != nil {
		t.Fatal(err)
	}
	if got != nil {
		t.Errorf("expected expired screenshot to be pruned, got %v", got)
	}
	// Fresh one should be readable.
	freshID, err := shots.Put(ctx, "after", 100, 100, []byte("fresh"))
	if err != nil {
		t.Fatal(err)
	}
	got, err = shots.Get(ctx, freshID)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(got, []byte("fresh")) {
		t.Errorf("fresh screenshot got %q, want %q", got, "fresh")
	}
}

func TestScreenshotStore_EncryptedOnDisk(t *testing.T) {
	// The bytes on disk must NOT match the plaintext (encryption works).
	dir := t.TempDir()
	db, err := storage.Open(context.Background(), storage.Config{Path: filepath.Join(dir, "test.db")})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	shots, err := NewScreenshotStore(db.SQL(), dir, db.MasterKey())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = shots.Close() }()
	png := []byte("this-is-the-plaintext-png-bytes")
	id, err := shots.Put(context.Background(), "before", 1, 1, png)
	if err != nil {
		t.Fatal(err)
	}
	// Walk the day directory and read the file.
	day := time.Now().UTC().Format("2006-01-02")
	path := filepath.Join(dir, "replay", day, id+".bin")
	raw, err := osReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for i := 0; i+len(png) <= len(raw); i++ {
		if bytes.Equal(raw[i:i+len(png)], png) {
			t.Fatalf("plaintext found at offset %d in %s — encryption broken", i, path)
		}
	}
}

func TestScreenshotStore_RejectsBadPosition(t *testing.T) {
	_, _, shots := setupReplay(t, true)
	if _, err := shots.Put(context.Background(), "sideways", 1, 1, []byte("x")); err == nil {
		t.Error("expected error for bad position")
	}
}

// osReadFile is a thin alias so the test reads cleanly.
func osReadFile(p string) ([]byte, error) {
	return os.ReadFile(p)
}
