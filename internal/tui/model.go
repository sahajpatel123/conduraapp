package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"

	"github.com/sahajpatel123/synapticapp/internal/audit"
	"github.com/sahajpatel123/synapticapp/internal/config"
	"github.com/sahajpatel123/synapticapp/internal/conversation"
	"github.com/sahajpatel123/synapticapp/internal/health"
	"github.com/sahajpatel123/synapticapp/internal/ipc"
)

type viewTab int

const (
	tabChat viewTab = iota
	tabConversations
	tabAudit
	tabSettings
	tabHealth
	tabCount
)

var tabNames = []string{"Chat", "Conversations", "Audit", "Settings", "Health"}

type providerInfo struct {
	Name         string `json:"name"`
	DefaultModel string `json:"default_model"`
	Enabled      bool   `json:"enabled"`
}

type spendInfo struct {
	Spent     float64 `json:"spent"`
	Cap       float64 `json:"cap"`
	Remaining float64 `json:"remaining"`
}

type Model struct {
	client *IPCClient
	logger *slog.Logger
	ready  bool
	width  int
	height int

	activeTab viewTab

	chatInput    textinput.Model
	chatViewport viewport.Model

	convCursor int
	convVP     viewport.Model

	currentConv   *conversation.Conversation
	conversations []conversation.Meta
	messages      []conversation.Message

	auditEvents []audit.Event
	auditVP     viewport.Model

	providers   []providerInfo
	cfg         *config.Config
	cfgViewport viewport.Model

	healthSnap health.Snapshot
	healthVP   viewport.Model

	spend spendInfo

	loading   bool
	err       error
	statusMsg string
}

func FindDaemonAddr() string {
	home, err := os.UserHomeDir()
	if err != nil {
		return ""
	}
	for _, dataDir := range []string{filepath.Join(home, ".synaptic"), "/tmp/synaptic"} {
		addrFile := filepath.Join(dataDir, "synapticd.addr")
		data, err := os.ReadFile(addrFile)
		if err == nil {
			return strings.TrimSpace(string(data))
		}
		sockPath := filepath.Join(dataDir, "synapticd.sock")
		if _, err := os.Stat(sockPath); err == nil {
			return "unix://" + sockPath
		}
	}
	return ""
}

type tickMsg time.Time

func doTick() tea.Cmd {
	return tea.Tick(5*time.Second, func(t time.Time) tea.Msg {
		return tickMsg(t)
	})
}

type errMsg struct{ err error }

func InitialModel(client *IPCClient, logger *slog.Logger) Model {
	ti := textinput.New()
	ti.Placeholder = "Type a message and press Enter..."
	ti.CharLimit = 4096
	ti.Width = 80

	chatVP := viewport.New(80, 20)
	auditVP := viewport.New(80, 20)
	cfgVP := viewport.New(80, 20)
	healthVP := viewport.New(80, 20)
	convVP := viewport.New(80, 20)

	return Model{
		client:       client,
		logger:       logger,
		chatInput:    ti,
		chatViewport: chatVP,
		auditVP:      auditVP,
		cfgViewport:  cfgVP,
		healthVP:     healthVP,
		convVP:       convVP,
		convCursor:   -1,
		activeTab:    tabChat,
		providers:    []providerInfo{},
		auditEvents:  []audit.Event{},
		conversations: []conversation.Meta{},
		messages:     []conversation.Message{},
	}
}

func (m Model) Init() tea.Cmd {
	return tea.Batch(
		textinput.Blink,
		doTick(),
		m.refreshAll(),
	)
}

func (m Model) refreshAll() tea.Cmd {
	return tea.Batch(
		m.fetchConversationsCmd(),
		m.fetchAuditCmd(),
		m.fetchHealthCmd(),
		m.fetchSpendCmd(),
		m.fetchProvidersCmd(),
		m.fetchConfigCmd(),
	)
}

func (m Model) fetchConversationsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		resp, err := m.client.Call(ctx, "conversations.list", nil)
		if err != nil {
			return errMsg{err}
		}
		var list []conversation.Meta
		if err := decodeResp(resp, &list); err != nil {
			return errMsg{err}
		}
		return loadedConversationsMsg(list)
	}
}

func (m Model) fetchAuditCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		resp, err := m.client.Call(ctx, "audit.list", map[string]any{"limit": 100})
		if err != nil {
			return errMsg{err}
		}
		var events []audit.Event
		if err := decodeResp(resp, &events); err != nil {
			return errMsg{err}
		}
		return loadedAuditMsg(events)
	}
}

func (m Model) fetchHealthCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		resp, err := m.client.Call(ctx, "health.snapshot", nil)
		if err != nil {
			return errMsg{err}
		}
		var snap health.Snapshot
		if err := decodeResp(resp, &snap); err != nil {
			return errMsg{err}
		}
		return loadedHealthMsg(snap)
	}
}

func (m Model) fetchSpendCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		resp, err := m.client.Call(ctx, "spend.today", nil)
		if err != nil {
			return errMsg{err}
		}
		var s spendInfo
		if err := decodeResp(resp, &s); err != nil {
			return errMsg{err}
		}
		return loadedSpendMsg(s)
	}
}

func (m Model) fetchProvidersCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		resp, err := m.client.Call(ctx, "providers.list", nil)
		if err != nil {
			return errMsg{err}
		}
		var provs []providerInfo
		if err := decodeResp(resp, &provs); err != nil {
			return errMsg{err}
		}
		return loadedProvidersMsg(provs)
	}
}

func (m Model) fetchConfigCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		resp, err := m.client.Call(ctx, "config.get", nil)
		if err != nil {
			return errMsg{err}
		}
		var cfg config.Config
		if err := decodeResp(resp, &cfg); err != nil {
			return errMsg{err}
		}
		return loadedConfigMsg(&cfg)
	}
}

func decodeResp(resp *ipc.Response, v any) error {
	if resp.Error != nil {
		return fmt.Errorf("rpc error %d: %s", resp.Error.Code, resp.Error.Message)
	}
	raw, ok := resp.Result.(json.RawMessage)
	if ok {
		return json.Unmarshal(raw, v)
	}
	data, err := json.Marshal(resp.Result)
	if err != nil {
		return err
	}
	return json.Unmarshal(data, v)
}

type loadedConversationsMsg []conversation.Meta
type loadedAuditMsg []audit.Event
type loadedHealthMsg health.Snapshot
type loadedSpendMsg spendInfo
type loadedProvidersMsg []providerInfo
type loadedConfigMsg *config.Config
type loadedConversationMsg *conversation.Conversation
type chatRespMsg string

func (m Model) loadConversationCmd(id int64) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		resp, err := m.client.Call(ctx, "conversations.get", map[string]any{"id": id})
		if err != nil {
			return errMsg{err}
		}
		var conv conversation.Conversation
		if err := decodeResp(resp, &conv); err != nil {
			return errMsg{err}
		}
		return loadedConversationMsg(&conv)
	}
}

func (m Model) createConversationCmd(title string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		resp, err := m.client.Call(ctx, "conversations.create", map[string]any{"title": title})
		if err != nil {
			return errMsg{err}
		}
		var meta conversation.Meta
		if err := decodeResp(resp, &meta); err != nil {
			return errMsg{err}
		}
		return meta
	}
}

func (m Model) sendChatCmd(convID int64, content string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		resp, err := m.client.Call(ctx, "conversations.append", map[string]any{
			"id":      convID,
			"message": conversation.Message{Role: "user", Content: content},
		})
		if err != nil {
			return errMsg{err}
		}
		_ = resp
		return chatRespMsg("sent")
	}
}

func (m Model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.width = msg.Width
		m.height = msg.Height
		m.ready = true
		m.chatInput.Width = msg.Width - 6
		m.chatViewport.Width = msg.Width - 6
		m.chatViewport.Height = msg.Height - 8
		m.auditVP.Width = msg.Width - 6
		m.auditVP.Height = msg.Height - 8
		m.cfgViewport.Width = msg.Width - 6
		m.cfgViewport.Height = msg.Height - 8
		m.healthVP.Width = msg.Width - 6
		m.healthVP.Height = msg.Height - 8
		m.convVP.Width = msg.Width - 6
		m.convVP.Height = msg.Height - 8

	case tea.KeyMsg:
		switch msg.String() {
		case "ctrl+c":
			return m, tea.Quit
		case "tab":
			m.activeTab = (m.activeTab + 1) % tabCount
		case "shift+tab":
			m.activeTab = (m.activeTab - 1 + tabCount) % tabCount
		case "right":
			if m.activeTab == tabHealth {
				return m, nil
			}
			m.activeTab = (m.activeTab + 1) % tabCount
		case "left":
			if m.activeTab == tabChat {
				return m, nil
			}
			m.activeTab = (m.activeTab - 1 + tabCount) % tabCount
		}

		switch m.activeTab {
		case tabChat:
			return m.updateChat(msg)
		case tabConversations:
			return m.updateConversations(msg)
		case tabAudit:
			return m.updateAudit(msg)
		case tabSettings:
			return m.updateSettings(msg)
		case tabHealth:
			return m.updateHealth(msg)
		}

	case tickMsg:
		return m, tea.Batch(doTick(), m.refreshAll())

	case loadedConversationsMsg:
		m.conversations = []conversation.Meta(msg)
		m.err = nil

	case loadedAuditMsg:
		m.auditEvents = []audit.Event(msg)
		m.err = nil

	case loadedHealthMsg:
		m.healthSnap = health.Snapshot(msg)
		m.err = nil

	case loadedSpendMsg:
		m.spend = spendInfo(msg)
		m.err = nil

	case loadedProvidersMsg:
		m.providers = []providerInfo(msg)
		m.err = nil

	case loadedConfigMsg:
		m.cfg = (*config.Config)(msg)
		m.err = nil

	case loadedConversationMsg:
		m.currentConv = (*conversation.Conversation)(msg)
		m.messages = m.currentConv.Messages
		m.err = nil

	case chatRespMsg:
		m.statusMsg = string(msg)
		m.chatInput.SetValue("")
		if m.currentConv != nil {
			return m, m.loadConversationCmd(m.currentConv.ID)
		}
	}

	return m, nil
}

func (m Model) updateChat(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.chatInput, cmd = m.chatInput.Update(msg)

	if msg.Type == tea.KeyEnter && m.chatInput.Value() != "" {
		if m.currentConv == nil {
			title := m.chatInput.Value()
			if len(title) > 40 {
				title = title[:40]
			}
			return m, m.createConversationCmd(title)
		}
		content := m.chatInput.Value()
		return m, m.sendChatCmd(m.currentConv.ID, content)
	}

	return m, cmd
}

func (m Model) updateConversations(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up":
		if m.convCursor > 0 {
			m.convCursor--
		}
	case "down":
		if m.convCursor < len(m.conversations)-1 {
			m.convCursor++
		}
	case "enter":
		if m.convCursor >= 0 && m.convCursor < len(m.conversations) {
			id := m.conversations[m.convCursor].ID
			m.activeTab = tabChat
			return m, m.loadConversationCmd(id)
		}
	case "n":
		title := fmt.Sprintf("Conversation %d", len(m.conversations)+1)
		return m, m.createConversationCmd(title)
	case "d":
		if m.convCursor >= 0 && m.convCursor < len(m.conversations) {
			id := m.conversations[m.convCursor].ID
			ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()
			_, _ = m.client.Call(ctx, "conversations.delete", map[string]any{"id": id})
			return m, m.fetchConversationsCmd()
		}
	}
	return m, nil
}

func (m Model) updateAudit(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.auditVP, cmd = m.auditVP.Update(msg)
	return m, cmd
}

func (m Model) updateSettings(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.cfgViewport, cmd = m.cfgViewport.Update(msg)
	return m, cmd
}

func (m Model) updateHealth(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd
	m.healthVP, cmd = m.healthVP.Update(msg)
	return m, cmd
}
