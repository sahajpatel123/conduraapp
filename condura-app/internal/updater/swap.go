package updater

import (
	"fmt"
	"os"
	"path/filepath"
)

// swapExecutable atomically replaces target with staged. The staged file
// must already be verified (SHA256 + manifest signature).
func swapExecutable(staged, target string) error {
	if staged == "" || target == "" {
		return fmt.Errorf("updater: empty swap path")
	}
	if err := os.Chmod(staged, 0o700); err != nil { //nolint:gosec // staged before install
		return fmt.Errorf("updater: chmod staged: %w", err)
	}
	return swapExecutablePlatform(staged, target)
}

// currentExecutable returns the resolved path to the running binary.
func currentExecutable() (string, error) {
	exe, err := os.Executable()
	if err != nil {
		return "", err
	}
	return filepath.EvalSymlinks(exe)
}
