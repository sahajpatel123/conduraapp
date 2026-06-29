// Package backends provides platform-specific computer-use backend
// implementations. Each backend satisfies the computeruse.Backend
// interface and is selected by the 4-tier Router.
//
// Backend priority per MISSION S11.2:
//  1. ORAX Eye  - structured AX tree, free, ~50ms
//  2. mac-cua   - background-first, CGEventPostToPid
//  3. macOS-MCP - comprehensive, foreground interaction
//  4. Vision CUA - Anthropic/OpenAI, last resort
package backends

import (
	"context"

	"github.com/sahajpatel123/conduraapp/internal/computeruse"
)

// ORAXBackend implements computeruse.Backend using the macOS
// Accessibility API (AX tree) and CoreGraphics (screenshots, events).
// This is the first and cheapest tier in the 4-tier router.
//
// Build: requires CGO and the ApplicationServices + CoreGraphics
// frameworks (macOS only).
type ORAXBackend struct {
	impl oraXImpl
}

// oraXImpl is the platform-specific implementation. On darwin it uses
// CGO CGEvent/CGImage; on other platforms it returns stub errors.
type oraXImpl interface {
	name() string
	isAvailable() bool
	captureScreen() (*computeruse.Screenshot, error)
	getAXTree() (*computeruse.AXTree, error)
	execute(action *computeruse.Action) (*computeruse.ActionResult, error)
}

// Compile-time check: ORAXBackend satisfies computeruse.Backend.
var _ computeruse.Backend = (*ORAXBackend)(nil)

// NewORAX creates an ORAX Eye backend for the current platform.
func NewORAX() *ORAXBackend {
	return &ORAXBackend{impl: newORAXImpl()}
}

// Name returns the backend identifier.
func (b *ORAXBackend) Name() string { return b.impl.name() }

// Capabilities returns all capabilities supported by ORAX Eye.
func (b *ORAXBackend) Capabilities() []computeruse.Capability {
	return []computeruse.Capability{
		computeruse.CapScreenshot,
		computeruse.CapAXTree,
		computeruse.CapClick,
		computeruse.CapType,
		computeruse.CapScroll,
		computeruse.CapKeyPress,
		computeruse.CapLaunch,
		computeruse.CapFocus,
	}
}

// IsAvailable checks if the Accessibility API is available and
// the app has been granted the required permission.
func (b *ORAXBackend) IsAvailable(_ context.Context) bool {
	return b.impl.isAvailable()
}

// CaptureScreen captures a screenshot of the entire screen.
func (b *ORAXBackend) CaptureScreen(_ context.Context) (*computeruse.Screenshot, error) {
	return b.impl.captureScreen()
}

// GetAXTree reads the accessibility tree from the focused application.
func (b *ORAXBackend) GetAXTree(_ context.Context) (*computeruse.AXTree, error) {
	return b.impl.getAXTree()
}

// Execute performs a computer-use action through the AX API
// and CoreGraphics.
func (b *ORAXBackend) Execute(_ context.Context, action *computeruse.Action) (*computeruse.ActionResult, error) {
	return b.impl.execute(action)
}
