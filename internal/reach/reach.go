// Package reach implements messaging channel integrations
// (Phase 14C). Telegram is first; WhatsApp, iMessage, and
// Signal are planned for later releases.
package reach

import "context"

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
}

// NewManager returns a Manager with the given channel store.
func NewManager(store *Store) *Manager {
	return &Manager{
		channels: make(map[string]Channel),
		store:    store,
	}
}

// Store returns the underlying channel store.
func (m *Manager) Store() *Store { return m.store }

// List returns all registered channel statuses.
func (m *Manager) List(ctx context.Context) ([]ChannelStatus, error) {
	return m.store.List(ctx)
}

// Connect establishes a connection to a channel.
func (m *Manager) Connect(ctx context.Context, name, token string) (ChannelStatus, error) {
	ch, err := m.getOrCreateChannel(name)
	if err != nil {
		return ChannelStatus{}, err
	}
	if err := ch.Connect(ctx, token); err != nil {
		return ChannelStatus{}, err
	}
	status, _ := ch.Status(ctx)
	if err := m.store.Save(ctx, name, status.ChatID, true); err != nil {
		return status, err
	}
	return status, nil
}

// Disconnect tears down a channel connection.
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
	default:
		return nil, &UnsupportedError{Name: name}
	}
}

// UnsupportedError is returned for channels not yet implemented.
type UnsupportedError struct {
	Name string
}

func (e *UnsupportedError) Error() string {
	return "reach: unsupported channel: " + e.Name
}
