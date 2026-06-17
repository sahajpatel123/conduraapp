package voice

import (
	"bufio"
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"sync"
)

// OpenWakeWordDetector implements WakeWordDetector using the openWakeWord library.
// It runs a subprocess that processes audio and emits wake word events via JSON on stdout.
type OpenWakeWordDetector struct {
	binaryPath string
	modelPath  string
	threshold  float64
	mu         sync.Mutex
	cancel     context.CancelFunc
}

// OpenWakeWordOption configures an OpenWakeWordDetector.
type OpenWakeWordOption func(*OpenWakeWordDetector)

// WithThreshold sets the wake word detection threshold.
func WithThreshold(t float64) OpenWakeWordOption {
	return func(d *OpenWakeWordDetector) {
		d.threshold = t
	}
}

// NewOpenWakeWordDetector creates a new openWakeWord-based WakeWordDetector.
// The binaryPath is the path to the openWakeWord binary.
// The modelPath is the path to the ONNX model file.
func NewOpenWakeWordDetector(binaryPath, modelPath string, opts ...OpenWakeWordOption) *OpenWakeWordDetector {
	d := &OpenWakeWordDetector{
		binaryPath: binaryPath,
		modelPath:  modelPath,
		threshold:  0.5, //nolint:mnd // default threshold
	}
	for _, opt := range opts {
		opt(d)
	}
	return d
}

// openWakeWordEvent represents a JSON event from the openWakeWord binary.
type openWakeWordEvent struct {
	Type      string  `json:"type"`      // "detection" or "error"
	Name      string  `json:"name"`      // wake word name
	Score     float64 `json:"score"`     // detection confidence
	Timestamp float64 `json:"timestamp"` // audio timestamp in seconds
	Error     string  `json:"error"`     // error message if type is "error"
}

// Start begins listening for the wake word. Signals detection on the returned
// channel. Blocks until ctx is canceled or Stop is called.
func (d *OpenWakeWordDetector) Start(ctx context.Context) (<-chan WakeEvent, error) {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.binaryPath == "" {
		return nil, fmt.Errorf("openwake detector: binary path required")
	}

	ctx, cancel := context.WithCancel(ctx)
	d.cancel = cancel

	// Start the openWakeWord binary as a subprocess.
	cmd := exec.CommandContext(ctx, d.binaryPath, //nolint:gosec // binaryPath is from config
		"--model", d.modelPath,
		"--threshold", fmt.Sprintf("%.2f", d.threshold),
		"--format", "json",
	)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		cancel()
		d.cancel = nil
		return nil, fmt.Errorf("openwake detector: create stdout pipe: %w", err)
	}

	if err := cmd.Start(); err != nil {
		cancel()
		d.cancel = nil
		return nil, fmt.Errorf("openwake detector: start process: %w", err)
	}

	// Channel for wake events.
	events := make(chan WakeEvent, 10)

	// Goroutine to read events from stdout.
	go func() {
		defer close(events)
		defer func() {
			_ = cmd.Process.Kill()
			_ = cmd.Wait()
		}()

		scanner := bufio.NewScanner(stdout)
		for scanner.Scan() {
			line := scanner.Text()
			if line == "" {
				continue
			}

			var event openWakeWordEvent
			if err := json.Unmarshal([]byte(line), &event); err != nil {
				continue
			}

			if event.Type == "error" {
				continue
			}

			events <- WakeEvent{
				Detected:   true,
				Confidence: event.Score,
			}
		}
	}()

	return events, nil
}

// Stop halts detection.
func (d *OpenWakeWordDetector) Stop() error {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.cancel != nil {
		d.cancel()
	}
	return nil
}

// Status returns whether the detector is ready.
func (d *OpenWakeWordDetector) Status() WakeStatus {
	d.mu.Lock()
	defer d.mu.Unlock()

	return WakeStatus{
		Available:   d.binaryPath != "",
		ModelLoaded: d.modelPath != "",
		Listening:   d.cancel != nil,
		Hotword:     "hey_synaptic",
	}
}
