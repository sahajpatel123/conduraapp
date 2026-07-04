// Package daemon audit + actor constants shared across method
// registration files. Centralized to satisfy linting rules (goconst)
// and to keep method registration call-sites free of magic strings.
package daemon

// Audit actor strings (who initiated the action).
const (
	actorDaemon = "daemon"
	actorGUI    = "gui"
	actorUser   = "user"
	actorSystem = "system"
	// actorIPC records that an RPC arrived over the IPC bearer channel.
	// The transport cannot distinguish the GUI from any other
	// token-holder, so "ipc" is the honest label for RPC-initiated
	// halt/resume. A privileged GUI-confirmed resume path (T3b) will
	// record a "gui-human" actor instead.
	actorIPC = "ipc"
	// actorGUIHuman records a resume that was confirmed by a human via
	// the privileged non-IPC path (T3b sticky resume).
	actorGUIHuman = "gui-human"
)

// Audit app strings (which app produced the action).
const (
	appCondurad  = "condurad"
	appConduraG  = "condura-gui"
	appSynaptic  = "condura"
	appConduraCL = "synaptic-cli"
)

// Audit level strings.
const (
	auditLevelDebug = "debug"
	auditLevelInfo  = "info"
	auditLevelWarn  = "warn"
	auditLevelError = "error"
)

// Audit result strings.
const (
	auditResultAllow = "allow"
	auditResultDeny  = "deny"
	auditResultError = "error"
)

// JSON field keys used across method registration files.
const (
	keyConversationID = "conversation_id"
)

// Shared message + provider strings (centralized for goconst).
const (
	msgDeniedBySafetyPolicy = "denied by safety policy"
	providerGoogle          = "google"
)
