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

// roundTripperFunc is an http.RoundTripper that calls a function.
type roundTripperFunc struct {
	fn func(*http.Request) (*http.Response, error)
}

func (f *roundTripperFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f.fn(req)
}

func TestOpenAISpeaker_Speak(t *testing.T) {
	// Create a mock OpenAI API server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path.
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/audio/speech" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify API key header.
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-api-key" {
			t.Errorf("unexpected auth header: %s", auth)
		}

		// Parse request body.
		var reqBody openAITTSRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("failed to decode request body: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Verify request fields.
		if reqBody.Model != "tts-1" {
			t.Errorf("expected model 'tts-1', got '%s'", reqBody.Model)
		}
		if reqBody.Voice != "alloy" {
			t.Errorf("expected voice 'alloy', got '%s'", reqBody.Voice)
		}
		if reqBody.Input != "Hello, this is a test." {
			t.Errorf("unexpected input: %s", reqBody.Input)
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

	speaker := NewOpenAISpeaker("test-api-key", "tts-1", "alloy",
		WithOpenAIHTTPClient(client))

	// Test speaking (this will fail on audio playback, but we can test the API call).
	ctx := context.Background()
	err := speaker.Speak(ctx, "Hello, this is a test.")
	// We expect an error because afplay/aplay won't be available in test,
	// but the API call should succeed.
	if err != nil && strings.Contains(err.Error(), "openai speaker:") {
		t.Errorf("unexpected API error: %v", err)
	}
}

func TestOpenAISpeaker_Speak_NoAPIKey(t *testing.T) {
	speaker := NewOpenAISpeaker("", "tts-1", "alloy")

	ctx := context.Background()
	err := speaker.Speak(ctx, "Hello")
	if err == nil {
		t.Error("expected error for missing API key")
	}
}

func TestOpenAISpeaker_Speak_APIError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = io.WriteString(w, `{"error": {"message": "Invalid API key"}}`)
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

	speaker := NewOpenAISpeaker("invalid-key", "tts-1", "alloy",
		WithOpenAIHTTPClient(client))

	ctx := context.Background()
	err := speaker.Speak(ctx, "Hello")
	if err == nil {
		t.Error("expected error for API failure")
	}
}

func TestOpenAISpeaker_Stop(t *testing.T) {
	speaker := NewOpenAISpeaker("test-key", "tts-1", "alloy")

	// Stop should not panic even if no Speak is in progress.
	speaker.Stop()
}
