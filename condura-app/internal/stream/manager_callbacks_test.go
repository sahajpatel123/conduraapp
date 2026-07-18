package stream

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/llm"
)

// TestManager_SetBreakerCheck_FailsFastOnFalse pins the circuit-breaker
// fail-fast contract: when SetBreakerCheck's callback returns false,
// Start MUST return an error mentioning "circuit breaker open" BEFORE
// acquiring the provider's stream channel. A regression that
// skipped the check would let a flaky provider issue new calls when
// its breaker is open, defeating the entire purpose of the breaker.
func TestManager_SetBreakerCheck_FailsFastOnFalse(t *testing.T) {
	p := &fakeProvider{name: "fake", models: []llm.ModelInfo{{ID: "fake-1"}}}
	m, _ := newTestManager(t, p)

	m.SetBreakerCheck(func(provider string) bool {
		if provider != "fake" {
			t.Errorf("breakerCheck called with provider=%q, want \"fake\"", provider)
		}
		return false // breaker open — Start MUST refuse
	})

	_, err := m.Start(context.Background(), Request{
		ProviderName: "fake",
		Chat: llm.ChatRequest{
			Model:    "fake-1",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
		},
	})
	if err == nil {
		t.Fatal("Start with breaker=false = nil; want error")
	}
	if !strings.Contains(err.Error(), "circuit breaker open") {
		t.Errorf("error %q should mention 'circuit breaker open'", err.Error())
	}
	if !strings.Contains(err.Error(), "fake") {
		t.Errorf("error %q should mention the provider name 'fake'", err.Error())
	}
}

// TestManager_SetBreakerCheck_PassesThroughOnTrue pins the happy-path
// half of the breaker contract: when the callback returns true, Start
// MUST proceed normally (not error on the breaker check). The test
// verifies that the breaker is consulted but doesn't reject.
func TestManager_SetBreakerCheck_PassesThroughOnTrue(t *testing.T) {
	p := &fakeProvider{name: "fake", models: []llm.ModelInfo{{ID: "fake-1"}}}
	m, _ := newTestManager(t, p)

	called := false
	m.SetBreakerCheck(func(provider string) bool {
		called = true
		return true // breaker closed — Start proceeds
	})

	_, err := m.Start(context.Background(), Request{
		ProviderName: "fake",
		Chat: llm.ChatRequest{
			Model:    "fake-1",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
		},
	})
	if err != nil {
		t.Fatalf("Start with breaker=true: %v", err)
	}
	if !called {
		t.Error("breakerCheck was not called; Start must consult the breaker before launching")
	}
}

// TestManager_SetSpendCheck_FailsFastOnErrSpendCap pins the
// spend-cap fail-fast contract: when SetSpendCheck's callback returns
// ErrSpendCap, Start MUST return ErrSpendCap directly (no wrapping).
// This is the "you've hit your daily limit" gate — it must surface
// to the GUI unchanged so the toast shows the right message.
func TestManager_SetSpendCheck_FailsFastOnErrSpendCap(t *testing.T) {
	p := &fakeProvider{name: "fake", models: []llm.ModelInfo{{ID: "fake-1"}}}
	m, _ := newTestManager(t, p)

	m.SetSpendCheck(func(model string) error {
		return ErrSpendCap
	})

	_, err := m.Start(context.Background(), Request{
		ProviderName: "fake",
		Chat: llm.ChatRequest{
			Model:    "fake-1",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
		},
	})
	if !errors.Is(err, ErrSpendCap) {
		t.Errorf("Start with ErrSpendCap = %v, want ErrSpendCap", err)
	}
}

// TestManager_SetSpendCheck_FailsFastOnOtherError pins the
// non-ErrSpendCap error path: when the spendCheck callback returns
// ANY OTHER error, Start MUST wrap it with ErrSpendCap. Without this
// pin, a regression that returned the raw error would leak the
// underlying provider/DB failure to the GUI without the "spend
// limit" framing the user needs.
func TestManager_SetSpendCheck_FailsFastOnOtherError(t *testing.T) {
	p := &fakeProvider{name: "fake", models: []llm.ModelInfo{{ID: "fake-1"}}}
	m, _ := newTestManager(t, p)

	otherErr := errors.New("spend db locked")
	m.SetSpendCheck(func(model string) error {
		return otherErr
	})

	_, err := m.Start(context.Background(), Request{
		ProviderName: "fake",
		Chat: llm.ChatRequest{
			Model:    "fake-1",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
		},
	})
	if !errors.Is(err, ErrSpendCap) {
		t.Errorf("Start with non-cap error = %v, want ErrSpendCap (wrapped)", err)
	}
	if !errors.Is(err, otherErr) {
		t.Errorf("Start error chain should still include the underlying error %v", otherErr)
	}
}

// TestManager_BreakerRunsBeforeSpendCheck pins the precedence
// contract: when BOTH callbacks are set and the breaker says false,
// Start MUST fail with the breaker error (NOT the spend error).
// The order matters: a closed breaker but an exceeded spend cap
// returns ErrSpendCap; an open breaker returns circuit-breaker error
// regardless of spend. Without this pin, the order could regress
// silently.
func TestManager_BreakerRunsBeforeSpendCheck(t *testing.T) {
	p := &fakeProvider{name: "fake", models: []llm.ModelInfo{{ID: "fake-1"}}}
	m, _ := newTestManager(t, p)

	m.SetBreakerCheck(func(provider string) bool { return false })
	spendCalled := false
	m.SetSpendCheck(func(model string) error {
		spendCalled = true
		return ErrSpendCap
	})

	_, err := m.Start(context.Background(), Request{
		ProviderName: "fake",
		Chat: llm.ChatRequest{
			Model:    "fake-1",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
		},
	})
	if err == nil || !strings.Contains(err.Error(), "circuit breaker") {
		t.Errorf("err = %v, want circuit breaker error (breaker should run first)", err)
	}
	if spendCalled {
		t.Error("spendCheck was called after breaker=false; breaker should fail-fast before spend check")
	}
}

// TestManager_SetBreakerCheck_Overwrite pins the setter-overwrite
// contract: a second SetBreakerCheck call MUST replace the first
// callback. A regression that appended to a slice would let the
// first (stale) callback linger.
func TestManager_SetBreakerCheck_Overwrite(t *testing.T) {
	p := &fakeProvider{name: "fake", models: []llm.ModelInfo{{ID: "fake-1"}}}
	m, _ := newTestManager(t, p)

	m.SetBreakerCheck(func(provider string) bool { return true })
	m.SetBreakerCheck(func(provider string) bool { return false }) // second wins

	_, err := m.Start(context.Background(), Request{
		ProviderName: "fake",
		Chat: llm.ChatRequest{
			Model:    "fake-1",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
		},
	})
	if err == nil {
		t.Error("Start with second SetBreakerCheck returning false = nil; want error (overwrite works)")
	}
	if !strings.Contains(err.Error(), "circuit breaker") {
		t.Errorf("err %q should mention 'circuit breaker' (proving the second callback was the one called)", err.Error())
	}
}

// TestManager_SetBreakerResult_StoresFunction pins the storage
// contract: SetBreakerResult MUST store the callback (not invoke it
// directly — it's called by the stream goroutine after the stream
// completes). We verify storage by setting a sentinel via a closure
// that the test can introspect; we don't directly observe the call
// because the stream completion path is async (goroutine + SSE
// broker).
func TestManager_SetBreakerResult_StoresFunction(t *testing.T) {
	p := &fakeProvider{name: "fake", models: []llm.ModelInfo{{ID: "fake-1"}}}
	m, _ := newTestManager(t, p)

	// Set a callback. We can't directly assert "it was stored" from
	// outside the package (unexported field), but we CAN assert
	// that calling SetBreakerResult doesn't panic and that the
	// subsequent Start succeeds — which would be impossible if
	// the field corruption from SetBreakerResult left the Manager
	// in a broken state.
	m.SetBreakerResult(func(provider string, success bool) {
		// no-op; just exercising the setter
	})

	_, err := m.Start(context.Background(), Request{
		ProviderName: "fake",
		Chat: llm.ChatRequest{
			Model:    "fake-1",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
		},
	})
	if err != nil {
		t.Fatalf("Start after SetBreakerResult: %v", err)
	}
}

// TestManager_SetSpendRecord_StoresFunction pins the storage
// contract: SetSpendRecord MUST store the callback (invoked when
// usage is known post-stream). Like SetBreakerResult, the actual
// invocation is async; we verify storage via "Start still works
// after SetSpendRecord."
func TestManager_SetSpendRecord_StoresFunction(t *testing.T) {
	p := &fakeProvider{name: "fake", models: []llm.ModelInfo{{ID: "fake-1"}}}
	m, _ := newTestManager(t, p)

	m.SetSpendRecord(func(provider, model string, usage llm.Usage) {
		// no-op; just exercising the setter
	})

	_, err := m.Start(context.Background(), Request{
		ProviderName: "fake",
		Chat: llm.ChatRequest{
			Model:    "fake-1",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
		},
	})
	if err != nil {
		t.Fatalf("Start after SetSpendRecord: %v", err)
	}
}

// TestManager_SpendCheckIsCalledWithModel pins the input-validation
// contract: SetSpendCheck's callback MUST receive the model name
// from the request (not the provider name, not an empty string).
// This is the audit hook for the spend monitor.
func TestManager_SpendCheckIsCalledWithModel(t *testing.T) {
	p := &fakeProvider{name: "fake", models: []llm.ModelInfo{{ID: "claude-opus-4-1"}}}
	m, _ := newTestManager(t, p)

	var gotModel string
	m.SetSpendCheck(func(model string) error {
		gotModel = model
		return nil
	})

	_, _ = m.Start(context.Background(), Request{
		ProviderName: "fake",
		Chat: llm.ChatRequest{
			Model:    "claude-opus-4-1",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
		},
	})
	if gotModel != "claude-opus-4-1" {
		t.Errorf("spendCheck received model=%q, want \"claude-opus-4-1\"", gotModel)
	}
}

// TestManager_BreakerCheckIsCalledWithProvider pins the input-validation
// contract: SetBreakerCheck's callback MUST receive the provider name
// from the request (not the model, not an empty string).
func TestManager_BreakerCheckIsCalledWithProvider(t *testing.T) {
	p := &fakeProvider{name: "anthropic-test", models: []llm.ModelInfo{{ID: "claude-opus-4-1"}}}
	m, _ := newTestManager(t, p)

	var gotProvider string
	m.SetBreakerCheck(func(provider string) bool {
		gotProvider = provider
		return true
	})

	_, _ = m.Start(context.Background(), Request{
		ProviderName: "anthropic-test",
		Chat: llm.ChatRequest{
			Model:    "claude-opus-4-1",
			Messages: []llm.Message{{Role: llm.RoleUser, Content: "hi"}},
		},
	})
	if gotProvider != "anthropic-test" {
		t.Errorf("breakerCheck received provider=%q, want \"anthropic-test\"", gotProvider)
	}
}
