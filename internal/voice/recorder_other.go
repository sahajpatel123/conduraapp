//go:build !darwin

package voice

import (
	"context"
	"fmt"
	"sync"
)

// noopRecorder is a Recorder stub for non-darwin platforms.
// It compiles, but Start returns an error so callers know
// real capture is unavailable.
type noopRecorder struct {
	samplesCh  chan []float32
	sampleRate uint32
	channels   uint32
	mu         sync.Mutex
	captured   []float32
	cancel     context.CancelFunc
}

// NewRecorder creates a platform-specific Recorder. On
// non-darwin platforms this is a noop; Start returns an error.
func NewRecorder(sampleRate, channels int) Recorder {
	return &noopRecorder{
		samplesCh:  make(chan []float32, 100),
		sampleRate: uint32(sampleRate), //nolint:gosec // sampleRate is validated by config
		channels:   uint32(channels),   //nolint:gosec // channels is validated by config
	}
}

func (r *noopRecorder) Start(_ context.Context) error {
	return fmt.Errorf("audio capture not yet implemented on this platform")
}

func (r *noopRecorder) Stop() ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.cancel != nil {
		r.cancel()
	}
	samples := r.captured
	r.captured = nil
	return encodeWAV(samples, r.sampleRate, r.channels), nil
}

func (r *noopRecorder) Samples() <-chan []float32 { return r.samplesCh }
