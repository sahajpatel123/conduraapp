# Release Runbook — Condura v0.1.x

> **Fail-closed by design.** macOS notarized DMGs and Ed25519-signed update
> manifests are **not** published unless the required secrets are set.
> CLI/daemon archives can still ship from GoReleaser without Apple keys.

---

## 0. Secrets checklist (GitHub → Settings → Secrets and variables → Actions)

### Required for a complete public release

| Secret | Format | Used by | If missing |
|---|---|---|---|
| `UPDATE_SIGNING_KEY` | **64 hex chars** (Ed25519 seed = 32 raw bytes). Generate: `openssl rand -hex 32` | `release.yml` → Sign manifest; `release-verify.yml` → embedded-key-check | Job **fails closed** — no signed manifest uploaded |
| `APPLE_CERTIFICATE` | Base64-encoded `.p12` Developer ID Application cert | `macos-sign` job | Job **fails closed** — no macOS DMG published |
| `APPLE_CERTIFICATE_PASSWORD` | Password for the `.p12` | `macos-sign` | Fails closed |
| `APPLE_DEVELOPER_ID_APPLICATION` | Identity string, e.g. `Developer ID Application: Name (TEAMID)` | `codesign` | Fails closed |
| `APPLE_ID` | Apple ID email for notarytool | `macos-sign` | Fails closed |
| `APPLE_TEAM_ID` | 10-char Team ID | `macos-sign` | Fails closed |
| `APPLE_NOTARY_PASSWORD` | App-specific password (appleid.apple.com) | `notarytool` | Fails closed |

### Not wired in current CI (do not block on these)

| Secret | Status |
|---|---|
| `WINDOWS_SIGN_PFX` / `WINDOWS_SIGN_PASSWORD` | Documented historically; **not** used by `release.yml` today |
| `GPG_SIGNING_KEY` | **Not** used by `release.yml` today |

Key generation and rotation: see [`release-keys.md`](./release-keys.md).

---

## 1. Pre-release checklist

### Code health (no secrets)

- [ ] `go test -race -count=1 ./...` passes
- [ ] `golangci-lint run --timeout=5m ./...` is clean
- [ ] `release-verify` workflow green on `main` (snapshot + ephemeral Ed25519 roundtrip)
- [ ] `make release-snapshot` succeeds locally
- [ ] Frontend: `cd condura-gui/frontend && npm run check` (or project package manager)
- [ ] Marketing honesty freeze: no false signed/notarized / multi-provider claims (see `docs/roadmap-v0.2.0.md`)
- [ ] Phase 15 / on-device verification signed off for the target platforms (`docs/phase15-verification.md`)

### Keys (human-held)

- [ ] Ed25519 update key generated offline (see `release-keys.md`)
- [ ] Matching **public** key embedded in `condura-app/internal/updater/updater.go` (`PublicKey`)
- [ ] `UPDATE_SIGNING_KEY` set in repo secrets (hex seed; must match the public key)
- [ ] All 6 Apple secrets set (table above)
- [ ] Local dry-run of fail-closed path verified (section 2)

---

## 2. Dry-run (do this BEFORE any public `v0.1.0` / `v0.1.x` tag)

### 2a. Local packaging only (no GitHub secrets)

```bash
# One-shot local packaging dry-run (tests + snapshot + optional unsigned manifest)
make release-dry-run-local

# Or step-by-step:
make release-snapshot
make gen-manifest   # requires dist/checksums.txt from snapshot
make check-lockfiles
```

What this proves: GoReleaser config, artifact names, checksum generation.
What it does **not** prove: Apple notarization, real signed manifest, GitHub upload.

### 2b. CI path without production keys

- Push to a PR or open `release-verify` on `main` — uses **ephemeral** keys; must stay green.
- Confirms: snapshot build, sign/verify roundtrip with a throwaway key, updater unit tests.

### 2c. Draft tag dry-run (needs secrets; still not public)

Only after section 1 secrets are configured:

```bash
# Use a non-public test tag. Prefer draft release if your process allows.
git tag -a v0.0.0-test -m "Condura pipeline dry-run (do not publish)"
git push origin v0.0.0-test

# Monitor
gh run list --workflow=release.yml --limit 5
gh run watch
```

Expected outcomes when secrets **are** set:

| Job | Expectation |
|---|---|
| `goreleaser` | CLI / daemon / deb archives + checksums uploaded |
| `upload-gui` | Non-macOS GUI artifacts only (darwin skipped until notarized) |
| `macos-sign` | Codesign + notarytool + staple + DMG re-upload |
| `sign-manifest` | Ed25519-signed `update-manifest.signed.json` / `manifest.json` |

Expected outcomes when secrets **are missing** (fail-closed):

| Missing | Behavior |
|---|---|
| `UPDATE_SIGNING_KEY` | Sign-manifest step **exits 1**; no fake “signed” upload |
| Any Apple secret | `macos-sign` **exits 1**; **no** unsigned macOS DMG is published |

Cleanup after a successful dry-run:

```bash
# Delete test tag + draft release if created
git push origin :refs/tags/v0.0.0-test
git tag -d v0.0.0-test
gh release delete v0.0.0-test --yes 2>/dev/null || true
```

### 2d. Post-pipeline verify script

```bash
# After a real or test tag has artifacts
make verify-release TAG=v0.0.0-test
# or:
./condura-ops/scripts/verify-release-artifacts.sh v0.0.0-test
```

---

## 3. Real tag and build (only after dry-run is green)

```bash
git tag -a v0.1.0 -m "Condura v0.1.0"
git push origin v0.1.0

# CI: .github/workflows/release.yml
# Monitor: https://github.com/sahajpatel123/conduraapp/actions
```

**Do not** uncheck “draft” / run `gh release edit … --draft=false` until:

1. All release jobs green  
2. Checksums verified  
3. macOS `spctl` / codesign verified on a real Mac  
4. Marketing site download labels match real artifact names  

---

## 4. Verify artifacts

```bash
# Checksums (filename may be checksums.txt or SHA256SUMS depending on GoReleaser config)
curl -LO "https://github.com/sahajpatel123/conduraapp/releases/download/v0.1.0/checksums.txt"
# or SHA256SUMS — open the release page if unsure
shasum -a 256 -c checksums.txt   # macOS
# sha256sum -c checksums.txt     # Linux

# macOS: verify notarization after install
spctl -a -vv --type execute /Applications/Condura.app
codesign --verify --deep --strict /Applications/Condura.app
```

Windows Authenticode and Linux GPG verification are **not** part of the current
pipeline. Do not claim them on the marketing site until wired.

---

## 5. Publish the release

```bash
# GitHub Releases UI: convert draft → published
# Or:
gh release edit v0.1.0 --draft=false
```

---

## 6. Update manifest (CI path preferred)

On a green `release.yml` run, CI:

1. Generates `dist/update-manifest.json` from release assets  
2. Signs with `UPDATE_SIGNING_KEY` → `dist/update-manifest.signed.json`  
3. Uploads only if signing succeeded (fail-closed)

Manual fallback (local, with key):

```bash
export UPDATE_SIGNING_KEY=<64-hex-seed>

go run ./condura-app/cmd/gen-update-manifest generate \
  --version v0.1.0 \
  --checksums dist/checksums.txt \
  --base-url "https://github.com/sahajpatel123/conduraapp/releases/download/v0.1.0" \
  --out dist/update-manifest.json

go run ./condura-app/cmd/gen-update-manifest sign \
  dist/update-manifest.json \
  dist/update-manifest.signed.json
```

The daemon poller expects the stable signed manifest URL configured in
`internal/updater` / release notes — do not upload an **unsigned** file under
a signed name.

---

## 7. Post-release monitoring

- [ ] Opt-in crash telemetry for new patterns  
- [ ] GitHub Issues for install / permissions problems  
- [ ] Manifest download / update adoption  

---

## 8. Rollback

If a shipped version is bad:

```bash
# Point the signed manifest at a previous good version (or a hotfix tag).
# Auto-updater polls periodically and on launch.

# Emergency: publish a no-op / pin-previous manifest after re-signing with the same key.
```

Never ship an unsigned manifest to “fix” a broken release.

---

## 9. What “dry-run complete” means

| Gate | Owner | Status definition |
|---|---|---|
| `make release-snapshot` | Anyone | Local packaging OK |
| `release-verify` green | CI | Ephemeral sign path OK |
| 7 secrets configured | Human (repo admin) | Names + formats match this doc |
| `v0.0.0-test` pipeline green | Human + CI | Real sign + notary path OK |
| Public `v0.1.x` tag | Human product call | Only after Phase 15 + marketing honesty |

**No public tag until every row above is green.**
