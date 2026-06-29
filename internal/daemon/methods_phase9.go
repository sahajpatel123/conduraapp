package daemon

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"os"
	"path/filepath"

	"github.com/sahajpatel123/conduraapp/internal/gatekeeper"
	"github.com/sahajpatel123/conduraapp/internal/ipc"
)

// errUnknownConsentTicket is returned when a GUI approve/deny request
// references a ticket that has already expired, been answered, or never existed.
const errUnknownConsentTicket = "unknown or expired consent ticket"

// registerSafetyMethods registers safety/consent RPC methods.
//
// These methods are the legacy surface; the canonical GUI-facing
// surface is the gatekeeper.* namespace (see methods_gatekeeper.go).
// The safety.consent.* registrations remain as DEPRECATED aliases
// for backward compatibility with external test scripts and any
// third-party callers, but new code must use gatekeeper.pending_consent /
// gatekeeper.approve / gatekeeper.deny.
//
// safety.policy.reload and safety.halt are NOT aliased — they are
// distinct concepts (policy refresh and kill switch) that the
// safety namespace owns.
func registerSafetyMethods(srv *ipc.Server, subs *Subsystems) {
	if subs.Safety == nil {
		return
	}

	// safety.consent.approve (DEPRECATED alias for gatekeeper.approve).
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

	// safety.consent.deny (DEPRECATED alias for gatekeeper.deny).
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

	// safety.consent.pending (DEPRECATED alias for gatekeeper.pending_consent).
	srv.Register("safety.consent.pending", func(_ context.Context, _ json.RawMessage) (any, error) {
		tickets := subs.Safety.Engine.Pending()
		return map[string]any{"tickets": tickets}, nil
	})

	// safety.policy.reload: reload the gatekeeper policy from disk.
	//
	// Audit 2026-06-28 fix: previously this RPC always reloaded the
	// embedded default policy (`gatekeeper.DefaultPolicy()`), which
	// contradicted MISSION.md §10.2 documenting user-editable policy
	// in `~/.condura/policy.yaml`. The fix:
	//   1. If `~/.condura/policy.yaml` exists, parse and load it.
	//   2. If parse fails, return an error (do NOT silently fall back
	//      to the default; the user's YAML is broken and they need
	//      to know).
	//   3. If the file does not exist, fall back to the embedded
	//      default (the documented "no user override" path).
	//
	// The action is classified as WRITE (added to blastradius
	// classByKind in this change) so the gatekeeper consent gate
	// applies — without that gate, an attacker with the IPC token
	// could swap in a permissive policy.
	srv.Register("safety.policy.reload", func(ctx context.Context, _ json.RawMessage) (any, error) {
		// 2026-06-29 audit P1-2: gate this RPC through the gatekeeper
		// so an attacker with the IPC token cannot swap in a permissive
		// policy. policy.reload is classified WRITE per the policy
		// class table; the engine will require consent before this
		// path can change the active policy.
		if !subs.GatekeeperAllow(ctx, "policy.reload", "ipc: safety.policy.reload") {
			return nil, &ipc.Error{
				Code:    ipc.CodeInvalidRequest,
				Message: "policy.reload denied by gatekeeper",
			}
		}
		dataDir := subs.GeneralDataDir()
		var policyPath string
		if dataDir != "" {
			policyPath = filepath.Join(dataDir, "policy.yaml")
		}
		var (
			p   *gatekeeper.Policy
			src string
		)
		if policyPath != "" {
			//nolint:gosec // G304: policyPath is server-controlled
			// (built from subs.GeneralDataDir, the data dir the
			// daemon itself created). It is not user-influenced
			// across IPC. The contents are YAML-parsed by
			// gatekeeper.LoadPolicy which validates the rules.
			if b, err := os.ReadFile(policyPath); err == nil {
				parsed, perr := gatekeeper.LoadPolicy(b)
				if perr != nil {
					return nil, &ipc.Error{
						Code:    ipc.CodeInvalidParams,
						Message: fmt.Sprintf("policy reload: parse %s: %v", policyPath, perr),
					}
				}
				p = parsed
				src = policyPath
			} else if !errors.Is(err, os.ErrNotExist) {
				return nil, &ipc.Error{
					Code:    ipc.CodeInternalError,
					Message: fmt.Sprintf("policy reload: read %s: %v", policyPath, err),
				}
			}
		}
		if p == nil {
			p = gatekeeper.DefaultPolicy()
			src = "embedded default (no ~/.condura/policy.yaml)"
		}
		subs.Safety.Engine.ReloadPolicy(p)
		slog.Info("policy reloaded", "source", src)
		return auditOK(), nil
	})

	// safety.halt: trigger the kill switch (Layer 1 flag + Layer 3 network guard).
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
		// N3: also toggle the network guard so the agent's HTTP is blocked.
		if subs.NetGuard != nil {
			_ = subs.NetGuard.Halt(p.Reason)
		}
		return auditOK(), nil
	})
}
