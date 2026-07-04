//go:build windows

package updater

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

const pendingScriptName = "synaptic-apply-update.cmd"

func completePendingUpdatePlatform() (bool, error) {
	target, err := currentExecutable()
	if err != nil {
		return false, err
	}
	dir := filepath.Dir(target)
	script := filepath.Join(dir, pendingScriptName)
	if _, err := os.Stat(script); err != nil {
		return false, nil
	}
	cacheDir := filepath.Join(userHome(), ".condura", "cache")
	staged, err := findStagedUpdate(cacheDir)
	if err != nil || staged == "" {
		removePendingScript(dir)
		return false, nil
	}
	if err := renameSwap(staged, target); err != nil {
		return false, err
	}
	removePendingScript(dir)
	return true, nil
}

func findStagedUpdate(cacheDir string) (string, error) {
	entries, err := os.ReadDir(cacheDir)
	if err != nil {
		if os.IsNotExist(err) {
			return "", nil
		}
		return "", err
	}
	var newest string
	for _, e := range entries {
		if e.IsDir() {
			continue
		}
		name := e.Name()
		if !strings.HasPrefix(name, "synaptic-update-") {
			continue
		}
		path := filepath.Join(cacheDir, name)
		if newest == "" || name > filepath.Base(newest) {
			newest = path
		}
	}
	return newest, nil
}

func renameSwap(staged, target string) error {
	backup := target + ".old"
	_ = os.Remove(backup)
	if err := os.Rename(target, backup); err != nil && !os.IsNotExist(err) {
		return fmt.Errorf("updater: backup current: %w", err)
	}
	if err := os.Rename(staged, target); err != nil {
		_ = os.Rename(backup, target)
		return fmt.Errorf("updater: install staged: %w", err)
	}
	_ = os.Remove(backup)
	return nil
}

func removePendingScript(dir string) {
	_ = os.Remove(filepath.Join(dir, pendingScriptName))
}
