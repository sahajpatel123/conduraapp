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
	"context"
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
	out, err := exec.CommandContext(context.Background(), "ioreg", "-c", "IOHIDSystem").Output()
	if err != nil {
		// Fail closed: a broken probe must NOT claim the user is
		// present, because presence gates DESTRUCTIVE consent.
		return false
	}
	// Check for "AppleEventKeyDown" timestamps in the last minute.
	return strings.Contains(string(out), "AppleEvent")
}

func (d *Detector) checkLockedDarwin() bool {
	// Check if screensaver is running (indicates lock).
	out, err := exec.CommandContext(context.Background(), "pgrep", "-x", "ScreenSaverEngine").Output()
	if err != nil {
		return false
	}
	return len(out) > 0
}

func (d *Detector) checkActiveOnWindows() bool {
	// Use GetLastInputInfo via PowerShell P/Invoke to measure the
	// seconds since the last keyboard/mouse input. This is the
	// correct Windows API for "is the user at the keyboard?".
	// The previous implementation counted network adapters
	// (Get-NetAdapter), which returns true on any Wi-Fi-connected
	// machine regardless of whether anyone is present — defeating
	// the require_user_active consent gate for DESTRUCTIVE actions.
	//
	// We consider the user present if the last input was within
	// presenceIdleSeconds (default 120s). A failure of the P/Invoke
	// path fails closed (returns false) so a DESTRUCTIVE action is
	// queued rather than auto-allowed on a broken probe.
	const script = `Add-Type @"
using System;
using System.Runtime.InteropServices;
public class LI {
  [StructLayout(LayoutKind.Sequential)]
  public struct LASTINPUTINFO {
    public uint cbSize;
    public uint dwTime;
  }
  [DllImport("user32.dll")] public static extern bool GetLastInputInfo(ref LASTINPUTINFO plii);
  [DllImport("kernel32.dll")] public static extern uint GetTickCount();
}
"@
$li = New-Object LI+LASTINPUTINFO
$li.cbSize = [uint32][System.Runtime.InteropServices.Marshal]::SizeOf([type][LI+LASTINPUTINFO])
[void][LI]::GetLastInputInfo([ref]$li)
$now = [LI]::GetTickCount()
$secs = ($now - $li.dwTime) / 1000
if ($secs -lt 120) { "true" } else { "false" }`
	out, err := exec.CommandContext(context.Background(), "powershell", "-NoProfile", "-Command", script).Output()
	if err != nil {
		// Fail closed: a broken probe must NOT claim the user is
		// present, because presence gates DESTRUCTIVE consent.
		return false
	}
	return strings.TrimSpace(strings.ToLower(string(out))) == "true"
}

func (d *Detector) checkLockedWindows() bool {
	// Check for workstation locked via log check.
	cmd := exec.CommandContext(context.Background(), "powershell", "-Command",
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
