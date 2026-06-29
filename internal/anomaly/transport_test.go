package anomaly

import (
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"
)

// TestRecordingTransport_DelegatesAndRecords pins that every
// outbound request triggers RecordNetwork with the hostname
// (port-stripped, lowercased) AND delegates to the inner transport.
// This is the regression test for P0-2 of the 2026-06-29 audit:
// the fifth §5.6 trigger was dead because RecordNetwork had no
// production call sites.
func TestRecordingTransport_DelegatesAndRecords(t *testing.T) {
	var trips atomic.Int32
	det := NewDetector(func(_ Trip) {
		trips.Add(1)
	})
	defer det.Close()

	var innerCalls atomic.Int32
	inner := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		innerCalls.Add(1)
		return &http.Response{
			StatusCode: 200,
			Body:       http.NoBody,
			Request:    req,
		}, nil
	})

	rt := NewRecordingTransport(det, inner)

	// First call: api.anthropic.com — populates seenHosts, no trip.
	req1 := httptest.NewRequest("GET", "https://api.anthropic.com/v1/messages", nil)
	resp1, err := rt.RoundTrip(req1)
	if err != nil {
		t.Fatalf("first roundtrip: %v", err)
	}
	defer func() { _ = resp1.Body.Close() }()

	// Second call: same host — no trip.
	req2 := httptest.NewRequest("GET", "https://api.anthropic.com/v1/messages", nil)
	resp2, err := rt.RoundTrip(req2)
	if err != nil {
		t.Fatalf("second roundtrip: %v", err)
	}
	defer func() { _ = resp2.Body.Close() }()

	// Wait briefly for the async trip path. None should fire yet.
	time.Sleep(50 * time.Millisecond)
	if got := trips.Load(); got != 0 {
		t.Fatalf("expected 0 trips before new host, got %d", got)
	}

	// Third call: NEW host — trips the new-endpoint detector.
	req3 := httptest.NewRequest("GET", "https://api.openai.com/v1/chat/completions", nil)
	resp3, err := rt.RoundTrip(req3)
	if err != nil {
		t.Fatalf("third roundtrip: %v", err)
	}
	defer func() { _ = resp3.Body.Close() }()
	time.Sleep(50 * time.Millisecond)
	if got := trips.Load(); got != 1 {
		t.Fatalf("expected 1 trip after new host, got %d", got)
	}

	// Inner transport was called for all three requests.
	if got := innerCalls.Load(); got != 3 {
		t.Fatalf("expected inner to be called 3 times, got %d", got)
	}
}

// TestRecordingTransport_NilDetectorIsPassThrough pins that a nil
// Detector means no recording but the inner transport still runs.
// This is the safety: a daemon wired without an anomaly detector
// still produces correct HTTP behavior.
func TestRecordingTransport_NilDetectorIsPassThrough(t *testing.T) {
	var innerCalls atomic.Int32
	inner := roundTripperFunc(func(req *http.Request) (*http.Response, error) {
		innerCalls.Add(1)
		return &http.Response{StatusCode: 200, Body: http.NoBody, Request: req}, nil
	})

	rt := NewRecordingTransport(nil, inner)

	req := httptest.NewRequest("GET", "https://example.com", nil)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("roundtrip: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	if got := innerCalls.Load(); got != 1 {
		t.Fatalf("expected inner to be called once, got %d", got)
	}
}

// TestRecordingTransport_NilInnerFallsBackToDefault pins that a
// nil inner transport falls back to http.DefaultTransport. The
// daemon constructs providers that may have nil Transport fields;
// the wrapper must not panic on them.
func TestRecordingTransport_NilInnerFallsBackToDefault(t *testing.T) {
	det := NewDetector(func(_ Trip) {})
	defer det.Close()

	rt := NewRecordingTransport(det, nil)
	req := httptest.NewRequest("GET", "http://127.0.0.1:1/never-reaches-server", nil)
	// The default transport will fail the dial — we just need to
	// confirm we did NOT panic on nil Inner.
	respNID, err := rt.RoundTrip(req)
	if err == nil {
		t.Fatal("expected dial error against 127.0.0.1:1, got nil")
	}
	if respNID != nil && respNID.Body != nil {
		_ = respNID.Body.Close()
	}
}

// roundTripperFunc lets us write a RoundTripper as a function in tests.
type roundTripperFunc func(*http.Request) (*http.Response, error)

func (f roundTripperFunc) RoundTrip(r *http.Request) (*http.Response, error) {
	return f(r)
}
