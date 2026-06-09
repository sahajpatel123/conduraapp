package computeruse

import (
	"context"
	"time"
)

// NoopBackend is a fallback backend that returns errors for all operations.
// It's used when no real backend is available.
type NoopBackend struct{}

// Name returns the backend identifier.
func (n *NoopBackend) Name() string { return "noop" }

// Capabilities returns nil since this backend supports no operations.
func (n *NoopBackend) Capabilities() []Capability { return nil }

// CaptureScreen returns ErrNoBackend.
func (n *NoopBackend) CaptureScreen(_ context.Context) (*Screenshot, error) {
	return nil, ErrNoBackend
}

// GetAXTree returns ErrNoBackend.
func (n *NoopBackend) GetAXTree(_ context.Context) (*AXTree, error) {
	return nil, ErrNoBackend
}

// Execute returns ErrNoBackend.
func (n *NoopBackend) Execute(_ context.Context, action *Action) (*ActionResult, error) {
	return &ActionResult{
		Success:  false,
		Error:    ErrNoBackend,
		Duration: 0,
		Action:   action,
	}, ErrNoBackend
}

// IsAvailable returns false since this backend is never available.
func (n *NoopBackend) IsAvailable(_ context.Context) bool { return false }

const (
	mockWidth  = 1920
	mockHeight = 1080
	mockBtnX   = 100
	mockBtnY   = 200
	mockBtnW   = 80
	mockBtnH   = 30
)

// MockBackend is a test backend that can be configured to return specific results.
type MockBackend struct {
	BackendName      string
	Caps             []Capability
	Available        bool
	Screenshot       *Screenshot
	AXTree           *AXTree
	ExecuteResult    *ActionResult
	ExecuteErr       error
	ScreenshotErr    error
	AXTreeErr        error
	ExecuteCallCount int
}

// Name returns the backend identifier.
func (m *MockBackend) Name() string {
	if m.BackendName != "" {
		return m.BackendName
	}
	return "mock"
}

// Capabilities returns the supported capabilities.
func (m *MockBackend) Capabilities() []Capability {
	if m.Caps != nil {
		return m.Caps
	}
	return []Capability{
		CapScreenshot,
		CapAXTree,
		CapClick,
		CapType,
		CapScroll,
		CapKeyPress,
	}
}

// CaptureScreen returns a mock screenshot.
func (m *MockBackend) CaptureScreen(_ context.Context) (*Screenshot, error) {
	if m.ScreenshotErr != nil {
		return nil, m.ScreenshotErr
	}
	if m.Screenshot != nil {
		return m.Screenshot, nil
	}
	return &Screenshot{
		Image:     []byte("mock-png"),
		Width:     mockWidth,
		Height:    mockHeight,
		Timestamp: time.Now(),
	}, nil
}

// GetAXTree returns a mock accessibility tree.
func (m *MockBackend) GetAXTree(_ context.Context) (*AXTree, error) {
	if m.AXTreeErr != nil {
		return nil, m.AXTreeErr
	}
	if m.AXTree != nil {
		return m.AXTree, nil
	}
	return &AXTree{
		Root: &AXNode{
			Role:  "AXApplication",
			Title: "MockApp",
			Children: []*AXNode{
				{
					Role:  "AXWindow",
					Title: "MainWindow",
					Bounds: &Rect{
						X: 0, Y: 0, Width: mockWidth, Height: mockHeight,
					},
					Children: []*AXNode{
						{
							Role:  "AXButton",
							Title: "OK",
							Bounds: &Rect{
								X: mockBtnX, Y: mockBtnY, Width: mockBtnW, Height: mockBtnH,
							},
						},
					},
				},
			},
		},
		Timestamp: time.Now(),
	}, nil
}

// Execute performs a mock action.
func (m *MockBackend) Execute(_ context.Context, action *Action) (*ActionResult, error) {
	m.ExecuteCallCount++
	if m.ExecuteErr != nil {
		return &ActionResult{
			Success:  false,
			Error:    m.ExecuteErr,
			Duration: time.Millisecond,
			Action:   action,
		}, m.ExecuteErr
	}
	if m.ExecuteResult != nil {
		return m.ExecuteResult, nil
	}
	return &ActionResult{
		Success:  true,
		Duration: time.Millisecond,
		Action:   action,
	}, nil
}

// IsAvailable returns whether the backend is available.
func (m *MockBackend) IsAvailable(_ context.Context) bool {
	return m.Available
}
