# On-Device Verification Checklist (v0.1.0)

This is the acceptance gate. Run on a real, clean machine per target OS before
tagging v0.1.0. Automated tests cannot verify native AX, CGEvent, OS Gatekeeper,
or notarization — these must be verified manually.

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

- [ ] Install `.deb`/`.rpm`/`.AppImage`
- [ ] Same checks
- [ ] GPG signature verifies: `gpg --verify synaptic-0.1.0-linux-amd64.deb.sig`
- [ ] SHA256SUMS: `sha256sum -c SHA256SUMS`

## Release Evidence

- [ ] Screenshots of consent modal during computer-use action
- [ ] Screenshots of auto-update notification
- [ ] Screenshots of backup/restore flow
- [ ] Log output from a full end-to-end session
- [ ] Notarization staple verification: `stapler validate /Applications/Synaptic.app`
