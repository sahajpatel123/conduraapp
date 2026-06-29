package tui

import (
	"fmt"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"github.com/sahajpatel123/conduraapp/internal/health"
)

var (
	titleStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 1)

	activeTabStyle = lipgloss.NewStyle().
			Bold(true).
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#25A065")).
			Padding(0, 2)

	inactiveTabStyle = lipgloss.NewStyle().
				Foreground(lipgloss.Color("#25A065")).
				Padding(0, 2)

	statusBarStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FFFDF5")).
			Background(lipgloss.Color("#3C3C3C")).
			Padding(0, 1)

	infoStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#A6D1FA"))

	warnStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F5A623"))

	errorStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#FF6B6B"))

	successStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#25A065"))

	dimStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	helpStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#888888"))

	chatUserStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#5B8DEF"))

	chatAsstStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#25A065"))

	chatSystemStyle = lipgloss.NewStyle().
			Foreground(lipgloss.Color("#F5A623"))

	borderStyle = lipgloss.NewStyle().
			Border(lipgloss.RoundedBorder()).
			BorderForeground(lipgloss.Color("#3C3C3C")).
			Padding(0, 1)
)

func (m Model) View() string {
	if !m.ready {
		return "Connecting to daemon..."
	}

	var b strings.Builder

	b.WriteString(m.headerView())
	b.WriteByte('\n')

	content := m.activeView()
	b.WriteString(content)
	b.WriteByte('\n')

	b.WriteString(m.statusBar())
	return b.String()
}

func (m Model) headerView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" Synaptic TUI "))
	b.WriteByte(' ')

	for i := 0; i < int(tabCount); i++ {
		tab := viewTab(i)
		label := tabNames[i]
		if tab == m.activeTab {
			b.WriteString(activeTabStyle.Render(label))
		} else {
			b.WriteString(inactiveTabStyle.Render(label))
		}
		b.WriteByte(' ')
	}
	return b.String()
}

func (m Model) activeView() string {
	switch m.activeTab {
	case tabChat:
		return m.chatView()
	case tabConversations:
		return m.conversationsView()
	case tabHub:
		return m.hubView()
	case tabSync:
		return m.syncView()
	case tabSkills:
		return m.skillsView()
	case tabAudit:
		return m.auditView()
	case tabSettings:
		return m.settingsView()
	case tabHealth:
		return m.healthView()
	}
	return "unknown tab"
}

// hubView renders the Skills Hub tab. Shows search results and
// keyboard hints.
func (m Model) hubView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" Skills Hub "))
	b.WriteByte('\n')
	if m.hubErr != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("error: %v", m.hubErr)))
		b.WriteByte('\n')
	} else if len(m.hubResults) == 0 {
		b.WriteString(dimStyle.Render("Type a query in the chat box then press '/' on this tab to search."))
		b.WriteByte('\n')
		b.WriteString(dimStyle.Render("(Hub may be disabled in config — see `synaptic config`.)"))
	} else {
		for i, r := range m.hubResults {
			cursor := "  "
			if i == m.hubCursor {
				cursor = "▸ "
			}
			name, _ := r["name"].(string)
			ver, _ := r["version"].(string)
			author, _ := r["author"].(string)
			b.WriteString(fmt.Sprintf("%s%s v%s by %s\n", cursor, name, ver, author))
		}
	}
	b.WriteByte('\n')
	b.WriteString(dimStyle.Render("up/down: navigate · i: install selected · /: search"))
	return b.String()
}

// syncView renders the P2P sync tab. Shows status, peers, and
// paired devices.
func (m Model) syncView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" P2P Sync (encrypted) "))
	b.WriteByte('\n')

	if m.syncErr != nil {
		b.WriteString(errorStyle.Render(fmt.Sprintf("error: %v", m.syncErr)))
		b.WriteByte('\n')
	}

	if m.syncStatus != nil {
		enabled, _ := m.syncStatus["enabled"].(bool)
		if !enabled {
			b.WriteString(dimStyle.Render("sync not enabled in config (set sync.enabled: true)"))
			b.WriteByte('\n')
		}
		if id, ok := m.syncStatus["device_id"].(string); ok {
			b.WriteString(fmt.Sprintf("device_id: %s\n", id))
		}
		if n, ok := m.syncStatus["entries"].(float64); ok {
			b.WriteString(fmt.Sprintf("entries:   %d\n", int(n)))
		}
	}

	b.WriteByte('\n')
	b.WriteString(infoStyle.Render(" Discovered peers:"))
	b.WriteByte('\n')
	if len(m.syncPeers) == 0 {
		b.WriteString(dimStyle.Render("  (no peers on LAN)"))
		b.WriteByte('\n')
	} else {
		for _, p := range m.syncPeers {
			id, _ := p["device_id"].(string)
			name, _ := p["name"].(string)
			b.WriteString(fmt.Sprintf("  %s (%s)\n", name, id))
		}
	}
	b.WriteByte('\n')
	b.WriteString(infoStyle.Render(" Paired devices:"))
	b.WriteByte('\n')
	if len(m.syncPairs) == 0 {
		b.WriteString(dimStyle.Render("  (no paired devices — use `synaptic sync pair <id>`)"))
		b.WriteByte('\n')
	} else {
		for _, p := range m.syncPairs {
			id, _ := p["device_id"].(string)
			name, _ := p["device_name"].(string)
			b.WriteString(fmt.Sprintf("  %s (%s)\n", name, id))
		}
	}
	b.WriteByte('\n')
	b.WriteString(dimStyle.Render("p: pair first peer · r: revoke first paired device"))
	return b.String()
}

// skillsView renders the locally installed skills tab.
func (m Model) skillsView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" Installed Skills "))
	b.WriteByte('\n')
	if len(m.skills) == 0 {
		b.WriteString(dimStyle.Render("(no skills installed — use the Hub tab)"))
		b.WriteByte('\n')
	} else {
		for i, s := range m.skills {
			cursor := "  "
			if i == m.skillsCursor {
				cursor = "▸ "
			}
			name, _ := s["name"].(string)
			ver, _ := s["version"].(string)
			trust, _ := s["trust"].(string)
			b.WriteString(fmt.Sprintf("%s%s %s [%s]\n", cursor, name, ver, trust))
		}
	}
	b.WriteByte('\n')
	b.WriteString(dimStyle.Render("up/down: navigate · d: delete selected"))
	return b.String()
}

func (m Model) chatView() string {
	var b strings.Builder

	convTitle := "Chat"
	if m.currentConv != nil && m.currentConv.Title != "" {
		convTitle = m.currentConv.Title
	}
	b.WriteString(titleStyle.Render(" " + convTitle + " "))
	b.WriteByte('\n')

	content := ""
	for _, msg := range m.messages {
		style := chatUserStyle
		label := "user"
		switch msg.Role {
		case "assistant":
			label = "assistant"
			style = chatAsstStyle
		case "system":
			label = "system"
			style = chatSystemStyle
		}
		text := msg.Content
		if len(text) > 200 {
			text = text[:200] + "..."
		}
		content += style.Render(label+":") + " " + text + "\n\n"
	}
	if content == "" {
		content = dimStyle.Render("No messages yet.")
	}

	m.chatViewport.SetContent(content)
	b.WriteString(borderStyle.Width(m.width - 4).Render(m.chatViewport.View()))
	b.WriteByte('\n')

	if m.currentConv == nil {
		b.WriteString(helpStyle.Render("Type a message to create a new conversation"))
		b.WriteByte('\n')
	}
	b.WriteString(m.chatInput.View())
	b.WriteByte('\n')
	b.WriteString(helpStyle.Render("Tab/←→: switch view  Enter: send  Ctrl+C: quit"))

	return b.String()
}

func (m Model) conversationsView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" Conversations "))
	b.WriteByte('\n')

	if len(m.conversations) == 0 {
		b.WriteString(dimStyle.Render("  No conversations yet."))
		return b.String()
	}

	for i, conv := range m.conversations {
		cursor := "  "
		prefix := dimStyle.Render("-")
		if i == m.convCursor {
			cursor = "> "
			prefix = successStyle.Render("●")
		}
		b.WriteString(fmt.Sprintf("%s%s [%d] %s",
			cursor, prefix, conv.ID, conv.Title))
		b.WriteString(dimStyle.Render(fmt.Sprintf(" (%d msgs)", conv.MessageCount)))
		b.WriteByte('\n')
	}

	b.WriteByte('\n')
	b.WriteString(helpStyle.Render("  ↑/↓: navigate  Enter: open  n: new  d: delete"))
	b.WriteByte('\n')
	b.WriteString(helpStyle.Render("  Tab/←→: switch view  Ctrl+C: quit"))

	return b.String()
}

func (m Model) auditView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" Audit Log "))
	b.WriteByte('\n')

	if len(m.auditEvents) == 0 {
		b.WriteString(dimStyle.Render("  No audit events."))
		return b.String()
	}

	content := ""
	for _, ev := range m.auditEvents {
		levelStyle := infoStyle
		switch ev.Level {
		case "warn", "warning":
			levelStyle = warnStyle
		case "error", "critical":
			levelStyle = errorStyle
		}
		line := fmt.Sprintf("[%s] %s %s/%s → %s",
			ev.TS.Format("15:04:05"),
			levelStyle.Render(ev.Level),
			ev.Actor, ev.Action, ev.Result)
		if ev.Message != "" {
			line += " " + dimStyle.Render(ev.Message)
		}
		content += line + "\n"
	}

	m.auditVP.SetContent(content)
	b.WriteString(borderStyle.Width(m.width - 4).Render(m.auditVP.View()))
	b.WriteByte('\n')
	b.WriteString(helpStyle.Render("  ↑/↓: scroll  Tab/←→: switch view  Ctrl+C: quit"))

	return b.String()
}

func (m Model) settingsView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" Settings "))
	b.WriteByte('\n')

	b.WriteString(successStyle.Render("Providers") + "\n")
	if len(m.providers) == 0 {
		b.WriteString(dimStyle.Render("  No providers configured.\n"))
	} else {
		for _, p := range m.providers {
			status := dimStyle.Render("disabled")
			if p.Enabled {
				status = successStyle.Render("enabled")
			}
			model := p.Models
			if model == "" {
				model = "default"
			}
			b.WriteString(fmt.Sprintf("  %s: %s [%s]\n", p.Name, model, status))
		}
	}

	b.WriteByte('\n')
	b.WriteString(successStyle.Render("Spend Today") + "\n")
	b.WriteString(fmt.Sprintf("  Spent: $%.4f / $%.2f (remaining: $%.4f)\n",
		m.spend.Spent, m.spend.Cap, m.spend.Remaining))

	b.WriteByte('\n')

	if m.cfg != nil {
		content := fmt.Sprintf("Configuration\n")
		content += fmt.Sprintf("  Data Dir: %s\n", m.cfg.General.DataDir)
		content += fmt.Sprintf("  Log Level: %s\n", m.cfg.Logging.Level)
		content += fmt.Sprintf("  Telemetry Enabled: %v\n", m.cfg.Telemetry.Enabled)
		content += fmt.Sprintf("  Spend Limit/Day: $%.2f\n", m.cfg.Security.SpendLimitUSDPerDay)
		content += fmt.Sprintf("  Voice Enabled: %v\n", m.cfg.Voice.Enabled)
		m.cfgViewport.SetContent(content)
		b.WriteString(borderStyle.Width(m.width - 4).Render(m.cfgViewport.View()))
	}

	b.WriteByte('\n')
	b.WriteString(helpStyle.Render("  Tab/←→: switch view  Ctrl+C: quit"))

	return b.String()
}

func (m Model) healthView() string {
	var b strings.Builder
	b.WriteString(titleStyle.Render(" Health "))
	b.WriteByte('\n')

	state := m.healthSnap.State
	stateColor := successStyle
	stateLabel := string(state)
	switch state {
	case health.StateDegraded:
		stateColor = warnStyle
	case health.StateDown:
		stateColor = errorStyle
	}
	b.WriteString(fmt.Sprintf("Overall: %s\n\n", stateColor.Render(stateLabel)))

	content := ""
	for _, r := range m.healthSnap.Results {
		rStyle := successStyle
		switch r.State {
		case health.StateDegraded:
			rStyle = warnStyle
		case health.StateDown:
			rStyle = errorStyle
		}
		line := fmt.Sprintf("  %s: %s (%dms)", r.Name, rStyle.Render(string(r.State)), r.TookMs)
		if r.Error != "" {
			line += " " + errorStyle.Render(r.Error)
		}
		content += line + "\n"
	}
	m.healthVP.SetContent(content)
	b.WriteString(borderStyle.Width(m.width - 4).Render(m.healthVP.View()))
	b.WriteByte('\n')
	b.WriteString(helpStyle.Render("  ↑/↓: scroll  Tab/←→: switch view  Ctrl+C: quit"))

	return b.String()
}

func (m Model) statusBar() string {
	statusText := "Ready"
	if m.loading {
		statusText = "Loading..."
	} else if m.err != nil {
		statusText = fmt.Sprintf("Error: %v", m.err)
	} else if m.statusMsg != "" {
		statusText = m.statusMsg
	}

	help := "Tab/←→: switch  Enter: send  Ctrl+C: quit"
	return statusBarStyle.Render(fmt.Sprintf(" %s | %s", statusText, help))
}
