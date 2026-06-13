package skills

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct{ db *sql.DB }

// defaultText is the SQLite column type used for optional text
// provenance fields. Extracted to avoid repeated literals.
const defaultTextCol = "TEXT NOT NULL DEFAULT ''"

// NewSQLiteStore opens a SQLite-backed skill store.
func NewSQLiteStore(path string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", path)
	if err != nil {
		return nil, fmt.Errorf("skills: open: %w", err)
	}
	if _, err := db.ExecContext(context.Background(), "PRAGMA journal_mode=WAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("skills: wal: %w", err)
	}
	s := &SQLiteStore{db: db}
	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, err
	}
	return s, nil
}

func (s *SQLiteStore) migrate() error {
	schema := fmt.Sprintf(`
		CREATE TABLE IF NOT EXISTS skills (
			id TEXT PRIMARY KEY, name TEXT NOT NULL,
			description TEXT NOT NULL DEFAULT '', version TEXT NOT NULL DEFAULT '0.1.0',
			trust TEXT NOT NULL DEFAULT 'community', trigger_pattern TEXT NOT NULL DEFAULT '',
			steps TEXT NOT NULL DEFAULT '[]', dependencies TEXT NOT NULL DEFAULT '[]',
			success_count INTEGER NOT NULL DEFAULT 0, failure_count INTEGER NOT NULL DEFAULT 0,
			created_at DATETIME NOT NULL, updated_at DATETIME NOT NULL, last_used DATETIME NOT NULL,
			author %s, author_key %s,
			license %s, source TEXT NOT NULL DEFAULT 'local',
			hub_id %s, checksum %s,
			published_at DATETIME
		)`, defaultTextCol, defaultTextCol, defaultTextCol, defaultTextCol, defaultTextCol)
	_, err := s.db.ExecContext(context.Background(), schema)
	if err != nil {
		return err
	}
	// Ensure provenance columns exist (migration for existing installs).
	provenanceCols := []struct{ name, def string }{
		{"author", defaultTextCol},
		{"author_key", defaultTextCol},
		{"license", defaultTextCol},
		{"source", "TEXT NOT NULL DEFAULT 'local'"},
		{"hub_id", defaultTextCol},
		{"checksum", defaultTextCol},
		{"published_at", "DATETIME"},
	}
	for _, col := range provenanceCols {
		_, _ = s.db.ExecContext(context.Background(),
			fmt.Sprintf("ALTER TABLE skills ADD COLUMN %s %s DEFAULT ''", col.name, col.def))
	}
	return nil
}

// Create persists a new skill.
func (s *SQLiteStore) Create(ctx context.Context, sk *Skill) error {
	stepsJSON, _ := json.Marshal(sk.Steps)
	depsJSON, _ := json.Marshal(sk.Dependencies)
	_, err := s.db.ExecContext(ctx,
		`INSERT INTO skills VALUES(?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?,?)`,
		sk.ID, sk.Name, sk.Description, sk.Version, string(sk.Trust),
		sk.TriggerPattern, string(stepsJSON), string(depsJSON),
		sk.SuccessCount, sk.FailureCount, sk.CreatedAt, sk.UpdatedAt, sk.LastUsed,
		sk.Author, sk.AuthorKey, sk.License, sk.Source, sk.HubID, sk.Checksum, sk.PublishedAt,
	)
	return err
}

// Get retrieves a skill by ID.
func (s *SQLiteStore) Get(ctx context.Context, id string) (*Skill, error) {
	r := s.db.QueryRowContext(ctx, `SELECT * FROM skills WHERE id=?`, id)
	return scanSkill(r)
}

// List returns the most recently updated skills.
func (s *SQLiteStore) List(ctx context.Context, limit int) ([]*Skill, error) {
	rows, err := s.db.QueryContext(ctx, `SELECT * FROM skills ORDER BY updated_at DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanSkills(rows)
}

// Search finds skills by name or trigger pattern.
func (s *SQLiteStore) Search(ctx context.Context, query string, limit int) ([]*Skill, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT * FROM skills WHERE name LIKE ? OR trigger_pattern LIKE ? ORDER BY success_count DESC LIMIT ?`,
		"%"+query+"%", "%"+query+"%", limit)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()
	return scanSkills(rows)
}

// Update modifies an existing skill.
func (s *SQLiteStore) Update(ctx context.Context, sk *Skill) error {
	stepsJSON, _ := json.Marshal(sk.Steps)
	depsJSON, _ := json.Marshal(sk.Dependencies)
	_, err := s.db.ExecContext(ctx,
		`UPDATE skills SET name=?,description=?,version=?,trust=?,trigger_pattern=?,steps=?,dependencies=?,success_count=?,failure_count=?,updated_at=?,last_used=?,author=?,author_key=?,license=?,source=?,hub_id=?,checksum=? WHERE id=?`,
		sk.Name, sk.Description, sk.Version, string(sk.Trust),
		sk.TriggerPattern, string(stepsJSON), string(depsJSON),
		sk.SuccessCount, sk.FailureCount, sk.UpdatedAt, sk.LastUsed,
		sk.Author, sk.AuthorKey, sk.License, sk.Source, sk.HubID, sk.Checksum, sk.ID,
	)
	return err
}

// Delete removes a skill.
func (s *SQLiteStore) Delete(ctx context.Context, id string) error {
	_, err := s.db.ExecContext(ctx, `DELETE FROM skills WHERE id=?`, id)
	return err
}

// IncrementUsage bumps the success or failure counter.
func (s *SQLiteStore) IncrementUsage(ctx context.Context, id string, success bool) error {
	now := time.Now()
	if success {
		_, err := s.db.ExecContext(ctx, `UPDATE skills SET success_count=success_count+1,last_used=?,updated_at=? WHERE id=?`, now, now, id)
		return err
	}
	_, err := s.db.ExecContext(ctx, `UPDATE skills SET failure_count=failure_count+1,last_used=?,updated_at=? WHERE id=?`, now, now, id)
	return err
}

// Close shuts down the database connection.
func (s *SQLiteStore) Close() error { return s.db.Close() }

func scanSkill(row *sql.Row) (*Skill, error) {
	var sk Skill
	var stepsJSON, depsJSON, trust string
	err := row.Scan(&sk.ID, &sk.Name, &sk.Description, &sk.Version, &trust,
		&sk.TriggerPattern, &stepsJSON, &depsJSON,
		&sk.SuccessCount, &sk.FailureCount, &sk.CreatedAt, &sk.UpdatedAt, &sk.LastUsed,
		&sk.Author, &sk.AuthorKey, &sk.License, &sk.Source, &sk.HubID, &sk.Checksum, &sk.PublishedAt)
	if err != nil {
		return nil, err
	}
	sk.Trust = TrustLevel(trust)
	_ = json.Unmarshal([]byte(stepsJSON), &sk.Steps)
	_ = json.Unmarshal([]byte(depsJSON), &sk.Dependencies)
	return &sk, nil
}

func scanSkills(rows *sql.Rows) ([]*Skill, error) {
	var out []*Skill
	for rows.Next() {
		var sk Skill
		var stepsJSON, depsJSON, trust string
		if err := rows.Scan(&sk.ID, &sk.Name, &sk.Description, &sk.Version, &trust,
			&sk.TriggerPattern, &stepsJSON, &depsJSON,
			&sk.SuccessCount, &sk.FailureCount, &sk.CreatedAt, &sk.UpdatedAt, &sk.LastUsed,
			&sk.Author, &sk.AuthorKey, &sk.License, &sk.Source, &sk.HubID, &sk.Checksum, &sk.PublishedAt); err != nil {
			return out, err
		}
		sk.Trust = TrustLevel(trust)
		_ = json.Unmarshal([]byte(stepsJSON), &sk.Steps)
		_ = json.Unmarshal([]byte(depsJSON), &sk.Dependencies)
		out = append(out, &sk)
	}
	return out, rows.Err()
}
