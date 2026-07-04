// Package anomaly detects unusual agent behavior per MISSION S10.4.
// Package description is in this file.
//
//nolint:revive,mnd // threshold constants; exported types are self-documenting
package anomaly

import (
	"strings"
	"sync"
	"time"
)

// Detector tracks behavioral anomalies across an agent run.
type Detector struct {
	actions chan actionRecord
	stop    chan struct{}
	onTrip  func(Trip)
	mu      sync.Mutex
	state   detectorState
}

// Trip describes what threshold was hit.
type Trip struct {
	Type   TripType
	Reason string
}

// TripType is the kind of anomaly detected.
type TripType string

const (
	TripRate        TripType = "rate"
	TripDuration    TripType = "duration"
	TripLoop        TripType = "loop"
	TripFailures    TripType = "failures"
	TripNewEndpoint TripType = "new_endpoint"
)

type actionRecord struct {
	kind    string
	coordX  float64
	coordY  float64
	success bool
	time    time.Time
}

type detectorState struct {
	count        int
	failures     int
	startTime    time.Time
	lastActivity time.Time // time of most recent Record(); used by IdleReset
	coordWindow  [][2]float64
	// seenHosts tracks every network host the agent has contacted
	// this session. The 5th "agent went insane" trigger (§5.6)
	// fires when the agent sends to a network endpoint it has
	// never used before. A new session starts with an empty set;
	// the first contact of any host is NOT a trip (the set is
	// empty), but a subsequent contact of a *different* host
	// after the set is populated is. This catches an agent that
	// suddenly pivots to a new exfil endpoint mid-session.
	seenHosts map[string]bool
}

// NewDetector creates an anomaly detector with the given thresholds.
func NewDetector(onTrip func(Trip)) *Detector {
	d := &Detector{
		actions: make(chan actionRecord, 256),
		stop:    make(chan struct{}),
		onTrip:  onTrip,
	}
	d.state.coordWindow = make([][2]float64, 0, 5)
	d.state.startTime = time.Now()
	d.state.seenHosts = make(map[string]bool)

	go d.loop()
	return d
}

// Record stores an action for anomaly analysis. Non-blocking.
func (d *Detector) Record(kind string, x, y float64, success bool) {
	select {
	case d.actions <- actionRecord{kind: kind, coordX: x, coordY: y, success: success, time: time.Now()}:
	default:
	}
}

// RecordNetwork records a network endpoint the agent is about to
// contact. It trips TripNewEndpoint when the host has not been seen
// this session AND the session has already contacted at least one
// other host (so the first ever contact in a session is not a trip,
// but a pivot to a new endpoint mid-session is). The host is the
// URL's hostname (port-stripped); non-network kinds are ignored.
//
// This implements the 5th "agent went insane" trigger from
// CLAUDE.md §5.6: "Agent sends to network endpoints it has never
// used before." The detector is the right place for this because it
// already owns the behavioral-anomaly state and the onTrip callback.
func (d *Detector) RecordNetwork(host string) {
	host = normalizeHost(host)
	if host == "" {
		return
	}
	d.mu.Lock()
	defer d.mu.Unlock()
	d.state.lastActivity = time.Now()
	if d.state.seenHosts == nil {
		d.state.seenHosts = make(map[string]bool)
	}
	if !d.state.seenHosts[host] {
		if len(d.state.seenHosts) > 0 {
			// The session has already contacted at least one
			// other host, so a new host is a pivot — trip.
			d.trip(Trip{Type: TripNewEndpoint, Reason: "network endpoint not seen before this session: " + host})
		}
		d.state.seenHosts[host] = true
	}
}

// normalizeHost strips the port and lowercases the host. Returns
// empty for non-network inputs.
func normalizeHost(host string) string {
	host = strings.ToLower(strings.TrimSpace(host))
	// Strip port.
	if i := strings.LastIndex(host, ":"); i > 0 {
		host = host[:i]
	}
	return host
}

// Reset clears all counters for a new conversation.
func (d *Detector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.state = detectorState{}
	d.state.startTime = time.Now()
	d.state.coordWindow = make([][2]float64, 0, 5)
	d.state.seenHosts = make(map[string]bool)
}

// IdleReset returns true if the detector has been inactive longer
// than the idle threshold. Used by the daemon's idle-watcher to
// automatically call Reset() so cross-session noise doesn't
// accumulate (Phase 16, Rec 6).
//
// "Inactive" means: no Record() call in the last `idle` duration.
// The check is approximate (sampled on each Record call), so a
// session that's truly idle is detected at the next Record()
// OR at the next watcher tick, whichever fires first.
//
// `lastActivity` is tracked separately from `startTime` because
// startTime is reset every time Reset() is called.
func (d *Detector) IdleReset(idle time.Duration) bool {
	if idle <= 0 {
		return false
	}
	d.mu.Lock()
	last := d.state.lastActivity
	count := d.state.count
	failures := d.state.failures
	d.mu.Unlock()
	if count == 0 && failures == 0 {
		return false // nothing to reset
	}
	if last.IsZero() {
		return false
	}
	return time.Since(last) > idle
}

// LastActivity returns the time of the most recent Record() call,
// or the zero time if the detector has never been used. Callers
// (e.g. the idle-watcher) use this to decide when to Reset().
func (d *Detector) LastActivity() time.Time {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state.lastActivity
}

// Close stops the background goroutine.
func (d *Detector) Close() {
	close(d.stop)
}

func (d *Detector) loop() {
	ticker := time.NewTicker(30 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-d.stop:
			return
		case a := <-d.actions:
			d.process(a)
		case <-ticker.C:
			d.checkRate()
		}
	}
}

func (d *Detector) process(a actionRecord) {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.state.count++
	d.state.lastActivity = a.time
	if !a.success {
		d.state.failures++
	} else {
		d.state.failures = 0
	}

	// Duration check.
	if time.Since(d.state.startTime) > 30*time.Minute {
		d.trip(Trip{Type: TripDuration, Reason: "task duration exceeds 30 minutes"})
	}

	// Loop check: track last 5 coordinates (spatial actions only).
	// Zero coords mean no pointer position (chat, shell, etc.) — skip loop detection.
	if a.coordX != 0 || a.coordY != 0 {
		d.state.coordWindow = append(d.state.coordWindow, [2]float64{a.coordX, a.coordY})
		if len(d.state.coordWindow) > 5 {
			d.state.coordWindow = d.state.coordWindow[1:]
		}
		if d.checkLoop() {
			d.trip(Trip{Type: TripLoop, Reason: "same coordinates repeated 3+ times"})
		}
	}

	// Failure check.
	if d.state.failures >= 5 {
		d.trip(Trip{Type: TripFailures, Reason: "5+ consecutive failures"})
	}
}

func (d *Detector) checkRate() {
	d.mu.Lock()
	defer d.mu.Unlock()
	if timeframe := time.Since(d.state.startTime); timeframe > 0 {
		rate := float64(d.state.count) / timeframe.Minutes()
		if rate > 20 {
			d.trip(Trip{Type: TripRate, Reason: "exceeds 20 actions per minute"})
		}
	}
}

func (d *Detector) checkLoop() bool {
	if len(d.state.coordWindow) < 3 {
		return false
	}
	// Check if last 3 coordinates are identical.
	last := d.state.coordWindow[len(d.state.coordWindow)-1]
	count := 0
	for i := len(d.state.coordWindow) - 1; i >= 0; i-- {
		if d.state.coordWindow[i] == last {
			count++
		} else {
			break
		}
	}
	return count >= 3
}

func (d *Detector) trip(t Trip) {
	if d.onTrip != nil {
		d.onTrip(t)
	}
}
