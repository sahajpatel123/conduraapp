// Package crash implements privacy-first crash reporting per MISSION S2.
// Crashes are captured locally as stack+dump files. Telemetry is opt-in:
// only a stack-trace SHA256 hash + version leaves the machine, and only
// when the user explicitly enables crash reporting. Source, screen, and
// PII never leave the machine.
//
// Off by default. Honoring "No tracking, period."
package crash

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"os"
	"path/filepath"
	"runtime/debug"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/version"
)

// Report captures a crash locally.
type Report struct {
	StackHash string    `json:"stack_hash"`
	Version   string    `json:"version"`
	Platform  string    `json:"platform"`
	Time      time.Time `json:"time"`
	Stack     []byte    `json:"-"` // never leaves the machine
}

// Capture writes the current goroutine stack to a local crash report
// and returns the report for optional telemetry submission.
func Capture(recovered any) *Report {
	stack := debug.Stack()
	h := sha256.Sum256(stack)
	r := &Report{
		StackHash: hex.EncodeToString(h[:]),
		Version:   version.Get().Version,
		Platform:  fmt.Sprintf("%s/%s", version.Get().GoVersion, version.Get().Platform),
		Time:      time.Now().UTC(),
		Stack:     stack,
	}
	r.writeLocal()
	return r
}

// writeLocal persists the crash dump to ~/.condura/crashes/.
func (r *Report) writeLocal() {
	home, err := os.UserHomeDir()
	if err != nil {
		return
	}
	dir := filepath.Join(home, ".condura", "crashes")
	_ = os.MkdirAll(dir, 0o700)
	path := filepath.Join(dir, fmt.Sprintf("crash-%s.log", r.Time.Format("20060102-150405")))
	data := fmt.Sprintf("version: %s\nplatform: %s\ntime: %s\nstack_hash: %s\n\n%s",
		r.Version, r.Platform, r.Time.Format(time.RFC3339), r.StackHash, string(r.Stack))
	_ = os.WriteFile(path, []byte(data), 0o600)
}

// TelemetryPayload is the minimal, privacy-respecting payload that
// leaves the machine — only when crash reporting is explicitly enabled.
type TelemetryPayload struct {
	StackHash string `json:"stack_hash"`
	Version   string `json:"version"`
	Platform  string `json:"platform"`
}

// ToTelemetry returns the hash-only payload for opt-in submission.
func (r *Report) ToTelemetry() *TelemetryPayload {
	return &TelemetryPayload{
		StackHash: r.StackHash,
		Version:   r.Version,
		Platform:  r.Platform,
	}
}

// Recover is a defer-recover helper for daemon goroutines.
// Call as: defer crash.Recover()
func Recover() {
	if r := recover(); r != nil {
		_ = Capture(r)
	}
}
