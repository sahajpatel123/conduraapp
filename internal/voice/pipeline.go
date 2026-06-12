package voice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"sync/atomic"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/sse"
	"github.com/sahajpatel123/synapticapp/internal/status"
)

// ErrSHAMismatch is returned by the Pipeline when the binary or model
// file's SHA256 doesn't match the expected pin.
var ErrSHAMismatch = errors.New("voice: SHA256 pin mismatch")

// ErrAlreadyRunning is returned by Listen when a previous pipeline
// step is still in flight.
var ErrAlreadyRunning = errors.New("voice: pipeline already running")

// SHA256Pins holds the expected SHA256 hashes for the whisper binary
// and the model file. An empty entry means "no pin set" (allowed
// only in development; production callers should always pin).
type SHA256Pins struct {
	Binary string // hex-encoded SHA256 of the whisper binary
	Model  string // hex-encoded SHA256 of the model file
}

// Verify checks the binary and model files against their expected
// SHA256 pins. If a pin is empty, that file is skipped. This is the
// single trust anchor for the voice pipeline — see MISSION §19.3
// ("whisper.cpp local").
func (p SHA256Pins) Verify(binaryPath, modelPath string) error {
	if p.Binary != "" {
		ok, got, err := verifyFileSHA(binaryPath, p.Binary)
		if err != nil {
			return fmt.Errorf("voice: verify binary: %w", err)
		}
		if !ok {
			return fmt.Errorf("%w: binary got %s, want %s",
				ErrSHAMismatch, got, p.Binary)
		}
	}
	if p.Model != "" {
		ok, got, err := verifyFileSHA(modelPath, p.Model)
		if err != nil {
			return fmt.Errorf("voice: verify model: %w", err)
		}
		if !ok {
			return fmt.Errorf("%w: model got %s, want %s",
				ErrSHAMismatch, got, p.Model)
		}
	}
	return nil
}

// verifyFileSHA returns (true, expectedSHA, nil) when the file's
// SHA256 matches expectedHex; (false, actualSHA, nil) when it
// doesn't match; or (_, _, err) on a read failure.
func verifyFileSHA(path, expectedHex string) (bool, string, error) {
	f, err := os.Open(path) // #nosec G304 -- path is operator-provided
	if err != nil {
		return false, "", err
	}
	defer func() { _ = f.Close() }()

	h := sha256.New()
	if _, err := io.Copy(h, f); err != nil {
		return false, "", err
	}
	got := hex.EncodeToString(h.Sum(nil))
	return strings.EqualFold(got, expectedHex), got, nil
}

// Config configures a Pipeline.
type Config struct {
	Recorder    Recorder
	Transcriber Transcriber
	Speaker     Speaker
	BinaryPath  string
	ModelPath   string
	Pins        SHA256Pins
	SilenceMS   int    // silence threshold for auto-submit (default 1500)
	Language    string // ISO 639-1; empty means auto-detect
	// Broker is the SSE broker for publishing voice.partial and
	// voice.final events. If nil, no SSE events are published.
	Broker *sse.Broker
}

// Pipeline orchestrates the full voice session: listen, transcribe,
// speak. It owns a small state machine and emits status updates
// through a callback.
type Pipeline struct {
	cfg Config

	mu       sync.Mutex
	busy     bool
	curState atomic.Int32

	// OnStatus fires whenever the pipeline transitions state.
	// Set this to wire the tray/overlay status updates.
	OnStatus func(status.Status)
}

// NewPipeline constructs a Pipeline from a Config. It performs the
// SHA256 verification eagerly so that a misconfigured binary fails
// fast at startup.
func NewPipeline(cfg Config) (*Pipeline, error) {
	if cfg.Recorder == nil {
		return nil, errors.New("voice: Recorder is required")
	}
	if cfg.Transcriber == nil {
		return nil, errors.New("voice: Transcriber is required")
	}
	if cfg.Speaker == nil {
		return nil, errors.New("voice: Speaker is required")
	}
	if cfg.BinaryPath == "" {
		return nil, errors.New("voice: BinaryPath is required")
	}
	if cfg.ModelPath == "" {
		return nil, errors.New("voice: ModelPath is required")
	}
	if err := cfg.Pins.Verify(cfg.BinaryPath, cfg.ModelPath); err != nil {
		return nil, err
	}
	if cfg.SilenceMS <= 0 {
		cfg.SilenceMS = 1500
	}
	return &Pipeline{cfg: cfg}, nil
}

// State returns the current pipeline state.
func (p *Pipeline) State() status.Status {
	return status.Status(p.curState.Load())
}

// setStatus updates the internal state and fires the OnStatus
// callback if one is set.
func (p *Pipeline) setStatus(s status.Status) {
	// status.Status is a small int. The int32 conversion is safe
	// for any value the enum can hold.
	p.curState.Store(int32(s)) //nolint:gosec // bounded by status enum
	if p.OnStatus != nil {
		p.OnStatus(s)
	}
}

// Result is the outcome of a Listen+Transcribe cycle.
type Result struct {
	Transcript string
	Confidence float64
}

// ListenAndProcess runs a full voice session. Steps:
//
//  1. Set status to listening
//  2. Start the recorder
//  3. Stop the recorder, get WAV bytes
//  4. Set status to thinking
//  5. Transcribe the audio
//  6. Return to idle; the caller drives the agent loop + TTS
//
// The Pipeline is the voice half of the agent loop, not the full
// loop. The caller wires the transcript into the agent and feeds
// the response back through Speak.
func (p *Pipeline) ListenAndProcess(ctx context.Context) (Result, error) {
	p.mu.Lock()
	if p.busy {
		p.mu.Unlock()
		return Result{}, ErrAlreadyRunning
	}
	p.busy = true
	p.mu.Unlock()
	defer func() {
		p.mu.Lock()
		p.busy = false
		p.mu.Unlock()
	}()

	// 1. Listening
	p.setStatus(status.StatusListening)

	// Emit voice.partial events during recording so the frontend
	// can show a live recording indicator. These fire periodically
	// with the current sample count.
	recordCtx, cancelRecord := context.WithCancel(ctx)
	defer cancelRecord()

	if p.cfg.Broker != nil {
		go p.emitPartials(recordCtx)
	}

	// 2-3. Record
	if err := p.cfg.Recorder.Start(recordCtx); err != nil {
		p.setStatus(status.StatusError)
		return Result{}, fmt.Errorf("voice: recorder start: %w", err)
	}

	wav, err := p.cfg.Recorder.Stop()
	cancelRecord()
	if err != nil {
		p.setStatus(status.StatusError)
		return Result{}, fmt.Errorf("voice: recorder stop: %w", err)
	}
	if len(wav) == 0 {
		p.setStatus(status.StatusIdle)
		return Result{}, nil
	}

	// 4-5. Transcribe
	p.setStatus(status.StatusThinking)
	transcript, err := p.cfg.Transcriber.Transcribe(ctx, wav)
	if err != nil {
		p.setStatus(status.StatusError)
		return Result{}, fmt.Errorf("voice: transcribe: %w", err)
	}

	// Emit voice.final SSE event with the completed transcript.
	if p.cfg.Broker != nil {
		p.cfg.Broker.PublishJSON("voice.final", map[string]any{
			"text":       transcript.Text,
			"confidence": transcript.Confidence,
			"language":   transcript.Language,
		})
	}

	// 6. Return to idle; the agent loop drives the next state.
	p.setStatus(status.StatusIdle)
	return Result{
		Transcript: transcript.Text,
		Confidence: transcript.Confidence,
	}, nil
}

// Speak converts text to speech using the configured Speaker.
// During playback the status is "speaking"; on completion it
// returns to "idle".
func (p *Pipeline) Speak(ctx context.Context, text string) error {
	if text == "" {
		return nil
	}
	p.mu.Lock()
	if p.busy {
		p.mu.Unlock()
		return ErrAlreadyRunning
	}
	p.busy = true
	p.mu.Unlock()
	defer func() {
		p.mu.Lock()
		p.busy = false
		p.mu.Unlock()
	}()

	p.setStatus(status.StatusSpeaking)
	defer p.setStatus(status.StatusIdle)
	return p.cfg.Speaker.Speak(ctx, text)
}

// Cancel stops the speaker and resets state to idle. Use this when
// the user interrupts a session.
func (p *Pipeline) Cancel() {
	if p.cfg.Speaker != nil {
		p.cfg.Speaker.Stop()
	}
	p.setStatus(status.StatusIdle)
}

// Stop is the Speaker interface method. Implemented so the
// pipeline can be passed as a Speaker to the session factory.
// Currently an alias for Cancel.
func (p *Pipeline) Stop() { p.Cancel() }

// HashFile returns the SHA256 hex of a file. It is exposed for
// setup scripts that need to compute pins; production code paths
// should not call this directly.
func HashFile(path string) (string, error) {
	ok, got, err := verifyFileSHA(path, "")
	if err != nil {
		return "", err
	}
	_ = ok
	return got, nil
}

// SilenceThreshold returns the configured silence threshold in
// milliseconds. Used by callers that want to tune auto-submit
// behavior.
func (p *Pipeline) SilenceThreshold() time.Duration {
	return time.Duration(p.cfg.SilenceMS) * time.Millisecond
}

// emitPartials publishes voice.partial events at a throttled rate
// while the recorder is active. Each event includes the current
// sample count so the frontend can show a live recording indicator.
// The goroutine exits when ctx is canceled (recording ends).
func (p *Pipeline) emitPartials(ctx context.Context) {
	ticker := time.NewTicker(200 * time.Millisecond)
	defer ticker.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			// Read the sample count from the recorder's Samples channel
			// without consuming it. We use a non-blocking select to
			// peek at the latest sample.
			select {
			case samples := <-p.cfg.Recorder.Samples():
				p.cfg.Broker.PublishJSON("voice.partial", map[string]any{
					"recording": true,
					"samples":   len(samples),
				})
			default:
				// No new samples yet; emit a generic recording event.
				p.cfg.Broker.PublishJSON("voice.partial", map[string]any{
					"recording": true,
				})
			}
		}
	}
}
