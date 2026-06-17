package daemon

import (
	"context"
	"encoding/json"

	"github.com/sahajpatel123/synapticapp/internal/gatekeeper"
	"github.com/sahajpatel123/synapticapp/internal/ipc"
)

// errUnknownConsentTicket is returned when a GUI approve/deny request
// references a ticket that has already expired, been answered, or never existed.
const errUnknownConsentTicket = "unknown or expired consent ticket"

// registerSafetyMethods registers safety/consent RPC methods.
func registerSafetyMethods(srv *ipc.Server, subs *Subsystems) {
	if subs.Safety == nil {
		return
	}

	// safety.consent.approve: the GUI approves a pending consent ticket.
	srv.Register("safety.consent.approve", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Nonce string `json:"nonce"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if ok := subs.Safety.Engine.ApproveTicket(p.Nonce); !ok {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: errUnknownConsentTicket}
		}
		return auditOK(), nil
	})

	// safety.consent.deny: the GUI denies a pending consent ticket.
	srv.Register("safety.consent.deny", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Nonce string `json:"nonce"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if ok := subs.Safety.Engine.DenyTicket(p.Nonce); !ok {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: errUnknownConsentTicket}
		}
		return auditOK(), nil
	})

	// safety.consent.pending: list pending consent tickets for GUI enumeration.
	srv.Register("safety.consent.pending", func(_ context.Context, _ json.RawMessage) (any, error) {
		tickets := subs.Safety.Engine.Pending()
		return map[string]any{"tickets": tickets}, nil
	})

	// safety.policy.reload: reload the gatekeeper policy.
	srv.Register("safety.policy.reload", func(_ context.Context, _ json.RawMessage) (any, error) {
		p := gatekeeper.DefaultPolicy()
		subs.Safety.Engine.ReloadPolicy(p)
		return auditOK(), nil
	})

	// safety.halt: trigger the kill switch.
	srv.Register("safety.halt", func(ctx context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Reason string `json:"reason"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if _, err := subs.Halt.Halt(ctx, p.Reason); err != nil {
			return nil, &ipc.Error{Code: ipc.CodeInternalError, Message: err.Error()}
		}
		return auditOK(), nil
	})
}
