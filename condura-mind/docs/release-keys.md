# Release Signing Key Management

## The Crown Jewel

The Ed25519 update-signing key is the single most sensitive secret in the
Condura project. If it leaks, every user can be pushed a malicious update —
the auto-update system becomes a universal RCE vector. Treat it accordingly.

## Key Generation (offline, one-time)

CI expects the **private seed as 64 hex characters** (32 raw bytes), not a PEM
blob. Generate offline:

```bash
# Preferred: raw seed for GitHub secret UPDATE_SIGNING_KEY
openssl rand -hex 32
# → paste the 64-char hex string into GitHub Actions secrets as UPDATE_SIGNING_KEY

# Optional PEM form for offline archival (hardware token / encrypted backup):
openssl genpkey -algorithm ed25519 -out condura-update-private.pem
openssl pkey -in condura-update-private.pem -pubout -out condura-update-public.pem

# Extract the raw 32-byte public key for embedding in the binary:
openssl pkey -in condura-update-public.pem -pubin -outform DER | tail -c 32 | xxd -p -c 32
```

**Important:** The hex seed in `UPDATE_SIGNING_KEY` and the public key embedded
in `condura-app/internal/updater/updater.go` must be a matching keypair.
`release-verify.yml` and release signing fail closed if the secret is empty or
does not match the embedded public key.

## Where the Keys Live

| Key | Location | Access |
|-----|----------|--------|
| Private seed | Offline encrypted backup + hardware token if available | One person, offline |
| CI signing secret | GitHub Actions `secrets.UPDATE_SIGNING_KEY` (**64 hex chars**, Ed25519 seed) | CI only |
| Public key | Embedded in `condura-app/internal/updater/updater.go` (`PublicKey`) | Read-only in repo |

### Format mismatch (historical note)

Older docs said “base64 of private key PEM.” That format is **wrong for current
CI**. Use `openssl rand -hex 32` (or the raw 32-byte seed hex-encoded). The
release workflow error message also states: *“Ed25519 hex seed, 64 hex chars”*.

## Apple secrets (macOS codesign + notarize)

These are separate from the update key. Required by `.github/workflows/release.yml`
job `macos-sign` (all fail-closed if any are missing):

| Secret | Content |
|--------|---------|
| `APPLE_CERTIFICATE` | Base64 of Developer ID Application `.p12` |
| `APPLE_CERTIFICATE_PASSWORD` | Password for that `.p12` |
| `APPLE_DEVELOPER_ID_APPLICATION` | Codesign identity string |
| `APPLE_ID` | Apple ID email for notarytool |
| `APPLE_TEAM_ID` | 10-character Team ID |
| `APPLE_NOTARY_PASSWORD` | App-specific password |

Without these, CLI/daemon artifacts may still publish; **no notarized macOS
DMG** is uploaded (by design).

## Rotation

1. Generate a new keypair (hex seed + matching public key).
2. Prefer a dual-signature / forced-update window if users already trust the old key.
3. Embed the new public key; ship a release that accepts it.
4. After the window, remove the old public key; revoke the old seed from CI secrets.

## Embedding the Public Key

Replace the constant in `condura-app/internal/updater/updater.go`:

```go
var PublicKey = ed25519.PublicKey{ /* 32 bytes from xxd -p */ }
```

## Emergency Rollback

If a key is compromised:

1. Revoke the compromised public key from the binary (emergency release).
2. Push an emergency update signed with the **new** key only after the binary that trusts it has shipped.
3. Optional: pin the manifest to a known-good version URL until the emergency release propagates.
4. Never upload an unsigned manifest under a signed filename — CI refuses this (fail-closed).

## Related docs

- [`release-runbook.md`](./release-runbook.md) — full checklist, dry-run, tag, publish
- [`.github/workflows/release.yml`](../../.github/workflows/release.yml) — fail-closed gates
- [`.github/workflows/release-verify.yml`](../../.github/workflows/release-verify.yml) — ephemeral key CI
