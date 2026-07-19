package logtail

import (
	"os"
	"path/filepath"
	"testing"
)

// writeLines creates a log file at path with the given lines
// (one per line, with trailing newlines). Used to set up
// deterministic test fixtures.
func writeLines(t *testing.T, path string, lines []string) {
	t.Helper()
	if err := os.WriteFile(path, []byte(joinLines(lines)), 0o600); err != nil {
		t.Fatal(err)
	}
}

// joinLines is the inverse of strings.Split on \n: produces
// a log-shaped file with one line per element.
func joinLines(lines []string) string {
	out := ""
	for i, l := range lines {
		if i > 0 {
			out += "\n"
		}
		out += l
	}
	if len(lines) > 0 {
		out += "\n" // trailing newline, like a real log
	}
	return out
}

// TestTail_MissingFileReturnsNil pins the fresh-install
// contract: a missing log file returns (nil, nil), not an
// error. The operator runs this command before the daemon
// has ever started, and "file not found" is unhelpful —
// "no log file yet" is the right answer.
func TestTail_MissingFileReturnsNil(t *testing.T) {
	missing := filepath.Join(t.TempDir(), "no-such.log")
	got, err := Tail(missing, 100)
	if err != nil {
		t.Errorf("Tail(missing) err = %v, want nil (fresh-install contract)", err)
	}
	if got != nil {
		t.Errorf("Tail(missing) = %v, want nil", got)
	}
}

// TestTail_LastNLinesNewestFirst pins the basic contract: the
// result is the last n lines of the file, in newest-first
// order. The "newest first" semantic is the key one —
// operators want to see what just happened, not what
// happened at startup.
func TestTail_LastNLinesNewestFirst(t *testing.T) {
	path := filepath.Join(t.TempDir(), "condura.log")
	lines := []string{
		"startup: init",
		"startup: connect",
		"info: handler ping",
		"info: handler version",
		"warn: slow lsm seek",
		"error: backup create failed",
	}
	writeLines(t, path, lines)

	got, err := Tail(path, 3)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{
		"error: backup create failed", // newest
		"warn: slow lsm seek",
		"info: handler version", // oldest of the 3
	}
	if !sliceEqual(got, want) {
		t.Errorf("Tail(3):\n  got  %v\n  want %v", got, want)
	}
}

// TestTail_AllLinesWhenFewerThanN pins the "fewer than n"
// case: a 2-line log file requested with n=10 returns the
// 2 lines (not 10 padding entries, not an error).
func TestTail_AllLinesWhenFewerThanN(t *testing.T) {
	path := filepath.Join(t.TempDir(), "condura.log")
	writeLines(t, path, []string{"line1", "line2"})

	got, err := Tail(path, 10)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"line2", "line1"} // newest first
	if !sliceEqual(got, want) {
		t.Errorf("Tail(10, 2 lines):\n  got  %v\n  want %v", got, want)
	}
}

// TestTail_ZeroOrNegativeLinesReturnsNil pins the boundary:
// n <= 0 returns (nil, nil) so callers can pass through
// user-provided values without a special case.
func TestTail_ZeroOrNegativeLinesReturnsNil(t *testing.T) {
	path := filepath.Join(t.TempDir(), "condura.log")
	writeLines(t, path, []string{"line1", "line2"})

	for _, n := range []int{0, -1, -100} {
		got, err := Tail(path, n)
		if err != nil {
			t.Errorf("Tail(n=%d) err = %v, want nil", n, err)
		}
		if got != nil {
			t.Errorf("Tail(n=%d) = %v, want nil", n, got)
		}
	}
}

// TestTail_ReadsRotatedSiblings pins the rotation contract:
// when the active log doesn't have enough lines, the function
// reads rotated siblings (condura.log.1, .2, ...) in order.
// .1 is the most recent rotation, so its lines come AFTER the
// active log's lines in the "last n" view.
func TestTail_ReadsRotatedSiblings(t *testing.T) {
	dir := t.TempDir()
	active := filepath.Join(dir, "condura.log")
	rot1 := filepath.Join(dir, "condura.log.1")
	rot2 := filepath.Join(dir, "condura.log.2")

	// rot2 is the OLDEST (rotation history: rot2 -> rot1 -> active).
	// Lines in each file (in chronological order; we reverse for
	// "newest first" output).
	writeLines(t, rot2, []string{"r2-1", "r2-2", "r2-3"})
	writeLines(t, rot1, []string{"r1-1", "r1-2", "r1-3"})
	writeLines(t, active, []string{"a-1", "a-2"})

	// Request 5 lines: should get a-2, a-1, r1-3, r1-2, r1-1
	// (2 from active + 3 from rot1, NO rot2).
	got, err := Tail(active, 5)
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"a-2", "a-1", "r1-3", "r1-2", "r1-1"}
	if !sliceEqual(got, want) {
		t.Errorf("Tail(5, rotation):\n  got  %v\n  want %v", got, want)
	}

	// Request 8 lines: should reach into rot2 for the last 3.
	got, err = Tail(active, 8)
	if err != nil {
		t.Fatal(err)
	}
	want = []string{"a-2", "a-1", "r1-3", "r1-2", "r1-1", "r2-3", "r2-2", "r2-1"}
	if !sliceEqual(got, want) {
		t.Errorf("Tail(8, rotation):\n  got  %v\n  want %v", got, want)
	}
}

// TestTail_StopsAtMissingRotation pins the "no more
// rotations" contract: when rotation 3 doesn't exist, the
// function stops reading without erroring. This is the
// common case for a fresh install (only the active log
// exists, no rotations yet).
func TestTail_StopsAtMissingRotation(t *testing.T) {
	dir := t.TempDir()
	active := filepath.Join(dir, "condura.log")
	writeLines(t, active, []string{"a-1", "a-2", "a-3"})

	// Request 10 lines: only 3 exist in active, no rotations.
	// Should return just the 3 from active, no error.
	got, err := Tail(active, 10)
	if err != nil {
		t.Fatalf("Tail returned error: %v", err)
	}
	want := []string{"a-3", "a-2", "a-1"}
	if !sliceEqual(got, want) {
		t.Errorf("Tail(10, no rotations):\n  got  %v\n  want %v", got, want)
	}
}

// TestLogFilePath pins the canonical log file path: the CLI
// and any GUI consumer must agree on this. The path is
// <dataDir>/logs/condura.log — matches the logger package's
// Rotate writer convention.
func TestLogFilePath(t *testing.T) {
	got := LogFilePath("/var/lib/condura")
	want := "/var/lib/condura/logs/condura.log"
	if got != want {
		t.Errorf("LogFilePath = %q, want %q", got, want)
	}
}

// sliceEqual compares two []string for equality.
func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
