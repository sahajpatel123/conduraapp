package daemon

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/adaptive"
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
	observer   *adaptive.Observer
	engine     *adaptive.Engine

	skillPatterns   map[string]int
	skillPatternsMu sync.Mutex
}

// SetObserver wires the adaptive engine's observer.
func (e *PostSessionExtractor) SetObserver(o *adaptive.Observer) {
	e.observer = o
}

// SetEngine wires the adaptive engine for post-session analysis.
func (e *PostSessionExtractor) SetEngine(eng *adaptive.Engine) {
	e.engine = eng
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

	// Fire observer for adaptive engine (user-initiated evidence).
	if e.observer != nil {
		//nolint:gosec // intentional: async observer must survive request ctx
		go e.observer.Record(context.Background(), adaptive.Observation{
			SessionID:     newID("sess"),
			UserQuery:     query,
			AgentReply:    reply,
			UserInitiated: true,
		})
	}

	// Trigger adaptive engine analysis (async, best-effort).
	if e.engine != nil {
		//nolint:gosec // intentional: async engine must survive request ctx
		go e.engine.Run(context.Background())
	}

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
	// Skill auto-creation: when N similar session patterns appear,
	// create a skill from the cluster. The adaptive engine's
	// Observer already tracks sessions — we use a simple in-memory
	// pattern counter as a lightweight trigger.
	if e.skills == nil {
		return
	}

	pattern := normalizePattern(query)
	if pattern == "" || len(pattern) < 10 {
		return
	}

	e.skillPatternsMu.Lock()
	if e.skillPatterns == nil {
		e.skillPatterns = make(map[string]int)
	}
	e.skillPatterns[pattern]++
	count := e.skillPatterns[pattern]
	e.skillPatternsMu.Unlock()

	const minSamples = 3
	if count < minSamples {
		return
	}

	// Create the skill.
	sk := &skills.Skill{
		ID:             newID("skill"),
		Name:           pattern,
		Description:    "Auto-created skill for: " + query,
		TriggerPattern: pattern,
		Steps:          []string{reply},
		Version:        "0.1.0",
		Trust:          skills.TrustExperimental,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
		LastUsed:       time.Now(),
	}
	if err := e.skills.Create(ctx, sk); err != nil {
		e.log.Warn("postsession: skill create failed", "err", err, "pattern", pattern)
		return
	}
	// Reset counter so the same pattern doesn't keep creating skills.
	e.skillPatternsMu.Lock()
	delete(e.skillPatterns, pattern)
	e.skillPatternsMu.Unlock()
	e.log.Info("postsession: skill auto-created", "pattern", pattern, "samples", count)
}

func normalizePattern(query string) string {
	// Extract the core intent: lowercase, strip punctuation, limit length.
	q := strings.ToLower(strings.TrimSpace(query))
	// Truncate to first 80 chars as the pattern key.
	const maxPatternLen = 80
	if len(q) > maxPatternLen {
		q = q[:80]
	}
	// Normalize whitespace.
	q = strings.Join(strings.Fields(q), " ")
	return q
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
