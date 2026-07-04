// Tests for Config.OverrideDataDir.

package config

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOverrideDataDir_RederivesDefaultPaths(t *testing.T) {
	c := Default()
	// After Default(), Storage.Path etc. are empty; Loader.Load()
	// fills them in via resolveEmptyPaths. We simulate that here.
	resolveEmptyPaths(c)
	oldDir := c.General.DataDir

	c.OverrideDataDir("/tmp/new-data")

	assert.Equal(t, "/tmp/new-data", c.General.DataDir)
	assert.Equal(t, filepath.Join("/tmp/new-data", "condura.db"), c.Storage.Path)
	assert.Equal(t, filepath.Join("/tmp/new-data", "backups"), c.Storage.Backup.Dir)
	assert.Equal(t, filepath.Join("/tmp/new-data", "cache"), c.General.CacheDir)
	_ = oldDir
}

func TestOverrideDataDir_PreservesExplicitPaths(t *testing.T) {
	c := Default()
	resolveEmptyPaths(c)
	// User set a custom storage path in YAML that is NOT under the
	// default data dir. The override must not clobber it.
	c.Storage.Path = "/var/lib/synaptic/prod.db"
	c.Storage.Backup.Dir = "/var/backups/synaptic"

	c.OverrideDataDir("/tmp/new-data")

	assert.Equal(t, "/tmp/new-data", c.General.DataDir)
	assert.Equal(t, "/var/lib/synaptic/prod.db", c.Storage.Path,
		"explicit storage path outside the old data dir should be preserved")
	assert.Equal(t, "/var/backups/synaptic", c.Storage.Backup.Dir,
		"explicit backup dir outside the old data dir should be preserved")
}

func TestOverrideDataDir_RederivesPathUnderOldDataDir(t *testing.T) {
	c := Default()
	resolveEmptyPaths(c)
	// Simulate a user who set storage.path = "<oldDataDir>/custom.db".
	// The override should re-derive it to point at the new data dir,
	// because the old path is no longer meaningful.
	c.Storage.Path = filepath.Join(c.General.DataDir, "custom.db")

	c.OverrideDataDir("/tmp/new-data")

	assert.Equal(t, filepath.Join("/tmp/new-data", "condura.db"), c.Storage.Path,
		"a storage path that lived under the old data dir should be re-derived")
}

func TestOverrideDataDir_HandlesEmptyStoragePath(t *testing.T) {
	// If the user never set a storage path, the override should
	// still populate it under the new data dir.
	c := Default()
	c.OverrideDataDir("/tmp/new-data")

	assert.Equal(t, filepath.Join("/tmp/new-data", "condura.db"), c.Storage.Path)
	assert.Equal(t, filepath.Join("/tmp/new-data", "backups"), c.Storage.Backup.Dir)
	assert.Equal(t, filepath.Join("/tmp/new-data", "cache"), c.General.CacheDir)
}
