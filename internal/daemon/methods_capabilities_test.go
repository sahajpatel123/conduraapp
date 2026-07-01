package daemon

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/sahajpatel123/conduraapp/internal/ipc"
)

// TestCapabilities_MethodRegistered verifies daemon.capabilities
// is a registered JSON-RPC method on a freshly-constructed
// Server. The handler is registered in registerMethods; this test
// just confirms the wiring is intact.
func TestCapabilities_MethodRegistered(t *testing.T) {
	srv := ipc.NewServer()
	registerCapabilitiesMethods(srv)
	if !srv.HasMethod("daemon.capabilities") {
		t.Fatal("daemon.capabilities not registered")
	}
}

// TestCapabilities_Shape pins the exact JSON shape the GUI
// depends on. Renaming a field, removing a layer, or marking
// in_process=false without the GUI knowing would silently break
// the "What this build can and can't do" panel — and the user
// would be told lies about what the kill switch can do.
//
// CLAUDE.md §2.1 invariant #4 is the contract: "user can always
// stop the agent, four independent mechanisms, the agent cannot
// disable any of them." The Layer 3 in_process flag is the only
// piece of that contract the daemon can lie about, so we pin
// every layer here.
func TestCapabilities_Shape(t *testing.T) {
	srv := ipc.NewServer()
	registerCapabilitiesMethods(srv)
	resp, err := srv.Handle(context.Background(), &ipc.Request{
		JSONRPC: "2.0", Method: "daemon.capabilities", ID: json.RawMessage("1"),
	})
	if err != nil {
		t.Fatalf("Handle: %v", err)
	}
	if resp.Error != nil {
		t.Fatalf("Error: %+v", resp.Error)
	}
	// ipc.Response.Result is json.RawMessage.
	raw, err := json.Marshal(resp.Result)
	if err != nil {
		t.Fatalf("re-marshal: %v", err)
	}
	var got map[string]any
	if err := json.Unmarshal(raw, &got); err != nil {
		t.Fatalf("unmarshal result: %v", err)
	}

	// version
	if _, ok := got["version"]; !ok {
		t.Error("capabilities: missing 'version' field")
	}

	// kill_switch
	ks, ok := got["kill_switch"].(map[string]any)
	if !ok {
		t.Fatalf("capabilities.kill_switch is not an object: %T", got["kill_switch"])
	}
	if v, _ := ks["layer1_hotkey"].(bool); !v {
		t.Error("kill_switch.layer1_hotkey must be true (CLAUDE.md §5.3)")
	}
	if v, _ := ks["layer2_watchdog"].(bool); !v {
		t.Error("kill_switch.layer2_watchdog must be true (the watchdog is available; arming is the user's choice)")
	}
	l3, ok := ks["layer3_network_isolation"].(map[string]any)
	if !ok {
		t.Fatalf("kill_switch.layer3_network_isolation is not an object: %T", ks["layer3_network_isolation"])
	}
	// The honest answer for v0.1.0: the guard runs inside the
	// daemon process. CLAUDE.md §33.5.2 row C4.14.
	if v, _ := l3["in_process"].(bool); !v {
		t.Error("kill_switch.layer3_network_isolation.in_process must be true in v0.1.0")
	}
	if v, _ := l3["os_process"].(bool); v {
		t.Error("kill_switch.layer3_network_isolation.os_process must be false until v0.2.0")
	}
	if got, want := l3["deferred_to"], "v0.2.0"; got != want {
		t.Errorf("kill_switch.layer3_network_isolation.deferred_to = %v, want %v", got, want)
	}
	if got, want := l3["reference"], "CLAUDE.md §33.5.2 row C4.14"; got != want {
		t.Errorf("kill_switch.layer3_network_isolation.reference = %v, want %v", got, want)
	}

	// computer_use
	cu, ok := got["computer_use"].(map[string]any)
	if !ok {
		t.Fatalf("capabilities.computer_use is not an object: %T", got["computer_use"])
	}
	for _, k := range []string{"orax", "mac_cua", "macos_mcp", "vision_cua"} {
		if _, ok := cu[k].(string); !ok {
			t.Errorf("computer_use.%s is not a string", k)
		}
	}
	if got, want := cu["vision_cua"], "disabled_default"; got != want {
		t.Errorf("computer_use.vision_cua = %v, want %v (vision is opt-in for cost reasons)", got, want)
	}

	// audit
	au, ok := got["audit"].(map[string]any)
	if !ok {
		t.Fatalf("capabilities.audit is not an object: %T", got["audit"])
	}
	for _, k := range []string{"redaction", "prune_tombstone", "hmac_subkey"} {
		if v, _ := au[k].(bool); !v {
			t.Errorf("audit.%s must be true (the wiring landed 2026-06-29; this guard is the regression test)", k)
		}
	}
}
