// Package logger provides the structured logging for Synaptic.
//
// It wraps log/slog (Go 1.21+ standard library) and adds:
//
//   - Sensitive data redaction for keys/values matching known patterns
//     (api_key, token, password, secret, authorization, cookie, etc.).
//   - Standard attribute names for request_id, session_id, user_id, run_id.
//   - Two output formats: JSON (production) and Text (development).
//   - Two sinks: stderr (always) and optional file (with rotation, future).
//
// The redaction is applied via a wrapping slog.Handler, so any handler
// (slog.NewJSONHandler, slog.NewTextHandler, custom) is supported.
//
// Usage:
//
//	logger.SetDefault(logger.New(logger.Config{Level: "info", Format: "json"}))
//	logger.Info("daemon started", "version", version.Get().Version)
//	logger.With("request_id", id).Info("processing request")
package logger

import (
	"context"
	"io"
	"log/slog"
	"os"
	"path/filepath"
	"strings"
	"sync"
)

// -----------------------------------------------------------------------------
// Public types
// -----------------------------------------------------------------------------

// Level represents a log level. Valid string values: "debug", "info", "warn", "error".
type Level string

const (
	LevelDebug Level = "debug"
	LevelInfo  Level = "info"
	LevelWarn  Level = "warn"
	LevelError Level = "error"
)

// ParseLevel parses a string into a Level. Defaults to Info on unknown.
func ParseLevel(s string) Level {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "debug":
		return LevelDebug
	case "info", "":
		return LevelInfo
	case "warn", "warning":
		return LevelWarn
	case "error", "err":
		return LevelError
	default:
		return LevelInfo
	}
}

// Format represents the log output format.
type Format string

const (
	FormatJSON Format = "json"
	FormatText Format = "text"
)

// ParseFormat parses a string into a Format. Defaults to Text on unknown.
func ParseFormat(s string) Format {
	switch strings.ToLower(strings.TrimSpace(s)) {
	case "json":
		return FormatJSON
	case "text", "":
		return FormatText
	default:
		return FormatText
	}
}

// Config configures the logger.
type Config struct {
	// Level is the minimum log level. Default: "info".
	Level Level
	// Format is the output format. Default: "text".
	Format Format
	// AddSource adds the source file:line to each log entry. Default: false (in prod).
	AddSource bool
	// File is an optional file path to write logs to in addition to stderr.
	// Empty means stderr only.
	File string
	// Redact enables sensitive data redaction. Default: true.
	Redact *bool
}

// -----------------------------------------------------------------------------
// Defaults and globals
// -----------------------------------------------------------------------------

var (
	defaultMu sync.RWMutex
	defaultLg *slog.Logger
)

func init() {
	defaultLg = New(Config{Level: LevelInfo, Format: FormatText, Redact: boolPtr(true)})
	slog.SetDefault(defaultLg)
}

// File mode for an optional log file. We pick owner-only because the
// log may contain redacted-but-still-sensitive paths and IDs.
const (
	logDirPerm  = 0o700
	logFilePerm = 0o600
)

// New returns a new logger with the given config.
func New(cfg Config) *slog.Logger {
	if cfg.Level == "" {
		cfg.Level = LevelInfo
	}
	if cfg.Format == "" {
		cfg.Format = FormatText
	}
	redact := true
	if cfg.Redact != nil {
		redact = *cfg.Redact
	}

	opts := &slog.HandlerOptions{
		Level:     toSlogLevel(cfg.Level),
		AddSource: cfg.AddSource,
	}

	var writer io.Writer = os.Stderr
	if cfg.File != "" {
		writer = openFileOrStderr(cfg.File, writer)
	}

	var base slog.Handler
	switch cfg.Format {
	case FormatJSON:
		base = slog.NewJSONHandler(writer, opts)
	default:
		base = slog.NewTextHandler(writer, opts)
	}

	if redact {
		base = newRedactingHandler(base)
	}

	return slog.New(base)
}

// SetDefault sets the default logger returned by package-level helpers and by slog.
func SetDefault(lg *slog.Logger) {
	defaultMu.Lock()
	defer defaultMu.Unlock()
	defaultLg = lg
	slog.SetDefault(lg)
}

// Default returns the default logger.
func Default() *slog.Logger {
	defaultMu.RLock()
	defer defaultMu.RUnlock()
	return defaultLg
}

// -----------------------------------------------------------------------------
// Package-level convenience helpers (delegate to default logger).
// -----------------------------------------------------------------------------

func Debug(msg string, args ...any) { Default().Debug(msg, args...) }
func Info(msg string, args ...any)  { Default().Info(msg, args...) }
func Warn(msg string, args ...any)  { Default().Warn(msg, args...) }
func Error(msg string, args ...any) { Default().Error(msg, args...) }

func DebugContext(ctx context.Context, msg string, args ...any) {
	Default().DebugContext(ctx, msg, args...)
}
func InfoContext(ctx context.Context, msg string, args ...any) {
	Default().InfoContext(ctx, msg, args...)
}
func WarnContext(ctx context.Context, msg string, args ...any) {
	Default().WarnContext(ctx, msg, args...)
}
func ErrorContext(ctx context.Context, msg string, args ...any) {
	Default().ErrorContext(ctx, msg, args...)
}

// With returns a new logger with the given attributes attached.
func With(args ...any) *slog.Logger {
	return Default().With(args...)
}

// WithGroup returns a new logger with the given group name attached.
func WithGroup(name string) *slog.Logger {
	return Default().WithGroup(name)
}

// -----------------------------------------------------------------------------
// Internal helpers
// -----------------------------------------------------------------------------

func toSlogLevel(lvl Level) slog.Level {
	switch lvl {
	case LevelDebug:
		return slog.LevelDebug
	case LevelWarn:
		return slog.LevelWarn
	case LevelError:
		return slog.LevelError
	default:
		return slog.LevelInfo
	}
}

func boolPtr(b bool) *bool { return &b }

func openFileOrStderr(path string, fallback io.Writer) io.Writer {
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, logDirPerm); err != nil {
		return fallback
	}
	f, err := os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_APPEND, logFilePerm)
	if err != nil {
		return fallback
	}
	// We intentionally do not close f — the OS will clean up on process exit.
	// For long-running daemons, a future log rotation module can take over.
	return f
}

// -----------------------------------------------------------------------------
// Standard attribute keys
// -----------------------------------------------------------------------------

// Standard attribute keys used across Synaptic. Using constants avoids typos
// and makes refactors safe.
const (
	KeyRequestID = "request_id"
	KeySessionID = "session_id"
	KeyUserID    = "user_id"
	KeyRunID     = "run_id"
	KeyComponent = "component"
	KeyProvider  = "provider"
	KeyModel     = "model"
)
