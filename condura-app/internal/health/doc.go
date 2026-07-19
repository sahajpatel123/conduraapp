// Package health aggregates the health of internal subsystems
// (database, secrets, providers, spend monitor, etc.) and exposes
// the result via the IPC layer and a future /healthz HTTP
// endpoint.
//
// Each subsystem registers a Check function. Snapshot()
// runs all checks concurrently and returns an aggregate
// Status. The Check contract:
//
//   - Each check returns (ok, msg, err) where ok is a boolean,
//     msg is a short human-readable description, and err is
//     a non-recoverable error.
//   - Each check is bounded by ctx — checks that time out
//     (or hang) are marked as failing in the aggregate.
//   - A check that returns ok=true does not mean the system
//     is healthy overall — only that this subsystem is OK
//     in isolation. The aggregate combines them.
//
// The CLI uses this package for the condura status and
// condura health.snapshot subcommands. The future /healthz
// HTTP endpoint will reuse the same Snapshot function so
// the CLI and HTTP have consistent health signals.
package health
