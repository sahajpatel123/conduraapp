# Release Signing Key Management

## The Crown Jewel

The Ed25519 update-signing key is the single most sensitive secret in the
Synaptic project. If it leaks, every user can be pushed a malicious update —
the auto-update system becomes a universal RCE vector. Treat it accordingly.

## Key Generation (offline, one-time)

```bash
# Generate on an air-gapped machine. Never on a dev machine.
openssl genpkey -algorithm ed25519 -out synaptic-update-private.pem
openssl pkey -in synaptic-update-private.pem -pubout -out synaptic-update-public.pem

# Extract the raw 32-byte public key for embedding in the binary.
openssl pkey -in synaptic-update-public.pem -pubin -outform DER | tail -c 32 | xxd -p
```

## Where the Keys Live

| Key | Location | Access |
|-----|----------|--------|
| Private key | Hardware token (YubiKey / TPM) + offline encrypted backup | One person, offline |
| Public key | Embedded in `internal/updater/updater.go` (`PublicKey` constant) | Read-only in repo |
| CI signing secret | GitHub Actions `secrets.UPDATE_SIGNING_KEY` (base64 of private key PEM) | CI only |

## Rotation

1. Generate a new keypair.
2. Deploy a release signed with BOTH the old and new keys (dual-signature window).
3. After a forced-update window (2 weeks), remove the old public key from the binary.
4. Revoke the old private key from CI secrets.

## Embedding the Public Key

Replace the placeholder in `internal/updater/updater.go`:
```go
var PublicKey = ed25519.PublicKey{0x...} // 32 bytes from `xxd -p`
```

## Emergency Rollback

If a key is compromised:
1. Revoke the compromised public key from the binary (emergency release).
2. Push an emergency update signed with the NEW key.
3. Optional: point the manifest at a known-good version URL until the emergency
   release propagates.
