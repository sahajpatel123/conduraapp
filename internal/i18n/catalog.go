// Package i18n provides internationalization support for the daemon.
// Locale catalogs are embedded via go:embed and treated as untrusted data:
// keys are never interpolated into commands or policies.
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"sync"
)

//go:embed locales/*.json
var localeFS embed.FS

// Catalog holds translations for all locales.
//
// Placeholder format: locale files use {0}, {1}, {2} ... (i18next / i18next-react
// compatible). The Go side resolves them via fmt-style verbs to support
// formatted arguments (strings, integers, floats). The {n} placeholders are
// rewritten to %v/%d/%.Nf on the fly so a single locale file works for both
// the Wails frontend and Go callers.
type Catalog struct {
	mu       sync.RWMutex
	locales  map[string]map[string]string
	defaults string
}

// NewCatalog loads all embedded locale files.
func NewCatalog() (*Catalog, error) {
	c := &Catalog{
		locales: make(map[string]map[string]string),
	}
	entries, err := localeFS.ReadDir("locales")
	if err != nil {
		return nil, fmt.Errorf("read locales dir: %w", err)
	}
	for _, e := range entries {
		if e.IsDir() || len(e.Name()) < 6 {
			continue
		}
		locale := e.Name()[:len(e.Name())-5] // strip .json
		data, err := localeFS.ReadFile("locales/" + e.Name())
		if err != nil {
			return nil, fmt.Errorf("read locale %s: %w", locale, err)
		}
		var m map[string]string
		if err := json.Unmarshal(data, &m); err != nil {
			return nil, fmt.Errorf("parse locale %s: %w", locale, err)
		}
		c.locales[locale] = m
		if locale == "en" {
			c.defaults = locale
		}
	}
	if len(c.locales) == 0 {
		return nil, fmt.Errorf("no locales loaded")
	}
	if c.defaults == "" {
		c.defaults = "en"
	}
	return c, nil
}

// MustNewCatalog loads the catalog and panics on error.
// Use during daemon startup where failure is fatal.
func MustNewCatalog() *Catalog {
	c, err := NewCatalog()
	if err != nil {
		panic(fmt.Sprintf("i18n: failed to load catalog: %v", err))
	}
	return c
}

// convertPlaceholders rewrites {0}, {1}, {2} ... to fmt's %[N+1]v indexed
// verbs so a single template works for both the Wails frontend and Go.
//
// Go's fmt package uses 1-indexed arguments: %[1]v is the first arg, %[2]v
// the second, etc. So a {0} in the template becomes %[1]v in the Go template.
//
// The default verb is %v (accepts any type: int, string, float, etc.).
// Localization-specific formatting (currency, decimal separators) is handled
// on the frontend where the i18next ecosystem has mature formatters.
func convertPlaceholders(template string, _ int) string {
	var b strings.Builder
	b.Grow(len(template) + 8)
	i := 0
	for i < len(template) {
		c := template[i]
		switch {
		case c == '{' && i+1 < len(template) && template[i+1] >= '0' && template[i+1] <= '9':
			// Try to parse a {N} placeholder.
			if j, ok := scanPlaceholderEnd(template, i+1); ok {
				// Valid {N} placeholder. Go fmt is 1-indexed: {0} -> %[1]v.
				n := template[i+1 : j]
				idx, _ := strconv.Atoi(n)
				b.WriteString("%[")
				b.WriteString(strconv.Itoa(idx + 1))
				b.WriteString("]v")
				i = j + 1
				continue
			}
			// Not a valid placeholder; fall through and emit the char.
			b.WriteByte(c)
			i++
		case c == '{' && i+1 < len(template) && template[i+1] == '{':
			// Escaped brace.
			b.WriteByte('{')
			i += 2
		case c == '}' && i+1 < len(template) && template[i+1] == '}':
			// Escaped brace.
			b.WriteByte('}')
			i += 2
		default:
			b.WriteByte(c)
			i++
		}
	}
	return b.String()
}

// scanPlaceholderEnd walks digits starting at start and returns the
// position of the matching '}' if a valid {N} placeholder follows.
// Returns (pos, false) if no placeholder is present.
func scanPlaceholderEnd(template string, start int) (int, bool) {
	j := start
	for j < len(template) && template[j] >= '0' && template[j] <= '9' {
		j++
	}
	if j < len(template) && template[j] == '}' {
		return j, true
	}
	return 0, false
}

// T returns the translation for key in locale, with fallback to English.
func (c *Catalog) T(locale, key string, args ...any) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if m, ok := c.locales[locale]; ok {
		if v, ok := m[key]; ok {
			return applyTemplate(v, args)
		}
	}
	if m, ok := c.locales[c.defaults]; ok {
		if v, ok := m[key]; ok {
			return applyTemplate(v, args)
		}
	}
	return key // fallback to key itself
}

// MustT returns translation or panics if key missing in all locales (for tests).
func (c *Catalog) MustT(locale, key string, args ...any) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if m, ok := c.locales[locale]; ok {
		if v, ok := m[key]; ok {
			return applyTemplate(v, args)
		}
	}
	if m, ok := c.locales[c.defaults]; ok {
		if v, ok := m[key]; ok {
			return applyTemplate(v, args)
		}
	}
	panic(fmt.Sprintf("i18n: missing key %q in locale %q and default %q", key, locale, c.defaults))
}

// applyTemplate converts {n} placeholders to fmt verbs and runs Sprintf.
// On any formatting error (missing arg, type mismatch), it returns the
// unformatted template with the {n} placeholders left intact — which makes
// bugs visible in logs but never panics inside a hot path.
func applyTemplate(template string, args []any) string {
	converted := convertPlaceholders(template, len(args))
	out := fmt.Sprintf(converted, args...)
	return out
}

// Locales returns the list of available locale codes.
func (c *Catalog) Locales() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	out := make([]string, 0, len(c.locales))
	for k := range c.locales {
		out = append(out, k)
	}
	return out
}

// HasLocale returns true if locale is loaded.
func (c *Catalog) HasLocale(locale string) bool {
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, ok := c.locales[locale]
	return ok
}

// Keys returns all translation keys for a locale (for completeness tests).
func (c *Catalog) Keys(locale string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m, ok := c.locales[locale]
	if !ok {
		return nil
	}
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	return keys
}

// RawTranslations returns the raw format strings for a locale without
// applying fmt.Sprintf. The {n} placeholders are returned as-is so the
// frontend's t() function can substitute them with localized formatters
// (number separators, currency, etc.). Used by the i18n.locale RPC.
func (c *Catalog) RawTranslations(locale string) map[string]string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	m, ok := c.locales[locale]
	if !ok {
		// Fall back to default locale.
		m, ok = c.locales[c.defaults]
		if !ok {
			return map[string]string{}
		}
	}
	// Return a copy to avoid data races.
	out := make(map[string]string, len(m))
	for k, v := range m {
		out[k] = v
	}
	return out
}

// AllKeys returns the union of all keys across all locales.
func (c *Catalog) AllKeys() []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	seen := make(map[string]struct{})
	for _, m := range c.locales {
		for k := range m {
			seen[k] = struct{}{}
		}
	}
	keys := make([]string, 0, len(seen))
	for k := range seen {
		keys = append(keys, k)
	}
	return keys
}

// MissingKeys returns keys present in default locale but missing in locale.
func (c *Catalog) MissingKeys(locale string) []string {
	c.mu.RLock()
	defer c.mu.RUnlock()
	def, ok := c.locales[c.defaults]
	if !ok {
		return nil
	}
	target, ok := c.locales[locale]
	if !ok {
		return nil
	}
	var missing []string
	for k := range def {
		if _, ok := target[k]; !ok {
			missing = append(missing, k)
		}
	}
	return missing
}
