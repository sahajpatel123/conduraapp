package memory

import (
	"context"
	"errors"
	"testing"
	"time"
)

// fakeStore is a function-field-based spy that implements the Store
// interface. Methods called in the production code path under test
// route to the configured function field (capturing args for assertion);
// methods not exercised return zero values.
type fakeStore struct {
	listEpisodes func(ctx context.Context, limit int) ([]*Episode, error)
	listFacts    func(ctx context.Context, category string, limit int) ([]*Fact, error)
	listSkills   func(ctx context.Context, limit int) ([]*Skill, error)
	cleanup      func(ctx context.Context, olderThan time.Duration) (int64, error)
	close        func() error
	storeEpisode func(ctx context.Context, episode *Episode) error
	storeFact    func(ctx context.Context, fact *Fact) error
	storeSkill   func(ctx context.Context, skill *Skill) error

	// Captured args (populated when the corresponding function field
	// is called and a *value* parameter is filled in below).
	gotListEpisodesLimit int
	gotListFactsCategory string
	gotListFactsLimit    int
	gotListSkillsLimit   int
	gotCleanupDuration   time.Duration
	gotStoreEpisode      *Episode
	gotStoreFact         *Fact
	gotStoreSkill        *Skill
	gotCleanupCallCount  int
	gotCloseCallCount    int
}

func (f *fakeStore) ListEpisodes(ctx context.Context, limit int) ([]*Episode, error) {
	f.gotListEpisodesLimit = limit
	if f.listEpisodes != nil {
		return f.listEpisodes(ctx, limit)
	}
	return nil, nil
}

func (f *fakeStore) ListFacts(ctx context.Context, category string, limit int) ([]*Fact, error) {
	f.gotListFactsCategory = category
	f.gotListFactsLimit = limit
	if f.listFacts != nil {
		return f.listFacts(ctx, category, limit)
	}
	return nil, nil
}

func (f *fakeStore) ListSkills(ctx context.Context, limit int) ([]*Skill, error) {
	f.gotListSkillsLimit = limit
	if f.listSkills != nil {
		return f.listSkills(ctx, limit)
	}
	return nil, nil
}

func (f *fakeStore) Cleanup(ctx context.Context, olderThan time.Duration) (int64, error) {
	f.gotCleanupDuration = olderThan
	f.gotCleanupCallCount++
	if f.cleanup != nil {
		return f.cleanup(ctx, olderThan)
	}
	return 0, nil
}

func (f *fakeStore) Close() error {
	f.gotCloseCallCount++
	if f.close != nil {
		return f.close()
	}
	return nil
}

func (f *fakeStore) StoreEpisode(ctx context.Context, episode *Episode) error {
	f.gotStoreEpisode = episode
	if f.storeEpisode != nil {
		return f.storeEpisode(ctx, episode)
	}
	return nil
}

func (f *fakeStore) StoreFact(ctx context.Context, fact *Fact) error {
	f.gotStoreFact = fact
	if f.storeFact != nil {
		return f.storeFact(ctx, fact)
	}
	return nil
}

func (f *fakeStore) StoreSkill(ctx context.Context, skill *Skill) error {
	f.gotStoreSkill = skill
	if f.storeSkill != nil {
		return f.storeSkill(ctx, skill)
	}
	return nil
}

// Unused-by-Manager-methods: return a sentinel error so any test
// that accidentally calls them fails loudly. Listed explicitly
// (instead of via embedded panic-stub) so the test file compiles even
// if the Store interface grows new methods.
var errFakeStoreUnimplemented = errors.New("fakeStore: method not implemented in this test")

func (f *fakeStore) GetEpisode(ctx context.Context, id string) (*Episode, error) {
	return nil, errFakeStoreUnimplemented
}
func (f *fakeStore) SearchEpisodes(ctx context.Context, query string, limit int) ([]*Episode, error) {
	return nil, errFakeStoreUnimplemented
}
func (f *fakeStore) GetFact(ctx context.Context, id string) (*Fact, error) {
	return nil, errFakeStoreUnimplemented
}
func (f *fakeStore) UpdateFactConfidence(ctx context.Context, id string, confidence float64) error {
	return errFakeStoreUnimplemented
}
func (f *fakeStore) GetSkill(ctx context.Context, id string) (*Skill, error) {
	return nil, errFakeStoreUnimplemented
}
func (f *fakeStore) IncrementSkillUsage(ctx context.Context, id string, success bool) error {
	return errFakeStoreUnimplemented
}
func (f *fakeStore) Search(ctx context.Context, query *SearchQuery) ([]*SearchResult, error) {
	return nil, errFakeStoreUnimplemented
}

// --- Tests -----------------------------------------------------------------

// TestStoreManager_GetEpisodic_DelegatesLimit pins the delegation
// contract: StoreManager.GetEpisodic(ctx, limit) MUST call
// Store.ListEpisodes(ctx, limit) with the same limit and return its
// result. A regression that drops the limit (e.g., hardcodes 0 or 100)
// would silently truncate or bloat the episodic retrieval.
func TestStoreManager_GetEpisodic_DelegatesLimit(t *testing.T) {
	wantEpisodes := []*Episode{{ID: "e1"}, {ID: "e2"}}
	store := &fakeStore{
		listEpisodes: func(ctx context.Context, limit int) ([]*Episode, error) {
			return wantEpisodes, nil
		},
	}
	mgr := NewManager(store)

	got, err := mgr.GetEpisodic(context.Background(), 50)
	if err != nil {
		t.Fatalf("GetEpisodic: %v", err)
	}
	if store.gotListEpisodesLimit != 50 {
		t.Errorf("ListEpisodes called with limit = %d, want 50", store.gotListEpisodesLimit)
	}
	if len(got) != 2 || got[0].ID != "e1" || got[1].ID != "e2" {
		t.Errorf("GetEpisodic returned %v, want the same slice the store returned", got)
	}
}

// TestStoreManager_GetSemantic_DelegatesCategoryAndLimit pins the
// two-arg delegation contract: StoreManager.GetSemantic(ctx,
// category, limit) MUST call Store.ListFacts(ctx, category, limit)
// with both args unchanged. A regression that hardcodes category to
// "" or swaps the arg order would silently break category-filtered
// fact retrieval.
func TestStoreManager_GetSemantic_DelegatesCategoryAndLimit(t *testing.T) {
	wantFacts := []*Fact{{ID: "f1", Category: "preference"}}
	store := &fakeStore{
		listFacts: func(ctx context.Context, category string, limit int) ([]*Fact, error) {
			return wantFacts, nil
		},
	}
	mgr := NewManager(store)

	got, err := mgr.GetSemantic(context.Background(), "preference", 20)
	if err != nil {
		t.Fatalf("GetSemantic: %v", err)
	}
	if store.gotListFactsCategory != "preference" {
		t.Errorf("ListFacts called with category = %q, want \"preference\"", store.gotListFactsCategory)
	}
	if store.gotListFactsLimit != 20 {
		t.Errorf("ListFacts called with limit = %d, want 20", store.gotListFactsLimit)
	}
	if len(got) != 1 || got[0].ID != "f1" {
		t.Errorf("GetSemantic returned %v, want the same slice the store returned", got)
	}
}

// TestStoreManager_GetProcedural_DelegatesLimit pins the
// single-arg delegation contract: StoreManager.GetProcedural(ctx,
// limit) MUST call Store.ListSkills(ctx, limit) with the same limit.
func TestStoreManager_GetProcedural_DelegatesLimit(t *testing.T) {
	wantSkills := []*Skill{{ID: "s1", Name: "summarize"}}
	store := &fakeStore{
		listSkills: func(ctx context.Context, limit int) ([]*Skill, error) {
			return wantSkills, nil
		},
	}
	mgr := NewManager(store)

	got, err := mgr.GetProcedural(context.Background(), 30)
	if err != nil {
		t.Fatalf("GetProcedural: %v", err)
	}
	if store.gotListSkillsLimit != 30 {
		t.Errorf("ListSkills called with limit = %d, want 30", store.gotListSkillsLimit)
	}
	if len(got) != 1 || got[0].ID != "s1" {
		t.Errorf("GetProcedural returned %v, want the same slice the store returned", got)
	}
}

// TestStoreManager_Cleanup_DelegatesDuration pins the
// time.Duration delegation contract: StoreManager.Cleanup(ctx,
// olderThan) MUST call Store.Cleanup(ctx, olderThan) with the same
// duration and return its (count, error). A regression that flipped
// the duration sign or used a hardcoded value would either delete
// nothing (negative durations become no-ops) or delete everything
// (very long durations).
func TestStoreManager_Cleanup_DelegatesDuration(t *testing.T) {
	store := &fakeStore{
		cleanup: func(ctx context.Context, olderThan time.Duration) (int64, error) {
			return 42, nil
		},
	}
	mgr := NewManager(store)

	count, err := mgr.Cleanup(context.Background(), 24*time.Hour)
	if err != nil {
		t.Fatalf("Cleanup: %v", err)
	}
	if store.gotCleanupDuration != 24*time.Hour {
		t.Errorf("Cleanup called with duration = %v, want 24h", store.gotCleanupDuration)
	}
	if store.gotCleanupCallCount != 1 {
		t.Errorf("Cleanup called %d times, want 1", store.gotCleanupCallCount)
	}
	if count != 42 {
		t.Errorf("Cleanup returned count = %d, want 42", count)
	}
}

// TestStoreManager_Close_DelegatesToStore pins the Close contract:
// StoreManager.Close() MUST call Store.Close() exactly once and
// propagate any returned error. A regression that swallowed the
// error would leak file descriptors / DB connections; a regression
// that called Close twice would double-release resources.
func TestStoreManager_Close_DelegatesToStore(t *testing.T) {
	store := &fakeStore{
		close: func() error {
			return errors.New("close-failed")
		},
	}
	mgr := NewManager(store)

	err := mgr.Close()
	if err == nil || err.Error() != "close-failed" {
		t.Errorf("Close returned %v, want error \"close-failed\"", err)
	}
	if store.gotCloseCallCount != 1 {
		t.Errorf("Close called %d times, want 1", store.gotCloseCallCount)
	}
}

// TestStoreManager_Remember_NilMetadata pins the input-validation
// guard: Remember() MUST reject memories with nil metadata BEFORE
// any store call. A regression that allowed nil metadata would panic
// inside the switch (metadata["session_id"] on nil map).
func TestStoreManager_Remember_NilMetadata(t *testing.T) {
	store := &fakeStore{}
	mgr := NewManager(store)

	err := mgr.Remember(context.Background(), &Memory{
		ID:        "m1",
		Type:      Episodic,
		Content:   "hello",
		Metadata:  nil,
		Timestamp: time.Now(),
	})
	if !errors.Is(err, ErrInvalidMemoryType) {
		t.Errorf("Remember(nil metadata) = %v, want ErrInvalidMemoryType", err)
	}
	if store.gotStoreEpisode != nil || store.gotStoreFact != nil || store.gotStoreSkill != nil {
		t.Errorf("Remember(nil metadata) made a store call: episode=%v fact=%v skill=%v",
			store.gotStoreEpisode, store.gotStoreFact, store.gotStoreSkill)
	}
}

// TestStoreManager_Remember_InvalidType pins the default-branch
// guard: Remember() MUST reject memories with an unknown Type BEFORE
// any store call. A regression that allowed unknown types would
// silently store nothing (no error, no row).
func TestStoreManager_Remember_InvalidType(t *testing.T) {
	store := &fakeStore{}
	mgr := NewManager(store)

	err := mgr.Remember(context.Background(), &Memory{
		ID:        "m1",
		Type:      Type("invalid_type"),
		Content:   "hello",
		Metadata:  map[string]interface{}{},
		Timestamp: time.Now(),
	})
	if !errors.Is(err, ErrInvalidMemoryType) {
		t.Errorf("Remember(invalid type) = %v, want ErrInvalidMemoryType", err)
	}
	if store.gotStoreEpisode != nil || store.gotStoreFact != nil || store.gotStoreSkill != nil {
		t.Errorf("Remember(invalid type) made a store call")
	}
}

// TestStoreManager_Remember_Episodic pins the Episodic branch:
// Remember with Type=Episodic MUST extract session_id from metadata
// and call StoreEpisode with a populated Episode. A regression that
// extracted the wrong key (e.g., "sessionID") or dropped fields would
// silently corrupt episodic memory.
func TestStoreManager_Remember_Episodic(t *testing.T) {
	store := &fakeStore{}
	mgr := NewManager(store)

	ts := time.Now()
	err := mgr.Remember(context.Background(), &Memory{
		ID:        "ep-1",
		Type:      Episodic,
		Content:   "user said hello",
		Metadata:  map[string]interface{}{"session_id": "sess-42"},
		Timestamp: ts,
	})
	if err != nil {
		t.Fatalf("Remember(Episodic): %v", err)
	}
	if store.gotStoreEpisode == nil {
		t.Fatal("StoreEpisode not called")
	}
	ep := store.gotStoreEpisode
	if ep.ID != "ep-1" {
		t.Errorf("Episode.ID = %q, want ep-1", ep.ID)
	}
	if ep.SessionID != "sess-42" {
		t.Errorf("Episode.SessionID = %q, want sess-42", ep.SessionID)
	}
	if ep.UserMessage != "user said hello" {
		t.Errorf("Episode.UserMessage = %q, want %q", ep.UserMessage, "user said hello")
	}
	if !ep.Timestamp.Equal(ts) {
		t.Errorf("Episode.Timestamp = %v, want %v", ep.Timestamp, ts)
	}
}

// TestStoreManager_Remember_Semantic pins the Semantic branch:
// Remember with Type=Semantic MUST extract category from metadata and
// call StoreFact with a populated Fact (default confidence, created/updated
// timestamps from the memory's Timestamp).
func TestStoreManager_Remember_Semantic(t *testing.T) {
	store := &fakeStore{}
	mgr := NewManager(store)

	ts := time.Now()
	err := mgr.Remember(context.Background(), &Memory{
		ID:        "f-1",
		Type:      Semantic,
		Content:   "user likes coffee",
		Metadata:  map[string]interface{}{"category": "preference"},
		Timestamp: ts,
	})
	if err != nil {
		t.Fatalf("Remember(Semantic): %v", err)
	}
	if store.gotStoreFact == nil {
		t.Fatal("StoreFact not called")
	}
	f := store.gotStoreFact
	if f.ID != "f-1" {
		t.Errorf("Fact.ID = %q, want f-1", f.ID)
	}
	if f.Category != "preference" {
		t.Errorf("Fact.Category = %q, want preference", f.Category)
	}
	if f.Content != "user likes coffee" {
		t.Errorf("Fact.Content = %q, want %q", f.Content, "user likes coffee")
	}
	if !f.CreatedAt.Equal(ts) {
		t.Errorf("Fact.CreatedAt = %v, want %v", f.CreatedAt, ts)
	}
	if !f.UpdatedAt.Equal(ts) {
		t.Errorf("Fact.UpdatedAt = %v, want %v", f.UpdatedAt, ts)
	}
}

// TestStoreManager_Remember_Procedural pins the Procedural branch:
// Remember with Type=Procedural MUST extract name from metadata and
// call StoreSkill with a populated Skill.
func TestStoreManager_Remember_Procedural(t *testing.T) {
	store := &fakeStore{}
	mgr := NewManager(store)

	ts := time.Now()
	err := mgr.Remember(context.Background(), &Memory{
		ID:        "sk-1",
		Type:      Procedural,
		Content:   "summarize a webpage",
		Metadata:  map[string]interface{}{"name": "summarize"},
		Timestamp: ts,
	})
	if err != nil {
		t.Fatalf("Remember(Procedural): %v", err)
	}
	if store.gotStoreSkill == nil {
		t.Fatal("StoreSkill not called")
	}
	sk := store.gotStoreSkill
	if sk.ID != "sk-1" {
		t.Errorf("Skill.ID = %q, want sk-1", sk.ID)
	}
	if sk.Name != "summarize" {
		t.Errorf("Skill.Name = %q, want summarize", sk.Name)
	}
	if sk.Description != "summarize a webpage" {
		t.Errorf("Skill.Description = %q, want %q", sk.Description, "summarize a webpage")
	}
	if !sk.CreatedAt.Equal(ts) {
		t.Errorf("Skill.CreatedAt = %v, want %v", sk.CreatedAt, ts)
	}
}
