//go:build !windows

package updater

func completePendingUpdatePlatform() (bool, error) {
	return false, nil
}
