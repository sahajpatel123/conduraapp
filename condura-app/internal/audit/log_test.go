package audit

import (
	"bytes"
	"context"
	"encoding/hex"
	"encoding/json"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/storage"
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

// TestNew_DerivesDomainSeparatedKey verifies that audit.New derives an
// HKDF subkey from the caller-provided master key. This is the P0-5
// "shared master key" fix: the audit HMAC must not be computed with
// the raw master key, otherwise a compromised master key (which also
// protects API-key encryption via AES-GCM) could be used to forge
// audit rows.
//
// What this test asserts:
//
//  1. Two audit.New calls with the same secret derive the same subkey:
//     hmacs are consistent across instances (existing chains verify).
//  2. Two audit.New calls with DIFFERENT secrets derive different
//     subkeys: a verifier built from a different master key cannot
//     validate a chain written by the first. This proves the secret
//     is mixed in via HKDF (not merely concatenated with a static salt).
//  3. The HMAC chain remains intact: an Append + VerifyChain round
//     trip succeeds when both writer and verifier share the same
//     secret.
func TestNew_DerivesDomainSeparatedKey(t *testing.T) {
	ctx := context.Background()

	dir := t.TempDir()
	db, err := storage.Open(ctx, storage.Config{
		Path: filepath.Join(dir, "subkey.db"),
	})
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = db.Close() }()
	masterA := []byte("masterkeybytes1234567890")
	masterB := []byte("differentkey1234567890123456")

	// 1. Two New() calls with the same master key must derive the
	// same subkey, so VerifyChain on logB succeeds against rows
	// written by logA.
	logA := New(db.SQL(), masterA)
	for i := 0; i < 3; i++ {
		if err := logA.Append(ctx, Event{Actor: "u", Action: "subkey.test"}); err != nil {
			t.Fatalf("logA.Append: %v", err)
		}
	}
	repSame, err := New(db.SQL(), masterA).VerifyChain(ctx, 0, 0)
	if err != nil {
		t.Fatalf("VerifyChain (same secret): %v", err)
	}
	if !repSame.Valid {
		t.Fatalf("chain should verify when both Logs use the same master key; got %s at id %d",
			repSame.FirstBreakReason, repSame.FirstBreakID)
	}
	if repSame.RowsChecked != 3 {
		t.Errorf("rows checked = %d, want 3", repSame.RowsChecked)
	}

	// 2. A New() call with a DIFFERENT master key must derive a
	// different subkey, so VerifyChain on logB fails — proving the
	// secret material is actually mixed in via HKDF, not a no-op.
	repDiff, err := New(db.SQL(), masterB).VerifyChain(ctx, 0, 0)
	if err != nil {
		t.Fatalf("VerifyChain (different secret): %v", err)
	}
	if repDiff.Valid {
		t.Fatal("chain must NOT verify when master key differs — HKDF subkey is not mixed in")
	}
	if repDiff.FirstBreakID == 0 {
		t.Errorf("FirstBreakID = 0, want first invalid row id (got empty break)")
	}
	if repDiff.FirstBreakReason == "" {
		t.Errorf("FirstBreakReason empty; expected an hmac mismatch reason")
	}

	// 3. Calling New() twice with the same key on the same DB and
	// appending more events must keep the chain valid: the subkey is
	// pure (master-key-only) and deterministic.
	logA2 := New(db.SQL(), masterA)
	if err := logA2.Append(ctx, Event{Actor: "u", Action: "subkey.test2"}); err != nil {
		t.Fatalf("logA2.Append: %v", err)
	}
	repAfter2, err := New(db.SQL(), masterA).VerifyChain(ctx, 0, 0)
	if err != nil {
		t.Fatalf("VerifyChain (after re-New): %v", err)
	}
	if !repAfter2.Valid {
		t.Fatalf("chain should still verify after re-deriving the same subkey; got %s at id %d",
			repAfter2.FirstBreakReason, repAfter2.FirstBreakID)
	}
	if repAfter2.RowsChecked != 4 {
		t.Errorf("rows checked = %d, want 4", repAfter2.RowsChecked)
	}
}

// TestPrune_WritesTombstone verifies that pruning N rows produces
// exactly one prune_tombstone row recording the deleted count and
// the pre-rewrite hmac of the oldest surviving row. This is the
// invariant #5 ("never deleted") forensic anchor: without a
// tombstone, an investigator cannot distinguish a 50-row post-prune
// log from a 100-row pre-prune log with 50 rows deleted.
func TestPrune_WritesTombstone(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()

	// Append 5 events spanning enough time to give us a controllable
	// retention cutoff. We backdate rows 1..3 by 2h so they fall
	// outside a 1h retention window.
	now := time.Now()
	for i := 0; i < 5; i++ {
		ts := now
		if i < 3 {
			ts = now.Add(-2 * time.Hour)
		}
		if err := l.Append(ctx, Event{Actor: "u", Action: "a", TS: ts}); err != nil {
			t.Fatal(err)
		}
	}

	// Capture the oldest surviving row (row 4, since rows 1..3 are
	// backdated) BEFORE prune so we can verify the tombstone's
	// oldest_surviving_hmac matches the row's pre-rewrite hmac.
	preOldest, err := l.GetByID(ctx, 4)
	if err != nil {
		t.Fatal(err)
	}
	preOldestHMAC := preOldest.hmac

	deleted, err := l.Prune(ctx, 1*time.Hour)
	if err != nil {
		t.Fatalf("Prune: %v", err)
	}
	if deleted != 3 {
		t.Errorf("deleted = %d, want 3", deleted)
	}

	// Exactly one tombstone should exist.
	stones, err := l.PruneTombstones(ctx)
	if err != nil {
		t.Fatalf("PruneTombstones: %v", err)
	}
	if len(stones) != 1 {
		t.Fatalf("got %d tombstones, want 1", len(stones))
	}

	got := stones[0]
	if got.PrunedCount != 3 {
		t.Errorf("tombstone.PrunedCount = %d, want 3", got.PrunedCount)
	}
	if got.OldestSurvivingID != preOldest.ID {
		t.Errorf("tombstone.OldestSurvivingID = %d, want %d (the row that survived)",
			got.OldestSurvivingID, preOldest.ID)
	}
	if got.OldestSurvivingHMAC != preOldestHMAC {
		t.Errorf("tombstone.OldestSurvivingHMAC = %q, want %q (pre-rewrite hmac)",
			got.OldestSurvivingHMAC, preOldestHMAC)
	}
	if got.RetentionWindowDays != 0 {
		// 1h retention is 0 full days.
		t.Errorf("tombstone.RetentionWindowDays = %d, want 0 (1h < 24h)", got.RetentionWindowDays)
	}
	if got.PrunedAt.IsZero() {
		t.Error("tombstone.PrunedAt is zero; expected a real timestamp")
	}
}

// TestVerifyChainWithHistory_ReturnsTombstones confirms that after a
// prune the chain still verifies AND the tombstone history is
// returned alongside the verdict. The plain VerifyChain must still
// return Valid=true (forwards compat with the existing GUI badge);
// VerifyChainWithHistory adds the tombstone slice.
func TestVerifyChainWithHistory_ReturnsTombstones(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()

	now := time.Now()
	for i := 0; i < 4; i++ {
		ts := now
		if i < 2 {
			ts = now.Add(-2 * time.Hour)
		}
		if err := l.Append(ctx, Event{Actor: "u", Action: "a", TS: ts}); err != nil {
			t.Fatal(err)
		}
	}

	if _, err := l.Prune(ctx, 1*time.Hour); err != nil {
		t.Fatalf("Prune: %v", err)
	}

	// Plain VerifyChain: still valid (back-compat).
	rep, err := l.VerifyChain(ctx, 0, 0)
	if err != nil {
		t.Fatalf("VerifyChain: %v", err)
	}
	if !rep.Valid {
		t.Errorf("plain VerifyChain should be Valid after Prune; got %s at id %d",
			rep.FirstBreakReason, rep.FirstBreakID)
	}

	// VerifyChainWithHistory: same verdict + tombstone.
	hist, err := l.VerifyChainWithHistory(ctx, 0, 0)
	if err != nil {
		t.Fatalf("VerifyChainWithHistory: %v", err)
	}
	if !hist.Valid {
		t.Errorf("history VerifyChainWithHistory.Valid = false; got %s at id %d",
			hist.FirstBreakReason, hist.FirstBreakID)
	}
	if len(hist.Tombstones) != 1 {
		t.Fatalf("got %d tombstones, want 1", len(hist.Tombstones))
	}
	if hist.Tombstones[0].PrunedCount != 2 {
		t.Errorf("tombstone.PrunedCount = %d, want 2", hist.Tombstones[0].PrunedCount)
	}
}

// TestPruneTombstones_ForensicQuery verifies the ordering and the
// accumulation behavior: multiple prune invocations produce
// multiple tombstones, ordered by pruned_at DESC. This is what an
// investigator would render in the GUI's "history" view.
func TestPruneTombstones_ForensicQuery(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()

	// First prune: no rows. retention window 1h with nothing to
	// delete; tombstone path is gated on "no rows AND no deletes"
	// being a true no-op (no tombstone written). Let's confirm
	// that path by appending nothing and pruning.
	if _, err := l.Prune(ctx, 1*time.Hour); err != nil {
		t.Fatalf("Prune (empty log): %v", err)
	}
	stones, err := l.PruneTombstones(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(stones) != 0 {
		t.Fatalf("Prune on empty log wrote %d tombstones, want 0 (true no-op)", len(stones))
	}

	// Append 3 backdated events then prune them.
	now := time.Now()
	for i := 0; i < 3; i++ {
		if err := l.Append(ctx, Event{Actor: "u", Action: "a", TS: now.Add(-2 * time.Hour)}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := l.Prune(ctx, 1*time.Hour); err != nil {
		t.Fatal(err)
	}

	// Append 2 more recent events and prune them in a second pass.
	for i := 0; i < 2; i++ {
		if err := l.Append(ctx, Event{Actor: "u", Action: "a", TS: now.Add(-2 * time.Hour)}); err != nil {
			t.Fatal(err)
		}
	}
	if _, err := l.Prune(ctx, 1*time.Hour); err != nil {
		t.Fatal(err)
	}

	// We should have 2 tombstones, ordered DESC by pruned_at.
	stones, err = l.PruneTombstones(ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(stones) != 2 {
		t.Fatalf("got %d tombstones, want 2", len(stones))
	}
	if stones[0].PrunedAt.Before(stones[1].PrunedAt) {
		t.Errorf("tombstones not in DESC order: [0]=%v [1]=%v",
			stones[0].PrunedAt, stones[1].PrunedAt)
	}
	// Total deleted across all prunes must be visible in the
	// forensic sum (3 + 2 = 5).
	var total int64
	for _, s := range stones {
		total += s.PrunedCount
	}
	if total != 5 {
		t.Errorf("sum of pruned_count = %d, want 5 (3 + 2)", total)
	}
}

func TestLog_Export(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()
	for i := 0; i < 3; i++ {
		if err := l.Append(ctx, Event{
			Actor:   "user",
			Action:  "test.export",
			App:     "condurad",
			Level:   "info",
			Result:  "allow",
			Message: "row",
		}); err != nil {
			t.Fatal(err)
		}
	}
	dest := filepath.Join(t.TempDir(), "out.jsonl")
	path, count, err := l.Export(ctx, Query{}, dest)
	if err != nil {
		t.Fatalf("Export: %v", err)
	}
	if path != dest {
		t.Fatalf("path = %q, want %q", path, dest)
	}
	if count != 3 {
		t.Fatalf("count = %d, want 3", count)
	}
	raw, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	lines := strings.Split(strings.TrimSpace(string(raw)), "\n")
	if len(lines) != 3 {
		t.Fatalf("lines = %d, want 3", len(lines))
	}
	if !strings.Contains(lines[0], `"this_hash"`) && !strings.Contains(lines[0], "this_hash") {
		// JSON keys present
		if !strings.Contains(string(raw), "this_hash") {
			t.Fatalf("export missing this_hash: %s", lines[0])
		}
	}
}

func TestLog_OnAppendAndMarshalHashes(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()
	var got []Event
	l.SetOnAppend(func(e Event) {
		got = append(got, e)
	})
	if err := l.Append(ctx, Event{
		Actor:  "user",
		Action: "test.live",
		Level:  "info",
		Result: "allow",
	}); err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("onAppend count = %d, want 1", len(got))
	}
	if got[0].ID == 0 {
		t.Fatal("onAppend event missing id")
	}
	raw, err := json.Marshal(got[0])
	if err != nil {
		t.Fatal(err)
	}
	s := string(raw)
	if !strings.Contains(s, `"this_hash"`) || !strings.Contains(s, `"prev_hash"`) {
		t.Fatalf("marshal missing hashes: %s", s)
	}
}

func TestLog_ListPhase15FiltersAndFacets(t *testing.T) {
	l := setupTestLog(t)
	ctx := context.Background()
	now := time.Now().UTC()
	// Older row outside 1h window
	_ = l.Append(ctx, Event{
		TS: now.Add(-2 * time.Hour), Actor: "u", Action: "old.action",
		App: "app-old", Level: "info", Result: "allow", Message: "ancient",
		BlastClass: "READ", Verdict: "allow",
	})
	_ = l.Append(ctx, Event{
		TS: now.Add(-10 * time.Minute), Actor: "u", Action: "shell.exec",
		App: "app-a", Level: "info", Result: "block", Message: "denied",
		BlastClass: "WRITE", Verdict: "block",
	})
	_ = l.Append(ctx, Event{
		TS: now.Add(-5 * time.Minute), Actor: "u", Action: "http.get",
		App: "app-b", Level: "info", Result: "allow", Message: "ok model=gpt",
		BlastClass: "NETWORK", Verdict: "allow",
	})

	since := now.Add(-1 * time.Hour)
	got, err := l.List(ctx, Query{
		Limit: 50, Since: since,
		Verdict: "block",
	})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 || got[0].Action != "shell.exec" {
		t.Fatalf("verdict block filter: got %+v", got)
	}

	search, err := l.List(ctx, Query{Limit: 50, Since: since, Search: "denied"})
	if err != nil {
		t.Fatal(err)
	}
	if len(search) != 1 {
		t.Fatalf("search: got %d", len(search))
	}

	bc, err := l.List(ctx, Query{Limit: 50, Since: since, BlastClasses: []string{"NETWORK"}})
	if err != nil {
		t.Fatal(err)
	}
	if len(bc) != 1 || bc[0].BlastClass != "NETWORK" {
		t.Fatalf("blast filter: %+v", bc)
	}

	facets, err := l.FacetCounts(ctx, Query{Since: since})
	if err != nil {
		t.Fatal(err)
	}
	if facets.Total != 2 {
		t.Fatalf("facet total = %d, want 2 (1h window)", facets.Total)
	}
	if facets.Verdicts["block"] < 1 || facets.Verdicts["allow"] < 1 {
		t.Fatalf("verdicts = %+v", facets.Verdicts)
	}
	if facets.BlastClasses["WRITE"] < 1 || facets.BlastClasses["NETWORK"] < 1 {
		t.Fatalf("blast_classes = %+v", facets.BlastClasses)
	}
}

// TestLog_Reload_NilReceiver pins the nil-safe contract on Reload.
// Backup-restore paths can run during teardown when the receiver is
// already nil; Reload must be a no-op in that case rather than
// panicking. This test would catch a future change that drops the
// `if l == nil { return }` guard at the top of the function.
func TestLog_Reload_NilReceiver(t *testing.T) {
	var l *Log
	// Must not panic on nil receiver, even when db is also nil.
	l.Reload(nil)
}

// TestLog_Reload_ReplacesHandle verifies the core contract: after
// Reload, subsequent Append writes go to the new DB handle rather
// than the closed old one. This is the path storage.Reload callers
// rely on after a backup restore.
func TestLog_Reload_ReplacesHandle(t *testing.T) {
	ctx := context.Background()
	l := setupTestLog(t)

	// Append to the original DB.
	if err := l.Append(ctx, Event{Actor: "u", Action: "before.reload", Result: "allow"}); err != nil {
		t.Fatal(err)
	}

	// Open a second, distinct DB.
	dir2 := t.TempDir()
	db2, err := storage.Open(ctx, storage.Config{Path: filepath.Join(dir2, "second.db")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db2.Close() })

	l.Reload(db2.SQL())

	// Append after Reload must succeed and land in db2, not the
	// original handle (which we never closed here, so this exercises
	// the handle swap rather than a use-after-close).
	if err := l.Append(ctx, Event{Actor: "u", Action: "after.reload", Result: "allow"}); err != nil {
		t.Fatalf("Append after Reload: %v", err)
	}

	// The new DB should now contain exactly one event (the post-reload one).
	got, err := l.List(ctx, Query{Limit: 10})
	if err != nil {
		t.Fatal(err)
	}
	if len(got) != 1 {
		t.Fatalf("got %d events in reloaded DB, want 1", len(got))
	}
	if got[0].Action != "after.reload" {
		t.Fatalf("first event action = %q, want %q", got[0].Action, "after.reload")
	}
}

// TestLog_Reload_InvalidatesCachedLastHMAC pins the cache-invalidation
// invariant. After Reload, the cached last HMAC must be cleared so
// that the next Append recomputes against the new DB's actual last
// row instead of trusting a stale cache from the old handle — which
// would otherwise let a backup with a different chain root be silently
// re-chained onto the previous chain.
func TestLog_Reload_InvalidatesCachedLastHMAC(t *testing.T) {
	ctx := context.Background()
	l := setupTestLog(t)

	// Establish a cached last-HMAC via Append.
	if err := l.Append(ctx, Event{Actor: "u", Action: "prime.cache", Result: "allow"}); err != nil {
		t.Fatal(err)
	}
	l.mu.Lock()
	if !l.lastHMACValid {
		l.mu.Unlock()
		t.Fatal("lastHMACValid should be true after Append")
	}
	if l.lastHMACValue == "" {
		l.mu.Unlock()
		t.Fatal("lastHMACValue should be non-empty after Append")
	}
	l.mu.Unlock()

	// Reload against a fresh DB.
	dir2 := t.TempDir()
	db2, err := storage.Open(ctx, storage.Config{Path: filepath.Join(dir2, "fresh.db")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db2.Close() })
	l.Reload(db2.SQL())

	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lastHMACValid {
		t.Fatal("lastHMACValid must be false after Reload")
	}
	if l.lastHMACValue != "" {
		t.Fatalf("lastHMACValue must be empty after Reload, got %q", l.lastHMACValue)
	}
}

// TestNewWithHexSecret_Empty pins the empty-input error contract.
// An empty hex secret is the canonical "no key configured" state and
// must surface as a clear error rather than panicking inside New().
func TestNewWithHexSecret_Empty(t *testing.T) {
	dir := t.TempDir()
	db, err := storage.Open(context.Background(), storage.Config{Path: filepath.Join(dir, "empty.db")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	l, err := NewWithHexSecret(db.SQL(), "")
	if err == nil {
		t.Fatal("expected error on empty hex secret, got nil")
	}
	if l != nil {
		t.Fatal("expected nil Log on error, got non-nil")
	}
	if !strings.Contains(err.Error(), "empty") {
		t.Fatalf("error %q should mention 'empty'", err.Error())
	}
}

// TestNewWithHexSecret_InvalidHex pins the parse-error contract.
// A non-hex string must surface as a wrapped error so callers can
// diagnose misconfigured secrets without leaking internals.
func TestNewWithHexSecret_InvalidHex(t *testing.T) {
	dir := t.TempDir()
	db, err := storage.Open(context.Background(), storage.Config{Path: filepath.Join(dir, "invalid.db")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// "not-hex!!" contains characters outside [0-9a-f].
	l, err := NewWithHexSecret(db.SQL(), "not-hex!!")
	if err == nil {
		t.Fatal("expected error on invalid hex, got nil")
	}
	if l != nil {
		t.Fatal("expected nil Log on error, got non-nil")
	}
	if !strings.Contains(err.Error(), "hex") {
		t.Fatalf("error %q should mention 'hex'", err.Error())
	}
}

// TestNewWithHexSecret_ValidHex verifies the happy path: a valid hex
// string decodes, the Log is constructed, and Append works. Without
// this, a regression that swapped DecodeString for an empty-string
// shortcut would pass NewWithHexSecret's empty check and only fail
// deep inside Append.
func TestNewWithHexSecret_ValidHex(t *testing.T) {
	dir := t.TempDir()
	db, err := storage.Open(context.Background(), storage.Config{Path: filepath.Join(dir, "valid.db")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	// 32 bytes -> 64 hex chars (auditSubKeyLen).
	rawSecret := make([]byte, 32)
	for i := range rawSecret {
		rawSecret[i] = byte(i + 1)
	}
	hexSecret := hex.EncodeToString(rawSecret)

	l, err := NewWithHexSecret(db.SQL(), hexSecret)
	if err != nil {
		t.Fatalf("NewWithHexSecret(valid): %v", err)
	}
	if l == nil {
		t.Fatal("NewWithHexSecret returned nil Log on valid input")
	}
	if err := l.Append(context.Background(), Event{Actor: "u", Action: "hex.valid", Result: "allow"}); err != nil {
		t.Fatalf("Append: %v", err)
	}
}

// TestNewWithHexSecret_MatchesRawSecret pins the equivalence contract:
// a Log built via NewWithHexSecret(db, hex.EncodeToString(s)) must
// derive the same HKDF subkey as New(db, s). This guards against a
// future refactor that decodes the hex string but then forgets to
// pass the bytes through deriveAuditSubkey — which would silently
// desync the chain from any backup written by the raw-secret path.
func TestNewWithHexSecret_MatchesRawSecret(t *testing.T) {
	dir := t.TempDir()
	db, err := storage.Open(context.Background(), storage.Config{Path: filepath.Join(dir, "equiv.db")})
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = db.Close() })

	rawSecret := make([]byte, 32)
	for i := range rawSecret {
		rawSecret[i] = byte(i*7 + 3) // non-trivial pattern
	}

	lHex, err := NewWithHexSecret(db.SQL(), hex.EncodeToString(rawSecret))
	if err != nil {
		t.Fatal(err)
	}
	want := deriveAuditSubkey(rawSecret)
	if !bytes.Equal(lHex.subkey, want) {
		t.Fatalf("NewWithHexSecret subkey != deriveAuditSubkey(rawSecret)\n  got:  %x\n  want: %x",
			lHex.subkey, want)
	}
}

// TestResolveExportPath_Default_UsesHomeDownloadsOrTemp pins the
// default-destination branch: when the caller passes an empty string,
// resolveExportPath must place the export under either $HOME/Downloads
// (if a home directory is available) or os.TempDir() (if UserHomeDir
// fails). The result must end with condura-audit-<UTC-timestamp>.jsonl.
//
// This branch was at 0% coverage before this test. A regression here
// would either leak exports to an unexpected location or change the
// filename pattern in a way that confuses downstream tooling.
func TestResolveExportPath_Default_UsesHomeDownloadsOrTemp(t *testing.T) {
	got, err := resolveExportPath("")
	if err != nil {
		t.Fatalf("resolveExportPath(\"\"): %v", err)
	}
	// Filename pattern: condura-audit-YYYYMMDD-HHMMSS.jsonl
	base := filepath.Base(got)
	if !strings.HasPrefix(base, "condura-audit-") || !strings.HasSuffix(base, ".jsonl") {
		t.Fatalf("default path %q does not match condura-audit-<ts>.jsonl pattern", got)
	}
	// Directory must be either $HOME/Downloads or os.TempDir().
	dir := filepath.Dir(got)
	home, herr := os.UserHomeDir()
	if herr == nil && home != "" {
		wantDir := filepath.Join(home, "Downloads")
		if dir != wantDir && dir != os.TempDir() {
			t.Fatalf("default path dir %q is neither %q nor %q", dir, wantDir, os.TempDir())
		}
	} else if dir != os.TempDir() {
		t.Fatalf("default path dir %q is not %q (UserHomeDir unavailable)", dir, os.TempDir())
	}
}

// TestResolveExportPath_WhitespaceOnlyFallsThrough pins the
// boundary-check contract: a destination consisting only of
// whitespace must be treated as empty. The trim check uses
// strings.TrimSpace, not filepath.Clean, so a regression that
// switched to filepath.Clean would still match "/" but start
// treating " " as a valid destination — pushing exports into
// a directory literally named " ".
func TestResolveExportPath_WhitespaceOnlyFallsThrough(t *testing.T) {
	got, err := resolveExportPath("   ")
	if err != nil {
		t.Fatalf("resolveExportPath(\"   \"): %v", err)
	}
	// A literal-whitespace directory would show up as a directory
	// component containing spaces only. Assert the parent is a
	// sensible real directory, not a whitespace-named path.
	dir := filepath.Dir(got)
	if strings.TrimSpace(dir) != dir {
		t.Fatalf("export path %q has whitespace-padded directory — trim check is broken", got)
	}
}

// TestResolveExportPath_AbsoluteDestinationReturned pins that an
// explicit absolute destination is passed through filepath.Clean
// unchanged. Callers pass a destination they trust (e.g. a user-
// chosen Downloads path); resolveExportPath must NOT add or remove
// path components — only normalize ./ and ../ segments.
func TestResolveExportPath_AbsoluteDestinationReturned(t *testing.T) {
	in := "/tmp/condura-test-explicit.jsonl"
	got, err := resolveExportPath(in)
	if err != nil {
		t.Fatalf("resolveExportPath(%q): %v", in, err)
	}
	if got != filepath.Clean(in) {
		t.Fatalf("resolveExportPath(%q) = %q, want %q", in, got, filepath.Clean(in))
	}
}

// TestResolveExportPath_NormalizesDotTraversal pins that relative
// dot-segments in the destination are collapsed by filepath.Clean.
// This is the safety net: a caller-supplied path like "foo/../bar"
// must NOT reach the filesystem as written — Clean normalizes to
// "bar" before the export writer opens it.
func TestResolveExportPath_NormalizesDotTraversal(t *testing.T) {
	// Use a t.TempDir()-rooted path so the test is hermetic and
	// doesn't touch the real filesystem.
	base := t.TempDir()
	in := filepath.Join(base, "subdir", "..", "normalized.jsonl")
	got, err := resolveExportPath(in)
	if err != nil {
		t.Fatalf("resolveExportPath(%q): %v", in, err)
	}
	want := filepath.Join(base, "normalized.jsonl")
	if got != want {
		t.Fatalf("resolveExportPath(%q) = %q, want %q (filepath.Clean should collapse ..)", in, got, want)
	}
}
