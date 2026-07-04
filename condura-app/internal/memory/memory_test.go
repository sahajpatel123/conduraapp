package memory

import (
	"context"
	"errors"
	"os"
	"testing"
	"time"
)

func TestSQLiteStore(t *testing.T) {
	// Create a temporary database
	tmpFile, err := os.CreateTemp("", "memory-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	store, err := NewSQLiteStore(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer func() { _ = store.Close() }()

	ctx := context.Background()

	t.Run("episode operations", func(t *testing.T) {
		episode := &Episode{
			ID:            "ep1",
			SessionID:     "session1",
			UserMessage:   "Hello, how are you?",
			AgentResponse: "I'm doing well, thank you!",
			ActionsTaken:  []string{"read_file", "write_file"},
			Timestamp:     time.Now(),
			Summary:       "Greeting exchange",
		}

		// Store
		if err := store.StoreEpisode(ctx, episode); err != nil {
			t.Fatalf("failed to store episode: %v", err)
		}

		// Get
		got, err := store.GetEpisode(ctx, "ep1")
		if err != nil {
			t.Fatalf("failed to get episode: %v", err)
		}
		if got.UserMessage != episode.UserMessage {
			t.Errorf("UserMessage = %v, want %v", got.UserMessage, episode.UserMessage)
		}
		if len(got.ActionsTaken) != len(episode.ActionsTaken) {
			t.Errorf("ActionsTaken length = %d, want %d", len(got.ActionsTaken), len(episode.ActionsTaken))
		}

		// List
		episodes, err := store.ListEpisodes(ctx, 10)
		if err != nil {
			t.Fatalf("failed to list episodes: %v", err)
		}
		if len(episodes) != 1 {
			t.Errorf("expected 1 episode, got %d", len(episodes))
		}

		// Search
		results, err := store.SearchEpisodes(ctx, "Hello", 10)
		if err != nil {
			t.Fatalf("failed to search episodes: %v", err)
		}
		if len(results) != 1 {
			t.Errorf("expected 1 search result, got %d", len(results))
		}
	})

	t.Run("fact operations", func(t *testing.T) {
		fact := &Fact{
			ID:         "fact1",
			Category:   "preference",
			Content:    "User prefers dark mode",
			Confidence: 0.8,
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}

		// Store
		if err := store.StoreFact(ctx, fact); err != nil {
			t.Fatalf("failed to store fact: %v", err)
		}

		// Get
		got, err := store.GetFact(ctx, "fact1")
		if err != nil {
			t.Fatalf("failed to get fact: %v", err)
		}
		if got.Content != fact.Content {
			t.Errorf("Content = %v, want %v", got.Content, fact.Content)
		}
		if got.Confidence != fact.Confidence {
			t.Errorf("Confidence = %v, want %v", got.Confidence, fact.Confidence)
		}

		// List by category
		facts, err := store.ListFacts(ctx, "preference", 10)
		if err != nil {
			t.Fatalf("failed to list facts: %v", err)
		}
		if len(facts) != 1 {
			t.Errorf("expected 1 fact, got %d", len(facts))
		}

		// Update confidence
		if err := store.UpdateFactConfidence(ctx, "fact1", 0.9); err != nil {
			t.Fatalf("failed to update confidence: %v", err)
		}
		got, _ = store.GetFact(ctx, "fact1")
		if got.Confidence != 0.9 {
			t.Errorf("Confidence = %v, want 0.9", got.Confidence)
		}
	})

	t.Run("skill operations", func(t *testing.T) {
		skill := &Skill{
			ID:             "skill1",
			Name:           "File Reader",
			Description:    "Reads files from the filesystem",
			TriggerPattern: "read file",
			Steps:          []string{"open file", "read contents", "close file"},
			CreatedAt:      time.Now(),
			LastUsed:       time.Now(),
		}

		// Store
		if err := store.StoreSkill(ctx, skill); err != nil {
			t.Fatalf("failed to store skill: %v", err)
		}

		// Get
		got, err := store.GetSkill(ctx, "skill1")
		if err != nil {
			t.Fatalf("failed to get skill: %v", err)
		}
		if got.Name != skill.Name {
			t.Errorf("Name = %v, want %v", got.Name, skill.Name)
		}
		if len(got.Steps) != len(skill.Steps) {
			t.Errorf("Steps length = %d, want %d", len(got.Steps), len(skill.Steps))
		}

		// List
		skills, err := store.ListSkills(ctx, 10)
		if err != nil {
			t.Fatalf("failed to list skills: %v", err)
		}
		if len(skills) != 1 {
			t.Errorf("expected 1 skill, got %d", len(skills))
		}

		// Increment usage
		if err := store.IncrementSkillUsage(ctx, "skill1", true); err != nil {
			t.Fatalf("failed to increment usage: %v", err)
		}
		got, _ = store.GetSkill(ctx, "skill1")
		if got.SuccessCount != 1 {
			t.Errorf("SuccessCount = %d, want 1", got.SuccessCount)
		}
	})

	t.Run("search across types", func(t *testing.T) {
		results, err := store.Search(ctx, &SearchQuery{
			Query: "file",
			Limit: 10,
		})
		if err != nil {
			t.Fatalf("failed to search: %v", err)
		}
		if len(results) == 0 {
			t.Error("expected search results, got none")
		}
	})

	t.Run("cleanup", func(t *testing.T) {
		// Store an old episode
		oldEpisode := &Episode{
			ID:        "old",
			SessionID: "session_old",
			Timestamp: time.Now().Add(-24 * time.Hour),
		}
		if err := store.StoreEpisode(ctx, oldEpisode); err != nil {
			t.Fatalf("failed to store old episode: %v", err)
		}

		// Cleanup
		deleted, err := store.Cleanup(ctx, 12*time.Hour)
		if err != nil {
			t.Fatalf("failed to cleanup: %v", err)
		}
		if deleted != 1 {
			t.Errorf("expected 1 deleted, got %d", deleted)
		}
	})
}

func TestManager(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "memory-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	store, err := NewSQLiteStore(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}
	defer func() { _ = store.Close() }()

	manager := NewManager(store)
	ctx := context.Background()

	t.Run("remember and recall", func(t *testing.T) {
		memory := &Memory{
			ID:        "mem1",
			Type:      Semantic,
			Content:   "User likes coffee",
			Metadata:  map[string]interface{}{"category": "preference"},
			Timestamp: time.Now(),
		}

		if err := manager.Remember(ctx, memory); err != nil {
			t.Fatalf("failed to remember: %v", err)
		}

		memories, err := manager.Recall(ctx, "coffee", 10)
		if err != nil {
			t.Fatalf("failed to recall: %v", err)
		}
		if len(memories) == 0 {
			t.Error("expected memories, got none")
		}
	})
}

func TestValidation(t *testing.T) {
	tmpFile, err := os.CreateTemp("", "memory-test-*.db")
	if err != nil {
		t.Fatalf("failed to create temp file: %v", err)
	}
	_ = tmpFile.Close()
	defer func() { _ = os.Remove(tmpFile.Name()) }()

	store, err := NewSQLiteStore(tmpFile.Name())
	if err != nil {
		t.Fatalf("failed to create store: %v", err)
	}

	ctx := context.Background()

	t.Run("invalid confidence", func(t *testing.T) {
		fact := &Fact{
			ID:         "invalid",
			Category:   "test",
			Content:    "test",
			Confidence: 1.5, // Invalid
			CreatedAt:  time.Now(),
			UpdatedAt:  time.Now(),
		}
		if err := store.StoreFact(ctx, fact); !errors.Is(err, ErrInvalidConfidence) {
			t.Errorf("expected ErrInvalidConfidence, got %v", err)
		}
	})

	t.Run("closed store", func(t *testing.T) {
		_ = store.Close()
		_, err := store.ListEpisodes(ctx, 10)
		if !errors.Is(err, ErrStoreClosed) {
			t.Errorf("expected ErrStoreClosed, got %v", err)
		}
	})
}
