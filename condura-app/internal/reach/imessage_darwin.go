//go:build darwin

package reach

import (
	"context"
	"fmt"
	"os/exec"
	"strings"
)

type imessageChannel struct {
	recipient string
	connected bool
}

func newIMessageChannel() *imessageChannel {
	return &imessageChannel{}
}

func (i *imessageChannel) Connect(_ context.Context, token string) error {
	recipient := strings.TrimSpace(token)
	if recipient == "" {
		return fmt.Errorf("reach: imessage recipient is empty")
	}
	i.recipient = recipient
	i.connected = true
	return nil
}

func (i *imessageChannel) Disconnect(_ context.Context) error {
	i.recipient = ""
	i.connected = false
	return nil
}

func (i *imessageChannel) Send(ctx context.Context, chatID, text string) error {
	if !i.connected {
		return fmt.Errorf("reach: imessage not connected")
	}
	target := strings.TrimSpace(chatID)
	if target == "" {
		target = i.recipient
	}
	if target == "" {
		return fmt.Errorf("reach: imessage recipient is empty")
	}
	if strings.TrimSpace(text) == "" {
		return fmt.Errorf("reach: imessage message is empty")
	}
	// Build the AppleScript with explicit escaping. Go's %q verb
	// produces a Go-quoted string, not an AppleScript-safe string
	// literal — it does not escape backticks, which AppleScript
	// evaluates inside string literals, allowing a model-controlled
	// text or recipient to inject a `do shell script "..."` expression
	// (security audit F-11 / backend audit B-11). We escape backslash,
	// double-quote, and backtick explicitly and wrap in double quotes.
	script := fmt.Sprintf(
		`tell application "Messages" to send "%s" to buddy "%s" of (service 1 whose service type is iMessage)`,
		escapeAppleScriptString(text), escapeAppleScriptString(target),
	)
	cmd := exec.CommandContext(ctx, "osascript", "-e", script) //nolint:gosec // values are AppleScript-escaped below
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("reach: imessage send: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
}

// escapeAppleScriptString escapes a value for safe interpolation
// into an AppleScript double-quoted string literal. It escapes
// backslash, double-quote, and backtick. The backtick escape closes
// the osascript injection vector (F-11/B-11) where a model-controlled
// value could otherwise inject a `do shell script "..."` expression.
//
// Audit 2026-07-01: also escape the ampersand ("&" is the AppleScript
// string-concat operator and breaks the literal boundary when used
// unescaped), CR, LF, and TAB (which AppleScript treats as expression
// terminators inside double-quoted strings).
func escapeAppleScriptString(s string) string {
	s = strings.ReplaceAll(s, "\\", "\\\\")
	s = strings.ReplaceAll(s, "\"", "\\\"")
	s = strings.ReplaceAll(s, "`", "\\`")
	s = strings.ReplaceAll(s, "&", "\\&")
	s = strings.ReplaceAll(s, "\n", "\\n")
	s = strings.ReplaceAll(s, "\r", "\\r")
	s = strings.ReplaceAll(s, "\t", "\\t")
	return s
}

func (i *imessageChannel) Receive(_ context.Context) (<-chan Message, error) {
	return nil, fmt.Errorf("reach: imessage receive not implemented")
}

func (i *imessageChannel) Status(_ context.Context) (ChannelStatus, error) {
	return ChannelStatus{
		Name:      "imessage",
		Connected: i.connected,
		ChatID:    i.recipient,
	}, nil
}
