// Package blastradius classifies a proposed agent action by how much
// damage it could do — its blast radius — so the Gatekeeper can decide
// whether to allow it. The four classes come straight from MISSION §5.1:
// READ, WRITE, NETWORK, DESTRUCTIVE, in ascending order of risk.
//
// This is deterministic, pure-logic code with no dependencies. It is
// never a model. An action whose kind we do not recognize classifies as
// DESTRUCTIVE — the most conservative class — so the default everywhere
// is maximal caution (MISSION §2: "if a feature conflicts with the
// invariants, the feature is wrong").
package blastradius

import "strings"

// Class is the blast radius of an action, ascending in risk.
type Class int

const (
	// READ observes state without changing it: screenshots, AX reads,
	// clipboard reads, LLM completions, transcription, speech.
	READ Class = iota
	// WRITE mutates local state the user could undo: typing, pasting,
	// writing a file, a generic UI click.
	WRITE
	// NETWORK reaches outside the machine: HTTP requests, form
	// submissions, sending a message or email, clicking a link.
	NETWORK
	// DESTRUCTIVE is hard or impossible to undo: deleting files, running
	// shell commands, purchases, transfers, formatting, raw keystrokes.
	DESTRUCTIVE
)

// String renders the class for audit logs and Gatekeeper reasons.
func (c Class) String() string {
	switch c {
	case READ:
		return "READ"
	case WRITE:
		return "WRITE"
	case NETWORK:
		return "NETWORK"
	case DESTRUCTIVE:
		return "DESTRUCTIVE"
	default:
		return "DESTRUCTIVE"
	}
}

// Action describes a proposed agent action. Used by every caller
// that routes through the Gatekeeper. The payload fields are fed
// to sanitizers (command/path/URL/body), policy target-matching
// (target app/URL), and anomaly detection (coordinates).
type Action struct {
	Kind      string // e.g. "chat", "file.write", "shell.exec", "mcp.tool_call"
	TargetApp string // app/window name, e.g. "1Password", "Code"
	TargetURL string // for NETWORK actions
	Path      string // filesystem path for file ops
	Command   string // shell command text
	Body      string // message/text/code body (chat, type, email, python)
}

// classByKind maps known action kinds to their blast radius.
// Missing or empty kinds classify to DESTRUCTIVE (conservative default).
var classByKind = map[string]Class{
	// READ
	"chat":                   READ,
	"llm.complete":           READ,
	"transcribe":             READ,
	"speak":                  READ,
	"tts":                    READ,
	"screenshot.read":        READ,
	"ax.read":                READ,
	"clipboard.read":         READ,
	"file.read":              READ,
	"computeruse.read":       READ,
	"computeruse.screenshot": READ,
	"computeruse.axtree":     READ,
	// WRITE
	"file.write":         WRITE,
	"apikeys.set":        WRITE, // stores a secret; WRITE consent applies
	"apikeys.delete":     WRITE, // removes a secret; WRITE consent applies
	"policy.reload":      WRITE, // replaces the active gatekeeper policy
	"type":               WRITE,
	"paste":              WRITE,
	"clipboard.write":    WRITE,
	"click":              WRITE,
	"computeruse.click":  WRITE,
	"computeruse.type":   WRITE,
	"computeruse.scroll": WRITE,
	"computeruse.key":    WRITE,
	"computeruse.drag":   WRITE,
	"computeruse.focus":  WRITE,
	// NETWORK
	"http.request":       NETWORK,
	"form.submit":        NETWORK,
	"message.send":       NETWORK,
	"email.send":         NETWORK,
	"click.link":         NETWORK,
	"computeruse.launch": NETWORK,
	"delegation.spawn":   NETWORK,
	"reach.message.send": NETWORK,
	"reach.message.read": READ,
	// DESTRUCTIVE
	"file.delete":       DESTRUCTIVE,
	"shell.exec":        DESTRUCTIVE,
	"purchase":          DESTRUCTIVE,
	"transfer":          DESTRUCTIVE,
	"format":            DESTRUCTIVE,
	"key.send":          DESTRUCTIVE,
	"computeruse.shell": DESTRUCTIVE,
	"mcp.tool_call":     DESTRUCTIVE,
}

// Classify returns the blast radius of an action. The kind is normalized
// (trimmed and lowercased) before lookup. Unknown or empty kinds return
// DESTRUCTIVE.
func Classify(a Action) Class {
	kind := strings.ToLower(strings.TrimSpace(a.Kind))
	if c, ok := classByKind[kind]; ok {
		return c
	}
	return DESTRUCTIVE
}
