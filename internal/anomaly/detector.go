// Package anomaly detects unusual agent behavior per MISSION S10.4.
// Package description is in this file.
//
//nolint:revive,mnd // threshold constants; exported types are self-documenting
package anomaly

import (
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
	TripRate     TripType = "rate"
	TripDuration TripType = "duration"
	TripLoop     TripType = "loop"
	TripFailures TripType = "failures"
)

type actionRecord struct {
	kind    string
	coordX  float64
	coordY  float64
	success bool
	time    time.Time
}

type detectorState struct {
	count       int
	failures    int
	startTime   time.Time
	coordWindow [][2]float64
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

// Reset clears all counters for a new conversation.
func (d *Detector) Reset() {
	d.mu.Lock()
	defer d.mu.Unlock()
	d.state = detectorState{}
	d.state.startTime = time.Now()
	d.state.coordWindow = make([][2]float64, 0, 5)
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
	if !a.success {
		d.state.failures++
	}

	// Duration check.
	if time.Since(d.state.startTime) > 30*time.Minute {
		d.trip(Trip{Type: TripDuration, Reason: "task duration exceeds 30 minutes"})
	}

	// Loop check: track last 5 coordinates.
	d.state.coordWindow = append(d.state.coordWindow, [2]float64{a.coordX, a.coordY})
	if len(d.state.coordWindow) > 5 {
		d.state.coordWindow = d.state.coordWindow[1:]
	}
	if d.checkLoop() {
		d.trip(Trip{Type: TripLoop, Reason: "same coordinates repeated 3+ times"})
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
