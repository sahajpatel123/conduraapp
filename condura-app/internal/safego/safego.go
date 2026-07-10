// Package safego runs goroutines with crash.Recover so a panic in
// background work cannot take down the daemon process.
//
// Prefer safego.Go over bare `go` for production background work.
// Named long-lived entrypoints (e.g. Watchdog.Run) may still call
// defer crash.Recover() at the top of the method for defense-in-depth.
package safego

import "github.com/sahajpatel123/conduraapp/condura-app/internal/crash"

// Go starts fn in a new goroutine with defer crash.Recover().
func Go(fn func()) {
	if fn == nil {
		return
	}
	go func() {
		defer crash.Recover()
		fn()
	}()
}
