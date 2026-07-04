package voice

import "errors"

var (
	// ErrNoMic is returned when no microphone is available.
	ErrNoMic = errors.New("voice: no microphone available")
	// ErrNoTTSEngine is returned when no TTS engine is found.
	ErrNoTTSEngine = errors.New("voice: no TTS engine found")
	// ErrModelDownloadFailed is returned when model download fails.
	ErrModelDownloadFailed = errors.New("voice: model download failed")
	// ErrWhisperBinaryMissing is returned when whisper-cli is not found.
	ErrWhisperBinaryMissing = errors.New("voice: whisper binary not found")
	// ErrMicPermissionDenied is returned when mic permission is denied.
	ErrMicPermissionDenied = errors.New("voice: microphone permission denied")
	// ErrCaptureTimeout is returned when capture exceeds max duration.
	ErrCaptureTimeout = errors.New("voice: capture timeout")
)
