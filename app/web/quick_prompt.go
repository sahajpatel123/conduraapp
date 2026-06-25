package main

import (
	"strings"

	"github.com/wailsapp/wails/v2/pkg/menu"
	"github.com/wailsapp/wails/v2/pkg/menu/keys"
)

// DefaultQuickPromptHotkey is the global shortcut when the user has
// not configured hotkey.overlay during onboarding.
const DefaultQuickPromptHotkey = "Ctrl+S"

// resolveOverlayHotkey returns the configured overlay hotkey or the
// product default (Ctrl+S on every OS).
func resolveOverlayHotkey(cfg string) string {
	if strings.TrimSpace(cfg) != "" {
		return cfg
	}
	return DefaultQuickPromptHotkey
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
