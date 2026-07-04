package voice

import "math"

// DetectVoiceActivity checks if a buffer of PCM float32 samples contains
// voice activity by computing the RMS energy and comparing it to the threshold.
// Returns true if the RMS exceeds the threshold, indicating speech.
func DetectVoiceActivity(samples []float32, threshold float64) bool {
	if len(samples) == 0 {
		return false
	}
	rms := computeRMS(samples)
	return rms > threshold
}

// computeRMS computes the root mean square of PCM samples.
func computeRMS(samples []float32) float64 {
	if len(samples) == 0 {
		return 0
	}
	var sum float64
	for _, s := range samples {
		sum += float64(s) * float64(s)
	}
	return math.Sqrt(sum / float64(len(samples)))
}

// TrimSilence trims leading and trailing silence from a buffer of PCM samples.
// Returns the trimmed slice. Silence is defined as samples below the threshold.
func TrimSilence(samples []float32, threshold float64) []float32 {
	if len(samples) == 0 {
		return samples
	}

	start := 0
	for start < len(samples) && math.Abs(float64(samples[start])) < threshold {
		start++
	}

	end := len(samples)
	for end > start && math.Abs(float64(samples[end-1])) < threshold {
		end--
	}

	if start >= end {
		return nil
	}
	return samples[start:end]
}
