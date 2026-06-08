//go:build darwin

package hotkey

import (
	"strings"

	xhotkey "golang.design/x/hotkey"
)

// modifierByName maps the spec's modifier names to xhotkey.Modifier
// constants on macOS. "cmd" maps to ModCmd, "alt"/"option" to ModOption.
func modifierByName(s string) (xhotkey.Modifier, bool) {
	switch strings.ToLower(s) {
	case "cmd", "command":
		return xhotkey.ModCmd, true
	case "ctrl", "control":
		return xhotkey.ModCtrl, true
	case "alt", "option", "opt":
		return xhotkey.ModOption, true
	case "shift":
		return xhotkey.ModShift, true
	}
	return 0, false
}
