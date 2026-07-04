# On-Device Verification Checklist (v0.1.0)

This is the acceptance gate for **public** v0.1.0. Automated CI covers
packaging, checksums, and Ed25519 manifest signing; this checklist covers
OS-native behavior that only a real machine can prove.

---

## Operator playbook (human action required)

**Who:** Product lead or QA operator with physical access to each target OS.
**Time:** ~4–6 hours per OS (macOS first; Windows and Linux after).
**Cannot be done by CI:** TCC permissions, Gatekeeper/SmartScreen, overlay
hotkey latency, consent modals, voice mic/TTS, and real computer-use clicks.

### Before you start

1. **Get a release build** — use the tagged v0.1.0 artifacts from GitHub
   Releases (`condura-gui-*.dmg`, `*-setup.exe`, `.deb`), not a local
   `wails dev` build, unless you are explicitly testing a pre-release candidate.
2. **Use a clean machine** (or a clean user account):
   - No prior Condura/Synaptic install
   - No `~/.synaptic/` directory
   - No Ollama, no API keys, no developer tools beyond a browser
   - macOS: prefer a real Mac (not a VM) for Accessibility + Screen Recording
3. **Prepare evidence folder** on the test machine, e.g.
   `~/Desktop/condura-v0.1.0-verify-<os>-<date>/` for screenshots,
   screen recordings, and log excerpts.
4. **Open both checklists:**
   - This file (`docs/on-device-verification.md`) — smoke-level acceptance
   - `docs/phase15-verification.md` — full Phase 15 matrix (required for ship sign-off)

### Execution order (per OS)

| Step | Action | Record |
|------|--------|--------|
| 1 | Run automated CI checks locally if verifying a new tag (see below) | PASS/FAIL + command output |
| 2 | Download from `https://condura.app/download` | Screenshot of download page |
| 3 | Verify checksum + code signature | Paste `sha256sum` / `codesign` output |
| 4 | Install and complete onboarding (EULA → permissions → hotkey → ready) | Screen recording |
| 5 | Chat: send one message with Ollama (if available) or API key you add in Settings | Screenshot of streaming reply |
| 6 | Computer use: trigger a WRITE action → approve consent modal | Screenshot of modal + result |
| 7 | Voice (macOS): speak a question → transcription + reply | Short screen recording |
| 8 | Safety: Cmd+Shift+Escape (or platform kill hotkey) → agent halts | Screenshot of halted tray state |
| 9 | Backup → restore → uninstall → reinstall | File paths + screenshots |
| 10 | Fill every **Status** cell in `docs/phase15-verification.md` | PASS / FAIL / N/A + Notes |
| 11 | Complete the **Sign-off** table at the bottom of phase15 | Names + dates |

### How to record results

- Mark each checkbox in this file and each table row in phase15.
- For any **FAIL**: stop the run, open a GitHub issue labeled `P0` or `P1`,
  attach logs from `~/.synaptic/logs/` (redact API keys), and do not sign off.
- **Ship-ready** means: all phase15 rows PASS on at least one clean machine
  per OS (macOS arm64, Windows 11 amd64, Ubuntu 22.04 amd64) and sign-off table complete.

### Quick log locations

| OS | Daemon / app logs |
|----|-------------------|
| macOS | `~/.synaptic/logs/`, Console.app → filter "condura" |
| Windows | `%USERPROFILE%\.synaptic\logs\` |
| Linux | `~/.synaptic/logs/`, `journalctl --user -u condura` if using systemd user unit |

### When you are done

1. Commit or attach the filled `docs/phase15-verification.md` (or store in your release evidence repo).
2. Update `LOGBOOK.md` with verdict: ship-ready or blocked + issue links.
3. Only then tag or promote the public release.

---

## Automated (CI — no manual step)

- [x] `go build ./...` and `golangci-lint` clean on `main`
- [x] `release-verify` — GoReleaser snapshot + manifest sign roundtrip
- [x] `embedded-key-check` — `UPDATE_SIGNING_KEY` matches embedded `PublicKey` (when secret set)
- [x] Updater unit tests + `update_e2e_test.go` through IPC
- [x] `scripts/verify-release-artifacts.sh v0.1.0` after tag (checksums + manifest sig)

Run locally after a tag:

```bash
make verify-release TAG=v0.1.0
go run ./cmd/gen-update-manifest verify dist/verify-v0.1.0/manifest.json
```

## macOS (primary target)

- [ ] Install `.dmg` — drag to Applications, no Gatekeeper block
- [ ] First run launches onboarding wizard
- [ ] Onboarding: grant Accessibility, Screen Recording, Microphone permissions
- [ ] Hotkey opens overlay (<100ms)
- [ ] Overlay: type a message, get a response via configured LLM
- [ ] Voice: speak a question, get transcription + TTS response
- [ ] Computer use: execute a real click action → consent modal appears → approve → click executes
- [ ] Delegation: `delegate.spawn claude "list files"` → consent → spawns
- [ ] Auto-update: signed manifest check → "update available" notification
- [ ] Backup: create backup from Settings → verify `.synaptic-backup` exists
- [ ] Restore: restore from backup → verify data intact
- [ ] Uninstall: drag to Trash + delete `~/.synaptic/` → clean state
- [ ] Re-install: no second-install message (clean install after removal)
- [ ] Second-install block: re-run installer without removing → "already installed" message
- [ ] Halt: press Cmd+Shift+Escape → agent stops, menu bar shows halted
- [ ] No telemetry: check `~/.synaptic/` — no crash reports sent (opt-in disabled)

## Windows

- [ ] Install `.exe` (NSIS) — no SmartScreen block
- [ ] Same checks as macOS (adapted for Windows equivalents)
- [ ] Uninstall from Control Panel

## Linux

- [ ] Install `.deb`
- [ ] Same checks
- [ ] SHA256: `sha256sum -c checksums.txt`

## Release Evidence

- [ ] Screenshots of consent modal during computer-use action
- [ ] Screenshots of auto-update notification
- [ ] Screenshots of backup/restore flow
- [ ] Log output from a full end-to-end session
- [ ] Notarization staple verification: `stapler validate /Applications/Condura.app` (when Apple secrets configured)
