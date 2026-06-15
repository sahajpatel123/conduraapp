// Command gen-update-manifest builds and/or signs Synaptic auto-update manifests.
//
// Usage:
//
//	gen-update-manifest sign <unsigned.json> <signed.json>
//	gen-update-manifest generate --version v0.1.0 --checksums dist/checksums.txt \
//	  --base-url https://github.com/org/repo/releases/download/v0.1.0 \
//	  --out dist/update-manifest.json
package main

import (
	"crypto/ed25519"
	"encoding/hex"
	"encoding/json"
	"flag"
	"fmt"
	"os"

	"github.com/sahajpatel123/synapticapp/internal/updater"
)

const manifestDirPerm = 0o750

func main() {
	if len(os.Args) < 2 {
		usage()
		os.Exit(2)
	}
	switch os.Args[1] {
	case "sign":
		if len(os.Args) != 4 {
			usage()
			os.Exit(2)
		}
		if err := signFile(os.Args[2], os.Args[3]); err != nil {
			fatal(err)
		}
	case "verify":
		if len(os.Args) != 3 {
			usage()
			os.Exit(2)
		}
		if err := verifyFile(os.Args[2]); err != nil {
			fatal(err)
		}
	case "generate":
		if err := generateCmd(os.Args[2:]); err != nil {
			fatal(err)
		}
	default:
		usage()
		os.Exit(2)
	}
}

func usage() {
	fmt.Fprintf(os.Stderr, `gen-update-manifest — build and sign auto-update manifests

  sign <unsigned.json> <signed.json>
  verify <signed.json>
  generate --version vX.Y.Z --checksums dist/checksums.txt --base-url URL --out dist/update-manifest.json

Requires UPDATE_SIGNING_KEY (hex Ed25519 seed or private key) for sign/generate.
`)
}

func generateCmd(args []string) error {
	fs := flag.NewFlagSet("generate", flag.ExitOnError)
	version := fs.String("version", "", "release version (v0.1.0)")
	channel := fs.String("channel", "stable", "update channel")
	checksums := fs.String("checksums", "", "path to checksums.txt")
	baseURL := fs.String("base-url", "", "release asset base URL")
	notes := fs.String("notes", "", "release notes snippet")
	out := fs.String("out", "dist/update-manifest.json", "output path")
	unsigned := fs.Bool("unsigned", false, "write manifest without signature")
	_ = fs.Parse(args)

	if *version == "" || *checksums == "" || *baseURL == "" {
		return fmt.Errorf("generate requires --version, --checksums, and --base-url")
	}
	raw, err := os.ReadFile(*checksums) //nolint:gosec // CLI tool
	if err != nil {
		return err
	}
	entries, err := updater.ParseChecksums(string(raw))
	if err != nil {
		return err
	}
	payload, err := updater.BuildManifestFromChecksums(*version, *channel, *baseURL, *notes, entries)
	if err != nil {
		return err
	}
	sm := updater.SignedManifest{
		Version:    payload.Version,
		Channel:    payload.Channel,
		Platforms:  payload.Platforms,
		Mandatory:  payload.Mandatory,
		MinVersion: payload.MinVersion,
		Notes:      payload.Notes,
	}
	if !*unsigned {
		priv, err := loadPrivateKey(os.Getenv("UPDATE_SIGNING_KEY"))
		if err != nil {
			return err
		}
		sig, err := updater.SignPayload(payload, priv)
		if err != nil {
			return err
		}
		sm.Ed25519Sig = sig
	}
	data, err := json.MarshalIndent(sm, "", "  ")
	if err != nil {
		return err
	}
	data = append(data, '\n')
	if err := os.MkdirAll(dirOf(*out), manifestDirPerm); err != nil {
		return err
	}
	return os.WriteFile(*out, data, 0o600) //nolint:gosec // CLI tool
}

func signFile(inPath, outPath string) error {
	raw, err := os.ReadFile(inPath) //nolint:gosec // CLI tool
	if err != nil {
		return err
	}
	var sm updater.SignedManifest
	if err := json.Unmarshal(raw, &sm); err != nil {
		return err
	}
	priv, err := loadPrivateKey(os.Getenv("UPDATE_SIGNING_KEY"))
	if err != nil {
		return err
	}
	p, err := sm.Payload()
	if err != nil {
		return err
	}
	sig, err := updater.SignPayload(p, priv)
	if err != nil {
		return err
	}
	sm.Ed25519Sig = sig
	out, err := json.MarshalIndent(sm, "", "  ")
	if err != nil {
		return err
	}
	out = append(out, '\n')
	return os.WriteFile(outPath, out, 0o600) //nolint:gosec // CLI tool
}

func verifyFile(path string) error {
	raw, err := os.ReadFile(path) //nolint:gosec // CLI tool
	if err != nil {
		return err
	}
	var sm updater.SignedManifest
	if err := json.Unmarshal(raw, &sm); err != nil {
		return err
	}
	if sm.Ed25519Sig == "" {
		return fmt.Errorf("manifest has no ed25519_sig")
	}
	p, err := sm.Payload()
	if err != nil {
		return err
	}
	if err := updater.VerifyPayload(p, updater.PublicKey, sm.Ed25519Sig); err != nil {
		return fmt.Errorf("signature invalid: %w", err)
	}
	fmt.Fprintf(os.Stderr, "manifest %s verified (version %s, channel %s)\n", path, sm.Version, sm.Channel)
	return nil
}

func loadPrivateKey(raw string) (ed25519.PrivateKey, error) {
	if raw == "" {
		return nil, fmt.Errorf("UPDATE_SIGNING_KEY not set")
	}
	b, err := hex.DecodeString(raw)
	if err != nil {
		return nil, fmt.Errorf("decode UPDATE_SIGNING_KEY: %w", err)
	}
	if len(b) == ed25519.SeedSize {
		return ed25519.NewKeyFromSeed(b), nil
	}
	if len(b) == ed25519.PrivateKeySize {
		return ed25519.PrivateKey(b), nil
	}
	return nil, fmt.Errorf("UPDATE_SIGNING_KEY must be %d-byte seed or %d-byte private key", ed25519.SeedSize, ed25519.PrivateKeySize)
}

func dirOf(path string) string {
	for i := len(path) - 1; i >= 0; i-- {
		if path[i] == '/' || path[i] == '\\' {
			return path[:i]
		}
	}
	return "."
}

func fatal(err error) {
	fmt.Fprintf(os.Stderr, "gen-update-manifest: %v\n", err)
	os.Exit(1)
}
