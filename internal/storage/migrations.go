package storage

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// migrations is the ordered list of schema migrations.
// Each entry's Version is the schema_version that becomes active after that
// migration runs. A migration is a transaction containing one or more SQL
// statements.
type migration struct {
	Version int
	Name    string
	SQL     string
}

var migrations = []migration{
	{
		Version: 1,
		Name:    "initial schema",
		SQL: `
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at TEXT NOT NULL
);

-- API keys: provider + encrypted secret material.
CREATE TABLE IF NOT EXISTS api_keys (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    provider TEXT NOT NULL,
    label TEXT,
    auth_kind TEXT NOT NULL CHECK (auth_kind IN ('api_key', 'oauth')),
    -- Encrypted with the master key. See DB.EncryptString.
    secret_ciphertext TEXT NOT NULL,
    -- OAuth-specific metadata (encrypted; may be NULL for API keys).
    refresh_token_ciphertext TEXT,
    scopes TEXT,
    -- Free-form JSON metadata (e.g. account email, expiry). NOT encrypted
    -- unless the user opts in. Phase 1: keep unencrypted.
    metadata_json TEXT,
    expires_at TEXT,
    last_used_at TEXT,
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    UNIQUE(provider, label)
);

CREATE INDEX IF NOT EXISTS idx_api_keys_provider ON api_keys(provider);

-- LLM call ledger (for cost tracking, failover, audit).
CREATE TABLE IF NOT EXISTS llm_calls (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ts TEXT NOT NULL DEFAULT (datetime('now')),
    provider TEXT NOT NULL,
    model TEXT NOT NULL,
    task TEXT NOT NULL, -- 'chat' | 'embedding' | 'vision' | 'tool'
    input_tokens INTEGER NOT NULL DEFAULT 0,
    output_tokens INTEGER NOT NULL DEFAULT 0,
    cost_usd REAL NOT NULL DEFAULT 0,
    latency_ms INTEGER NOT NULL DEFAULT 0,
    success INTEGER NOT NULL,
    error TEXT,
    -- The full request/response is NOT stored. We log a hash + first 200 chars
    -- of the user prompt for audit, encrypted with the master key.
    prompt_hash TEXT,
    prompt_preview_ciphertext TEXT
);

CREATE INDEX IF NOT EXISTS idx_llm_calls_ts ON llm_calls(ts);
CREATE INDEX IF NOT EXISTS idx_llm_calls_provider ON llm_calls(provider);

-- Per-day spend rollup (updated by triggers in v1; computed by query in Phase 1).
CREATE TABLE IF NOT EXISTS spend_daily (
    day TEXT NOT NULL, -- YYYY-MM-DD
    provider TEXT NOT NULL,
    cost_usd REAL NOT NULL DEFAULT 0,
    PRIMARY KEY (day, provider)
);

-- Audit log of privileged operations.
CREATE TABLE IF NOT EXISTS audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ts TEXT NOT NULL DEFAULT (datetime('now')),
    actor TEXT NOT NULL, -- 'user' | 'daemon' | 'cli' | provider name
    action TEXT NOT NULL, -- 'config.change' | 'secrets.set' | 'llm.failover' | etc.
    target TEXT,
    -- Free-form JSON. Sensitive fields are encrypted at the application layer.
    details_ciphertext TEXT,
    details_json TEXT
);

CREATE INDEX IF NOT EXISTS idx_audit_log_ts ON audit_log(ts);
CREATE INDEX IF NOT EXISTS idx_audit_log_action ON audit_log(action);

-- Provider health snapshot (last seen error, latency, last success).
CREATE TABLE IF NOT EXISTS provider_health (
    provider TEXT PRIMARY KEY,
    last_success_at TEXT,
    last_failure_at TEXT,
    consecutive_failures INTEGER NOT NULL DEFAULT 0,
    circuit_state TEXT NOT NULL DEFAULT 'closed'
        CHECK (circuit_state IN ('closed', 'open', 'half_open')),
    last_error TEXT
);

-- Memory entries (Phase 4 — schema only here).
CREATE TABLE IF NOT EXISTS memory_entries (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    scope TEXT NOT NULL CHECK (scope IN ('user', 'project', 'session', 'ephemeral')),
    key TEXT NOT NULL,
    value_ciphertext TEXT NOT NULL,
    source TEXT,
    embedding BLOB, -- future: vector embedding
    created_at TEXT NOT NULL DEFAULT (datetime('now')),
    updated_at TEXT NOT NULL DEFAULT (datetime('now')),
    expires_at TEXT,
    UNIQUE(scope, key)
);

CREATE INDEX IF NOT EXISTS idx_memory_entries_scope ON memory_entries(scope);
CREATE INDEX IF NOT EXISTS idx_memory_entries_expires ON memory_entries(expires_at);
`,
	},
	{
		Version: 2,
		Name:    "conversations + audit + halt + telemetry + first-run + window state",
		SQL: `
-- Conversations: one row per chat thread. Phase 2 stores the
-- current conversation only (per the locked-in decision) but the
-- schema is general enough for full history if we want it later.
CREATE TABLE IF NOT EXISTS conversations (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    title TEXT NOT NULL DEFAULT 'New conversation',
    created_at TEXT NOT NULL,
    updated_at TEXT NOT NULL
);

-- Conversation messages: the per-message payload. We store the
-- raw JSON (role + content + tool_calls) as TEXT to keep this
-- migration simple. Future versions can split this into columns
-- or move to a separate messages table keyed by conversation_id.
CREATE TABLE IF NOT EXISTS conversation_messages (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    conversation_id INTEGER NOT NULL,
    role TEXT NOT NULL,
    content TEXT NOT NULL,
    tool_calls_json TEXT,
    tool_call_id TEXT,
    created_at TEXT NOT NULL,
    FOREIGN KEY (conversation_id) REFERENCES conversations(id) ON DELETE CASCADE
);
CREATE INDEX IF NOT EXISTS idx_conv_msgs ON conversation_messages(conversation_id, id);

-- Audit log: append-only. One row per auditable action.
-- Phase 2 replaces the v1 audit_log schema (which had a
-- different column set) with the new one. Production hasn't
-- shipped yet, so this is a clean cut.
DROP TABLE IF EXISTS audit_log;
CREATE TABLE audit_log (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    ts TEXT NOT NULL,
    actor TEXT NOT NULL,
    action TEXT NOT NULL,
    app TEXT NOT NULL DEFAULT '',
    level TEXT NOT NULL DEFAULT 'info',
    result TEXT NOT NULL DEFAULT 'allow',
    message TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_audit_ts ON audit_log(ts DESC);
CREATE INDEX IF NOT EXISTS idx_audit_action ON audit_log(action);

-- Halt flag: a single-row table that all subsystems check before
-- performing work. Updated by daemon.halt / daemon.resume.
CREATE TABLE IF NOT EXISTS halt_state (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    halted INTEGER NOT NULL DEFAULT 0,
    since TEXT,
    reason TEXT
);
INSERT OR IGNORE INTO halt_state (id, halted) VALUES (1, 0);

-- First-run marker: 0 if wizard has not been completed, 1 if it has.
CREATE TABLE IF NOT EXISTS first_run (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    complete INTEGER NOT NULL DEFAULT 0,
    completed_at TEXT
);
INSERT OR IGNORE INTO first_run (id, complete) VALUES (1, 0);

-- Window state: persisted GUI window position/size + last
-- conversation. Read on app launch, written on resize/move.
CREATE TABLE IF NOT EXISTS window_state (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    width INTEGER NOT NULL DEFAULT 1200,
    height INTEGER NOT NULL DEFAULT 800,
    x INTEGER,
    y INTEGER,
    last_conversation_id INTEGER DEFAULT 0
);
INSERT OR IGNORE INTO window_state (id, width, height, x, y, last_conversation_id) VALUES (1, 1200, 800, NULL, NULL, 0);

-- Telemetry counters: anonymous usage counters. Aggregated on
-- disk; flushed to the (opt-in) endpoint periodically.
CREATE TABLE IF NOT EXISTS telemetry_counters (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    enabled INTEGER NOT NULL DEFAULT 0,
    session_starts INTEGER NOT NULL DEFAULT 0,
    messages_sent INTEGER NOT NULL DEFAULT 0,
    tools_called INTEGER NOT NULL DEFAULT 0,
    errors INTEGER NOT NULL DEFAULT 0,
    last_flush_ts TEXT
);
INSERT OR IGNORE INTO telemetry_counters (id, enabled) VALUES (1, 0);

-- Update manifest cache: stores the most recent update-check
-- result so the GUI can show "update available" without making
-- a network call on every launch.
CREATE TABLE IF NOT EXISTS update_cache (
    id INTEGER PRIMARY KEY CHECK (id = 1),
    last_check_ts TEXT,
    latest_version TEXT,
    download_url TEXT
);
INSERT OR IGNORE INTO update_cache (id) VALUES (1);
`,
	},
	{
		// Phase 11 (Trust & Recovery): Action Replay.
		// 1. HMAC chain on audit_log so tampering is detectable (MISSION §5.4).
		// 2. Structured fields on audit_log so Replay reconstructs the
		//    timeline from real data, not string-parsing the Message column.
		// Additive, backfill NULL: existing rows remain valid; new rows
		// carry the structured payload. The first row to use the chain
		// (id=1) has prev_hash=0; every subsequent row's prev_hash
		// matches the prior row's hmac.
		Version: 3,
		Name:    "audit log HMAC chain + structured fields",
		SQL: `
-- HMAC chain: prev_hash is the hex SHA-256 of the prior row's HMAC
-- (or 64 zeros for the first row). hmac is the hex SHA-256 of the
-- canonical serialization of this row (excluding the hmac column).
ALTER TABLE audit_log ADD COLUMN prev_hash TEXT NOT NULL DEFAULT '';
ALTER TABLE audit_log ADD COLUMN hmac TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_audit_hmac ON audit_log(id, hmac);

-- Structured fields for Action Replay. All NULL-safe.
ALTER TABLE audit_log ADD COLUMN kind TEXT NOT NULL DEFAULT '';
ALTER TABLE audit_log ADD COLUMN blast_class TEXT NOT NULL DEFAULT '';
ALTER TABLE audit_log ADD COLUMN verdict TEXT NOT NULL DEFAULT '';
ALTER TABLE audit_log ADD COLUMN target_app TEXT NOT NULL DEFAULT '';
ALTER TABLE audit_log ADD COLUMN target_url TEXT NOT NULL DEFAULT '';
ALTER TABLE audit_log ADD COLUMN path TEXT NOT NULL DEFAULT '';
ALTER TABLE audit_log ADD COLUMN command TEXT NOT NULL DEFAULT '';
ALTER TABLE audit_log ADD COLUMN consent_result TEXT NOT NULL DEFAULT '';
ALTER TABLE audit_log ADD COLUMN screenshot_before_ref TEXT NOT NULL DEFAULT '';
ALTER TABLE audit_log ADD COLUMN screenshot_after_ref TEXT NOT NULL DEFAULT '';
ALTER TABLE audit_log ADD COLUMN session_id TEXT NOT NULL DEFAULT '';
CREATE INDEX IF NOT EXISTS idx_audit_session ON audit_log(session_id) WHERE session_id != '';
CREATE INDEX IF NOT EXISTS idx_audit_kind ON audit_log(kind) WHERE kind != '';

-- Replay screenshots: on-disk store referenced by audit_log.
-- The actual image bytes live in <data_dir>/replay/; this table
-- is a metadata index. Encryption is the storage.DB's domain
-- (encrypted via the same master key when written to a sidecar
-- metadata file).
CREATE TABLE IF NOT EXISTS replay_screenshots (
    id TEXT PRIMARY KEY,
    captured_at TEXT NOT NULL,
    audit_event_id INTEGER NOT NULL,
    position TEXT NOT NULL CHECK (position IN ('before', 'after')),
    width INTEGER NOT NULL DEFAULT 0,
    height INTEGER NOT NULL DEFAULT 0,
    byte_size INTEGER NOT NULL DEFAULT 0
);
CREATE INDEX IF NOT EXISTS idx_replay_captured_at ON replay_screenshots(captured_at);
CREATE INDEX IF NOT EXISTS idx_replay_audit_event ON replay_screenshots(audit_event_id);
`,
	},
	{
		Version: 4,
		Name:    "rollback checkpoints persisted to disk",
		SQL: `
-- Rollback checkpoints: persisted so they survive daemon restarts.
-- Previously CreateCheckpoint returned an in-memory struct only.
CREATE TABLE IF NOT EXISTS rollback_checkpoints (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    created_at TEXT NOT NULL,
    reason TEXT NOT NULL DEFAULT ''
);
CREATE INDEX IF NOT EXISTS idx_rb_cp_created ON rollback_checkpoints(created_at);
`,
	},
}

// migrate applies all pending migrations in order. Idempotent.
func (d *DB) migrate(ctx context.Context, onMigrate func(int) error) error {
	// Ensure schema_version table exists (handles fresh databases).
	if _, err := d.sql.ExecContext(ctx, `
CREATE TABLE IF NOT EXISTS schema_version (
    version INTEGER NOT NULL PRIMARY KEY,
    name TEXT NOT NULL,
    applied_at TEXT NOT NULL
);`); err != nil {
		return fmt.Errorf("create schema_version: %w", err)
	}

	current, err := d.currentVersion(ctx)
	if err != nil {
		return err
	}

	for _, m := range migrations {
		if m.Version <= current {
			continue
		}
		if err := d.applyMigration(ctx, m, onMigrate); err != nil {
			return err
		}
	}
	return nil
}

func (d *DB) currentVersion(ctx context.Context) (int, error) {
	var v sql.NullInt64
	err := d.sql.QueryRowContext(ctx, `SELECT MAX(version) FROM schema_version`).Scan(&v)
	if err != nil {
		return 0, fmt.Errorf("query schema_version: %w", err)
	}
	if !v.Valid {
		return 0, nil
	}
	return int(v.Int64), nil
}

func (d *DB) applyMigration(ctx context.Context, m migration, onMigrate func(int) error) error {
	tx, err := d.sql.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin tx (v%d %s): %w", m.Version, m.Name, err)
	}
	defer func() { _ = tx.Rollback() }()

	if _, err := tx.ExecContext(ctx, m.SQL); err != nil {
		return fmt.Errorf("apply v%d %s: %w", m.Version, m.Name, err)
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO schema_version (version, name, applied_at) VALUES (?, ?, datetime('now'))`,
		m.Version, m.Name,
	); err != nil {
		return fmt.Errorf("record v%d: %w", m.Version, err)
	}
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("commit v%d: %w", m.Version, err)
	}
	if onMigrate != nil {
		if err := onMigrate(m.Version); err != nil {
			return fmt.Errorf("onMigrate v%d: %w", m.Version, err)
		}
	}
	return nil
}

// ErrNoMigration is returned by EnsureVersion when the requested version
// is not in the migration set.
var ErrNoMigration = errors.New("storage: no such migration")

// EnsureVersion returns nil if the database is at the given version, or
// ErrNoMigration if the version is not defined.
func (d *DB) EnsureVersion(ctx context.Context, want int) error {
	current, err := d.currentVersion(ctx)
	if err != nil {
		return err
	}
	if current == want {
		return nil
	}
	for _, m := range migrations {
		if m.Version == want {
			if current < want {
				return fmt.Errorf("storage: schema at v%d, want v%d (not yet applied)", current, want)
			}
			return fmt.Errorf("storage: schema at v%d, want v%d (downgrade not supported)", current, want)
		}
	}
	return ErrNoMigration
}
