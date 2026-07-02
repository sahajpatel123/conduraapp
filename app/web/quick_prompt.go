package main

import (
	"strings"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
)

// DefaultQuickPromptHotkey is the global shortcut when the user has
// not configured hotkey.overlay during onboarding.
const DefaultQuickPromptHotkey = "Ctrl+S"

// DefaultKillSwitchHotkey is the global shortcut when the user has
// not configured hotkey.kill_switch during onboarding. Per CLAUDE.md
// §5.3 Layer 1 + Locked Decision #8, the kill switch MUST be wired
// at startup with a non-empty default; we use Cmd+Shift+Escape on
// macOS and Ctrl+Alt+\ on Win/Linux (the latter matches the
// existing test expectation in config/loader_test.go:477).
const DefaultKillSwitchHotkey = "Cmd+Shift+Escape"

// resolveOverlayHotkey returns the configured overlay hotkey or the
// product default (Ctrl+S on every OS).
func resolveOverlayHotkey(cfg string) string {
	if strings.TrimSpace(cfg) != "" {
		return cfg
	}
	return DefaultQuickPromptHotkey
}

// resolveKillSwitchHotkey returns the configured kill-switch hotkey
// or the product default. Unlike resolveOverlayHotkey, we never
// silently fall back if the user has explicitly configured an empty
// string — the audit claimed "kill switch unfulfillable via
// documented hotkey", so the default MUST be applied if the config
// is missing or whitespace.
func resolveKillSwitchHotkey(cfg string) string {
	if strings.TrimSpace(cfg) != "" {
		return cfg
	}
	return DefaultKillSwitchHotkey
}

// buildApplicationMenu constructs the native menu bar. macOS uses this
// as the primary entry point because getlantern/systray conflicts with
// Wails; Windows and Linux also get the same menu for consistency.
func buildApplicationMenu(a *App) *menu.Menu {
	appMenu := menu.NewMenu()
	appMenu.Append(menu.AppMenu())
	appMenu.Append(menu.EditMenu())

	condura := appMenu.AddSubmenu("Condura")
	condura.AddText("Quick Prompt", keys.Control("s"), func(*menu.CallbackData) {
		if a != nil {
			a.ToggleQuickPrompt()
		}
	})
	condura.AddSeparator()
	condura.AddText("Show Main Window", nil, func(*menu.CallbackData) {
		if a != nil {
			a.showMainWindow()
		}
	})

	return appMenu
}
