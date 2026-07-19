// Package audit is an append-only, HMAC-chained audit log for
// the daemon.
//
// Every security-relevant action is recorded here with timestamp,
// actor, action, and outcome. The log powers:
//
//   - Phase 2 (sub-phase 2.6) — the audit-log viewer in the GUI.
//     The user can browse recent events, filter by actor or
//     action, and see the structured fields.
//   - Phase 11 (Trust & Recovery) — the HMAC chain and the
//     structured fields (Kind, BlastClass, Verdict,
//     TargetApp/URL/Path/Command, ConsentResult, screenshot
//     refs, SessionID) used by Action Replay to detect
//     tampering.
//
// The HMAC chain (MISSION §5.4): each row stores `prev_hash`
// (the hex SHA-256 of the previous row's hmac, or 64 zeros for
// the first row) and `hmac` (the hex SHA-256 of the canonical
// serialization of this row's payload, excluding the hmac
// column itself). Any modification to a past row invalidates
// every subsequent row's hmac.
//
// Files in this package:
//
//   - log.go: Event, Severity, Outcome types + Append/Read API.
//     The Append call computes the HMAC; the Read calls don't
//     verify the chain (that's the "replay" subsystem).
//   - test_helpers.go: helpers used by the tests across
//     multiple packages. The audit log is depended on by
//     dozens of tests (gatekeeper, agent, etc.) that need a
//     clean per-test audit log.
package audit
