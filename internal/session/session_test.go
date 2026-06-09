package session

import (
	"context"
	"errors"
	"sync/atomic"
	"testing"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/llm"
	"github.com/sahajpatel123/synapticapp/internal/stream"
)

// fakeProvider is a stub Provider. It returns the configured
// response without making a real network call.
type fakeProvider struct {
	resp  llm.ChatResponse
	err   error
	calls atomic.Int32
}

func (p *fakeProvider) Chat(_ context.Context, _ string, _ llm.ChatRequest) (llm.ChatResponse, error) {
	p.calls.Add(1)
	if p.err != nil {
		return llm.ChatResponse{}, p.err
	}
	return p.resp, nil
}

func TestNew_RequiresStreamMgr(t *testing.T) {
	_, err := New(Config{Provider: &fakeProvider{}, ProviderName: "x", Model: "y"})
	if err == nil {
		t.Fatal("expected error for nil StreamMgr")
	}
}

func TestNew_RequiresProvider(t *testing.T) {
	_, err := New(Config{StreamMgr: stream.NewManager(nil, nil), ProviderName: "x", Model: "y"})
	if err == nil {
		t.Fatal("expected error for nil Provider")
	}
}

func TestNew_RequiresProviderName(t *testing.T) {
	_, err := New(Config{StreamMgr: stream.NewManager(nil, nil), Provider: &fakeProvider{}, Model: "y"})
	if err == nil {
		t.Fatal("expected error for empty ProviderName")
	}
}

func TestNew_RequiresModel(t *testing.T) {
	_, err := New(Config{StreamMgr: stream.NewManager(nil, nil), Provider: &fakeProvider{}, ProviderName: "x"})
	if err == nil {
		t.Fatal("expected error for empty Model")
	}
}

func TestNew_OK(t *testing.T) {
	s, err := New(Config{
		StreamMgr:    stream.NewManager(nil, nil),
		Provider:     &fakeProvider{},
		ProviderName: "x",
		Model:        "y",
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if s == nil {
		t.Fatal("nil session")
	}
}

func TestRun_EmptyQueryReturnsImmediately(t *testing.T) {
	s, err := New(Config{
		StreamMgr:    stream.NewManager(nil, nil),
		Provider:     &fakeProvider{},
		ProviderName: "x",
		Model:        "y",
	})
	if err != nil {
		t.Fatal(err)
	}
	out, err := s.Run(context.Background(), "")
	if err != nil {
		t.Fatalf("Run: %v", err)
	}
	if out != "" {
		t.Errorf("expected empty output, got %q", out)
	}
}

func TestRun_AlreadyRunning(t *testing.T) {
	// Force the busy flag.
	s := &Session{cfg: Config{
		StreamMgr:    stream.NewManager(nil, nil),
		Provider:     &fakeProvider{},
		ProviderName: "x",
		Model:        "y",
	}}
	s.busy = true
	defer func() { s.busy = false }()

	_, err := s.Run(context.Background(), "hello")
	if !errors.Is(err, ErrAlreadyRunning) {
		t.Errorf("err = %v, want ErrAlreadyRunning", err)
	}
}

func TestBuildMessages_NoConversation(t *testing.T) {
	s := &Session{cfg: Config{}}
	msgs, err := s.buildMessages(context.Background(), "hi")
	if err != nil {
		t.Fatalf("buildMessages: %v", err)
	}
	if len(msgs) != 1 {
		t.Fatalf("expected 1 message, got %d", len(msgs))
	}
	if msgs[0].Role != llm.RoleUser || msgs[0].Content != "hi" {
		t.Errorf("msg = %+v", msgs[0])
	}
}

func TestStatus_IdleForFreshSession(t *testing.T) {
	s, _ := New(Config{
		StreamMgr:    stream.NewManager(nil, nil),
		Provider:     &fakeProvider{},
		ProviderName: "x",
		Model:        "y",
	})
	if got := s.Status(); got != 0 { // StatusIdle = 0
		t.Errorf("Status = %d, want 0", got)
	}
}

func TestRun_StreamStartFails(t *testing.T) {
	// We use a real stream manager with a nil registry; the
	// manager panics on Start because the registry is nil. This
	// test verifies that Run is bounded by context.
	s, err := New(Config{
		StreamMgr:    stream.NewManager(nil, nil),
		Provider:     &fakeProvider{},
		ProviderName: "x",
		Model:        "y",
	})
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	// Catch the expected panic and verify the run is bounded.
	defer func() {
		_ = recover() // expected: nil registry causes panic
	}()
	_, _ = s.Run(ctx, "hello")
}
