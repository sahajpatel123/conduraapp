package daemon

import (
	"context"

	"github.com/sahajpatel123/conduraapp/condura-app/internal/ipc"
)

// gate is the canonical "ask the Gatekeeper, return an IPC error
// on deny" pattern used by every destructive or filesystem-reading
// IPC method. Replaces the 5-line pattern:
//
//	if !subs.GatekeeperAllow(ctx, action, detail) {
//	    return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: msgDeniedBySafetyPolicy}
//	}
//
// with the one-line:
//
//	if err := gate(ctx, subs, action, detail); err != nil {
//	    return nil, err
//	}
//
// The Gatekeeper's contract is fail-closed: nil Safety or nil
// Engine means GatekeeperAllow returns false, and gate returns
// the deny error. This matches the production behavior of the
// methods that previously inlined the pattern.
//
// Callers should use gate when the action is either:
//   - destructive (delete files, restore from archive, etc.), OR
//   - reads from an arbitrary filesystem path that a local IPC
//     peer could use as an arbitrary-read primitive
//     (e.g. backup.preview, backup.inspect).
//
// Plain reads from the daemon's own data dir (e.g. apikeys.list)
// do NOT need a gate — the daemon owns the data dir and the
// peer is authorized to see it via the auth token.
func gate(ctx context.Context, subs *Subsystems, action, detail string) error {
	if !subs.GatekeeperAllow(ctx, action, detail) {
		return &ipc.Error{Code: ipc.CodeInternalError, Message: msgDeniedBySafetyPolicy}
	}
	return nil
}
