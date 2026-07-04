package memory

import (
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"time"

	_ "modernc.org/sqlite"
)

// SQLiteStore implements Store using SQLite.
type SQLiteStore struct {
	db     *sql.DB
	closed bool
}

// NewSQLiteStore creates a new SQLite-backed memory store.
func NewSQLiteStore(dbPath string) (*SQLiteStore, error) {
	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("memory: failed to open database: %w", err)
	}

	// Enable WAL mode for better concurrent performance
	if _, err := db.ExecContext(context.Background(), "PRAGMA journal_mode=WAL"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("memory: failed to enable WAL mode: %w", err)
	}

	// Busy timeout and single connection for WAL safety.
	_ = db.PingContext(context.Background())
	db.SetMaxOpenConns(1)
	if _, err := db.ExecContext(context.Background(), "PRAGMA busy_timeout=5000"); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("memory: failed to set busy timeout: %w", err)
	}

	store := &SQLiteStore{db: db}
	if err := store.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("memory: failed to migrate database: %w", err)
	}

	return store, nil
}

// migrate creates the database schema.
func (s *SQLiteStore) migrate() error {
	queries := []string{
		`CREATE TABLE IF NOT EXISTS episodes (
			id TEXT PRIMARY KEY,
			session_id TEXT NOT NULL,
			user_message TEXT,
			agent_response TEXT,
			actions_taken TEXT,
			timestamp DATETIME NOT NULL,
			summary TEXT
		)`,
		`CREATE TABLE IF NOT EXISTS facts (
			id TEXT PRIMARY KEY,
			category TEXT NOT NULL,
			content TEXT NOT NULL,
			confidence REAL DEFAULT 0.5,
			source_episode_id TEXT,
			created_at DATETIME NOT NULL,
			updated_at DATETIME NOT NULL
		)`,
		`CREATE TABLE IF NOT EXISTS skills (
			id TEXT PRIMARY KEY,
			name TEXT NOT NULL,
			description TEXT,
			trigger_pattern TEXT,
			steps TEXT,
			success_count INTEGER DEFAULT 0,
			failure_count INTEGER DEFAULT 0,
			created_at DATETIME NOT NULL,
			last_used DATETIME
		)`,
		`CREATE INDEX IF NOT EXISTS idx_episodes_session ON episodes(session_id)`,
		`CREATE INDEX IF NOT EXISTS idx_episodes_timestamp ON episodes(timestamp)`,
		`CREATE INDEX IF NOT EXISTS idx_facts_category ON facts(category)`,
		`CREATE INDEX IF NOT EXISTS idx_skills_name ON skills(name)`,
		// FTS5 virtual table for full-text search across episodes.
		`CREATE VIRTUAL TABLE IF NOT EXISTS episodes_fts USING fts5(
			id,
			session_id,
			user_message,
			agent_response,
			summary,
			content='episodes',
			content_rowid='rowid'
		)`,
		// FTS5 virtual table for full-text search across facts.
		`CREATE VIRTUAL TABLE IF NOT EXISTS facts_fts USING fts5(
			id,
			category,
			content,
			content='facts',
			content_rowid='rowid'
		)`,
		// FTS5 virtual table for full-text search across skills.
		`CREATE VIRTUAL TABLE IF NOT EXISTS skills_fts USING fts5(
			id,
			name,
			description,
			content='skills',
			content_rowid='rowid'
		)`,
		// Triggers to keep episodes_fts in sync.
		`CREATE TRIGGER IF NOT EXISTS episodes_ai AFTER INSERT ON episodes BEGIN
			INSERT INTO episodes_fts(rowid, id, session_id, user_message, agent_response, summary)
			VALUES (new.rowid, new.id, new.session_id, new.user_message, new.agent_response, new.summary);
		END`,
		`CREATE TRIGGER IF NOT EXISTS episodes_ad AFTER DELETE ON episodes BEGIN
			INSERT INTO episodes_fts(episodes_fts, rowid, id, session_id, user_message, agent_response, summary)
			VALUES ('delete', old.rowid, old.id, old.session_id, old.user_message, old.agent_response, old.summary);
		END`,
		`CREATE TRIGGER IF NOT EXISTS episodes_au AFTER UPDATE ON episodes BEGIN
			INSERT INTO episodes_fts(episodes_fts, rowid, id, session_id, user_message, agent_response, summary)
			VALUES ('delete', old.rowid, old.id, old.session_id, old.user_message, old.agent_response, old.summary);
			INSERT INTO episodes_fts(rowid, id, session_id, user_message, agent_response, summary)
			VALUES (new.rowid, new.id, new.session_id, new.user_message, new.agent_response, new.summary);
		END`,
		// Triggers to keep facts_fts in sync.
		`CREATE TRIGGER IF NOT EXISTS facts_ai AFTER INSERT ON facts BEGIN
			INSERT INTO facts_fts(rowid, id, category, content)
			VALUES (new.rowid, new.id, new.category, new.content);
		END`,
		`CREATE TRIGGER IF NOT EXISTS facts_ad AFTER DELETE ON facts BEGIN
			INSERT INTO facts_fts(facts_fts, rowid, id, category, content)
			VALUES ('delete', old.rowid, old.id, old.category, old.content);
		END`,
		`CREATE TRIGGER IF NOT EXISTS facts_au AFTER UPDATE ON facts BEGIN
			INSERT INTO facts_fts(facts_fts, rowid, id, category, content)
			VALUES ('delete', old.rowid, old.id, old.category, old.content);
			INSERT INTO facts_fts(rowid, id, category, content)
			VALUES (new.rowid, new.id, new.category, new.content);
		END`,
		// Triggers to keep skills_fts in sync.
		`CREATE TRIGGER IF NOT EXISTS skills_ai AFTER INSERT ON skills BEGIN
			INSERT INTO skills_fts(rowid, id, name, description)
			VALUES (new.rowid, new.id, new.name, new.description);
		END`,
		`CREATE TRIGGER IF NOT EXISTS skills_ad AFTER DELETE ON skills BEGIN
			INSERT INTO skills_fts(skills_fts, rowid, id, name, description)
			VALUES ('delete', old.rowid, old.id, old.name, old.description);
		END`,
		`CREATE TRIGGER IF NOT EXISTS skills_au AFTER UPDATE ON skills BEGIN
			INSERT INTO skills_fts(skills_fts, rowid, id, name, description)
			VALUES ('delete', old.rowid, old.id, old.name, old.description);
			INSERT INTO skills_fts(rowid, id, name, description)
			VALUES (new.rowid, new.id, new.name, new.description);
		END`,
	}

	for _, q := range queries {
		if _, err := s.db.ExecContext(context.Background(), q); err != nil {
			return fmt.Errorf("migration failed: %w", err)
		}
	}

	return nil
}

// StoreEpisode stores an episode in the database.
func (s *SQLiteStore) StoreEpisode(ctx context.Context, episode *Episode) error {
	if s.closed {
		return ErrStoreClosed
	}

	actionsJSON, err := json.Marshal(episode.ActionsTaken)
	if err != nil {
		return fmt.Errorf("failed to marshal actions: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO episodes (id, session_id, user_message, agent_response, actions_taken, timestamp, summary)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		episode.ID, episode.SessionID, episode.UserMessage, episode.AgentResponse,
		string(actionsJSON), episode.Timestamp, episode.Summary,
	)
	return err
}

// GetEpisode retrieves an episode by ID.
func (s *SQLiteStore) GetEpisode(ctx context.Context, id string) (*Episode, error) {
	if s.closed {
		return nil, ErrStoreClosed
	}

	row := s.db.QueryRowContext(ctx,
		`SELECT id, session_id, user_message, agent_response, actions_taken, timestamp, summary
		 FROM episodes WHERE id = ?`, id,
	)

	episode := &Episode{}
	var actionsJSON string
	err := row.Scan(&episode.ID, &episode.SessionID, &episode.UserMessage,
		&episode.AgentResponse, &actionsJSON, &episode.Timestamp, &episode.Summary)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(actionsJSON), &episode.ActionsTaken); err != nil {
		return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
	}

	return episode, nil
}

// ListEpisodes lists recent episodes.
func (s *SQLiteStore) ListEpisodes(ctx context.Context, limit int) ([]*Episode, error) {
	if s.closed {
		return nil, ErrStoreClosed
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, session_id, user_message, agent_response, actions_taken, timestamp, summary
		 FROM episodes ORDER BY timestamp DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var episodes []*Episode
	for rows.Next() {
		episode := &Episode{}
		var actionsJSON string
		if err := rows.Scan(&episode.ID, &episode.SessionID, &episode.UserMessage,
			&episode.AgentResponse, &actionsJSON, &episode.Timestamp, &episode.Summary); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(actionsJSON), &episode.ActionsTaken); err != nil {
			return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
		}
		episodes = append(episodes, episode)
	}

	return episodes, rows.Err()
}

// SearchEpisodes searches episodes by text content using FTS5.
func (s *SQLiteStore) SearchEpisodes(ctx context.Context, query string, limit int) ([]*Episode, error) {
	if s.closed {
		return nil, ErrStoreClosed
	}

	// Use FTS5 MATCH for full-text search.
	// The FTS5 table has the same rowids as the episodes table,
	// so we join on rowid to get the full episode data.
	rows, err := s.db.QueryContext(ctx,
		`SELECT e.id, e.session_id, e.user_message, e.agent_response, e.actions_taken, e.timestamp, e.summary
		 FROM episodes e
		 INNER JOIN episodes_fts f ON e.rowid = f.rowid
		 WHERE episodes_fts MATCH ?
		 ORDER BY e.timestamp DESC LIMIT ?`,
		query, limit,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var episodes []*Episode
	for rows.Next() {
		episode := &Episode{}
		var actionsJSON string
		if err := rows.Scan(&episode.ID, &episode.SessionID, &episode.UserMessage,
			&episode.AgentResponse, &actionsJSON, &episode.Timestamp, &episode.Summary); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(actionsJSON), &episode.ActionsTaken); err != nil {
			return nil, fmt.Errorf("failed to unmarshal actions: %w", err)
		}
		episodes = append(episodes, episode)
	}

	return episodes, rows.Err()
}

// StoreFact stores a fact in the database.
func (s *SQLiteStore) StoreFact(ctx context.Context, fact *Fact) error {
	if s.closed {
		return ErrStoreClosed
	}

	if fact.Confidence < 0 || fact.Confidence > 1 {
		return ErrInvalidConfidence
	}

	_, err := s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO facts (id, category, content, confidence, source_episode_id, created_at, updated_at)
		 VALUES (?, ?, ?, ?, ?, ?, ?)`,
		fact.ID, fact.Category, fact.Content, fact.Confidence,
		fact.SourceEpisodeID, fact.CreatedAt, fact.UpdatedAt,
	)
	return err
}

// GetFact retrieves a fact by ID.
func (s *SQLiteStore) GetFact(ctx context.Context, id string) (*Fact, error) {
	if s.closed {
		return nil, ErrStoreClosed
	}

	row := s.db.QueryRowContext(ctx,
		`SELECT id, category, content, confidence, source_episode_id, created_at, updated_at
		 FROM facts WHERE id = ?`, id,
	)

	fact := &Fact{}
	err := row.Scan(&fact.ID, &fact.Category, &fact.Content, &fact.Confidence,
		&fact.SourceEpisodeID, &fact.CreatedAt, &fact.UpdatedAt)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	return fact, nil
}

// ListFacts lists facts, optionally filtered by category.
func (s *SQLiteStore) ListFacts(ctx context.Context, category string, limit int) ([]*Fact, error) {
	if s.closed {
		return nil, ErrStoreClosed
	}

	var rows *sql.Rows
	var err error

	if category != "" {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, category, content, confidence, source_episode_id, created_at, updated_at
			 FROM facts WHERE category = ? ORDER BY confidence DESC LIMIT ?`, category, limit,
		)
	} else {
		rows, err = s.db.QueryContext(ctx,
			`SELECT id, category, content, confidence, source_episode_id, created_at, updated_at
			 FROM facts ORDER BY confidence DESC LIMIT ?`, limit,
		)
	}
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var facts []*Fact
	for rows.Next() {
		fact := &Fact{}
		if err := rows.Scan(&fact.ID, &fact.Category, &fact.Content, &fact.Confidence,
			&fact.SourceEpisodeID, &fact.CreatedAt, &fact.UpdatedAt); err != nil {
			return nil, err
		}
		facts = append(facts, fact)
	}

	return facts, rows.Err()
}

// UpdateFactConfidence updates the confidence of a fact.
func (s *SQLiteStore) UpdateFactConfidence(ctx context.Context, id string, confidence float64) error {
	if s.closed {
		return ErrStoreClosed
	}

	if confidence < 0 || confidence > 1 {
		return ErrInvalidConfidence
	}

	result, err := s.db.ExecContext(ctx,
		`UPDATE facts SET confidence = ?, updated_at = ? WHERE id = ?`,
		confidence, time.Now(), id,
	)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// StoreSkill stores a skill in the database.
func (s *SQLiteStore) StoreSkill(ctx context.Context, skill *Skill) error {
	if s.closed {
		return ErrStoreClosed
	}

	stepsJSON, err := json.Marshal(skill.Steps)
	if err != nil {
		return fmt.Errorf("failed to marshal steps: %w", err)
	}

	_, err = s.db.ExecContext(ctx,
		`INSERT OR REPLACE INTO skills (id, name, description, trigger_pattern, steps, success_count, failure_count, created_at, last_used)
		 VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)`,
		skill.ID, skill.Name, skill.Description, skill.TriggerPattern,
		string(stepsJSON), skill.SuccessCount, skill.FailureCount,
		skill.CreatedAt, skill.LastUsed,
	)
	return err
}

// GetSkill retrieves a skill by ID.
func (s *SQLiteStore) GetSkill(ctx context.Context, id string) (*Skill, error) {
	if s.closed {
		return nil, ErrStoreClosed
	}

	row := s.db.QueryRowContext(ctx,
		`SELECT id, name, description, trigger_pattern, steps, success_count, failure_count, created_at, last_used
		 FROM skills WHERE id = ?`, id,
	)

	skill := &Skill{}
	var stepsJSON string
	err := row.Scan(&skill.ID, &skill.Name, &skill.Description, &skill.TriggerPattern,
		&stepsJSON, &skill.SuccessCount, &skill.FailureCount, &skill.CreatedAt, &skill.LastUsed)
	if err == sql.ErrNoRows {
		return nil, ErrNotFound
	}
	if err != nil {
		return nil, err
	}

	if err := json.Unmarshal([]byte(stepsJSON), &skill.Steps); err != nil {
		return nil, fmt.Errorf("failed to unmarshal steps: %w", err)
	}

	return skill, nil
}

// ListSkills lists all skills.
func (s *SQLiteStore) ListSkills(ctx context.Context, limit int) ([]*Skill, error) {
	if s.closed {
		return nil, ErrStoreClosed
	}

	rows, err := s.db.QueryContext(ctx,
		`SELECT id, name, description, trigger_pattern, steps, success_count, failure_count, created_at, last_used
		 FROM skills ORDER BY last_used DESC LIMIT ?`, limit,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var skills []*Skill
	for rows.Next() {
		skill := &Skill{}
		var stepsJSON string
		if err := rows.Scan(&skill.ID, &skill.Name, &skill.Description, &skill.TriggerPattern,
			&stepsJSON, &skill.SuccessCount, &skill.FailureCount, &skill.CreatedAt, &skill.LastUsed); err != nil {
			return nil, err
		}
		if err := json.Unmarshal([]byte(stepsJSON), &skill.Steps); err != nil {
			return nil, fmt.Errorf("failed to unmarshal steps: %w", err)
		}
		skills = append(skills, skill)
	}

	return skills, rows.Err()
}

// IncrementSkillUsage increments the usage count of a skill.
func (s *SQLiteStore) IncrementSkillUsage(ctx context.Context, id string, success bool) error {
	if s.closed {
		return ErrStoreClosed
	}

	var query string
	if success {
		query = `UPDATE skills SET success_count = success_count + 1, last_used = ? WHERE id = ?`
	} else {
		query = `UPDATE skills SET failure_count = failure_count + 1, last_used = ? WHERE id = ?`
	}

	result, err := s.db.ExecContext(ctx, query, time.Now(), id)
	if err != nil {
		return err
	}

	rows, err := result.RowsAffected()
	if err != nil {
		return err
	}
	if rows == 0 {
		return ErrNotFound
	}

	return nil
}

// Search searches across all memory types.
func (s *SQLiteStore) Search(ctx context.Context, query *SearchQuery) ([]*SearchResult, error) {
	if s.closed {
		return nil, ErrStoreClosed
	}

	var results []*SearchResult

	// Search episodes
	if query.Type == "" || query.Type == Episodic {
		epResults, err := s.findEpisodeResults(ctx, query)
		if err != nil {
			return nil, err
		}
		results = append(results, epResults...)
	}

	// Search facts
	if query.Type == "" || query.Type == Semantic {
		factResults, err := s.findFactResults(ctx, query)
		if err != nil {
			return nil, err
		}
		results = append(results, factResults...)
	}

	// Search skills
	if query.Type == "" || query.Type == Procedural {
		skillResults, err := s.findSkillResults(ctx, query)
		if err != nil {
			return nil, err
		}
		results = append(results, skillResults...)
	}

	return results, nil
}

// findEpisodeResults searches episodes and returns results.
func (s *SQLiteStore) findEpisodeResults(ctx context.Context, query *SearchQuery) ([]*SearchResult, error) {
	episodes, err := s.SearchEpisodes(ctx, query.Query, query.Limit)
	if err != nil {
		return nil, err
	}

	var results []*SearchResult
	for _, ep := range episodes {
		results = append(results, &SearchResult{
			Memory: &Memory{
				ID:        ep.ID,
				Type:      Episodic,
				Content:   ep.UserMessage + " " + ep.AgentResponse,
				Timestamp: ep.Timestamp,
			},
			Score: 1.0,
		})
	}
	return results, nil
}

// findFactResults searches facts and returns results using FTS5.
func (s *SQLiteStore) findFactResults(ctx context.Context, query *SearchQuery) ([]*SearchResult, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT f.id, f.category, f.content, f.confidence, f.source_episode_id, f.created_at, f.updated_at
		 FROM facts f
		 INNER JOIN facts_fts ft ON f.rowid = ft.rowid
		 WHERE facts_fts MATCH ?
		 ORDER BY f.confidence DESC LIMIT ?`,
		query.Query, query.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []*SearchResult
	for rows.Next() {
		fact := &Fact{}
		if err := rows.Scan(&fact.ID, &fact.Category, &fact.Content, &fact.Confidence,
			&fact.SourceEpisodeID, &fact.CreatedAt, &fact.UpdatedAt); err != nil {
			return nil, err
		}
		results = append(results, &SearchResult{
			Memory: &Memory{
				ID:        fact.ID,
				Type:      Semantic,
				Content:   fact.Content,
				Timestamp: fact.CreatedAt,
			},
			Score: fact.Confidence,
		})
	}
	return results, rows.Err()
}

// findSkillResults searches skills and returns results using FTS5.
func (s *SQLiteStore) findSkillResults(ctx context.Context, query *SearchQuery) ([]*SearchResult, error) {
	rows, err := s.db.QueryContext(ctx,
		`SELECT sk.id, sk.name, sk.description, sk.success_count, sk.failure_count, sk.created_at
		 FROM skills sk
		 INNER JOIN skills_fts st ON sk.rowid = st.rowid
		 WHERE skills_fts MATCH ?
		 ORDER BY sk.success_count DESC LIMIT ?`,
		query.Query, query.Limit,
	)
	if err != nil {
		return nil, err
	}
	defer func() { _ = rows.Close() }()

	var results []*SearchResult
	for rows.Next() {
		skill := &Skill{}
		if err := rows.Scan(&skill.ID, &skill.Name, &skill.Description,
			&skill.SuccessCount, &skill.FailureCount, &skill.CreatedAt); err != nil {
			return nil, err
		}
		score := float64(skill.SuccessCount) / float64(skill.SuccessCount+skill.FailureCount+1)
		results = append(results, &SearchResult{
			Memory: &Memory{
				ID:        skill.ID,
				Type:      Procedural,
				Content:   skill.Description,
				Timestamp: skill.CreatedAt,
			},
			Score: score,
		})
	}
	return results, rows.Err()
}

// Cleanup removes old memories.
func (s *SQLiteStore) Cleanup(ctx context.Context, olderThan time.Duration) (int64, error) {
	if s.closed {
		return 0, ErrStoreClosed
	}

	cutoff := time.Now().Add(-olderThan)

	result, err := s.db.ExecContext(ctx,
		`DELETE FROM episodes WHERE timestamp < ?`, cutoff,
	)
	if err != nil {
		return 0, err
	}

	return result.RowsAffected()
}

// Close closes the database connection.
func (s *SQLiteStore) Close() error {
	if s.closed {
		return ErrStoreClosed
	}

	s.closed = true
	return s.db.Close()
}
