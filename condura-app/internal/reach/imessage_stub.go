//go:build !darwin

package reach

import (
	"context"
	"fmt"
)

const imessageDarwinOnly = "iMessage is only available on macOS."

type imessageChannel struct{}

func newIMessageChannel() *imessageChannel {
	return &imessageChannel{}
}

func (i *imessageChannel) Connect(_ context.Context, _ string) error {
	return &UnsupportedError{Name: "imessage", Message: imessageDarwinOnly}
}

func (i *imessageChannel) Disconnect(_ context.Context) error { return nil }

func (i *imessageChannel) Send(_ context.Context, _, _ string) error {
	return fmt.Errorf("reach: imessage not connected")
}

func (i *imessageChannel) Receive(_ context.Context) (<-chan Message, error) {
	return nil, fmt.Errorf("reach: imessage not connected")
}

func (i *imessageChannel) Status(_ context.Context) (ChannelStatus, error) {
	return ChannelStatus{Name: "imessage", Connected: false}, nil
}
