//go:build !darwin

package backends

import "github.com/sahajpatel123/synapticapp/internal/computeruse"

type noopMCP struct{}

func newMCPImpl() macOSMCPImpl { return &noopMCP{} }

func (n *noopMCP) name() string      { return "macos-mcp" }
func (n *noopMCP) isAvailable() bool { return false }
func (n *noopMCP) captureScreen() (*computeruse.Screenshot, error) {
	return nil, computeruse.ErrNoBackend
}
func (n *noopMCP) getAXTree() (*computeruse.AXTree, error) { return nil, computeruse.ErrNoBackend }
func (n *noopMCP) execute(_ *computeruse.Action) (*computeruse.ActionResult, error) {
	return nil, computeruse.ErrNoBackend
}
