//go:build darwin && cgo

package backends

/*
#cgo LDFLAGS: -framework ApplicationServices -framework CoreGraphics -framework Foundation

#include <ApplicationServices/ApplicationServices.h>
#include <CoreGraphics/CoreGraphics.h>
#include <stdint.h>

// Accessibility: is trusted
Boolean orax_isTrusted() {
	return AXIsProcessTrusted();
}

// Accessibility: get focused application
AXUIElementRef orax_getFocusedApp() {
	AXUIElementRef systemWide = AXUIElementCreateSystemWide();
	if (!systemWide) return NULL;
	CFTypeRef app = NULL;
	AXError err = AXUIElementCopyAttributeValue(
		systemWide, kAXFocusedApplicationAttribute, &app);
	CFRelease(systemWide);
	if (err != kAXErrorSuccess || !app) return NULL;
	return (AXUIElementRef)app;
}

// Accessibility: get string attribute
char* orax_getStringAttr(AXUIElementRef el, CFStringRef attr) {
	CFTypeRef val = NULL;
	if (AXUIElementCopyAttributeValue(el, attr, &val) != kAXErrorSuccess || !val) return NULL;
	if (CFGetTypeID(val) != CFStringGetTypeID()) { CFRelease(val); return NULL; }
	CFStringRef s = (CFStringRef)val;
	CFIndex len = CFStringGetLength(s);
	CFIndex max = CFStringGetMaximumSizeForEncoding(len, kCFStringEncodingUTF8) + 1;
	char* buf = (char*)malloc(max);
	if (!buf || !CFStringGetCString(s, buf, max, kCFStringEncodingUTF8)) { free(buf); CFRelease(val); return NULL; }
	CFRelease(val);
	return buf;
}

// Accessibility: get size
int orax_getSize(AXUIElementRef el, double* w, double* h) {
	CFTypeRef val = NULL;
	if (AXUIElementCopyAttributeValue(el, kAXSizeAttribute, &val) != kAXErrorSuccess || !val) return 0;
	CGSize sz;
	int ok = AXValueGetValue((AXValueRef)val, kAXValueCGSizeType, &sz);
	CFRelease(val);
	if (!ok) return 0;
	*w = sz.width; *h = sz.height;
	return 1;
}

// Accessibility: get position
int orax_getPos(AXUIElementRef el, double* x, double* y) {
	CFTypeRef val = NULL;
	if (AXUIElementCopyAttributeValue(el, kAXPositionAttribute, &val) != kAXErrorSuccess || !val) return 0;
	CGPoint pt;
	int ok = AXValueGetValue((AXValueRef)val, kAXValueCGPointType, &pt);
	CFRelease(val);
	if (!ok) return 0;
	*x = pt.x; *y = pt.y;
	return 1;
}

// Accessibility: get children array (caller must CFRelease)
CFArrayRef orax_getChildren(AXUIElementRef el) {
	CFTypeRef val = NULL;
	if (AXUIElementCopyAttributeValue(el, kAXChildrenAttribute, &val) != kAXErrorSuccess || !val) return NULL;
	return (CFArrayRef)val;
}

// Accessibility: get PID
int orax_getPid(AXUIElementRef el, pid_t* pid) {
	return AXUIElementGetPid(el, pid) == kAXErrorSuccess;
}

// Accessibility: set string value (for text fields)
int orax_setStringValue(AXUIElementRef el, const char* text) {
	CFStringRef s = CFStringCreateWithCString(NULL, text, kCFStringEncodingUTF8);
	if (!s) return 0;
	AXError err = AXUIElementSetAttributeValue(el, kAXValueAttribute, s);
	CFRelease(s);
	return err == kAXErrorSuccess;
}

// Accessibility: press action
int orax_press(AXUIElementRef el) {
	return AXUIElementPerformAction(el, kAXPressAction) == kAXErrorSuccess;
}

// Accessibility: confirm action
int orax_confirm(AXUIElementRef el) {
	return AXUIElementPerformAction(el, kAXConfirmAction) == kAXErrorSuccess;
}

// Screen capture: uses the Go-level screencapture(1) fork+exec
// in the Go code below. C-level helpers are not needed.
// (orax_captureDisplay and orax_imageToPNG are unused; captureScreen
// calls screencapture(1) via exec.CommandContext.)

int orax_clickAt(double x, double y) {
	CGEventRef down = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseDown, CGPointMake(x, y), kCGMouseButtonLeft);
	CGEventRef up   = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseUp,   CGPointMake(x, y), kCGMouseButtonLeft);
	if (!down || !up) {
		if (down) CFRelease(down);
		if (up)   CFRelease(up);
		return 0;
	}
	CGEventPost(kCGHIDEventTap, down);
	usleep(10000); // 10ms between down/up
	CGEventPost(kCGHIDEventTap, up);
	CFRelease(down);
	CFRelease(up);
	return 1;
}

int orax_doubleClickAt(double x, double y) {
	CGEventRef down = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseDown, CGPointMake(x, y), kCGMouseButtonLeft);
	CGEventRef up   = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseUp,   CGPointMake(x, y), kCGMouseButtonLeft);
	if (!down || !up) {
		if (down) CFRelease(down);
		if (up)   CFRelease(up);
		return 0;
	}
	CGEventSetIntegerValueField(down, kCGMouseEventClickState, 2);
	CGEventSetIntegerValueField(up,   kCGMouseEventClickState, 2);
	CGEventPost(kCGHIDEventTap, down);
	usleep(10000);
	CGEventPost(kCGHIDEventTap, up);
	CFRelease(down);
	CFRelease(up);
	return 1;
}

// Keyboard: send keystrokes
int orax_typeString(const char* text) {
	CFStringRef s = CFStringCreateWithCString(NULL, text, kCFStringEncodingUTF8);
	if (!s) return 0;
	CGEventRef ev = CGEventCreateKeyboardEvent(NULL, 0, true);
	if (!ev) { CFRelease(s); return 0; }
	CGEventPost(kCGHIDEventTap, ev);
	usleep(10000);
	CFRelease(ev);

	// Paste the string via the pasteboard (more reliable for Unicode)
	CGEventRef cmdV = CGEventCreateKeyboardEvent(NULL, (CGKeyCode)9, true);  // V key
	CGEventRef cmdVDown = cmdV;
	if (!cmdVDown) { CFRelease(s); return 0; }
	CGEventSetFlags(cmdVDown, kCGEventFlagMaskCommand);
	CGEventRef cmdVUp = CGEventCreateKeyboardEvent(NULL, (CGKeyCode)9, false);
	CGEventSetFlags(cmdVUp, kCGEventFlagMaskCommand);

	CGEventPost(kCGHIDEventTap, cmdVDown);
	usleep(10000);
	CGEventPost(kCGHIDEventTap, cmdVUp);
	CFRelease(cmdVDown);
	CFRelease(cmdVUp);
	CFRelease(s);
	return 1;
}

int orax_typeChar(CGKeyCode keyCode) {
	CGEventRef down = CGEventCreateKeyboardEvent(NULL, keyCode, true);
	CGEventRef up   = CGEventCreateKeyboardEvent(NULL, keyCode, false);
	if (!down || !up) {
		if (down) CFRelease(down);
		if (up)   CFRelease(up);
		return 0;
	}
	CGEventPost(kCGHIDEventTap, down);
	usleep(5000);
	CGEventPost(kCGHIDEventTap, up);
	CFRelease(down);
	CFRelease(up);
	return 1;
}

int orax_typeModKey(CGKeyCode keyCode, CGEventFlags flags) {
	CGEventRef down = CGEventCreateKeyboardEvent(NULL, keyCode, true);
	CGEventRef up   = CGEventCreateKeyboardEvent(NULL, keyCode, false);
	if (!down || !up) {
		if (down) CFRelease(down);
		if (up)   CFRelease(up);
		return 0;
	}
	CGEventSetFlags(down, flags);
	CGEventSetFlags(up,   flags);
	CGEventPost(kCGHIDEventTap, down);
	usleep(5000);
	CGEventPost(kCGHIDEventTap, up);
	CFRelease(down);
	CFRelease(up);
	return 1;
}

// Scroll
int orax_scrollAt(double x, double y, double dx, double dy) {
	CGEventRef ev = CGEventCreateScrollWheelEvent(NULL, kCGScrollEventUnitPixel, 2, (int32_t)dy, (int32_t)dx);
	if (!ev) return 0;
	CGEventPost(kCGHIDEventTap, ev);
	CFRelease(ev);
	return 1;
}

// Launch app via open command (simplest cross-version approach)
int orax_launchApp(const char* bundleID) {
	CFStringRef bid = CFStringCreateWithCString(NULL, bundleID, kCFStringEncodingUTF8);
	if (!bid) return 0;
	CFArrayRef apps = LSCopyApplicationURLsForBundleIdentifier(bid, NULL);
	CFRelease(bid);
	if (!apps || CFArrayGetCount(apps) == 0) { if (apps) CFRelease(apps); return 0; }
	CFURLRef url = (CFURLRef)CFArrayGetValueAtIndex(apps, 0);
	CFRetain(url);
	CFRelease(apps);
	OSStatus st = LSOpenCFURLRef(url, NULL);
	CFRelease(url);
	return st == noErr;
}

// Focus an app by bringing it to the foreground
int orax_focusApp(pid_t pid) {
	ProcessSerialNumber psn = {0, kNoProcess};
	while (GetNextProcess(&psn) == noErr) {
		pid_t p;
		if (GetProcessPID(&psn, &p) == noErr && p == pid) {
			return SetFrontProcessWithOptions(&psn, kSetFrontProcessFrontWindowOnly) == noErr;
		}
	}
	return 0;
}
*/
import "C"
import (
	"bytes"
	"context"
	"fmt"
	"image/png"
	"os/exec"
	"time"
	"unsafe"

	"github.com/sahajpatel123/synapticapp/internal/computeruse"
)

// darwinORAX is the CGo implementation of oraXImpl for macOS.
type darwinORAX struct{}

func newORAXImpl() oraXImpl { return &darwinORAX{} }

func (d *darwinORAX) name() string { return "orax" }

func (d *darwinORAX) isAvailable() bool {
	return C.orax_isTrusted() != 0
}

// captureScreen captures the main display as a PNG screenshot via the
// system screencapture(1) command. This avoids the macOS 15+
// CGDisplayCreateImage deprecation.
func (d *darwinORAX) captureScreen() (*computeruse.Screenshot, error) {
	cmd := exec.CommandContext(context.Background(), "screencapture", "-x", "-t", "png", "-")
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("orax: screencapture: %w", err)
	}

	// Decode PNG header to get dimensions.
	cfg, err := png.DecodeConfig(bytes.NewReader(out))
	w, h := 0, 0
	if err == nil {
		w = cfg.Width
		h = cfg.Height
	}

	return &computeruse.Screenshot{
		Image:     out,
		Width:     w,
		Height:    h,
		Timestamp: time.Now(),
	}, nil
}

// getAXTree reads the accessibility tree from the focused application.
func (d *darwinORAX) getAXTree() (*computeruse.AXTree, error) {
	app := C.orax_getFocusedApp()
	if app == 0 {
		return nil, fmt.Errorf("orax: no focused application")
	}
	defer C.CFRelease(C.CFTypeRef(app))

	var pid C.pid_t
	if C.orax_getPid(app, &pid) == 0 {
		return nil, fmt.Errorf("orax: failed to get PID")
	}

	root, err := d.buildNode(app, 0)
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

func (d *darwinORAX) buildNode(element C.AXUIElementRef, depth int) (*computeruse.AXNode, error) {
	if depth > maxDepth {
		return nil, fmt.Errorf("orax: max depth exceeded at %d", depth)
	}

	node := &computeruse.AXNode{
		Attributes: make(map[string]interface{}),
	}

	// Role
	if s := C.orax_getStringAttr(element, C.kAXRoleAttribute); s != nil {
		node.Role = C.GoString(s)
		C.free(unsafe.Pointer(s))
	}
	// Title
	if s := C.orax_getStringAttr(element, C.kAXTitleAttribute); s != nil {
		node.Title = C.GoString(s)
		C.free(unsafe.Pointer(s))
	}
	// Value
	if s := C.orax_getStringAttr(element, C.kAXValueAttribute); s != nil {
		node.Value = C.GoString(s)
		C.free(unsafe.Pointer(s))
	}
	// Description
	if s := C.orax_getStringAttr(element, C.kAXDescriptionAttribute); s != nil {
		node.Description = C.GoString(s)
		C.free(unsafe.Pointer(s))
	}
	// Bounds
	var x, y, w, h C.double
	if C.orax_getPos(element, &x, &y) != 0 && C.orax_getSize(element, &w, &h) != 0 {
		node.Bounds = &computeruse.Rect{
			X: float64(x), Y: float64(y),
			Width: float64(w), Height: float64(h),
		}
	}
	// Children
	if arr := C.orax_getChildren(element); arr != 0 {
		count := int(C.CFArrayGetCount(arr))
		if count > 0 {
			node.Children = make([]*computeruse.AXNode, 0, count)
			for i := 0; i < count; i++ {
				child := C.CFArrayGetValueAtIndex(arr, C.CFIndex(i))
				if child == nil {
					continue
				}
				C.CFRetain(C.CFTypeRef(child))
				cn, err := d.buildNode(C.AXUIElementRef(child), depth+1)
				C.CFRelease(C.CFTypeRef(child))
				if err != nil {
					continue
				}
				node.Children = append(node.Children, cn)
			}
		}
		C.CFRelease(C.CFTypeRef(arr))
	}

	return node, nil
}

// execute performs a computer-use action using CGEvent and CoreGraphics.
func (d *darwinORAX) execute(action *computeruse.Action) (*computeruse.ActionResult, error) {
	if action == nil {
		return nil, fmt.Errorf("orax: nil action")
	}

	start := time.Now()
	var err error

	switch action.Type {
	case computeruse.ActionClick:
		err = d.execClick(action)
	case computeruse.ActionTypeText:
		err = d.execType(action)
	case computeruse.ActionScroll:
		err = d.execScroll(action)
	case computeruse.ActionKeyPress:
		err = d.execKeyPress(action)
	case computeruse.ActionLaunch:
		err = d.execLaunch(action)
	case computeruse.ActionFocus:
		err = d.execFocus(action)
	default:
		err = computeruse.ErrUnsupportedAction
	}

	result := &computeruse.ActionResult{
		Success:  err == nil,
		Error:    err,
		Duration: time.Since(start),
		Action:   action,
	}
	return result, err
}

func (d *darwinORAX) execClick(action *computeruse.Action) error {
	x, y, err := d.resolveCoords(action)
	if err != nil {
		return err
	}
	if C.orax_clickAt(C.double(x), C.double(y)) == 0 {
		return fmt.Errorf("orax: click failed at (%.0f, %.0f)", x, y)
	}
	return nil
}

func (d *darwinORAX) execType(action *computeruse.Action) error {
	target, err := d.findTargetElement(action)
	if err != nil {
		return err
	}
	if target != 0 {
		defer C.CFRelease(C.CFTypeRef(target))
		// Try to set the text value directly on the text field.
		cStr := C.CString(action.Value)
		defer C.free(unsafe.Pointer(cStr))
		if C.orax_setStringValue(target, cStr) != 0 {
			// Also fire the confirm action so the app registers the change.
			_ = C.orax_confirm(target)
			return nil
		}
	}
	// Fallback: type via CGEvent paste.
	if C.orax_typeString(C.CString(action.Value)) == 0 {
		return fmt.Errorf("orax: type failed")
	}
	return nil
}

func (d *darwinORAX) execScroll(action *computeruse.Action) error {
	x, y, err := d.resolveCoords(action)
	if err != nil {
		return err
	}
	const defaultScrollLines = 3.0
	dx, dy := 0.0, -defaultScrollLines
	switch action.Value {
	case "up":
		dx, dy = 0.0, defaultScrollLines
	case "down":
		dx, dy = 0.0, -defaultScrollLines
	case "left":
		dx, dy = -defaultScrollLines, 0.0
	case "right":
		dx, dy = defaultScrollLines, 0.0
	}
	if C.orax_scrollAt(C.double(x), C.double(y), C.double(dx), C.double(dy)) == 0 {
		return fmt.Errorf("orax: scroll failed")
	}
	return nil
}

// keyMap maps action values to (keyCode, flags) pairs.
var keyMap = map[string]struct {
	code  C.CGKeyCode
	flags C.CGEventFlags
}{
	"return": {36, 0},
	"enter":  {36, 0},
	"tab":    {48, 0},
	"space":  {49, 0},
	"escape": {53, 0},
	"esc":    {53, 0},
	"delete": {51, 0},
	"left":   {123, 0},
	"right":  {124, 0},
	"down":   {125, 0},
	"up":     {126, 0},
	"cmd+c":  {8, C.kCGEventFlagMaskCommand},
	"cmd+v":  {9, C.kCGEventFlagMaskCommand},
	"cmd+q":  {12, C.kCGEventFlagMaskCommand},
	"cmd+w":  {13, C.kCGEventFlagMaskCommand},
}

func (d *darwinORAX) execKeyPress(action *computeruse.Action) error {
	k, ok := keyMap[action.Value]
	if !ok {
		return fmt.Errorf("orax: unsupported key %q", action.Value)
	}
	if k.flags != 0 {
		if C.orax_typeModKey(k.code, k.flags) == 0 {
			return fmt.Errorf("orax: key press failed for %q", action.Value)
		}
		return nil
	}
	if C.orax_typeChar(k.code) == 0 {
		return fmt.Errorf("orax: key press failed for %q", action.Value)
	}
	return nil
}

func (d *darwinORAX) execLaunch(action *computeruse.Action) error {
	cStr := C.CString(action.Value)
	defer C.free(unsafe.Pointer(cStr))
	if C.orax_launchApp(cStr) == 0 {
		return fmt.Errorf("orax: launch failed for %q", action.Value)
	}
	return nil
}

func (d *darwinORAX) execFocus(action *computeruse.Action) error {
	var pid C.pid_t = C.pid_t(action.AppPID)
	if pid == 0 {
		return fmt.Errorf("orax: no PID for focus action")
	}
	if C.orax_focusApp(pid) == 0 {
		return fmt.Errorf("orax: focus failed for PID %d", pid)
	}
	return nil
}

// resolveCoords gets the (x, y) center of the action's target element,
// or falls back to the action's Bounds.
func (d *darwinORAX) resolveCoords(action *computeruse.Action) (float64, float64, error) {
	if action.Bounds != nil {
		return action.Bounds.X + action.Bounds.Width/2,
			action.Bounds.Y + action.Bounds.Height/2, nil
	}
	el, err := d.findTargetElement(action)
	if err != nil {
		return 0, 0, err
	}
	if el == 0 {
		return 0, 0, fmt.Errorf("orax: cannot resolve coordinates")
	}
	defer C.CFRelease(C.CFTypeRef(el))
	var x, y, w, h C.double
	if C.orax_getPos(el, &x, &y) == 0 || C.orax_getSize(el, &w, &h) == 0 {
		return 0, 0, fmt.Errorf("orax: element has no bounds")
	}
	return float64(x) + float64(w)/2, float64(y) + float64(h)/2, nil
}

// findTargetElement locates an AXUIElement matching the action's target.
// Returns the element (with retain count +1; caller must CFRelease) or 0.
func (d *darwinORAX) findTargetElement(action *computeruse.Action) (C.AXUIElementRef, error) {
	if action.Target == nil {
		var zero C.AXUIElementRef
		return zero, nil
	}
	app := C.orax_getFocusedApp()
	if app == 0 {
		return 0, fmt.Errorf("orax: no focused application")
	}
	defer C.CFRelease(C.CFTypeRef(app))

	return d.findInTree(app, action.Target), nil
}

func (d *darwinORAX) findInTree(root C.AXUIElementRef, target *computeruse.Target) C.AXUIElementRef {
	if root == 0 {
		return 0
	}

	// Check if this element matches.
	if d.matchesTarget(root, target) {
		C.CFRetain(C.CFTypeRef(root))
		return root
	}

	// Search children.
	if arr := C.orax_getChildren(root); arr != 0 {
		defer C.CFRelease(C.CFTypeRef(arr))
		for i, n := 0, int(C.CFArrayGetCount(arr)); i < n; i++ {
			child := C.CFArrayGetValueAtIndex(arr, C.CFIndex(i))
			if child == nil {
				continue
			}
			if found := d.findInTree(C.AXUIElementRef(child), target); found != 0 {
				return found
			}
		}
	}
	return 0
}

func (d *darwinORAX) matchesTarget(el C.AXUIElementRef, target *computeruse.Target) bool {
	if target.Role != "" {
		s := C.orax_getStringAttr(el, C.kAXRoleAttribute)
		if s == nil {
			return false
		}
		match := C.GoString(s) == target.Role
		C.free(unsafe.Pointer(s))
		if !match {
			return false
		}
	}
	if target.Title != "" {
		s := C.orax_getStringAttr(el, C.kAXTitleAttribute)
		if s == nil {
			return false
		}
		match := C.GoString(s) == target.Title
		C.free(unsafe.Pointer(s))
		if !match {
			return false
		}
	}
	if target.Value != "" {
		s := C.orax_getStringAttr(el, C.kAXValueAttribute)
		if s == nil {
			return false
		}
		match := C.GoString(s) == target.Value
		C.free(unsafe.Pointer(s))
		if !match {
			return false
		}
	}
	return true
}
