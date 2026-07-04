package computeruse

import "errors"

var (
	// ErrNoBackend is returned when no backend is available for the requested action.
	ErrNoBackend = errors.New("computeruse: no available backend")

	// ErrPermissionDenied is returned when the backend lacks required permissions.
	ErrPermissionDenied = errors.New("computeruse: permission denied (accessibility permission required)")

	// ErrElementNotFound is returned when the target UI element cannot be found.
	ErrElementNotFound = errors.New("computeruse: target element not found")

	// ErrActionTimeout is returned when an action exceeds its timeout.
	ErrActionTimeout = errors.New("computeruse: action timed out")

	// ErrStaleState is returned when twin-snapshot verification detects stale state.
	ErrStaleState = errors.New("computeruse: stale state detected, aborting action")

	// ErrUserInterruption is returned when the user interacts during an action.
	ErrUserInterruption = errors.New("computeruse: user interruption detected")

	// ErrUnsupportedAction is returned when the backend doesn't support the action.
	ErrUnsupportedAction = errors.New("computeruse: action not supported by backend")

	// ErrScreenshotFailed is returned when screenshot capture fails.
	ErrScreenshotFailed = errors.New("computeruse: screenshot capture failed")

	// ErrAXTreeFailed is returned when AX tree capture fails.
	ErrAXTreeFailed = errors.New("computeruse: accessibility tree capture failed")
)
