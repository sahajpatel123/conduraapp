package reach

import (
	"context"
	"database/sql"
	"fmt"
)

// Store persists channel state in SQLite.
type Store struct {
	db *sql.DB
}

// NewStore creates the reach_channels table and returns a Store.
func NewStore(db *sql.DB) (*Store, error) {
	_, err := db.ExecContext(context.Background(), `
CREATE TABLE IF NOT EXISTS reach_channels (
    name         TEXT PRIMARY KEY,
    token        TEXT DEFAULT '',
    chat_id      TEXT DEFAULT '',
    enabled      INTEGER DEFAULT 0,
    connected_at TEXT DEFAULT ''
);
`)
	if err != nil {
		return nil, fmt.Errorf("reach: create channels table: %w", err)
	}
	return &Store{db: db}, nil
}

// Save creates or updates a channel record.
func (s *Store) Save(ctx context.Context, name, chatID string, enabled bool) error {
	enabledInt := 0
	if enabled {
		enabledInt = 1
	}
	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO reach_channels (name, chat_id, enabled) VALUES (?, ?, ?)`,
		name, chatID, enabledInt,
	)
	return err
}

// List returns all channels.
func (s *Store) List(ctx context.Context) ([]ChannelStatus, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT name, chat_id, enabled FROM reach_channels ORDER BY name`)
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

// Delete removes a channel record.
func (s *Store) Delete(ctx context.Context, name string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM reach_channels WHERE name = ?`, name)
	return err
}
