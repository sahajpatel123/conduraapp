package updater

import (
	"crypto/ed25519"
	"testing"
)

func TestParseChecksums(t *testing.T) {
	raw := `aabbcc condurad-1.0.0-linux-amd64.tar.gz
ddeeff condurad-1.0.0-darwin-arm64.zip
`
	entries, err := ParseChecksums(raw)
	if err != nil {
		t.Fatal(err)
	}
	if len(entries) != 2 {
		t.Fatalf("expected 2 entries, got %d", len(entries))
	}
}

func TestBuildManifestFromChecksums(t *testing.T) {
	entries := []ChecksumEntry{
		{Hash: "aa", Name: "condurad-1.0.0-linux-amd64.tar.gz"},
		{Hash: "bb", Name: "condurad-1.0.0-darwin-arm64.tar.gz"},
		{Hash: "cc", Name: "condura-cli-1.0.0-linux-amd64.tar.gz"},
	}
	p, err := BuildManifestFromChecksums("v1.0.0", "stable", "https://example.com/v1.0.0", "notes", entries)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Platforms) != 2 {
		t.Fatalf("expected 2 platforms, got %d", len(p.Platforms))
	}
	if p.Platforms["linux/amd64"].SHA256 != "aa" {
		t.Error("linux amd64 hash mismatch")
	}
}

func TestMultiPlatformManifestSignVerify(t *testing.T) {
	pub, priv, err := ed25519.GenerateKey(nil)
	if err != nil {
		t.Fatal(err)
	}
	sm := SignedManifest{
		Version: "1.0.0",
		Channel: "stable",
		Platforms: map[string]PlatformArtifact{
			PlatformKey(): {DownloadURL: "http://example.com/bin", SHA256: "abc"},
		},
	}
	p, err := sm.Payload()
	if err != nil {
		t.Fatal(err)
	}
	sig, err := SignPayload(p, priv)
	if err != nil {
		t.Fatal(err)
	}
	if err := VerifyPayload(p, pub, sig); err != nil {
		t.Fatal(err)
	}
}

func TestResolveArtifactMultiPlatform(t *testing.T) {
	sm := SignedManifest{
		Platforms: map[string]PlatformArtifact{
			PlatformKey(): {DownloadURL: "http://x/y", SHA256: "deadbeef"},
		},
	}
	url, sha, err := sm.ResolveArtifact()
	if err != nil {
		t.Fatal(err)
	}
	if url != "http://x/y" || sha != "deadbeef" {
		t.Fatalf("got %q %q", url, sha)
	}
}
