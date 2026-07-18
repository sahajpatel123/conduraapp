package ipc

import (
	"os"
	"path/filepath"
	"testing"
)

// TestClient_Addr_ReturnsConfigured pins the Addr getter contract:
// Addr() MUST return the address the client was configured with.
// A regression that returned a different string (e.g., the resolved
// TCP host) would break callers that use Addr() to log the connection
// target.
func TestClient_Addr_ReturnsConfigured(t *testing.T) {
	c := &Client{addr: "127.0.0.1:9999", scheme: "http", host: "127.0.0.1:9999"}
	if got := c.Addr(); got != "127.0.0.1:9999" {
		t.Errorf("Addr() = %q, want %q", got, "127.0.0.1:9999")
	}
}

// TestClient_Addr_EmptyForUnconfigured pins the empty-Client
// contract: a zero-value Client (never Dialed) returns Addr() == "".
// Defensive — a regression that panicked on nil receiver or returned
// a literal default would surface here.
func TestClient_Addr_EmptyForUnconfigured(t *testing.T) {
	c := &Client{}
	if got := c.Addr(); got != "" {
		t.Errorf("zero Client.Addr() = %q, want \"\"", got)
	}
}

// TestReadAddrFile_ReadsExistingFile pins the happy-path contract:
// ReadAddrFile MUST return the file's trimmed contents when the
// addr file exists. This is the daemon's IPC discovery path: the
// GUI calls ReadAddrFile to find where the daemon is listening.
func TestReadAddrFile_ReadsExistingFile(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "condurad.addr")
	if err := os.WriteFile(path, []byte("127.0.0.1:9999"), 0o600); err != nil {
		t.Fatal(err)
	}

	got := ReadAddrFile(dir)
	if got != "127.0.0.1:9999" {
		t.Errorf("ReadAddrFile = %q, want %q", got, "127.0.0.1:9999")
	}
}

// TestReadAddrFile_TrimsWhitespace pins the trim contract: leading
// and trailing whitespace (newlines from `echo` or shell) MUST be
// stripped before returning. Without this, the GUI would try to
// dial "127.0.0.1:9999\n" and fail.
func TestReadAddrFile_TrimsWhitespace(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "condurad.addr")
	if err := os.WriteFile(path, []byte("  127.0.0.1:9999  \n"), 0o600); err != nil {
		t.Fatal(err)
	}

	got := ReadAddrFile(dir)
	if got != "127.0.0.1:9999" {
		t.Errorf("ReadAddrFile = %q, want %q (trimmed)", got, "127.0.0.1:9999")
	}
}

// TestReadAddrFile_MissingFileReturnsEmpty pins the missing-file
// contract: ReadAddrFile MUST return "" (NOT an error) when the
// addr file doesn't exist. The GUI checks for "" to decide
// "daemon not running" vs "daemon running on X". Returning an
// error would force every caller to wrap with tryRecover.
func TestReadAddrFile_MissingFileReturnsEmpty(t *testing.T) {
	dir := t.TempDir() // no condurad.addr written here
	got := ReadAddrFile(dir)
	if got != "" {
		t.Errorf("ReadAddrFile on missing file = %q, want \"\"", got)
	}
}

// TestReadAddrFile_EmptyFileReturnsEmpty pins the empty-file
// contract: ReadAddrFile MUST return "" when the file exists but
// is empty (e.g., daemon just started, addr file not yet written).
func TestReadAddrFile_EmptyFileReturnsEmpty(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "condurad.addr")
	if err := os.WriteFile(path, []byte{}, 0o600); err != nil {
		t.Fatal(err)
	}

	got := ReadAddrFile(dir)
	if got != "" {
		t.Errorf("ReadAddrFile on empty file = %q, want \"\"", got)
	}
}

// TestReadAddrFile_DirectoryNotFoundReturnsEmpty pins the
// path-resolution contract: ReadAddrFile on a non-existent
// directory MUST return "" (NOT panic, NOT error). The daemon's
// data dir may not exist on first launch.
func TestReadAddrFile_DirectoryNotFoundReturnsEmpty(t *testing.T) {
	got := ReadAddrFile("/this/directory/definitely/does/not/exist")
	if got != "" {
		t.Errorf("ReadAddrFile on nonexistent dir = %q, want \"\"", got)
	}
}

// TestDefaultDataDir_ReturnsHomeSlashCondura pins the home-dir
// fallback contract: DefaultDataDir MUST return
// $HOME/.condura. The GUI and daemon both rely on this default
// when no explicit --data-dir is passed.
func TestDefaultDataDir_ReturnsHomeSlashCondura(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home) // Windows

	got := DefaultDataDir()
	want := home + string(filepath.Separator) + ".condura"
	if got != want {
		t.Errorf("DefaultDataDir() = %q, want %q", got, want)
	}
}

// TestDefaultDataDir_PathSeparatorIsCorrect pins the separator
// contract: DefaultDataDir MUST use the OS-native path separator
// (forward slash on Unix, backslash on Windows). A regression to
// hardcoded "/" would fail on Windows.
func TestDefaultDataDir_PathSeparatorIsCorrect(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)

	got := DefaultDataDir()
	// The path MUST end with ".condura" (no separator between dir
	// name and .condura).
	if filepath.Base(got) != ".condura" {
		t.Errorf("DefaultDataDir() base = %q, want \".condura\"", filepath.Base(got))
	}
	// The parent MUST be the home dir we set.
	if filepath.Dir(got) != home {
		t.Errorf("DefaultDataDir() parent = %q, want %q", filepath.Dir(got), home)
	}
}
