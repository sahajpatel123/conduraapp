package backends

import (
	"context"

	"github.com/sahajpatel123/conduraapp/internal/computeruse"
)

// MacOSMCPBackend implements computeruse.Backend using AppleScript
// and System Events. It provides comprehensive foreground interaction
// — click, type, scroll, menu navigation, window management.
//
// This is the third tier in the 4-tier router, falling back from
// mac-cua when PID-targeted background execution isn't sufficient.
type MacOSMCPBackend struct {
	impl macOSMCPImpl
}

type macOSMCPImpl interface {
	name() string
	isAvailable() bool
	captureScreen() (*computeruse.Screenshot, error)
	getAXTree() (*computeruse.AXTree, error)
	execute(action *computeruse.Action) (*computeruse.ActionResult, error)
}

var _ computeruse.Backend = (*MacOSMCPBackend)(nil)

// NewMacOSMCP creates a macOS-MCP backend for the current platform.
func NewMacOSMCP() *MacOSMCPBackend {
	return &MacOSMCPBackend{impl: newMCPImpl()}
}

// Name returns the backend identifier.
func (b *MacOSMCPBackend) Name() string { return b.impl.name() }

// Capabilities returns all supported operations including drag.
func (b *MacOSMCPBackend) Capabilities() []computeruse.Capability {
	return []computeruse.Capability{
		computeruse.CapScreenshot,
		computeruse.CapAXTree,
		computeruse.CapClick,
		computeruse.CapType,
		computeruse.CapScroll,
		computeruse.CapKeyPress,
		computeruse.CapLaunch,
		computeruse.CapFocus,
		computeruse.CapDrag,
	}
}

// IsAvailable checks if osascript is accessible.
func (b *MacOSMCPBackend) IsAvailable(_ context.Context) bool { return b.impl.isAvailable() }

// CaptureScreen captures the screen via screencapture(1).
func (b *MacOSMCPBackend) CaptureScreen(_ context.Context) (*computeruse.Screenshot, error) {
	return b.impl.captureScreen()
}

// GetAXTree reads the AX tree via osascript System Events.
func (b *MacOSMCPBackend) GetAXTree(_ context.Context) (*computeruse.AXTree, error) {
	return b.impl.getAXTree()
}

// Execute performs an action via AppleScript.
func (b *MacOSMCPBackend) Execute(_ context.Context, action *computeruse.Action) (*computeruse.ActionResult, error) {
	return b.impl.execute(action)
}
