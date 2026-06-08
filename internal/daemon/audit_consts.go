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
)

// Audit app strings (which app produced the action).
const (
	appSynapticd  = "synapticd"
	appSynapticG  = "synaptic-gui"
	appSynaptic   = "synaptic"
	appSynapticCL = "synaptic-cli"
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
