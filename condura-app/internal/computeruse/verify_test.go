package computeruse

import (
	"context"
	"testing"
)

func TestVerifySnapshots(t *testing.T) {
	tests := []struct {
		name      string
		pre       *Snapshot
		post      *Snapshot
		action    *Action
		wantValid bool
		wantAbort bool
	}{
		{
			name: "identical snapshots",
			pre: &Snapshot{
				WindowID: 123,
				AXTree: &AXTree{
					Root: &AXNode{
						Role:  "AXButton",
						Title: "OK",
					},
				},
			},
			post: &Snapshot{
				WindowID: 123,
				AXTree: &AXTree{
					Root: &AXNode{
						Role:  "AXButton",
						Title: "OK",
					},
				},
			},
			action:    &Action{Type: ActionClick},
			wantValid: true,
			wantAbort: false,
		},
		{
			name: "window focus changed",
			pre: &Snapshot{
				WindowID: 123,
				AXTree: &AXTree{
					Root: &AXNode{Role: "AXButton"},
				},
			},
			post: &Snapshot{
				WindowID: 456,
				AXTree: &AXTree{
					Root: &AXNode{Role: "AXButton"},
				},
			},
			action:    &Action{Type: ActionClick},
			wantValid: false,
			wantAbort: true,
		},
		{
			name: "role changed",
			pre: &Snapshot{
				WindowID: 123,
				AXTree: &AXTree{
					Root: &AXNode{
						Role:  "AXButton",
						Title: "OK",
					},
				},
			},
			post: &Snapshot{
				WindowID: 123,
				AXTree: &AXTree{
					Root: &AXNode{
						Role:  "AXTextField",
						Title: "OK",
					},
				},
			},
			action:    &Action{Type: ActionClick},
			wantValid: false,
			wantAbort: false,
		},
		{
			name: "title changed",
			pre: &Snapshot{
				WindowID: 123,
				AXTree: &AXTree{
					Root: &AXNode{
						Role:  "AXButton",
						Title: "OK",
					},
				},
			},
			post: &Snapshot{
				WindowID: 123,
				AXTree: &AXTree{
					Root: &AXNode{
						Role:  "AXButton",
						Title: "Cancel",
					},
				},
			},
			action:    &Action{Type: ActionClick},
			wantValid: false,
			wantAbort: false,
		},
		{
			name: "node removed",
			pre: &Snapshot{
				WindowID: 123,
				AXTree: &AXTree{
					Root: &AXNode{
						Role: "AXWindow",
						Children: []*AXNode{
							{Role: "AXButton", Title: "OK"},
						},
					},
				},
			},
			post: &Snapshot{
				WindowID: 123,
				AXTree: &AXTree{
					Root: &AXNode{
						Role:     "AXWindow",
						Children: []*AXNode{},
					},
				},
			},
			action:    &Action{Type: ActionClick},
			wantValid: false,
			wantAbort: true,
		},
		{
			name: "node added",
			pre: &Snapshot{
				WindowID: 123,
				AXTree: &AXTree{
					Root: &AXNode{
						Role:     "AXWindow",
						Children: []*AXNode{},
					},
				},
			},
			post: &Snapshot{
				WindowID: 123,
				AXTree: &AXTree{
					Root: &AXNode{
						Role: "AXWindow",
						Children: []*AXNode{
							{Role: "AXButton", Title: "OK"},
						},
					},
				},
			},
			action:    &Action{Type: ActionClick},
			wantValid: false,
			wantAbort: false,
		},
		{
			name: "nil pre snapshot",
			pre:  nil,
			post: &Snapshot{
				WindowID: 123,
			},
			action:    &Action{Type: ActionClick},
			wantValid: false,
			wantAbort: false,
		},
		{
			name: "nil post snapshot",
			pre: &Snapshot{
				WindowID: 123,
			},
			post:      nil,
			action:    &Action{Type: ActionClick},
			wantValid: false,
			wantAbort: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := VerifySnapshots(tt.pre, tt.post, tt.action)
			if result.Valid != tt.wantValid {
				t.Errorf("Valid = %v, want %v", result.Valid, tt.wantValid)
			}
			if result.Aborted != tt.wantAbort {
				t.Errorf("Aborted = %v, want %v", result.Aborted, tt.wantAbort)
			}
		})
	}
}

func TestCompareAXTrees(t *testing.T) {
	tests := []struct {
		name      string
		pre       *AXNode
		post      *AXNode
		wantDiffs int
	}{
		{
			name: "identical nodes",
			pre: &AXNode{
				Role:  "AXButton",
				Title: "OK",
			},
			post: &AXNode{
				Role:  "AXButton",
				Title: "OK",
			},
			wantDiffs: 0,
		},
		{
			name: "role changed",
			pre: &AXNode{
				Role:  "AXButton",
				Title: "OK",
			},
			post: &AXNode{
				Role:  "AXTextField",
				Title: "OK",
			},
			wantDiffs: 1,
		},
		{
			name: "title changed",
			pre: &AXNode{
				Role:  "AXButton",
				Title: "OK",
			},
			post: &AXNode{
				Role:  "AXButton",
				Title: "Cancel",
			},
			wantDiffs: 1,
		},
		{
			name: "multiple changes",
			pre: &AXNode{
				Role:  "AXButton",
				Title: "OK",
				Value: "original",
			},
			post: &AXNode{
				Role:  "AXTextField",
				Title: "Cancel",
				Value: "changed",
			},
			wantDiffs: 3,
		},
		{
			name: "child added",
			pre: &AXNode{
				Role:     "AXWindow",
				Children: []*AXNode{},
			},
			post: &AXNode{
				Role: "AXWindow",
				Children: []*AXNode{
					{Role: "AXButton", Title: "OK"},
				},
			},
			wantDiffs: 1,
		},
		{
			name: "child removed",
			pre: &AXNode{
				Role: "AXWindow",
				Children: []*AXNode{
					{Role: "AXButton", Title: "OK"},
				},
			},
			post: &AXNode{
				Role:     "AXWindow",
				Children: []*AXNode{},
			},
			wantDiffs: 1,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diffs := compareAXTrees(tt.pre, tt.post, "")
			if len(diffs) != tt.wantDiffs {
				t.Errorf("got %d diffs, want %d", len(diffs), tt.wantDiffs)
			}
		})
	}
}

func TestNewSnapshot(t *testing.T) {
	tree := &AXTree{
		Root: &AXNode{
			Role:  "AXButton",
			Title: "OK",
		},
	}

	snapshot := NewSnapshot(tree, 123)
	if snapshot == nil {
		t.Fatal("expected snapshot, got nil")
	}
	if snapshot.WindowID != 123 {
		t.Errorf("WindowID = %d, want 123", snapshot.WindowID)
	}
	if snapshot.AXTree != tree {
		t.Error("AXTree mismatch")
	}
	if snapshot.Timestamp.IsZero() {
		t.Error("expected non-zero timestamp")
	}
}

func TestVerifyActionCapture(t *testing.T) {
	mock := &MockBackend{
		Available: true,
	}

	ctx := context.TODO()
	snapshot := VerifyActionCapture(ctx, mock)
	if snapshot == nil {
		t.Fatal("expected snapshot, got nil")
	}
	if snapshot.AXTree == nil {
		t.Error("expected AXTree in snapshot")
	}
}

func TestVerifyActionVerify(t *testing.T) {
	mock := &MockBackend{
		Available: true,
	}

	ctx := context.TODO()
	pre := VerifyActionCapture(ctx, mock)
	if pre == nil {
		t.Fatal("expected pre snapshot, got nil")
	}

	result := VerifyActionVerify(pre, mock, &Action{Type: ActionClick})
	if result == nil {
		t.Fatal("expected result, got nil")
	}
	if !result.Valid {
		t.Errorf("expected valid, got invalid: %s", result.Reason)
	}
}
