//go:build darwin && !cgo

package voice

import (
	"context"
	"fmt"
	"sync"
)

// noopDarwinRecorder is a Recorder stub for darwin when CGO is disabled.
// It compiles, but Start returns an error so callers know real capture
// is unavailable. This file is only compiled during cross-compilation
// (e.g., CI building for darwin from Linux with CGO_ENABLED=0).
type noopDarwinRecorder struct {
	samplesCh  chan []float32
	sampleRate uint32
	channels   uint32
	mu         sync.Mutex
	captured   []float32
	cancel     context.CancelFunc
}

// NewRecorder creates a platform-specific Recorder. On darwin without
// CGO this is a noop; Start returns an error.
func NewRecorder(sampleRate, channels int) Recorder {
	return &noopDarwinRecorder{
		samplesCh:  make(chan []float32, 100),
		sampleRate: uint32(sampleRate), //nolint:gosec // sampleRate is validated by config
		channels:   uint32(channels),   //nolint:gosec // channels is validated by config
	}
}

// RecorderAvailable reports whether the platform can capture audio.
// On darwin without CGO this always returns false.
func RecorderAvailable() bool { return false }

func (r *noopDarwinRecorder) Start(_ context.Context) error {
	return fmt.Errorf("audio capture requires CGO on darwin")
}

func (r *noopDarwinRecorder) Stop() ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		r.cancel()
	}
	samples := r.captured
	r.captured = nil
	return encodeWAV(samples, r.sampleRate, r.channels), nil
}

func (r *noopDarwinRecorder) Samples() <-chan []float32 { return r.samplesCh }
