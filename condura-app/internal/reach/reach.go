// Package reach implements messaging channel integrations
// (Phase 14C). Telegram is first; WhatsApp, iMessage, and
// Signal are planned for later releases.
package reach

import (
	"context"
	"fmt"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/safego"
)

// Cipher is the envelope encryption API the Manager uses to
// protect channel tokens at rest. The interface matches the shape
// of storage.DB.EncryptStringWithAAD / DecryptStringWithAAD, so
// *storage.DB satisfies it without an explicit assertion. Defining
// the interface here (rather than importing storage) keeps the
// reach package decoupled from storage and unit-testable with a
// fake cipher.
//
// On encrypt, AAD is the caller-chosen byte slice (here we use
// "reach:<channel-name>") and travels inside the envelope — the
// caller does NOT need to remember the AAD for decryption; the
// envelope is self-describing.
type Cipher interface {
	EncryptStringWithAAD(plaintext string, aad []byte) (string, error)
	DecryptStringWithAAD(envelope string) (string, error)
}

// EventBroker is the minimal broker surface the Manager uses to
// publish channel events. The real sse.Broker satisfies it; tests
// can pass nil or a fake recorder.
type EventBroker interface {
	PublishJSON(name string, data interface{})
}

// Channel is a single messaging integration.
type Channel interface {
	Send(ctx context.Context, chatID, text string) error
	Receive(ctx context.Context) (<-chan Message, error)
	Connect(ctx context.Context, token string) error
	Disconnect(ctx context.Context) error
	Status(ctx context.Context) (ChannelStatus, error)
}

// Message is an incoming message from a channel.
type Message struct {
	ChatID  string `json:"chat_id"`
	Sender  string `json:"sender"`
	Text    string `json:"text"`
	Channel string `json:"channel"`
}

// ChannelStatus describes the current state of a channel.
type ChannelStatus struct {
	Name      string `json:"name"`
	Connected bool   `json:"connected"`
	ChatID    string `json:"chat_id,omitempty"`
	Error     string `json:"error,omitempty"`
}

// Manager orchestrates messaging channels.
type Manager struct {
	channels map[string]Channel
	store    *Store
	cipher   Cipher
	broker   EventBroker
}

// NewManager returns a Manager with the given store. cipher may be
// nil in tests; production always passes a real storage.DB.
//
// broker may be nil — the Manager tolerates a nil broker and skips
// event publishing. Production wires the real sse.Broker via the
// dedicated SetBroker method (or by passing it here).
func NewManager(store *Store, cipher Cipher, broker EventBroker) *Manager {
	return &Manager{
		channels: make(map[string]Channel),
		store:    store,
		cipher:   cipher,
		broker:   broker,
	}
}

// SetBroker wires a broker post-construction. Used when the broker
// is built later than the Manager (e.g., daemon wiring where
// subsystems.Broker is created mid-initSubsystems).
func (m *Manager) SetBroker(broker EventBroker) {
	m.broker = broker
}

// Store returns the underlying channel store.
func (m *Manager) Store() *Store { return m.store }

// List returns all registered channel statuses.
func (m *Manager) List(ctx context.Context) ([]ChannelStatus, error) {
	return m.store.List(ctx)
}

// Connect establishes a connection to a channel. The plaintext
// token is encrypted via the Cipher and stored in reach_channels;
// the Manager never writes the plaintext to disk or to the
// audit log.
func (m *Manager) Connect(ctx context.Context, name, token string) (ChannelStatus, error) {
	ch, err := m.getOrCreateChannel(name)
	if err != nil {
		return ChannelStatus{}, err
	}
	if err := ch.Connect(ctx, token); err != nil {
		return ChannelStatus{}, err
	}
	status, _ := ch.Status(ctx)

	// Encrypt the token before persisting. Failure to encrypt is
	// fatal for Connect — we cannot safely store the plaintext.
	tokenCT := ""
	if m.cipher != nil && token != "" {
		aad := []byte("reach:" + name)
		tokenCT, err = m.cipher.EncryptStringWithAAD(token, aad)
		if err != nil {
			// Best-effort: roll back the channel Connect so the
			// in-memory state matches the on-disk state (no row).
			_ = ch.Disconnect(ctx)
			return ChannelStatus{}, fmt.Errorf("reach: encrypt token: %w", err)
		}
	}

	if err := m.store.Save(ctx, name, tokenCT, status.ChatID, true); err != nil {
		return status, err
	}

	// Kick off Receive in the background to capture chat_id and
	// forward incoming messages to the broker. Only channels that
	// implement Receive usefully (Telegram) will actually produce
	// messages; iMessage and the stubs return an error from Receive
	// and we silently skip them.
	m.startReceive(name, ch)

	return status, nil
}

// Disconnect tears down a channel connection and removes its row
// from the store. The encrypted token is dropped along with the
// row — Disconnect is a full tear-down, not a "pause". Users who
// want to re-enable later re-run Connect.
func (m *Manager) Disconnect(ctx context.Context, name string) error {
	ch, err := m.getOrCreateChannel(name)
	if err != nil {
		return err
	}
	if err := ch.Disconnect(ctx); err != nil {
		return err
	}
	return m.store.Delete(ctx, name)
}

// Status returns the status of a specific channel.
func (m *Manager) Status(ctx context.Context, name string) (ChannelStatus, error) {
	ch, err := m.getOrCreateChannel(name)
	if err != nil {
		return ChannelStatus{}, err
	}
	return ch.Status(ctx)
}

// Restore re-establishes channels that were connected at the last
// daemon shutdown. Called from buildReach on boot. Channels without
// a stored token, or whose decryption fails, are silently skipped
// (a logged audit row is the right next step; deferred until the
// audit pattern for "channel restore failed" is established).
func (m *Manager) Restore(ctx context.Context) error {
	statuses, err := m.store.List(ctx)
	if err != nil {
		return fmt.Errorf("reach: list for restore: %w", err)
	}
	for _, s := range statuses {
		if !s.Connected {
			continue
		}
		tokenCT, err := m.store.GetTokenCiphertext(ctx, s.Name)
		if err != nil || tokenCT == "" {
			continue
		}
		if m.cipher == nil {
			continue
		}
		token, err := m.cipher.DecryptStringWithAAD(tokenCT)
		if err != nil {
			continue
		}
		ch, err := m.getOrCreateChannel(s.Name)
		if err != nil {
			continue
		}
		if err := ch.Connect(ctx, token); err != nil {
			continue
		}
		// Pre-populate the channel's in-memory chat_id from the
		// stored value so the first Send can route correctly
		// without waiting for an inbound message. Receive will
		// overwrite this on the next inbound update.
		if s.ChatID != "" {
			if tc, ok := ch.(*telegramChannel); ok {
				tc.chatID = s.ChatID
			}
		}
		m.startReceive(s.Name, ch)
	}
	return nil
}

// startReceive kicks off the channel's Receive loop and forwards
// every inbound message to the broker as a "channel.message" SSE
// event. The chat_id is persisted to the store on the first
// message — that is the moment we finally know which chat the user
// has been reaching us from.
func (m *Manager) startReceive(name string, ch Channel) {
	msgCh, err := ch.Receive(context.Background())
	if err != nil {
		// iMessage and the stubs return errors here; that's fine.
		return
	}
	captured := false
	safego.Go(func() {
		for msg := range msgCh {
			if !captured && msg.ChatID != "" {
				captured = true
				_ = m.store.UpdateChatID(context.Background(), name, msg.ChatID)
			}
			if m.broker != nil {
				m.broker.PublishJSON("channel.message", map[string]any{
					"channel": name,
					"chat_id": msg.ChatID,
					"sender":  msg.Sender,
					"text":    msg.Text,
				})
			}
		}
	})
}

// getOrCreateChannel returns an existing channel or creates one.
func (m *Manager) getOrCreateChannel(name string) (Channel, error) {
	if ch, ok := m.channels[name]; ok {
		return ch, nil
	}
	switch name {
	case "telegram":
		ch := newTelegramChannel()
		m.channels[name] = ch
		return ch, nil
	case "whatsapp":
		ch := newWhatsAppChannel()
		m.channels[name] = ch
		return ch, nil
	case "signal":
		ch := newSignalChannel()
		m.channels[name] = ch
		return ch, nil
	case "imessage":
		ch := newIMessageChannel()
		m.channels[name] = ch
		return ch, nil
	default:
		return nil, &UnsupportedError{Name: name}
	}
}

// UnsupportedError is returned for channels not yet implemented.
type UnsupportedError struct {
	Name    string
	Message string
}

func (e *UnsupportedError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	return "reach: unsupported channel: " + e.Name
}
