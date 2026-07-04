package reach

import (
	"context"
	"fmt"
)

const signalComingSoon = "Signal integration coming in v0.2.0. Use Telegram in the meantime."

type signalChannel struct{}

func newSignalChannel() *signalChannel {
	return &signalChannel{}
}

func (s *signalChannel) Connect(_ context.Context, _ string) error {
	return &UnsupportedError{Name: "signal", Message: signalComingSoon}
}

func (s *signalChannel) Disconnect(_ context.Context) error { return nil }

func (s *signalChannel) Send(_ context.Context, _, _ string) error {
	return fmt.Errorf("reach: signal not connected")
}

func (s *signalChannel) Receive(_ context.Context) (<-chan Message, error) {
	return nil, fmt.Errorf("reach: signal not connected")
}

func (s *signalChannel) Status(_ context.Context) (ChannelStatus, error) {
	return ChannelStatus{Name: "signal", Connected: false}, nil
}
