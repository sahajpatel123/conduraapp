//go:build unix

package main

import (
	"os"
	"syscall"
)

// syscallStat is a type alias for the platform's syscall.Stat_t
// (which differs per-Unix: *syscall.Stat_t on Linux/Darwin,
// *syscall.FreebsdStat_t on FreeBSD, etc.). The follow_unix.go
// implementation works for any Unix with Ino/Dev in its
// Stat_t. If your platform needs a different type, add a
// build-tag-specific file.
type syscallStat = syscall.Stat_t

// fileIdentity returns the (device, inode) pair for f, used
// for rotation detection. On Unix systems this is the real
// syscall.Stat_t.Dev / .Ino fields. On other platforms (see
// follow_other.go) the implementation returns zeros — rotation
// detection falls back to "trust the file size" mode.
func fileIdentity(f *os.File) (dev, ino uint64) {
	fi, err := f.Stat()
	if err != nil {
		return 0, 0
	}
	return fileIdentityFromStat(fi)
}

// fileIdentityFromStat extracts the (device, inode) pair from
// an os.FileInfo. Unix-specific via syscall.Stat_t; on other
// platforms (follow_other.go) returns zeros.
func fileIdentityFromStat(fi os.FileInfo) (dev, ino uint64) {
	stat, ok := fi.Sys().(*syscallStat)
	if !ok {
		return 0, 0
	}
	return uint64(stat.Dev), uint64(stat.Ino)
}
