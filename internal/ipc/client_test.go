// Tests for the ipc.Client. They spin up a ServerTransport in-process
// (no socket binding) and call the client over a real loopback HTTP
// listener so the wire format is exercised.

package ipc_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/synapticapp/synaptic/internal/ipc"
)

func TestClientCall(t *testing.T) {
	srv := ipc.NewServer()
	srv.Register("ping", func(_ context.Context, _ json.RawMessage) (any, error) {
		return map[string]any{"pong": true}, nil
	})
	srv.Register("echo", func(_ context.Context, params json.RawMessage) (any, error) {
		return params, nil
	})
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Reuse the server's HandleRaw over the real HTTP request.
		body := make([]byte, 0, 256)
		buf := make([]byte, 256)
		for {
			n, err := r.Body.Read(buf)
			if n > 0 {
				body = append(body, buf[:n]...)
			}
			if err != nil {
				break
			}
		}
		out, err := srv.HandleRaw(r.Context(), body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if out == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(out)
	}))
	defer ts.Close()

	addr := strings.TrimPrefix(ts.URL, "http://")
	c, err := ipc.Dial("tcp://"+addr, "")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = c.Close() }()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	var out map[string]any
	if err := c.Call(ctx, "ping", nil, &out); err != nil {
		t.Fatalf("ping: %v", err)
	}
	if pong, _ := out["pong"].(bool); !pong {
		t.Fatalf("expected pong=true, got %v", out)
	}

	// Echo back a struct.
	var echoOut map[string]any
	if err := c.Call(ctx, "echo", map[string]any{"hello": "world"}, &echoOut); err != nil {
		t.Fatalf("echo: %v", err)
	}
	if got, _ := echoOut["hello"].(string); got != "world" {
		t.Fatalf("expected hello=world, got %v", echoOut)
	}
}

func TestClientUnknownMethod(t *testing.T) {
	srv := ipc.NewServer()
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body := make([]byte, 0, 256)
		buf := make([]byte, 256)
		for {
			n, err := r.Body.Read(buf)
			if n > 0 {
				body = append(body, buf[:n]...)
			}
			if err != nil {
				break
			}
		}
		out, err := srv.HandleRaw(r.Context(), body)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		if out == nil {
			w.WriteHeader(http.StatusNoContent)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write(out)
	}))
	defer ts.Close()

	c, err := ipc.Dial("tcp://"+strings.TrimPrefix(ts.URL, "http://"), "")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = c.Close() }()

	var out any
	err = c.Call(context.Background(), "nope", nil, &out)
	if err == nil {
		t.Fatalf("expected error, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Fatalf("expected method-not-found error, got %v", err)
	}
}

func TestClientConnRefused(t *testing.T) {
	// 127.0.0.1:1 is almost certainly not bound; should be refused.
	c, err := ipc.Dial("tcp://127.0.0.1:1", "")
	if err != nil {
		t.Fatalf("dial: %v", err)
	}
	defer func() { _ = c.Close() }()
	err = c.Call(context.Background(), "ping", nil, nil)
	if err == nil {
		t.Fatalf("expected connection refused")
	}
	if !ipc.IsConnRefused(err) {
		t.Fatalf("IsConnRefused = false for %v", err)
	}
}
