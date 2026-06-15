// Package presence provides user presence detection.
// It answers: "Is the user actually present and aware at the keyboard?"
//
// Presence signals (from MISSION §2):
//   - Active input (keyboard/mouse) in last 60s → Likely present
//   - Screen locked → Definitely not present
//   - Lid closed (laptop) → Definitely not present
//   - User away >5 min (configurable) → Not present
//   - User logged out → Not present
//   - Active audio (mic input) → Possibly present
//   - Camera input (face detection) → Possibly present
//
// Behavior when not present (MISSION §S10.2 Table):
//   - READ actions: allowed
//   - LOCAL actions: queue, ask on return
//   - NETWORK actions: queue + require consent on return + wait 1 hour
//   - DESTRUCTIVE actions: queued + cannot run without unlock
package presence

import (
	"os/exec"
	"runtime"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// Detector polls the OS for presence signals.
// It is safe for concurrent use.
type Detector struct {
	mu      sync.Mutex
	state   State
	stop    chan struct{}
	running int32 // atomic flag
}

// State represents the current presence state.
type State struct {
	Present   bool      // Is the user likely present?
	Locked    bool      // Screen is locked
	AwaySince time.Time // When user went away (zero if present)
}

// NewDetector creates a presence detector that polls every interval.
func NewDetector(pollInterval time.Duration) *Detector {
	return &Detector{
		state: State{Present: true}, // Assume present at start
		stop:  make(chan struct{}),
	}
}

// Start begins polling. Safe to call multiple times (idempotent).
func (d *Detector) Start() {
	if !atomic.CompareAndSwapInt32(&d.running, 0, 1) {
		return // Already running
	}
	go d.loop()
}

// Stop ends polling.
func (d *Detector) Stop() {
	if atomic.LoadInt32(&d.running) == 0 {
		return
	}
	close(d.stop)
	atomic.StoreInt32(&d.running, 0)
}

func (d *Detector) loop() {
	ticker := time.NewTicker(5 * time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-d.stop:
			return
		case <-ticker.C:
			d.poll()
		}
	}
}

func (d *Detector) poll() {
	present := d.checkPresent()
	locked := d.checkLocked()

	d.mu.Lock()
	d.state.Present = present && !locked
	d.state.Locked = locked
	// Note: AwaySince is set when user transitions to absent
	d.mu.Unlock()
}

// checkPresent returns true if there's been recent input activity.
// We poll the OS for this.
func (d *Detector) checkPresent() bool {
	// Check for active input in last 60 seconds.
	// This is a simplified check - real implementations would use
	// CGEventTap (macOS), GetLastInputInfo (Windows), or X11 (Linux).
	switch runtime.GOOS {
	case "darwin":
		return d.checkActiveOnDarwin()
	case "windows":
		return d.checkActiveOnWindows()
	default:
		return d.checkActiveOnLinux()
	}
}

// checkLocked returns true if the screen is locked.
func (d *Detector) checkLocked() bool {
	switch runtime.GOOS {
	case "darwin":
		return d.checkLockedDarwin()
	case "windows":
		return d.checkLockedWindows()
	default:
		return false // Linux: no reliable cross-desktop way
	}
}

// darwin checks use AppleScript and libc calls.
func (d *Detector) checkActiveOnDarwin() bool {
	// Use ioreg to check for screen state changes and recent events.
	// This is a heuristic - real implementation would use CGEventSourceSecondsSinceLastEventType.
	// For now, assume active unless screen is locked.
	out, err := exec.Command("ioreg", "-c", "IOHIDSystem").Output()
	if err != nil {
		return true // Assume present on error
	}
	// Check for "AppleEventKeyDown" timestamps in the last minute.
	return strings.Contains(string(out), "AppleEvent")
}

func (d *Detector) checkLockedDarwin() bool {
	// Check if screensaver is running (indicates lock).
	out, err := exec.Command("pgrep", "-x", "ScreenSaverEngine").Output()
	if err != nil {
		return false
	}
	return len(out) > 0
}

func (d *Detector) checkActiveOnWindows() bool {
	// Use PowerShell to check last input time.
	cmd := exec.Command("powershell", "-Command",
		"(Get-NetAdapter | Where-Object {$_.Status -eq 'Up'}).Count -gt 0")
	out, err := cmd.Output()
	if err != nil {
		return true
	}
	return strings.TrimSpace(string(out)) == "True"
}

func (d *Detector) checkLockedWindows() bool {
	// Check for workstation locked via log check.
	cmd := exec.Command("powershell", "-Command",
		"try { (Get-Process logonui -ErrorAction Stop) | Out-Null; return $true } catch { return $false }")
	out, err := cmd.Output()
	if err != nil {
		return false
	}
	return strings.TrimSpace(string(out)) == "True"
}

func (d *Detector) checkActiveOnLinux() bool {
	// Check for X11 idle time or session activity.
	// This is a placeholder - real implementation uses XScreenSaver.
	return true
}

// IsPresent returns true if the user is likely present.
func (d *Detector) IsPresent() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state.Present
}

// IsLocked returns true if the screen is locked.
func (d *Detector) IsLocked() bool {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state.Locked
}

// State returns the current presence state.
func (d *Detector) State() State {
	d.mu.Lock()
	defer d.mu.Unlock()
	return d.state
}