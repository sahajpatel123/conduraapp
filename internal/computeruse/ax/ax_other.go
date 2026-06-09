//go:build !darwin

package ax

import (
	"context"

	"github.com/sahajpatel123/synapticapp/internal/computeruse"
)

// Backend is not available on non-Darwin platforms.
type Backend struct{}

// New creates a new backend (no-op on non-Darwin).
func New() *Backend {
	return &Backend{}
}

// Name returns the backend identifier.
func (b *Backend) Name() string { return "ax-unavailable" }

// Capabilities returns nil.
func (b *Backend) Capabilities() []computeruse.Capability { return nil }

// IsAvailable returns false on non-Darwin platforms.
func (b *Backend) IsAvailable(_ context.Context) bool { return false }

// CaptureScreen returns ErrNoBackend.
func (b *Backend) CaptureScreen(_ context.Context) (*computeruse.Screenshot, error) {
	return nil, computeruse.ErrNoBackend
}

// GetAXTree returns ErrNoBackend.
func (b *Backend) GetAXTree(_ context.Context) (*computeruse.AXTree, error) {
	return nil, computeruse.ErrNoBackend
}

// Execute returns ErrNoBackend.
func (b *Backend) Execute(_ context.Context, action *computeruse.Action) (*computeruse.ActionResult, error) {
	return &computeruse.ActionResult{
		Success: false,
		Error:   computeruse.ErrNoBackend,
		Action:  action,
	}, computeruse.ErrNoBackend
}
