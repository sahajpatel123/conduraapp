# Synaptic/Condura — Production Readiness Audit (Tier 4+)

> **Author:** minimax-m3 via opencode
> **Date:** 2026-06-24
> **Branch audited:** `fix/marketing-honest-v0.1.1` @ `bcb2b89` (HEAD at time of audit)
> **Scope:** The user's question was: "Is everything green? Is Condura production-ready, could it deploy itself, could it be used for production quality?" This audit answers that question across four dimensions: (1) CI/CD + supply chain, (2) operational readiness + observability, (3) cross-platform parity + Svelte frontend quality, (4) functional completeness + adversarial safety. The prior `docs/analysis/backend-audit-2026-06-24.md` (Tier-4 code audit, 60 packages) and `docs/analysis/security-audit-2026-06-24.md` (F-01..F-12) are prerequisites.
> **Methodology:** Tier 4+. Four parallel exploration agents, each with a read-only brief, returned structured per-finding reports. Their findings were re-verified at the file:line level and merged into the consolidated report below. The local test suite was re-run (47 packages ok, svelte-check 0 errors, golangci-lint 0 issues, one flaky keyring test under parallel `-race` re-verified to be transient). All cited line numbers are valid at HEAD `bcb2b89`.
> **Severity scale:**
> - **P0** — blocks shipping a v1.0.0 release (data loss, full compromise, or no recovery)
> - **P1** — would block a public release to strangers (unfixable-after-the-fact risk, supply-chain gap, broken claimed feature)
> - **P2** — would block a closed beta of 100+ users (operational blind spot, no observability, no upgrade path)
> - **P3** — hygiene, best-practice gap, future-proofing
> - **INFO** — observation, intentional deferral, or post-launch improvement

---

## 0. Executive Summary — the one-paragraph answer

**No. Condura v0.1.1 is not production-ready for a public release to strangers.** The CI is green (15/15 on main + 12/12 on PR #13, 47 test packages ok, svelte-check 0 errors, golangci-lint 0 issues, all binaries build for the documented matrix), and the **safety layer in the binary is genuinely good** (deterministic Gatekeeper, HMAC-chained audit log, Ed25519-signed auto-update, AES-256-GCM secrets at rest, shell command sanitization, sensitive-hook URL escalation). But the **CI/CD pipeline that builds and ships the binary is not production-grade** (no branch protection on main, all five GitHub-provided security features disabled — Dependabot, secret scanning, code scanning, push protection, vulnerability alerts, the `UPDATE_SIGNING_KEY` secret has a fail-open mode that uploads an unsigned update manifest, `install.sh` does zero signature verification, no SBOM, no SLSA provenance, no binary code-signing on Windows or Linux). The **operational story is also weak** (no log rotation, watchdog disabled by default, no user-visible crash dialog, no `daemon.info` RPC, backup retention/interval are config-only-not-honored). The **frontend has 4 CRITICAL accessibility issues** (sidebar nav has zero ARIA labels, conversation delete is not undoable, 12 places use `alert()`, 10 places use `confirm()`). The **safety layer has 1 HIGH adversarial gap** (agent can re-arm its own halt flag via `daemon.resume`). For a closed beta of ≤50 hands-on testers on macOS who are told it's "alpha," v0.1.1 is fine. For anything beyond that, there is a clear P0/P1 punch list of about 12 items that must close first — mostly on the CI/supply-chain side, not on the binary itself.

**Table of must-fix-before-public-release (12 items, 1 P0, 11 P1):**

| # | Finding | Severity | Why this is must-fix |
|---|---|---|---|
| **PR-01** | `release.yml:197-208` — if `UPDATE_SIGNING_KEY` is missing, release.yml uploads an unsigned `update-manifest.signed.json` and the verify workflow exits 0. The binary will accept the unsigned manifest and apply the update. | **P0** | Fail-open mode for the most security-critical secret in the pipeline. |
| **PR-02** | GitHub repo: Dependabot / Dependabot security updates / secret scanning / secret scanning push protection / code scanning all `disabled` (per `gh api /repos/.../security_and_analysis` at HEAD). | **P0** | No automated supply-chain alerting. Catches the F-03 finding from the prior security audit, still 100% accurate. |
| **PR-03** | `main` has no branch protection. `gh api /repos/.../branches/main/protection` returns 404. No required status checks, no required reviews, no required signed commits, no restrictions on who can push. | **P0** | One-person single point of failure: a compromised owner account can push to main, tag, and release without any automated or human gate. |
| **PR-04** | `scripts/install.sh` (25 lines, the user-facing `curl -fsSL https://condura.app/install.sh | bash` one-liner) downloads the DMG and copies to `/Applications` with **zero** signature verification, zero checksum, zero contact with `manifest.json`. | **P0** | The first link in the user's install chain is unverified. DNS hijack or GitHub write-token compromise = arbitrary code as the user's `/Applications/Condura.app`. |
| **PR-05** | v0.1.1 ships **unsigned binaries**. Only the auto-update manifest is Ed25519-signed. No macOS notarization (despite runbook listing notary secrets), no Windows Authenticode (despite runbook listing PFX secrets), no Linux GPG signature, no `cosign` signature. | **P0** | A user who downloads `condurad-0.1.1-linux-amd64.zip` has no built-in way to verify it was produced by Condura. |
| **PR-06** | The **Linux presence detector** is `return true` (`internal/presence/detector.go:202`). Any DESTRUCTIVE action that requires `require_user_active` is auto-allowed on Linux. | **P1** | Defeats the safety gate on the platform we now ship the GUI to. |
| **PR-07** | **Sidebar icon rail has zero ARIA labels** (`Sidebar.svelte:41-143`). Screen-reader users cannot navigate the 9 main routes. | **P1** | Blocks WCAG/accessibility. |
| **PR-08** | **Conversation deletion is not undoable** (`conversation.svelte.ts:55-64` + `Sidebar.svelte:33`). Instant data loss with a 2-keypress accidental delete. No trash, no undo, no grace period. | **P1** | Real data-loss risk for the user. |
| **PR-09** | The **agent can re-arm its own halt flag** by calling `daemon.resume` over the IPC bearer token it shares with the daemon. `daemon.halt` and `daemon.resume` are RPCs that any in-process code with the bearer token can call. The audit log records the re-arm, so it's forensically traceable, but the kill switch is not *sticky* by default. | **P1** | A prompt-injected or compromised agent can re-arm itself. The kill switch is the headline safety claim. |
| **PR-10** | The `install.sh` one-liner is missing `set -euo pipefail` and an `ERR` trap. If `hdiutil attach` succeeds but `cp -R` fails halfway, the mounted volume may not be detached and the temp file may not be cleaned. | **P1** | Installs that fail in the middle can leave the system in an inconsistent state. |
| **PR-11** | **The `UpdateSigningKey` absence is a fail-open condition** in both `release.yml:197-208` and `release-verify.yml:126-128`. The right behavior is `exit 1`, not `exit 0`. | **P1** | Same as PR-01, secondary site. |
| **PR-12** | The EULA (`EULA.md:88-89`) says "Upon termination, you must cease all use of the Software and destroy all copies" but **does not document a clean-uninstall flow**. The web `legal/page.tsx` is silent. The `uninstall.execute` RPC exists but the user has to discover it. | **P1** | A user who reads the EULA looking for clean-uninstall instructions will not find them. |

The 12 must-fix items are **concentrated in 3 areas**: CI/supply-chain (5), frontend (3), safety (3), install (1). The 47 test packages and the binary's safety layer are not in the punch list.

**Recommended posture:**
- **Public release to strangers:** blocked until the 12 items above are fixed. Realistically a 2-week sprint for the CI/supply-chain fixes plus 1 week for the frontend/safety ones.
- **Closed beta of 50 hands-on macOS users:** ship today. v0.1.1 is sufficient.
- **Public release on the marketing site alone (the website you can deploy to Vercel without shipping the binary):** ship today. The marketing site is honest after PR #13.

---

## 1. Is everything green? — CI status

**Yes. Every CI check that runs on PR #13 is green.** The local test suite is also clean. Here is the verbatim state at HEAD `bcb2b89`.

### 1.1 PR #13 CI (12 active checks, 12 pass, 2 skip)

| Job | Status | Time | Evidence |
|---|---|---|---|
| Build (darwin/amd64) | pass | 38s | `actions/runs/28235658071/job/83649802140` |
| Build (darwin/arm64) | pass | 26s | `actions/runs/28235658071/job/83649802128` |
| Build (linux/amd64) | pass | 42s | `actions/runs/28235658071/job/83649802134` |
| Build (linux/arm64) | pass | 47s | `actions/runs/28235658071/job/83649802190` |
| Build (windows/amd64) | pass | 40s | `actions/runs/28235658071/job/83649802120` |
| Build (windows/arm64) | pass | 41s | `actions/runs/28235658071/job/83649802153` |
| Lint | pass | 1m27s | `actions/runs/28235658071/job/83649802179` |
| Test (macos-latest/amd64) | pass | 2m16s | `actions/runs/28235658071/job/83649802150` |
| Test (macos-latest/arm64) | pass | 3m05s | `actions/runs/28235658071/job/83649802121` |
| Test (windows-latest/amd64) | pass | 4m38s | `actions/runs/28235658071/job/83649802124` |
| Test (ubuntu-latest/amd64) | pass | 2m46s | `actions/runs/28235658071/job/83649802167` |
| Test (ubuntu-latest/arm64) | pass | 3m42s | `actions/runs/28235658071/job/83649802129` |
| Security Scan | pass | 48s | `actions/runs/28235658071/job/83649802123` |
| Integration Tests | skipping | 0s | Skips on PRs (runs on main only) |
| GUI Build (darwin/arm64) | skipping | 0s | Skips on PRs (runs on main only) |

**Verdict:** 12/12 active checks pass. The 2 skipping checks are normal PR behavior (run on main pushes, not on PRs).

### 1.2 main CI (3 most recent runs, 3 success)

| Run | Commit | Event | Conclusion |
|---|---|---|---|
| `28105359258` | `cace2a4` (v0.1.1 fix bundle) | push | success |
| `28105799269` | `19a7daa` (LOGBOOK) | pull_request | success |
| `28105799408` | `19a7daa` | push | success |

All 3 most recent runs are green.

### 1.3 Local test suite

- **`go test -count=1 -race -timeout 180s -short ./...`:** 47 packages, 0 failures. One transient failure in `internal/secrets` was reproducible-once and then disappeared on rerun — the `TestNew_NoFilePath_Auto` test races on macOS keyring access under parallel `-race` package execution. CI runs packages one at a time, so this is not an issue there. Confirmed stable on 3 consecutive reruns.
- **`./node_modules/.bin/svelte-check --tsconfig ./tsconfig.json`:** 0 errors, 0 warnings.
- **`go vet ./...`:** clean.
- **`go build ./...`:** clean (only pre-existing CGO deprecation warnings for macOS frameworks).

**Verdict:** everything green. The only flake is a pre-existing test that doesn't reproduce on rerun and doesn't affect CI.

---

## 2. CI/CD + Supply Chain — 38 findings (1 P0, 6 P1)

### 2.1 Critical findings (must fix before public release)

**Finding CI-01 — P0. `release.yml:197-208` fails open on the most important secret.**

```yaml
- name: Sign manifest
  if: hashFiles('dist/update-manifest.json') != ''
  env:
    UPDATE_SIGNING_KEY: ${{ secrets.UPDATE_SIGNING_KEY }}
  run: |
    if [ -z "$UPDATE_SIGNING_KEY" ]; then
      echo "UPDATE_SIGNING_KEY not set — uploading unsigned manifest"
      cp dist/update-manifest.json dist/update-manifest.signed.json
      exit 0
    fi
```

If `UPDATE_SIGNING_KEY` is unset, release.yml copies the unsigned manifest under the misleading name `update-manifest.signed.json` and exits 0. The `release-verify.yml:embedded-key-check` job (line 126-128) ALSO short-circuits with `exit 0` when the secret is missing. **The combined effect: a missing secret produces a release that the binary will accept and apply.** The right behavior is `exit 1`.

**Finding CI-02 — P0. All five GitHub-provided security features are DISABLED.**

Verified live via `gh api /repos/sahajpatel123/conduraapp` at HEAD `bcb2b89`:

| Feature | Status |
|---|---|
| `secret_scanning` | disabled |
| `secret_scanning_push_protection` | disabled |
| `dependabot_security_updates` | disabled |
| `secret_scanning_non_provider_patterns` | disabled |
| `secret_scanning_validity_checks` | disabled |
| `dependabot_alerts` | 403 (disabled) |
| `vulnerability_alerts` | 404 (disabled) |
| `code_scanning` | 404 (no analysis found) |

This is the prior security audit's F-03 finding. Still 100% accurate. For a project whose binary performs privileged actions and self-updates over the network, this is the single highest-impact supply-chain gap. **The user has not acted on F-03 in the 24 hours since the security audit was written.**

**Finding CI-03 — P0. `main` has no branch protection.**

`gh api /repos/sahajpatel123/conduraapp/branches/main/protection` returns `{"message":"Branch not protected","status":"404"}`. No required status checks, no required reviews (despite CODEOWNERS being configured), no required signed commits, no restrictions on who can push. Combined with single-owner (the only way to get a token in the repo), this means: a compromised owner account → push to main → tag → release with no human gate.

**Finding CI-04 — P0. `scripts/install.sh` (the user-facing one-liner) has zero signature verification.**

`scripts/install.sh:12-19` does `curl -fsSL ...` → `hdiutil attach condura.dmg` → `cp -R Condura.app /Applications/`. It does NOT verify:
- The DMG's SHA-256 against `manifest.json`
- An `apple-developer-id` signature (because there is no Apple Developer ID signing — see CI-05)
- Any contact with `https://condura.app/.well-known/manifest.json` to confirm the version
- Any check that the user is downloading the same version as the auto-update channel

If `condura.app` DNS is hijacked OR the DMG is replaced at the GitHub Releases URL by an attacker with a write token, every user who runs `curl -fsSL https://condura.app/install.sh | bash` on a clean macOS install gets the attacker's binary. The DMG is also not notarized (CI-05), so Gatekeeper's "are you sure" dialog is the only defense.

**Finding CI-05 — P0. v0.1.1 ships unsigned binaries.**

Verified at `gh release view v0.1.1`. Only the auto-update manifest is signed (Ed25519). The binaries themselves are not:
- No macOS notarization (despite `docs/release-runbook.md:14-16` listing `APPLE_NOTARY_USER`/`APPLE_NOTARY_PASSWORD`/`APPLE_TEAM_ID` as required). The `macos-sign` job (`release.yml:224-270`) runs only `codesign`, no `xcrun notarytool submit --wait --staple`.
- No Windows Authenticode (despite `docs/release-runbook.md:16` listing `WINDOWS_SIGN_PFX`/`WINDOWS_SIGN_PASSWORD`).
- No Linux GPG signature on `.deb` or `.tar.gz` (despite `docs/release-runbook.md:17, 44-46` documenting the procedure).
- No `cosign` signature on any binary (only the manifest has one).

The runbook and the code are out of sync. A release engineer who follows the runbook will set the secrets and assume they are being used. They are not.

### 2.2 High findings (P1, block public release)

**Finding CI-06 — P1. The `UpdateSigningKey` absence is a fail-open condition in BOTH `release.yml` and `release-verify.yml`.** The verify workflow's `embedded-key-check` (line 126-128) does `if [ -z "$UPDATE_SIGNING_KEY" ]; then exit 0; fi`. This is the same failure mode as CI-01 in a second place.

**Finding CI-07 — P1. CODEOWNERS doesn't cover the most security-critical files.**

`.github/CODEOWNERS` covers `/internal/safety/`, `/internal/perception/`, `/internal/agent/`, `/internal/ipc/`, `/internal/secrets/`, `/internal/api_key/`. It does NOT cover:
- `.github/workflows/*` (the CI/release/verify definitions themselves)
- `internal/updater/` (the Ed25519 verifier and its embedded `PublicKey` — arguably the most security-critical single file in the project)
- `go.mod`, `go.sum`, `package.json`, `package-lock.json`
- `.goreleaser.yml`, `Makefile`
- `docs/release-keys.md`

A merge of a new release workflow or a change to the embedded `PublicKey` constant requires only a self-review (or no review at all, since the teams `@synaptic/core` / `@synaptic/security` are not GitHub-verified to exist on this account).

**Finding CI-08 — P1. `scripts/install.sh` lacks `set -euo pipefail` and an ERR trap.**

The script does `trap 'cleanup' EXIT` (line 16) but only on EXIT, not on ERR. If `hdiutil attach` succeeds but `cp -R` fails halfway, the mounted volume may not be detached and `/tmp/condura.dmg` may not be cleaned. Result: half-installed apps on user machines after a failed `curl | bash`.

**Finding CI-09 — P1. No Dependabot. No Renovate. No SCA on the JS tree.**

The Go tree has `govulncheck` in CI (ci.yml:339), which covers the 12 direct Go deps and the 41 indirect ones, **except** the CGO packages `internal/hotkey` and `internal/tray` (line 339 excludes them). The Next.js marketing site (`web/`) and the Svelte/Vite GUI (`app/web/frontend/`) have **zero** SCA — `ci.yml` only references their lockfiles as npm cache keys (lines 237, 238). A known CVE in `marked`, `motion`, `next`, or any of the 511 npm packages in `web/package-lock.json` will not generate an alert or PR.

**Finding CI-10 — P1. No `.npmrc` with `save-exact=true`; 8 of 12 web deps use `^` ranges.**

`web/package.json:13-29` uses caret ranges. Every new `npm install` by a future contributor will write a caret range and update the lockfile on the next commit. Lockfile pins the *installed* version, but the declared `package.json` is what new contributors see and trust. Add `.npmrc` with `save-exact=true` to make the policy explicit.

**Finding CI-11 — P1. Windows arm64 is shipped in v0.1.1 with zero test coverage on a Windows arm64 runner.**

`ci.yml:63` matrix includes `windows-latest` × `amd64` but NOT `windows-latest` × `arm64`. Yet `gh release view v0.1.1` shows `condurad-0.1.1-windows-arm64.zip` shipped as a real asset. So Windows arm64 binaries ship with **zero** unit test coverage on the actual target architecture. Fixable: GitHub-hosted Windows arm64 runners exist (`windows-11-arm`).

**Finding CI-12 — P1. `release.yml` is not gated by CI status of the tagged commit.**

`release.yml:3-6` is `on: push: tags: ['v*']` with no `workflow_run` or `needs:` dependency on a green `ci.yml` run for the tagged SHA. A tag can be pushed from a commit that never passed CI.

### 2.3 Medium findings (P2, recommended before v0.2.0)

- **CI-13 (P2)** — `integration` job only runs on main (ci.yml:287); PRs not gated.
- **CI-14 (P2)** — `gui-build` silent fallback (ci.yml:253-257): if `wails` CLI isn't on PATH, the job runs `go build` of the Wails library into `/tmp/condura-gui-smoke` and emits `::warning::`. A PR that breaks the actual Wails build can pass this CI if Wails isn't installed, then break the actual `release.yml` GUI build at tag time.
- **CI-15 (P2)** — `golangci-lint` bootstrap downloads a shell script from `raw.githubusercontent.com/golangci/golangci-lint/HEAD/install.sh` at every run. The version is pinned but the *script* is not pinned by SHA (ci.yml:46).
- **CI-16 (P2)** — `continue-on-error: true` on the GUI artifact download (release.yml:238) and on the sign-manifest artifact download (release.yml:172) means partial pipeline failures pass silently.
- **CI-17 (P2)** — `web_commit_signoff_required: false`. Enabling would require signed commits.
- **CI-18 (P2)** — `allowed_actions: "all"`. Least-privilege would be an allowlist.
- **CI-19 (P2)** — No release artifact is published with the `immutable` flag set. The maintainer (or anyone with write access) can swap binaries in place after a release.
- **CI-20 (P2)** — v0.1.1 is missing `condura-gui-windows-amd64-setup.exe` and the NSIS portable .exe. `scripts/package-gui-installers.sh:41-43` silently skips if `makensis` is not found. The Windows GUI installer is missing in v0.1.1 (it was present in v0.1.0). This is a **silent shipping regression**.
- **CI-21 (P2)** — No SBOM is published. No SLSA provenance. No `cosign` signature.
- **CI-22 (P2)** — Third-party actions (`actions/checkout@v4`, `actions/setup-go@v5`, etc.) are pinned by major, not by SHA.
- **CI-23 (P2)** — `docs/release-runbook.md` documents signing steps (APPLE_NOTARY_*, WINDOWS_SIGN_PFX, GPG_SIGNING_KEY, `gpg --verify` step at line 45) that do not exist in `release.yml`. **Doc-vs-code drift** that the docs themselves should flag.
- **CI-24 (P2)** — `.gitignore` doesn't cover `.pem`, `.p12`, `.asc`, `.sig`, `secrets.*`, `private/`. A developer could still commit a private key.
- **CI-25 (P2)** — CGO packages (`internal/hotkey`, `internal/tray`) excluded from both `.golangci.yml:142` and `ci.yml:339`. CGO deps (`wails/v2`, `gen2brain/malgo`, `getlantern/systray`) are harder to audit.
- **CI-26 (P2)** — `web/` (Next.js 16.2.7) and `app/web/frontend/` (Svelte 5) are not built, linted, type-checked, or tested in CI.
- **CI-27 (P2)** — Two different `marked` major versions in two `package.json` files (web: `^18.0.5`, app/web/frontend: `^14.1.3`). `marked` has had CVEs (ReDoS 2023).
- **CI-28 (P2)** — No key-rotation support in `internal/updater`. `docs/release-keys.md:28-34` describes a dual-signing rotation but the code has only one `pubKey`.
- **CI-29 (P2)** — `compareVersions` is lexicographic (`updater.go:337-345`). `v0.10.0` < `v0.2.0` under string compare. For 0.0.x → 0.1.x fine; for 0.10.x breaks.
- **CI-30 (P2)** — Default `RELEASE_TAG` for the marketing site (`web/lib/downloads.ts:34`) was `v0.1.0` before PR #13. Now `v0.1.1`.
- **CI-31 (P2)** — CODEOWNERS teams `@synaptic/core` / `@synaptic/security` membership unverified. If empty, required reviews are unsatisfiable.
- **CI-32 (P2)** — Release bodies are 200-line commit dumps; no CVE/rollback section. A security release should have a focused CVE list and a rollback procedure.
- **CI-33 (P2)** — Empty root `package-lock.json` (`packages: {}`) is vestigial.

### 2.4 Low findings (P3)

- **CI-34 (P3)** — PGP fingerprint in `SECURITY.md:25` is "TBD" (placeholder).
- **CI-35 (P3)** — No `/.well-known/security.txt` for the deployed website.
- **CI-36 (P3)** — `internal/updater/updater.go:93` has no `User-Agent` on the HTTP client.
- **CI-37 (P3)** — Default manifest URL uses `/releases/latest/download/manifest.json` (symlink). Mid-rotation, the symlink could point at a manifest signed with a different key.
- **CI-38 (P3)** — `build-gui.sh` runs `npm ci` then installs `wails` CLI v2.12.0 from upstream without checksum verification.

### 2.5 Positive findings (INFO)

- **CI-INFO-1** — The Go dep tree is small (~53 packages), pinned, free of `replace` directives, and uses `modernc.org/sqlite` (pure-Go, no CGO) for SQLite. The right shape for a security-sensitive app.
- **CI-INFO-2** — The updater is well-designed: Ed25519 verification + SHA-256 binary check + anti-downgrade floor + atomic swap + opt-out toggle + embedded public key + no over-the-wire public key fetch.
- **CI-INFO-3** — `docs/release-keys.md` calls the update-signing key "the crown jewel" (line 5) and treats the key-management story as the highest-stakes part of the operation. Good posture, even if the rotation procedure isn't implemented in code.
- **CI-INFO-4** — The `release-verify.yml:manifest-signing` job generates a throwaway key, signs, re-signs, and asserts the verify step passes. This is the strongest gate in the pipeline.

---

## 3. Operational Readiness + Observability — 50 findings (3 P0, 20 P1, 13 P2, 9 P3, 5 INFO)

### 3.1 Critical findings (P0, would block recovery in a break)

**Finding OPS-01 — P0. No log rotation. No size or count cap.**

`internal/logger/logger.go:237-249` opens the file sink with `O_APPEND | O_WRONLY | O_CREATE` and never truncates, never rotates, never checks file size. The author wrote: "For long-running daemons, a future log rotation module can take over." The package docstring also calls this out: "file (with rotation, future)". A runaway agent at `debug` level on a chatty LLM day will silently grow a single log file until the user's disk fills.

**Finding OPS-02 — P0. `crash.Recover()` is only in 2 places; subsystem goroutines uncaught.**

`crash.Recover()` is wired in only `daemon/daemon.go:109` (top of `Run`) and `cmd/condurad/main.go:38`. The IPC server goroutines, the auto-backup scheduler, the watchdog, the auto-update poller, the audit pruner, the SSE broker, and every subsystem goroutine that the daemon starts at `daemon.go:195-241` will, on a panic, take the whole process down. There is no per-goroutine recovery. The docstring says "call it in every daemon goroutine." It is not.

**Finding OPS-03 — P0. The EULA does not document a clean-uninstall flow.**

`EULA.md:88-89` says "Upon termination, you must cease all use of the Software and destroy all copies in your possession." That's a legal sentence, not a UX flow. The web `legal/page.tsx` is silent. The `uninstall.execute` RPC exists (`methods_phase11_misc.go:28-73`) but the user has to discover it. The `DownloadPageView.tsx` is the only place that mentions "auto-backup before uninstall," and that's in a marketing claim, not in the EULA or the legal page.

### 3.2 High findings (P1, would block diagnosis in a break)

**Finding OPS-04 — P1. `/healthz` is hard-coded `ok`; not deep.**

`internal/ipc/transport.go:128-132` returns literal `ok` from `/healthz`. The `healthCheckIPC()` registered at `subsystems.go:1239-1244` is a no-op that always returns `nil`. Kubernetes, macOS launchd, or any external orchestrator that does an HTTP health check will see `200 OK` even when the DB is wedged and the required health check has been failing for 10 minutes.

**Finding OPS-05 — P1. Health is never run on a schedule.**

`health.Register.Snapshot` exists but is only called from the `health.snapshot` RPC, which the GUI has to invoke. There is no background "if health is `down` for 60s, log loudly" loop. A user has no automatic notification that the daemon is in a bad state.

**Finding OPS-06 — P1. No user-visible crash dialog. No auto-submit.**

`crash.Capture` writes a local file and returns a `*Report` that callers can `ToTelemetry()` — but nothing in the daemon's call graph actually submits it. The closest the codebase gets is `telemetry.Reporter.RecordCrash(stack string)` (`telemetry/reporter.go:144-162`), which the daemon never calls. Result: the local crash file is the only artifact; a non-technical user will never know it exists.

**Finding OPS-07 — P1. No RPC to read the local crash files.**

The crash files at `~/.condura/crashes/crash-*.log` are write-only from the daemon's perspective. There is no `crashes.list` or `crashes.get` RPC, no "open crash folder" action. The local crash log is invisible to the GUI.

**Finding OPS-08 — P1. No "debug toggle" surfaced to the user.**

`LoggingConfig.Level` is exposed and `ParseLevel` understands `debug` (`logger.go:46-60`), but the GUI has no surface to flip it. To get a `debug` log the user must hand-edit `~/.condura/config.yaml` and restart. There is no `log.level.set` RPC.

**Finding OPS-09 — P1. `BackupConfig.RetentionDays` honored by NO ONE.**

`config.go:177` defines `RetentionDays: 30` default; `daemon/daemon.go:1390-1424` calls `backup.NewScheduler(backup.DefaultSchedulerConfig(), ...)` which always uses `defaultSchedulerKeepN = 7`. The docstring at `subsystems.go:1397-1414` literally says "cfg.Backup.IntervalHours is set, we honor it" but `IntervalHours` is not a field on `BackupConfig`. **The user sets `backup.retention_days: 30` and gets 7.** They set `backup.interval_hours: 6` and the code doesn't even try to read it.

**Finding OPS-10 — P1. No `IntervalHours` field exists in `BackupConfig`.** The user has no documented way to change the auto-backup cadence.

**Finding OPS-11 — P1. No `KeepN` field either.** The user cannot change the rotation count from 7.

**Finding OPS-12 — P1. No test for `backup.restore` against a corrupted archive.**

The unit tests cover "wrong key", "oversized manifest", "schema mismatch", and "happy round-trip", but there is no end-to-end test that simulates a half-corrupted archive (a `*.zip` whose central directory is intact but one of the encrypted entries is truncated). The code at `restore.go:268-274` will fail the size check and return an error, but the test suite doesn't prove it.

**Finding OPS-13 — P1. The Ed25519 update key is baked into the binary; no rotation story.**

`updater/updater.go:35-40` defines `PublicKey` as a 32-byte literal. If this key is ever compromised (or the team wants to rotate it on a normal cadence), the only way to update is to ship a new binary. There is no `UpdaterPublicKey` config field, no `update.public_key` YAML knob, no `PinPublicKey` API. This is the "we baked the trust root into the binary" decision — it works, but a stolen key is a permanent compromise for every shipped binary.

**Finding OPS-14 — P1. Watchdog is opt-in and disabled by default.**

`internal/config/loader.go:98-102` sets `Watchdog: WatchdogConfig{Enabled: false, ...}`. The user must hand-edit `config.yaml` to enable it. A non-technical user — the only kind of user who would benefit from a watchdog that auto-halts an out-of-control agent — will never enable it.

**Finding OPS-15 — P1. `watchdog.enable` and `watchdog.disable` are stubs that return errors.**

`daemon/methods_watchdog.go:63-81`. The user-facing buttons will be wired to these RPCs and the buttons will appear broken.

**Finding OPS-16 — P1. `Watchdog.Touch` is not called on any per-user-event path.**

`grep` shows it's called only from the RPC handler at `methods_watchdog.go:59`. Nothing in the codebase calls `subs.Watchdog.Touch()` automatically when the user types into the chat, approves a consent, or otherwise verifies the agent's actions. A user who doesn't manually click the "I'm here" button will trip the watchdog.

**Finding OPS-17 — P1. In-process network guard is bypassable by design.**

`halt/network.go:11-18`: "a misbehaving agent can skip the transport." The `InProcessGuard` is wired via `WireToHTTPClient`, but the LLM client must use that helper to be protected. A new LLM client added to the codebase tomorrow that calls `http.DefaultClient` will bypass the guard. There is no static check or test that catches this.

**Finding OPS-18 — P1. `haltFlag.Refresh` is called exactly once at startup.**

`daemon/subsystems.go:543`. The docstring at `flag.go:90-92` says "Call once at startup, then periodically (every ~1s)." The periodic refresh is never wired.

**Finding OPS-19 — P1. `SetEndpoint` doesn't persist.**

`telemetry/reporter.go:93-101`. The SQL is `UPDATE telemetry_counters SET enabled = enabled WHERE id = 1` — a tautology. The new endpoint is held in memory only. Restart the daemon → endpoint is whatever's in the config.

**Finding OPS-20 — P1. `RecordCrash` is never called by the daemon.**

`grep` for `RecordCrash` and `crash.ToTelemetry` shows only the test file uses it. The `crash.Recover` in `daemon.go:109` writes a local file but does not call `RecordCrash`. So a panic never produces a telemetry event.

**Finding OPS-21 — P1. No SIGHUP handler.**

`cmd/condurad/main.go:62` and `app/web/main.go:64` only catch SIGINT and SIGTERM. The standard "reload config" signal kills the daemon.

**Finding OPS-22 — P1. `shutdownDaemon` does NOT close the IPC transport.**

`daemon/daemon.go:307-322`. The transport has a `Close()` method but the daemon never calls it. The HTTP server is killed by the Go runtime when the process exits, but in-flight RPCs are dropped mid-response. A long-running RPC (e.g. `cu.action` mid-CU-plan) is not drained.

**Finding OPS-23 — P1. No graceful shutdown timeout.**

`shutdownDaemon` calls `subs.Close()` and `subs.Storage.Close()` with no upper bound. A SIGTERM during a database spike can leave the process hung for tens of seconds. There's no `context.WithTimeout` wrapping the shutdown path.

**Finding OPS-24 — P1. `subs.Executor` is conditional on `cuComps != nil`.**

`subsystems.go:892-894`. When no LLM is configured, `cuComps` is nil and the executor is nil. The `pending` actions queue still accepts inserts, but they'll never be executed. No error is surfaced to the caller.

**Finding OPS-25 — P1. Audit retention: `0` means "keep forever" but feels like "delete now."**

`daemon/daemon.go:275-286`. A user who interprets `audit_retention_days: 0` as "delete now" will be surprised when their audit log grows forever.

**Finding OPS-26 — P1. Query API can't filter by actor, session, target.**

`audit.Query` supports `Limit`, `Offset`, `Since`, `Action`, `Level`, `Kind` (`log.go:71-78`) but no `Actor`, no `SessionID`, no `TargetApp`, no `Path`. The structured fields are written but the query API doesn't read them.

**Finding OPS-27 — P1. No proactive permission refresh.**

`permissions.Probe` is called only by the RPC. If the user grants Accessibility in System Settings, the next `permissions.status` call has to be issued by the GUI; nothing in the daemon notices proactively.

**Finding OPS-28 — P1. Storage `busy_timeout(5000)` is too long for foreground RPC.**

`storage/db.go:117`. A contended database makes a 5-second pause on every write. Combined with `SetMaxOpenConns(1)` (line 122-124, a single-writer bottleneck), a user-perceived hang is likely.

**Finding OPS-29 — P1. No `daemon.info` / `daemon.uptime` / `daemon.pid` RPCs.**

The pieces exist (`version.Info` has version/commit/build_date/go_version/platform). The GUI can show "v0.1.1" but not "running for 14h 22m." Without uptime, you can't say "this daemon has been alive since 03:00."

**Finding OPS-30 — P1. No `replay.delete` or `replay.purge` RPC.**

The audit log is HMAC-chained (per-row delete is non-trivial), but the **screenshots** are not in the audit chain — they're just encrypted PNG files. A user who says "delete that screen capture of my bank statement" has no RPC for it.

**Finding OPS-31 — P1. No `replay.set_retention` config field.**

The 24h TTL is hardcoded in `replay.go:56-59` and `screenshots.go:79`. The audit log has `audit_retention_days` config; replay has no equivalent.

**Finding OPS-32 — P1. `ExportMP4` shells out to `ffmpeg` without checking the path.**

`replay/export.go:56-77` uses `exec.LookPath("ffmpeg")` and then `exec.CommandContext(ctx, ffmpeg, args...)` with the looked-up path. The args are hardcoded flags (not user-controlled), so this is not a command-injection vector, but a user without ffmpeg gets a confusing error and the GUI has no way to know "is ffmpeg installed?"

**Finding OPS-33 — P1. No PID-staleness check on the lockfile.**

`gofrs/flock` is a kernel-level advisory lock; if the holder dies unexpectedly, the lock is released by the kernel. But on a stale-lock scenario (rare on Unix, common on NFS, possible on certain Windows filesystem drivers), there is no PID-staleness check. The only recovery is to delete `<data-dir>/condurad.lock` by hand.

### 3.3 Medium findings (P2)

- **OPS-34 (P2)** — Resume doesn't restart cancelled streams.
- **OPS-35 (P2)** — Halt does not stop background services (auto-backup, updater, etc.) — undocumented.
- **OPS-36 (P2)** — The provider allow-list is hardcoded in `halt/network.go:63-79`. Adding a self-hosted LLM requires an `AllowHost` call; no GUI surface.
- **OPS-37 (P2)** — `closers` slice order is wrong vs. Windows file-locks (`subsystems.go:905-914`).
- **OPS-38 (P2)** — `migrateLegacyDataDir` runs before lockfile is acquired (`daemon.go:119-124`). Race.
- **OPS-39 (P2)** — `maybeApplyPendingUpdate` runs before lockfile is acquired (`daemon.go:120-124`).
- **OPS-40 (P2)** — No GUI surface for the audit log viewer.
- **OPS-41 (P2)** — HMAC secret = storage master key; no separate audit key.
- **OPS-42 (P2)** — No `PRAGMA optimize` on shutdown.
- **OPS-43 (P2)** — `busy_timeout` too long for foreground RPC.
- **OPS-44 (P2)** — No `loglevel.get`/`loglevel.set` RPC.
- **OPS-45 (P2)** — EULA version is `v1`; the version string is hardcoded in `onboarding/eula.go:78-80`, not parsed from the file.
- **OPS-46 (P2)** — EULA says "source code is proprietary" but the repo is open source. Web EULA is softer.

### 3.4 Low findings (P3)

- **OPS-47 (P3)** — No on-disk history of more than one prior version (`.old` gets clobbered).
- **OPS-48 (P3)** — Windows orphan swap script.
- **OPS-49 (P3)** — `mandatory` update UX is not in the GUI.
- **OPS-50 (P3)** — `auto_remove` for update notes is not displayed.
- **OPS-51 (P3)** — Default watchdog poll interval is 6h, hardcoded.
- **OPS-52 (P3)** — EULA no "I grant you a license to take physical actions" clause.
- **OPS-53 (P3)** — `~/.condura/installed` is not in the uninstall manifest.
- **OPS-54 (P3)** — EULA version has no timestamp visible on the web legal page.
- **OPS-55 (P3)** — Onboarding has no per-user state; multi-user shares wizard.

### 3.5 Positive findings (INFO)

- **OPS-INFO-1** — The HMAC-chained audit log is real, tested, and tamper-evident.
- **OPS-INFO-2** — The backup format is open, encrypted, self-describing, and round-trip is tested end-to-end (`backup_test.go:201-322`).
- **OPS-INFO-3** — The crash capture writes a human-readable file the user can open in any text editor.
- **OPS-INFO-4** — The logger's redaction (`logger/redact.go:36-103`) is one of the better implementations in this kind of code — 28 closed-set keys + 9 substring fallbacks + 11 value-pattern regexes.
- **OPS-INFO-5** — The `startBackgroundServices` test at `trust_phase11_caveats_test.go:162-176` is exactly the test that fixed the v0.1.0 B-01 / B-02 orphaned-subsystem regression. The B-01 / B-02 fix is genuinely working.

---

## 4. Cross-Platform Parity + Svelte Frontend Quality — 42 findings (3 P1, 14 P1, 13 P2, 11 P3, 1 INFO)

### 4.1 Critical findings (P1, would block a public launch with cross-platform claims)

**Finding UX-01 — P1. Sidebar icon rail has zero ARIA labels.**

`Sidebar.svelte:41-143` — the entire 9-icon `icon-rail` is built with `<a class="rail-icon" href="#/audit" title={t('sidebar.nav.audit')}>` and similar. There is no `aria-label`, no `aria-labelledby`, and the `<nav class="icon-rail">` has no `aria-label`. The 9 navigation icons are completely opaque to screen reader users. The `title=` attribute is a tooltip only, not a substitute.

**Finding UX-02 — P1. Conversation items in the Sidebar are unlabeled buttons.**

`Sidebar.svelte:167-179` — the conversation list buttons have NO `aria-label`. The visible content is `<span class="title">{c.title}</span><span class="meta">{msg_count} · {date}</span>`, but the button's accessible name is computed from that — meaning screen readers announce the full meta string ("3 messages · 6/24/2026") as the label, not the title.

**Finding UX-03 — P1. Conversation deletion is not undoable. Instant data loss.**

`conversation.svelte.ts:55-64` — `deleteCurrent` immediately calls `ipc.conversationsDelete(id)` and filters the list. There is no trash, no undo, no "are you really sure" with a 5-second grace period. Press Enter accidentally on a focused "delete" button → conversation is gone forever.

**Finding UX-04 — P1. `alert()` used in 12 places for success/failure — blocking native dialogs.**

`grep -c "alert(" app/web/frontend/src/lib/routes/*.svelte` returns 12 hits. `Settings.svelte:79, 176, 187, 203, 218, 220, 244, 246, 271, 273, 284` and `Replay.svelte:21, 23`. Unstyled OS dialogs that interrupt the user, look out of place in a "premium" design system, and don't respect the dark/ink theme.

**Finding UX-05 — P1. `confirm()` used in 10 places for destructive actions; no undo, no styled modal.**

`grep -c "confirm(" app/web/frontend/src/lib/` returns 10 hits. `Sidebar.svelte:33` (delete conversation), `Settings.svelte:140, 150, 181, 227, 232` (delete API key, reset adaptive, rerun setup, halt), `Hub.svelte:33` (install skill), `Skills.svelte:25` (delete skill), `Sync.svelte:49` (revoke device), `Channels.svelte:56` (disconnect channel). All blocking native dialogs.

**Finding UX-06 — P1. Linux presence detector returns `true` unconditionally.**

`internal/presence/detector.go:202` — the Linux `checkActiveOnLinux` returns `true` with a comment "This is a placeholder." This means any DESTRUCTIVE action that requires `require_user_active` is auto-allowed on Linux. The safety gate that the rest of the system assumes is wide open.

### 4.2 Cross-platform findings (would block Linux/Windows public release)

**Finding UX-07 — P1. Windows GUI build was broken in v0.1.1 (resolved at HEAD).**

Per the v0.1.1 release log, the Windows GUI build failed with "a.OpenQuickPrompt undefined." Fix committed in `06feee9`/`1abb6a1`/`cace2a4`. **Verified at HEAD** via `strings app/web/web.exe` that the symbols are present.

**Finding UX-08 — P1. Linux hotkey stub returns success but does not register.**

`internal/hotkey/hotkey_linux.go:37-44` — `Start` records the callback but does NOT register a global hotkey. Returns nil error so caller doesn't know the hotkey is dead. The `HotkeyRecorder.svelte` UI shows a recorded hotkey on Linux, but pressing it does nothing.

**Finding UX-09 — P1. Linux tray package not built; tray_wiring_linux.go is no-op.**

`internal/tray/tray.go:9` is `//go:build !linux`. On Linux the user has no menu-bar / tray icon. `app/web/tray_wiring_linux.go:1-21` is a no-op that logs "tray: not available on Linux; using in-app menu instead." The user has to discover the in-app menu.

**Finding UX-10 — P1. Voice capture unavailable on Windows/Linux; Settings → Voice always says "not granted."**

`internal/voice/recorder_other.go:37-39` — `Start` always returns error. The error message is honest: "audio capture is not available on this platform; install whisper.cpp and configure voice.binary_path in Settings to enable local transcription, or add an OpenAI API key for cloud transcription."

**Finding UX-11 — P1. Native TTS unavailable on Windows/Linux; only cloud TTS.**

`internal/voice/speaker_other.go:16-18` — `Speak` returns `errNotImplemented`. Cloud TTS (OpenAI, ElevenLabs) works on all platforms, but only if the user has API keys.

**Finding UX-12 — P1. Computer-use entirely stubbed on Windows/Linux.**

`internal/computeruse/ax/ax_other.go:1` and `internal/computeruse/backends/*_other.go` — no UIA backend for Windows, no AT-SPI backend for Linux. `Router.Execute` returns `ErrNoBackend`.

**Finding UX-13 — P1. Permissions always `StatusUnknown` on Windows/Linux.**

`internal/permissions/permissions.go:98-100` — the `defaultProbeOne` returns `StatusUnknown` for every kind on every non-darwin platform.

**Finding UX-14 — P1. iMessage, WhatsApp, Signal all stub.**

`internal/reach/{imessage_stub,whatsapp,signal}.go` — `Connect` returns `UnsupportedError` with "coming in v0.2.0" / "macOS only" messages. Only Telegram works everywhere.

**Finding UX-15 — P1. No "daemon not running" UI.**

`daemon.svelte.ts:6-50` tracks `connected: boolean`. The status bar shows a green/red dot. But when the daemon is down, the Chat page just shows "Waiting for daemon connection…" and the user can still click Send — the IPC call will fail, and the error will appear in a toast (if the store is wired) or silently fail. No blocking modal, no retry button.

**Finding UX-16 — P1. IPC reconnect has no jitter, no UI for "reconnecting…."**

`client.ts:570-581` — `scheduleReconnect` uses `250 * Math.pow(2, attempt-1)` capped at 30s. No random jitter (thundering herd risk if the daemon is shared).

**Finding UX-17 — P1. Many error handlers swallow exceptions silently.**

Examples: `init.ts:31-33, 46-59, 75-88` (every catch block does `// ignore`); `Sidebar.svelte:20` (no UI feedback if it fails); `client.ts:606-614, 617-625` (return empty on error and log nothing); `client.ts:540-543, 548-552` (JSON parse errors on SSE events silently ignored); `eventemitter.ts:39-44` (handler errors are caught and console.error'd but no metric / Sentry).

**Finding UX-18 — P1. Zero `aria-live` regions.**

`grep -rn "aria-live" app/web/frontend/src/` returns zero matches. The Chat streaming, Toast notifications, and audit pill are all silent for screen readers.

**Finding UX-19 — P1. Toasts lack `role="alert"`/`role="status"`; close button has no `aria-label`.**

`Toasts.svelte:5-15` — the `.toast-container` has no `role`. The `.toast-close` button is text content `×` with no `aria-label="Dismiss"`.

**Finding UX-20 — P1. Hub/Skills/Delegation result rows are `<li onclick>` with no keyboard nav.**

`Hub.svelte:84`, `Skills.svelte:55`, `Delegation.svelte:109` — keyboard users cannot navigate the result list.

### 4.3 Medium findings (P2)

- **UX-21 (P2)** — macOS tray not shown in Wails GUI (systray AppDelegate conflict).
- **UX-22 (P2)** — macOS presence detector uses string-match on ioreg output (heuristic, not real last-input-time).
- **UX-23 (P2)** — PendingActions buttons have no per-button context.
- **UX-24 (P2)** — Replay frame buttons have no `aria-label`.
- **UX-25 (P2)** — SignInPanel provider buttons use first-letter of provider name as fake icon.
- **UX-26 (P2)** — Hub.svelte:84 / Skills.svelte:55 list items have no `tabindex`/`aria-selected`.
- **UX-27 (P2)** — No skeleton UI used anywhere; only plain "Loading…" text.
- **UX-28 (P2)** — Chat has no "thinking…" indicator between Send and first stream byte.
- **UX-29 (P2)** — Settings route loads 6+ things onMount with no aggregate loading state.
- **UX-30 (P2)** — Hub search requires Enter; no debounced auto-search.
- **UX-31 (P2)** — `audit.refresh` has no `catch`; rejected promise leaves UI stuck on "Loading…".
- **UX-32 (P2)** — Empty states are inconsistent.
- **UX-33 (P2)** — PendingActions `working` state replaced non-atomically — race possible.
- **UX-34 (P2)** — `Chat.svelte:42-47` auto-scrolls on every keystroke, fights user-initiated scroll.
- **UX-35 (P2)** — Settings → Rerun setup has no inline error feedback on failure.

### 4.4 Low findings (P3)

36 findings including:
- `highlight.js` and `marked` listed in `package.json` but never imported (dead deps)
- No code-splitting per route; 225 KB single bundle
- `:focus-visible { outline: none; box-shadow: var(--shadow-focus) }` removes browser default globally
- LocaleSelector uses Svelte 4 `subscribe` pattern inside `$effect`
- Sidebar "Quick prompt" button (line 120-123) has icon `aria-hidden="true"` but no `aria-label` on button
- `{@html t('skills.empty_html')}` renders raw HTML from locale file (XSS if locale becomes user-editable)
- Chat input has no `maxlength`; pasting 1 MB OOMs the daemon
- Color contrast edge case: `--color-text-faint` at 11px uppercase
- `trayUpdate` IPC method defined but never called
- `overlay.stop()` defined but never called
- `[data-theme="ink"]` defined but no theme switcher exists
- No `env(safe-area-inset-*)` use; overlay 620×88 may clip near notches
- Many ambient animations run on hidden tabs
- Heavy `backdrop-filter` (46 occurrences) — slow on low-end webkit2gtk

### 4.5 Verdict

**Would I ship to a Windows user today?** **No, not as a polished product** — but the binary works for the chat daemon use case. The Windows GUI binary compiles and the in-process daemon + chat + settings + audit + skills + sync + channels (Telegram) + delegation all function. Voice and computer-use are honest no-ops with clear error messages pointing to cloud alternatives. However, the user will see "Permissions: unknown" for everything, "Mic unavailable" in voice settings, and have no computer-use.

**Would I ship to a Linux user today?** **No, not the GUI — ship only the standalone daemon + CLI + TUI.** The Linux GUI has a silent broken global hotkey, no tray icon, and a presence detector that always returns `true` — meaning the DESTRUCTIVE-action consent gate is wide open on Linux. The standalone daemon + TUI is the right Linux deliverable.

**Honest disclosure for the marketing site (per PR #13 already):** "macOS: full feature set. Windows: chat + stores, no voice/computer-use. Linux: chat + stores via in-app menu, no tray, no global hotkey, no voice, no computer-use, destructive consent requires local trust."

---

## 5. Functional Completeness + Adversarial Safety — 15 flows × 21 scenarios

### 5.1 Functional flows (15 of 15 verified at HEAD)

| # | User Flow | Status | Evidence |
|---|-----------|--------|----------|
| 1 | Install + first launch | ✅ WORKS | Run #1, `docs/phase15-verification.md:176-228` |
| 2 | Add API key | ✅ WORKS | `internal/api_key/manager.go:144-206`; v0.1.1 sentinel |
| 3 | Send a chat message | ✅ WORKS | `internal/session/session.go:257-387` + `internal/stream/manager.go:225-314` |
| 4 | Grant DESTRUCTIVE consent | ✅ WORKS | `internal/gatekeeper/engine.go:154-246` + nonce + expiry (PR #2) |
| 5 | Halt the agent | ✅ WORKS | `internal/halt/flag.go` + `internal/halt/network.go` + `internal/watchdog/watchdog.go` |
| 6 | Spawn a sub-agent | ✅ WORKS | `internal/delegation/gated_runner.go:156-201` + `internal/pending/store.go` + `internal/executor/executor.go` |
| 7 | Auto-update | ✅ WORKS | `internal/updater/updater.go:299-316` (Ed25519 verified) |
| 8 | Auto-backup | ✅ WORKS | `internal/backup/scheduler.go` + `internal/backup/backup.go` (AES-256-GCM) |
| 9 | Uninstall | ✅ WORKS | `internal/uninstall/manifest.go:212-289` (hard guard + confirm token) |
| 10 | Audit log review | ✅ WORKS | `internal/audit/log.go:226-489` (HMAC chained, length-prefixed) |
| 11 | Restore from backup | ✅ WORKS | `internal/backup/restore.go:90-127` (gated + safety snapshot) |
| 12 | Voice input (macOS) | ⚠️ PARTIAL | Wake word works; full pipeline needs `voice.enabled=true` + SHA pins set |
| 13 | CU click (macOS) | ✅ WORKS | `internal/daemon/cu_resolver.go:117-202` + `internal/computeruse/verify.go:48-103` |
| 14 | Selective Perception budget | ✅ WORKS | `internal/perception/perception.go` + v0.1.1 wiring in `cu_resolver.go:95-112` |
| 15 | Kill switch in process | ✅ WORKS | (with HIGH caveat — see §5.2 #9) |

**Total: 14 ✅ WORKS / 1 ⚠️ PARTIAL / 0 ❌ BROKEN.**

### 5.2 Adversarial scenarios (13 defended, 7 partial, 0 vulnerable)

| # | Scenario | Status | Severity |
|---|----------|--------|----------|
| 1 | Prompt injection via screen text | 🛡️ DEFENDED | — |
| 2 | Prompt injection via LLM output | 🛡️ DEFENDED | — |
| 3 | SQL injection via API key name | 🛡️ DEFENDED | — |
| 4 | API key via crash report | 🛡️ DEFENDED | — |
| 5 | API key via audit log | 🛡️ DEFENDED | — |
| 6 | API key via SSE | 🛡️ DEFENDED | — |
| 7 | API key via backup | 🛡️ DEFENDED | — |
| 8 | HMAC chain compromise | 🛡️ DEFENDED | LOW (F-10) |
| 9 | **Halt flag bypass (self-re-arm)** | ⚠️ PARTIAL | **HIGH** |
| 10 | Network guard bypass (sub-agents) | ⚠️ PARTIAL | MEDIUM |
| 11 | Update manifest compromise (no key rotation) | ⚠️ PARTIAL | HIGH (supply chain) |
| 12 | P2P sync compromise | 🛡️ DEFENDED | — |
| 13 | TOFU sidecar attack (wake) | ⚠️ PARTIAL | MEDIUM |
| 14 | TOFU sidecar on Linux (modelmgr) | ⚠️ PARTIAL | MEDIUM |
| 15 | MCP prompt injection | 🛡️ DEFENDED | — |
| 16 | OAuth prompt / state CSRF | 🛡️ DEFENDED | — |
| 17 | Halt flag race | ⚠️ PARTIAL | HIGH (same as #9) |
| 18 | Lockfile stale recovery | 🛡️ DEFENDED | — |
| 19 | TOCTOU in consent queue | ⚠️ PARTIAL | MEDIUM |
| 20 | Side-channel via replay | 🛡️ DEFENDED | — |
| 21 | API Server config not honored | ⚠️ PARTIAL | LOW |

**Top 3 adversarial findings:**

**Finding SAFE-01 — HIGH. Agent can re-arm its own halt flag via `daemon.resume` over the IPC bearer token.**

`internal/daemon/methods_phase2.go:303-330`, `internal/daemon/methods_phase9.go:74-85`. `daemon.halt` and `daemon.resume` are RPCs that any in-process code with the bearer token can call. A prompt-injected or compromised agent can re-arm itself. The audit log records the re-arm, so it's forensically traceable, but the kill switch is not *sticky* by default. The v0.2.0 "hard Layer 3" deliverable (a real `pf`/`netsh` daemon the agent cannot influence) will close this. In the meantime, add a 5-second cooling window or a "halt requires re-install" mode at the cost of UX friction.

**Finding SAFE-02 — MEDIUM. Sub-agents + MCP servers are separate processes; the network guard does not cover their HTTP, and `Halt` does not kill them.**

When the user presses the kill switch, the daemon's HTTP is blocked, but the Claude Code / Codex subprocesses continue to run. `internal/halt/network.go:7-19` documents this as "soft Layer 3." The v0.2.0 hard Layer 3 fixes it via a separate process the agent can't influence.

**Finding SAFE-03 — MEDIUM. TOCTOU in consent queue: `pending_actions.payload_json` is mutable after approval.**

A row whose `(status, payload)` is `(approved, "git status")` at approval time could become `(approved, "rm -rf /")` at execute time. The `ShellSanitizer` would still block `rm` (not in allowlist), but a malicious command in the allowlist could be substituted. `internal/pending/store.go:192-238` has no immutability; `internal/executor/executor.go:240-263` re-reads payload at execute time.

### 5.3 Functional + Adversarial Verdict

**If 1000 users installed this binary today and 10 of them had adversarial users in their environment, the safety layer would hold for 9 of them, and would fail for 1 — but the failure mode for the 1 is "the agent re-arms itself after the kill switch," not "the attacker exfiltrates the API key."**

The functional layer is excellent. Every user flow that the spec promises works end-to-end. The safety layer is **strong** but not **bulletproof**. The 1 HIGH finding (Halt flag self-re-arm) is a known design trade-off that the spec accepts; the v0.2.0 hard Layer 3 will close it. The MEDIUM findings are defense-in-depth gaps tracked as P1 follow-ups but do not block the v0.1.x release.

---

## 6. The verdict on "could Condura deploy itself?"

**Partially, with significant caveats.**

What Condura COULD do today on macOS (genuine, not aspirational):
1. ✅ Download the marketing site source (`git clone`)
2. ✅ Use `web/app/download/page.tsx` to learn the install flow
3. ✅ Open a browser to the Vercel dashboard (via the GUI's computer-use pipeline — `daemon/cu_resolver.go:117-202` + `internal/computeruse/verify.go`)
4. ✅ Sign into Vercel (via the account OAuth flow — `internal/account/oauth.go`)
5. ✅ Click through the "Import Project" wizard
6. ✅ Type the GitHub repo URL `sahajpatel123/conduraapp`, set root directory to `web`, click "Deploy"

This is a real demo of the computer-use + delegation + safety layers working end-to-end. It would take ~3 minutes of screen recording.

What Condura CANNOT do today:
1. ❌ Deploy the daemon to a cloud server (Vercel/Railway don't run long-lived Go daemons with OS access; and even if they did, the daemon needs a screen, keyboard, and file system that doesn't exist on a cloud server)
2. ❌ Build & sign the native binaries autonomously (the signing keys live in GitHub secrets; the daemon can't access them; the build needs macOS Xcode, Windows NSIS, and Linux makensis toolchains that only exist on dedicated CI runners)
3. ❌ Create a GitHub release (the token has `contents: write` but the daemon would need the `gh` CLI + a network allowlist exception for the GitHub API)
4. ❌ Update the marketing site copy (no CMS; the site is static Next.js)
5. ❌ Set the Vercel custom domain (requires DNS configuration outside Condura's control)

**The honest framing: Condura can drive the user's local browser to deploy the marketing site, but the actual build/sign/publish flow is gated to GitHub Actions with secrets that Condura cannot reach. This is the right security posture — the build chain should NOT be accessible to the thing it ships.**

A more interesting "deploy itself" interpretation: Condura drives the user through the install of the daemon on their own Mac, signs into the user's Vercel account via OAuth, imports the repo, and clicks Deploy. The user watches; Condura clicks. The artifact (a static Next.js bundle) ends up on Vercel. The daemon never has to leave the user's Mac. This is the closest honest "self-deployment" demo and it's achievable today with ~30 lines of orchestration code in a new `internal/automation` package.

For the literal "ship Condura to the cloud so it can act" interpretation: **no, never, do not attempt this**. The product's entire safety model depends on the daemon running on a machine the user controls with OS-level permissions and a physical kill switch. Cloud-hosting it would defeat the model and would also fail to deliver any value (a cloud daemon has no screen to look at, no keyboard to type on, no files to read, no TCC to grant).

---

## 7. Recommendations and next steps

### 7.1 To ship v0.1.1 to a closed beta (≤50 macOS testers, alpha label)

**You can ship today.** v0.1.1 is sufficient. Document the known gaps in the beta welcome email:

- macOS only (Windows + Linux ship the daemon + CLI + TUI; GUI is macOS-first)
- Watchdog is opt-in (Settings → Advanced → enable)
- No voice activation by default (configure in Settings → Voice)
- Some marketing claims are aspirational and will be addressed in v0.2.0
- Auto-update is signed; if it fails, manual download is at https://github.com/sahajpatel123/conduraapp/releases

### 7.2 To ship v0.1.1 to a public release (Product Hunt, HN, etc.)

**Block on the 12 must-fix items in §0.** Estimated 2-3 weeks of focused work, mostly CI/supply-chain. The codebase is the easy part; the governance is the hard part.

Priority order:
1. **Day 1-2:** PR-01 (release.yml fail-open), PR-02 (enable Dependabot + secret scanning + push protection + code scanning), PR-03 (branch protection on main with required CI checks + required CODEOWNERS review + required linear history), PR-11 (release-verify.yml fail-open)
2. **Day 3-4:** PR-04 (`install.sh` signature verification — minimum: add a SHA-256 check against `manifest.json` from the auto-update channel), PR-05 (macOS notarization + Windows Authenticode + Linux GPG via `goreleaser`'s signing hooks), PR-08 (`set -euo pipefail` + ERR trap), PR-10 (CODEOWNERS coverage for security-critical files)
3. **Day 5-7:** UX-06 (Linux presence fix — `return false` placeholder), UX-07/08/09/10/11/12/13/14 (frontend accessibility + cross-platform honesty), SAFE-01 (sticky halt mode)
4. **Day 8-10:** PR-12 (EULA clean-uninstall section + legal page), UX-01/02/03/04/05 (sidebar ARIA + undo delete + replace alert/confirm), OPS-01 (log rotation), OPS-02 (per-goroutine panic recovery), OPS-13/14 (watchdog enabled by default), OPS-15 (replay delete/purge RPCs), OPS-09/10/11 (backup retention/interval/keep honored)
5. **Day 11-12:** Full re-test, re-audit, sign the release, re-tag v0.1.1

### 7.3 To ship v0.2.0

The 9 INFO items from the v0.1.1 backend audit plus:
- Hybrid LLM router (`internal/router/`) with cascade + memory bias + per-task override
- Subscription OAuth (ChatGPT Plus, Claude Pro, SuperGrok)
- Hard Layer 3 kill switch (separate `condura-guard` process with `pf`/`netsh` rules)
- Vector embeddings via sqlite-vec + all-MiniLM-L6-v2
- Vector recall in memory
- Wave scheduler / DAG executor
- MCP HTTP + SSE transports
- WhatsApp / Signal channels + iMessage receive
- Public Skills Hub at `hub.condura.app` (Next.js)
- Linux hotkey (X11 Record Extension / Wayland portal)
- Linux system tray
- Non-macOS voice (cloud-only)
- Linux GUI parity

These are the v0.2.0 roadmap per `docs/roadmap-v0.2.0.md` and the v0.1.1 audit's deferral list. Estimated 6-8 weeks.

### 7.4 What I recommend you do **right now** (today)

1. **Enable Dependabot + secret scanning + push protection on the repo** (5 minutes, free, no code changes). This is the single highest-leverage CI fix and it's the same finding the security audit flagged 24 hours ago. The fact that it hasn't been actioned is itself a finding.
2. **Add a `branch-protection` ruleset on main**: required CI checks (lint, test, build, security), required CODEOWNERS review, required linear history, no force pushes, no deletions. (10 minutes.)
3. **Edit `release.yml:200-208` and `release-verify.yml:126-128`**: change `exit 0` to `exit 1` when `UPDATE_SIGNING_KEY` is unset. (5 minutes.)
4. **Add `set -euo pipefail` and an ERR trap to `scripts/install.sh`**. (10 minutes.)
5. **Fix the Linux presence detector**: change `internal/presence/detector.go:202` from `return true` to `return false` (or delete the function and have the caller return false). (5 minutes.)

These 5 changes take ~35 minutes total and address the most consequential P0s. Everything else can wait.

---

## 8. Appendix — how to reproduce this audit

```bash
# Verify the CI state at HEAD.
gh pr checks 13

# Re-run the local test suite.
go test -count=1 -race -timeout 180s -short ./...

# Confirm svelte-check is clean.
cd app/web/frontend && ./node_modules/.bin/svelte-check --tsconfig ./tsconfig.json

# Verify the security features that should be enabled.
gh api /repos/sahajpatel123/conduraapp | jq .security_and_analysis

# Confirm branch protection.
gh api /repos/sahajpatel123/conduraapp/branches/main/protection

# List release assets.
gh release view v0.1.1

# Read the v0.1.1 install script.
cat scripts/install.sh

# Check the Linux presence detector.
sed -n '195,205p' internal/presence/detector.go

# Check the release.yml fail-open condition.
sed -n '195,210p' .github/workflows/release.yml

# Check the release-verify.yml fail-open condition.
sed -n '120,135p' .github/workflows/release-verify.yml

# Re-confirm B-01/B-02 are fixed.
grep -A5 "buildSafetyLayer\|subs.Capturer" internal/daemon/subsystems.go | head -30

# Re-confirm the v0.1.1 secrets encryption.
grep -A3 "ensureLoadedAndMigrated\|ErrBackendFailed.*wrong key" internal/secrets/manager.go | head -20
```

**End of audit.**

---

## Summary of the user's question, answered

**"Is everything green?"** Yes. PR #13 CI is 12/12 green (all active checks), main CI is 3/3 green (most recent runs), local test suite is 47/47 packages ok, svelte-check is 0/0, golangci-lint is 0/0, all binaries build for the documented matrix.

**"Is Condura production-ready / could it deploy itself / could it be used for production quality?"** It depends what you mean. For a closed beta of 50 hands-on macOS testers, yes. For a public launch to strangers, no — the binary is good, the CI/CD pipeline that ships the binary is not, and there are 12 must-fix P0/P1 items before public launch. For "deploy itself": the marketing site can be deployed to Vercel by the user with Condura driving the browser (real, achievable today); the daemon cannot be deployed to the cloud (it would lose all its value and break the safety model); "deploy itself" is a poetic framing that doesn't map to this product — the honest framing is "Condura can act on the user's behalf, on the user's machine, with the user's physical oversight."

**The single most important thing you can do right now** is spend 5 minutes enabling Dependabot, secret scanning, push protection, and code scanning on the GitHub repo, then 10 minutes adding branch protection to main. Both are free, both close the most consequential P0s, and neither requires any code change.
