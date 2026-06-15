// Phase 12 Hub E2E — drives the Skills Hub RPCs end-to-end.
// These tests verify the RPC surface shape and error handling when
// the hub is not configured (the default state).
package daemon

import (
	"testing"
)

// TestTrustE2E_HubSearchRequiresConfiguration verifies hub.search
// returns an error when the hub is not configured.
func TestTrustE2E_HubSearchRequiresConfiguration(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	_, err := trustCallRPC(t, addr, "hub.search", map[string]any{"query": "test", "limit": 10})
	if err == nil {
		t.Fatalf("expected error when hub not configured")
	}
}

// TestTrustE2E_HubGetRequiresConfiguration verifies hub.get returns
// an error when the hub is not configured.
func TestTrustE2E_HubGetRequiresConfiguration(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	_, err := trustCallRPC(t, addr, "hub.get", map[string]any{"id": "test-skill"})
	if err == nil {
		t.Fatalf("expected error when hub not configured")
	}
}

// TestTrustE2E_HubInstallRequiresConfiguration verifies hub.install
// returns an error when the hub is not configured.
func TestTrustE2E_HubInstallRequiresConfiguration(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	_, err := trustCallRPC(t, addr, "hub.install", map[string]any{"id": "test-skill"})
	if err == nil {
		t.Fatalf("expected error when hub not configured")
	}
}

// TestTrustE2E_HubPublishRequiresConfiguration verifies hub.publish
// returns an error when the hub is not configured.
func TestTrustE2E_HubPublishRequiresConfiguration(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()
	_, err := trustCallRPC(t, addr, "hub.publish", map[string]any{"id": "test-skill"})
	if err == nil {
		t.Fatalf("expected error when hub not configured")
	}
}
