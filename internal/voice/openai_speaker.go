package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os/exec"
	"runtime"
	"sync"
)

// OpenAISpeaker implements Speaker using the OpenAI TTS API.
// It sends text to POST https://api.openai.com/v1/audio/speech
// and plays the returned audio bytes.
type OpenAISpeaker struct {
	apiKey     string
	model      string
	voice      string
	httpClient *http.Client
	mu         sync.Mutex
	cancel     context.CancelFunc
}

// OpenAISpeakerOption configures an OpenAISpeaker.
type OpenAISpeakerOption func(*OpenAISpeaker)

// WithOpenAIHTTPClient sets a custom HTTP client.
func WithOpenAIHTTPClient(c *http.Client) OpenAISpeakerOption {
	return func(s *OpenAISpeaker) {
		s.httpClient = c
	}
}

// NewOpenAISpeaker creates a new OpenAI-based Speaker.
// The apiKey is the OpenAI API key. The model is typically "tts-1" or "tts-1-hd".
// The voice is one of "alloy", "echo", "fable", "onyx", "nova", "shimmer".
func NewOpenAISpeaker(apiKey, model, voice string, opts ...OpenAISpeakerOption) Speaker {
	s := &OpenAISpeaker{
		apiKey:     apiKey,
		model:      model,
		voice:      voice,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(s)
	}
	if s.model == "" {
		s.model = "tts-1"
	}
	if s.voice == "" {
		s.voice = "alloy"
	}
	return s
}

// openAITTSRequest is the request body for the OpenAI TTS API.
type openAITTSRequest struct {
	Input string `json:"input"`
	Model string `json:"model"`
	Voice string `json:"voice"`
}

// Speak converts text to speech using the OpenAI TTS API and plays it.
func (s *OpenAISpeaker) Speak(ctx context.Context, text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.apiKey == "" {
		return fmt.Errorf("openai speaker: API key required")
	}

	// Build request body.
	reqBody := openAITTSRequest{
		Input: text,
		Model: s.model,
		Voice: s.voice,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("openai speaker: marshal request: %w", err)
	}

	// Create HTTP request.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.openai.com/v1/audio/speech", bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("openai speaker: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Authorization", "Bearer "+s.apiKey)

	// Send request.
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("openai speaker: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("openai speaker: API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Read audio bytes.
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("openai speaker: read response: %w", err)
	}

	// Play audio using platform-specific method.
	return playAudio(ctx, audioData)
}

// Stop halts any in-progress speech.
func (s *OpenAISpeaker) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
	}
}

// playAudio plays audio data using the platform's native audio player.
func playAudio(ctx context.Context, data []byte) error {
	// Write data to a temporary file.
	tmpFile, err := createTempWAV(data)
	if err != nil {
		return fmt.Errorf("play audio: create temp file: %w", err)
	}
	defer cleanupTempFile(tmpFile)

	// Play based on platform.
	switch runtime.GOOS {
	case "darwin":
		return playAudioDarwin(ctx, tmpFile)
	case "linux":
		return playAudioLinux(ctx, tmpFile)
	case "windows":
		return playAudioWindows(ctx, tmpFile)
	default:
		return fmt.Errorf("play audio: unsupported platform: %s", runtime.GOOS)
	}
}

// playAudioDarwin plays audio on macOS using afplay.
//
//nolint:gosec // path is from temp file, not user input
func playAudioDarwin(ctx context.Context, path string) error {
	cmd := exec.CommandContext(ctx, "afplay", path)
	return cmd.Run()
}

// playAudioLinux plays audio on Linux using aplay or paplay.
//
//nolint:gosec // path is from temp file, not user input
func playAudioLinux(ctx context.Context, path string) error {
	// Try paplay first (PulseAudio), then aplay.
	cmd := exec.CommandContext(ctx, "paplay", path)
	if err := cmd.Run(); err != nil {
		cmd = exec.CommandContext(ctx, "aplay", path)
		return cmd.Run()
	}
	return nil
}

// playAudioWindows plays audio on Windows using PowerShell.
//
//nolint:gosec // path is from temp file, not user input
func playAudioWindows(ctx context.Context, path string) error {
	cmd := exec.CommandContext(ctx, "powershell", "-c",
		fmt.Sprintf("(New-Object Media.SoundPlayer '%s').PlaySync()", path))
	return cmd.Run()
}
