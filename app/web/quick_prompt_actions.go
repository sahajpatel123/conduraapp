package main

import (
	"context"

	"github.com/sahajpatel123/conduraapp/internal/overlay"
	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// OpenQuickPrompt shows the compact always-on-top prompt. When the
// presence orchestrator is wired (normal runtime), it follows the same
// summon path as the global hotkey. Safe to call from menu, tray, or
// the Wails JS bridge.
func (a *App) OpenQuickPrompt() {
	if a.presenceOrch != nil {
		if a.presenceOrch.IsActive() {
			return
		}
		if err := a.presenceOrch.Summon(context.Background()); err != nil {
			return
		}
		a.syncQuickPromptOpen()
		return
	}
	a.openQuickPromptDirect(false)
}

// CloseQuickPrompt dismisses the quick prompt and restores the main
// window layout.
func (a *App) CloseQuickPrompt() {
	if a.presenceOrch != nil && a.presenceOrch.IsActive() {
		a.presenceOrch.Dismiss()
		a.syncQuickPromptClosed()
		return
	}
	if a.IsOverlay() {
		a.closeQuickPromptDirect()
	}
}

// ToggleQuickPrompt opens the prompt when hidden and closes it when
// visible. Bound to the application menu and available from JS.
func (a *App) ToggleQuickPrompt() {
	if a.isQuickPromptVisible() {
		a.CloseQuickPrompt()
		return
	}
	a.OpenQuickPrompt()
}

func (a *App) isQuickPromptVisible() bool {
	if a.presenceOrch != nil && a.presenceOrch.IsActive() {
		return true
	}
	return a.IsOverlay()
}

// syncQuickPromptOpen updates frontend state after the overlay
// controller has already been shown by the presence orchestrator or
// openQuickPromptDirect.
func (a *App) syncQuickPromptOpen() {
	a.overlay.Store(true)
	a.focusMainWindow()
	a.emitOverlayState(true)
}

func (a *App) syncQuickPromptClosed() {
	a.overlay.Store(false)
	a.emitOverlayState(false)
}

func (a *App) openQuickPromptDirect(atCursor bool) {
	a.overlay.Store(true)
	a.focusMainWindow()
	if a.overlayCtrl != nil {
		_ = a.overlayCtrl.Show(context.Background(), overlay.ShowOpts{AtCursor: atCursor})
	}
	a.emitOverlayState(true)
}

func (a *App) closeQuickPromptDirect() {
	a.overlay.Store(false)
	if a.overlayCtrl != nil {
		_ = a.overlayCtrl.Hide()
	}
	a.emitOverlayState(false)
}

func (a *App) showMainWindow() {
	if a.IsOverlay() {
		a.CloseQuickPrompt()
	}
	a.focusMainWindow()
}

func (a *App) focusMainWindow() {
	if a.ctx == nil {
		return
	}
	wailsruntime.WindowShow(a.ctx)
	wailsruntime.WindowUnminimise(a.ctx)
}

func (a *App) emitOverlayState(active bool) {
	if a.ctx == nil {
		return
	}
	wailsruntime.EventsEmit(a.ctx, "condura:overlay", map[string]interface{}{
		"active": active,
	})
}
