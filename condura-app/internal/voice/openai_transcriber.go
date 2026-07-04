package voice

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"mime/multipart"
	"net/http"
)

// OpenAITranscriber implements Transcriber using the OpenAI Whisper API.
// It sends audio data to POST https://api.openai.com/v1/audio/transcriptions
// and returns the transcribed text.
type OpenAITranscriber struct {
	apiKey     string
	model      string
	httpClient *http.Client
}

// OpenAITranscriberOption configures an OpenAITranscriber.
type OpenAITranscriberOption func(*OpenAITranscriber)

// WithHTTPClient sets a custom HTTP client.
func WithHTTPClient(c *http.Client) OpenAITranscriberOption {
	return func(t *OpenAITranscriber) {
		t.httpClient = c
	}
}

// NewOpenAITranscriber creates a new OpenAI-based Transcriber.
// The apiKey is the OpenAI API key. The model is typically "whisper-1".
func NewOpenAITranscriber(apiKey, model string, opts ...OpenAITranscriberOption) Transcriber {
	t := &OpenAITranscriber{
		apiKey:     apiKey,
		model:      model,
		httpClient: http.DefaultClient,
	}
	for _, opt := range opts {
		opt(t)
	}
	if t.model == "" {
		t.model = "whisper-1"
	}
	return t
}

// Transcribe performs a one-shot transcription of a complete audio clip.
func (t *OpenAITranscriber) Transcribe(ctx context.Context, audio []byte) (Transcript, error) {
	if t.apiKey == "" {
		return Transcript{}, fmt.Errorf("openai transcriber: API key required")
	}

	// Build multipart form.
	body := &bytes.Buffer{}
	writer := multipart.NewWriter(body)

	// Add audio file.
	part, err := writer.CreateFormFile("file", "audio.wav")
	if err != nil {
		return Transcript{}, fmt.Errorf("openai transcriber: create form file: %w", err)
	}
	if _, err := part.Write(audio); err != nil {
		return Transcript{}, fmt.Errorf("openai transcriber: write audio: %w", err)
	}

	// Add model field.
	if err := writer.WriteField("model", t.model); err != nil {
		return Transcript{}, fmt.Errorf("openai transcriber: write model: %w", err)
	}

	if err := writer.Close(); err != nil {
		return Transcript{}, fmt.Errorf("openai transcriber: close writer: %w", err)
	}

	// Create request.
	req, err := http.NewRequestWithContext(ctx, http.MethodPost,
		"https://api.openai.com/v1/audio/transcriptions", body)
	if err != nil {
		return Transcript{}, fmt.Errorf("openai transcriber: create request: %w", err)
	}

	req.Header.Set("Content-Type", writer.FormDataContentType())
	req.Header.Set("Authorization", "Bearer "+t.apiKey)

	// Send request.
	resp, err := t.httpClient.Do(req)
	if err != nil {
		return Transcript{}, fmt.Errorf("openai transcriber: request failed: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return Transcript{}, fmt.Errorf("openai transcriber: API error (status %d): %s", resp.StatusCode, string(respBody))
	}

	// Parse response.
	var result struct {
		Text string `json:"text"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return Transcript{}, fmt.Errorf("openai transcriber: decode response: %w", err)
	}

	return Transcript{
		Text: result.Text,
	}, nil
}

// TranscribeStream processes live audio samples and emits partial results.
func (t *OpenAITranscriber) TranscribeStream(ctx context.Context, audio <-chan []float32) (<-chan Partial, error) {
	// Read all audio data from channel and transcribe.
	// Streaming is not supported by the Whisper API, so we collect all samples
	// and transcribe at once when the channel is closed.
	ch := make(chan Partial, 1)

	go func() {
		defer close(ch)

		var samples []float32
		for data := range audio {
			samples = append(samples, data...)
		}

		// Convert float32 samples to WAV bytes.
		// For now, we'll just send an empty result since we need proper WAV encoding.
		// This will be implemented when we have the full audio pipeline.
		_ = samples
		ch <- Partial{
			Text:    "",
			IsFinal: true,
		}
	}()

	return ch, nil
}
