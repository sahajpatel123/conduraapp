// Package logger provides structured logging for Condura.
//
// It wraps log/slog (Go 1.21+ standard library) and adds:
//
//   - Sensitive data redaction for keys/values matching known patterns
//     (api_key, token, password, secret, authorization, cookie, etc.).
//   - Standard attribute names for request_id, session_id, user_id, run_id.
//   - Two output formats: JSON (production) and Text (development).
//   - Two sinks: stderr (always) and optional size-rotated file.
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

// Log level values.
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

// Log format values.
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

// Default rotation limits when a file sink is configured.
// Sized so a chatty debug day cannot fill a disk while still
// retaining enough history for post-incident forensics.
const (
	DefaultMaxSizeMB  = 50
	DefaultMaxBackups = 5
	DefaultMaxAgeDays = 30
)

// Config configures the logger.
type Config struct {
	// Level is the minimum log level. Default: "info".
	Level Level
	// Format is the output format. Default: "text".
	Format Format
	// AddSource adds the source file:line to each log entry. Default: false (in prod).
	AddSource bool
	// File is an optional file path to write logs to in addition to stderr.
	// Empty means stderr only. When set, size-based rotation is applied.
	File string
	// MaxSizeMB is the maximum size in megabytes of a single log file
	// before rotation. 0 uses DefaultMaxSizeMB. Only applied when File is set.
	MaxSizeMB int
	// MaxBackups is the maximum number of rotated files to retain
	// (name.log.1 … name.log.N). 0 uses DefaultMaxBackups. Negative disables pruning.
	MaxBackups int
	// MaxAgeDays is the maximum age in days of a rotated file before
	// it is deleted. 0 uses DefaultMaxAgeDays. Negative disables age pruning.
	MaxAgeDays int
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
	var fileCloser io.Closer
	if cfg.File != "" {
		writer, fileCloser = openFileOrStderr(cfg, writer)
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

	lg := slog.New(base)
	if fileCloser != nil {
		// Stash on a package map keyed by logger pointer so tests
		// (and optional daemon shutdown) can release the Windows file lock.
		registerFileCloser(lg, fileCloser)
	}
	return lg
}

// CloseFileSink closes any on-disk rotating file associated with lg.
// Safe if lg has no file sink. Required on Windows before removing
// the temp log directory in tests.
func CloseFileSink(lg *slog.Logger) error {
	return closeFileCloser(lg)
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

// Debug logs at debug level on the default logger.
func Debug(msg string, args ...any) { Default().Debug(msg, args...) }

// Info logs at info level on the default logger.
func Info(msg string, args ...any) { Default().Info(msg, args...) }

// Warn logs at warn level on the default logger.
func Warn(msg string, args ...any) { Default().Warn(msg, args...) }

// Error logs at error level on the default logger.
func Error(msg string, args ...any) { Default().Error(msg, args...) }

// DebugContext logs at debug level on the default logger, attaching ctx.
func DebugContext(ctx context.Context, msg string, args ...any) {
	Default().DebugContext(ctx, msg, args...)
}

// InfoContext logs at info level on the default logger, attaching ctx.
func InfoContext(ctx context.Context, msg string, args ...any) {
	Default().InfoContext(ctx, msg, args...)
}

// WarnContext logs at warn level on the default logger, attaching ctx.
func WarnContext(ctx context.Context, msg string, args ...any) {
	Default().WarnContext(ctx, msg, args...)
}

// ErrorContext logs at error level on the default logger, attaching ctx.
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

func openFileOrStderr(cfg Config, fallback io.Writer) (io.Writer, io.Closer) {
	dir := filepath.Dir(cfg.File)
	if err := os.MkdirAll(dir, logDirPerm); err != nil {
		return fallback, nil
	}

	maxSize := cfg.MaxSizeMB
	if maxSize <= 0 {
		maxSize = DefaultMaxSizeMB
	}
	maxBackups := cfg.MaxBackups
	if maxBackups == 0 {
		maxBackups = DefaultMaxBackups
	}
	maxAge := cfg.MaxAgeDays
	if maxAge == 0 {
		maxAge = DefaultMaxAgeDays
	}

	w, err := newRotatingWriter(rotatingConfig{
		Filename:   cfg.File,
		MaxSize:    int64(maxSize) * 1024 * 1024,
		MaxBackups: maxBackups,
		MaxAgeDays: maxAge,
	})
	if err != nil {
		return fallback, nil
	}
	// Tee to stderr so operators still see live output when a file sink
	// is configured (daemon journal + on-disk forensics).
	return io.MultiWriter(os.Stderr, w), w
}

// fileClosers tracks open rotating file sinks so CloseFileSink can release
// OS locks (required on Windows before TempDir cleanup).
var (
	fileClosersMu sync.Mutex
	fileClosers   = map[*slog.Logger]io.Closer{}
)

func registerFileCloser(lg *slog.Logger, c io.Closer) {
	fileClosersMu.Lock()
	fileClosers[lg] = c
	fileClosersMu.Unlock()
}

func closeFileCloser(lg *slog.Logger) error {
	fileClosersMu.Lock()
	c, ok := fileClosers[lg]
	if ok {
		delete(fileClosers, lg)
	}
	fileClosersMu.Unlock()
	if !ok || c == nil {
		return nil
	}
	return c.Close()
}

// -----------------------------------------------------------------------------
// Standard attribute keys
// -----------------------------------------------------------------------------

// Standard attribute keys used across Condura. Using constants avoids typos
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
