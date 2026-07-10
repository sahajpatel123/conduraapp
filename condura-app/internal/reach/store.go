package reach

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
)

// Store persists channel state in SQLite. Channel tokens are stored
// only as ciphertext (see Cipher interface in reach.go); the
// plaintext token never appears in this package's SQL.
type Store struct {
	db *sql.DB
}

// NewStore creates the reach_channels table and returns a Store.
//
// The `token` column is legacy: declared on first install before
// token persistence was designed, it was never written to. Migration
// v8 (storage/migrations.go) added `token_ciphertext` which is the
// real field. The legacy `token` column is harmless — always
// empty in practice — and removing it would require a destructive
// rename. A future migration can drop it when convenient.
func NewStore(db *sql.DB) (*Store, error) {
	_, err := db.ExecContext(context.Background(), `
CREATE TABLE IF NOT EXISTS reach_channels (
    name             TEXT PRIMARY KEY,
    token            TEXT DEFAULT '',
    chat_id          TEXT DEFAULT '',
    enabled          INTEGER DEFAULT 0,
    connected_at     TEXT DEFAULT '',
    token_ciphertext TEXT NOT NULL DEFAULT ''
);
`)
	if err != nil {
		return nil, fmt.Errorf("reach: create channels table: %w", err)
	}
	return &Store{db: db}, nil
}

// Save creates or updates a channel record.
//
// tokenCiphertext is the result of Cipher.EncryptStringWithAAD —
// the Store never sees the plaintext token. Pass "" if no token
// (e.g., when only updating chat_id via a future Save variant;
// currently Save is the single write path and the caller passes
// the existing ciphertext).
func (s *Store) Save(ctx context.Context, name, tokenCiphertext, chatID string, enabled bool) error {
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO reach_channels (name, token_ciphertext, chat_id, enabled) VALUES (?, ?, ?, ?)`,
		name, tokenCiphertext, chatID, enabledInt,
	)
	return err
}

// UpdateChatID writes the captured chat_id without touching the
// token. Used by the receive loop after the first inbound message
// from a Telegram bot — at that point we finally know which chat
// the user has been reaching us from.
func (s *Store) UpdateChatID(ctx context.Context, name, chatID string) error {
	_, err := s.db.ExecContext(ctx,
		`UPDATE reach_channels SET chat_id = ? WHERE name = ?`,
		chatID, name,
	)
	return err
}

// GetTokenCiphertext returns the encrypted token for a channel.
// Returns ("", nil) if the channel has no row.
func (s *Store) GetTokenCiphertext(ctx context.Context, name string) (string, error) {
	var ct string
	err := s.db.QueryRowContext(ctx,
		`SELECT token_ciphertext FROM reach_channels WHERE name = ?`,
		name,
	).Scan(&ct)
	if errors.Is(err, sql.ErrNoRows) {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("reach: get token ciphertext: %w", err)
	}
	return ct, nil
}

// List returns all channels. Token ciphertext is intentionally NOT
// returned — the GUI has no business seeing it.
func (s *Store) List(ctx context.Context) ([]ChannelStatus, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT name, chat_id, enabled FROM reach_channels ORDER BY name`,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	var statuses []ChannelStatus
	for rows.Next() {
		var name, chatID string
		var enabled int
		if err := rows.Scan(&name, &chatID, &enabled); err != nil {
			return nil, fmt.Errorf("reach: scan channel: %w", err)
		}
		statuses = append(statuses, ChannelStatus{
			Name:      name,
			Connected: enabled == 1,
			ChatID:    chatID,
		})
	}
	return statuses, nil
}

// Delete removes a channel record. Called from Manager.Disconnect.
func (s *Store) Delete(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM reach_channels WHERE name = ?`, name)
	return err
}
