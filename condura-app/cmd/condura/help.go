package main

import "fmt"

// helpFor prints per-command usage for the named subcommand.
// `condura help explain` shows the explain command's usage.
// `condura help backup` shows the backup subcommand's usage
// (and the per-subcommand list of sub-subcommands like list,
// inspect, delete, prune).
//
// The daemon-required commands (ping, version, status, config,
// llm, apikeys, hub, sync, skills, resume, i18n) all print
// "requires daemon" because they have no local-only equivalent.
// The local-only commands (diag, validate, logs, path, explain,
// env, where, doctor, backup subcommands) are flagged too.
//
// Unknown commands print a "no help for X" message that lists
// the known commands, so the operator can discover what's
// available without reading docs.
func helpFor(name string) error {
	usage, ok := commandHelp[name]
	if !ok {
		fmt.Printf("no help for %q\n", name)
		fmt.Println()
		fmt.Println("known commands:")
		for _, n := range helpOrder {
			fmt.Printf("  %-12s  %s\n", n, commandHelp[n])
		}
		return nil
	}
	fmt.Printf("usage: condura %s\n\n", usage)
	return nil
}

// helpAll prints the full cheatsheet: every command + its
// description, in tabular form. The "no help for ..." fallback
// in helpFor ALSO uses this pattern (the "known commands"
// list at the bottom of the error message) — but this
// version is the explicit "give me everything" mode.
func helpAll() error {
	fmt.Println("Condura CLI cheatsheet (all commands):")
	fmt.Println()
	for _, n := range helpOrder {
		fmt.Printf("  %-12s  %s\n", n, commandHelp[n])
	}
	return nil
}

// commandHelp is the per-command usage registry. Adding a
// new subcommand = 1 line here.
var commandHelp = map[string]string{
	"ping":     "ping                                Send a JSON-RPC ping; prints pong and a timestamp (requires daemon)",
	"version":  "version [--local]                   Print the daemon's version info (--local = CLI version, no daemon)",
	"status":   "status                              Print health snapshot, registered providers, and spend (requires daemon)",
	"config":   "config                              Print the daemon's effective config (requires daemon)",
	"llm":      "llm chat|providers                  Chat with a provider / list providers (requires daemon)",
	"apikeys":  "apikeys <list|set|delete>           Manage stored API keys (requires daemon)",
	"hub":      "hub <search|get|install|publish|serve>   Manage Skills Hub (requires daemon)",
	"sync":     "sync <peers|pair|revoke|status>     Manage P2P encrypted sync (requires daemon)",
	"skills":   "skills <list|get|delete>             Manage locally installed skills (requires daemon)",
	"resume":   "resume <request|confirm|cancel>     T3b sticky human-confirmed resume (requires daemon)",
	"i18n":     "i18n <locales|locale>                Manage locale catalogs (requires daemon)",
	"backup":   "backup <list|inspect|delete|prune>   Manage local backup archives (local-only)",
	"diag":     "diag [--json]                        Dump local diagnostic snapshot for support (local-only)",
	"validate": "validate [--json]                   Run local install health checks (local-only)",
	"logs":     "logs [--lines N] [--follow]         Read the last N lines of the daemon log (local-only)",
	"path":     "path [--exists] [--json]            Print the standard install paths (local-only)",
	"explain":  "explain <code>                      Explain an IPC error code (local-only)",
	"env":      "env [--json]                        Print the env vars that affect Condura behavior (local-only)",
	"where":    "where <key> [--list]                 Print a single install path; inverse of 'path' (local-only)",
	"doctor":   "doctor [--json] [--fix]              Run validate + print remediation steps for each failure (local-only)",
}

// helpOrder is the iteration order for the command list in
// helpFor's "no help for X" message. Matches the order in
// printUsage so the operator gets a consistent view.
var helpOrder = []string{
	"ping", "version", "status", "config",
	"llm", "apikeys", "hub", "sync", "skills", "resume", "i18n",
	"backup", "diag", "validate", "logs", "path",
	"explain", "env", "where", "doctor",
}
