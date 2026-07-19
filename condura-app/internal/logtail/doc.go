// Package logtail reads the last N lines from a log file,
// including rotated siblings (condura.log.1, .2, ...).
//
// Used by `condura logs` and the (future) GUI "Recent
// errors" panel. Pure local file I/O — no daemon required.
//
// Design constraints:
//
//   - Missing log file returns (nil, nil) — the operator runs
//     this command before the daemon has ever started, and
//     a "file not found" toast is unhelpful. "No log file yet"
//     is the right answer.
//   - Newest line first. The operator wants the most recent
//     context (what just happened?), not the oldest (what
//     happened at startup?).
//   - Reads rotated siblings in time order: condura.log.1 is
//     NEWER than condura.log.2 (the rotation scheme keeps
//     lower numbers for more recent rotations).
//   - O(lines) memory. The whole point of "last N lines" is
//     bounded memory — we don't slurp the whole log file
//     into memory if it's 50MB.
package logtail
