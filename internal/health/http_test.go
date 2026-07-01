package health

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
)

// TestLivez_AlwaysOK verifies the livez probe returns 200 with the
// expected body and never invokes the supplied livez func.
func TestLivez_AlwaysOK(t *testing.T) {
	called := 0
	livez := func() error { called++; return nil }
	readyz := func() error { return nil }
	h := HTTPHandler(livez, readyz)

	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("livez status = %d, want %d", rr.Code, http.StatusOK)
	}
	body, _ := io.ReadAll(rr.Body)
	if got := string(body); got != "alive\n" {
		t.Errorf("livez body = %q, want %q", got, "alive\n")
	}
	if called != 0 {
		t.Errorf("livez func should not be invoked by /livez, was called %d times", called)
	}
	if cc := rr.Header().Get("Cache-Control"); cc != "no-store" {
		t.Errorf("livez Cache-Control = %q, want no-store", cc)
	}
	if ct := rr.Header().Get("Content-Type"); ct == "" {
		t.Error("livez should set Content-Type")
	}
}

// TestLivez_NoAuthRequired verifies /livez returns 200 even with
// no Authorization header and no credentials at all. This is the
// orchestrator probe contract: it must be reachable with curl +
// bare GET, no auth.
func TestLivez_NoAuthRequired(t *testing.T) {
	h := HTTPHandler(nil, func() error { return nil })
	req := httptest.NewRequest(http.MethodGet, "/livez", nil)
	// No Authorization header on purpose.
	if req.Header.Get("Authorization") != "" {
		t.Fatal("test bug: Authorization header leaked into request")
	}
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusOK {
		t.Fatalf("livez without auth: status = %d, want 200", rr.Code)
	}
}

// TestReadyz_OK returns 200 when the supplied func returns nil.
func TestReadyz_OK(t *testing.T) {
	h := HTTPHandler(nil, func() error { return nil })
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusOK {
		t.Fatalf("readyz status = %d, want %d", rr.Code, http.StatusOK)
	}
	body, _ := io.ReadAll(rr.Body)
	if got := string(body); got != "ready\n" {
		t.Errorf("readyz body = %q, want %q", got, "ready\n")
	}
}

// TestReadyz_Down returns 503 with the reason when the func fails.
func TestReadyz_Down(t *testing.T) {
	h := HTTPHandler(nil, func() error { return errors.New("storage closed") })
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("readyz status = %d, want %d", rr.Code, http.StatusServiceUnavailable)
	}
	body, _ := io.ReadAll(rr.Body)
	want := "not ready: storage closed\n"
	if got := string(body); got != want {
		t.Errorf("readyz body = %q, want %q", got, want)
	}
}

// TestReadyz_TruncatesLongReasons guards against a runaway readyz
// func flooding the response. 256 bytes is the documented cap.
func TestReadyz_TruncatesLongReasons(t *testing.T) {
	long := make([]byte, 4096)
	for i := range long {
		long[i] = 'x'
	}
	h := HTTPHandler(nil, func() error { return errors.New(string(long)) })
	req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)

	if rr.Code != http.StatusServiceUnavailable {
		t.Fatalf("status = %d, want 503", rr.Code)
	}
	body, _ := io.ReadAll(rr.Body)
	// "not ready: " (11) + 256 chars + "\n" (1) = 268 bytes
	if len(body) > 268 {
		t.Errorf("readyz body length = %d, expected <= 268 (truncation missing)", len(body))
	}
}

// TestReadyz_InvokesFuncOnEveryCall verifies the func is called
// per request (so the orchestrator always sees fresh state).
func TestReadyz_InvokesFuncOnEveryCall(t *testing.T) {
	calls := 0
	h := HTTPHandler(nil, func() error {
		calls++
		return nil
	})
	for i := 0; i < 3; i++ {
		req := httptest.NewRequest(http.MethodGet, "/readyz", nil)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusOK {
			t.Fatalf("iter %d: status = %d, want 200", i, rr.Code)
		}
	}
	if calls != 3 {
		t.Errorf("readyz func called %d times, want 3", calls)
	}
}

// TestMethodNotAllowed_405 returns 405 (not 200 or 500) for
// non-GET/HEAD methods on /livez and /readyz. Orchestrators that
// accidentally send POST should get a clear error.
func TestMethodNotAllowed_405(t *testing.T) {
	h := HTTPHandler(nil, func() error { return nil })
	for _, path := range []string{"/livez", "/readyz"} {
		req := httptest.NewRequest(http.MethodPost, path, nil)
		rr := httptest.NewRecorder()
		h.ServeHTTP(rr, req)
		if rr.Code != http.StatusMethodNotAllowed {
			t.Errorf("%s POST status = %d, want 405", path, rr.Code)
		}
		if allow := rr.Header().Get("Allow"); allow != "GET, HEAD" {
			t.Errorf("%s Allow = %q, want GET, HEAD", path, allow)
		}
	}
}

// TestHEAD_Supported verifies HEAD returns the same status as GET
// (but the body is dropped by net/http automatically).
func TestHEAD_Supported(t *testing.T) {
	h := HTTPHandler(nil, func() error { return errors.New("nope") })
	req := httptest.NewRequest(http.MethodHead, "/readyz", nil)
	rr := httptest.NewRecorder()
	h.ServeHTTP(rr, req)
	if rr.Code != http.StatusServiceUnavailable {
		t.Errorf("HEAD readyz status = %d, want 503", rr.Code)
	}
}
