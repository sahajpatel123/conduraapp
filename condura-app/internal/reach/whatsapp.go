package reach

import (
	"context"
	"fmt"
)

const whatsAppComingSoon = "WhatsApp integration coming in v0.2.0. Use Telegram in the meantime."

type whatsAppChannel struct{}

func newWhatsAppChannel() *whatsAppChannel {
	return &whatsAppChannel{}
}

func (w *whatsAppChannel) Connect(_ context.Context, _ string) error {
	return &UnsupportedError{Name: "whatsapp", Message: whatsAppComingSoon}
}

func (w *whatsAppChannel) Disconnect(_ context.Context) error { return nil }

func (w *whatsAppChannel) Send(_ context.Context, _, _ string) error {
	return fmt.Errorf("reach: whatsapp not connected")
}

func (w *whatsAppChannel) Receive(_ context.Context) (<-chan Message, error) {
	return nil, fmt.Errorf("reach: whatsapp not connected")
}

func (w *whatsAppChannel) Status(_ context.Context) (ChannelStatus, error) {
	return ChannelStatus{Name: "whatsapp", Connected: false}, nil
}
