//go:build darwin

package backends

import (
	"context"
	"errors"
	"fmt"
	"os/exec"
	"strings"
	"time"

	"github.com/sahajpatel123/synapticapp/internal/computeruse"
)

const (
	mcpScrollDirUp    = "up"
	mcpScrollDirDown  = "down"
	mcpScrollDirLeft  = "left"
	mcpScrollDirRight = "right"
)

// darwinMCP uses osascript subprocess calls for AppleScript execution.
type darwinMCP struct{}

func newMCPImpl() macOSMCPImpl { return &darwinMCP{} }

func (d *darwinMCP) name() string { return "macos-mcp" }

func (d *darwinMCP) isAvailable() bool {
	_, err := exec.LookPath("osascript")
	return err == nil
}

func (d *darwinMCP) runAppleScript(script string) (string, error) {
	out, err := exec.CommandContext(context.Background(), "osascript", "-e", script).Output() //nolint:gosec // script is built from escapeAppleScript-escaped model-controlled values
	if err != nil {
		var exitErr *exec.ExitError
		if errors.As(err, &exitErr) {
			return "", fmt.Errorf("osascript: %s", strings.TrimSpace(string(exitErr.Stderr)))
		}
		return "", fmt.Errorf("osascript: %w", err)
	}
	return strings.TrimSpace(string(out)), nil
}

func (d *darwinMCP) captureScreen() (*computeruse.Screenshot, error) {
	out, err := exec.CommandContext(context.Background(), "screencapture", "-x", "-t", "png", "-").Output()
	if err != nil {
		return nil, fmt.Errorf("macos-mcp: screencapture: %w", err)
	}
	return &computeruse.Screenshot{
		Image:     out,
		Width:     0,
		Height:    0,
		Timestamp: time.Now(),
	}, nil
}

func (d *darwinMCP) getAXTree() (*computeruse.AXTree, error) {
	// Read the frontmost application's UI via System Events.
	script := `tell application "System Events"
		set frontApp to first application process whose frontmost is true
		set appName to name of frontApp
		set pid to unix id of frontApp
		set info to "app:" & appName & " pid:" & pid
		try
			set mainWindow to first window of frontApp
			set numElems to count of every UI element of mainWindow
			set info to info & " elements:" & numElems
		end try
		return info
	end tell`
	out, err := d.runAppleScript(script)
	if err != nil {
		return nil, fmt.Errorf("macos-mcp: ax_tree: %w", err)
	}
	// Parse the compact output: "app:Name pid:12345 elements:42"
	root := &computeruse.AXNode{
		Role:       "AXApplication",
		Title:      out,
		Attributes: make(map[string]interface{}),
	}
	return &computeruse.AXTree{
		Root:      root,
		Timestamp: time.Now(),
	}, nil
}

func (d *darwinMCP) execute(action *computeruse.Action) (*computeruse.ActionResult, error) {
	if action == nil {
		return nil, fmt.Errorf("macos-mcp: nil action")
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
	r := &computeruse.ActionResult{Success: err == nil, Error: err, Duration: time.Since(start), Action: action}
	return r, err
}

func (d *darwinMCP) execClick(action *computeruse.Action) error {
	title := ""
	if action.Target != nil {
		title = action.Target.Title
	}
	script := fmt.Sprintf(`tell application "System Events"
		set frontApp to first application process whose frontmost is true
		set targetButton to first button of window 1 of frontApp whose title is "%s"
		click targetButton
	end tell`, escapeAppleScript(title))
	_, err := d.runAppleScript(script)
	return err
}

func (d *darwinMCP) execType(action *computeruse.Action) error {
	script := fmt.Sprintf(`tell application "System Events"
		keystroke "%s"
	end tell`, escapeAppleScript(action.Value))
	_, err := d.runAppleScript(script)
	return err
}

func (d *darwinMCP) execScroll(action *computeruse.Action) error {
	dir := action.Value
	count := "3"
	script := fmt.Sprintf(`tell application "System Events"
		repeat %s times
			key code %s
		end repeat
	end tell`, count, scrollKeyCode(dir))
	_, err := d.runAppleScript(script)
	return err
}

func (d *darwinMCP) execKeyPress(action *computeruse.Action) error {
	script := fmt.Sprintf(`tell application "System Events"
		keystroke "%s"
	end tell`, escapeAppleScript(action.Value))
	_, err := d.runAppleScript(script)
	return err
}

func (d *darwinMCP) execLaunch(action *computeruse.Action) error {
	script := fmt.Sprintf(`tell application "%s" to activate`, escapeAppleScript(action.Value))
	_, err := d.runAppleScript(script)
	return err
}

func (d *darwinMCP) execFocus(_ *computeruse.Action) error {
	script := `tell application "System Events"
		set frontApp to first application process whose frontmost is true
		set frontmost of frontApp to true
	end tell`
	_, err := d.runAppleScript(script)
	return err
}

// scrollKeyCode returns the AppleScript key code for a direction.
func scrollKeyCode(dir string) string {
	switch dir {
	case mcpScrollDirUp:
		return "126"
	case mcpScrollDirDown:
		return "125"
	case mcpScrollDirLeft:
		return "123"
	case mcpScrollDirRight:
		return "124"
	default:
		return "125"
	}
}

func escapeAppleScript(s string) string {
	// Escape backslashes and double quotes for AppleScript strings.
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	// Escape backticks. AppleScript evaluates backtick-quoted
	// expressions inside string literals, so an unescaped backtick
	// in a model-controlled value can inject a `do shell script "..."`
	// expression. This is the osascript injection vector flagged in
	// the security audit (F-11) and the backend audit (B-11).
	s = strings.ReplaceAll(s, "`", "\\`")
	return s
}
