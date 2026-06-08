// Package hotkey spec parsing. Internal to the hotkey package; not
// part of the public surface.
//
// Supported spec format (human-readable):
//
//	"Cmd+Shift+K", "Ctrl+Alt+\\", "Cmd+Shift+Space"
//
// We accept the following modifier names (case-insensitive):
//
//	cmd, ctrl (alias: control), alt (alias: option), shift, win (alias: super, meta)
//
// For the key, we accept the printable character or a name from the
// table below. We do NOT support the "Hyper"/"Fn" macOS modifiers.
//
//go:build !linux

package hotkey

import (
	"fmt"
	"strings"

	xhotkey "golang.design/x/hotkey"
)

// ParseSpec parses a spec like "Cmd+Shift+Space" into the matching
// golang.design/x/hotkey modifier set and key code.
//
// Returns an error if the spec is empty, contains an unknown
// modifier, or names a key that isn't in our small supported set.
func ParseSpec(spec string) (mods []xhotkey.Modifier, key xhotkey.Key, err error) {
	spec = strings.TrimSpace(spec)
	if spec == "" {
		return nil, 0, fmt.Errorf("empty spec")
	}
	parts := strings.Split(spec, "+")
	if len(parts) < 2 {
		return nil, 0, fmt.Errorf("spec %q must contain at least one modifier + a key", spec)
	}
	// Last part is the key; everything before is a modifier.
	keyName := strings.TrimSpace(parts[len(parts)-1])
	key, ok := keyByName(keyName)
	if !ok {
		return nil, 0, fmt.Errorf("unknown key %q", keyName)
	}
	for _, raw := range parts[:len(parts)-1] {
		m, ok := modifierByName(strings.TrimSpace(raw))
		if !ok {
			return nil, 0, fmt.Errorf("unknown modifier %q", raw)
		}
		mods = append(mods, m)
	}
	return mods, key, nil
}

// modifierByName maps the spec's modifier names to xhotkey.Modifier
// constants. Aliases are accepted (e.g. "control" -> ctrl).
//
// We deliberately do NOT map "win"/"super"/"meta" to a constant —
// golang.design/x/hotkey does not export a Win/Super modifier in the
// version we depend on, so we reject those names explicitly.
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

// namedKeys is the set of friendly key names that the spec parser
// recognizes. Anything else must be a single printable ASCII rune.
var namedKeys = map[string]xhotkey.Key{
	"space":  xhotkey.KeySpace,
	"esc":    xhotkey.KeyEscape,
	"escape": xhotkey.KeyEscape,
	"tab":    xhotkey.KeyTab,
	"enter":  xhotkey.KeyReturn,
	"return": xhotkey.KeyReturn,
	"delete": xhotkey.KeyDelete,
	"del":    xhotkey.KeyDelete,
	"left":   xhotkey.KeyLeft,
	"right":  xhotkey.KeyRight,
	"up":     xhotkey.KeyUp,
	"down":   xhotkey.KeyDown,
	"f1":     xhotkey.KeyF1,
	"f2":     xhotkey.KeyF2,
	"f3":     xhotkey.KeyF3,
	"f4":     xhotkey.KeyF4,
	"f5":     xhotkey.KeyF5,
	"f6":     xhotkey.KeyF6,
	"f7":     xhotkey.KeyF7,
	"f8":     xhotkey.KeyF8,
	"f9":     xhotkey.KeyF9,
	"f10":    xhotkey.KeyF10,
	"f11":    xhotkey.KeyF11,
	"f12":    xhotkey.KeyF12,
}

// keyByName maps a small set of named keys + any single printable
// rune to xhotkey.Key. We deliberately do NOT accept every printable
// rune — that would make the function the target of weird inputs.
//
// Note: golang.design/x/hotkey only exports a limited set of named
// constants (A-Z, digits, Space, Escape, Tab, Return, Delete, the
// arrows, and the F-keys). For everything else we fall through to
// "single printable ASCII rune" and cast to xhotkey.Key — the
// underlying Carbon/HID codes work for letters, digits, and the
// common punctuation characters in the same range.
func keyByName(name string) (xhotkey.Key, bool) {
	if k, ok := namedKeys[strings.ToLower(name)]; ok {
		return k, true
	}
	// Accept a single printable ASCII rune (e.g. "K", "\\", "0").
	// golang.design/x/hotkey treats this as a keycode on macOS/Windows
	// and as a keysym on Linux.
	if len(name) == 1 {
		c := name[0]
		if c >= 0x20 && c <= 0x7E {
			return xhotkey.Key(c), true
		}
	}
	return 0, false
}
