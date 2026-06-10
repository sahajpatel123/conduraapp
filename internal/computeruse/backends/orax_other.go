//go:build !darwin || !cgo

package backends

import (
	"github.com/sahajpatel123/synapticapp/internal/computeruse"
)

// noopORAX is the stub implementation for non-darwin platforms.
type noopORAX struct{}

func newORAXImpl() oraXImpl { return &noopORAX{} }

func (n *noopORAX) name() string { return "orax" }

func (n *noopORAX) isAvailable() bool { return false }

func (n *noopORAX) captureScreen() (*computeruse.Screenshot, error) {
	return nil, computeruse.ErrNoBackend
}

func (n *noopORAX) getAXTree() (*computeruse.AXTree, error) {
	return nil, computeruse.ErrNoBackend
}

func (n *noopORAX) execute(_ *computeruse.Action) (*computeruse.ActionResult, error) {
	return nil, computeruse.ErrNoBackend
}
