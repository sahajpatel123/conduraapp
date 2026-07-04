//go:build windows

package hotkey

import (
	"strings"

	xhotkey "golang.design/x/hotkey"
)

// modifierByName maps the spec's modifier names to xhotkey.Modifier
// constants on Windows. "cmd" maps to ModWin, "alt"/"option" to ModAlt.
func modifierByName(s string) (xhotkey.Modifier, bool) {
	switch strings.ToLower(s) {
	case "cmd", "command":
		return xhotkey.ModWin, true
	case "ctrl", "control":
		return xhotkey.ModCtrl, true
	case "alt", "option", "opt":
		return xhotkey.ModAlt, true
	case "shift":
		return xhotkey.ModShift, true
	}
	return 0, false
}
