package i18n

import (
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
