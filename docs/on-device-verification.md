# On-Device Verification Checklist (v0.1.0)

This is the acceptance gate for **public** v0.1.0. Automated CI covers
packaging, checksums, and Ed25519 manifest signing; this checklist covers
OS-native behavior that only a real machine can prove.

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
- [ ] Notarization staple verification: `stapler validate /Applications/Synaptic.app` (when Apple secrets configured)
