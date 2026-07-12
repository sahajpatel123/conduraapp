package daemon

import (
	"errors"
	"strings"
	"testing"
)

func TestHubClientNotConfigured(t *testing.T) {
	t.Parallel()
	_, err := hubClient(&Phase12Components{})
	if err == nil {
		t.Fatal("expected error when hub client is nil")
	}
	var ipcErr interface{ Error() string }
	if !errors.As(err, &ipcErr) {
		t.Fatalf("expected ipc.Error, got %T", err)
	}
	if !strings.Contains(err.Error(), errHubNotConfigured) {
		t.Fatalf("expected %q, got %v", errHubNotConfigured, err)
	}
}

func TestHubRPCErrorMapsNetworkFailures(t *testing.T) {
	t.Parallel()
	err := hubRPCError(errors.New(`hub search: Get "https://hub.example": dial tcp: connection refused`))
	if err == nil || !strings.Contains(err.Message, errHubUnreachable) {
		t.Fatalf("expected %q, got %v", errHubUnreachable, err)
	}
}

func TestHubRPCErrorPreservesAuthMessage(t *testing.T) {
	t.Parallel()
	err := hubRPCError(errors.New("hub search: authentication required (set a token in config)"))
	if err == nil || err.Message != errHubSignInRequired {
		t.Fatalf("expected %q, got %v", errHubSignInRequired, err)
	}
}
