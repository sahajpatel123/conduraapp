// Package i18n provides internationalization support for the daemon.
// Locale catalogs are embedded via go:embed and treated as untrusted data:
// keys are never interpolated into commands or policies.
package i18n

import (
	"embed"
	"encoding/json"
	"fmt"
	"sync"
)

//go:embed locales/*.json
var localeFS embed.FS

// Catalog holds translations for all locales.
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

// T returns the translation for key in locale, with fallback to English.
func (c *Catalog) T(locale, key string, args ...any) string {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if m, ok := c.locales[locale]; ok {
		if v, ok := m[key]; ok {
			return fmt.Sprintf(v, args...)
		}
	}
	if m, ok := c.locales[c.defaults]; ok {
		if v, ok := m[key]; ok {
			return fmt.Sprintf(v, args...)
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
			return fmt.Sprintf(v, args...)
		}
	}
	if m, ok := c.locales[c.defaults]; ok {
		if v, ok := m[key]; ok {
			return fmt.Sprintf(v, args...)
		}
	}
	panic(fmt.Sprintf("i18n: missing key %q in locale %q and default %q", key, locale, c.defaults))
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
