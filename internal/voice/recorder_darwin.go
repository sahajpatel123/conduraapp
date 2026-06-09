package voice

import (
	"bytes"
	"context"
	"encoding/binary"
	"fmt"
	"math"
	"sync"
)

const maxAmplitude = 32767

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

// encodeWAV encodes PCM float32 samples to WAV format.
func encodeWAV(samples []float32, sampleRate, channels uint32) []byte {
	if len(samples) == 0 {
		return nil
	}

	// Convert float32 to int16.
	intSamples := make([]int16, len(samples))
	for i, s := range samples {
		s = float32(math.Max(-1, math.Min(1, float64(s))))
		intSamples[i] = int16(s * maxAmplitude) //nolint:gosec // safe after clamping
	}

	var buf bytes.Buffer

	// RIFF header.
	buf.WriteString("RIFF")
	dataSize := uint32(len(intSamples)) * 2 //nolint:gosec // len is bounded by audio duration
	buf.Write(LE32(8 + dataSize))
	buf.WriteString("WAVE")

	// fmt chunk.
	buf.WriteString("fmt ")
	buf.Write(LE32(16))
	buf.Write(LE16(1))                // PCM format.
	buf.Write(LE16(uint16(channels))) //nolint:gosec // channels is 1 or 2
	buf.Write(LE32(sampleRate))
	buf.Write(LE32(sampleRate * channels * 2))
	buf.Write(LE16(uint16(channels) * 2)) //nolint:gosec // channels is 1 or 2
	buf.Write(LE16(16))

	// data chunk.
	buf.WriteString("data")
	buf.Write(LE32(dataSize))

	for _, s := range intSamples {
		_ = binary.Write(&buf, binary.LittleEndian, s)
	}

	return buf.Bytes()
}

// LE16 encodes a uint16 as little-endian bytes.
func LE16(v uint16) []byte {
	b := make([]byte, 2)
	binary.LittleEndian.PutUint16(b, v)
	return b
}

// LE32 encodes a uint32 as little-endian bytes.
func LE32(v uint32) []byte {
	b := make([]byte, 4)
	binary.LittleEndian.PutUint32(b, v)
	return b
}
