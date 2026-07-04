package computeruse

import (
	"context"
	"errors"
	"testing"
)

func TestClassifyComputerUseActions(t *testing.T) {
	tests := []struct {
		name     string
		action   *Action
		wantType ActionType
	}{
		{
			name:     "click action",
			action:   &Action{Type: ActionClick, Target: &Target{Role: "AXButton", Title: "OK"}},
			wantType: ActionClick,
		},
		{
			name:     "type action",
			action:   &Action{Type: ActionTypeText, Value: "hello"},
			wantType: ActionTypeText,
		},
		{
			name:     "scroll action",
			action:   &Action{Type: ActionScroll, Bounds: &Rect{X: 100, Y: 100, Width: 50, Height: 50}},
			wantType: ActionScroll,
		},
		{
			name:     "key press action",
			action:   &Action{Type: ActionKeyPress, Value: "Return"},
			wantType: ActionKeyPress,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.action.Type != tt.wantType {
				t.Errorf("action type = %v, want %v", tt.action.Type, tt.wantType)
			}
		})
	}
}

func TestFindNode(t *testing.T) {
	tree := &AXTree{
		Root: &AXNode{
			Role:  "AXApplication",
			Title: "TestApp",
			Children: []*AXNode{
				{
					Role:  "AXWindow",
					Title: "MainWindow",
					Children: []*AXNode{
						{
							Role:   "AXButton",
							Title:  "OK",
							Bounds: &Rect{X: 100, Y: 200, Width: 80, Height: 30},
						},
						{
							Role:   "AXButton",
							Title:  "Cancel",
							Bounds: &Rect{X: 200, Y: 200, Width: 80, Height: 30},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name      string
		query     *ElementQuery
		wantNil   bool
		wantTitle string
	}{
		{
			name:      "find by role and title",
			query:     &ElementQuery{Role: "AXButton", Title: "OK"},
			wantNil:   false,
			wantTitle: "OK",
		},
		{
			name:      "find by role only",
			query:     &ElementQuery{Role: "AXButton"},
			wantNil:   false,
			wantTitle: "OK",
		},
		{
			name:    "not found",
			query:   &ElementQuery{Role: "AXTextField"},
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := findNode(tree.Root, tt.query)
			if tt.wantNil {
				if node != nil {
					t.Errorf("expected nil, got %v", node)
				}
				return
			}
			if node == nil {
				t.Fatal("expected node, got nil")
			}
			if node.Title != tt.wantTitle {
				t.Errorf("title = %v, want %v", node.Title, tt.wantTitle)
			}
		})
	}
}

func TestFindNodeAtPoint(t *testing.T) {
	tree := &AXTree{
		Root: &AXNode{
			Role:  "AXApplication",
			Title: "TestApp",
			Children: []*AXNode{
				{
					Role:   "AXWindow",
					Title:  "MainWindow",
					Bounds: &Rect{X: 0, Y: 0, Width: 1920, Height: 1080},
					Children: []*AXNode{
						{
							Role:   "AXButton",
							Title:  "OK",
							Bounds: &Rect{X: 100, Y: 200, Width: 80, Height: 30},
						},
					},
				},
			},
		},
	}

	tests := []struct {
		name      string
		x, y      float64
		wantNil   bool
		wantTitle string
	}{
		{
			name:      "point inside button",
			x:         140,
			y:         215,
			wantNil:   false,
			wantTitle: "OK",
		},
		{
			name:    "point outside window",
			x:       2000,
			y:       2000,
			wantNil: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			node := findNodeAtPoint(tree.Root, tt.x, tt.y)
			if tt.wantNil {
				if node != nil {
					t.Errorf("expected nil, got %v", node)
				}
				return
			}
			if node == nil {
				t.Fatal("expected node, got nil")
			}
			if node.Title != tt.wantTitle {
				t.Errorf("title = %v, want %v", node.Title, tt.wantTitle)
			}
		})
	}
}

func TestRouter(t *testing.T) {
	available := &MockBackend{BackendName: "available", Available: true}
	unavailable := &MockBackend{BackendName: "unavailable", Available: false}

	router := NewRouter(unavailable, available)

	ctx := context.Background()

	t.Run("execute on available backend", func(t *testing.T) {
		result, err := Execute(router, ctx, func(b Backend) (string, error) {
			return b.Name(), nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "available" {
			t.Errorf("expected 'available', got %v", result)
		}
	})

	t.Run("skip unavailable backend", func(t *testing.T) {
		result, err := Execute(router, ctx, func(b Backend) (string, error) {
			return b.Name(), nil
		})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if result != "available" {
			t.Errorf("expected 'available', got %v", result)
		}
	})

	t.Run("find backend by capability", func(t *testing.T) {
		b := router.FindBackend(ctx, CapScreenshot)
		if b == nil {
			t.Fatal("expected backend, got nil")
		}
		if b.Name() != "available" {
			t.Errorf("expected 'available', got %v", b.Name())
		}
	})

	t.Run("available backends", func(t *testing.T) {
		backends := router.AvailableBackends(ctx)
		if len(backends) != 1 {
			t.Errorf("expected 1 available backend, got %d", len(backends))
		}
	})
}

func TestComputerUse(t *testing.T) {
	mock := &MockBackend{
		Available: true,
		Screenshot: &Screenshot{
			Image:    []byte("test-png"),
			Width:    1920,
			Height:   1080,
			WindowID: 12345,
		},
	}

	cu := New(mock)
	ctx := context.Background()

	t.Run("capture screen", func(t *testing.T) {
		screenshot, err := cu.CaptureScreen(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if screenshot.Width != 1920 {
			t.Errorf("width = %d, want 1920", screenshot.Width)
		}
	})

	t.Run("get ax tree", func(t *testing.T) {
		tree, err := cu.GetAXTree(ctx)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if tree.Root == nil {
			t.Error("expected root node")
		}
	})

	t.Run("execute action", func(t *testing.T) {
		action := &Action{
			Type:   ActionClick,
			Target: &Target{Role: "AXButton", Title: "OK"},
		}
		result, err := cu.Execute(ctx, action)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !result.Success {
			t.Error("expected success")
		}
		if mock.ExecuteCallCount != 1 {
			t.Errorf("expected 1 execute call, got %d", mock.ExecuteCallCount)
		}
	})

	t.Run("find element", func(t *testing.T) {
		node, err := cu.FindElement(ctx, &ElementQuery{Role: "AXButton", Title: "OK"})
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if node == nil {
			t.Fatal("expected node, got nil")
		}
		if node.Title != "OK" {
			t.Errorf("title = %v, want OK", node.Title)
		}
	})

	t.Run("get element at point", func(t *testing.T) {
		node, err := cu.GetElementAtPoint(ctx, 140, 215)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if node == nil {
			t.Fatal("expected node, got nil")
		}
		if node.Title != "OK" {
			t.Errorf("title = %v, want OK", node.Title)
		}
	})
}

func TestNoopBackend(t *testing.T) {
	noop := &NoopBackend{}
	ctx := context.Background()

	if noop.IsAvailable(ctx) {
		t.Error("expected noop backend to be unavailable")
	}

	_, err := noop.CaptureScreen(ctx)
	if !errors.Is(err, ErrNoBackend) {
		t.Errorf("expected ErrNoBackend, got %v", err)
	}

	_, err = noop.GetAXTree(ctx)
	if !errors.Is(err, ErrNoBackend) {
		t.Errorf("expected ErrNoBackend, got %v", err)
	}

	result, err := noop.Execute(ctx, &Action{Type: ActionClick})
	if !errors.Is(err, ErrNoBackend) {
		t.Errorf("expected ErrNoBackend, got %v", err)
	}
	if result.Success {
		t.Error("expected failure")
	}
}
