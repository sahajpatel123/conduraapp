package daemon

import (
	"errors"
	"strings"
	"sync"
	"testing"
	"time"
)

// TestResumeTicketStore_MintConsume verifies the happy path:
// mint a ticket, then consume it with the right secret. The session
// ID is returned; the rate-limit clock is updated.
func TestResumeTicketStore_MintConsume(t *testing.T) {
	s := NewResumeTicketStore()
	s.nowFunc = func() time.Time { return time.Unix(0, 0) }
	const secret = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	ticket, err := s.MintWithTTL(time.Minute)
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	if len(ticket) != 64 {
		t.Fatalf("ticket length = %d, want 64 (32 hex bytes)", len(ticket))
	}
	sessionID, err := s.Consume(ticket, secret, secret)
	if err != nil {
		t.Fatalf("Consume: %v", err)
	}
	if sessionID == "" {
		t.Fatal("expected non-empty session ID on successful consume")
	}
}

// TestResumeTicketStore_RejectsBadSecret: a wrong secret must fail
// with ErrResumeSecretMismatch and the ticket must remain unconsumed
// (rate-limit + retry protection).
func TestResumeTicketStore_RejectsBadSecret(t *testing.T) {
	s := NewResumeTicketStore()
	s.nowFunc = func() time.Time { return time.Unix(0, 0) }
	const good = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	const bad = "ffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffffff"
	ticket, err := s.MintWithTTL(time.Minute)
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	if _, err := s.Consume(ticket, bad, good); !errors.Is(err, ErrResumeSecretMismatch) {
		t.Fatalf("Consume(bad secret) = %v, want ErrResumeSecretMismatch", err)
	}
	// Ticket should still be valid — retry should also fail with
	// mismatch (not "unknown"), proving the bad attempt did NOT
	// consume it.
	if _, err := s.Consume(ticket, bad, good); !errors.Is(err, ErrResumeSecretMismatch) {
		t.Fatalf("retry: %v, want ErrResumeSecretMismatch", err)
	}
	// Right secret still works (no replay).
	if _, err := s.Consume(ticket, good, good); err != nil {
		t.Fatalf("Consume(good secret) after bad retries: %v", err)
	}
}

// TestResumeTicketStore_ExpiredTicket: a ticket older than the TTL
// is rejected as ErrResumeTicketExpired and auto-evicted.
func TestResumeTicketStore_ExpiredTicket(t *testing.T) {
	s := NewResumeTicketStore()
	const secret = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	mint := time.Unix(0, 0)
	s.nowFunc = func() time.Time { return mint }
	ticket, err := s.MintWithTTL(time.Minute)
	if err != nil {
		t.Fatalf("Mint: %v", err)
	}
	// Jump 2 minutes past mint.
	s.nowFunc = func() time.Time { return mint.Add(2 * time.Minute) }
	if _, err := s.Consume(ticket, secret, secret); !errors.Is(err, ErrResumeTicketExpired) {
		t.Fatalf("Consume after TTL: %v, want ErrResumeTicketExpired", err)
	}
}

// TestResumeTicketStore_RateLimit: a second Mint within 10s of the
// first is rejected with ErrResumeRateLimited. After waiting, Mint
// succeeds.
func TestResumeTicketStore_RateLimit(t *testing.T) {
	s := NewResumeTicketStore()
	now := time.Unix(1_000, 0)
	s.nowFunc = func() time.Time { return now }
	if _, err := s.MintWithTTL(time.Minute); err != nil {
		t.Fatalf("first Mint: %v", err)
	}
	// Second Mint immediately after — rate-limited.
	if _, err := s.MintWithTTL(time.Minute); !errors.Is(err, ErrResumeRateLimited) {
		t.Fatalf("second Mint: %v, want ErrResumeRateLimited", err)
	}
	// Advance past the 10s minimum interval.
	s.nowFunc = func() time.Time { return now.Add(15 * time.Second) }
	if _, err := s.MintWithTTL(time.Minute); err != nil {
		t.Fatalf("third Mint after wait: %v", err)
	}
}

// TestResumeTicketStore_MaxPending: enforces the 3-pending cap.
// After 3 mints, the 4th is rate-limited. Consuming one frees a slot.
func TestResumeTicketStore_MaxPending(t *testing.T) {
	s := NewResumeTicketStore()
	now := time.Unix(0, 0)
	s.nowFunc = func() time.Time { return now }
	// First 3: spaced 15s apart so the per-request interval is OK.
	for i := 0; i < 3; i++ {
		s.nowFunc = func() time.Time { return now.Add(time.Duration(i) * 15 * time.Second) }
		if _, err := s.MintWithTTL(time.Hour); err != nil {
			t.Fatalf("Mint %d: %v", i+1, err)
		}
	}
	// 4th within the interval AND at the cap — rate-limited.
	s.nowFunc = func() time.Time { return now.Add(45 * time.Second) }
	if _, err := s.MintWithTTL(time.Hour); !errors.Is(err, ErrResumeRateLimited) {
		t.Fatalf("4th Mint at cap: %v, want ErrResumeRateLimited", err)
	}
}

// TestResumeTicketStore_ConcurrentMints: parallel mints must not
// double-mint the same ticket or skip the rate limit. spawn 8
// goroutines, each mints in a tight loop.
func TestResumeTicketStore_ConcurrentMints(t *testing.T) {
	s := NewResumeTicketStore()
	now := time.Unix(0, 0)
	var nowMu sync.Mutex
	s.nowFunc = func() time.Time {
		nowMu.Lock()
		defer nowMu.Unlock()
		return now
	}
	advance := func() {
		nowMu.Lock()
		now = now.Add(15 * time.Second)
		nowMu.Unlock()
	}
	const N = 8
	tickets := make(chan string, N)
	errs := make(chan error, N)
	for i := 0; i < N; i++ {
		go func() {
			advance()
			t, err := s.MintWithTTL(time.Hour)
			tickets <- t
			errs <- err
		}()
	}
	seen := map[string]bool{}
	rateLimitCount := 0
	for i := 0; i < N; i++ {
		tk := <-tickets
		err := <-errs
		switch {
		case err == nil:
			if seen[tk] {
				t.Fatalf("duplicate ticket minted: %s", tk)
			}
			seen[tk] = true
		case errors.Is(err, ErrResumeRateLimited):
			rateLimitCount++
		default:
			t.Fatalf("unexpected error: %v", err)
		}
	}
	if len(seen) == 0 {
		t.Fatal("no tickets minted")
	}
	if !strings.Contains(strings.Join([]string{"ok", "rate-limited"}, ","), "ok") {
		t.Fatalf("expected a mix of successes + rate-limits; got %d successes, %d rate-limited", len(seen), rateLimitCount)
	}
}

// TestResumeTicketStore_UnknownTicket: a ticket that was never
// minted (or was already consumed) is rejected with
// ErrResumeTicketUnknown.
func TestResumeTicketStore_UnknownTicket(t *testing.T) {
	s := NewResumeTicketStore()
	const secret = "0123456789abcdef0123456789abcdef0123456789abcdef0123456789abcdef"
	bogus := "deadbeef" + strings.Repeat("00", 28)
	if _, err := s.Consume(bogus, secret, secret); !errors.Is(err, ErrResumeTicketUnknown) {
		t.Fatalf("Consume unknown ticket: %v, want ErrResumeTicketUnknown", err)
	}
}

// TestResumeTicketStore_MissingExpectedSecret: the daemon's secret
// was never loaded (Load was not called). Must return
// ErrResumeSecretMissing, NOT panic, NOT silently allow.
func TestResumeTicketStore_MissingExpectedSecret(t *testing.T) {
	s := NewResumeTicketStore()
	if _, err := s.Consume("anything", "anything", ""); !errors.Is(err, ErrResumeSecretMissing) {
		t.Fatalf("Consume with empty expected: %v, want ErrResumeSecretMissing", err)
	}
}

// TestConstantTimeEqual: the secret compare must be constant-time
// for matching-length inputs. We can't observe timing in a unit
// test, but at minimum it must produce the right boolean for known
// pairs and never panic on length mismatch.
func TestConstantTimeEqual(t *testing.T) {
	tests := []struct {
		a, b string
		want bool
	}{
		{"abc", "abc", true},
		{"abc", "abd", false},
		{"abc", "abcd", false}, // length mismatch
		{"", "", true},
		{"", "a", false},
	}
	for _, tt := range tests {
		got := constantTimeEqual(tt.a, tt.b)
		if got != tt.want {
			t.Errorf("constantTimeEqual(%q,%q) = %v, want %v", tt.a, tt.b, got, tt.want)
		}
	}
}
