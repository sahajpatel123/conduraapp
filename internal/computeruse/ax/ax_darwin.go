// Package ax provides macOS Accessibility API bindings for reading
// the AX tree.
//
// This package uses CGo to access the ApplicationServices framework.
// It requires the Accessibility permission to be granted to the application.
//
// Build: CGO_ENABLED=1 go build -tags darwin
// Runtime: System Preferences > Privacy & Accessibility > add your app
package ax

/*
#cgo LDFLAGS: -framework ApplicationServices
#include <ApplicationServices/ApplicationServices.h>

// Helper to check if accessibility is trusted
Boolean isAccessibilityTrusted() {
    return AXIsProcessTrusted();
}

// Helper to get focused application
AXUIElementRef getFocusedApplication() {
    AXUIElementRef systemWide = AXUIElementCreateSystemWide();
    if (!systemWide) return NULL;

    CFTypeRef focusedApp = NULL;
    AXError err = AXUIElementCopyAttributeValue(
        systemWide,
        kAXFocusedApplicationAttribute,
        &focusedApp
    );
    CFRelease(systemWide);

    if (err != kAXErrorSuccess || !focusedApp) return NULL;
    return (AXUIElementRef)focusedApp;
}

// Helper to get attribute value
AXError getAttributeValue(AXUIElementRef element, CFStringRef attribute, CFTypeRef *value) {
    return AXUIElementCopyAttributeValue(element, attribute, value);
}

// Helper to get string attribute
const char* getStringAttribute(AXUIElementRef element, CFStringRef attribute) {
    CFTypeRef value = NULL;
    AXError err = AXUIElementCopyAttributeValue(element, attribute, &value);
    if (err != kAXErrorSuccess || !value) return NULL;

    if (CFGetTypeID(value) != CFStringGetTypeID()) {
        CFRelease(value);
        return NULL;
    }

    CFStringRef str = (CFStringRef)value;
    CFIndex length = CFStringGetLength(str);
    CFIndex maxSize = CFStringGetMaximumSizeForEncoding(length, kCFStringEncodingUTF8) + 1;
    char *buffer = (char *)malloc(maxSize);
    if (!buffer) {
        CFRelease(value);
        return NULL;
    }

    if (!CFStringGetCString(str, buffer, maxSize, kCFStringEncodingUTF8)) {
        free(buffer);
        CFRelease(value);
        return NULL;
    }

    CFRelease(value);
    return buffer;
}

// Helper to get child at index (deprecated — use getChildren for bulk access)
AXUIElementRef getChildAtIndex(AXUIElementRef element, CFIndex index) {
    CFTypeRef children = NULL;
    AXError err = AXUIElementCopyAttributeValue(element, kAXChildrenAttribute, &children);
    if (err != kAXErrorSuccess || !children) return NULL;

    CFIndex count = CFArrayGetCount((CFArrayRef)children);
    if (index >= count) {
        CFRelease(children);
        return NULL;
    }

    AXUIElementRef child = (AXUIElementRef)CFArrayGetValueAtIndex((CFArrayRef)children, index);
    CFRetain(child);
    CFRelease(children);
    return child;
}

// getChildren fetches the children array once. Follows the Create rule:
// the returned CFArrayRef has retain count 1; the caller must CFRelease
// it when done. Returns NULL on error.
CFArrayRef getChildren(AXUIElementRef element) {
    CFTypeRef children = NULL;
    AXError err = AXUIElementCopyAttributeValue(element, kAXChildrenAttribute, &children);
    if (err != kAXErrorSuccess || !children) return NULL;
    return (CFArrayRef)children;
}

// Helper to get position
int getPosition(AXUIElementRef element, double *x, double *y) {
    CFTypeRef position = NULL;
    AXError err = AXUIElementCopyAttributeValue(element, kAXPositionAttribute, &position);
    if (err != kAXErrorSuccess || !position) return 0;

    CGPoint point;
    if (!AXValueGetValue((AXValueRef)position, kAXValueCGPointType, &point)) {
        CFRelease(position);
        return 0;
    }

    *x = point.x;
    *y = point.y;
    CFRelease(position);
    return 1;
}

// Helper to get size
int getSize(AXUIElementRef element, double *width, double *height) {
    CFTypeRef size = NULL;
    AXError err = AXUIElementCopyAttributeValue(element, kAXSizeAttribute, &size);
    if (err != kAXErrorSuccess || !size) return 0;

    CGSize sz;
    if (!AXValueGetValue((AXValueRef)size, kAXValueCGSizeType, &sz)) {
        CFRelease(size);
        return 0;
    }

    *width = sz.width;
    *height = sz.height;
    CFRelease(size);
    return 1;
}

// Helper to get PID
int getPid(AXUIElementRef element, pid_t *pid) {
    return AXUIElementGetPid(element, pid) == kAXErrorSuccess;
}
*/
import "C"
import (
	"context"
	"fmt"
	"time"
	"unsafe"

	"github.com/sahajpatel123/conduraapp/internal/computeruse"
)

// Backend implements computeruse.Backend using the macOS Accessibility API.
type Backend struct{}

// New creates a new macOS Accessibility backend.
func New() *Backend {
	return &Backend{}
}

// Name returns the backend identifier.
func (b *Backend) Name() string { return "ax-darwin" }

// Capabilities returns the supported capabilities.
func (b *Backend) Capabilities() []computeruse.Capability {
	return []computeruse.Capability{
		computeruse.CapAXTree,
	}
}

// IsAvailable checks if the Accessibility API is available and has permission.
func (b *Backend) IsAvailable(_ context.Context) bool {
	return C.isAccessibilityTrusted() != 0
}

// CaptureScreen is not implemented in this backend.
func (b *Backend) CaptureScreen(_ context.Context) (*computeruse.Screenshot, error) {
	return nil, computeruse.ErrUnsupportedAction
}

// GetAXTree reads the accessibility tree from the focused application.
func (b *Backend) GetAXTree(_ context.Context) (*computeruse.AXTree, error) {
	// Get the focused application
	app := C.getFocusedApplication()
	if app == 0 {
		return nil, fmt.Errorf("ax: no focused application")
	}
	defer C.CFRelease(C.CFTypeRef(app))

	// Get PID
	var pid C.pid_t
	if C.getPid(app, &pid) == 0 {
		return nil, fmt.Errorf("ax: failed to get PID")
	}

	// Build the AX tree
	root, err := b.buildNode(app, 0)
	if err != nil {
		return nil, err
	}

	return &computeruse.AXTree{
		Root:      root,
		Timestamp: time.Now(),
		PID:       int32(pid),
	}, nil
}

const maxDepth = 50

// buildNode recursively builds an AXNode from an AXUIElement.
func (b *Backend) buildNode(element C.AXUIElementRef, depth int) (*computeruse.AXNode, error) {
	if depth > maxDepth {
		return nil, fmt.Errorf("ax: maximum depth exceeded")
	}

	node := &computeruse.AXNode{
		Attributes: make(map[string]interface{}),
	}

	// Get role
	role := C.getStringAttribute(element, C.kAXRoleAttribute)
	if role != nil {
		node.Role = C.GoString(role)
		C.free(unsafe.Pointer(role))
	}

	// Get title
	title := C.getStringAttribute(element, C.kAXTitleAttribute)
	if title != nil {
		node.Title = C.GoString(title)
		C.free(unsafe.Pointer(title))
	}

	// Get value
	value := C.getStringAttribute(element, C.kAXValueAttribute)
	if value != nil {
		node.Value = C.GoString(value)
		C.free(unsafe.Pointer(value))
	}

	// Get description
	desc := C.getStringAttribute(element, C.kAXDescriptionAttribute)
	if desc != nil {
		node.Description = C.GoString(desc)
		C.free(unsafe.Pointer(desc))
	}

	// Get bounds
	var x, y, width, height C.double
	if C.getPosition(element, &x, &y) != 0 && C.getSize(element, &width, &height) != 0 {
		node.Bounds = &computeruse.Rect{
			X:      float64(x),
			Y:      float64(y),
			Width:  float64(width),
			Height: float64(height),
		}
	}

	// Get children — fetch once, iterate (O(N) instead of O(N²))
	childrenArr := C.getChildren(element)
	if childrenArr != 0 {
		count := int(C.CFArrayGetCount(childrenArr))
		if count > 0 {
			node.Children = make([]*computeruse.AXNode, 0, count)
			for i := 0; i < count; i++ {
				child := C.CFArrayGetValueAtIndex(childrenArr, C.CFIndex(i))
				if child == nil {
					continue
				}
				// CFRetain for the child before buildNode, which may
				// recurse; CFRelease after.
				C.CFRetain(C.CFTypeRef(child))
				childNode, err := b.buildNode(C.AXUIElementRef(child), depth+1)
				C.CFRelease(C.CFTypeRef(child))
				if err != nil {
					continue
				}
				node.Children = append(node.Children, childNode)
			}
		}
		C.CFRelease(C.CFTypeRef(childrenArr))
	}

	return node, nil
}

// Execute performs a computer-use action (not implemented in AX backend).
func (b *Backend) Execute(_ context.Context, action *computeruse.Action) (*computeruse.ActionResult, error) {
	return &computeruse.ActionResult{
		Success: false,
		Error:   computeruse.ErrUnsupportedAction,
		Action:  action,
	}, computeruse.ErrUnsupportedAction
}
