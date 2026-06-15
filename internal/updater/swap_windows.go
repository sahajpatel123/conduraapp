//go:build windows

package updater

import (
	"fmt"
	"os"
	"path/filepath"
)

// On Windows the running executable cannot be replaced in-place.
// We stage a .cmd script that swaps after the process exits.
func swapExecutablePlatform(staged, target string) error {
	dir := filepath.Dir(target)
	script := filepath.Join(dir, "synaptic-apply-update.cmd")
	body := fmt.Sprintf("@echo off\r\n"+
		"timeout /t 2 /nobreak >nul\r\n"+
		"move /y %q %q\r\n"+
		"del %q\r\n",
		staged, target, script)
	if err := os.WriteFile(script, []byte(body), 0o700); err != nil {
		return fmt.Errorf("updater: write swap script: %w", err)
	}
	return ErrRestartRequired
}
