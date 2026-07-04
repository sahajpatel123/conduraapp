package logger

import (
	"bytes"
	"context"
	"encoding/json"
	"log/slog"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// -----------------------------------------------------------------------------
// ParseLevel
// -----------------------------------------------------------------------------

func TestParseLevel(t *testing.T) {
	tests := []struct {
		in   string
		want Level
	}{
		{"debug", LevelDebug},
		{"DEBUG", LevelDebug},
		{"info", LevelInfo},
		{"", LevelInfo},
		{"warn", LevelWarn},
		{"warning", LevelWarn},
		{"error", LevelError},
		{"err", LevelError},
		{"trace", LevelInfo}, // unknown -> info
		{"  info  ", LevelInfo},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, ParseLevel(tt.in))
		})
	}
}

func TestParseFormat(t *testing.T) {
	tests := []struct {
		in   string
		want Format
	}{
		{"json", FormatJSON},
		{"JSON", FormatJSON},
		{"text", FormatText},
		{"", FormatText},
		{"xml", FormatText}, // unknown -> text
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			assert.Equal(t, tt.want, ParseFormat(tt.in))
		})
	}
}

// -----------------------------------------------------------------------------
// New
// -----------------------------------------------------------------------------

func TestNew_Defaults(t *testing.T) {
	lg := New(Config{})
	require.NotNil(t, lg)
	// Should not panic; just ensure it returns a valid logger.
	assert.NotNil(t, lg)
}

func TestNew_JSON(t *testing.T) {
	var buf bytes.Buffer
	// Build a logger that writes to our buffer via a JSON handler.
	var handler slog.Handler = slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	handler = newRedactingHandler(handler)
	lg := slog.New(handler)

	lg.Info("hello", "key", "value")
	out := buf.String()
	assert.Contains(t, out, `"msg":"hello"`)
	assert.Contains(t, out, `"key":"value"`)
}

// -----------------------------------------------------------------------------
// Redaction
// -----------------------------------------------------------------------------

func newJSONLoggerWithRedaction(buf *bytes.Buffer, lvl slog.Level) *slog.Logger {
	var h slog.Handler = slog.NewJSONHandler(buf, &slog.HandlerOptions{Level: lvl})
	h = newRedactingHandler(h)
	return slog.New(h)
}

func TestRedact_SensitiveKeys(t *testing.T) {
	tests := []struct {
		key  string
		want string
	}{
		{"api_key", "REDACTED"},
		{"API_KEY", "REDACTED"},
		{"apiKey", "REDACTED"},
		{"Api-Key", "REDACTED"},
		{"Authorization", "REDACTED"},
		{"authorization", "REDACTED"},
		{"password", "REDACTED"},
		{"secret", "REDACTED"},
		{"access_token", "REDACTED"},
		{"refresh_token", "REDACTED"},
		{"client_secret", "REDACTED"},
		{"private_key", "REDACTED"},
		{"session_token", "REDACTED"},
		{"bearer", "REDACTED"},
		{"cookie", "REDACTED"},
		{"x-api-key", "REDACTED"},
		{"encryption_key", "REDACTED"},
	}

	for _, tt := range tests {
		t.Run(tt.key, func(t *testing.T) {
			var buf bytes.Buffer
			lg := newJSONLoggerWithRedaction(&buf, slog.LevelInfo)
			lg.Info("test", tt.key, "supersecretvalue123")
			out := buf.String()
			assert.NotContains(t, out, "supersecretvalue123", "value should be redacted for key %q", tt.key)
			assert.Contains(t, out, "[REDACTED]", "output should contain redaction marker for key %q", tt.key)
		})
	}
}

func TestRedact_NonSensitiveKeys(t *testing.T) {
	var buf bytes.Buffer
	lg := newJSONLoggerWithRedaction(&buf, slog.LevelInfo)
	lg.Info("test", "username", "alice", "email", "alice@example.com", "model", "claude-sonnet-4-5")
	out := buf.String()
	assert.Contains(t, out, "alice")
	assert.Contains(t, out, "alice@example.com")
	assert.Contains(t, out, "claude-sonnet-4-5")
}

func TestRedact_SensitiveValues(t *testing.T) {
	tests := []struct {
		name  string
		key   string
		value string
	}{
		{"OpenAI key", "config", "sk-abcdefghijklmnopqrstuvwxyz1234567890"},
		{"Anthropic key", "config", "sk-ant-api03-abcdefghijklmnopqrstuvwxyz1234567890"},
		{"Google key", "config", "AIzaSyAbcdefghijklmnopqrstuvwxyz1234567"},
		{"GitHub PAT", "config", "ghp_abcdefghijklmnopqrstuvwxyz1234567890"},
		{"Slack token", "config", "xoxb-1234567890-abcdefghijklmnopqrstuvwx"},
		{"AWS key", "config", "AKIAIOSFODNN7EXAMPLE"},
		{"JWT", "config", "eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiIxMjM0NTY3ODkwIn0.SflKxwRJSMeKKF2QT4fwpMeJf36POk6yJV_adQssw5c"},
		{"Bearer in plain value", "config", "Bearer abcdefghijklmnopqrstuvwxyz1234567890"},
		{"PEM private key", "config", "-----BEGIN RSA PRIVATE KEY-----"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var buf bytes.Buffer
			lg := newJSONLoggerWithRedaction(&buf, slog.LevelInfo)
			lg.Info("test", tt.key, tt.value)
			out := buf.String()
			assert.NotContains(t, out, tt.value, "sensitive value should be redacted")
			assert.Contains(t, out, "[REDACTED]")
		})
	}
}

func TestRedact_Group(t *testing.T) {
	var buf bytes.Buffer
	lg := newJSONLoggerWithRedaction(&buf, slog.LevelInfo)
	lg.Info("test", slog.Group("auth",
		slog.String("username", "alice"),
		slog.String("password", "hunter2"),
		slog.String("api_key", "sk-abcdefghijklmnopqrstuvwxyz1234567890"),
	))
	out := buf.String()
	// Parsed JSON check
	var rec map[string]any
	require.NoError(t, json.Unmarshal([]byte(strings.TrimSpace(out)), &rec))
	auth, ok := rec["auth"].(map[string]any)
	require.True(t, ok, "auth should be a group: %v", rec)
	assert.Equal(t, "alice", auth["username"])
	assert.Equal(t, "[REDACTED]", auth["password"])
	assert.Equal(t, "[REDACTED]", auth["api_key"])
}

func TestRedact_WithAttrs(t *testing.T) {
	var buf bytes.Buffer
	var h slog.Handler = slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	h = newRedactingHandler(h)
	lg := slog.New(h).With("api_key", "supersecret", "user", "alice")
	lg.Info("test")
	out := buf.String()
	assert.NotContains(t, out, "supersecret")
	assert.Contains(t, out, "alice")
}

func TestRedact_WithGroup(t *testing.T) {
	var buf bytes.Buffer
	var h slog.Handler = slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	h = newRedactingHandler(h)
	lg := slog.New(h).WithGroup("cfg").With("api_key", "supersecret", "model", "claude-sonnet-4-5")
	lg.Info("test")
	out := buf.String()
	assert.NotContains(t, out, "supersecret")
	assert.Contains(t, out, "claude-sonnet-4-5")
}

func TestRedact_Disabled(t *testing.T) {
	var buf bytes.Buffer
	var h slog.Handler = slog.NewJSONHandler(&buf, &slog.HandlerOptions{Level: slog.LevelInfo})
	// No redaction wrapping.
	lg := slog.New(h)
	lg.Info("test", "api_key", "supersecret")
	out := buf.String()
	assert.Contains(t, out, "supersecret", "without redaction, value should pass through")
}

// -----------------------------------------------------------------------------
// Levels
// -----------------------------------------------------------------------------

func TestLevel_Filtering(t *testing.T) {
	var buf bytes.Buffer
	lg := newJSONLoggerWithRedaction(&buf, slog.LevelWarn)
	lg.Debug("d")
	lg.Info("i")
	lg.Warn("w")
	lg.Error("e")
	out := buf.String()
	assert.NotContains(t, out, `"msg":"d"`)
	assert.NotContains(t, out, `"msg":"i"`)
	assert.Contains(t, out, `"msg":"w"`)
	assert.Contains(t, out, `"msg":"e"`)
}

// -----------------------------------------------------------------------------
// Context-aware logging
// -----------------------------------------------------------------------------

func TestContextHelpers(t *testing.T) {
	// Just ensure the helpers exist and don't panic.
	ctx := context.Background()
	InfoContext(ctx, "test", "key", "value")
	DebugContext(ctx, "test", "key", "value")
	WarnContext(ctx, "test", "key", "value")
	ErrorContext(ctx, "test", "key", "value")
}

func TestWith(t *testing.T) {
	lg := With("request_id", "abc123")
	require.NotNil(t, lg)
}

func TestWithGroup(t *testing.T) {
	lg := WithGroup("rpc")
	require.NotNil(t, lg)
}

// -----------------------------------------------------------------------------
// Default
// -----------------------------------------------------------------------------

func TestDefault(t *testing.T) {
	lg := Default()
	require.NotNil(t, lg)

	original := Default()
	defer SetDefault(original)

	custom := New(Config{Level: LevelDebug, Format: FormatJSON})
	SetDefault(custom)
	assert.Equal(t, custom, Default())
}
