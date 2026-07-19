package logtail

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// LogFilePath returns the canonical path to the daemon's
// log file: <dataDir>/logs/condura.log. Centralized so the
// CLI and the GUI agree on the same path (and so a future
// log-rotation scheme change touches one line, not two).
func LogFilePath(dataDir string) string {
	return filepath.Join(dataDir, "logs", "condura.log")
}

// Tail returns the last n lines of path, newest first. If n
// is <= 0, returns nil. If path doesn't exist, returns
// (nil, nil) — see package doc for rationale.
//
// Rotated siblings (condura.log.1, condura.log.2, ...) are
// merged transparently if more than n lines are needed to
// reach "the last n". The merge is time-ordered: the .1
// file is the most recent rotation, so its lines come
// AFTER condura.log's lines and BEFORE .2's lines in the
// "last n lines" view.
//
// Memory bound: O(n) for the result slice. The implementation
// uses a reverse-seek strategy on the active log + a streaming
// line counter on rotated siblings.
func Tail(path string, n int) ([]string, error) {
	if n <= 0 {
		return nil, nil
	}
	// Active log first.
	active, err := tailFile(path, n)
	if err != nil {
		return nil, fmt.Errorf("logtail: read %s: %w", path, err)
	}
	// If we got fewer than n lines from the active log, look
	// at rotated siblings (condura.log.1, .2, ...) until we
	// have n or we run out of siblings.
	if len(active) < n {
		needed := n - len(active)
		for i := 1; ; i++ {
			rotated := path + "." + itoa(i)
			more, err := tailFile(rotated, needed)
			if err != nil {
				if os.IsNotExist(err) {
					break // no more rotations
				}
				return nil, fmt.Errorf("logtail: read %s: %w", rotated, err)
			}
			if len(more) == 0 {
				break
			}
			// tailFile already returns newest-first, so `more`
			// is the most-recent rotation's lines in newest-
			// first order. We want to APPEND them to `active`
			// (which is also newest-first) and then take the
			// first n. The .1 file is NEWER than .2, so its
			// lines are more recent than .2's — append in
			// rotation-number order.
			active = append(active, more...)
			if len(active) >= n {
				active = active[:n]
				break
			}
			needed = n - len(active)
		}
	}
	return active, nil
}

// tailFile reads the last n lines of path. Used for both the
// active log and rotated siblings. Returns (nil, nil) for
// missing files so Tail() can treat "no more rotations" as
// "we're done" rather than a hard error.
func tailFile(path string, n int) ([]string, error) {
	f, err := os.Open(path) //nolint:gosec // path is operator-supplied via --data-dir; not user-tainted.
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, err
	}
	defer func() { _ = f.Close() }()

	// Reverse-seek strategy: read the file in chunks from the
	// end, count newlines, stop when we have n+1 (the +1 lets
	// us strip the partial-line at the start of the first
	// chunk we landed on). Memory bound: O(chunk + n).
	const chunkSize = 64 * 1024
	stat, err := f.Stat()
	if err != nil {
		return nil, err
	}
	size := stat.Size()
	if size == 0 {
		return nil, nil
	}

	// Read chunks from the end until we have enough newlines.
	var (
		buf       []byte
		offset    = size
		readChunk = func() error {
			readSize := int64(chunkSize)
			if readSize > offset {
				readSize = offset
			}
			chunk := make([]byte, readSize)
			n, err := f.ReadAt(chunk, offset-readSize)
			if err != nil {
				return err
			}
			buf = append(chunk[:n], buf...)
			offset -= int64(n)
			return nil
		}
	)
	for {
		if err := readChunk(); err != nil {
			return nil, err
		}
		// Count newlines; if we have at least n+1, we have
		// enough to produce n full lines.
		newlines := strings.Count(string(buf), "\n")
		if newlines > n {
			break
		}
		if offset == 0 {
			break // reached start of file
		}
	}

	// Split into lines and take the last n.
	lines := strings.Split(string(buf), "\n")
	// bufio.ScanLines-style trim: remove trailing empty string
	// from a final newline.
	if len(lines) > 0 && lines[len(lines)-1] == "" {
		lines = lines[:len(lines)-1]
	}
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return reverse(lines), nil
}

// reverse returns the reverse of s in-place, returning the
// same slice (callers don't need to allocate).
func reverse(s []string) []string {
	for i, j := 0, len(s)-1; i < j; i, j = i+1, j-1 {
		s[i], s[j] = s[j], s[i]
	}
	return s
}

// itoa is a tiny base-10 int-to-string helper that avoids
// importing strconv for one call site.
func itoa(n int) string {
	if n == 0 {
		return "0"
	}
	var b [12]byte
	i := len(b)
	for n > 0 {
		i--
		b[i] = byte('0' + n%10)
		n /= 10
	}
	return string(b[i:])
}
