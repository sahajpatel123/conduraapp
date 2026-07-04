package computeruse

import (
	"context"
	"fmt"
	"time"
)

// VerificationResult is the result of comparing pre and post action snapshots.
type VerificationResult struct {
	Valid   bool     // Whether the verification passed
	Reason  string   // Human-readable reason for failure
	Diffs   []AXDiff // List of differences found
	Aborted bool     // Whether the action was aborted
}

// AXDiff describes a difference between two AX trees.
type AXDiff struct {
	Type     DiffType // The type of difference
	Path     string   // Path to the differing node
	Expected string   // Expected value (pre-action)
	Actual   string   // Actual value (post-action)
}

// DiffType describes the type of AX tree difference.
type DiffType string

// Difference types for AX tree comparison.
const (
	DiffRoleChanged   DiffType = "role_changed"   // Role attribute changed
	DiffTitleChanged  DiffType = "title_changed"  // Title attribute changed
	DiffValueChanged  DiffType = "value_changed"  // Value attribute changed
	DiffBoundsChanged DiffType = "bounds_changed" // Bounds changed
	DiffNodeAdded     DiffType = "node_added"     // New node appeared
	DiffNodeRemoved   DiffType = "node_removed"   // Node disappeared
	DiffWindowFocused DiffType = "window_focused" // Window focus changed
	DiffWindowChanged DiffType = "window_changed" // Window changed
)

// VerifySnapshots compares pre and post action snapshots to detect
// stale state. This implements the twin-snapshot verification from
// MISSION §5.2.
//
// The verification checks:
// 1. Window focus hasn't changed unexpectedly
// 2. AX tree structure hasn't changed in unexpected ways
// 3. Target element still exists and is in the expected state
func VerifySnapshots(pre, post *Snapshot, action *Action) *VerificationResult {
	result := &VerificationResult{
		Valid: true,
		Diffs: make([]AXDiff, 0),
	}

	if pre == nil || post == nil {
		result.Valid = false
		result.Reason = "missing snapshot"
		return result
	}

	// Check if window focus changed
	if pre.WindowID != 0 && post.WindowID != 0 && pre.WindowID != post.WindowID {
		result.Valid = false
		result.Reason = "window focus changed during action"
		result.Aborted = true
		result.Diffs = append(result.Diffs, AXDiff{
			Type:     DiffWindowFocused,
			Path:     "window",
			Expected: fmt.Sprintf("%d", pre.WindowID),
			Actual:   fmt.Sprintf("%d", post.WindowID),
		})
		return result
	}

	// Check if AX trees are available
	if pre.AXTree == nil || post.AXTree == nil {
		// Can't verify without AX trees, allow action
		return result
	}

	// Compare AX trees
	diffs := compareAXTrees(pre.AXTree.Root, post.AXTree.Root, "")
	if len(diffs) > 0 {
		result.Diffs = diffs

		// Check if any diffs are critical (window change, node removal)
		for _, diff := range diffs {
			if diff.Type == DiffWindowChanged || diff.Type == DiffNodeRemoved {
				result.Valid = false
				result.Reason = fmt.Sprintf("critical change detected: %s at %s", diff.Type, diff.Path)
				result.Aborted = true
				return result
			}
		}

		// Non-critical diffs still make it invalid but not aborted
		if len(diffs) > 0 {
			result.Valid = false
			result.Reason = fmt.Sprintf("%d differences detected", len(diffs))
		}
	}

	return result
}

// compareAXTrees recursively compares two AX trees and returns differences.
func compareAXTrees(pre, post *AXNode, path string) []AXDiff {
	if pre == nil && post == nil {
		return nil
	}
	if pre == nil {
		return []AXDiff{{
			Type: DiffNodeAdded,
			Path: path,
		}}
	}
	if post == nil {
		return []AXDiff{{
			Type: DiffNodeRemoved,
			Path: path,
		}}
	}

	var diffs []AXDiff

	// Compare role
	if pre.Role != post.Role {
		diffs = append(diffs, AXDiff{
			Type:     DiffRoleChanged,
			Path:     path,
			Expected: pre.Role,
			Actual:   post.Role,
		})
	}

	// Compare title
	if pre.Title != post.Title {
		diffs = append(diffs, AXDiff{
			Type:     DiffTitleChanged,
			Path:     path,
			Expected: pre.Title,
			Actual:   post.Title,
		})
	}

	// Compare value
	if pre.Value != post.Value {
		diffs = append(diffs, AXDiff{
			Type:     DiffValueChanged,
			Path:     path,
			Expected: pre.Value,
			Actual:   post.Value,
		})
	}

	// Compare bounds
	if pre.Bounds != nil && post.Bounds != nil {
		if pre.Bounds.X != post.Bounds.X || pre.Bounds.Y != post.Bounds.Y ||
			pre.Bounds.Width != post.Bounds.Width || pre.Bounds.Height != post.Bounds.Height {
			diffs = append(diffs, AXDiff{
				Type:     DiffBoundsChanged,
				Path:     path,
				Expected: fmt.Sprintf("(%.0f,%.0f,%.0f,%.0f)", pre.Bounds.X, pre.Bounds.Y, pre.Bounds.Width, pre.Bounds.Height),
				Actual:   fmt.Sprintf("(%.0f,%.0f,%.0f,%.0f)", post.Bounds.X, post.Bounds.Y, post.Bounds.Width, post.Bounds.Height),
			})
		}
	}

	// Compare children
	childDiffs := compareChildren(pre.Children, post.Children, path)
	diffs = append(diffs, childDiffs...)

	return diffs
}

// compareChildren compares two slices of child nodes.
func compareChildren(pre, post []*AXNode, parentPath string) []AXDiff {
	var diffs []AXDiff

	// Check for removed children
	for i, child := range pre {
		childPath := fmt.Sprintf("%s/child[%d]", parentPath, i)
		if i >= len(post) {
			diffs = append(diffs, AXDiff{
				Type:     DiffNodeRemoved,
				Path:     childPath,
				Expected: child.Role + ":" + child.Title,
			})
			continue
		}
		childDiffs := compareAXTrees(child, post[i], childPath)
		diffs = append(diffs, childDiffs...)
	}

	// Check for added children
	for i := len(pre); i < len(post); i++ {
		childPath := fmt.Sprintf("%s/child[%d]", parentPath, i)
		diffs = append(diffs, AXDiff{
			Type:   DiffNodeAdded,
			Path:   childPath,
			Actual: post[i].Role + ":" + post[i].Title,
		})
	}

	return diffs
}

// Snapshot represents a point-in-time capture of the screen state.
type Snapshot struct {
	Timestamp time.Time
	WindowID  uint32
	AXTree    *AXTree
}

// NewSnapshot creates a new snapshot from an AX tree.
func NewSnapshot(tree *AXTree, windowID uint32) *Snapshot {
	return &Snapshot{
		Timestamp: time.Now(),
		WindowID:  windowID,
		AXTree:    tree,
	}
}

// VerifyActionCapture captures a pre-action snapshot for twin-snapshot verification.
// Call this before executing an action, then call VerifyActionVerify after.
//
// Usage:
//
//	pre := VerifyActionCapture(ctx, backend)
//	result, err := backend.Execute(ctx, action)
//	if err != nil { ... }
//	verification := VerifyActionVerify(pre, backend, action)
//	if !verification.Valid { abort }
func VerifyActionCapture(ctx context.Context, backend Backend) *Snapshot {
	tree, err := backend.GetAXTree(ctx)
	if err != nil {
		return nil
	}
	return NewSnapshot(tree, tree.WindowID)
}

// VerifyActionVerify captures a post-action snapshot and verifies it
// against the pre-action snapshot.
func VerifyActionVerify(pre *Snapshot, backend Backend, action *Action) *VerificationResult {
	if pre == nil {
		return &VerificationResult{
			Valid:   false,
			Reason:  "no pre-action snapshot",
			Aborted: true,
		}
	}

	ctx := context.Background()
	post := VerifyActionCapture(ctx, backend)
	if post == nil {
		return &VerificationResult{
			Valid:   false,
			Reason:  "failed to capture post-action snapshot",
			Aborted: true,
		}
	}

	return VerifySnapshots(pre, post, action)
}
