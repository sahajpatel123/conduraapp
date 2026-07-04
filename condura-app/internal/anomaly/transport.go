package anomaly

import "net/http"

// RecordingTransport wraps an http.RoundTripper and calls
// Detector.RecordNetwork for every outbound request. It is the
// production wire for the CLAUDE.md §5.6 fifth "agent went insane"
// trigger ("agent sends to network endpoints it has never used
// before"): every LLM provider, OAuth flow, reach subsystem, and
// IPC HTTP egress that uses the standard library will pass through
// it via the daemon's wrapProviderHTTPClient wiring.
//
// The detector may be nil, in which case the transport is a
// pass-through (no behavior change). The inner transport is also
// called when nil, falling back to http.DefaultTransport so the
// daemon can construct clients without a custom Transport set.
//
// Construction is wrapped via NewRecordingTransport rather than
// direct struct literal so the nil-vs-empty handling is consistent
// with the InProcessGuard pattern in internal/halt/network.go.
type RecordingTransport struct {
	Detector *Detector
	Inner    http.RoundTripper
}

// NewRecordingTransport constructs a RecordingTransport that calls
// det.RecordNetwork before each request and delegates to inner.
// Either may be nil; nil Detector → no recording, nil Inner →
// http.DefaultTransport.
func NewRecordingTransport(det *Detector, inner http.RoundTripper) *RecordingTransport {
	return &RecordingTransport{Detector: det, Inner: inner}
}

// RoundTrip implements http.RoundTripper. It records the host then
// delegates. Recording is best-effort and never fails the request —
// the trip callback runs in the detector's goroutine and never
// blocks the caller.
func (t *RecordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if t.Detector != nil && req != nil && req.URL != nil {
		t.Detector.RecordNetwork(req.URL.Hostname())
	}
	inner := t.Inner
	if inner == nil {
		inner = http.DefaultTransport
	}
	return inner.RoundTrip(req)
}
