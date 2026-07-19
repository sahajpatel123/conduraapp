// Package diag produces a structured support-ticket snapshot of
// the local Condura installation. Used by `condura diag` and by
// future GUI support flows.
//
// The snapshot is gathered entirely from the local filesystem
// — no daemon IPC, no auth token, no DB connections. This
// matters: when a user's daemon won't start (or the install is
// corrupt enough that IPC can't connect), the operator still
// needs a way to dump "what does this machine look like to
// Condura?" for support.
//
// Design constraints:
//
//   - No panics on missing files. Every Stat/Read is best-effort;
//     missing artifacts are reported as null/empty in the
//     snapshot, not as errors.
//   - No sensitive data. The snapshot intentionally omits
//     secrets, OAuth tokens, and master key material. The
//     version string and config keys are fine; raw values are
//     not.
//   - Stable JSON shape. Adding fields is allowed; renaming
//     or removing is not. Support tickets stay readable across
//     versions.
package diag
