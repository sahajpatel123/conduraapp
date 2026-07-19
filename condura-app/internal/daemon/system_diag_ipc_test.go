package daemon

import (
	"context"
	"encoding/json"
	"path/filepath"
	"testing"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/diag"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/storage"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/validate"
)

// newTestSubsystemsWithDataDir constructs a minimal Subsystems
// with a real storage.DB at dir, so that GeneralDataDir() returns
// the expected value. The DB is closed via t.Cleanup.
//
// The other fields of Subsystems stay at zero values — the
// system.diag / system.validate handlers only read Storage and
// the general data dir.
func newTestSubsystemsWithDataDir(t *testing.T, dir string) *Subsystems {
	t.Helper()
	db, err := storage.Open(context.Background(), storage.Config{
		Path:      filepath.Join(dir, "condura.db"),
		MasterKey: "",
	})
	if err != nil {
		t.Fatalf("storage.Open: %v", err)
	}
	t.Cleanup(func() { _ = db.Close() })
	return &Subsystems{Storage: db}
}

// TestSystemDiagIPC_ReturnsSnapshot pins the contract for the
// system.diag IPC method: it must return a Snapshot with the
// data_dir path matching Subsystems.GeneralDataDir(), and the
// JSON shape must round-trip (the GUI parses this).
func TestSystemDiagIPC_ReturnsSnapshot(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CONDURA_BACKUP_DIR", filepath.Join(dir, "no-backups"))
	subs := newTestSubsystemsWithDataDir(t, dir)

	got, err := systemDiagHandler(context.Background(), subs, json.RawMessage(nil))
	if err != nil {
		t.Fatalf("system.diag handler: %v", err)
	}
	snap, ok := got.(diag.Snapshot)
	if !ok {
		t.Fatalf("system.diag returned %T, want diag.Snapshot", got)
	}
	if snap.Paths.DataDir != dir {
		t.Errorf("Snapshot.Paths.DataDir = %q, want %q", snap.Paths.DataDir, dir)
	}
	if snap.Version == "" {
		t.Error("Snapshot.Version is empty; want the build version")
	}
	if snap.Timestamp == "" {
		t.Error("Snapshot.Timestamp is empty; want RFC3339")
	}
}

// TestSystemValidateIPC_ReturnsReport pins the contract for
// system.validate: returns a Report with the data_dir matching
// GeneralDataDir(), and the checks are populated.
func TestSystemValidateIPC_ReturnsReport(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CONDURA_BACKUP_DIR", filepath.Join(dir, "no-backups"))
	subs := newTestSubsystemsWithDataDir(t, dir)

	got, err := systemValidateHandler(context.Background(), subs, json.RawMessage(nil))
	if err != nil {
		t.Fatalf("system.validate handler: %v", err)
	}
	rep, ok := got.(validate.Report)
	if !ok {
		t.Fatalf("system.validate returned %T, want validate.Report", got)
	}
	if rep.DataDir != dir {
		t.Errorf("Report.DataDir = %q, want %q", rep.DataDir, dir)
	}
	if len(rep.Checks) != 7 {
		t.Errorf("len(Report.Checks) = %d, want 7", len(rep.Checks))
	}
}

// TestSystemDiagIPC_JSONRoundTrip pins the JSON-shape contract:
// the snapshot must marshal cleanly and the GUI can unmarshal
// it back into the same shape. This is the public contract
// for the IPC method.
func TestSystemDiagIPC_JSONRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CONDURA_BACKUP_DIR", filepath.Join(dir, "no-backups"))
	subs := newTestSubsystemsWithDataDir(t, dir)

	got, err := systemDiagHandler(context.Background(), subs, json.RawMessage(nil))
	if err != nil {
		t.Fatalf("system.diag handler: %v", err)
	}
	js, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var back diag.Snapshot
	if err := json.Unmarshal(js, &back); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if back.Version != got.(diag.Snapshot).Version {
		t.Errorf("Version drift: %q vs %q", back.Version, got.(diag.Snapshot).Version)
	}
}

// TestSystemValidateIPC_JSONRoundTrip pins the JSON-shape
// contract for system.validate: the report must marshal cleanly
// and the GUI can unmarshal it back.
func TestSystemValidateIPC_JSONRoundTrip(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("CONDURA_BACKUP_DIR", filepath.Join(dir, "no-backups"))
	subs := newTestSubsystemsWithDataDir(t, dir)

	got, err := systemValidateHandler(context.Background(), subs, json.RawMessage(nil))
	if err != nil {
		t.Fatalf("system.validate handler: %v", err)
	}
	js, err := json.Marshal(got)
	if err != nil {
		t.Fatalf("json.Marshal: %v", err)
	}
	var back validate.Report
	if err := json.Unmarshal(js, &back); err != nil {
		t.Fatalf("json.Unmarshal: %v", err)
	}
	if back.DataDir != got.(validate.Report).DataDir {
		t.Errorf("DataDir drift: %q vs %q", back.DataDir, got.(validate.Report).DataDir)
	}
}

// systemDiagHandler and systemValidateHandler are defined in
// methods.go (production code); the test file just calls them
// directly without re-defining them.
