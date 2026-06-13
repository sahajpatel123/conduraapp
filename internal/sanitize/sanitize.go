// Package sanitize provides defense-in-depth input sanitizers that
// run before the Gatekeeper evaluates an action. Sanitizers run first,
// gatekeeping runs second — passing a sanitizer never makes an action
// "safe"; it's still blast-radius/consent-gated.
//
// Five sanitizers (MISSION S10.6):
//
//	Shell: binary allowlist + arg-pattern checks
//	Path: no .., no system paths, no traversal
//	URL: SSRF prevention via resolved-IP check (RFC1918, loopback)
//	PII: regex heuristics (CC+Luhn, SSN, email, phone)
//	PythonImport: banned import/call denylist
//
// Package description is at the top of this file.
//
//nolint:revive // error sentinels are self-documenting
package sanitize

import "errors"

// Sanitizer validates and optionally modifies an input string.
type Sanitizer interface {
	Sanitize(input string) (string, error)
	Name() string
}

var (
	ErrShellDenied  = errors.New("sanitize: shell command denied")
	ErrPathDenied   = errors.New("sanitize: path denied")
	ErrURLDenied    = errors.New("sanitize: URL denied (SSRF)")
	ErrPIIDetected  = errors.New("sanitize: PII detected")
	ErrImportDenied = errors.New("sanitize: banned import")
)

// Chain runs multiple sanitizers in order.
func Chain(sanitizers []Sanitizer, input string) (string, error) {
	var err error
	for _, s := range sanitizers {
		input, err = s.Sanitize(input)
		if err != nil {
			return input, err
		}
	}
	return input, nil
}
