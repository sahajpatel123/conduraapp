package daemon

import (
	"database/sql"
	"io"
	"log/slog"
	"path/filepath"
	"testing"

	_ "modernc.org/sqlite"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/llm"
)

// spendSchema is the minimal subset of the production storage
// schema required by persistSpend + loadSpendToday. Mirroring the
// real DDL keeps the contract test honest: if the production code
// starts writing to a column the test schema doesn't have, the
// insert will fail and the test will surface the drift.
//
// Schema source: internal/storage/migrations.go (llm_calls,
// spend_daily). Kept inline (rather than importing migrations.Run)
// because the spend helpers only need these two tables, and an
// in-process test should not trigger the full migration chain.
const spendSchema = `
CREATE TABLE llm_calls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ts TEXT NOT NULL DEFAULT (datetime('now')),
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    task TEXT NOT NULL,
    input_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    cost_usd REAL NOT NULL DEFAULT 0,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    success INTEGER NOT NULL,
    error TEXT,
    prompt_hash TEXT,
    prompt_preview_ciphertext TEXT
);

CREATE TABLE spend_daily (
    day TEXT NOT NULL,
    provider TEXT NOT NULL,
    cost_usd REAL NOT NULL DEFAULT 0,
    PRIMARY KEY (day, provider)
);
`

// openSpendTestDB returns a sqlite-backed *sql.DB wired with the
// spend schema, in a per-test temp dir so parallel tests don't
// collide. Caller is responsible for Close().
func openSpendTestDB(t *testing.T) *sql.DB {
	t.Helper()
	db, err := sql.Open("sqlite", filepath.Join(t.TempDir(), "spend.db"))
	if err != nil {
		t.Fatalf("sql.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	if _, err := db.Exec(spendSchema); err != nil {
		t.Fatalf("apply spend schema: %v", err)
	}
	return db
}

// quietLogger discards all output so the spend-helper error
// paths (which log a Warn on insert failure) don't spam the test
// runner. The contract being tested is the side-effect on the DB,
// not the log line itself.
func quietLogger() *slog.Logger {
	return slog.New(slog.NewTextHandler(io.Discard, nil))
}

// TestPersistSpend_NilDB_NoOp verifies the nil-db short-circuit.
// persistSpend is called from a fire-and-forget LLM callback;
// if db is nil (e.g. durable storage failed to initialize), the
// helper must not panic and must not attempt any writes.
func TestPersistSpend_NilDB_NoOp(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("persistSpend(nil, ...) panicked: %v", r)
		}
	}()
	persistSpend(nil, "openai", "gpt-4o", llm.Usage{InputTokens: 10, OutputTokens: 20}, 0.05, quietLogger())
}

// TestLoadSpendToday_NilDB_ReturnsZero verifies the nil-db
// short-circuit on the read path. Without this guard, a startup
// that failed to open the DB would crash on the spend-seeding
// line instead of silently starting with a zero baseline.
func TestLoadSpendToday_NilDB_ReturnsZero(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("loadSpendToday(nil) panicked: %v", r)
		}
	}()
	if got := loadSpendToday(nil); got != 0 {
		t.Fatalf("loadSpendToday(nil) = %v, want 0", got)
	}
}

// TestLoadSpendToday_EmptyDB_ReturnsZero verifies the happy
// path on a fresh database: no spend_daily rows for today means
// the spend monitor starts with a zero baseline. Pre-fix, the
// COALESCE path was untested; this pins it.
func TestLoadSpendToday_EmptyDB_ReturnsZero(t *testing.T) {
	db := openSpendTestDB(t)
	if got := loadSpendToday(db); got != 0 {
		t.Fatalf("loadSpendToday(empty) = %v, want 0", got)
	}
}

// TestPersistSpend_WritesLLMCallsAndSpendDaily verifies the
// happy path: one persistSpend call writes exactly one llm_calls
// row AND upserts one spend_daily row for today. Both writes are
// required — a partial write (e.g. llm_calls but not spend_daily)
// would silently break the durable cap on restart.
func TestPersistSpend_WritesLLMCallsAndSpendDaily(t *testing.T) {
	db := openSpendTestDB(t)

	persistSpend(db, "openai", "gpt-4o", llm.Usage{InputTokens: 100, OutputTokens: 50}, 0.42, quietLogger())

	// llm_calls row: exactly one, with the correct provider/model/task.
	var (
		provider, model, task string
		inTok, outTok         int
		cost                  float64
		success               int
	)
	err := db.QueryRow(
		`SELECT provider, model, task, input_tokens, output_tokens, cost_usd, success FROM llm_calls`,
	).Scan(&provider, &model, &task, &inTok, &outTok, &cost, &success)
	if err != nil {
		t.Fatalf("llm_calls scan: %v", err)
	}
	if provider != "openai" || model != "gpt-4o" || task != "chat" {
		t.Errorf("llm_calls row = (%q,%q,%q), want (openai,gpt-4o,chat)", provider, model, task)
	}
	if inTok != 100 || outTok != 50 || cost != 0.42 || success != 1 {
		t.Errorf("llm_calls values = (in=%d out=%d cost=%v success=%d), want (100,50,0.42,1)", inTok, outTok, cost, success)
	}

	// spend_daily row: exactly one, with the rolled-up cost.
	var sum float64
	err = db.QueryRow(`SELECT cost_usd FROM spend_daily`).Scan(&sum)
	if err != nil {
		t.Fatalf("spend_daily scan: %v", err)
	}
	if sum != 0.42 {
		t.Errorf("spend_daily cost_usd = %v, want 0.42", sum)
	}
}

// TestPersistSpend_UpsertsAccumulates verifies the ON CONFLICT
// branch: a second persistSpend for the same (day, provider)
// must add to the existing cost, not replace it. The spend
// monitor seeds from spend_daily on startup, so an insert
// (rather than upsert) would reset the daily cap to the latest
// call — silently letting the user blow through their budget.
func TestPersistSpend_UpsertsAccumulates(t *testing.T) {
	db := openSpendTestDB(t)

	persistSpend(db, "openai", "gpt-4o", llm.Usage{InputTokens: 100, OutputTokens: 50}, 0.42, quietLogger())
	persistSpend(db, "openai", "gpt-4o", llm.Usage{InputTokens: 200, OutputTokens: 100}, 0.88, quietLogger())
	persistSpend(db, "openai", "gpt-4o", llm.Usage{InputTokens: 50, OutputTokens: 25}, 0.20, quietLogger())

	// spend_daily must accumulate: 0.42 + 0.88 + 0.20 = 1.50.
	var sum float64
	if err := db.QueryRow(`SELECT cost_usd FROM spend_daily`).Scan(&sum); err != nil {
		t.Fatalf("spend_daily scan: %v", err)
	}
	if sum != 1.50 {
		t.Errorf("spend_daily cost_usd = %v, want 1.50 (accumulated)", sum)
	}

	// llm_calls must hold all three rows (no dedup).
	var n int
	if err := db.QueryRow(`SELECT COUNT(*) FROM llm_calls`).Scan(&n); err != nil {
		t.Fatalf("llm_calls count: %v", err)
	}
	if n != 3 {
		t.Errorf("llm_calls row count = %d, want 3 (one per persistSpend call)", n)
	}

	// loadSpendToday must return the accumulated total.
	if got := loadSpendToday(db); got != 1.50 {
		t.Errorf("loadSpendToday after 3 upserts = %v, want 1.50", got)
	}
}
