package daemon

import (
	"context"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/adaptive"
	"github.com/sahajpatel123/synapticapp/internal/memory"
	"github.com/sahajpatel123/synapticapp/internal/skills"
)

// PostSessionExtractor runs async memory extraction and skill
// auto-creation after a session completes.
type PostSessionExtractor struct {
	memory     *memory.StoreManager
	autoCreate *skills.AutoCreate
	log        *slog.Logger
	enabled    bool
	skillStore io.Closer
	observer   *adaptive.Observer
	engine     *adaptive.Engine
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
	var ac *skills.AutoCreate
	var closer io.Closer
	if skillStore != nil {
		ac = skills.NewAutoCreate(skillStore)
		if c, ok := skillStore.(io.Closer); ok {
			closer = c
		}
	}
	return &PostSessionExtractor{
		memory:     mem,
		autoCreate: ac,
		log:        log,
		enabled:    enabled,
		skillStore: closer,
	}
}

// Close releases resources held by the extractor.
func (e *PostSessionExtractor) Close() error {
	if e.skillStore != nil {
		return e.skillStore.Close()
	}
	return nil
}

// AfterSession is called after a session completes.
func (e *PostSessionExtractor) AfterSession(ctx context.Context, userMessage, assistantReply string, conversationID int64) {
	if !e.enabled {
		return
	}

	query := userMessage
	reply := assistantReply

	if e.observer != nil {
		go e.observer.Record(context.Background(), adaptive.Observation{
			SessionID:     newID("sess"),
			UserQuery:     query,
			AgentReply:    reply,
			UserInitiated: true,
		})
	}

	if e.engine != nil {
		go e.engine.Run(context.Background())
	}

	if e.memory != nil {
		go e.storeEpisode(ctx, query, reply, conversationID)
	}

	if e.autoCreate != nil {
		go e.runAutoCreate(ctx, query, reply, conversationID)
	}
}

func (e *PostSessionExtractor) storeEpisode(ctx context.Context, query, reply string, conversationID int64) {
	epID := newID("ep")
	err := e.memory.Remember(ctx, &memory.Memory{
		ID:      epID,
		Type:    memory.Episodic,
		Content: query,
		Metadata: map[string]interface{}{
			"session_id":        epID,
			keyConversationID:   conversationID,
			"reply":             reply,
		},
	})
	if err != nil {
		e.log.Warn("postsession: store episode failed", "err", err)
	}
	e.extractFacts(ctx, query, reply)
}

func (e *PostSessionExtractor) extractFacts(ctx context.Context, query, reply string) {
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

func extractPreference(query, reply string) string {
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

func (e *PostSessionExtractor) runAutoCreate(ctx context.Context, query, reply string, conversationID int64) {
	sessionID := fmt.Sprintf("conv-%d", conversationID)
	steps := []string{reply}
	err := e.autoCreate.Observe(ctx, sessionID, query, steps)
	switch {
	case err == nil:
		e.log.Info("postsession: skill auto-created", "session", sessionID)
	case errors.Is(err, skills.ErrNoSkillCreated), errors.Is(err, skills.ErrEmptyQuery):
		// below threshold — not an error
	default:
		e.log.Warn("postsession: autocreate failed", "err", err)
	}
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
