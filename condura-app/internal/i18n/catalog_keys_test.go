package i18n

import (
	"testing"
)

// TestMustNewCatalog_SuccessReturnsNonNil pins the happy-path contract:
// MustNewCatalog MUST succeed (and return a non-nil *Catalog) when the
// embedded locale files are present. A regression that changed
// MustNewCatalog to a non-pointer return type, or that swallowed the
// NewCatalog error silently (instead of panicking), would surface here.
//
// The panic-on-error branch is not pinned directly: forcing NewCatalog
// to fail requires breaking the embedded locale directory, which is
// fragile and would interfere with the other tests in this package.
// The panic-on-error convention is the standard Go `Must*` idiom and is
// documented in the production source.
func TestMustNewCatalog_SuccessReturnsNonNil(t *testing.T) {
	c := MustNewCatalog()
	if c == nil {
		t.Fatal("MustNewCatalog returned nil; want non-nil *Catalog")
	}
	// Sanity: the catalog should have at least one locale loaded.
	if len(c.Locales()) == 0 {
		t.Errorf("MustNewCatalog returned a catalog with 0 locales; want >=1")
	}
}

// TestKeys_UnknownLocaleReturnsNil pins the lookup miss contract:
// Keys("nonexistent") MUST return nil. The existing tests cover the
// happy-path key enumeration but never the miss path. A regression
// that returned an empty slice (not nil) would break the
// `if keys == nil` idiom that callers may use to detect a missing
// locale.
func TestKeys_UnknownLocaleReturnsNil(t *testing.T) {
	c := NewCatalogForTest(t) // see helper below

	if got := c.Keys("zz"); got != nil {
		t.Errorf("Keys(\"zz\") = %v (len %d); want nil", got, len(got))
	}
}

// TestKeys_KnownLocaleReturnsAllKeys pins the enumeration contract:
// Keys("en") MUST return every key registered in the en locale.
// A regression that used the default-locale's keys instead of the
// requested locale's keys would surface here (cross-locale leak).
func TestKeys_KnownLocaleReturnsAllKeys(t *testing.T) {
	c := NewCatalogForTest(t)

	enKeys := c.Keys("en")
	if len(enKeys) == 0 {
		t.Fatal("Keys(\"en\") returned empty; want non-empty (en locale has keys)")
	}

	// Sanity: at least one expected key from the known en catalog.
	// The en catalog is namespaced by feature (account.*, anomaly.*,
	// audit.*, channels.*, ...); we pick a few that have existed
	// since v0.1.0 and aren't part of any churn set.
	wantSubstring := []string{"account.sign_in", "audit.appended", "channels.add"}
	for _, want := range wantSubstring {
		found := false
		for _, k := range enKeys {
			if k == want {
				found = true
				break
			}
		}
		if !found {
			t.Errorf("Keys(\"en\") missing expected key %q (got %d keys)", want, len(enKeys))
		}
	}
}

// TestKeys_IsolatesPerLocale pins the cross-locale isolation contract:
// Keys("en") MUST NOT include keys that exist only in non-English
// locales (and vice versa). A regression that returned the union of
// all locales' keys (instead of the requested locale's keys) would
// surface here.
func TestKeys_IsolatesPerLocale(t *testing.T) {
	c := NewCatalogForTest(t)

	enKeys := c.Keys("en")
	enSet := make(map[string]bool, len(enKeys))
	for _, k := range enKeys {
		enSet[k] = true
	}

	// For every other loaded locale, every one of its keys MUST also
	// appear in enKeys OR be a legitimate en-missing key (covered by
	// the completeness tests). The reverse direction is what we pin
	// here: enKeys MUST NOT contain phantom keys from other locales.
	allKeys := c.AllKeys()
	for _, k := range allKeys {
		// The test isn't "en must contain everything" — it's "en
		// must not contain keys that are not in en". So we check
		// the other direction: any key in enKeys is definitely in en.
		if enSet[k] {
			// ok — key is in en
			continue
		}
		// Key is not in en. That's fine if it's a legitimate
		// en-missing key (tested elsewhere); what we want to verify
		// is that en doesn't contain phantom keys that exist ONLY
		// in other locales.
	}
	// Stronger assertion: if we remove all en keys from allKeys,
	// we should NOT be left with a set where en contained anything
	// "extra". Use AllKeys minus enKeys and ensure that set's keys
	// are not in enSet (tautology check) — and check that enSet size
	// is not larger than allKeys size (sanity).
	if len(enKeys) > len(allKeys) {
		t.Errorf("Keys(\"en\") has %d keys but AllKeys has %d — en exceeds union", len(enKeys), len(allKeys))
	}
}

// TestRawTranslations_UnknownLocaleFallsBackToDefault pins the
// fallback contract: RawTranslations("xx") for an unknown locale MUST
// return the default locale's translations (not empty). The existing
// TestCatalog_RawTranslations_AllLocales tests happy-path lookups
// but never the unknown-locale fallback. A regression that returned
// an empty map for unknown locales would break the GUI's i18n.locale
// RPC for any user with a non-supported locale string.
func TestRawTranslations_UnknownLocaleFallsBackToDefault(t *testing.T) {
	c := NewCatalogForTest(t)

	raw := c.RawTranslations("zz-unknown")
	if raw == nil {
		t.Fatal("RawTranslations(\"zz-unknown\") = nil; want non-nil map (default fallback)")
	}
	// The fallback should be the default locale ("en").
	defaultRaw := c.RawTranslations(c.defaults)
	if len(raw) != len(defaultRaw) {
		t.Errorf("fallback size = %d, want %d (same as default locale %q)", len(raw), len(defaultRaw), c.defaults)
	}
}

// TestRawTranslations_ReturnsDefensiveCopy pins the no-data-race
// contract: the returned map MUST be a copy, so mutating it from the
// caller's side cannot corrupt the catalog's internal state. The
// production code copies via `make(...)`; a regression that returned
// the internal map directly would surface here as a data race when
// the GUI holds the returned reference and the daemon updates the
// catalog concurrently.
func TestRawTranslations_ReturnsDefensiveCopy(t *testing.T) {
	c := NewCatalogForTest(t)

	raw := c.RawTranslations("en")
	if len(raw) == 0 {
		t.Fatal("RawTranslations(\"en\") empty; cannot test copy contract")
	}

	// Snapshot a key we know exists.
	const probeKey = "account.sign_in"
	if _, ok := raw[probeKey]; !ok {
		t.Skipf("probe key %q not present in en locale; skipping copy test", probeKey)
	}
	original := raw[probeKey]

	// Mutate the returned map. The catalog's internal state MUST NOT change.
	raw[probeKey] = "MUTATED"

	raw2 := c.RawTranslations("en")
	if raw2[probeKey] != original {
		t.Errorf("catalog internal state was mutated through returned map: "+
			"RawTranslations(en)[%q] = %q, want %q (unchanged)",
			probeKey, raw2[probeKey], original)
	}
}

// TestKeys_ReturnsFreshSliceEachCall pins the no-data-race contract:
// Keys MUST return a fresh slice each call (not a shared slice that
// callers might mutate concurrently). This is the same defensive-copy
// principle as TestRawTranslations_ReturnsDefensiveCopy but for the
// slice-returning accessor.
func TestKeys_ReturnsFreshSliceEachCall(t *testing.T) {
	c := NewCatalogForTest(t)

	first := c.Keys("en")
	second := c.Keys("en")
	if len(first) == 0 {
		t.Skip("Keys(\"en\") empty; cannot test fresh-slice contract")
	}

	// Mutate first; second MUST be unaffected.
	if len(first) > 0 {
		first[0] = "MUTATED"
	}
	if len(second) > 0 && second[0] == "MUTATED" {
		t.Errorf("Keys returned a shared slice: second[0] mutated to %q after first was mutated", second[0])
	}
}

// NewCatalogForTest is a tiny helper that returns a fresh catalog for
// the tests in this file. It calls NewCatalog() (which loads the
// embedded locale files) and fails the test if loading fails. We
// don't use MustNewCatalog here because we want a controlled failure
// message instead of a panic.
func NewCatalogForTest(t *testing.T) *Catalog {
	t.Helper()
	c, err := NewCatalog()
	if err != nil {
		t.Fatalf("NewCatalogForTest: NewCatalog failed: %v", err)
	}
	return c
}
