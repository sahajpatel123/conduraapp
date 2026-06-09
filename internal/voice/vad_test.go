package voice

import (
	"testing"
)

func TestDetectVoiceActivity_Silence(t *testing.T) {
	samples := make([]float32, 16000) // 1 second of silence
	if DetectVoiceActivity(samples, 0.015) {
		t.Error("expected silence to be detected as not voice")
	}
}

func TestDetectVoiceActivity_Speech(t *testing.T) {
	// Generate a sine wave (simulates speech).
	samples := make([]float32, 16000)
	for i := range samples {
		samples[i] = float32(0.5) // loud signal
	}
	if !DetectVoiceActivity(samples, 0.015) {
		t.Error("expected speech to be detected as voice")
	}
}

func TestDetectVoiceActivity_Empty(t *testing.T) {
	if DetectVoiceActivity(nil, 0.015) {
		t.Error("expected empty buffer to be detected as not voice")
	}
}

func TestTrimSilence_LeadingAndTrailing(t *testing.T) {
	// 10 silence, 5 signal, 10 silence.
	samples := make([]float32, 25)
	for i := 10; i < 15; i++ {
		samples[i] = 0.5
	}
	trimmed := TrimSilence(samples, 0.015)
	if len(trimmed) != 5 {
		t.Errorf("expected 5 samples, got %d", len(trimmed))
	}
}

func TestTrimSilence_AllSilence(t *testing.T) {
	samples := make([]float32, 100)
	trimmed := TrimSilence(samples, 0.015)
	if trimmed != nil {
		t.Error("expected nil for all-silence input")
	}
}

func TestTrimSilence_NoSilence(t *testing.T) {
	samples := []float32{0.5, 0.5, 0.5}
	trimmed := TrimSilence(samples, 0.015)
	if len(trimmed) != 3 {
		t.Errorf("expected 3 samples, got %d", len(trimmed))
	}
}

func TestTrimSilence_Empty(t *testing.T) {
	trimmed := TrimSilence(nil, 0.015)
	if trimmed != nil {
		t.Error("expected nil for empty input")
	}
}

func TestComputeRMS_Silence(t *testing.T) {
	samples := make([]float32, 100)
	rms := computeRMS(samples)
	if rms != 0 {
		t.Errorf("expected 0, got %f", rms)
	}
}

func TestComputeRMS_KnownValue(t *testing.T) {
	// RMS of constant 0.5 signal should be 0.5.
	samples := make([]float32, 100)
	for i := range samples {
		samples[i] = 0.5
	}
	rms := computeRMS(samples)
	if rms < 0.49 || rms > 0.51 {
		t.Errorf("expected ~0.5, got %f", rms)
	}
}
