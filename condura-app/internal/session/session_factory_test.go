package session

import (
	"context"
	"sync/atomic"
	"testing"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/blastradius"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/gatekeeper"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/llm"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/sse"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/status"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/stream"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/voice"
)

// fakeFactoryProvider satisfies session.Provider for the
// minimum surface the Factory exercises (NewFactory validation;
// Factory.New field-copy). It does NOT need Stream / Models —
// the Factory never calls them.
type fakeFactoryProvider struct{}

func (p *fakeFactoryProvider) Chat(_ context.Context, _ string, _ llm.ChatRequest) (llm.ChatResponse, error) {
	return llm.ChatResponse{}, nil
}

// fakeSpeaker satisfies voice.Speaker. The setter test does not
// exercise Speak/Stop; it only checks the pointer round-trip.
type fakeSpeaker struct{ id int }

func (s *fakeSpeaker) Speak(_ context.Context, _ string) error { return nil }
func (s *fakeSpeaker) Stop()                                   {}

// fakeGatekeeper satisfies gatekeeper.Gatekeeper. The setter
// test only checks that the injected gatekeeper pointer
// round-trips through SetGatekeeper onto the Factory.
type fakeGatekeeper struct{ allow atomic.Bool }

func (g *fakeGatekeeper) Evaluate(_ context.Context, _ blastradius.Action) (gatekeeper.Decision, string) {
	if g.allow.Load() {
		return gatekeeper.Allow, "test-allow"
	}
	return gatekeeper.Deny, "test-deny"
}

// fakeMemoryStore satisfies session.MemoryStore.
type fakeMemoryStore struct{}

func (m *fakeMemoryStore) Recall(_ context.Context, _ string, _ int) ([]string, error) {
	return nil, nil
}

// fakePredictorStore satisfies session.PredictorStore.
type fakePredictorStore struct{}

func (p *fakePredictorStore) Predict(_ context.Context, _ string) (string, error) {
	return "", nil
}

// newTestFactory builds a Factory with the minimum viable deps
// (streamMgr, Provider, broker) so the setter tests can call
// f.New(...) and inspect the resulting Session's cfg. The
// cleanup closes the broker + manager so parallel runs don't
// leak file descriptors.
func newTestFactory(t *testing.T, providerName, model string) *Factory {
	t.Helper()
	broker := sse.NewBroker()
	t.Cleanup(func() { broker.Close() })
	reg := llm.NewRegistry()
	mgr := stream.NewManager(broker, reg)
	t.Cleanup(func() { mgr.Close() })
	f, err := NewFactory(mgr, &fakeFactoryProvider{}, providerName, model, nil, broker)
	if err != nil {
		t.Fatalf("NewFactory: %v", err)
	}
	return f
}

// TestNewFactory_RejectsNilStreamMgr verifies the first guard.
// The session's hot path (Run) calls streamMgr directly; a nil
// streamMgr would NPE on the first message. Failing closed at
// construction time is the contract.
func TestNewFactory_RejectsNilStreamMgr(t *testing.T) {
	broker := sse.NewBroker()
	defer broker.Close()
	if _, err := NewFactory(nil, &fakeFactoryProvider{}, "test", "gpt-4o", nil, broker); err == nil {
		t.Fatal("NewFactory(nil streamMgr) should return an error")
	}
}

// TestNewFactory_RejectsNilProvider verifies the second guard.
// The session's Provider is called on every chat. A nil provider
// would NPE on the first user message.
func TestNewFactory_RejectsNilProvider(t *testing.T) {
	broker := sse.NewBroker()
	defer broker.Close()
	reg := llm.NewRegistry()
	mgr := stream.NewManager(broker, reg)
	defer mgr.Close()
	if _, err := NewFactory(mgr, nil, "test", "gpt-4o", nil, broker); err == nil {
		t.Fatal("NewFactory(nil Provider) should return an error")
	}
}

// TestNewFactory_RejectsNilBroker verifies the third guard.
// The session fans status transitions out through the broker; a
// nil broker would NPE on the first state transition.
func TestNewFactory_RejectsNilBroker(t *testing.T) {
	broker := sse.NewBroker()
	defer broker.Close()
	reg := llm.NewRegistry()
	mgr := stream.NewManager(broker, reg)
	defer mgr.Close()
	if _, err := NewFactory(mgr, &fakeFactoryProvider{}, "test", "gpt-4o", nil, nil); err == nil {
		t.Fatal("NewFactory(nil broker) should return an error")
	}
}

// TestNewFactory_AcceptsEmptyProviderName verifies the
// documented "first-launch" state: an empty providerName is
// allowed because session.Run will surface a clear error
// ("no provider configured") instead of failing closed here.
// The test pins the contract: empty providerName + empty model
// both succeed, and the resulting Session carries the empty
// values forward.
func TestNewFactory_AcceptsEmptyProviderName(t *testing.T) {
	f := newTestFactory(t, "", "")
	if f == nil {
		t.Fatal("NewFactory with empty providerName/model should succeed (first-launch state)")
	}
	s := f.New(1)
	if s == nil {
		t.Fatal("Factory.New should return a Session even with empty providerName")
	}
	if s.cfg.ProviderName != "" || s.cfg.Model != "" {
		t.Errorf("Session.cfg providerName=%q model=%q, want both empty", s.cfg.ProviderName, s.cfg.Model)
	}
}

// TestFactory_SetSpeaker verifies the setter contract: after
// SetSpeaker(s), every Session built from this Factory has its
// cfg.Speaker pointer == s. The session speaks the reply on
// success; a nil Speaker field would silently swallow the TTS
// path without an error.
func TestFactory_SetSpeaker(t *testing.T) {
	f := newTestFactory(t, "test", "gpt-4o")
	spk := &fakeSpeaker{id: 1}
	f.SetSpeaker(spk)

	s := f.New(1)
	if s == nil {
		t.Fatal("Factory.New returned nil")
	}
	if s.cfg.Speaker == nil {
		t.Fatal("after SetSpeaker, Session.cfg.Speaker is nil; want non-nil")
	}
	if got, ok := s.cfg.Speaker.(*fakeSpeaker); !ok || got != spk {
		t.Errorf("Session.cfg.Speaker = %v, want pointer identity match on the injected fakeSpeaker", s.cfg.Speaker)
	}
}

// TestFactory_SetOnStatus verifies that the status callback is
// threaded through to every Session built from the factory.
// The callback is what fans state transitions out to the SSE
// broker so the GUI's tray can react.
func TestFactory_SetOnStatus(t *testing.T) {
	f := newTestFactory(t, "test", "gpt-4o")
	var hits atomic.Int32
	cb := func(_ status.Status) { hits.Add(1) }
	f.SetOnStatus(cb)

	s := f.New(1)
	if s == nil {
		t.Fatal("Factory.New returned nil")
	}
	if s.OnStatus == nil {
		t.Fatal("after SetOnStatus, Session.OnStatus is nil; want non-nil")
	}
	// Exercise the injected callback directly to confirm pointer
	// identity (not just non-nil).
	s.OnStatus(status.StatusThinking)
	if got := hits.Load(); got != 1 {
		t.Errorf("OnStatus callback hit count = %d, want 1 (pointer identity mismatch?)", got)
	}
}

// TestFactory_SetGatekeeper verifies that the gatekeeper pointer
// round-trips through to the Session's cfg. SetGatekeeper is
// documented to also set the audit log; the audit field's
// pointer identity is covered by internal/audit's own tests
// (we don't construct a real *audit.Log here — the constructor
// requires a sql.DB + HMAC secret, which is out of scope for
// a Factory setter contract).
func TestFactory_SetGatekeeper(t *testing.T) {
	f := newTestFactory(t, "test", "gpt-4o")
	gate := &fakeGatekeeper{}
	gate.allow.Store(true)
	// Pass nil for auditLog — the setter assigns it as-is, and
	// the test does not need a real *audit.Log to verify the
	// gatekeeper pointer round-trip.
	f.SetGatekeeper(gate, nil)

	s := f.New(1)
	if s == nil {
		t.Fatal("Factory.New returned nil")
	}
	if s.cfg.Gatekeeper == nil {
		t.Fatal("after SetGatekeeper, Session.cfg.Gatekeeper is nil; want non-nil")
	}
	if got, ok := s.cfg.Gatekeeper.(*fakeGatekeeper); !ok || got != gate {
		t.Errorf("Session.cfg.Gatekeeper = %v, want injected fakeGatekeeper (pointer identity)", s.cfg.Gatekeeper)
	}
	// Audit was nil-in / nil-out — verify the setter didn't
	// silently substitute a non-nil value.
	if s.cfg.Audit != nil {
		t.Errorf("Session.cfg.Audit = %v, want nil (passed nil to SetGatekeeper)", s.cfg.Audit)
	}
}

// TestFactory_SetMemory verifies the memory-store round-trip.
// MemoryStore.Recall is what surfaces prior user context into
// the prompt; a nil field would silently produce context-free
// replies.
func TestFactory_SetMemory(t *testing.T) {
	f := newTestFactory(t, "test", "gpt-4o")
	mem := &fakeMemoryStore{}
	f.SetMemory(mem)

	s := f.New(1)
	if s == nil {
		t.Fatal("Factory.New returned nil")
	}
	if s.cfg.Memory == nil {
		t.Fatal("after SetMemory, Session.cfg.Memory is nil; want non-nil")
	}
	if got, ok := s.cfg.Memory.(*fakeMemoryStore); !ok || got != mem {
		t.Errorf("Session.cfg.Memory = %v, want injected fakeMemoryStore (pointer identity)", s.cfg.Memory)
	}
}

// TestFactory_SetPredictor verifies the adaptive-engine
// predictor round-trip. PredictorStore.Predict is consulted on
// each query to surface user-adaptive suggestions.
func TestFactory_SetPredictor(t *testing.T) {
	f := newTestFactory(t, "test", "gpt-4o")
	pred := &fakePredictorStore{}
	f.SetPredictor(pred)

	s := f.New(1)
	if s == nil {
		t.Fatal("Factory.New returned nil")
	}
	if s.cfg.Predictor == nil {
		t.Fatal("after SetPredictor, Session.cfg.Predictor is nil; want non-nil")
	}
	if got, ok := s.cfg.Predictor.(*fakePredictorStore); !ok || got != pred {
		t.Errorf("Session.cfg.Predictor = %v, want injected fakePredictorStore (pointer identity)", s.cfg.Predictor)
	}
}

// TestFactory_UpdatePrimary verifies the live-reload contract.
// UpdatePrimary is called after the user enables a new provider
// or adds an API key; sessions created AFTER the call must use
// the new (providerName, model).
func TestFactory_UpdatePrimary(t *testing.T) {
	f := newTestFactory(t, "openai", "gpt-4o")

	// Pre-update: initial values pinned.
	if got := f.New(1).cfg.ProviderName; got != "openai" {
		t.Errorf("initial ProviderName = %q, want openai", got)
	}
	if got := f.New(2).cfg.Model; got != "gpt-4o" {
		t.Errorf("initial Model = %q, want gpt-4o", got)
	}

	// Update.
	f.UpdatePrimary("anthropic", "claude-sonnet-4.5")

	// Post-update: new values, on a fresh Session.
	if got := f.New(3).cfg.ProviderName; got != "anthropic" {
		t.Errorf("post-update ProviderName = %q, want anthropic", got)
	}
	if got := f.New(4).cfg.Model; got != "claude-sonnet-4.5" {
		t.Errorf("post-update Model = %q, want claude-sonnet-4.5", got)
	}
}

// TestFactory_UpdatePrimary_PreservesStreamMgrAndBroker verifies
// the second half of the live-reload contract: UpdatePrimary
// MUST NOT touch the stream manager, broker, or any other
// pinned dependency — only (providerName, model). Resetting the
// stream manager mid-session would orphan every in-flight SSE
// subscription.
func TestFactory_UpdatePrimary_PreservesStreamMgrAndBroker(t *testing.T) {
	f := newTestFactory(t, "openai", "gpt-4o")
	pre := f.New(1)

	// Snapshot the deps that must survive UpdatePrimary.
	preStreamMgr := pre.cfg.StreamMgr
	preBroker := pre.cfg.Broker
	preProvider := pre.cfg.Provider

	f.UpdatePrimary("anthropic", "claude-sonnet-4.5")
	post := f.New(2)

	if post.cfg.StreamMgr != preStreamMgr {
		t.Error("UpdatePrimary mutated StreamMgr; only (providerName, model) should change")
	}
	if post.cfg.Broker != preBroker {
		t.Error("UpdatePrimary mutated Broker; only (providerName, model) should change")
	}
	if post.cfg.Provider != preProvider {
		t.Error("UpdatePrimary mutated Provider; only (providerName, model) should change")
	}
}

// Compile-time check that fakeSpeaker satisfies voice.Speaker.
// The setter test relies on this; if the interface changes, the
// build breaks here before any test runs.
var _ voice.Speaker = (*fakeSpeaker)(nil)
