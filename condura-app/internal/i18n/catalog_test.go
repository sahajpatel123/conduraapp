package i18n

import (
	"strings"
	"testing"
)

func TestCatalog_LoadsAllLocales(t *testing.T) {
	c, err := NewCatalog()
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	locales := c.Locales()
	if len(locales) != 6 {
		t.Fatalf("expected 6 locales, got %d: %v", len(locales), locales)
	}
	for _, loc := range []string{"en", "es", "fr", "de", "ja", "zh"} {
		if !c.HasLocale(loc) {
			t.Errorf("missing locale %q", loc)
		}
	}
}

func TestCatalog_TranslationWithFallback(t *testing.T) {
	c, err := NewCatalog()
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}

	// English should work
	en := c.T("en", "daemon.ready")
	if en == "daemon.ready" {
		t.Errorf("English translation missing for daemon.ready")
	}

	// Spanish should fall back to English (since placeholders are English)
	es := c.T("es", "daemon.ready")
	if es == "daemon.ready" {
		t.Errorf("Spanish fallback missing for daemon.ready")
	}

	// Unknown locale should fall back to English
	xx := c.T("xx", "daemon.ready")
	if xx == "daemon.ready" {
		t.Errorf("Unknown locale fallback missing for daemon.ready")
	}

	// Missing key returns key itself
	missing := c.T("en", "nonexistent.key.12345")
	if missing != "nonexistent.key.12345" {
		t.Errorf("missing key should return key itself, got %q", missing)
	}
}

func TestCatalog_Formatting(t *testing.T) {
	c, err := NewCatalog()
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}

	// Test with args
	formatted := c.T("en", "llm.provider.registered", "anthropic", 3)
	expected := "Registered provider: anthropic (3 models)"
	if formatted != expected {
		t.Errorf("formatting failed: got %q, want %q", formatted, expected)
	}
}

func TestCatalog_AllKeys(t *testing.T) {
	c, err := NewCatalog()
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}

	keys := c.AllKeys()
	if len(keys) == 0 {
		t.Fatal("no keys found")
	}
	t.Logf("Total unique keys across all locales: %d", len(keys))
}

func TestCatalog_NoMissingKeys(t *testing.T) {
	c, err := NewCatalog()
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}

	for _, loc := range []string{"es", "fr", "de", "ja", "zh"} {
		missing := c.MissingKeys(loc)
		if len(missing) > 0 {
			t.Errorf("locale %q missing %d keys: %v", loc, len(missing), missing)
		}
	}
}

func TestCatalog_MustT_PanicsOnMissing(t *testing.T) {
	c, err := NewCatalog()
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}

	defer func() {
		if r := recover(); r == nil {
			t.Error("MustT should panic on missing key")
		}
	}()
	c.MustT("en", "this.key.does.not.exist.at.all")
}

func TestCatalog_PlaceholderFormat(t *testing.T) {
	c, err := NewCatalog()
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}

	// Single placeholder with string arg
	if got := c.T("en", "daemon.halted", "out of memory"); got != "Daemon halted: out of memory" {
		t.Errorf("daemon.halted with string: got %q", got)
	}
	// Single placeholder with int arg
	if got := c.T("en", "llm.registry.ready", 5); got != "LLM registry ready (5 providers)" {
		t.Errorf("llm.registry.ready with int: got %q", got)
	}
	// Two placeholders
	if got := c.T("en", "llm.provider.registered", "anthropic", 3); got != "Registered provider: anthropic (3 models)" {
		t.Errorf("llm.provider.registered: got %q", got)
	}
	// Currency-like (float) via string formatting — frontend handles
	// locale-aware currency, Go just substitutes the value.
	if got := c.T("en", "spend.limit_exceeded", 50.25); got != "Daily spend limit exceeded ($50.25)" {
		t.Errorf("spend.limit_exceeded with float: got %q", got)
	}
	// Two placeholders for spend — Go's %v uses Go's default float formatting.
	// Frontend applies locale-aware currency/decimal formatting when it
	// receives the raw translation. So on the Go side we just substitute.
	if got := c.T("en", "spend.warning", 12.5, 50.0); got != "Spend warning: $12.5 / $50 today" {
		t.Errorf("spend.warning: got %q", got)
	}
	// Raw translations should preserve {n} for the frontend
	raw := c.RawTranslations("en")["daemon.halted"]
	if raw != "Daemon halted: {0}" {
		t.Errorf("RawTranslations should preserve {0} placeholders, got %q", raw)
	}
}

func TestConvertPlaceholders(t *testing.T) {
	tests := []struct {
		in   string
		want string
	}{
		{"hello {0}", "hello %[1]v"},
		{"{0} and {1}", "%[1]v and %[2]v"},
		{"no placeholders", "no placeholders"},
		{"escaped {{literal}}", "escaped {literal}"},
		{"{0} {{ {1} }}", "%[1]v { %[2]v }"},
		{"", ""},
		// Edge: lone { at end
		{"trailing {", "trailing {"},
		// Edge: {12} - larger indices become {13} for Go's 1-indexed fmt
		{"{12} deep", "%[13]v deep"},
	}
	for _, tt := range tests {
		got := convertPlaceholders(tt.in, 0)
		if got != tt.want {
			t.Errorf("convertPlaceholders(%q) = %q, want %q", tt.in, got, tt.want)
		}
	}
}

func TestCatalog_MultipleArgs_FloatFormatting(t *testing.T) {
	c, err := NewCatalog()
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	// The `spend.warning` key has two placeholders: dollar amounts.
	// Go's %v renders 12.5 as "12.5" not "12.50". The frontend will
	// re-format with locale-aware currency; on the Go side, the
	// raw value is what we get. The point of this test is just to
	// confirm no panic and the substitution lands in the right spot.
	got := c.T("en", "spend.warning", 1.23, 4.56)
	if got == "spend.warning" {
		t.Errorf("spend.warning key not resolved")
	}
}

func TestCatalog_RawTranslations_AllLocales(t *testing.T) {
	c, err := NewCatalog()
	if err != nil {
		t.Fatalf("NewCatalog: %v", err)
	}
	// Each locale's translation should contain a {0} placeholder (when the
	// template uses one), not Go-style %s/%d. This guards against
	// regressions when translators update locale files.
	wantForms := map[string]string{
		"en": "Daemon halted: {0}",
		"es": "Daemon detenido: {0}",
		"fr": "Daemon arrêté : {0}",
		"de": "Daemon angehalten: {0}",
		"ja": "デーモンが停止しました: {0}",
		"zh": "守护进程已停止: {0}",
	}
	for _, loc := range c.Locales() {
		raw := c.RawTranslations(loc)
		v, ok := raw["daemon.halted"]
		if !ok {
			t.Errorf("locale %q missing daemon.halted", loc)
			continue
		}
		if v != wantForms[loc] {
			t.Errorf("locale %q daemon.halted = %q, want %q", loc, v, wantForms[loc])
		}
		// Make sure no Go-style %s/%d leaked through
		if strings.Contains(v, "%s") || strings.Contains(v, "%d") {
			t.Errorf("locale %q has Go-style placeholder: %q", loc, v)
		}
	}
}
