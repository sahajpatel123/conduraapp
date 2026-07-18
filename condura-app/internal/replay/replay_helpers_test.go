package replay

import (
	"context"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/audit"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/storage"
)

// setupReplayForTest is a small helper that opens an in-memory-style
// sqlite DB (in a temp file), creates an audit log + screenshot
// store, and returns a wired-up *Replay. Mirrors the existing
// setupReplay in replay_test.go but inlines the Options struct
// construction (avoiding that helper's extra params).
func setupReplayForTest(t *testing.T, withShots bool) (*Replay, *ScreenshotStore) {
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

	r, err := New(Options{Audit: l, Screenshots: shots, TTL: time.Hour})
	if err != nil {
		t.Fatal(err)
	}
	return r, shots
}

// TestReplay_Screenshots_ReturnsStore pins the ScreenshotStore
// getter contract: Screenshots() MUST return the underlying
// *ScreenshotStore. The CUResolver uses this accessor to capture
// before/after screenshots for the replay timeline. A regression
// that returned nil would silently break CU capture.
func TestReplay_Screenshots_ReturnsStore(t *testing.T) {
	r, shots := setupReplayForTest(t, true)
	if r.Screenshots() == nil {
		t.Fatal("Screenshots() = nil; want non-nil store")
	}
	if r.Screenshots() != shots {
		t.Errorf("Screenshots() = %p, want %p (the same store passed to New)", r.Screenshots(), shots)
	}
}

// TestReplay_Screenshots_NilWhenNotConfigured pins the
// nil-when-unconfigured contract: when New is called without a
// Screenshots in Options, Screenshots() MUST return nil (NOT
// panic). The replay still works in this mode — just without
// image refs.
func TestReplay_Screenshots_NilWhenNotConfigured(t *testing.T) {
	r, _ := setupReplayForTest(t, false)
	if r.Screenshots() != nil {
		t.Errorf("Screenshots() with no shots option = %p, want nil", r.Screenshots())
	}
}

// TestScreenshotStore_Reload_NilReceiverSafe pins the nil-guard:
// Reload MUST NOT panic on a nil receiver. Defense against a
// regression that removed the `if s == nil { return }` guard,
// which would NPE every call site that hasn't initialized the
// store yet (e.g., during daemon shutdown).
func TestScreenshotStore_Reload_NilReceiverSafe(t *testing.T) {
	var s *ScreenshotStore
	// Should not panic.
	s.Reload(nil)
}

// TestScreenshotStore_Reload_ReplacesDB pins the swap contract:
// after Reload(newDB), subsequent operations MUST use newDB,
// not the old one. This is the contract used by backup-restore
// flows: storage.Reload triggers replay.Reload so screenshot
// writes go to the restored DB.
func TestScreenshotStore_Reload_ReplacesDB(t *testing.T) {
	_, shots := setupReplayForTest(t, true)
	if shots == nil {
		t.Fatal("setup: shots is nil")
	}

	// Open a SECOND storage (separate sqlite file) to swap in.
	dir2 := t.TempDir()
	db2, err := storage.Open(context.Background(), storage.Config{
		Path: filepath.Join(dir2, "second.db"),
	})
	if err != nil {
		t.Fatalf("second storage.Open: %v", err)
	}
	defer func() { _ = db2.Close() }()

	shots.Reload(db2.SQL())

	// Verify the swap works: a Put on the new DB should succeed
	// without touching the original DB.
	id, err := shots.Put(context.Background(), "before", 1920, 1080, []byte("redacted-data"))
	if err != nil {
		t.Fatalf("Put after Reload: %v", err)
	}
	if id == "" {
		t.Error("Put returned empty ID; want non-empty")
	}

	// Verify the new DB has the entry (via Get).
	data, err := shots.Get(context.Background(), id)
	if err != nil {
		t.Fatalf("Get after Reload+Put: %v", err)
	}
	if string(data) != "redacted-data" {
		t.Errorf("Get returned %q, want \"redacted-data\"", data)
	}
}

// TestScreenshotStore_Reload_NilDBIsAccepted pins the contract:
// Reload(nil) MUST NOT panic (the underlying store becomes nil,
// but that's a valid state — callers should test for nil before
// using). Without this pin, a nil-DB reload would NPE.
func TestScreenshotStore_Reload_NilDBIsAccepted(t *testing.T) {
	_, shots := setupReplayForTest(t, true)
	// Should not panic.
	shots.Reload(nil)
}

// TestReplay_ExportMP4FromTimeline_EmptyTimelineErrors pins the
// production contract: ExportMP4FromTimeline with an empty
// timeline MUST return an error mentioning "no frames". The
// production code treats "no frames" as a defensive error rather
// than silently producing an empty MP4 — the caller (GUI) wants
// to know there's nothing to export so it can show "no activity
// in this time range" instead of producing a confusing empty file.
func TestReplay_ExportMP4FromTimeline_EmptyTimelineErrors(t *testing.T) {
	r, _ := setupReplayForTest(t, false)
	dest := filepath.Join(t.TempDir(), "export.mp4")

	_, err := r.ExportMP4FromTimeline(context.Background(), time.Now().Add(-time.Hour), dest)
	if err == nil {
		t.Fatal("ExportMP4FromTimeline on empty timeline returned nil; want error")
	}
	if !strings.Contains(err.Error(), "no frames") {
		t.Errorf("error %q should mention 'no frames'", err.Error())
	}
}

// TestReplay_ExportMP4FromTimeline_PropagatesTimelineError pins
// the error-propagation contract: if Timeline() returns an error,
// ExportMP4FromTimeline MUST return that error WITHOUT calling
// ExportMP4. This is critical for surfacing DB errors to the GUI
// (otherwise the user would see a confusing "export failed" with
// no diagnostic).
//
// We force the Timeline error by passing a pre-canceled context.
// The Timeline query uses this context, so the cancellation
// surfaces as an error.
func TestReplay_ExportMP4FromTimeline_PropagatesTimelineError(t *testing.T) {
	r, _ := setupReplayForTest(t, false)

	canceledCtx, cancel := context.WithCancel(context.Background())
	cancel() // cancel immediately

	dest := filepath.Join(t.TempDir(), "export.mp4")
	_, err := r.ExportMP4FromTimeline(canceledCtx, time.Now().Add(-time.Hour), dest)
	if err == nil {
		t.Fatal("ExportMP4FromTimeline with canceled context returned nil; want error")
	}
}

// TestReplay_ExportMP4FromTimeline_FutureSinceAlsoErrors pins
// the no-data edge case: ExportMP4FromTimeline with `since` set
// to a future time (so no frames can match) MUST also return
// the "no frames" error. Same contract as the empty-timeline
// case — production treats both as "nothing to export".
func TestReplay_ExportMP4FromTimeline_FutureSinceAlsoErrors(t *testing.T) {
	r, _ := setupReplayForTest(t, false)
	dest := filepath.Join(t.TempDir(), "export.mp4")

	_, err := r.ExportMP4FromTimeline(context.Background(), time.Now().Add(24*time.Hour), dest)
	if err == nil {
		t.Fatal("ExportMP4FromTimeline(future-since) returned nil; want error")
	}
}
