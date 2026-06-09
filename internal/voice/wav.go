package voice

import (
	"bytes"
	"encoding/binary"
	"math"
)

const maxAmplitude = 32767

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
