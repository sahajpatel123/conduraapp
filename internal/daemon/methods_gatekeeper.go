// Package daemon registers the Gatekeeper consent RPC surface.
//
// These methods give the GUI a first-class way to enumerate pending
// consent tickets and to approve or deny them. They mirror the
// safety.consent.* methods registered in methods_phase9.go but use
// the public "gatekeeper" namespace expected by the consent modal.
package daemon

import (
	"context"
	"encoding/json"

	"github.com/sahajpatel123/conduraapp/internal/ipc"
)

// registerGatekeeperMethods wires gatekeeper.* consent methods.
// If the safety layer is not initialized, the methods are not registered.
func registerGatekeeperMethods(srv *ipc.Server, subs *Subsystems) {
	if subs.Safety == nil || subs.Safety.Engine == nil {
		return
	}

	engine := subs.Safety.Engine

	// gatekeeper.pending_consent returns the list of in-flight consent
	// tickets so the GUI can render the consent modal.
	srv.Register("gatekeeper.pending_consent", func(_ context.Context, _ json.RawMessage) (any, error) {
		tickets := engine.Pending()
		return map[string]any{"tickets": tickets}, nil
	})

	// gatekeeper.approve approves a pending consent ticket by nonce.
	srv.Register("gatekeeper.approve", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Nonce string `json:"nonce"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if ok := engine.ApproveTicket(p.Nonce); !ok {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: errUnknownConsentTicket}
		}
		return auditOK(), nil
	})

	// gatekeeper.deny denies a pending consent ticket by nonce.
	srv.Register("gatekeeper.deny", func(_ context.Context, params json.RawMessage) (any, error) {
		var p struct {
			Nonce string `json:"nonce"`
		}
		if err := decodeParams(params, &p); err != nil {
			return nil, err
		}
		if ok := engine.DenyTicket(p.Nonce); !ok {
			return nil, &ipc.Error{Code: ipc.CodeInvalidParams, Message: errUnknownConsentTicket}
		}
		return auditOK(), nil
	})
}
