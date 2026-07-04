// Package memory provides a 3-layer memory system for Synaptic:
// - Episodic: Past sessions and conversations
// - Semantic: Facts about the user
// - Procedural: Skills and learned behaviors
//
// Memory is stored in SQLite with FTS5 for full-text search.
// All data is local-first, encrypted at rest.
package memory

import (
	"context"
	"time"
)

// Type identifies the type of memory.
type Type string

// Memory types.
const (
	Episodic   Type = "episodic"   // Past sessions and conversations
	Semantic   Type = "semantic"   // Facts about the user
	Procedural Type = "procedural" // Skills and learned behaviors

	defaultConfidence = 0.5
)

// Memory represents a single memory entry.
type Memory struct {
	ID        string
	Type      Type
	Content   string
	Metadata  map[string]interface{}
	Timestamp time.Time
	ExpiresAt *time.Time // Optional expiration
}

// Episode represents a past session or conversation.
type Episode struct {
	ID            string
	SessionID     string
	UserMessage   string
	AgentResponse string
	ActionsTaken  []string // JSON array of actions
	Timestamp     time.Time
	Summary       string
}

// Fact represents a semantic fact about the user.
type Fact struct {
	ID              string
	Category        string // preference, identity, expertise, etc.
	Content         string
	Confidence      float64 // 0.0 to 1.0
	SourceEpisodeID string
	CreatedAt       time.Time
	UpdatedAt       time.Time
}

// Skill represents a procedural memory (learned behavior).
type Skill struct {
	ID             string
	Name           string
	Description    string
	TriggerPattern string
	Steps          []string // JSON array of steps
	SuccessCount   int
	FailureCount   int
	CreatedAt      time.Time
	LastUsed       time.Time
}

// SearchQuery describes a memory search query.
type SearchQuery struct {
	Query    string
	Type     Type // Optional: filter by type
	Limit    int
	MinScore float64
}

// SearchResult is a single search result.
type SearchResult struct {
	Memory *Memory
	Score  float64
}

// Store is the interface for memory storage operations.
type Store interface {
	// Episodic memory
	StoreEpisode(ctx context.Context, episode *Episode) error
	GetEpisode(ctx context.Context, id string) (*Episode, error)
	ListEpisodes(ctx context.Context, limit int) ([]*Episode, error)
	SearchEpisodes(ctx context.Context, query string, limit int) ([]*Episode, error)

	// Semantic memory
	StoreFact(ctx context.Context, fact *Fact) error
	GetFact(ctx context.Context, id string) (*Fact, error)
	ListFacts(ctx context.Context, category string, limit int) ([]*Fact, error)
	UpdateFactConfidence(ctx context.Context, id string, confidence float64) error

	// Procedural memory
	StoreSkill(ctx context.Context, skill *Skill) error
	GetSkill(ctx context.Context, id string) (*Skill, error)
	ListSkills(ctx context.Context, limit int) ([]*Skill, error)
	IncrementSkillUsage(ctx context.Context, id string, success bool) error

	// Search
	Search(ctx context.Context, query *SearchQuery) ([]*SearchResult, error)

	// Cleanup
	Cleanup(ctx context.Context, olderThan time.Duration) (int64, error)
	Close() error
}

// StoreManager provides high-level memory operations.
type StoreManager struct {
	store Store
}

// NewManager creates a new memory manager.
func NewManager(store Store) *StoreManager {
	return &StoreManager{store: store}
}

// Remember stores a memory entry.
func (m *StoreManager) Remember(ctx context.Context, memory *Memory) error {
	if memory.Metadata == nil {
		return ErrInvalidMemoryType
	}
	switch memory.Type {
	case Episodic:
		sessionID, _ := memory.Metadata["session_id"].(string)
		episode := &Episode{
			ID:          memory.ID,
			SessionID:   sessionID,
			UserMessage: memory.Content,
			Timestamp:   memory.Timestamp,
		}
		return m.store.StoreEpisode(ctx, episode)
	case Semantic:
		category, _ := memory.Metadata["category"].(string)
		fact := &Fact{
			ID:         memory.ID,
			Category:   category,
			Content:    memory.Content,
			Confidence: defaultConfidence,
			CreatedAt:  memory.Timestamp,
			UpdatedAt:  memory.Timestamp,
		}
		return m.store.StoreFact(ctx, fact)
	case Procedural:
		name, _ := memory.Metadata["name"].(string)
		skill := &Skill{
			ID:          memory.ID,
			Name:        name,
			Description: memory.Content,
			CreatedAt:   memory.Timestamp,
		}
		return m.store.StoreSkill(ctx, skill)
	default:
		return ErrInvalidMemoryType
	}
}

// Recall searches memory for relevant entries.
func (m *StoreManager) Recall(ctx context.Context, query string, limit int) ([]*Memory, error) {
	results, err := m.store.Search(ctx, &SearchQuery{
		Query: query,
		Limit: limit,
	})
	if err != nil {
		return nil, err
	}

	memories := make([]*Memory, len(results))
	for i, result := range results {
		memories[i] = result.Memory
	}
	return memories, nil
}

// GetEpisodic retrieves past episodes.
func (m *StoreManager) GetEpisodic(ctx context.Context, limit int) ([]*Episode, error) {
	return m.store.ListEpisodes(ctx, limit)
}

// GetSemantic retrieves facts about the user.
func (m *StoreManager) GetSemantic(ctx context.Context, category string, limit int) ([]*Fact, error) {
	return m.store.ListFacts(ctx, category, limit)
}

// GetProcedural retrieves learned skills.
func (m *StoreManager) GetProcedural(ctx context.Context, limit int) ([]*Skill, error) {
	return m.store.ListSkills(ctx, limit)
}

// Cleanup removes old memories.
func (m *StoreManager) Cleanup(ctx context.Context, olderThan time.Duration) (int64, error) {
	return m.store.Cleanup(ctx, olderThan)
}

// Close closes the underlying store.
func (m *StoreManager) Close() error {
	return m.store.Close()
}
