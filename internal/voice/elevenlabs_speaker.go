package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
)

// ElevenLabsSpeaker implements Speaker using the ElevenLabs TTS API.
// It sends text to POST https://api.elevenlabs.io/v1/text-to-speech/{voice_id}
// and plays the returned audio bytes.
type ElevenLabsSpeaker struct {
	apiKey     string
	voiceID    string
	modelID    string
	httpClient *http.Client
	mu         sync.Mutex
	cancel     context.CancelFunc
}

// ElevenLabsSpeakerOption configures an ElevenLabsSpeaker.
type ElevenLabsSpeakerOption func(*ElevenLabsSpeaker)

// WithElevenLabsHTTPClient sets a custom HTTP client.
func WithElevenLabsHTTPClient(c *http.Client) ElevenLabsSpeakerOption {
	return func(s *ElevenLabsSpeaker) {
		s.httpClient = c
	}
}

// NewElevenLabsSpeaker creates a new ElevenLabs-based Speaker.
// The apiKey is the ElevenLabs API key. The voiceID is the voice to use.
// The modelID is typically "eleven_monolingual_v1" or "eleven_multilingual_v1".
func NewElevenLabsSpeaker(apiKey, voiceID, modelID string, opts ...ElevenLabsSpeakerOption) Speaker {
	s := &ElevenLabsSpeaker{
		apiKey:     apiKey,
		voiceID:    voiceID,
		modelID:    modelID,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(s)
	}
	if s.modelID == "" {
		s.modelID = "eleven_monolingual_v1"
	}
	return s
}

// elevenLabsTTSRequest is the request body for the ElevenLabs TTS API.
type elevenLabsTTSRequest struct {
	Text    string `json:"text"`
	ModelID string `json:"model_id"`
}

// Speak converts text to speech using the ElevenLabs TTS API and plays it.
func (s *ElevenLabsSpeaker) Speak(ctx context.Context, text string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.apiKey == "" {
		return fmt.Errorf("elevenlabs speaker: API key required")
	}

	// Build request body.
	reqBody := elevenLabsTTSRequest{
		Text:    text,
		ModelID: s.modelID,
	}

	jsonBody, err := json.Marshal(reqBody)
	if err != nil {
		return fmt.Errorf("elevenlabs speaker: marshal request: %w", err)
	}

	// Create HTTP request.
	url := fmt.Sprintf("https://api.elevenlabs.io/v1/text-to-speech/%s", s.voiceID)
	req, err := http.NewRequestWithContext(ctx, http.MethodPost, url, bytes.NewReader(jsonBody))
	if err != nil {
		return fmt.Errorf("elevenlabs speaker: create request: %w", err)
	}

	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("xi-api-key", s.apiKey)

	// Send request.
	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("elevenlabs speaker: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("elevenlabs speaker: API error (status %d): %s", resp.StatusCode, string(body))
	}

	// Read audio bytes.
	audioData, err := io.ReadAll(resp.Body)
	if err != nil {
		return fmt.Errorf("elevenlabs speaker: read response: %w", err)
	}

	// Play audio using platform-specific method.
	return playAudio(ctx, audioData)
}

// Stop halts any in-progress speech.
func (s *ElevenLabsSpeaker) Stop() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.cancel != nil {
		s.cancel()
	}
}
