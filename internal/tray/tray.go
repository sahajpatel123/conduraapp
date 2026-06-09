// Package tray wraps getlantern/systray with a Synaptic-specific menu:
// "Show/Hide Synaptic", "Pause (halt)", "Resume", "Spend today: $X.XX",
// and "Quit". The tray runs in its own goroutine and reports user
// clicks back via a small channel of events.
//
// The implementation is platform-agnostic via systray; the bundle's
// icon + title are set from the same strings the GUI uses.
//
//go:build !linux

package tray

import (
	"context"
	"fmt"
	"sync/atomic"

	"github.com/getlantern/systray"
)

// Event is a user-driven action from the tray menu.
type Event int

// Event categories. EventNone is the zero value, used to mean "no
// event pending" by the Run loop's default branch.
const (
	EventNone Event = iota
	EventShow
	EventHide
	EventToggleHalt
	EventQuit
)

// Menu holds the live systray state. Construct with New, then call
// Start. The Events channel emits one value per click. Stop the
// tray with Stop().
type Menu struct {
	title   string
	tooltip string

	events chan Event

	halted   atomic.Bool
	spend    atomic.Uint64 // cents, to keep it lock-free
	voiceStr string        // current voice state label

	mShow  *systray.MenuItem
	mHide  *systray.MenuItem
	mHalt  *systray.MenuItem
	mSpend *systray.MenuItem
	mVoice *systray.MenuItem
	mQuit  *systray.MenuItem
}

// New constructs the menu shell; Start begins the systray goroutine
// and returns immediately.
func New(title, tooltip string) *Menu {
	return &Menu{
		title:   title,
		tooltip: tooltip,
		events:  make(chan Event, 16),
	}
}

// Events returns the channel of user-driven menu actions. The
// channel is closed by Stop().
func (m *Menu) Events() <-chan Event { return m.events }

// SetHalted updates the "Pause"/"Resume" label based on the
// current halt-flag value. Safe to call from any goroutine.
func (m *Menu) SetHalted(halted bool) {
	m.halted.Store(halted)
	if m.mHalt == nil {
		return
	}
	if halted {
		m.mHalt.SetTitle("Resume")
	} else {
		m.mHalt.SetTitle("Pause (kill switch)")
	}
}

// SetSpendUSD updates the "Spend today: $X.XX" line. Cheap; the
// setter is lock-free via atomic.
func (m *Menu) SetSpendUSD(usd float64) {
	cents := uint64(usd * 100)
	m.spend.Store(cents)
	if m.mSpend != nil {
		m.mSpend.SetTitle(fmt.Sprintf("Spend today: $%.2f", usd))
	}
}

// SetTooltip updates the systray icon tooltip text.
func (m *Menu) SetTooltip(s string) {
	m.tooltip = s
	if m.mShow != nil {
		systray.SetTooltip(s)
	}
}

// SetVoiceState updates the voice status indicator in the tray.
// Valid states: "idle", "listening", "thinking", "speaking".
func (m *Menu) SetVoiceState(state string) {
	m.voiceStr = state
	if m.mVoice != nil {
		switch state {
		case "listening":
			m.mVoice.SetTitle("Voice: Listening...")
		case "thinking":
			m.mVoice.SetTitle("Voice: Thinking...")
		case "speaking":
			m.mVoice.SetTitle("Voice: Speaking...")
		default:
			m.mVoice.SetTitle("Voice: Idle")
		}
	}
}

// Start blocks; run it in a goroutine. systray.Run installs signal
// handlers and a platform event loop; calling it from the main
// goroutine is the supported way.
func (m *Menu) Start() {
	systray.Run(m.onReady, m.onExit)
}

// onReady is called by systray once the OS tray is ready. We build
// the menu items here and start a goroutine that listens for clicks.
func (m *Menu) onReady() {
	systray.SetTitle(m.title)
	systray.SetTooltip(m.tooltip)

	m.mShow = systray.AddMenuItem("Show Synaptic", "Bring the window to the front")
	m.mHide = systray.AddMenuItem("Hide Synaptic", "Hide the main window")
	systray.AddSeparator()
	m.mHalt = systray.AddMenuItem("Pause (kill switch)", "Halt all agent activity")
	systray.AddSeparator()
	m.mVoice = systray.AddMenuItem("Voice: Idle", "Current voice state")
	m.mSpend = systray.AddMenuItem("Spend today: $0.00", "Today's spend in USD")
	systray.AddSeparator()
	m.mQuit = systray.AddMenuItem("Quit", "Shut Synaptic down completely")

	go m.watchClicks()
}

// watchClicks translates systray menu clicks into Event values.
func (m *Menu) watchClicks() {
	for {
		select {
		case <-m.mShow.ClickedCh:
			m.events <- EventShow
		case <-m.mHide.ClickedCh:
			m.events <- EventHide
		case <-m.mHalt.ClickedCh:
			m.events <- EventToggleHalt
		case <-m.mQuit.ClickedCh:
			m.events <- EventQuit
			return
		}
	}
}

// onExit is called by systray when the OS is shutting the tray down.
func (m *Menu) onExit() {
	close(m.events)
}

// Stop tears down the systray. systray.Quit is the only documented
// way to ask the loop to exit; the goroutine then calls onExit,
// which closes the Events channel.
func (m *Menu) Stop() {
	systray.Quit()
}

// Run is a convenience that drives a Menu until ctx is canceled or
// the user picks Quit. handlers is the callback that turns each
// Event into a side effect (typically: window.show, window.hide,
// daemon.halt/resume, daemon.shutdown).
//
// Run blocks; call it in a goroutine. It returns when ctx is done
// OR when EventQuit is received.
func Run(ctx context.Context, m *Menu, handler func(Event)) {
	go m.Start()
	defer m.Stop()
	for {
		select {
		case <-ctx.Done():
			return
		case ev, ok := <-m.Events():
			if !ok {
				return
			}
			handler(ev)
			if ev == EventQuit {
				return
			}
		}
	}
}
