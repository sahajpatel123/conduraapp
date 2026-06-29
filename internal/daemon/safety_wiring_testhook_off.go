//go:build !synaptictest

// Package daemon — production stub for the test-only consent override.
//
// In production builds, the autoApproveConsentProvider type and the
// SYNAPTIC_TEST_AUTO_CONSENT env-var check are stripped from the
// binary. The call site in buildSafetyLayer still calls
// maybeAutoApproveConsent; this stub returns nil so production
// code never sees the override.
package daemon

import (
	"log/slog"

	"github.com/sahajpatel123/conduraapp/internal/gatekeeper"
)

func maybeAutoApproveConsent(_ *slog.Logger) gatekeeper.ConsentProvider {
	return nil
}
