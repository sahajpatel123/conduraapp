package backends

import (
	"context"

	"github.com/sahajpatel123/synapticapp/internal/computeruse"
)

// MacCUABackend implements computeruse.Backend using CGEventPostToPid
// for background-first operation. It targets specific applications by
// PID without stealing focus — the agent works while the user continues
// interacting with other apps.
//
// This is the second tier in the 4-tier router, falling back from
// ORAX Eye when foreground interaction isn't needed or available.
type MacCUABackend struct {
	impl macCUAImpl
}

type macCUAImpl interface {
	name() string
	isAvailable() bool
	captureScreen() (*computeruse.Screenshot, error)
	getAXTree() (*computeruse.AXTree, error)
	execute(action *computeruse.Action) (*computeruse.ActionResult, error)
}

var _ computeruse.Backend = (*MacCUABackend)(nil)

// NewMacCUA creates a mac-cua backend for the current platform.
func NewMacCUA() *MacCUABackend {
	return &MacCUABackend{impl: newMCImpl()}
}

// Name returns the backend identifier.
func (b *MacCUABackend) Name() string { return b.impl.name() }

// Capabilities returns the set of supported operations.
func (b *MacCUABackend) Capabilities() []computeruse.Capability {
	return []computeruse.Capability{
		computeruse.CapAXTree,
		computeruse.CapClick,
		computeruse.CapType,
		computeruse.CapScroll,
		computeruse.CapKeyPress,
		computeruse.CapFocus,
	}
}

// IsAvailable checks if Accessibility permission is granted.
func (b *MacCUABackend) IsAvailable(_ context.Context) bool { return b.impl.isAvailable() }

// CaptureScreen is not supported by mac-cua.
func (b *MacCUABackend) CaptureScreen(_ context.Context) (*computeruse.Screenshot, error) {
	return b.impl.captureScreen()
}

// GetAXTree reads the accessibility tree.
func (b *MacCUABackend) GetAXTree(_ context.Context) (*computeruse.AXTree, error) {
	return b.impl.getAXTree()
}

// Execute performs a background action via CGEventPostToPid.
func (b *MacCUABackend) Execute(_ context.Context, action *computeruse.Action) (*computeruse.ActionResult, error) {
	return b.impl.execute(action)
}
