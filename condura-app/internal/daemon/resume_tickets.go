// Package daemon — resume ticket store.
//
// Lifecycle:
//  1. The IPC handler `daemon.resume_request` mints a single-use
//     ticket (32-byte hex) with a 5-minute TTL and records it here.
//  2. The human runs `condura resume --confirm <ticket>` in a terminal
//     (or the GUI calls halt.confirm_resume with the secret).
//  3. The IPC handler `halt.confirm_resume` consumes the ticket here
//     and resumes the daemon.
//
// Rate limits (P0-1 spam hardening):
//   - max 3 pending tickets per halt cycle (older ones expire naturally)
//   - min 10s between requests (otherwise the handler returns a
//     rate-limit error; the user can retry shortly)
package daemon

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"sync"
	"time"
)

const (
	// resumeTicketTTL bounds how long a minted ticket is valid.
	resumeTicketTTL = 5 * time.Minute
	// resumeTicketMaxPending caps the number of pending tickets per
	// daemon lifetime (older ones are naturally evicted by TTL).
	resumeTicketMaxPending = 3
	// resumeRequestMinInterval rate-limits ticket minting so a compromised
	// conductor can't spam the IPC channel.
	resumeRequestMinInterval = 10 * time.Second
)

// ErrResumeRateLimited is returned when a resume_request arrives too
// soon after the previous one.
var ErrResumeRateLimited = errors.New("resume: rate-limited (too many requests)")

// ErrResumeTicketUnknown is returned when the ticket does not exist
// (already consumed, never minted, or evicted by TTL).
var ErrResumeTicketUnknown = errors.New("resume: ticket unknown (consumed, never minted, or expired)")

// ErrResumeTicketExpired is returned for a ticket that exists but
// exceeded the TTL.
var ErrResumeTicketExpired = errors.New("resume: ticket expired")

// ErrResumeSecretMissing is returned when the daemon has no secret
// configured (Load failed or was not called).
var ErrResumeSecretMissing = errors.New("resume: daemon has no resume secret (Load not called)")

// ErrResumeSecretMismatch is returned when the supplied secret does
// not match the daemon's.
var ErrResumeSecretMismatch = errors.New("resume: secret mismatch")

// resumeTicket is one minted ticket.
type resumeTicket struct {
	ticket string
	// sessionID disambiguates tickets across halt cycles: the rate
	// limit and max-pending count are scoped per halt cycle so the
	// user can have several pending tickets while the daemon is
	// halted, and minting a new ticket after auto-resume/re-halt
	// resets the count.
	sessionID string
	mintedAt  time.Time
	// ttl is the per-ticket validity duration (set at Mint). Honors
	// MintWithTTL (test seam). Defaults to resumeTicketTTL if zero.
	ttl time.Duration
}

// ResumeTicketStore is the in-memory map of pending tickets + the
// rate-limit clock. It is safe for concurrent use.
type ResumeTicketStore struct {
	mu               sync.Mutex
	tickets          map[string]resumeTicket // ticket -> meta
	lastRequestAt    time.Time
	sessionCounter   uint64 // bumped on each consume() so future mints are in a new "session"
	currentSessionID string

	nowFunc func() time.Time // injectable for tests
}

// NewResumeTicketStore constructs an empty store.
func NewResumeTicketStore() *ResumeTicketStore {
	return &ResumeTicketStore{
		tickets:          map[string]resumeTicket{},
		currentSessionID: "0",
		nowFunc:          time.Now,
	}
}

// Mint creates a new single-use ticket. Returns ErrResumeRateLimited
// if the previous request was too recent, or if too many tickets are
// already pending. ticketTTL is the validity duration (test seam).
func (s *ResumeTicketStore) Mint() (string, error) {
	return s.MintWithTTL(resumeTicketTTL)
}

// MintWithTTL is Mint with an explicit TTL (for tests).
func (s *ResumeTicketStore) MintWithTTL(ttl time.Duration) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	now := s.nowFunc()
	if !s.lastRequestAt.IsZero() && now.Sub(s.lastRequestAt) < resumeRequestMinInterval {
		return "", ErrResumeRateLimited
	}
	// GC stale entries first.
	for k, t := range s.tickets {
		if now.Sub(t.mintedAt) > ttl {
			delete(s.tickets, k)
		}
	}
	if len(s.tickets) >= resumeTicketMaxPending {
		return "", ErrResumeRateLimited
	}
	raw := make([]byte, 32)
	if _, err := rand.Read(raw); err != nil {
		return "", err
	}
	ticket := hex.EncodeToString(raw)
	s.tickets[ticket] = resumeTicket{
		ticket:    ticket,
		sessionID: s.currentSessionID,
		mintedAt:  now,
		ttl:       ttl,
	}
	s.lastRequestAt = now
	return ticket, nil
}

// Consume validates a ticket and the supplied secret, marks the ticket
// consumed, and (only on success) bumps sessionID so future mints
// reset the per-session rate-limit counters. secret is compared in
// constant time. Returns (consumedSessionID, nil) on success, or
// one of the sentinel errors on failure.
//
// The secret parameter is the hex secret loaded by ResumeSecretManager.
func (s *ResumeTicketStore) Consume(ticket, secret string, expectedSecret string) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if expectedSecret == "" {
		return "", ErrResumeSecretMissing
	}
	if !constantTimeEqual(secret, expectedSecret) {
		return "", ErrResumeSecretMismatch
	}
	t, ok := s.tickets[ticket]
	if !ok {
		return "", ErrResumeTicketUnknown
	}
	now := s.nowFunc()
	effective := t.ttl
	if effective <= 0 {
		effective = resumeTicketTTL
	}
	if now.Sub(t.mintedAt) > effective {
		delete(s.tickets, ticket)
		return "", ErrResumeTicketExpired
	}
	delete(s.tickets, ticket)
	s.sessionCounter++
	s.currentSessionID = hexEncodeUint64(s.sessionCounter)
	return t.sessionID, nil
}

// constantTimeEqual is a constant-time string compare (avoids timing
// oracles on the secret or ticket).
func constantTimeEqual(a, b string) bool {
	if len(a) != len(b) {
		return false
	}
	var diff byte
	for i := 0; i < len(a); i++ {
		diff |= a[i] ^ b[i]
	}
	return diff == 0
}

func hexEncodeUint64(n uint64) string {
	const hexDigits = "0123456789abcdef"
	if n == 0 {
		return "0"
	}
	var buf [16]byte
	i := len(buf)
	for n > 0 {
		i--
		buf[i] = hexDigits[n&0xF]
		n >>= 4
	}
	return string(buf[i:])
}
