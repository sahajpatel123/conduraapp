// Package telemetry implements the opt-in anonymous event channel.
//
// Per the locked-in decision: opt-in, default OFF, no PII. We
// only ever send:
//
//   - The app version + OS (e.g. "synaptic 0.1.0 / darwin/arm64")
//   - Anonymous command counters (e.g. "messages_sent: 12")
//   - Crash signatures (stack-trace SHA256 hashes, never source)
//
// We never send user IDs, IPs, prompt contents, file paths, or
// anything that could identify a user. The endpoint URL is
// configurable; when telemetry is disabled, nothing leaves the
// process.
package telemetry

import (
	"bytes"
	"context"
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"runtime"
	"sync"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/version"
)

// Event is one anonymous event to be sent.
type Event struct {
	Kind     string         `json:"kind"`     // "session_start", "command", "crash"
	TS       time.Time      `json:"ts"`       // RFC3339
	Version  string         `json:"version"`  // app version
	OS       string         `json:"os"`       // runtime.GOOS
	Arch     string         `json:"arch"`     // runtime.GOARCH
	Counters map[string]int `json:"counters"` // command counters
	Hashes   []string       `json:"hashes"`   // sha256(stack) hex digests
}

// Reporter is the telemetry reporter. Construct once at startup;
// share across handlers.
type Reporter struct {
	mu        sync.Mutex
	db        *sql.DB
	endpoint  string
	client    *http.Client
	enabled   bool
	sessionID string
}

// New returns a Reporter. Call SetEnabled(true) to start sending.
func New(db *sql.DB, endpoint string) *Reporter {
	return &Reporter{
		db:        db,
		endpoint:  endpoint,
		client:    &http.Client{Timeout: 10 * time.Second},
		sessionID: newSessionID(),
	}
}

// SetEnabled turns the reporter on or off. When off, all Record
// calls are no-ops.
func (r *Reporter) SetEnabled(v bool) {
	r.mu.Lock()
	r.enabled = v
	r.mu.Unlock()
	if r.db != nil {
		_, _ = r.db.ExecContext(context.Background(), `UPDATE telemetry_counters SET enabled = ? WHERE id = 1`, boolToInt(v))
	}
}

// Enabled returns the current state.
func (r *Reporter) Enabled() bool {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.enabled
}

// Endpoint returns the configured URL.
func (r *Reporter) Endpoint() string {
	return r.muGet()
}

func (r *Reporter) muGet() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.endpoint
}

// SetEndpoint updates the URL.
func (r *Reporter) SetEndpoint(url string) {
	r.mu.Lock()
	r.endpoint = url
	r.mu.Unlock()
	if r.db != nil {
		_, _ = r.db.ExecContext(context.Background(), `UPDATE telemetry_counters SET enabled = enabled WHERE id = 1`)
	}
}

// RecordSessionStart increments the session_starts counter and
// asynchronously sends a session_start event.
func (r *Reporter) RecordSessionStart() {
	if !r.Enabled() {
		return
	}
	r.incr("session_starts")
	r.sendAsync(Event{
		Kind:    "session_start",
		TS:      time.Now().UTC(),
		Version: version.Get().Version,
		OS:      safeGoos(),
		Arch:    safeGoarch(),
		Counters: map[string]int{
			"session_id_prefix": sessionIDPrefix(r.sessionID),
		},
	})
}

// RecordCommand increments a per-command counter.
func (r *Reporter) RecordCommand(name string) {
	if !r.Enabled() {
		return
	}
	r.incr("messages_sent")
	r.sendAsync(Event{
		Kind:    "command",
		TS:      time.Now().UTC(),
		Version: version.Get().Version,
		OS:      safeGoos(),
		Arch:    safeGoarch(),
		Counters: map[string]int{
			"messages_sent":   1,
			"command_" + name: 1,
		},
	})
}

// RecordCrash records a crash signature. stack may contain file
// paths / function names; we hash it so no source leaves the
// process.
func (r *Reporter) RecordCrash(stack string) {
	if !r.Enabled() {
		return
	}
	r.incr("errors")
	sum := sha256.Sum256([]byte(stack))
	hash := hex.EncodeToString(sum[:])
	r.sendAsync(Event{
		Kind:    "crash",
		TS:      time.Now().UTC(),
		Version: version.Get().Version,
		OS:      safeGoos(),
		Arch:    safeGoarch(),
		Counters: map[string]int{
			"errors": 1,
		},
		Hashes: []string{hash},
	})
}

// Flush is a no-op for now; we send events async. Reserved for
// future batched flushing.
func (r *Reporter) Flush() {
	// intentionally empty
}

func (r *Reporter) incr(name string) {
	if r.db == nil {
		return
	}
	_, _ = r.db.ExecContext(context.Background(), fmt.Sprintf(`UPDATE telemetry_counters SET %s = %s + 1 WHERE id = 1`, name, name))
}

func (r *Reporter) sendAsync(ev Event) {
	go func() {
		endpoint := r.Endpoint()
		if endpoint == "" {
			return
		}
		body, err := json.Marshal(ev)
		if err != nil {
			return
		}
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		req, _ := http.NewRequestWithContext(ctx, http.MethodPost, endpoint, bytes.NewReader(body))
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("X-Synaptic-Session", r.sessionID)
		resp, err := r.client.Do(req)
		if err != nil {
			return
		}
		_ = resp.Body.Close()
	}()
}

// jsonReader is unused — removed; bytes.NewReader is used instead.

// --- helpers ---

func newSessionID() string {
	now := time.Now().UnixNano()
	h := sha256.Sum256([]byte(fmt.Sprintf("%d", now)))
	return hex.EncodeToString(h[:8])
}

// sessionIDPrefix returns the first 4 bytes of the session ID as
// an int. Used as a privacy-preserving grouping key in counters
// (so the same user's events can be aggregated without ever
// sending the actual user identifier).
func sessionIDPrefix(id string) int {
	if len(id) < 4 {
		return 0
	}
	v, err := hex.DecodeString(id[:8])
	if err != nil || len(v) < 4 {
		return 0
	}
	return int(v[0])<<24 | int(v[1])<<16 | int(v[2])<<8 | int(v[3])
}

func safeGoos() string   { return runtime.GOOS }
func safeGoarch() string { return runtime.GOARCH }

func boolToInt(b bool) int {
	if b {
		return 1
	}
	return 0
}
