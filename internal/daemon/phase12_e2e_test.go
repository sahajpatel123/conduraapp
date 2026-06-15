// Phase 12 E2E — drives the telemetry.status, skills.list/get/delete,
// and audit.list RPCs through the IPC layer to verify the fixes from
// the Phase 12 bug audit.
package daemon

import (
	"encoding/json"
	"testing"
)

func TestE2E_TelemetryStatus(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()

	// Default state: telemetry disabled.
	raw, err := trustCallRPC(t, addr, "telemetry.status", map[string]any{})
	if err != nil {
		t.Fatalf("telemetry.status: %v", err)
	}
	var status struct {
		Enabled  bool   `json:"enabled"`
		Endpoint string `json:"endpoint"`
	}
	if err := json.Unmarshal(raw, &status); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, raw)
	}
	if status.Enabled {
		t.Error("telemetry should be disabled by default")
	}

	// Enable telemetry.
	if _, err := trustCallRPC(t, addr, "telemetry.setEnabled", map[string]any{"enabled": true}); err != nil {
		t.Fatalf("telemetry.setEnabled: %v", err)
	}

	// Verify status reflects the change.
	raw, err = trustCallRPC(t, addr, "telemetry.status", map[string]any{})
	if err != nil {
		t.Fatalf("telemetry.status after enable: %v", err)
	}
	if err := json.Unmarshal(raw, &status); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if !status.Enabled {
		t.Error("telemetry should be enabled after setEnabled(true)")
	}
}

func TestE2E_SkillsListEmpty(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()

	raw, err := trustCallRPC(t, addr, "skills.list", map[string]any{})
	if err != nil {
		t.Fatalf("skills.list: %v", err)
	}
	// Should return an empty array, not null.
	var list []json.RawMessage
	if err := json.Unmarshal(raw, &list); err != nil {
		t.Fatalf("unmarshal skills.list: %v\nraw: %s", err, raw)
	}
	if len(list) != 0 {
		t.Errorf("expected empty skills list, got %d", len(list))
	}
}

func TestE2E_SkillsGetNotFound(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()

	_, err := trustCallRPC(t, addr, "skills.get", map[string]any{"id": "nonexistent"})
	if err == nil {
		t.Error("expected error for nonexistent skill")
	}
}

func TestE2E_SkillsDeleteNotFound(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()

	// Deleting a nonexistent skill is idempotent (standard SQL behavior).
	raw, err := trustCallRPC(t, addr, "skills.delete", map[string]any{"id": "nonexistent"})
	if err != nil {
		t.Fatalf("skills.delete: %v", err)
	}
	var result map[string]any
	if err := json.Unmarshal(raw, &result); err != nil {
		t.Fatalf("unmarshal: %v", err)
	}
	if result["ok"] != true {
		t.Error("skills.delete should return ok:true for nonexistent skill (idempotent)")
	}
}

func TestE2E_AuditListEmpty(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()

	raw, err := trustCallRPC(t, addr, "audit.list", map[string]any{"limit": 10})
	if err != nil {
		t.Fatalf("audit.list: %v", err)
	}
	// Should return an empty array, not null.
	var list []json.RawMessage
	if err := json.Unmarshal(raw, &list); err != nil {
		t.Fatalf("unmarshal audit.list: %v\nraw: %s", err, raw)
	}
	if list == nil {
		t.Error("audit.list returned nil slice (JSON null), expected empty array")
	}
}

func TestE2E_AuditListAfterAction(t *testing.T) {
	addr, _, cleanup := startTrustDaemon(t)
	defer cleanup()

	// Perform an action that creates an audit entry.
	if _, err := trustCallRPC(t, addr, "daemon.halt", map[string]any{"reason": "test"}); err != nil {
		t.Fatalf("daemon.halt: %v", err)
	}

	// Now audit.list should have at least one entry.
	raw, err := trustCallRPC(t, addr, "audit.list", map[string]any{"limit": 10})
	if err != nil {
		t.Fatalf("audit.list: %v", err)
	}
	var list []json.RawMessage
	if err := json.Unmarshal(raw, &list); err != nil {
		t.Fatalf("unmarshal: %v\nraw: %s", err, raw)
	}
	if len(list) == 0 {
		t.Error("audit.list should have entries after daemon.halt, got empty")
	}
}
