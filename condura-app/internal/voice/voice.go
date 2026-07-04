package voice

import "context"

// Recorder captures audio from the microphone.
type Recorder interface {
	// Start begins mic capture. Blocks until ctx is canceled or Stop is called.
	Start(ctx context.Context) error
	// Stop halts capture and returns the recorded audio as WAV bytes.
	Stop() ([]byte, error)
	// Samples returns a channel of live PCM samples for waveform/VAD.
	Samples() <-chan []float32
}

// Transcriber converts audio bytes to text.
type Transcriber interface {
	// Transcribe performs a one-shot transcription of a complete audio clip.
	Transcribe(ctx context.Context, audio []byte) (Transcript, error)
	// TranscribeStream processes live audio samples and emits partial results.
	TranscribeStream(ctx context.Context, audio <-chan []float32) (<-chan Partial, error)
}

// Speaker converts text to speech using the OS-native voice.
type Speaker interface {
	// Speak blocks until the text is spoken or ctx is canceled.
	Speak(ctx context.Context, text string) error
	// Stop halts any in-progress speech.
	Stop()
}

// Transcript is the result of a one-shot transcription.
type Transcript struct {
	Text       string
	Language   string
	Confidence float64
	Segments   []Segment
}

// Partial is an intermediate transcription result (streaming).
type Partial struct {
	Text    string
	IsFinal bool
}

// Segment is a time-aligned piece of a transcript.
type Segment struct {
	Start float64
	End   float64
	Text  string
}
