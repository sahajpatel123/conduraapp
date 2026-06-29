package audit

import (
	"context"
	"path/filepath"
	"testing"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/storage"
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
	return New(db.SQL(), db.MasterKey())
}

func TestLog_AppendAndList(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		err := l.Append(ctx, Event{
			Actor:   "user",
			Action:  "test.append",
			App:     "condurad",
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
	_ = l.Append(ctx, Event{Actor: "u", Action: "a", Level: "info"})
	_ = l.Append(ctx, Event{Actor: "u", Action: "b", Level: "warn"})
	_ = l.Append(ctx, Event{Actor: "u", Action: "c", Level: "error"})

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
	_ = l.Append(ctx, Event{Action: "old", Actor: "u", TS: time.Now().Add(-1 * time.Hour)})
	_ = l.Append(ctx, Event{Action: "new", Actor: "u", TS: time.Now()})

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
		_ = l.Append(ctx, Event{Action: "x", Actor: "u"})
	}
	n, err := l.Count(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if n != 5 {
		t.Fatalf("count = %d, want 5", n)
	}
}

// TestLog_ChainGenesis verifies the first row in the chain has the
// genesis prev_hash and a non-empty hmac.
func TestLog_ChainGenesis(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()
	if err := l.Append(ctx, Event{Actor: "user", Action: "first"}); err != nil {
		t.Fatal(err)
	}
	evs, err := l.List(ctx, Query{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 {
		t.Fatalf("got %d events, want 1", len(evs))
	}
	row, err := l.GetByID(ctx, evs[0].ID)
	if err != nil {
		t.Fatal(err)
	}
	if row.prevHash != genesisHash {
		t.Errorf("first row prev_hash = %q, want genesis", row.prevHash)
	}
	if len(row.hmac) != 64 {
		t.Errorf("first row hmac length = %d, want 64 hex chars", len(row.hmac))
	}
}

// TestLog_ChainLinks verifies each row's prev_hash matches the prior
// row's hmac, so the chain is contiguous.
func TestLog_ChainLinks(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()
	for i := 0; i < 5; i++ {
		if err := l.Append(ctx, Event{Actor: "user", Action: "x"}); err != nil {
			t.Fatal(err)
		}
	}
	rep, err := l.VerifyChain(ctx, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if !rep.Valid {
		t.Errorf("chain invalid: %s (row %d)", rep.FirstBreakReason, rep.FirstBreakID)
	}
	if rep.RowsChecked != 5 {
		t.Errorf("rows checked = %d, want 5", rep.RowsChecked)
	}
}

// TestLog_ChainDetectsTampering mutates one row and confirms the chain
// verifier catches it.
func TestLog_ChainDetectsTampering(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := l.Append(ctx, Event{Actor: "user", Action: "a"}); err != nil {
			t.Fatal(err)
		}
	}
	// Tamper with the message column of row 2 directly in the DB.
	if _, err := l.db.ExecContext(ctx, `UPDATE audit_log SET message = 'tampered' WHERE id = 2`); err != nil {
		t.Fatal(err)
	}
	rep, err := l.VerifyChain(ctx, 0, 0)
	if err != nil {
		t.Fatal(err)
	}
	if rep.Valid {
		t.Fatal("chain should NOT be valid after tampering")
	}
	if rep.FirstBreakID != 2 {
		t.Errorf("break at id = %d, want 2", rep.FirstBreakID)
	}
	if rep.RowsChecked != 2 {
		t.Errorf("rows checked = %d, want 2 (chain stops at first break)", rep.RowsChecked)
	}
}

// TestLog_StructuredFieldsRoundTrip confirms the Phase 11 fields
// survive an Append + List round trip.
func TestLog_StructuredFieldsRoundTrip(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()
	in := Event{
		Actor:         "gatekeeper",
		Action:        "consent.request",
		App:           "condura-gui",
		Level:         "info",
		Result:        "allow",
		Kind:          "consent.request",
		BlastClass:    "DESTRUCTIVE",
		Verdict:       "require_presence_and_consent",
		TargetApp:     "Safari",
		TargetURL:     "https://example.com",
		Path:          "/Users/syn/Downloads/report.pdf",
		Command:       "rm -rf /tmp/foo",
		ConsentResult: "approved",
		SSBeforeRef:   "replay/2026-06-11/abc-before.png",
		SSAfterRef:    "replay/2026-06-11/abc-after.png",
		SessionID:     "sess-001",
	}
	if err := l.Append(ctx, in); err != nil {
		t.Fatal(err)
	}
	evs, err := l.List(ctx, Query{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(evs) != 1 {
		t.Fatalf("got %d, want 1", len(evs))
	}
	got := evs[0]
	if got.Kind != in.Kind || got.BlastClass != in.BlastClass ||
		got.Verdict != in.Verdict || got.TargetApp != in.TargetApp ||
		got.TargetURL != in.TargetURL || got.Path != in.Path ||
		got.Command != in.Command || got.ConsentResult != in.ConsentResult ||
		got.SSBeforeRef != in.SSBeforeRef || got.SSAfterRef != in.SSAfterRef ||
		got.SessionID != in.SessionID {
		t.Errorf("structured fields lost in round trip:\n got: %+v\n want: %+v", got, in)
	}
}

// TestLog_HMAC_LengthPrefixing prevents delimiter collisions: two
// different events whose concatenated fields differ only in where the
// boundary lies must produce different HMACs.
func TestLog_HMAC_LengthPrefixing(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()
	// These two events would collide under a naive pipe-separated
	// canonicalization because "a|b" + "c" vs "a" + "b|c" serialize
	// identically. Length-prefixing must distinguish them.
	e1 := Event{Actor: "a|b", Action: "c", App: "x"}
	e2 := Event{Actor: "a", Action: "b|c", App: "x"}
	if err := l.Append(ctx, e1); err != nil {
		t.Fatal(err)
	}
	if err := l.Append(ctx, e2); err != nil {
		t.Fatal(err)
	}
	e1Stored, err := l.GetByID(ctx, 1)
	if err != nil {
		t.Fatal(err)
	}
	e2Stored, err := l.GetByID(ctx, 2)
	if err != nil {
		t.Fatal(err)
	}
	if e1Stored.hmac == e2Stored.hmac {
		t.Fatal("length-prefixing failed: HMACs collide for delimiter-shifted fields")
	}
}

// TestLog_AppendRequiresActorAndAction guards against the silent
// "Actor=” Action=”" footgun that the old code allowed.
func TestLog_AppendRequiresActorAndAction(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()
	if err := l.Append(ctx, Event{Action: "x"}); err == nil {
		t.Error("empty Actor should error")
	}
	if err := l.Append(ctx, Event{Actor: "x"}); err == nil {
		t.Error("empty Action should error")
	}
}
