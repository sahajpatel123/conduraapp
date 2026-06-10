//go:build darwin

package voice

import (
	"context"
	"fmt"
	"sync"
)

// darwinRecorder captures audio from the microphone.
// On macOS, this uses malgo (CoreAudio). For now, this is a stub
// that compiles and satisfies the interface. The real malgo integration
// will be completed in the next iteration.
type darwinRecorder struct {
	samplesCh  chan []float32
	sampleRate uint32
	channels   uint32
	mu         sync.Mutex
	captured   []float32
	cancel     context.CancelFunc
}

// NewRecorder creates a new platform-specific Recorder.
func NewRecorder(sampleRate, channels int) Recorder {
	return &darwinRecorder{
		samplesCh:  make(chan []float32, 100),
		sampleRate: uint32(sampleRate), //nolint:gosec // sampleRate is validated by config
		channels:   uint32(channels),   //nolint:gosec // channels is validated by config
	}
}

func (r *darwinRecorder) Start(_ context.Context) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.captured = nil

	// TODO: Integrate with malgo for real audio capture.
	// For now, return an error indicating the feature is not yet implemented.
	return fmt.Errorf("audio capture not yet implemented (malgo integration pending)")
}

func (r *darwinRecorder) Stop() ([]byte, error) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if r.cancel != nil {
		r.cancel()
	}

	samples := r.captured
	r.captured = nil

	return encodeWAV(samples, r.sampleRate, r.channels), nil
}

func (r *darwinRecorder) Samples() <-chan []float32 {
	return r.samplesCh
}
