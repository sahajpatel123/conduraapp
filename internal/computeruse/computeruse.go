// Package computeruse provides the core interfaces and types for
// interacting with the operating system — reading the accessibility tree,
// capturing screenshots, and executing actions like clicks and keystrokes.
//
// The package defines a Backend interface that platform-specific
// implementations satisfy. A 4-tier router selects the cheapest
// available backend for each action.
//
// All computer-use actions go through the Gatekeeper (Phase 4.0)
// before execution. Twin-snapshot verification (Phase 5.2) prevents
// stale-state actions.
package computeruse

import (
	"context"
	"fmt"
	"time"

	"github.com/sahajpatel123/conduraapp/internal/blastradius"
	"github.com/sahajpatel123/conduraapp/internal/gatekeeper"
)

// Backend is the interface that platform-specific computer-use
// implementations must satisfy. Each backend provides different
// capabilities and tradeoffs.
type Backend interface {
	// Name returns the backend identifier (e.g., "orax-eye", "mac-cua").
	Name() string

	// Capabilities returns the set of actions this backend supports.
	Capabilities() []Capability

	// CaptureScreen captures a screenshot of the current screen
	// or focused window.
	CaptureScreen(ctx context.Context) (*Screenshot, error)

	// GetAXTree reads the accessibility tree from the OS.
	GetAXTree(ctx context.Context) (*AXTree, error)

	// Execute performs a computer-use action (click, type, scroll, etc.).
	Execute(ctx context.Context, action *Action) (*ActionResult, error)

	// IsAvailable checks if the backend is available and has
	// required permissions.
	IsAvailable(ctx context.Context) bool
}

// Capability describes what a backend can do.
type Capability string

// Capabilities supported by computer-use backends.
const (
	CapScreenshot Capability = "screenshot" // Capture screen image
	CapAXTree     Capability = "ax_tree"    // Read accessibility tree
	CapClick      Capability = "click"      // Click UI element
	CapType       Capability = "type"       // Type text
	CapScroll     Capability = "scroll"     // Scroll view
	CapKeyPress   Capability = "key_press"  // Press keyboard key
	CapDrag       Capability = "drag"       // Drag element
	CapLaunch     Capability = "launch"     // Launch application
	CapFocus      Capability = "focus"      // Focus window
)

// ActionType is the type of computer-use action.
type ActionType string

// Action types supported by the computer-use system.
const (
	ActionClick    ActionType = "click"     // Click UI element
	ActionTypeText ActionType = "type"      // Type text
	ActionScroll   ActionType = "scroll"    // Scroll view
	ActionKeyPress ActionType = "key_press" // Press keyboard key
	ActionDrag     ActionType = "drag"      // Drag element
	ActionLaunch   ActionType = "launch"    // Launch application
	ActionFocus    ActionType = "focus"     // Focus window
	ActionWait     ActionType = "wait"      // Wait for condition
)

// Action describes a single computer-use action to execute.
type Action struct {
	// Type is the action to perform.
	Type ActionType

	// Target identifies the UI element to act on.
	// Either Target or Bounds must be set, not both.
	Target *Target

	// Value is the text to type or key to press.
	Value string

	// Bounds provides fallback coordinates when no AX element is available.
	Bounds *Rect

	// Timeout for the action (default: 5s).
	Timeout time.Duration

	// AppPID is the process ID of the target application.
	// If zero, the action applies to the focused window.
	AppPID int32
}

// Target identifies a UI element by its accessibility attributes.
type Target struct {
	// Role is the AX role (e.g., "AXButton", "AXTextField").
	Role string

	// Title is the AX title or label.
	Title string

	// Value is the AX value.
	Value string

	// Description is the AX description.
	Description string

	// Index is the element index when multiple elements match.
	Index int
}

// Rect represents a bounding box in screen coordinates.
type Rect struct {
	X      float64
	Y      float64
	Width  float64
	Height float64
}

// Point represents a point in screen coordinates.
type Point struct {
	X float64
	Y float64
}

// Screenshot is a captured image of the screen or window.
type Screenshot struct {
	Image     []byte    // PNG image data
	Width     int       // Image width in pixels
	Height    int       // Image height in pixels
	Timestamp time.Time // When the screenshot was taken
	WindowID  uint32    // 0 for full screen
	PID       int32     // process ID of captured window
}

// AXTree is the accessibility tree of the screen or window.
type AXTree struct {
	Root      *AXNode
	Timestamp time.Time
	WindowID  uint32
	PID       int32
}

// AXNode is a single node in the accessibility tree.
type AXNode struct {
	Role        string
	Title       string
	Value       string
	Description string
	Bounds      *Rect
	Children    []*AXNode
	Attributes  map[string]interface{}
}

// ActionResult is the result of executing a computer-use action.
type ActionResult struct {
	Success    bool
	Error      error
	Screenshot *Screenshot // post-action screenshot (for verification)
	Duration   time.Duration
	Action     *Action
}

// ToBlastRadius converts a computeruse.Action to a blastradius.Action
// for safety classification. This is the structural bridge between the
// three action types (blastradius, computeruse, agent).
func (a *Action) ToBlastRadius() blastradius.Action {
	kind := "computeruse." + string(a.Type)
	ba := blastradius.Action{Kind: kind, Body: a.Value}
	if a.Target != nil {
		ba.TargetApp = a.Target.Title
	}
	return ba
}

// Execute performs the action through the Gatekeeper. If the
// gatekeeper denies the action, an error is returned and no
// physical execution occurs.
func (ge *GatedExecutor) Execute(ctx context.Context, action *Action) (*ActionResult, error) {
	ba := action.ToBlastRadius()
	decision, reason := ge.gate.Evaluate(ctx, ba)
	if decision != gatekeeper.Allow {
		return &ActionResult{
			Success: false,
			Error:   fmt.Errorf("gatekeeper denied: %s", reason),
			Action:  action,
		}, fmt.Errorf("gatekeeper denied: %s", reason)
	}
	return ge.cu.Execute(ctx, action)
}

// ElementQuery describes how to find a UI element.
type ElementQuery struct {
	Role        string
	Title       string
	Value       string
	Description string
	Index       int
}

// ComputerUse provides the main interface for interacting with the OS.
type ComputerUse struct {
	router *Router
}

// GatedExecutor wraps execution through the Gatekeeper, ensuring
// every computer-use action is safety-checked before running.
type GatedExecutor struct {
	cu   *ComputerUse
	gate gatekeeper.Gatekeeper
}

// NewGatedExecutor creates a GatedExecutor that routes all
// computer-use actions through the given gatekeeper.
func NewGatedExecutor(cu *ComputerUse, gate gatekeeper.Gatekeeper) *GatedExecutor {
	return &GatedExecutor{cu: cu, gate: gate}
}

// CU returns the underlying ComputerUse pipeline. Used by the
// agent loop in subsystems.go to wrap a real CU pipeline through
// agent.NewComputerUseExecutor so agent.Actions flow into the
// same gated backends. The GatedExecutor's own gate still applies
// via computeruse.Execute → ge.cu.Execute; the agent executor
// calls CU() to skip the GatedExecutor's redundant gate check.
func (ge *GatedExecutor) CU() *ComputerUse {
	if ge == nil {
		return nil
	}
	return ge.cu
}

// New creates a new ComputerUse instance with the given backends.
func New(backends ...Backend) *ComputerUse {
	return &ComputerUse{
		router: NewRouter(backends...),
	}
}

// CaptureScreen captures a screenshot using the best available backend.
func (c *ComputerUse) CaptureScreen(ctx context.Context) (*Screenshot, error) {
	return Execute(c.router, ctx, func(b Backend) (*Screenshot, error) {
		return b.CaptureScreen(ctx)
	})
}

// GetAXTree reads the accessibility tree using the best available backend.
func (c *ComputerUse) GetAXTree(ctx context.Context) (*AXTree, error) {
	return Execute(c.router, ctx, func(b Backend) (*AXTree, error) {
		return b.GetAXTree(ctx)
	})
}

// Execute performs a computer-use action using the best available backend.
func (c *ComputerUse) Execute(ctx context.Context, action *Action) (*ActionResult, error) {
	return c.router.ExecuteAction(ctx, action)
}

// FindElement finds a UI element matching the query in the AX tree.
func (c *ComputerUse) FindElement(ctx context.Context, query *ElementQuery) (*AXNode, error) {
	tree, err := c.GetAXTree(ctx)
	if err != nil {
		return nil, err
	}
	return findNode(tree.Root, query), nil
}

// GetElementAtPoint finds the UI element at the given screen coordinates.
func (c *ComputerUse) GetElementAtPoint(ctx context.Context, x, y float64) (*AXNode, error) {
	tree, err := c.GetAXTree(ctx)
	if err != nil {
		return nil, err
	}
	return findNodeAtPoint(tree.Root, x, y), nil
}

// findNode recursively searches for a node matching the query.
// When query.Index > 0, it returns the Nth match instead of the first.
func findNode(node *AXNode, query *ElementQuery) *AXNode {
	if node == nil {
		return nil
	}

	matches := findNodeAll(node, query, 1)
	if len(matches) > 0 {
		idx := query.Index
		if idx < 0 || idx >= len(matches) {
			idx = 0
		}
		return matches[idx]
	}
	return nil
}

// findNodeAll recursively collects all nodes matching the query, up to
// limit results. A limit of 0 means collect all.
func findNodeAll(node *AXNode, query *ElementQuery, limit int) []*AXNode {
	if node == nil {
		return nil
	}
	var result []*AXNode
	if matchesQuery(node, query) {
		result = append(result, node)
		if limit > 0 && len(result) >= limit {
			return result
		}
	}
	for _, child := range node.Children {
		result = append(result, findNodeAll(child, query, limit)...)
		if limit > 0 && len(result) >= limit {
			return result
		}
	}
	return result
}

// matchesQuery checks if a node matches the element query.
func matchesQuery(node *AXNode, query *ElementQuery) bool {
	if query.Role != "" && node.Role != query.Role {
		return false
	}
	if query.Title != "" && node.Title != query.Title {
		return false
	}
	if query.Value != "" && node.Value != query.Value {
		return false
	}
	if query.Description != "" && node.Description != query.Description {
		return false
	}
	return true
}

// findNodeAtPoint recursively searches for a node containing the point.
// It returns the most specific (deepest) node that contains the point.
func findNodeAtPoint(node *AXNode, x, y float64) *AXNode {
	if node == nil {
		return nil
	}

	// If node has bounds, check if point is within them
	if node.Bounds != nil {
		if x < node.Bounds.X || x > node.Bounds.X+node.Bounds.Width ||
			y < node.Bounds.Y || y > node.Bounds.Y+node.Bounds.Height {
			return nil
		}
	}

	// Check children first (more specific elements), in reverse order
	// so we check topmost elements first
	for i := len(node.Children) - 1; i >= 0; i-- {
		if found := findNodeAtPoint(node.Children[i], x, y); found != nil {
			return found
		}
	}

	// No child matched and node has bounds, this node is the target
	if node.Bounds != nil {
		return node
	}

	return nil
}
