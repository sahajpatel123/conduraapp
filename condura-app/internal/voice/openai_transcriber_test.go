package voice

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

func TestOpenAITranscriber_Transcribe(t *testing.T) {
	// Create a mock OpenAI API server.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path.
		if r.Method != http.MethodPost {
			t.Errorf("expected POST, got %s", r.Method)
		}
		if r.URL.Path != "/v1/audio/transcriptions" {
			t.Errorf("unexpected path: %s", r.URL.Path)
		}

		// Verify API key header.
		auth := r.Header.Get("Authorization")
		if auth != "Bearer test-api-key" {
			t.Errorf("unexpected auth header: %s", auth)
		}

		// Parse multipart form.
		if err := r.ParseMultipartForm(10 << 20); err != nil {
			t.Errorf("failed to parse form: %v", err)
			http.Error(w, "bad request", http.StatusBadRequest)
			return
		}

		// Verify model field.
		model := r.FormValue("model")
		if model != "whisper-1" {
			t.Errorf("expected model 'whisper-1', got '%s'", model)
		}

		// Return mock response.
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"text": "Hello, this is a test transcription.",
		})
	}))
	defer server.Close()

	// Create transcriber with custom HTTP client.
	client := &http.Client{
		Transport: &roundTripperFunc{
			fn: func(req *http.Request) (*http.Response, error) {
				req.URL.Scheme = "http"
				req.URL.Host = server.Listener.Addr().String()
				return http.DefaultTransport.RoundTrip(req)
			},
		},
	}

	transcriber := NewOpenAITranscriber("test-api-key", "whisper-1",
		WithHTTPClient(client))

	// Test transcription.
	ctx := context.Background()
	result, err := transcriber.Transcribe(ctx, []byte("fake audio data"))
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if result.Text != "Hello, this is a test transcription." {
		t.Errorf("unexpected text: %s", result.Text)
	}
}

func TestOpenAITranscriber_Transcribe_NoAPIKey(t *testing.T) {
	transcriber := NewOpenAITranscriber("", "whisper-1")

	ctx := context.Background()
	_, err := transcriber.Transcribe(ctx, []byte("audio"))
	if err == nil {
		t.Error("expected error for missing API key")
	}
}

func TestOpenAITranscriber_Transcribe_APIError(t *testing.T) {
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

	transcriber := NewOpenAITranscriber("invalid-key", "whisper-1",
		WithHTTPClient(client))

	ctx := context.Background()
	_, err := transcriber.Transcribe(ctx, []byte("audio"))
	if err == nil {
		t.Error("expected error for API failure")
	}
}

func TestOpenAITranscriber_TranscribeStream(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(map[string]string{
			"text": "Streamed transcription.",
		})
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

	transcriber := NewOpenAITranscriber("test-api-key", "whisper-1",
		WithHTTPClient(client))

	ctx := context.Background()
	audio := make(chan []float32, 1)
	close(audio)

	ch, err := transcriber.TranscribeStream(ctx, audio)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// Read from channel.
	partial := <-ch
	if !partial.IsFinal {
		t.Error("expected final partial")
	}
}
