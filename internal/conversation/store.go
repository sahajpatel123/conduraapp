// Package conversation provides SQLite-backed storage for chat
// conversations. Phase 2 (per the locked-in decision) stores the
// current conversation only; the schema is general enough to
// support full history if we want it later.
package conversation

import (
	"context"
	"database/sql"
	"encoding/json"
	"errors"
	"fmt"
	"time"
)

// Message is a single chat message. Mirrors the JSON shape sent
// to/from the GUI.
type Message struct {
	Role       string          `json:"role"`
	Content    string          `json:"content"`
	ToolCalls  json.RawMessage `json:"tool_calls,omitempty"`
	ToolCallID string          `json:"tool_call_id,omitempty"`
}

// Meta is the sidebar entry for a conversation.
type Meta struct {
	ID           int64     `json:"id"`
	Title        string    `json:"title"`
	CreatedAt    time.Time `json:"created_at"`
	UpdatedAt    time.Time `json:"updated_at"`
	MessageCount int       `json:"message_count"`
}

// Conversation is a full conversation (with messages).
type Conversation struct {
	Meta
	Messages []Message `json:"messages"`
}

// Store provides CRUD over the conversations + conversation_messages
// tables. Constructed once at daemon startup, shared across all
// IPC handlers.
type Store struct {
	db *sql.DB
}

// New returns a Store wrapping the given database. The database
// must have the conversation tables (created by storage migration
// v2).
func New(db *sql.DB) *Store {
	return &Store{db: db}
}

// Create creates a new empty conversation and returns its meta.
// title defaults to "New conversation" if empty.
func (s *Store) Create(ctx context.Context, title string) (Meta, error) {
	if title == "" {
		title = "New conversation"
	}
	now := time.Now().UTC()
	res, err := s.db.ExecContext(ctx,
		`INSERT INTO conversations (title, created_at, updated_at) VALUES (?, ?, ?)`,
		title, now.Format(time.RFC3339), now.Format(time.RFC3339),
	)
	if err != nil {
		return Meta{}, fmt.Errorf("insert conversation: %w", err)
	}
	id, err := res.LastInsertId()
	if err != nil {
		return Meta{}, fmt.Errorf("last insert id: %w", err)
	}
	return Meta{ID: id, Title: title, CreatedAt: now, UpdatedAt: now, MessageCount: 0}, nil
}

// Get returns the full conversation (with messages) by id.
// Returns ErrNotFound if the id does not exist.
func (s *Store) Get(ctx context.Context, id int64) (Conversation, error) {
	var c Conversation
	var created, updated string
	row := s.db.QueryRowContext(ctx,
		`SELECT id, title, created_at, updated_at FROM conversations WHERE id = ?`, id,
	)
	if err := row.Scan(&c.ID, &c.Title, &created, &updated); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return Conversation{}, ErrNotFound
		}
		return Conversation{}, fmt.Errorf("select conversation: %w", err)
	}
	c.CreatedAt, _ = time.Parse(time.RFC3339, created)
	c.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
	rows, err := s.db.QueryContext(ctx,
		`SELECT role, content, tool_calls_json, tool_call_id
		 FROM conversation_messages
		 WHERE conversation_id = ?
		 ORDER BY id ASC`, id,
	)
	if err != nil {
		return Conversation{}, fmt.Errorf("select messages: %w", err)
	}
	defer func() { _ = rows.Close() }()
	for rows.Next() {
		var m Message
		var tcs sql.NullString
		var tcid sql.NullString
		if err := rows.Scan(&m.Role, &m.Content, &tcs, &tcid); err != nil {
			return Conversation{}, fmt.Errorf("scan message: %w", err)
		}
		if tcs.Valid {
			m.ToolCalls = json.RawMessage(tcs.String)
		}
		if tcid.Valid {
			m.ToolCallID = tcid.String
		}
		c.Messages = append(c.Messages, m)
	}
	c.MessageCount = len(c.Messages)
	return c, nil
}

// List returns all conversations ordered by updated_at desc.
func (s *Store) List(ctx context.Context) ([]Meta, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT c.id, c.title, c.created_at, c.updated_at, COUNT(m.id)
		 FROM conversations c
		 LEFT JOIN conversation_messages m ON m.conversation_id = c.id
		 GROUP BY c.id
		 ORDER BY c.updated_at DESC`,
	)
	if err != nil {
		return nil, fmt.Errorf("query conversations: %w", err)
	}
	defer func() { _ = rows.Close() }()
	var out []Meta
	for rows.Next() {
		var m Meta
		var created, updated string
		if err := rows.Scan(&m.ID, &m.Title, &created, &updated, &m.MessageCount); err != nil {
			return nil, fmt.Errorf("scan: %w", err)
		}
		m.CreatedAt, _ = time.Parse(time.RFC3339, created)
		m.UpdatedAt, _ = time.Parse(time.RFC3339, updated)
		out = append(out, m)
	}
	return out, nil
}

// Delete removes the conversation (and its messages, via FK cascade).
func (s *Store) Delete(ctx context.Context, id int64) error {
	res, err := s.db.ExecContext(ctx, `DELETE FROM conversations WHERE id = ?`, id)
	if err != nil {
		return fmt.Errorf("delete: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return ErrNotFound
	}
	return nil
}

// Append appends a message to the conversation and bumps updated_at.
func (s *Store) Append(ctx context.Context, id int64, m Message) error {
	now := time.Now().UTC()
	tx, err := s.db.BeginTx(ctx, nil)
	if err != nil {
		return fmt.Errorf("begin: %w", err)
	}
	defer func() { _ = tx.Rollback() }()

	var tcs interface{}
	if len(m.ToolCalls) > 0 {
		tcs = string(m.ToolCalls)
	}
	var tcid interface{}
	if m.ToolCallID != "" {
		tcid = m.ToolCallID
	}
	if _, err := tx.ExecContext(ctx,
		`INSERT INTO conversation_messages (conversation_id, role, content, tool_calls_json, tool_call_id, created_at)
		 VALUES (?, ?, ?, ?, ?, ?)`,
		id, m.Role, m.Content, tcs, tcid, now.Format(time.RFC3339),
	); err != nil {
		return fmt.Errorf("insert message: %w", err)
	}
	if _, err := tx.ExecContext(ctx,
		`UPDATE conversations SET updated_at = ? WHERE id = ?`,
		now.Format(time.RFC3339), id,
	); err != nil {
		return fmt.Errorf("update conversation: %w", err)
	}
	return tx.Commit()
}

// ErrNotFound is returned when a conversation is not found.
var ErrNotFound = errors.New("conversation: not found")
