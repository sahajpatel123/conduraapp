package tui

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
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
	tabHub
	tabSync
	tabSkills
	tabAudit
	tabSettings
	tabHealth
	tabCount
)

var tabNames = []string{"Chat", "Conversations", "Hub", "Sync", "Skills", "Audit", "Settings", "Health"}

type providerInfo struct {
	Name    string `json:"name"`
	Models  string `json:"models"`
	Enabled bool   `json:"enabled"`
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

	// Hub tab: search results + selected skill
	hubQuery    string
	hubResults  []map[string]any
	hubCursor   int
	hubViewport viewport.Model
	hubErr      error

	// Sync tab: status + peers + paired devices
	syncStatus   map[string]any
	syncPeers    []map[string]any
	syncPairs    []map[string]any
	syncViewport viewport.Model
	syncErr      error

	// Skills tab: locally installed skills
	skills       []map[string]any
	skillsCursor int
	skillsVP     viewport.Model

	spend spendInfo

	loading   bool
	err       error
	statusMsg string
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
	hubVP := viewport.New(80, 20)
	syncVP := viewport.New(80, 20)
	skillsVP := viewport.New(80, 20)

	return Model{
		client:        client,
		logger:        logger,
		chatInput:     ti,
		chatViewport:  chatVP,
		auditVP:       auditVP,
		cfgViewport:   cfgVP,
		healthVP:      healthVP,
		convVP:        convVP,
		hubViewport:   hubVP,
		syncViewport:  syncVP,
		skillsVP:      skillsVP,
		convCursor:    -1,
		activeTab:     tabChat,
		providers:     []providerInfo{},
		auditEvents:   []audit.Event{},
		conversations: []conversation.Meta{},
		messages:      []conversation.Message{},
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
		m.fetchHubCmd(),
		m.fetchSyncCmd(),
		m.fetchSkillsCmd(),
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
		// providers.list returns []llm.Provider which has different fields.
		// Extract name + enabled from the raw JSON.
		raw, _ := resp.Result.(json.RawMessage)
		if raw == nil {
			return loadedProvidersMsg([]providerInfo{})
		}
		var rawList []map[string]any
		if err := json.Unmarshal(raw, &rawList); err != nil {
			return loadedProvidersMsg([]providerInfo{})
		}
		provs := make([]providerInfo, 0, len(rawList))
		for _, r := range rawList {
			name, _ := r["name"].(string)
			if name == "" {
				name, _ = r["Name"].(string)
			}
			enabled, _ := r["enabled"].(bool)
			if !enabled {
				enabled, _ = r["Enabled"].(bool)
			}
			models := ""
			if m, ok := r["default_model"].(string); ok {
				models = m
			}
			provs = append(provs, providerInfo{
				Name:    name,
				Models:  models,
				Enabled: enabled,
			})
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

// fetchHubCmd queries the Skills Hub for the current query string.
// When the query is empty, returns the configured status (count
// of installed skills, hub enabled/disabled).
func (m Model) fetchHubCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if m.hubQuery == "" {
			return loadedHubMsg{results: nil, err: nil}
		}
		resp, err := m.client.Call(ctx, "hub.search", map[string]any{
			"query": m.hubQuery,
			"limit": 20,
		})
		if err != nil {
			return loadedHubMsg{err: err}
		}
		var result struct {
			Skills []map[string]any `json:"skills"`
			Total  int              `json:"total"`
		}
		if err := decodeResp(resp, &result); err != nil {
			return loadedHubMsg{err: err}
		}
		return loadedHubMsg{results: result.Skills}
	}
}

// fetchSyncCmd gets sync status, peers, and paired devices.
func (m Model) fetchSyncCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		out := loadedSyncMsg{}
		// Status
		if resp, err := m.client.Call(ctx, "sync.status", nil); err == nil {
			raw, _ := resp.Result.(json.RawMessage)
			if raw != nil {
				_ = json.Unmarshal(raw, &out.Status)
			}
		} else {
			out.Err = err
		}
		// Peers
		if resp, err := m.client.Call(ctx, "sync.peers", nil); err == nil {
			var wrap struct {
				Peers []map[string]any `json:"peers"`
			}
			if err := decodeResp(resp, &wrap); err == nil {
				out.Peers = wrap.Peers
			}
		}
		// Paired devices
		if resp, err := m.client.Call(ctx, "sync.list_pairs", nil); err == nil {
			var wrap struct {
				Devices []map[string]any `json:"devices"`
			}
			if err := decodeResp(resp, &wrap); err == nil {
				out.Pairs = wrap.Devices
			}
		}
		return out
	}
}

// fetchSkillsCmd lists locally installed skills.
func (m Model) fetchSkillsCmd() tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		resp, err := m.client.Call(ctx, "skills.list", map[string]any{"limit": 100})
		if err != nil {
			return loadedSkillsMsg{err: err}
		}
		var list []map[string]any
		if err := decodeResp(resp, &list); err != nil {
			return loadedSkillsMsg{err: err}
		}
		return loadedSkillsMsg{results: list}
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
type loadedHubMsg struct {
	results []map[string]any
	err     error
}
type loadedSyncMsg struct {
	Status map[string]any
	Peers  []map[string]any
	Pairs  []map[string]any
	Err    error
}
type loadedSkillsMsg struct {
	results []map[string]any
	err     error
}
type convCreatedMsg struct {
	meta conversation.Meta
	cmd  tea.Cmd
}

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

func (m Model) createAndSendCmd(title, content string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		// Create conversation.
		resp, err := m.client.Call(ctx, "conversations.create", map[string]any{"title": title})
		if err != nil {
			return errMsg{err}
		}
		var meta conversation.Meta
		if err := decodeResp(resp, &meta); err != nil {
			return errMsg{err}
		}
		// Send first message.
		_, err = m.client.Call(ctx, "conversations.append", map[string]any{
			"id":      meta.ID,
			"message": conversation.Message{Role: "user", Content: content},
		})
		if err != nil {
			return errMsg{err}
		}
		// Load the full conversation.
		resp2, err := m.client.Call(ctx, "conversations.get", map[string]any{"id": meta.ID})
		if err != nil {
			return errMsg{err}
		}
		var conv conversation.Conversation
		if err := decodeResp(resp2, &conv); err != nil {
			return errMsg{err}
		}
		return loadedConversationMsg(&conv)
	}
}

func (m Model) deleteConversationCmd(id int64) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		_, err := m.client.Call(ctx, "conversations.delete", map[string]any{"id": id})
		if err != nil {
			return errMsg{err}
		}
		// Refresh list.
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
		m.hubViewport.Width = msg.Width - 6
		m.hubViewport.Height = msg.Height - 8
		m.syncViewport.Width = msg.Width - 6
		m.syncViewport.Height = msg.Height - 8
		m.skillsVP.Width = msg.Width - 6
		m.skillsVP.Height = msg.Height - 8

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
		case tabHub:
			return m.updateHub(msg)
		case tabSync:
			return m.updateSync(msg)
		case tabSkills:
			return m.updateSkills(msg)
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

	case loadedHubMsg:
		m.hubResults = msg.results
		m.hubErr = msg.err
		if msg.err == nil {
			m.err = nil
		}

	case loadedSyncMsg:
		m.syncStatus = msg.Status
		m.syncPeers = msg.Peers
		m.syncPairs = msg.Pairs
		m.syncErr = msg.Err
		if msg.Err == nil {
			m.err = nil
		}

	case loadedSkillsMsg:
		m.skills = msg.results
		if msg.err != nil {
			// Not fatal — Hub may be unconfigured; keep silent.
			m.skills = nil
		}

	case chatRespMsg:
		m.statusMsg = string(msg)
		m.chatInput.SetValue("")
		if m.currentConv != nil {
			return m, m.loadConversationCmd(m.currentConv.ID)
		}

	case errMsg:
		m.err = msg.err
		m.statusMsg = ""
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
			content := m.chatInput.Value()
			m.chatInput.SetValue("")
			return m, m.createAndSendCmd(title, content)
		}
		content := m.chatInput.Value()
		m.chatInput.SetValue("")
		return m, tea.Batch(
			m.sendChatCmd(m.currentConv.ID, content),
			m.loadConversationCmd(m.currentConv.ID),
		)
	}

	return m, cmd
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
		return m, m.createAndSendCmd(title, "")
	case "d":
		if m.convCursor >= 0 && m.convCursor < len(m.conversations) {
			id := m.conversations[m.convCursor].ID
			return m, m.deleteConversationCmd(id)
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

// updateHub handles key events on the Hub tab. Pressing Enter
// fires a search using the current chat-input value; up/down
// moves the cursor; pressing 'i' on a selected skill installs it.
func (m Model) updateHub(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up":
		if m.hubCursor > 0 {
			m.hubCursor--
		}
	case "down":
		if m.hubCursor < len(m.hubResults)-1 {
			m.hubCursor++
		}
	case "enter":
		// The chat input is shared across tabs. When the user is
		// on the Hub tab, the text in the input is treated as a
		// search query. Pressing Enter triggers the search.
		query := strings.TrimSpace(m.chatInput.Value())
		if query == "" {
			return m, nil
		}
		m.hubQuery = query
		m.chatInput.SetValue("")
		return m, m.fetchHubCmd()
	case "i":
		if m.hubCursor >= 0 && m.hubCursor < len(m.hubResults) {
			id, _ := m.hubResults[m.hubCursor]["id"].(string)
			if id != "" {
				return m, m.installHubSkillCmd(id)
			}
		}
	}
	var cmd tea.Cmd
	m.hubViewport, cmd = m.hubViewport.Update(msg)
	return m, cmd
}

func (m Model) installHubSkillCmd(id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
		defer cancel()
		if _, err := m.client.Call(ctx, "hub.install", map[string]any{"id": id}); err != nil {
			return errMsg{err}
		}
		return statusMsg{text: "installed " + id}
	}
}

// updateSync handles key events on the Sync tab. Pressing 'p'
// begins pairing on the selected peer; pressing 'r' revokes.
func (m Model) updateSync(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up", "down":
		// No cursor yet — viewport scroll only.
	case "p":
		// Begin pairing on first peer (simplified UX).
		if len(m.syncPeers) > 0 {
			id, _ := m.syncPeers[0]["device_id"].(string)
			if id != "" {
				return m, m.pairSyncCmd(id)
			}
		}
	case "r":
		// Revoke first paired device.
		if len(m.syncPairs) > 0 {
			id, _ := m.syncPairs[0]["device_id"].(string)
			if id != "" {
				return m, m.revokeSyncCmd(id)
			}
		}
	}
	var cmd tea.Cmd
	m.syncViewport, cmd = m.syncViewport.Update(msg)
	return m, cmd
}

func (m Model) pairSyncCmd(deviceID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()
		if _, err := m.client.Call(ctx, "sync.pair_begin", map[string]any{"device_id": deviceID}); err != nil {
			return errMsg{err}
		}
		return statusMsg{text: "pairing initiated for " + deviceID}
	}
}

func (m Model) revokeSyncCmd(deviceID string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		if _, err := m.client.Call(ctx, "sync.revoke", map[string]any{"device_id": deviceID}); err != nil {
			return errMsg{err}
		}
		return statusMsg{text: "revoked " + deviceID}
	}
}

// updateSkills handles key events on the Skills tab. Mostly
// viewport scrolling; pressing 'd' deletes the selected skill.
func (m Model) updateSkills(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "up":
		if m.skillsCursor > 0 {
			m.skillsCursor--
		}
	case "down":
		if m.skillsCursor < len(m.skills)-1 {
			m.skillsCursor++
		}
	case "d":
		if m.skillsCursor >= 0 && m.skillsCursor < len(m.skills) {
			id, _ := m.skills[m.skillsCursor]["id"].(string)
			if id != "" {
				return m, m.deleteSkillCmd(id)
			}
		}
	}
	var cmd tea.Cmd
	m.skillsVP, cmd = m.skillsVP.Update(msg)
	return m, cmd
}

func (m Model) deleteSkillCmd(id string) tea.Cmd {
	return func() tea.Msg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		if _, err := m.client.Call(ctx, "skills.delete", map[string]any{"id": id}); err != nil {
			return errMsg{err}
		}
		return statusMsg{text: "deleted " + id}
	}
}

type statusMsg struct{ text string }
