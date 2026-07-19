// Package homedir provides a single, cached entry point for the
// current user's home directory.
//
// The standard library's os.UserHomeDir() is fine in isolation,
// but condura-app calls it 10+ times across 5 packages during
// startup (lockfile.IsInstalled, config.loader platform branches,
// uninstall.manifest, crash reporting, TUI client init). Each
// call repeats the platform-specific lookup logic.
//
// homedir.Dir() collapses all those calls into a single cached
// lookup via sync.Once, and gives the rest of the codebase one
// place to mock the home directory in tests.
package homedir
