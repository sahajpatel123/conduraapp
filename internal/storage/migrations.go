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
