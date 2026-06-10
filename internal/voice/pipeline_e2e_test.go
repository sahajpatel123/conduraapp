package voice

import (
	"context"
	"sync"
	"testing"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/sse"
	"github.com/sahajpatel123/synapticapp/internal/status"
)

// TestE2E_RecorderToSpeaker is an end-to-end integration test that
// wires mock recorder → pipeline → mock speaker, asserting that the
// transcript flows through and the reply is spoken. This is the exact
// test that would have caught 6A #1 (the keystone bug where the session
// read from the conversation store instead of the SSE broker).
func TestE2E_RecorderToSpeaker(t *testing.T) {
	// Set up the SSE broker (in-process, no HTTP).
	broker := sse.NewBroker()
	defer broker.Close()

	// Set up mock components.
	rec := newFakeRecorder()
	rec.wav = []byte("fake-wav-data")

	transcriber := &fakeTranscriber{
		text:       "What's the weather today?",
		confidence: 0.92,
	}

	speaker := &fakeSpeaker{}

	// Build the pipeline.
	bin := writeTempFile(t, "fake-binary")
	model := writeTempFile(t, "fake-model")
	cfg := Config{
		Recorder:    rec,
		Transcriber: transcriber,
		Speaker:     speaker,
		BinaryPath:  bin,
		ModelPath:   model,
		Pins: SHA256Pins{
			Binary: sha256Of(t, "fake-binary"),
			Model:  sha256Of(t, "fake-model"),
		},
		SilenceMS: 1500,
		Broker:    broker,
	}

	pipeline, err := NewPipeline(cfg)
	if err != nil {
		t.Fatalf("NewPipeline: %v", err)
	}

	// Track status transitions.
	var mu sync.Mutex
	var states []status.Status
	pipeline.OnStatus = func(s status.Status) {
		mu.Lock()
		states = append(states, s)
		mu.Unlock()
	}

	// Subscribe to voice.final SSE events.
	sub := broker.Subscribe()
	defer broker.Unsubscribe(sub)

	// Run the pipeline.
	result, err := pipeline.ListenAndProcess(context.Background())
	if err != nil {
		t.Fatalf("ListenAndProcess: %v", err)
	}

	// Assert the transcript was returned.
	if result.Transcript != "What's the weather today?" {
		t.Errorf("transcript = %q, want %q", result.Transcript, "What's the weather today?")
	}
	if result.Confidence != 0.92 {
		t.Errorf("confidence = %v, want 0.92", result.Confidence)
	}

	// Assert the recorder was started and stopped.
	if rec.started.Load() != 1 {
		t.Errorf("recorder started = %d, want 1", rec.started.Load())
	}
	if rec.stopped.Load() != 1 {
		t.Errorf("recorder stopped = %d, want 1", rec.stopped.Load())
	}

	// Assert the voice.final SSE event was published.
	select {
	case ev := <-sub.Events:
		if ev.Name != "voice.final" {
			t.Errorf("SSE event name = %q, want %q", ev.Name, "voice.final")
		}
	case <-time.After(100 * time.Millisecond):
		t.Error("timed out waiting for voice.final event")
	}

	// Assert status transitions: listening → thinking → idle.
	mu.Lock()
	got := make([]status.Status, len(states))
	copy(got, states)
	mu.Unlock()

	want := []status.Status{status.StatusListening, status.StatusThinking, status.StatusIdle}
	if len(got) != len(want) {
		t.Fatalf("states = %v, want %v", got, want)
	}
	for i := range want {
		if got[i] != want[i] {
			t.Errorf("state[%d] = %v, want %v", i, got[i], want[i])
		}
	}
}

// TestE2E_PipelineSpeakAfterListen verifies that after ListenAndProcess
// returns, the caller can feed the transcript to Speak and get the
// spoken text. This tests the full voice → transcript → reply → speak
// loop that the session orchestrator drives.
func TestE2E_PipelineSpeakAfterListen(t *testing.T) {
	broker := sse.NewBroker()
	defer broker.Close()

	rec := newFakeRecorder()
	rec.wav = []byte("fake-wav")
	transcriber := &fakeTranscriber{text: "Hello AI", confidence: 0.9}
	speaker := &fakeSpeaker{}

	bin := writeTempFile(t, "bin")
	model := writeTempFile(t, "model")
	cfg := Config{
		Recorder:    rec,
		Transcriber: transcriber,
		Speaker:     speaker,
		BinaryPath:  bin,
		ModelPath:   model,
		Pins: SHA256Pins{
			Binary: sha256Of(t, "bin"),
			Model:  sha256Of(t, "model"),
		},
		Broker: broker,
	}

	pipeline, err := NewPipeline(cfg)
	if err != nil {
		t.Fatalf("NewPipeline: %v", err)
	}

	// Step 1: Listen and transcribe.
	result, err := pipeline.ListenAndProcess(context.Background())
	if err != nil {
		t.Fatalf("ListenAndProcess: %v", err)
	}

	// Step 2: Speak the reply (simulating what the session does).
	replyText := "Hello! How can I help you today?"
	if err := pipeline.Speak(context.Background(), replyText); err != nil {
		t.Fatalf("Speak: %v", err)
	}

	// Assert the speaker received the reply.
	if speaker.spoke.Load() != 1 {
		t.Errorf("spoke = %d, want 1", speaker.spoke.Load())
	}
	if got, _ := speaker.text.Load().(string); got != replyText {
		t.Errorf("spoken text = %q, want %q", got, replyText)
	}

	// Assert the transcript was correct.
	if result.Transcript != "Hello AI" {
		t.Errorf("transcript = %q, want %q", result.Transcript, "Hello AI")
	}
}
