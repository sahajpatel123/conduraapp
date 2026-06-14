// AutoCreate is Phase 11 sub-phase 11F — the skills auto-create
// pipeline. After a session ends, the post-session extractor
// hands the transcript to AutoCreate. If a normalized trigger
// pattern appears in the user's natural-language command at
// least MinSamples (3) times across distinct sessions, a
// community-trust skill is created. We never auto-create
// official-trust skills — promotion to official requires a
// human review pass (see Skills Hub, Phase 12C).
//
// The pending map is bounded by MaxPending; least-recently-used
// triggers are evicted. This prevents a malicious or
// pathological user (or a runaway test) from filling memory.
package skills

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"
)

// Tunables. Exported as constants so tests can reference them.
const (
	// MinSamples is the threshold of session-end observations
	// required before a skill is auto-created. Per the spec.
	MinSamples = 3
	// MaxPending bounds the trigger map. Beyond this, the
	// least-recently-used entry is evicted.
	MaxPending = 256
	// MaxSteps is a soft cap on the captured step list of an
	// auto-created skill. Beyond this we just keep the first N.
	MaxSteps = 32
)

// Sentinel errors. Callers can errors.Is() against these.
var (
	// ErrNoSkillCreated is returned by Observe when the
	// observation didn't reach MinSamples. This is NOT a
	// failure — the extractor filters it out and keeps
	// running. We expose it as a sentinel so tests can
	// distinguish "still building evidence" from real errors.
	ErrNoSkillCreated = errors.New("skills: no skill created (below threshold)")

	// ErrEmptyQuery is returned when the user query is empty or
	// whitespace-only. Treated as a soft no-op.
	ErrEmptyQuery = errors.New("skills: empty query")

	// ErrStoreMissing is returned when AutoCreate was
	// constructed without a Store. Constructors should
	// enforce this at boot; this sentinel exists for tests
	// and for callers that introspect errors.
	ErrStoreMissing = errors.New("skills: no store configured")
)

// AutoCreate is the auto-create pipeline. It is safe for
// concurrent Observe calls.
type AutoCreate struct {
	store Store // may be nil during construction, before SetStore
	mu    sync.Mutex
	// pending[normalizedTrigger] is the list of (sessionID,
	// query) observations waiting to reach MinSamples. Bounded
	// to MaxPending by LRU eviction.
	pending map[string]*pendingEntry
	// lru is the eviction order. Front is oldest, back is
	// newest. We move a key to the back on each touch.
	lru []string
}

type pendingEntry struct {
	queries   []observation
	firstSeen time.Time
	lastSeen  time.Time
	lruIndex  int // index into AutoCreate.lru
}

type observation struct {
	sessionID string
	query     string
	steps     []string
}

// NewAutoCreate returns an AutoCreate ready to Observe. If
// store is nil, the constructor still succeeds but every
// Observe returns ErrStoreMissing. This is intentional: the
// daemon may construct AutoCreate before the skill store
// finishes opening.
func NewAutoCreate(store Store) *AutoCreate {
	return &AutoCreate{
		store:   store,
		pending: make(map[string]*pendingEntry),
		lru:     make([]string, 0, MaxPending),
	}
}

// SetStore wires the skill store. Call this from the daemon
// once the store has opened. Calling with nil is a no-op.
func (a *AutoCreate) SetStore(s Store) {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.store = s
}

// Observe records one session-end observation. If the
// normalized trigger reaches MinSamples for the first time,
// a community-trust skill is created and stored. Subsequent
// observations of the same trigger increment usage counters.
//
// Returns ErrNoSkillCreated if the threshold wasn't reached,
// ErrEmptyQuery if query is empty, or nil on success or
// already-created.
func (a *AutoCreate) Observe(ctx context.Context, sessionID, query string, steps []string) error {
	if strings.TrimSpace(query) == "" {
		return ErrEmptyQuery
	}
	if sessionID == "" {
		// Defensive: never let an empty sessionID enter the
		// pending map (would cause key collisions on reset).
		sessionID = fmt.Sprintf("anon-%d", time.Now().UnixNano())
	}
	trigger := normalize(query)
	if trigger == "" {
		return ErrEmptyQuery
	}
	a.mu.Lock()
	defer a.mu.Unlock()
	if a.store == nil {
		return ErrStoreMissing
	}
	now := time.Now().UTC()
	// Already a skill for this trigger? Just bump its usage.
	existing, err := a.findByTrigger(ctx, trigger)
	if err != nil {
		return fmt.Errorf("skills: lookup: %w", err)
	}
	if existing != nil {
		// Increment success count and update last_used.
		_ = a.store.IncrementUsage(ctx, existing.ID, true)
		// Also touch the pending entry so we don't keep
		// building evidence unnecessarily. (We don't
		// actively trim the map, but the LRU eviction
		// will eventually drop it.)
		a.touchLocked(trigger)
		return nil
	}
	// Not yet a skill — accumulate evidence.
	entry, ok := a.pending[trigger]
	if !ok {
		if len(a.pending) >= MaxPending {
			a.evictLRULocked()
		}
		entry = &pendingEntry{
			firstSeen: now,
			lruIndex:  len(a.lru),
		}
		a.pending[trigger] = entry
		a.lru = append(a.lru, trigger)
	} else {
		a.touchLocked(trigger)
	}
	entry.queries = append(entry.queries, observation{
		sessionID: sessionID,
		query:     query,
		steps:     steps,
	})
	entry.lastSeen = now
	if len(entry.queries) < MinSamples {
		return ErrNoSkillCreated
	}
	// Threshold reached. Create the skill.
	sk := a.buildSkill(trigger, entry.queries)
	if err := a.store.Create(ctx, sk); err != nil {
		// Roll back the trigger accumulation so the next
		// observation can retry. (We don't delete pending;
		// we just decrement so the user doesn't have to wait
		// for 3 more sessions after a transient DB error.)
		if len(entry.queries) > 0 {
			entry.queries = entry.queries[:len(entry.queries)-1]
		}
		return fmt.Errorf("skills: create: %w", err)
	}
	// Clean up the pending entry — the skill is now persisted
	// and we'll find it via findByTrigger next time.
	delete(a.pending, trigger)
	a.removeFromLRULocked(trigger)
	return nil
}

// PendingCount returns the number of triggers currently
// accumulating evidence. Used by tests and by the GUI's
// "Skills being learned" indicator.
func (a *AutoCreate) PendingCount() int {
	a.mu.Lock()
	defer a.mu.Unlock()
	return len(a.pending)
}

// Reset clears the pending map. Used by tests and by the
// "forget learned triggers" user action.
func (a *AutoCreate) Reset() {
	a.mu.Lock()
	defer a.mu.Unlock()
	a.pending = make(map[string]*pendingEntry)
	a.lru = a.lru[:0]
}

// findByTrigger checks whether a skill with this normalized
// trigger already exists. Caller must hold a.mu.
func (a *AutoCreate) findByTrigger(ctx context.Context, trigger string) (*Skill, error) {
	// Search is a LIKE on name+trigger_pattern. We store the
	// normalized form in trigger_pattern, so a literal match
	// on the normalized query is what we want.
	hits, err := a.store.Search(ctx, trigger, 5)
	if err != nil {
		return nil, err
	}
	for _, s := range hits {
		if s.TriggerPattern == trigger {
			return s, nil
		}
	}
	return nil, nil
}

// buildSkill constructs a Skill from accumulated observations.
// Caller must hold a.mu. We pick the longest step list (the
// most complete run) and dedupe adjacent identical steps.
func (a *AutoCreate) buildSkill(trigger string, obs []observation) *Skill {
	now := time.Now().UTC()
	// Pick the observation with the most steps.
	best := obs[0]
	for _, o := range obs[1:] {
		if len(o.steps) > len(best.steps) {
			best = o
		}
	}
	steps := dedupeAdjacent(best.steps)
	if len(steps) > MaxSteps {
		steps = steps[:MaxSteps]
	}
	// ID: a content hash of trigger + best.query, plus a
	// timestamp suffix to allow re-creation after a Reset
	// without colliding with the prior auto-created row.
	idSeed := trigger + "|" + best.query + "|" + now.Format(time.RFC3339Nano)
	idHash := sha256.Sum256([]byte(idSeed))
	id := "skill-auto-" + hex.EncodeToString(idHash[:8])
	return &Skill{
		ID:             id,
		Name:           humanize(trigger),
		Description:    fmt.Sprintf("Auto-created from %d observations of user pattern.", len(obs)),
		Version:        "0.1.0",
		Trust:          TrustCommunity, // NEVER auto-official.
		TriggerPattern: trigger,
		Steps:          steps,
		SuccessCount:   len(obs),
		CreatedAt:      now,
		UpdatedAt:      now,
		LastUsed:       now,
		Source:         "auto",
		Author:         "synaptic-auto",
	}
}

// touchLocked moves a trigger to the back of the LRU list.
// Caller must hold a.mu.
func (a *AutoCreate) touchLocked(trigger string) {
	entry, ok := a.pending[trigger]
	if !ok {
		return
	}
	// Remove the old lru entry.
	a.removeFromLRULocked(trigger)
	// Append to back.
	entry.lruIndex = len(a.lru)
	a.lru = append(a.lru, trigger)
}

// removeFromLRULocked drops a trigger from the LRU list
// regardless of its current position. Caller must hold a.mu.
func (a *AutoCreate) removeFromLRULocked(trigger string) {
	for i, t := range a.lru {
		if t == trigger {
			a.lru = append(a.lru[:i], a.lru[i+1:]...)
			// Fix lruIndex on the moved-back element, if any.
			if i < len(a.lru) {
				if e, ok := a.pending[a.lru[i]]; ok {
					e.lruIndex = i
				}
			}
			return
		}
	}
}

// evictLRULocked drops the oldest pending entry. Caller must
// hold a.mu and have already verified len(pending) >= MaxPending.
func (a *AutoCreate) evictLRULocked() {
	if len(a.lru) == 0 {
		return
	}
	oldest := a.lru[0]
	a.lru = a.lru[1:]
	delete(a.pending, oldest)
	// Fix indices on the rest.
	for i, t := range a.lru {
		if e, ok := a.pending[t]; ok {
			e.lruIndex = i
		}
	}
}

// normalize extracts a stable, lowercase, whitespace-collapsed
// trigger phrase from a free-form user query. We don't try to
// be clever — the goal is to bucket "open chrome", "Open Chrome",
// and "  open   chrome  " into the same bucket.
func normalize(query string) string {
	q := strings.ToLower(strings.TrimSpace(query))
	if q == "" {
		return ""
	}
	// Collapse internal whitespace.
	var b strings.Builder
	prevSpace := false
	for _, r := range q {
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			if !prevSpace {
				b.WriteByte(' ')
			}
			prevSpace = true
			continue
		}
		prevSpace = false
		b.WriteRune(r)
	}
	// Cap at 80 chars to keep the trigger_pattern column small.
	out := strings.TrimSpace(b.String())
	if len(out) > 80 {
		out = out[:80]
	}
	return out
}

// humanize turns a normalized trigger into a Title-Case skill
// name suitable for display in the Skills Hub.
func humanize(trigger string) string {
	parts := strings.Fields(trigger)
	for i, p := range parts {
		if p == "" {
			continue
		}
		parts[i] = strings.ToUpper(p[:1]) + p[1:]
	}
	return strings.Join(parts, " ")
}

// dedupeAdjacent collapses consecutive identical strings.
func dedupeAdjacent(steps []string) []string {
	if len(steps) == 0 {
		return steps
	}
	out := make([]string, 0, len(steps))
	out = append(out, steps[0])
	for _, s := range steps[1:] {
		if s != out[len(out)-1] {
			out = append(out, s)
		}
	}
	return out
}
