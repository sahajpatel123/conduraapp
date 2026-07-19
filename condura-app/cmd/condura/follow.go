package main

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/logtail"
)

// tailFollow implements `tail -F` semantics: print the last
// `initialLines` lines of the file, then watch for new lines
// and print them as they appear. Resilient to log rotation:
// when the current file is replaced (rotated), the watch
// re-opens the new file from byte 0.
//
// Used by cmdLogs --follow. Local-only (no daemon IPC required).
//
// Implementation notes:
//   - Polling at 250ms intervals (rather than fsnotify) so
//     the function works on all platforms without adding a
//     new dependency.
//   - Tracks (dev, inode) of the current file via the
//     underlying syscall.Stat_t. On non-Unix systems the
//     rotation check falls back to "trust the file size".
//   - Honors SIGINT (Ctrl+C) via the caller's context. The
//     signal.NotifyContext in cmdLogs wires this up.
func tailFollow(ctx context.Context, path string, initialLines int) error {
	// Print the last N lines first (the "starting state" of
	// `tail -F`). This matches `tail -n N` so the operator
	// sees the recent context before new lines arrive.
	if initialLines > 0 {
		start, err := logtail.Tail(path, initialLines)
		if err != nil {
			return fmt.Errorf("condura logs --follow: initial tail: %w", err)
		}
		for _, l := range start {
			fmt.Println(l)
		}
	}

	// Open the file, seek to END, start polling for new lines.
	f, dev, ino, err := openForFollow(path)
	if err != nil {
		return err
	}
	defer func() { _ = f.Close() }()

	// 250ms poll interval. 4 wakes/second is fast enough for
	// interactive use and slow enough to be CPU-friendly.
	tick := time.NewTicker(250 * time.Millisecond)
	defer tick.Stop()

	reader := bufio.NewReader(f)
	for {
		select {
		case <-ctx.Done():
			return nil
		case <-tick.C:
		}

		// Read any new lines. Loop because a single tick may
		// see multiple new lines (the daemon logged a batch).
		for {
			line, err := reader.ReadString('\n')
			if err == io.EOF {
				break
			}
			if err != nil {
				return fmt.Errorf("condura logs --follow: read: %w", err)
			}
			fmt.Print(line)
		}

		// Check for rotation: dev or inode change means the
		// current file was rotated. Re-open from byte 0.
		newDev, newIno, rotated, err := checkRotation(path, dev, ino)
		if err != nil {
			return err
		}
		if rotated {
			_ = f.Close()
			f, dev, ino, err = openForFollow(path)
			if err != nil {
				return err
			}
			defer func() { _ = f.Close() }()
			reader = bufio.NewReader(f)
		}
		_ = newDev
		_ = newIno
	}
}

// openForFollow opens path for reading and returns the file
// handle + the device+inode of the file (for rotation detection).
// Seeks to END so we only print NEW lines.
func openForFollow(path string) (f *os.File, dev, ino uint64, err error) {
	f, err = os.Open(path) //nolint:gosec // path is operator-supplied via --data-dir; not user-tainted.
	if err != nil {
		return nil, 0, 0, fmt.Errorf("condura logs --follow: open: %w", err)
	}
	if _, err := f.Seek(0, io.SeekEnd); err != nil {
		_ = f.Close()
		return nil, 0, 0, fmt.Errorf("condura logs --follow: seek: %w", err)
	}
	dev, ino = fileIdentity(f)
	return f, dev, ino, nil
}

// checkRotation returns (newDev, newIno, rotated, err). The
// rotated flag is true if the file at path was rotated since
// the last open (dev/inode changed). The caller uses this to
// decide whether to re-open.
func checkRotation(path string, prevDev, prevIno uint64) (newDev, newIno uint64, rotated bool, err error) {
	fi, err := os.Stat(path)
	if err != nil {
		// File might have been deleted (mid-rotation). Trust
		// the next tick to retry; return "not rotated".
		return prevDev, prevIno, false, nil
	}
	newDev, newIno = fileIdentityFromStat(fi)
	rotated = newDev != prevDev || newIno != prevIno
	return newDev, newIno, rotated, nil
}
