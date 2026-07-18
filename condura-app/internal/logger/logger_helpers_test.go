package logger

import (
	"bytes"
	"log/slog"
	"strings"
	"testing"
)

// TestDebug_DelegatesToDefault pins the package-level Debug helper:
// `Debug(msg, args...)` MUST route to whatever logger SetDefault()
// installed. The existing TestContextHelpers (for the *Context
// variants) only checks that the call doesn't panic; this stronger
// pin verifies the message actually lands in the configured default
// logger's output.
//
// A regression where Debug wrote to a hard-coded logger (ignoring
// SetDefault) would silently route debug output to the wrong
// destination, and callers using SetDefault to redirect logs would
// lose their log lines.
func TestDebug_DelegatesToDefault(t *testing.T) {
	var buf bytes.Buffer
	lg := newJSONLoggerWithRedaction(&buf, slog.LevelDebug)
	original := Default()
	t.Cleanup(func() { SetDefault(original) })
	SetDefault(lg)

	Debug("test-debug-message", "key1", "value1")
	out := buf.String()

	if !strings.Contains(out, `"msg":"test-debug-message"`) {
		t.Errorf("Debug did not route to SetDefault'd logger; output:\n  %s\nwant substring %q", out, `"msg":"test-debug-message"`)
	}
	if !strings.Contains(out, `"key1":"value1"`) {
		t.Errorf("Debug did not pass args through; output:\n  %s", out)
	}
}

// TestInfo_DelegatesToDefault — same contract as TestDebug_... but
// for the Info level.
func TestInfo_DelegatesToDefault(t *testing.T) {
	var buf bytes.Buffer
	lg := newJSONLoggerWithRedaction(&buf, slog.LevelInfo)
	original := Default()
	t.Cleanup(func() { SetDefault(original) })
	SetDefault(lg)

	Info("test-info-message", "k", "v")
	out := buf.String()

	if !strings.Contains(out, `"msg":"test-info-message"`) {
		t.Errorf("Info did not route to SetDefault'd logger; output:\n  %s", out)
	}
	if !strings.Contains(out, `"k":"v"`) {
		t.Errorf("Info did not pass args through; output:\n  %s", out)
	}
}

// TestWarn_DelegatesToDefault — same contract for Warn.
func TestWarn_DelegatesToDefault(t *testing.T) {
	var buf bytes.Buffer
	lg := newJSONLoggerWithRedaction(&buf, slog.LevelWarn)
	original := Default()
	t.Cleanup(func() { SetDefault(original) })
	SetDefault(lg)

	Warn("test-warn-message", "k", "v")
	out := buf.String()

	if !strings.Contains(out, `"msg":"test-warn-message"`) {
		t.Errorf("Warn did not route to SetDefault'd logger; output:\n  %s", out)
	}
}

// TestError_DelegatesToDefault — same contract for Error.
func TestError_DelegatesToDefault(t *testing.T) {
	var buf bytes.Buffer
	lg := newJSONLoggerWithRedaction(&buf, slog.LevelError)
	original := Default()
	t.Cleanup(func() { SetDefault(original) })
	SetDefault(lg)

	Error("test-error-message", "k", "v")
	out := buf.String()

	if !strings.Contains(out, `"msg":"test-error-message"`) {
		t.Errorf("Error did not route to SetDefault'd logger; output:\n  %s", out)
	}
}

// TestHelpers_LevelFilteringRespected pins the level-filtering
// contract: package-level helpers MUST respect the level configured
// on the default logger. A regression that bypassed the level filter
// (e.g., always emitted regardless of threshold) would either spam
// logs (debug spam at info level) or swallow them (info at error
// level).
//
// Specifically: with default level = ERROR, Info/Debug/Warn calls
// MUST NOT appear in the output; only Error calls do.
func TestHelpers_LevelFilteringRespected(t *testing.T) {
	var buf bytes.Buffer
	lg := newJSONLoggerWithRedaction(&buf, slog.LevelError)
	original := Default()
	t.Cleanup(func() { SetDefault(original) })
	SetDefault(lg)

	Debug("filtered-debug")
	Info("filtered-info")
	Warn("filtered-warn")
	Error("kept-error")

	out := buf.String()
	if !strings.Contains(out, `"msg":"kept-error"`) {
		t.Errorf("Error() output missing; got:\n  %s", out)
	}
	for _, filtered := range []string{"filtered-debug", "filtered-info", "filtered-warn"} {
		if strings.Contains(out, filtered) {
			t.Errorf("level filter not respected: %q leaked into output at ERROR level:\n  %s", filtered, out)
		}
	}
}
