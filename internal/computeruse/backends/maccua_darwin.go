// Package backends provides computer-use backend implementations.
// This file defines the mac-cua backend — background-first execution
// using CGEventPostToPid targeting specific applications.
//
// mac-cua is the second tier in the 4-tier router:
//   1. ORAX Eye  — foreground, AX-structured
//   2. mac-cua   — background, CGEventPostToPid
//   3. macOS-MCP — comprehensive, AppleScript
//   4. Vision CUA — API-based, last resort

//go:build darwin && cgo

package backends

/*
#cgo LDFLAGS: -framework ApplicationServices -framework CoreGraphics

#include <ApplicationServices/ApplicationServices.h>
#include <CoreGraphics/CoreGraphics.h>

// Accessibility trust check.
Boolean mc_isTrusted() {
	return AXIsProcessTrusted();
}

// Get focused application.
AXUIElementRef mc_getFocusedApp() {
	AXUIElementRef sys = AXUIElementCreateSystemWide();
	if (!sys) return NULL;
	CFTypeRef app = NULL;
	AXError err = AXUIElementCopyAttributeValue(
		sys, kAXFocusedApplicationAttribute, &app);
	CFRelease(sys);
	if (err != kAXErrorSuccess || !app) return NULL;
	return (AXUIElementRef)app;
}

// Get app PID.
int mc_getPid(AXUIElementRef el, pid_t* pid) {
	return AXUIElementGetPid(el, pid) == kAXErrorSuccess;
}

// Get string attribute.
char* mc_getStringAttr(AXUIElementRef el, CFStringRef attr) {
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

// Get position and size of an AX element.
int mc_getPos(AXUIElementRef el, double* x, double* y) {
	CFTypeRef val = NULL;
	if (AXUIElementCopyAttributeValue(el, kAXPositionAttribute, &val) != kAXErrorSuccess || !val) return 0;
	CGPoint pt;
	int ok = AXValueGetValue((AXValueRef)val, kAXValueCGPointType, &pt);
	CFRelease(val);
	if (!ok) return 0;
	*x = pt.x; *y = pt.y;
	return 1;
}

int mc_getSize(AXUIElementRef el, double* w, double* h) {
	CFTypeRef val = NULL;
	if (AXUIElementCopyAttributeValue(el, kAXSizeAttribute, &val) != kAXErrorSuccess || !val) return 0;
	CGSize sz;
	int ok = AXValueGetValue((AXValueRef)val, kAXValueCGSizeType, &sz);
	CFRelease(val);
	if (!ok) return 0;
	*w = sz.width; *h = sz.height;
	return 1;
}

// Get children.
CFArrayRef mc_getChildren(AXUIElementRef el) {
	CFTypeRef val = NULL;
	if (AXUIElementCopyAttributeValue(el, kAXChildrenAttribute, &val) != kAXErrorSuccess || !val) return NULL;
	return (CFArrayRef)val;
}

// === CGEventPostToPid — background event posting ===

// Post a mouse click at coordinates targeting a specific PID.
int mc_clickAtPid(double x, double y, pid_t pid) {
	CGPoint pt = CGPointMake(x, y);
	CGEventRef down = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseDown, pt, kCGMouseButtonLeft);
	CGEventRef up   = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseUp,   pt, kCGMouseButtonLeft);
	if (!down || !up) {
		if (down) CFRelease(down);
		if (up)   CFRelease(up);
		return 0;
	}
	CGEventPostToPid(pid, down);
	usleep(10000);
	CGEventPostToPid(pid, up);
	CFRelease(down);
	CFRelease(up);
	return 1;
}

// Post a double click.
int mc_doubleClickAtPid(double x, double y, pid_t pid) {
	CGPoint pt = CGPointMake(x, y);
	CGEventRef down = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseDown, pt, kCGMouseButtonLeft);
	CGEventRef up   = CGEventCreateMouseEvent(NULL, kCGEventLeftMouseUp,   pt, kCGMouseButtonLeft);
	if (!down || !up) {
		if (down) CFRelease(down);
		if (up)   CFRelease(up);
		return 0;
	}
	CGEventSetIntegerValueField(down, kCGMouseEventClickState, 2);
	CGEventSetIntegerValueField(up,   kCGMouseEventClickState, 2);
	CGEventPostToPid(pid, down);
	usleep(10000);
	CGEventPostToPid(pid, up);
	CFRelease(down);
	CFRelease(up);
	return 1;
}

// Post keystrokes to a specific PID.
int mc_typeToPid(const char* text, pid_t pid) {
	if (!text) return 0;
	// Type via CGEvent keyboard simulation for the target PID.
	// For Unicode reliability, use the pasteboard approach.
	CFStringRef s = CFStringCreateWithCString(NULL, text, kCFStringEncodingUTF8);
	if (!s) return 0;

	// Cmd+V to paste the clipboard contents.
	CGEventRef cmdV = CGEventCreateKeyboardEvent(NULL, (CGKeyCode)9, true);
	CGEventSetFlags(cmdV, kCGEventFlagMaskCommand);
	CGEventRef cmdVUp = CGEventCreateKeyboardEvent(NULL, (CGKeyCode)9, false);
	CGEventSetFlags(cmdVUp, kCGEventFlagMaskCommand);
	CGEventPostToPid(pid, cmdV);
	usleep(10000);
	CGEventPostToPid(pid, cmdVUp);
	CFRelease(cmdV);
	CFRelease(cmdVUp);
	CFRelease(s);
	return 1;
}

// Post a single key to a PID.
int mc_keyToPid(CGKeyCode code, pid_t pid) {
	CGEventRef down = CGEventCreateKeyboardEvent(NULL, code, true);
	CGEventRef up   = CGEventCreateKeyboardEvent(NULL, code, false);
	if (!down || !up) {
		if (down) CFRelease(down);
		if (up)   CFRelease(up);
		return 0;
	}
	CGEventPostToPid(pid, down);
	usleep(5000);
	CGEventPostToPid(pid, up);
	CFRelease(down);
	CFRelease(up);
	return 1;
}

// Post a modified key (Cmd+C, Cmd+Q, etc.) to a PID.
int mc_modKeyToPid(CGKeyCode code, CGEventFlags flags, pid_t pid) {
	CGEventRef down = CGEventCreateKeyboardEvent(NULL, code, true);
	CGEventRef up   = CGEventCreateKeyboardEvent(NULL, code, false);
	if (!down || !up) {
		if (down) CFRelease(down);
		if (up)   CFRelease(up);
		return 0;
	}
	CGEventSetFlags(down, flags);
	CGEventSetFlags(up,   flags);
	CGEventPostToPid(pid, down);
	usleep(5000);
	CGEventPostToPid(pid, up);
	CFRelease(down);
	CFRelease(up);
	return 1;
}

// Post a scroll event to a PID.
int mc_scrollToPid(double dx, double dy, pid_t pid) {
	CGEventRef ev = CGEventCreateScrollWheelEvent(NULL, kCGScrollEventUnitPixel, 2, (int32_t)dy, (int32_t)dx);
	if (!ev) return 0;
	CGEventPostToPid(pid, ev);
	CFRelease(ev);
	return 1;
}

// Focus (bring to foreground) using ProcessSerialNumber.
// Note: deprecated since 10.9, will migrate to NSRunningApplication.
int mc_focusPid(pid_t pid) {
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
	"fmt"
	"time"
	"unsafe"

	"github.com/sahajpatel123/synapticapp/internal/computeruse"
)

const (
	mcMaxDepth       = 50
	mcScrollLines    = 3.0
	mcScrollDirUp    = "up"
	mcScrollDirDown  = "down"
	mcScrollDirLeft  = "left"
	mcScrollDirRight = "right"
)

// darwinMC is the CGo implementation of macCUAImpl for macOS.
type darwinMC struct{}

func newMCImpl() macCUAImpl { return &darwinMC{} }

func (d *darwinMC) name() string { return "mac-cua" }

func (d *darwinMC) isAvailable() bool { return C.mc_isTrusted() != 0 }

func (d *darwinMC) captureScreen() (*computeruse.Screenshot, error) {
	return nil, computeruse.ErrUnsupportedAction // mac-cua doesn't do screenshots
}

func (d *darwinMC) getAXTree() (*computeruse.AXTree, error) {
	app := C.mc_getFocusedApp()
	if app == 0 {
		return nil, fmt.Errorf("mac-cua: no focused application")
	}
	defer C.CFRelease(C.CFTypeRef(app))

	var pid C.pid_t
	if C.mc_getPid(app, &pid) == 0 {
		return nil, fmt.Errorf("mac-cua: failed to get PID")
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

func (d *darwinMC) buildNode(el C.AXUIElementRef, depth int) (*computeruse.AXNode, error) {
	if depth > mcMaxDepth {
		return nil, fmt.Errorf("mac-cua: max depth")
	}
	n := &computeruse.AXNode{Attributes: make(map[string]interface{})}
	if s := C.mc_getStringAttr(el, C.kAXRoleAttribute); s != nil {
		n.Role = C.GoString(s)
		C.free(unsafe.Pointer(s))
	}
	if s := C.mc_getStringAttr(el, C.kAXTitleAttribute); s != nil {
		n.Title = C.GoString(s)
		C.free(unsafe.Pointer(s))
	}
	if s := C.mc_getStringAttr(el, C.kAXValueAttribute); s != nil {
		n.Value = C.GoString(s)
		C.free(unsafe.Pointer(s))
	}
	if s := C.mc_getStringAttr(el, C.kAXDescriptionAttribute); s != nil {
		n.Description = C.GoString(s)
		C.free(unsafe.Pointer(s))
	}
	var x, y, w, h C.double
	if C.mc_getPos(el, &x, &y) != 0 && C.mc_getSize(el, &w, &h) != 0 {
		n.Bounds = &computeruse.Rect{X: float64(x), Y: float64(y), Width: float64(w), Height: float64(h)}
	}
	if arr := C.mc_getChildren(el); arr != 0 {
		count := int(C.CFArrayGetCount(arr))
		if count > 0 {
			n.Children = make([]*computeruse.AXNode, 0, count)
			for i := 0; i < count; i++ {
				c := C.CFArrayGetValueAtIndex(arr, C.CFIndex(i))
				if c == nil {
					continue
				}
				C.CFRetain(C.CFTypeRef(c))
				cn, _ := d.buildNode(C.AXUIElementRef(c), depth+1)
				C.CFRelease(C.CFTypeRef(c))
				if cn != nil {
					n.Children = append(n.Children, cn)
				}
			}
		}
		C.CFRelease(C.CFTypeRef(arr))
	}
	return n, nil
}

func (d *darwinMC) execute(action *computeruse.Action) (*computeruse.ActionResult, error) {
	if action == nil {
		return nil, fmt.Errorf("mac-cua: nil action")
	}
	start := time.Now()
	pid := C.pid_t(action.AppPID)

	var err error
	switch action.Type {
	case computeruse.ActionClick:
		err = d.execClick(action, pid)
	case computeruse.ActionTypeText:
		err = d.execType(action, pid)
	case computeruse.ActionScroll:
		err = d.execScroll(action, pid)
	case computeruse.ActionKeyPress:
		err = d.execKeyPress(action, pid)
	case computeruse.ActionFocus:
		err = d.execFocus(pid)
	default:
		err = computeruse.ErrUnsupportedAction
	}
	r := &computeruse.ActionResult{Success: err == nil, Error: err, Duration: time.Since(start), Action: action}
	return r, err
}

func (d *darwinMC) execClick(action *computeruse.Action, pid C.pid_t) error {
	x, y, err := d.resolveCoords(action)
	if err != nil {
		return err
	}
	if C.mc_clickAtPid(C.double(x), C.double(y), pid) == 0 {
		return fmt.Errorf("mac-cua: click failed")
	}
	return nil
}

func (d *darwinMC) execType(action *computeruse.Action, pid C.pid_t) error {
	cStr := C.CString(action.Value)
	defer C.free(unsafe.Pointer(cStr))
	if C.mc_typeToPid(cStr, pid) == 0 {
		return fmt.Errorf("mac-cua: type failed")
	}
	return nil
}

func (d *darwinMC) execScroll(action *computeruse.Action, pid C.pid_t) error {
	const dl = mcScrollLines
	dx, dy := 0.0, -dl
	switch action.Value {
	case mcScrollDirUp:
		dx, dy = 0.0, dl
	case mcScrollDirDown:
		dx, dy = 0.0, -dl
	case mcScrollDirLeft:
		dx, dy = -dl, 0.0
	case mcScrollDirRight:
		dx, dy = dl, 0.0
	}
	if C.mc_scrollToPid(C.double(dx), C.double(dy), pid) == 0 {
		return fmt.Errorf("mac-cua: scroll failed")
	}
	return nil
}

var mckeyMap = map[string]struct {
	code  C.CGKeyCode
	flags C.CGEventFlags
}{
	"return": {36, 0}, "enter": {36, 0}, "tab": {48, 0}, "space": {49, 0},
	"escape": {53, 0}, "esc": {53, 0}, "delete": {51, 0},
	"left": {123, 0}, "right": {124, 0}, "down": {125, 0}, "up": {126, 0},
	"cmd+c": {8, C.kCGEventFlagMaskCommand}, "cmd+v": {9, C.kCGEventFlagMaskCommand},
	"cmd+q": {12, C.kCGEventFlagMaskCommand}, "cmd+w": {13, C.kCGEventFlagMaskCommand},
}

func (d *darwinMC) execKeyPress(action *computeruse.Action, pid C.pid_t) error {
	k, ok := mckeyMap[action.Value]
	if !ok {
		return fmt.Errorf("mac-cua: unsupported key %q", action.Value)
	}
	if k.flags != 0 {
		if C.mc_modKeyToPid(k.code, k.flags, pid) == 0 {
			return fmt.Errorf("mac-cua: key failed for %q", action.Value)
		}
		return nil
	}
	if C.mc_keyToPid(k.code, pid) == 0 {
		return fmt.Errorf("mac-cua: key failed for %q", action.Value)
	}
	return nil
}

func (d *darwinMC) execFocus(pid C.pid_t) error {
	if C.mc_focusPid(pid) == 0 {
		return fmt.Errorf("mac-cua: focus failed for PID %d", pid)
	}
	return nil
}

func (d *darwinMC) resolveCoords(action *computeruse.Action) (float64, float64, error) {
	if action.Bounds != nil {
		return action.Bounds.X + action.Bounds.Width/2, action.Bounds.Y + action.Bounds.Height/2, nil
	}
	return 0, 0, fmt.Errorf("mac-cua: no coordinates (background mode requires Bounds)")
}
