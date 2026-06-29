package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/ipc"
	"github.com/sahajpatel123/conduraapp/internal/overlay"
	"github.com/sahajpatel123/conduraapp/internal/sse"
)

// TestPhase6_VoiceStatusUnavailable verifies voice.status works
// when voice is not configured (returns available=false).
func TestPhase6_VoiceStatusUnavailable(t *testing.T) {
	subs := &Subsystems{
		Voice:   nil,
		Broker:  sse.NewBroker(),
		Audit:   nil,
		Overlay: overlay.NewNoopController(),
	}
	srv := ipc.NewServer()
	registerPhase6Methods(srv, subs)

	resp, err := phase6CallRPC(t, srv, "voice.status", nil)
	if err != nil {
		t.Fatalf("voice.status: %v", err)
	}
	m, ok := resp.(map[string]any)
	if !ok {
		t.Fatalf("expected map, got %T", resp)
	}
	if avail, _ := m["available"].(bool); avail {
		t.Error("expected available=false when voice is nil")
	}
	if state, _ := m["status"].(string); state != "idle" {
		t.Errorf("expected status=idle, got %q", state)
	}
}

// TestPhase6_PresenceSummonAndDismiss verifies presence.summon
// and presence.dismiss route to the real overlay.
func TestPhase6_PresenceSummonAndDismiss(t *testing.T) {
	ctrl := overlay.NewNoopController()
	subs := &Subsystems{
		Voice:   nil,
		Broker:  sse.NewBroker(),
		Audit:   nil,
		Overlay: ctrl,
	}
	srv := ipc.NewServer()
	registerPhase6Methods(srv, subs)

	// Initially hidden.
	if got := ctrl.State(); got != overlay.StateHidden {
		t.Fatalf("initial state = %v, want hidden", got)
	}

	// Summon.
	if _, err := phase6CallRPC(t, srv, "presence.summon", nil); err != nil {
		t.Fatalf("presence.summon: %v", err)
	}
	if got := ctrl.State(); got == overlay.StateHidden {
		t.Errorf("after summon, state still hidden")
	}

	// Dismiss.
	if _, err := phase6CallRPC(t, srv, "presence.dismiss", nil); err != nil {
		t.Fatalf("presence.dismiss: %v", err)
	}
	if got := ctrl.State(); got != overlay.StateHidden {
		t.Errorf("after dismiss, state = %v, want hidden", got)
	}
}

// TestPhase6_PresenceState verifies presence.state returns the
// current state as a string.
func TestPhase6_PresenceState(t *testing.T) {
	ctrl := overlay.NewNoopController()
	subs := &Subsystems{
		Voice:   nil,
		Broker:  sse.NewBroker(),
		Audit:   nil,
		Overlay: ctrl,
	}
	srv := ipc.NewServer()
	registerPhase6Methods(srv, subs)

	resp, err := phase6CallRPC(t, srv, "presence.state", nil)
	if err != nil {
		t.Fatalf("presence.state: %v", err)
	}
	m, _ := resp.(map[string]any)
	if state, _ := m["state"].(string); state != "hidden" {
		t.Errorf("state = %q, want hidden", state)
	}
}

// TestPhase6_VoiceCancelNoVoice verifies voice.cancel returns a
// well-typed error when voice is unavailable (does not panic).
func TestPhase6_VoiceCancelNoVoice(t *testing.T) {
	subs := &Subsystems{
		Voice:   nil,
		Broker:  sse.NewBroker(),
		Audit:   nil,
		Overlay: overlay.NewNoopController(),
	}
	srv := ipc.NewServer()
	registerPhase6Methods(srv, subs)

	_, err := phase6CallRPC(t, srv, "voice.cancel", nil)
	if err == nil {
		t.Fatal("expected error from voice.cancel when voice is nil")
	}
	var ipcErr *ipc.Error
	if !errors.As(err, &ipcErr) {
		t.Fatalf("expected *ipc.Error, got %T", err)
	}
	if ipcErr.Code != ipc.CodeMethodNotFound {
		t.Errorf("code = %d, want %d", ipcErr.Code, ipc.CodeMethodNotFound)
	}
}

// TestPhase6_AgentAskEmptyQuery verifies agent.ask rejects empty
// queries with a clean error.
func TestPhase6_AgentAskEmptyQuery(t *testing.T) {
	subs := &Subsystems{
		Voice:   nil,
		Broker:  sse.NewBroker(),
		Audit:   nil,
		Overlay: overlay.NewNoopController(),
	}
	srv := ipc.NewServer()
	registerPhase6Methods(srv, subs)

	_, err := phase6CallRPC(t, srv, "agent.ask", []byte(`{"query":""}`))
	if err == nil {
		t.Fatal("expected error for empty query")
	}
}

// TestPhase6_OverlayShowRoutesToController verifies the
// overlay.show RPC actually invokes the overlay controller.
func TestPhase6_OverlayShowRoutesToController(t *testing.T) {
	ctrl := overlay.NewNoopController()
	subs := &Subsystems{
		Voice:   nil,
		Broker:  sse.NewBroker(),
		Audit:   nil,
		Overlay: ctrl,
	}
	srv := ipc.NewServer()
	registerWindowMethods(srv, subs)

	if _, err := phase6CallRPC(t, srv, "overlay.show", []byte(`{"at_cursor":true}`)); err != nil {
		t.Fatalf("overlay.show: %v", err)
	}
	if ctrl.State() == overlay.StateHidden {
		t.Error("overlay should not be hidden after overlay.show")
	}

	if _, err := phase6CallRPC(t, srv, "overlay.hide", nil); err != nil {
		t.Fatalf("overlay.hide: %v", err)
	}
	if ctrl.State() != overlay.StateHidden {
		t.Error("overlay should be hidden after overlay.hide")
	}
}

// callRPC invokes a method on the server and returns the
// unmarshaled result. Returns the result and an error if the
// call returned a JSON-RPC error.
func phase6CallRPC(t *testing.T, srv *ipc.Server, method string, params json.RawMessage) (any, error) {
	t.Helper()
	resp, err := srv.Handle(context.Background(), &ipc.Request{
		JSONRPC: "2.0",
		Method:  method,
		Params:  params,
		ID:      json.RawMessage("1"),
	})
	if err != nil {
		return nil, err
	}
	if resp.Error != nil {
		return nil, resp.Error
	}
	return resp.Result, nil
}
