//go:build !unix

package main

import "os"

// fileIdentity is the no-Unix fallback. Returns zeros, which
// means rotation detection falls back to "trust the file
// size" mode (always-rotated = false). On non-Unix platforms
// the log rotation scheme uses file copies (not in-place
// truncation) so the inode-based detection isn't critical.
func fileIdentity(f *os.File) (dev, ino uint64) { return 0, 0 }
func fileIdentityFromStat(fi os.FileInfo) (dev, ino uint64) { return 0, 0 }
