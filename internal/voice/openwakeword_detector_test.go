package voice

import (
	"context"
	"testing"
)

func TestOpenWakeWordDetector_Status(t *testing.T) {
	detector := NewOpenWakeWordDetector("/nonexistent/binary", "/nonexistent/model")

	// Status should show available and model loaded.
	status := detector.Status()
	if !status.Available {
		t.Error("expected available to be true")
	}
	if !status.ModelLoaded {
		t.Error("expected model loaded to be true")
	}
	if status.Listening {
		t.Error("expected listening to be false")
	}
}

func TestOpenWakeWordDetector_Stop(t *testing.T) {
	detector := NewOpenWakeWordDetector("/nonexistent/binary", "/nonexistent/model")

	// Stop should not panic even if not started.
	err := detector.Stop()
	if err != nil {
		t.Errorf("unexpected error: %v", err)
	}
}

func TestOpenWakeWordDetector_Start_MissingBinary(t *testing.T) {
	detector := NewOpenWakeWordDetector("/nonexistent/binary", "/nonexistent/model")

	ctx := context.Background()

	_, err := detector.Start(ctx)
	if err == nil {
		t.Error("expected error for missing binary")
	}
}

func TestOpenWakeWordDetector_Threshold(t *testing.T) {
	// Test that threshold option works.
	detector := NewOpenWakeWordDetector("/nonexistent/binary", "/nonexistent/model",
		WithThreshold(0.8))

	// We can't easily test the internal threshold value, but we can verify
	// that the constructor doesn't panic with the option.
	if detector == nil {
		t.Error("expected non-nil detector")
	}
}

func TestOpenWakeWordDetector_Start_Cancel(t *testing.T) {
	// Test that context cancellation stops the detector.
	detector := NewOpenWakeWordDetector("/nonexistent/binary", "/nonexistent/model")

	ctx, cancel := context.WithCancel(context.Background())

	// This will fail because binary doesn't exist, but we just cancel and check status.
	_, _ = detector.Start(ctx)
	cancel()

	// Status should show not listening after cancel (or not listening if start failed).
	status := detector.Status()
	if status.Listening {
		t.Error("expected listening to be false after cancel")
	}
}
