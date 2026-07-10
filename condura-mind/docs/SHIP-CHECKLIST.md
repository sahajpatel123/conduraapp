# Condura v0.1.x — Ship checklist (operator)

> Code/CI work for the quality pass is complete. This page is the **human** path to a public tag.

## 0. Preconditions (already true on `main` when CI is green)

- [x] Marketing honesty freeze (no false signed/notarized / multi-provider / hard Layer-3 claims)
- [x] P1 reliability (streams, a11y, safego, i18n, npm)
- [x] P3 hygiene (log rotation, brand honesty, LOGBOOK archive)
- [x] Fail-closed `release.yml` for missing Apple secrets / `UPDATE_SIGNING_KEY`

## 1. Secrets (GitHub → Settings → Secrets)

| Secret | Status |
|--------|--------|
| `UPDATE_SIGNING_KEY` | Must be set (hex seed, 64 chars) |
| `APPLE_CERTIFICATE` | Required for notarized macOS DMG |
| `APPLE_CERTIFICATE_PASSWORD` | Required |
| `APPLE_DEVELOPER_ID_APPLICATION` | Required |
| `APPLE_ID` | Required |
| `APPLE_TEAM_ID` | Required |
| `APPLE_NOTARY_PASSWORD` | Required |

See `docs/release-keys.md` and `docs/release-runbook.md`.

## 2. Local packaging dry-run (no secrets)

```bash
make check-lockfiles
make release-dry-run-local
```

## 3. Draft pipeline dry-run (needs secrets)

```bash
git tag -a v0.0.0-test -m "pipeline dry-run"
git push origin v0.0.0-test
gh run watch   # release.yml
# cleanup
git push origin :refs/tags/v0.0.0-test && git tag -d v0.0.0-test
gh release delete v0.0.0-test --yes 2>/dev/null || true
```

## 4. Phase 15 (real Mac)

Follow `docs/phase15-verification.md` + `docs/macos-verification-runbook.md`.  
Agent CLI rows may already be PASS; **GUI / TCC / overlay** rows need a human.

## 5. Public tag (only if 1–4 green)

```bash
git tag -a v0.1.x -m "Condura v0.1.x"
git push origin v0.1.x
# keep release draft until spctl + checksums verified
```

## Explicitly deferred (P2 — after public v0.1.0)

Hybrid router, hard Layer-3 guard, subscription OAuth, Wave/DAG, Skills Hub, SDK, messaging channels.
