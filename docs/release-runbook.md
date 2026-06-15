# Release Runbook — Synaptic v0.1.0

## Pre-release checklist

- [ ] Phase 11 backup/restore passes tests (no data loss)
- [ ] Phase 12 on-device verification passes on real macOS/Windows/Linux machines
- [ ] All 48+ packages pass `go test -race ./...`
- [ ] `golangci-lint run ./...` is clean
- [ ] Release signing key has been generated (see `docs/release-keys.md`)
- [ ] Public key is embedded in `internal/updater/updater.go` (`PublicKey`)
- [ ] CI secrets are set:
  - `APPLE_DEVELOPER_ID_APPLICATION`
  - `APPLE_NOTARY_USER`, `APPLE_NOTARY_PASSWORD`, `APPLE_TEAM_ID`
  - `WINDOWS_SIGN_PFX`, `WINDOWS_SIGN_PASSWORD`
  - `GPG_SIGNING_KEY`
  - `UPDATE_SIGNING_KEY`

## Tag and build

```bash
# Tag the release
git tag -a v0.1.0 -m "Synaptic v0.1.0"
git push origin v0.1.0

# CI builds, signs, notarizes, and uploads to GitHub Releases.
# Monitor: https://github.com/sahajpatel123/synapticapp/actions
```

## Verify artifacts

```bash
# Download and verify checksums
curl -LO https://github.com/sahajpatel123/synapticapp/releases/download/v0.1.0/SHA256SUMS
sha256sum -c SHA256SUMS

# macOS: verify notarization
spctl -a -v /Applications/Synaptic.app

# Windows: verify Authenticode
signtool verify /pa /v synaptic.exe

# Linux: verify GPG
gpg --verify synaptic-0.1.0-linux-amd64.deb.sig synaptic-0.1.0-linux-amd64.deb
```

## Publish the release

```bash
# In GitHub Releases UI: uncheck "draft"
# Or via CLI:
gh release edit v0.1.0 --draft=false
```

## Publish the update manifest

GoReleaser writes an unsigned `dist/update-manifest.json` (multi-platform).
Sign and upload with:

```bash
export UPDATE_SIGNING_KEY=<hex-ed25519-seed>
go run ./cmd/gen-update-manifest sign dist/update-manifest.json dist/update-manifest.signed.json
```

Or generate from checksums manually:

```bash
go run ./cmd/gen-update-manifest generate \
  --version v0.1.0 \
  --checksums dist/checksums.txt \
  --base-url "https://github.com/sahajpatel123/synapticapp/releases/download/v0.1.0" \
  --out dist/update-manifest.signed.json
```

Push the signed manifest to the update server (stable URL for the daemon poller):
```json
{
  "version": "0.1.0",
  "channel": "stable",
  "download_url": "https://github.com/sahajpatel123/synapticapp/releases/download/v0.1.0/synaptic-0.1.0-darwin-arm64.tar.gz",
  "sha256": "<SHA256 from SHA256SUMS>",
  "ed25519_sig": "<signature from signing tool>",
  "mandatory": false,
  "notes": "Initial release of Synaptic v0.1.0"
}
```

## Post-release monitoring

- [ ] Check opt-in crash telemetry for new crash patterns
- [ ] Monitor GitHub Issues for installation problems
- [ ] Check update adoption rates via manifest download counts

## Rollback (if needed)

If v0.1.0 has a critical bug:

```bash
# Point the manifest at the previous version or a hotfix.
# The auto-updater will detect it on the next poll (every 6h + on launch).

# Emergency: push an empty manifest to disable updates temporarily.
echo '{"version":"none","channel":"stable"}' > manifest.json
```

## Dry-run this runbook BEFORE the real release

Run through every step with a `v0.0.0-test` tag to verify the pipeline works
end-to-end before tagging `v0.1.0`.
