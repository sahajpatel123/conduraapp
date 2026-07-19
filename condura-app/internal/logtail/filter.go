package logtail

import (
	"regexp"
	"strings"
)

// Level is the slog log level (we mirror slog.Level here to
// avoid forcing cmd/condura to import log/slog just for one
// constant). The order is the same as slog: Debug < Info <
// Warn < Error.
type Level int

// Log levels. The zero value (LevelDebug) means "all levels
// pass" (no level filter). The order matches slog: Debug
// < Info < Warn < Error.
const (
	LevelDebug Level = iota
	LevelInfo
	LevelWarn
	LevelError
)

// levelNames maps the canonical name to the Level. Used for
// parsing the --level flag value.
var levelNames = map[string]Level{
	"debug": LevelDebug,
	"info":  LevelInfo,
	"warn":  LevelWarn,
	"error": LevelError,
}

// levelByName is the inverse of levelNames; built once at
// init for fast lookup in ParseLevel. We pre-build the
// inverse map so ParseLevel is O(1) per call.
var levelByName = func() map[Level]string {
	m := make(map[Level]string, len(levelNames))
	for k, v := range levelNames {
		m[v] = k
	}
	return m
}()

// ParseLevel parses a slog-style level name (case-insensitive)
// and returns the corresponding Level. Returns (level, false)
// if the name is not recognized.
func ParseLevel(name string) (Level, bool) {
	l, ok := levelNames[strings.ToLower(name)]
	return l, ok
}

// String returns the canonical lowercase name for a Level.
func (l Level) String() string {
	return levelByName[l]
}

// Filter decides whether a single log line should be included
// in the output. Created via NewFilter; safe for concurrent
// use after construction (the regex is compiled once).
//
// Two filters compose:
//   - level: minimum level to show. A line with level < minLevel
//     is dropped. If the line can't be parsed as a structured
//     log (e.g. it's plain text), the level filter passes it
//     through (we err on the side of "show the line").
//   - grep: regex that the line must match. If the regex is
//     empty, the grep filter is a no-op.
//
// A nil filter (or any field-zero filter) is a no-op pass-through.
type Filter struct {
	minLevel Level
	regex    *regexp.Regexp
}

// NewFilter returns a Filter that requires level >= minLevel
// AND the line to match regex. Either argument can be the
// zero value for "no filter":
//   - minLevel = LevelDebug (0) → all levels pass
//   - regex = "" → grep filter is skipped
func NewFilter(minLevel Level, regex string) *Filter {
	f := &Filter{minLevel: minLevel}
	if regex != "" {
		// Compile with the case-insensitive flag: the operator
		// usually types "ERROR" or "error" and expects either
		// to match. (Default regex is case-sensitive, which
		// surprises users who don't read Go's regexp docs.)
		f.regex = regexp.MustCompile(`(?i)` + regex)
	}
	return f
}

// Matches returns true if line should be included in the
// output, false if it should be dropped.
//
// The check order is: level first (cheap), then regex
// (more expensive, only on lines that already passed the
// level filter). This minimizes the per-line cost when the
// level filter drops most lines.
func (f *Filter) Matches(line string) bool {
	if f == nil {
		return true
	}
	// Level filter: try to parse the line as a structured
	// slog log. The format is:
	//   {"time":"...","level":"INFO","msg":"..."}
	// We only check the level field; the rest is opaque.
	if !matchesLevel(line, f.minLevel) {
		return false
	}
	// Regex filter: if set, the line must match.
	if f.regex != nil && !f.regex.MatchString(line) {
		return false
	}
	return true
}

// matchesLevel parses a structured log line and returns true
// if the line's level >= minLevel. Returns true (pass-through)
// for lines that don't look like structured slog output —
// we err on the side of "show the line" rather than "drop it".
//
// The implementation is intentionally cheap: it scans the
// line for `"level":"..."` (the slog JSON field) and
// compares the value. It does NOT parse the whole JSON — a
// non-JSON line is just returned as pass-through.
func matchesLevel(line string, minLevel Level) bool {
	const key = `"level":`
	idx := strings.Index(line, key)
	if idx < 0 {
		// Not a structured log line (e.g. plain text written
		// by fmt.Println). Pass through — the level filter
		// shouldn't hide non-JSON lines.
		return true
	}
	// Skip past the key, find the opening quote of the
	// value, then the value text, then the closing quote.
	rest := line[idx+len(key):]
	// Skip whitespace (Go's json encoder doesn't add any
	// after the colon, but other encoders might).
	rest = strings.TrimLeft(rest, " ")
	if len(rest) == 0 || rest[0] != '"' {
		return true
	}
	rest = rest[1:]
	// Find the closing quote. slog values don't contain
	// quotes (the level is "INFO", "WARN", etc., no escapes).
	end := strings.IndexByte(rest, '"')
	if end < 0 {
		return true
	}
	name := rest[:end]
	// Normalize: case-insensitive comparison. "WARN" ==
	// "warn" == "Warn".
	lvl, ok := levelNames[strings.ToLower(name)]
	if !ok {
		// Unknown level name in the log line. Pass through —
		// the operator probably has a custom level; not our
		// problem to filter.
		return true
	}
	return lvl >= minLevel
}
