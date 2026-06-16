// Package voice — wake word / hotword detection (Phase 14E).
//
// WakeWordDetector listens for a specific phrase ("hey synaptic")
// on the microphone stream. When detected, it signals the voice
// pipeline to begin active listening.
//
// The default implementation uses openWakeWord (ONNX model) but
// the interface allows swapping for different backends.
package voice

import "context"

// WakeWordDetector listens for a hotword on the microphone stream.
type WakeWordDetector interface {
	// Start begins listening. Signals detection on the returned
	// channel. Blocks until ctx is canceled or Stop is called.
	Start(ctx context.Context) (<-chan WakeEvent, error)
	// Stop halts detection.
	Stop() error
	// Status returns whether the detector is ready.
	Status() WakeStatus
}

// WakeEvent is emitted when the wake word is detected or on error.
type WakeEvent struct {
	Detected   bool    `json:"detected"`
	Confidence float64 `json:"confidence"`
	Error      string  `json:"error,omitempty"`
}

// WakeStatus reports the detector's current state.
type WakeStatus struct {
	Available   bool   `json:"available"`
	ModelLoaded bool   `json:"model_loaded"`
	Listening   bool   `json:"listening"`
	Hotword     string `json:"hotword"`
	Error       string `json:"error,omitempty"`
}

// NoopWakeDetector is a stub that never detects the wake word.
// Used when wake word is disabled or the ONNX model is unavailable.
type NoopWakeDetector struct{}

func (NoopWakeDetector) Start(ctx context.Context) (<-chan WakeEvent, error) {
	ch := make(chan WakeEvent)
	go func() {
		defer close(ch)
		<-ctx.Done()
	}()
	return ch, nil
}
func (NoopWakeDetector) Stop() error        { return nil }
func (NoopWakeDetector) Status() WakeStatus { return WakeStatus{Available: false} }
