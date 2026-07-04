package ipc

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSSETicket_ExchangeAndConsume verifies the full ticket flow:
// 1. POST /sse-ticket with Authorization header returns a ticket.
// 2. GET /events?ticket=<ticket> passes auth (200, not 401).
// 3. A second use of the same ticket is rejected (single-use).
// 4. POST /sse-ticket without auth header is rejected.
func TestSSETicket_ExchangeAndConsume(t *testing.T) {
	s := NewServer()
	st := &ServerTransport{S: s, Token: "secret"}
	require.NoError(t, st.Listen(context.Background(), "tcp://127.0.0.1:0"))
	defer func() { _ = st.Close() }()
	addr := "http://" + st.Addr()

	// 1. No auth header -> 401.
	resp, err := http.Post(addr+"/sse-ticket", "application/json", strings.NewReader(""))
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode)

	// 2. With auth header -> get ticket.
	req, _ := http.NewRequest(http.MethodPost, addr+"/sse-ticket", strings.NewReader(""))
	req.Header.Set("Authorization", "Bearer secret")
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	require.Equal(t, http.StatusOK, resp.StatusCode)

	var ticketResp struct {
		Ticket    string `json:"ticket"`
		ExpiresIn int    `json:"expires_in"`
	}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&ticketResp))
	require.NotEmpty(t, ticketResp.Ticket)
	assert.Equal(t, 30, ticketResp.ExpiresIn)

	// 3. Use ticket on /events — should pass auth.
	// (SSE broker is nil, so we'll get 501 Not Implemented, but
	// that proves auth passed — 401 means auth failed.)
	req, _ = http.NewRequest(http.MethodGet, addr+"/events?ticket="+ticketResp.Ticket, nil)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.NotEqual(t, http.StatusUnauthorized, resp.StatusCode,
		"valid ticket should pass auth (got 501 if SSE broker is nil, not 401)")

	// 4. Reuse same ticket -> rejected (single-use).
	req, _ = http.NewRequest(http.MethodGet, addr+"/events?ticket="+ticketResp.Ticket, nil)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
		"reused ticket should be rejected")

	// 5. No ticket at all -> 401 (token query param no longer accepted).
	req, _ = http.NewRequest(http.MethodGet, addr+"/events?token=secret", nil)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode,
		"token= query param should no longer be accepted for SSE")
}

// TestSSETicket_ExpiredRejected verifies that an expired ticket is
// rejected. We issue a ticket, wait for it to expire, then try to
// use it.
func TestSSETicket_ExpiredRejected(t *testing.T) {
	s := NewServer()
	st := &ServerTransport{S: s, Token: "secret"}
	require.NoError(t, st.Listen(context.Background(), "tcp://127.0.0.1:0"))
	defer func() { _ = st.Close() }()
	addr := "http://" + st.Addr()

	// Issue a ticket with a custom short expiry by manipulating the
	// store directly.
	ticket, err := st.issueSSETicket()
	require.NoError(t, err)

	// Force it to expire.
	st.sseTicketsMu.Lock()
	st.sseTickets[ticket] = time.Now().Add(-1 * time.Second)
	st.sseTicketsMu.Unlock()

	// Use expired ticket -> 401.
	req, _ := http.NewRequest(http.MethodGet, addr+"/events?ticket="+ticket, nil)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	body, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusUnauthorized, resp.StatusCode, "expired ticket should be rejected: %s", body)
}
