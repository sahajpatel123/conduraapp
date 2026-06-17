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
	script := fmt.Sprintf(
		`tell application "Messages" to send %q to buddy %q of (service 1 whose service type is iMessage)`,
		text, target,
	)
	cmd := exec.CommandContext(ctx, "osascript", "-e", script) //nolint:gosec // AppleScript send via system osascript
	if out, err := cmd.CombinedOutput(); err != nil {
		return fmt.Errorf("reach: imessage send: %w: %s", err, strings.TrimSpace(string(out)))
	}
	return nil
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
