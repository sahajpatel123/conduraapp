package backup

import (
	"os"
	"path/filepath"
)

// ResolveBackupDir returns the directory where encrypted backup
// archives are stored. Priority:
//  1. CONDURA_BACKUP_DIR environment variable (absolute path)
//  2. ~/Documents/condura-backups (MISSION §24.1 / decision #17)
//  3. <data-dir>/backups (daemon-local default)
func ResolveBackupDir(dataDir string) string {
	if dir := os.Getenv("CONDURA_BACKUP_DIR"); dir != "" {
		return dir
	}
	if home := userHomeDir(); home != "" {
		return filepath.Join(home, "Documents", "condura-backups")
	}
	return filepath.Join(dataDir, "backups")
}

func userHomeDir() string {
	if h := os.Getenv("HOME"); h != "" {
		return h
	}
	if h := os.Getenv("USERPROFILE"); h != "" {
		return h
	}
	h, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	return h
}
