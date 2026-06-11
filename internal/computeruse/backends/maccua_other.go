//go:build !darwin || !cgo

package backends

import "github.com/sahajpatel123/synapticapp/internal/computeruse"

type noopMC struct{}

func newMCImpl() macCUAImpl { return &noopMC{} }

func (n *noopMC) name() string                                     { return "mac-cua" }
func (n *noopMC) isAvailable() bool                                { return false }
func (n *noopMC) captureScreen() (*computeruse.Screenshot, error)  { return nil, computeruse.ErrNoBackend }
func (n *noopMC) getAXTree() (*computeruse.AXTree, error)          { return nil, computeruse.ErrNoBackend }
func (n *noopMC) execute(_ *computeruse.Action) (*computeruse.ActionResult, error) {
	return nil, computeruse.ErrNoBackend
}
