//go:build synaptictest

// Package daemon — test-only build hooks for buildSafetyLayer.
//
// This file is gated behind the `synaptictest` build tag so the
// SYNAPTIC_TEST_AUTO_CONSENT override (and the autoApproveConsentProvider
// type) does NOT compile into production binaries. The production
// daemon binary always uses rpcConsentProvider (the GUI SSE bridge).
//
// To run an E2E test that needs auto-approved consent, pass the
// `synaptictest` build tag:
//
//	go test -tags=synaptictest ./internal/daemon/...
//
// Production builds must never set this tag. CI does not.
//
// 2026-06-29 audit P1-1: the previous code shipped the autoApprove
// type inside the production binary and gated it only on an env
// var check. A misconfigured systemd unit, a leaked .env, or a
// packaging bug that set SYNAPTIC_TEST_AUTO_CONSENT would have made
// every gate a rubber stamp — consent granted, no human in the
// loop. The build-tag isolation makes that misconfiguration
// impossible at the binary level.
package daemon

import (
	"context"
	"log/slog"
	"os"

	"github.com/sahajpatel123/conduraapp/internal/gatekeeper"
)

// maybeAutoApproveConsent returns the autoApproveConsentProvider
// when SYNAPTIC_TEST_AUTO_CONSENT is set; otherwise nil. It is
// called from buildSafetyLayer only when compiled with the
// `synaptictest` build tag — in production builds the env-var
// check and the provider type simply do not exist.
//
// The function logs a WARN at the env-var hit so a developer who
// accidentally sets the env var sees the override in the logs.
func maybeAutoApproveConsent(log *slog.Logger) gatekeeper.ConsentProvider {
	if os.Getenv("SYNAPTIC_TEST_AUTO_CONSENT") != "" {
		log.Warn("SYNAPTIC_TEST_AUTO_CONSENT is set — gatekeeper consent will be auto-approved; this build must NEVER run in production")
		return &autoApproveConsentProvider{log: log}
	}
	return nil
}

// autoApproveConsentProvider approves every ticket immediately.
// Compiled into the binary ONLY when -tags=synaptictest is set.
// Production binaries contain neither this type nor the call
// site that constructs it.
type autoApproveConsentProvider struct {
	log *slog.Logger
}

func (p *autoApproveConsentProvider) Show(ctx context.Context, ticket *gatekeeper.ConsentTicket) (bool, error) {
	if ticket.Result == nil {
		return true, nil
	}
	select {
	case ticket.Result <- true:
		return true, nil
	case <-ctx.Done():
		return false, ctx.Err()
	}
}

func (p *autoApproveConsentProvider) IsAvailable() bool { return true }
