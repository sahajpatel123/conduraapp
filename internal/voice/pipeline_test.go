package voice

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/sahajpatel123/synapticapp/internal/status"
)

// fakeRecorder is a stub Recorder. It records the calls and returns
// a configurable WAV byte slice.
type fakeRecorder struct {
	wav      []byte
	startErr error
	stopErr  error
	started  atomic.Int32
	stopped  atomic.Int32
	samples  chan []float32
}

func newFakeRecorder() *fakeRecorder {
	return &fakeRecorder{samples: make(chan []float32, 8)}
}

func (r *fakeRecorder) Start(_ context.Context) error {
	r.started.Add(1)
	return r.startErr
}

func (r *fakeRecorder) Stop() ([]byte, error) {
	r.stopped.Add(1)
	if r.stopErr != nil {
		return nil, r.stopErr
	}
	return r.wav, nil
}

func (r *fakeRecorder) Samples() <-chan []float32 { return r.samples }

// fakeTranscriber is a stub Transcriber.
type fakeTranscriber struct {
	text       string
	confidence float64
	err        error
	calls      atomic.Int32
}

func (t *fakeTranscriber) Transcribe(_ context.Context, _ []byte) (Transcript, error) {
	t.calls.Add(1)
	if t.err != nil {
		return Transcript{}, t.err
	}
	return Transcript{Text: t.text, Confidence: t.confidence}, nil
}

func (t *fakeTranscriber) TranscribeStream(_ context.Context, _ <-chan []float32) (<-chan Partial, error) {
	out := make(chan Partial)
	close(out)
	return out, nil
}

// fakeSpeaker is a stub Speaker.
type fakeSpeaker struct {
	spoke atomic.Int32
	err   error
	text  atomic.Value // string
}

func (s *fakeSpeaker) Speak(_ context.Context, text string) error {
	s.spoke.Add(1)
	s.text.Store(text)
	return s.err
}

func (s *fakeSpeaker) Stop() {}

// writeTempFile writes content to a temp file and returns its path.
// The file is auto-cleaned by t.TempDir().
func writeTempFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "blob")
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatalf("write temp: %v", err)
	}
	return path
}

func sha256Of(t *testing.T, content string) string {
	t.Helper()
	h := sha256.Sum256([]byte(content))
	return hex.EncodeToString(h[:])
}

func validConfig(t *testing.T) Config {
	t.Helper()
	bin := writeTempFile(t, "fake-binary")
	model := writeTempFile(t, "fake-model")
	return Config{
		Recorder:    newFakeRecorder(),
		Transcriber: &fakeTranscriber{text: "hello", confidence: 0.95},
		Speaker:     &fakeSpeaker{},
		BinaryPath:  bin,
		ModelPath:   model,
		Pins: SHA256Pins{
			Binary: sha256Of(t, "fake-binary"),
			Model:  sha256Of(t, "fake-model"),
		},
		SilenceMS: 1500,
	}
}

func TestNewPipeline_RequiresComponents(t *testing.T) {
	tests := []struct {
		name string
		mut  func(c *Config)
	}{
		{"nil recorder", func(c *Config) { c.Recorder = nil }},
		{"nil transcriber", func(c *Config) { c.Transcriber = nil }},
		{"nil speaker", func(c *Config) { c.Speaker = nil }},
		{"empty binary", func(c *Config) { c.BinaryPath = "" }},
		{"empty model", func(c *Config) { c.ModelPath = "" }},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			cfg := validConfig(t)
			tt.mut(&cfg)
			if _, err := NewPipeline(cfg); err == nil {
				t.Fatal("expected error, got nil")
			}
		})
	}
}

func TestNewPipeline_VerifiesSHA(t *testing.T) {
	cfg := validConfig(t)
	cfg.Pins.Binary = "0000000000000000000000000000000000000000000000000000000000000000"
	_, err := NewPipeline(cfg)
	if err == nil {
		t.Fatal("expected SHA mismatch error")
	}
}

func TestNewPipeline_BadBinaryPath(t *testing.T) {
	cfg := validConfig(t)
	cfg.BinaryPath = filepath.Join(t.TempDir(), "does-not-exist")
	_, err := NewPipeline(cfg)
	if err == nil {
		t.Fatal("expected error for missing binary")
	}
}

func TestNewPipeline_DefaultsSilenceMS(t *testing.T) {
	cfg := validConfig(t)
	cfg.SilenceMS = 0
	p, err := NewPipeline(cfg)
	if err != nil {
		t.Fatalf("NewPipeline: %v", err)
	}
	if got := p.SilenceThreshold().Milliseconds(); got != 1500 {
		t.Errorf("default silence = %dms, want 1500ms", got)
	}
}

func TestSHA256Pins_VerifyOK(t *testing.T) {
	bin := writeTempFile(t, "data")
	model := writeTempFile(t, "more")
	pins := SHA256Pins{
		Binary: sha256Of(t, "data"),
		Model:  sha256Of(t, "more"),
	}
	if err := pins.Verify(bin, model); err != nil {
		t.Fatalf("Verify: %v", err)
	}
}

func TestSHA256Pins_VerifyEmptyPinsAllowed(t *testing.T) {
	bin := writeTempFile(t, "data")
	model := writeTempFile(t, "more")
	if err := (SHA256Pins{}).Verify(bin, model); err != nil {
		t.Fatalf("empty pins should be allowed in dev, got: %v", err)
	}
}

func TestSHA256Pins_VerifyMismatch(t *testing.T) {
	bin := writeTempFile(t, "data")
	model := writeTempFile(t, "more")
	pins := SHA256Pins{Binary: sha256Of(t, "different")}
	err := pins.Verify(bin, model)
	if err == nil {
		t.Fatal("expected mismatch")
	}
}

func TestPipeline_ListenAndProcess_TranscribesAndReturns(t *testing.T) {
	cfg := validConfig(t)
	rec := newFakeRecorder()
	rec.wav = []byte("fake-wav-bytes")
	cfg.Recorder = rec
	cfg.Transcriber = &fakeTranscriber{text: "what's the weather", confidence: 0.88}

	p, err := NewPipeline(cfg)
	if err != nil {
		t.Fatalf("NewPipeline: %v", err)
	}

	var states []status.Status
	p.OnStatus = func(s status.Status) { states = append(states, s) }

	result, err := p.ListenAndProcess(context.Background())
	if err != nil {
		t.Fatalf("ListenAndProcess: %v", err)
	}
	if result.Transcript != "what's the weather" {
		t.Errorf("transcript = %q", result.Transcript)
	}
	if result.Confidence != 0.88 {
		t.Errorf("confidence = %v", result.Confidence)
	}
	if rec.started.Load() != 1 {
		t.Errorf("recorder started = %d, want 1", rec.started.Load())
	}
	if rec.stopped.Load() != 1 {
		t.Errorf("recorder stopped = %d, want 1", rec.stopped.Load())
	}
	// States: listening → thinking → idle.
	want := []status.Status{status.StatusListening, status.StatusThinking, status.StatusIdle}
	if len(states) != len(want) {
		t.Fatalf("states = %v, want %v", states, want)
	}
	for i := range want {
		if states[i] != want[i] {
			t.Errorf("state[%d] = %v, want %v", i, states[i], want[i])
		}
	}
}

func TestPipeline_ListenAndProcess_EmptyAudio(t *testing.T) {
	cfg := validConfig(t)
	rec := newFakeRecorder()
	rec.wav = nil
	cfg.Recorder = rec

	p, err := NewPipeline(cfg)
	if err != nil {
		t.Fatalf("NewPipeline: %v", err)
	}
	result, err := p.ListenAndProcess(context.Background())
	if err != nil {
		t.Fatalf("ListenAndProcess: %v", err)
	}
	if result.Transcript != "" {
		t.Errorf("expected empty transcript, got %q", result.Transcript)
	}
	if p.State() != status.StatusIdle {
		t.Errorf("state = %v, want idle", p.State())
	}
}

func TestPipeline_ListenAndProcess_RecorderStartError(t *testing.T) {
	cfg := validConfig(t)
	rec := newFakeRecorder()
	rec.startErr = errFake
	cfg.Recorder = rec

	p, err := NewPipeline(cfg)
	if err != nil {
		t.Fatalf("NewPipeline: %v", err)
	}
	_, err = p.ListenAndProcess(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if p.State() != status.StatusError {
		t.Errorf("state = %v, want error", p.State())
	}
}

func TestPipeline_ListenAndProcess_TranscribeError(t *testing.T) {
	cfg := validConfig(t)
	rec := newFakeRecorder()
	rec.wav = []byte("data")
	cfg.Recorder = rec
	cfg.Transcriber = &fakeTranscriber{err: errFake}

	p, err := NewPipeline(cfg)
	if err != nil {
		t.Fatalf("NewPipeline: %v", err)
	}
	_, err = p.ListenAndProcess(context.Background())
	if err == nil {
		t.Fatal("expected error")
	}
	if p.State() != status.StatusError {
		t.Errorf("state = %v, want error", p.State())
	}
}

func TestPipeline_ListenAndProcess_AlreadyRunning(t *testing.T) {
	cfg := validConfig(t)
	rec := newFakeRecorder()
	rec.wav = []byte("data")
	cfg.Recorder = rec

	p, err := NewPipeline(cfg)
	if err != nil {
		t.Fatalf("NewPipeline: %v", err)
	}
	// Force the busy flag via the first call.
	p.busy = true
	defer func() { p.busy = false }()

	_, err = p.ListenAndProcess(context.Background())
	if !errors.Is(err, ErrAlreadyRunning) {
		t.Errorf("err = %v, want ErrAlreadyRunning", err)
	}
}

func TestPipeline_Speak(t *testing.T) {
	cfg := validConfig(t)
	sp := &fakeSpeaker{}
	cfg.Speaker = sp

	p, err := NewPipeline(cfg)
	if err != nil {
		t.Fatalf("NewPipeline: %v", err)
	}
	var states []status.Status
	p.OnStatus = func(s status.Status) { states = append(states, s) }

	if err := p.Speak(context.Background(), "hello world"); err != nil {
		t.Fatalf("Speak: %v", err)
	}
	if sp.spoke.Load() != 1 {
		t.Errorf("spoke = %d, want 1", sp.spoke.Load())
	}
	if got, _ := sp.text.Load().(string); got != "hello world" {
		t.Errorf("spoken text = %q", got)
	}
	want := []status.Status{status.StatusSpeaking, status.StatusIdle}
	if len(states) != len(want) {
		t.Fatalf("states = %v, want %v", states, want)
	}
}

func TestPipeline_SpeakEmpty(t *testing.T) {
	cfg := validConfig(t)
	sp := &fakeSpeaker{}
	cfg.Speaker = sp
	p, err := NewPipeline(cfg)
	if err != nil {
		t.Fatalf("NewPipeline: %v", err)
	}
	if err := p.Speak(context.Background(), ""); err != nil {
		t.Fatalf("Speak(\"\"): %v", err)
	}
	if sp.spoke.Load() != 0 {
		t.Errorf("empty text should not invoke speaker, got %d", sp.spoke.Load())
	}
}

func TestPipeline_Cancel(t *testing.T) {
	cfg := validConfig(t)
	sp := &fakeSpeaker{}
	cfg.Speaker = sp
	p, err := NewPipeline(cfg)
	if err != nil {
		t.Fatalf("NewPipeline: %v", err)
	}
	p.setStatus(status.StatusSpeaking)
	p.Cancel()
	if p.State() != status.StatusIdle {
		t.Errorf("state = %v, want idle", p.State())
	}
}

func TestHashFile(t *testing.T) {
	path := writeTempFile(t, "abc")
	got, err := HashFile(path)
	if err != nil {
		t.Fatalf("HashFile: %v", err)
	}
	if got != sha256Of(t, "abc") {
		t.Errorf("HashFile = %s, want %s", got, sha256Of(t, "abc"))
	}
}

func TestHashFile_Missing(t *testing.T) {
	_, err := HashFile("/nonexistent/path")
	if err == nil {
		t.Fatal("expected error for missing file")
	}
}

// errFake is a sentinel used in the tests above.
var errFake = errFakeSentinel{}

type errFakeSentinel struct{}

func (errFakeSentinel) Error() string { return "fake error" }
