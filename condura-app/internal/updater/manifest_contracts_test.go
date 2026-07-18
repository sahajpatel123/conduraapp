package updater

import (
	"crypto/ed25519"
	"strings"
	"testing"
)

// TestPlatformFromArchiveName_StandardFormats pins the parsing
// contract: archive names of the form `<project>-<version>-<goos>-<goarch>.tar.gz`
// MUST be parsed to the `<goos>/<goarch>` platform key. A regression
// in the parser (wrong separator, wrong field selection) would
// silently miscategorize every release artifact.
func TestPlatformFromArchiveName_StandardFormats(t *testing.T) {
	cases := []struct {
		name     string
		archive  string
		project  string
		wantPlat string
		wantOK   bool
	}{
		{"linux-amd64-targz", "condurad-1.0.0-linux-amd64.tar.gz", "condurad", "linux/amd64", true},
		{"darwin-arm64-targz", "condurad-1.0.0-darwin-arm64.tar.gz", "condurad", "darwin/arm64", true},
		{"windows-amd64-zip", "condurad-1.0.0-windows-amd64.zip", "condurad", "windows/amd64", true},
		{"darwin-arm64-zip", "condurad-1.0.0-darwin-arm64.zip", "condurad", "darwin/arm64", true},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, ok := PlatformFromArchiveName(tc.archive, tc.project)
			if ok != tc.wantOK {
				t.Fatalf("ok = %v, want %v", ok, tc.wantOK)
			}
			if got != tc.wantPlat {
				t.Errorf("platform = %q, want %q", got, tc.wantPlat)
			}
		})
	}
}

// TestPlatformFromArchiveName_WrongPrefixRejected pins the
// prefix-mismatch contract: an archive whose name doesn't start with
// `<project>-` MUST return ("", false). This is the safety net that
// prevents BuildManifestFromChecksums from picking up
// `condura-cli-*` archives as if they were `condurad-*` daemons.
func TestPlatformFromArchiveName_WrongPrefixRejected(t *testing.T) {
	cases := []string{
		"condura-cli-1.0.0-linux-amd64.tar.gz", // different project
		"someother-1.0.0-linux-amd64.tar.gz",   // totally different
		"linux-amd64.tar.gz",                   // no project prefix at all
		"",                                     // empty
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			got, ok := PlatformFromArchiveName(name, "condurad")
			if ok {
				t.Errorf("PlatformFromArchiveName(%q, \"condurad\") = (%q, true); want (\"\", false)", name, got)
			}
			if got != "" {
				t.Errorf("PlatformFromArchiveName(%q, \"condurad\") = (%q, false); want (\"\", false)", name, got)
			}
		})
	}
}

// TestPlatformFromArchiveName_MalformedRejected pins the structural
// guards: an archive name without a version separator (hyphen)
// MUST return ("", false). Catches inputs like "condurad-linux-amd64"
// (no version between project and platform).
func TestPlatformFromArchiveName_MalformedRejected(t *testing.T) {
	cases := []string{
		"condurad-linux-amd64.tar.gz",  // no version separator
		"condurad-1.0.0.tar.gz",        // no platform separator
		"condurad-1.0.0-linux.tar.gz",  // empty arch
		"condurad-1.0.0--amd64.tar.gz", // empty goos (leading hyphen)
	}
	for _, name := range cases {
		t.Run(name, func(t *testing.T) {
			got, ok := PlatformFromArchiveName(name, "condurad")
			if ok {
				t.Errorf("malformed name %q parsed to (%q, true); want (\"\", false)", name, got)
			}
		})
	}
}

// TestVerifyPayload_TamperedPayloadRejected pins the integrity
// contract: VerifyPayload MUST reject a payload that was modified
// after signing. This is the core defense against MITM tampering —
// without it, an attacker could substitute a different download URL
// or version while keeping the signature valid.
func TestVerifyPayload_TamperedPayloadRejected(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	original := ManifestPayload{
		Version:     "1.0.0",
		Channel:     "stable",
		DownloadURL: "https://updates.example.com/condurad-1.0.0.tar.gz",
		SHA256:      "abc123",
	}
	sig, err := SignPayload(original, priv)
	if err != nil {
		t.Fatal(err)
	}

	// Tamper with the URL after signing.
	tampered := original
	tampered.DownloadURL = "https://attacker.example.com/malware.tar.gz"

	if err := VerifyPayload(tampered, pub, sig); err == nil {
		t.Error("VerifyPayload accepted tampered payload; want error")
	}
}

// TestVerifyPayload_WrongKeyRejected pins the cross-key defense:
// signing with priv1 and verifying with pub2 MUST fail. Without this
// pin, an attacker who intercepts a manifest signed for the official
// key could pass it off as signed for their own (rogue) key.
func TestVerifyPayload_WrongKeyRejected(t *testing.T) {
	pub1, priv1, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	pub2, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	p := ManifestPayload{Version: "1.0.0", Channel: "stable"}
	sig, err := SignPayload(p, priv1)
	if err != nil {
		t.Fatal(err)
	}

	if err := VerifyPayload(p, pub2, sig); err == nil {
		t.Error("VerifyPayload accepted sig from a different key; want error")
	}
	// Sanity: the right key still works.
	if err := VerifyPayload(p, pub1, sig); err != nil {
		t.Errorf("VerifyPayload rejected sig with the right key: %v", err)
	}
}

// TestVerifyPayload_InvalidHexRejected pins the input-validation
// contract: VerifyPayload MUST return a clear error when the
// signature is not valid hex (not 64 bytes of hex chars). Without
// this pin, a malformed sig would crash inside ed25519.Verify with a
// confusing error.
func TestVerifyPayload_InvalidHexRejected(t *testing.T) {
	pub, _, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	p := ManifestPayload{Version: "1.0.0"}

	cases := []string{
		"",                       // empty
		"not-hex-at-all",         // ASCII letters
		"abcd",                   // too short
		"deadbeefdeadbeef",       // 16 bytes, not 64
		strings.Repeat("z", 128), // wrong chars
	}
	for _, sigHex := range cases {
		t.Run(sigHex, func(t *testing.T) {
			err := VerifyPayload(p, pub, sigHex)
			if err == nil {
				t.Errorf("VerifyPayload with invalid hex %q returned nil; want error", sigHex)
			}
			// Sanity: error message should mention "signature" so
			// log readers can diagnose.
			if !strings.Contains(err.Error(), "signature") {
				t.Errorf("error %q should mention 'signature' for diagnostic clarity", err.Error())
			}
		})
	}
}

// TestResolveArtifact_MissingPlatformErrors pins the platform-key
// contract: when Platforms is set but has no entry for the current
// GOOS/GOARCH, ResolveArtifact MUST return an error mentioning the
// missing platform. Without this pin, a release that forgot to
// include the user's platform would silently return an empty URL,
// leading to a 404 download that fails opaquely downstream.
func TestResolveArtifact_MissingPlatformErrors(t *testing.T) {
	sm := SignedManifest{
		Platforms: map[string]PlatformArtifact{
			"plan9/amd64": {DownloadURL: "https://example.com/p9", SHA256: "deadbeef"},
		},
	}
	_, _, err := sm.ResolveArtifact()
	if err == nil {
		t.Fatal("ResolveArtifact with no entry for current platform returned nil; want error")
	}
	if !strings.Contains(err.Error(), "no artifact") {
		t.Errorf("error %q should mention 'no artifact' for diagnostic clarity", err.Error())
	}
}

// TestResolveArtifact_IncompleteArtifactErrors pins the
// per-platform integrity guard: an entry with either empty
// DownloadURL or empty SHA256 MUST be rejected (defense against a
// corrupt checksums.txt that produces a half-populated entry).
func TestResolveArtifact_IncompleteArtifactErrors(t *testing.T) {
	cases := []struct {
		name string
		art  PlatformArtifact
		want string
	}{
		{"empty-url", PlatformArtifact{DownloadURL: "", SHA256: "deadbeef"}, "incomplete"},
		{"empty-sha", PlatformArtifact{DownloadURL: "https://x/y", SHA256: ""}, "incomplete"},
		{"both-empty", PlatformArtifact{DownloadURL: "", SHA256: ""}, "incomplete"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			sm := SignedManifest{
				Platforms: map[string]PlatformArtifact{
					PlatformKey(): tc.art,
				},
			}
			_, _, err := sm.ResolveArtifact()
			if err == nil {
				t.Fatalf("ResolveArtifact with %v returned nil; want error", tc.art)
			}
			if !strings.Contains(err.Error(), tc.want) {
				t.Errorf("error %q should contain %q", err.Error(), tc.want)
			}
		})
	}
}

// TestResolveArtifact_LegacyFallback pins the legacy single-target
// contract: when Platforms is empty, ResolveArtifact MUST fall back
// to the top-level DownloadURL/SHA256 fields. A regression that
// required Platforms (ignoring legacy manifests) would break old
// releases.
func TestResolveArtifact_LegacyFallback(t *testing.T) {
	// Empty Platforms + valid DownloadURL/SHA: success.
	sm := SignedManifest{
		DownloadURL: "https://example.com/legacy.tar.gz",
		SHA256:      "cafebabe",
	}
	url, sha, err := sm.ResolveArtifact()
	if err != nil {
		t.Fatalf("legacy fallback: %v", err)
	}
	if url != "https://example.com/legacy.tar.gz" || sha != "cafebabe" {
		t.Errorf("legacy fallback returned url=%q sha=%q; want the top-level fields", url, sha)
	}
}

// TestResolveArtifact_LegacyMissingURLErrors pins the legacy path's
// own guard: if Platforms is empty AND DownloadURL is empty, MUST
// error (not return empty URL silently).
func TestResolveArtifact_LegacyMissingURLErrors(t *testing.T) {
	sm := SignedManifest{
		SHA256: "abc",
	}
	_, _, err := sm.ResolveArtifact()
	if err == nil {
		t.Error("ResolveArtifact with empty Platforms + empty DownloadURL returned nil; want error")
	}
}

// TestBuildManifestFromChecksums_EmptyVersionErrors pins the
// required-version guard: version MUST be non-empty (the manifest
// payload would otherwise be unusable).
func TestBuildManifestFromChecksums_EmptyVersionErrors(t *testing.T) {
	_, err := BuildManifestFromChecksums("", "stable", "https://x", "", []ChecksumEntry{
		{Hash: "aa", Name: "condurad-1.0.0-linux-amd64.tar.gz"},
	})
	if err == nil {
		t.Error("BuildManifestFromChecksums(\"\", ...) returned nil; want error")
	}
}

// TestBuildManifestFromChecksums_DefaultsChannelToStable pins the
// default-channel contract: when channel is empty, the manifest MUST
// default to "stable" (the only safe default for an unsigned-on-stable
// fallback). Without this pin, a caller who forgot to set channel
// would produce a manifest with empty Channel — which the verifier
// would reject as unknown.
func TestBuildManifestFromChecksums_DefaultsChannelToStable(t *testing.T) {
	p, err := BuildManifestFromChecksums("v1.0.0", "", "https://x", "", []ChecksumEntry{
		{Hash: "aa", Name: "condurad-1.0.0-linux-amd64.tar.gz"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Channel != "stable" {
		t.Errorf("Channel = %q, want \"stable\"", p.Channel)
	}
}

// TestBuildManifestFromChecksums_NoConduradArchivesErrors pins the
// filtering contract: if no entry matches `condurad-*`, the manifest
// MUST error (not silently produce an empty Platforms map).
func TestBuildManifestFromChecksums_NoConduradArchivesErrors(t *testing.T) {
	entries := []ChecksumEntry{
		{Hash: "aa", Name: "condura-cli-1.0.0-linux-amd64.tar.gz"}, // CLI, not daemon
		{Hash: "bb", Name: "condurad-stuff.tar.gz"},                // wrong name shape
	}
	_, err := BuildManifestFromChecksums("v1.0.0", "stable", "https://x", "", entries)
	if err == nil {
		t.Error("BuildManifestFromChecksums with no condurad archives returned nil; want error")
	}
	if !strings.Contains(err.Error(), "no condurad archives") {
		t.Errorf("error %q should mention 'no condurad archives'", err.Error())
	}
}

// TestBuildManifestFromChecksums_StripsVPrefix pins the version
// normalization contract: the "v" prefix (per Go convention) MUST
// be stripped from the version field. Without this pin, a manifest
// with "v1.0.0" would compare unequal to a runtime version of "1.0.0"
// (no v).
func TestBuildManifestFromChecksums_StripsVPrefix(t *testing.T) {
	p, err := BuildManifestFromChecksums("v1.0.0", "stable", "https://x", "", []ChecksumEntry{
		{Hash: "aa", Name: "condurad-1.0.0-linux-amd64.tar.gz"},
	})
	if err != nil {
		t.Fatal(err)
	}
	if p.Version != "1.0.0" {
		t.Errorf("Version = %q, want \"1.0.0\" (v prefix stripped)", p.Version)
	}
}

// TestParseChecksums_BadLineErrors pins the per-line validation
// contract: any line that doesn't have at least two whitespace-
// separated fields MUST error (defends against a corrupted
// checksums.txt where a line was truncated).
func TestParseChecksums_BadLineErrors(t *testing.T) {
	cases := []struct {
		name string
		raw  string
	}{
		{"single-token", "aabbcc\n"},                  // only hash, no filename
		{"trailing-newline-ok", "aabb cc\n"},          // valid one-line
		{"bad-line-among-good", "aabb cc\ngarbage\n"}, // second line bad
		{"only-filename", "filename-only.tar.gz\n"},   // no hash
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, err := ParseChecksums(tc.raw)
			// "trailing-newline-ok" is the only valid one.
			if tc.name == "trailing-newline-ok" {
				if err != nil {
					t.Errorf("expected no error for valid input, got %v", err)
				}
				return
			}
			if err == nil {
				t.Errorf("ParseChecksums with bad input %q returned nil; want error", tc.raw)
			}
			if !strings.Contains(err.Error(), "bad checksum line") {
				t.Errorf("error %q should mention 'bad checksum line'", err.Error())
			}
		})
	}
}

// TestParseChecksums_EmptyInputOK pins the empty-input contract:
// an empty checksums.txt MUST return an empty slice with no error.
// A regression that errored on empty input would break the early-
// release case where checksums.txt is generated only after the
// first artifact is built.
func TestParseChecksums_EmptyInputOK(t *testing.T) {
	for _, raw := range []string{"", "\n", "\n\n\n"} {
		entries, err := ParseChecksums(raw)
		if err != nil {
			t.Errorf("ParseChecksums(%q) errored: %v", raw, err)
		}
		if len(entries) != 0 {
			t.Errorf("ParseChecksums(%q) returned %d entries; want 0", raw, len(entries))
		}
	}
}
