//go:build !windows

package updater

import (
	"fmt"
	"os"
)

func swapExecutablePlatform(staged, target string) error {
	backup := target + ".old"
	_ = os.Remove(backup)
	if err := os.Rename(target, backup); err != nil {
		return fmt.Errorf("updater: backup current binary: %w", err)
	}
	if err := os.Rename(staged, target); err != nil {
		_ = os.Rename(backup, target)
		return fmt.Errorf("updater: install staged binary: %w", err)
	}
	_ = os.Chmod(target, 0o755) //nolint:gosec,mnd // installed binary must be executable
	return nil
}
