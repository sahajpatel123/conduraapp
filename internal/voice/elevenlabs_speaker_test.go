package voice

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestElevenLabsSpeaker_Speak(t *testing.T) {
	// Create a mock ElevenLabs API server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path.
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if !strings.HasPrefix(r.URL.Path, "/v1/text-to-speech/") {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify API key header.
		apiKey := r.Header.Get("xi-api-key")
		if apiKey != "test-api-key" {
			t.Errorf("unexpected API key: %s", apiKey)
		}

		// Parse request body.
		var reqBody elevenLabsTTSRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("failed to decode request body: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Verify request fields.
		if reqBody.ModelID != "eleven_monolingual_v1" {
			t.Errorf("expected model 'eleven_monolingual_v1', got '%s'", reqBody.ModelID)
		}
		if reqBody.Text != "Hello, this is a test." {
			t.Errorf("unexpected text: %s", reqBody.Text)
		}

		// Return mock audio data.
		w.Header().Set("Content-Type", "audio/mpeg")
		_, _ = w.Write([]byte("fake audio data"))
	}))
	defer server.Close()

	// Create speaker with custom HTTP client.
	client := &http.Client{
		Transport: &roundTripperFunc{
			fn: func(req *http.Request) (*http.Response, error) {
				req.URL.Scheme = "http"
				req.URL.Host = server.Listener.Addr().String()
				return http.DefaultTransport.RoundTrip(req)
			},
		},
	}

	speaker := NewElevenLabsSpeaker("test-api-key", "test-voice-id", "eleven_monolingual_v1",
		WithElevenLabsHTTPClient(client))

	// Test speaking (this will fail on audio playback, but we can test the API call).
	ctx := context.Background()
	err := speaker.Speak(ctx, "Hello, this is a test.")
	// We expect an error because afplay/aplay won't be available in test,
	// but the API call should succeed.
	if err != nil && strings.Contains(err.Error(), "elevenlabs speaker:") {
		t.Errorf("unexpected API error: %v", err)
	}
}

func TestElevenLabsSpeaker_Speak_NoAPIKey(t *testing.T) {
	speaker := NewElevenLabsSpeaker("", "test-voice-id", "eleven_monolingual_v1")

	ctx := context.Background()
	err := speaker.Speak(ctx, "Hello")
	if err == nil {
		t.Error("expected error for missing API key")
	}
}

func TestElevenLabsSpeaker_Speak_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"detail": {"message": "Invalid API key"}}`)
	}))
	defer server.Close()

	client := &http.Client{
		Transport: &roundTripperFunc{
			fn: func(req *http.Request) (*http.Response, error) {
				req.URL.Scheme = "http"
				req.URL.Host = server.Listener.Addr().String()
				return http.DefaultTransport.RoundTrip(req)
			},
		},
	}

	speaker := NewElevenLabsSpeaker("invalid-key", "test-voice-id", "eleven_monolingual_v1",
		WithElevenLabsHTTPClient(client))

	ctx := context.Background()
	err := speaker.Speak(ctx, "Hello")
	if err == nil {
		t.Error("expected error for API failure")
	}
}

func TestElevenLabsSpeaker_Stop(t *testing.T) {
	speaker := NewElevenLabsSpeaker("test-key", "test-voice-id", "eleven_monolingual_v1")

	// Stop should not panic even if no Speak is in progress.
	speaker.Stop()
}
