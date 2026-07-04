package updater

// CompletePendingUpdate applies a staged Windows update on restart.
// On Unix this is a no-op (in-place swap already happened).
// Returns true when a pending update was applied.
func CompletePendingUpdate() (bool, error) {
	return completePendingUpdatePlatform()
}
