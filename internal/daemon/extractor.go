package daemon

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/memory"
	"github.com/sahajpatel123/synapticapp/internal/skills"
)

// PostSessionExtractor runs async memory extraction and skill
// auto-creation after a session completes. Both operations are
// fire-and-forget, best-effort, and config-gated.
type PostSessionExtractor struct {
	memory     *memory.StoreManager
	skills     skills.Store
	log        *slog.Logger
	enabled    bool
	skillStore io.Closer
}

// NewPostSessionExtractor creates an async post-session processor.
func NewPostSessionExtractor(mem *memory.StoreManager, skillStore skills.Store, log *slog.Logger, enabled bool) *PostSessionExtractor {
	return &PostSessionExtractor{
		memory:     mem,
		skills:     skillStore,
		log:        log,
		enabled:    enabled,
		skillStore: skillStore.(io.Closer),
	}
}

// Close releases resources held by the extractor.
func (e *PostSessionExtractor) Close() error {
	if e.skillStore != nil {
		return e.skillStore.Close()
	}
	return nil
}

// AfterSession is called after a session completes. It fires
// async goroutines for memory extraction and skill creation.
// Never blocks the session return.
func (e *PostSessionExtractor) AfterSession(ctx context.Context, userMessage, assistantReply string, conversationID int64) {
	if !e.enabled {
		return
	}

	// Copy values for the goroutines since the caller may reuse.
	query := userMessage
	reply := assistantReply

	if e.memory != nil {
		go e.storeEpisode(ctx, query, reply, conversationID)
	}

	if e.skills != nil {
		go e.maybeCreateSkill(ctx, query, reply)
	}
}

func (e *PostSessionExtractor) storeEpisode(ctx context.Context, query, reply string, conversationID int64) {
	epID := newID("ep")
	err := e.memory.Remember(ctx, &memory.Memory{
		ID:      epID,
		Type:    memory.Episodic,
		Content: query,
		Metadata: map[string]interface{}{
			"session_id":      epID,
			keyConversationID: conversationID,
			"reply":           reply,
		},
	})
	if err != nil {
		e.log.Warn("postsession: store episode failed", "err", err)
	}

	// Also extract semantic facts from the reply.
	e.extractFacts(ctx, query, reply)
}

func (e *PostSessionExtractor) extractFacts(ctx context.Context, query, reply string) {
	// Simple heuristic extraction: look for preference/keyword patterns.
	// Full LLM-based extraction requires a provider (not wired yet).
	fact := extractPreference(query, reply)
	if fact != "" {
		err := e.memory.Remember(ctx, &memory.Memory{
			ID:      newID("fact"),
			Type:    memory.Semantic,
			Content: fact,
			Metadata: map[string]interface{}{
				"category": "preference",
			},
		})
		if err != nil {
			e.log.Warn("postsession: store fact failed", "err", err)
		}
	}
}

// extractPreference uses simple heuristics to detect user preferences.
func extractPreference(query, reply string) string {
	// Detect "I prefer/like/want/use" patterns.
	combined := query + " " + reply
	for _, marker := range []string{"I prefer ", "I like ", "I use ", "I want ", "my favorite "} {
		if idx := idxOf(combined, marker); idx >= 0 {
			end := idxOf(combined[idx+len(marker):], ".")
			if end < 0 {
				end = idxOf(combined[idx+len(marker):], ",")
			}
			if end < 0 {
				end = len(combined) - idx - len(marker)
			}
			return marker + combined[idx+len(marker):idx+len(marker)+end]
		}
	}
	return ""
}

func (e *PostSessionExtractor) maybeCreateSkill(ctx context.Context, query, reply string) {
	// Check if we have enough similar sessions to auto-create a skill.
	// For now, this is a placeholder — real implementation requires
	// tracking session similarity across multiple calls.
	_ = ctx
	_ = query
	_ = reply
}

func idxOf(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func newID(prefix string) string {
	return fmt.Sprintf("%s-%d", prefix, time.Now().UnixNano())
}
