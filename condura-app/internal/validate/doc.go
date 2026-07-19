// Package validate runs a series of local health checks against
// a Condura installation. Used by `condura validate` and by
// future GUI support flows.
//
// Design constraints mirror the diag package:
//
//   - No panics on missing files. Every check returns a Result
//     with a Status field, never panics on bad input.
//   - No sensitive data. The package reports file paths and
//     status codes, not file contents.
//   - Each check is independent. A failure in one check must
//     NOT prevent the others from running — the operator gets
//     a complete picture in one CLI call.
//   - Local-only. No daemon IPC required; the package reads
//     the filesystem directly. Works even when condurad can't
//     start.
package validate
