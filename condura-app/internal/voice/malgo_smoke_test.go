//go:build darwin

package voice

import (
	"testing"

	"github.com/gen2brain/malgo"
)

// TestMalgoSmoke verifies that malgo compiles and can initialize
// the audio subsystem on this platform. This is the gate test for
// the malgo dependency — if this fails, we evaluate PortAudio fallback.
//
// Skipped in CI where no audio device is available.
func TestMalgoSmoke(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping audio smoke test in short mode")
	}

	// Initialize the malgo context with default backends.
	// This verifies: (1) malgo compiles, (2) the CGO linkage works,
	// (3) the platform audio backend is accessible.
	maCtx, err := malgo.InitContext(nil, malgo.ContextConfig{}, func(msg string) {
		t.Logf("malgo: %s", msg)
	})
	if err != nil {
		t.Fatalf("malgo.InitContext failed (no audio device?): %v", err)
	}
	defer maCtx.Free()

	// Verify we can enumerate playback devices.
	playbackDevices, err := maCtx.Devices(malgo.Playback)
	if err != nil {
		t.Logf("warning: could not enumerate playback devices: %v", err)
	} else {
		t.Logf("found %d playback device(s)", len(playbackDevices))
		for i, d := range playbackDevices {
			t.Logf("  [%d] %s (id: %v)", i, d.Name(), d.ID)
		}
	}

	// Verify we can enumerate capture devices.
	captureDevices, err := maCtx.Devices(malgo.Capture)
	if err != nil {
		t.Logf("warning: could not enumerate capture devices: %v", err)
	} else {
		t.Logf("found %d capture device(s)", len(captureDevices))
		for i, d := range captureDevices {
			t.Logf("  [%d] %s (id: %v)", i, d.Name(), d.ID)
		}
	}

	t.Log("malgo smoke test passed")
}
