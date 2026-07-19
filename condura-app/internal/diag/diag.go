package diag

import (
	"os"
	"path/filepath"
	"time"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/backup"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/homedir"
	"github.com/sahajpatel123/conduraapp/condura-app/internal/version"
)

// Paths is the on-disk layout the snapshot reports. Centralized
// so the JSON field names are stable across refactors (the
// JSON tags are the public support-ticket contract).
type Paths struct {
	DataDir    string `json:"data_dir"`    // e.g. ~/.condura
	BackupDir  string `json:"backup_dir"`  // standard backup location
	ConfigFile string `json:"config_file"` // data_dir/config.yaml
	MainDB     string `json:"main_db"`     // data_dir/condura.db
	MemoryDB   string `json:"memory_db"`   // data_dir/memory.db
	SkillsDB   string `json:"skills_db"`   // data_dir/skills.db
	LogsDir    string `json:"logs_dir"`    // data_dir/logs
	LockFile   string `json:"lock_file"`   // data_dir/condurad.lock
	AddrFile   string `json:"addr_file"`   // data_dir/condurad.addr
}

// FileInfo describes one on-disk artifact: its path, size in
// bytes, and mtime (RFC3339). If the file is missing, Size is
// 0 and MTime is the empty string.
type FileInfo struct {
	Path  string `json:"path"`
	Size  int64  `json:"size"`
	MTime string `json:"mtime,omitempty"`
}

// Snapshot is the support-ticket payload. The JSON shape is
// versioned via the Version field; the client (GUI / CLI /
// scraper) can branch on that if the schema needs to evolve.
type Snapshot struct {
	Version   string     `json:"version"`
	Timestamp string     `json:"timestamp"` // RFC3339 when the snapshot was taken
	Paths     Paths      `json:"paths"`
	MainDB    FileInfo   `json:"main_db"`
	MemoryDB  FileInfo   `json:"memory_db"`
	SkillsDB  FileInfo   `json:"skills_db"`
	Config    FileInfo   `json:"config"`
	Backups   []FileInfo `json:"backups"` // newest-first
}

// Take gathers a snapshot for the given data dir (or the default
// ~/.condura if empty). The Snapshot's Timestamp is set to
// time.Now().UTC(); callers don't need to set it.
//
// Backups are gathered via backup.ListBackupArchives (the
// newest-first path filter from iter-9). Missing files are
// reported as zero-size FileInfo entries (not errors) so the
// snapshot is always well-formed JSON.
func Take(dataDir string) Snapshot {
	if dataDir == "" {
		if home, err := homedir.Dir(); err == nil {
			dataDir = filepath.Join(home, ".condura")
		} else {
			dataDir = ".condura" // best-effort fallback
		}
	}
	backupDir := backup.ResolveBackupDir(dataDir)

	now := time.Now().UTC()
	s := Snapshot{
		Version:   version.Version,
		Timestamp: now.Format(time.RFC3339),
		Paths: Paths{
			DataDir:    dataDir,
			BackupDir:  backupDir,
			ConfigFile: filepath.Join(dataDir, "config.yaml"),
			MainDB:     filepath.Join(dataDir, "condura.db"),
			MemoryDB:   filepath.Join(dataDir, "memory.db"),
			SkillsDB:   filepath.Join(dataDir, "skills.db"),
			LogsDir:    filepath.Join(dataDir, "logs"),
			LockFile:   filepath.Join(dataDir, "condurad.lock"),
			AddrFile:   filepath.Join(dataDir, "condurad.addr"),
		},
	}

	s.MainDB = stat(filepath.Join(dataDir, "condura.db"))
	s.MemoryDB = stat(filepath.Join(dataDir, "memory.db"))
	s.SkillsDB = stat(filepath.Join(dataDir, "skills.db"))
	s.Config = stat(filepath.Join(dataDir, "config.yaml"))

	// Backups — gather via the same path the CLI uses, so the
	// snapshot reflects what the operator would see in
	// `condura backup list`.
	//
	// Initialize as a non-nil empty slice so the JSON field
	// marshals as "[]" rather than "null". Support scrapers
	// don't have to special-case the null branch.
	s.Backups = []FileInfo{}
	if paths, err := backup.ListBackupArchives(backupDir); err == nil {
		for _, p := range paths {
			s.Backups = append(s.Backups, stat(p))
		}
	}
	// Missing backup dir / unreadable entries → empty slice;
	// the snapshot is still valid JSON.

	return s
}

// stat wraps os.Stat with the FileInfo shape. Missing files
// return a zero FileInfo with the path set (so the JSON
// reports "we know about this path; it's absent").
func stat(path string) FileInfo {
	fi, err := os.Stat(path)
	if err != nil {
		return FileInfo{Path: path, Size: 0}
	}
	return FileInfo{
		Path:  path,
		Size:  fi.Size(),
		MTime: fi.ModTime().UTC().Format(time.RFC3339),
	}
}
