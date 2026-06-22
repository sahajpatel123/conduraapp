# Phase 15 Verification Checklist

> Final pre-public-launch verification for Condura v0.1.0.
> Goal: prove that a real user can download, install, onboard, and use the product safely.
>
> Execute this checklist on **at least one clean machine per OS**:
> macOS (arm64), Windows 11 (amd64), Ubuntu 22.04 (amd64).
> A "clean machine" means no prior Condura install, no pre-existing `~/.synaptic/`,
> no Ollama, no API keys configured, and no developer tooling beyond a browser.

---

## How to Use This Checklist

1. Read **`docs/on-device-verification.md`** first — it has the operator
   playbook (clean machine setup, evidence folder, execution order).
2. Run each step in order on a clean machine.
3. For each step, record: **PASS**, **FAIL**, or **N/A**.
4. If a step fails, stop the run, file a P0/P1 issue, and re-run the checklist after the fix.
5. Attach logs, screenshots, or screen recordings for any failure.
6. A completed checklist with zero failures is required before declaring v0.1.0 ship-ready.
7. Complete the **Sign-off** table at the bottom of this document.

---

## 1. Download

| # | Step | Expected Result | Status | Notes |
|---|------|-----------------|--------|-------|
| 1.1 | Visit `https://condura.app/download` | Page loads, platform auto-detected, no console errors | | |
| 1.2 | Download the artifact for the current OS | Download completes; file size matches release manifest | | |
| 1.3 | Verify checksum (macOS/Linux: `sha256sum`; Windows: `CertUtil`) | Checksum matches `manifest.json` entry | | |
| 1.4 | Verify code signature | macOS: `codesign -dv --verbose=4`; Windows: signature present in file properties; Linux: GPG signature if provided | | |

---

## 2. Install

| # | Step | Expected Result | Status | Notes |
|---|------|-----------------|--------|-------|
| 2.1 | macOS: open `.dmg`, drag to Applications | App copies without error; quarantine dialog appears on first launch | | |
| 2.2 | Windows: run `-setup.exe` | Installer completes; shortcut created; no antivirus false positive | | |
| 2.3 | Linux: `chmod +x` binary or install `.deb` | Binary runs; dependencies resolved | | |
| 2.4 | Launch app for the first time | App opens; menu bar / tray icon appears | | |
| 2.5 | Confirm only one instance can run | Second launch shows "already running" or focuses first instance | | |

---

## 3. Onboarding

| # | Step | Expected Result | Status | Notes |
|---|------|-----------------|--------|-------|
| 3.1 | EULA screen appears | Scroll-to-bottom + checkbox required before Accept | | |
| 3.2 | Permissions screen | Shows Accessibility + Screen Recording status; deep links open System Settings | | |
| 3.3 | Grant Accessibility permission | `permissions.status` reports `granted` within 2 polling cycles | | |
| 3.4 | Grant Screen Recording permission | `permissions.status` reports `granted` within 2 polling cycles | | |
| 3.5 | Hotkey screen | Records a valid combo; Continue enabled only after record | | |
| 3.6 | Ready screen | Detects Ollama if present; otherwise shows API key / CLI options | | |
| 3.7 | Click "Start using Condura" | Wizard dismisses; main chat UI appears | | |
| 3.8 | Re-run setup from Settings | Wizard re-appears at EULA; does not lose existing config | | |

---

## 4. Chat

| # | Step | Expected Result | Status | Notes |
|---|------|-----------------|--------|-------|
| 4.1 | Type a simple message and send | Message appears in conversation; streaming response begins | | |
| 4.2 | With Ollama reachable | Response generated locally; no API cost | | |
| 4.3 | With API key configured | Response generated via configured provider; cost recorded | | |
| 4.4 | Cancel a streaming response | `llm.cancel` stops tokens; UI returns to idle | | |
| 4.5 | Start a new conversation | New thread created; previous history preserved | | |
| 4.6 | Switch conversation | Messages load correctly | | |
| 4.7 | Delete a conversation | Confirms before delete; conversation removed | | |

---

## 5. Computer Use

| # | Step | Expected Result | Status | Notes |
|---|------|-----------------|--------|-------|
| 5.1 | Ask agent to "open Finder and create a new folder" | Gatekeeper prompts for WRITE action | | |
| 5.2 | Approve the Gatekeeper prompt | Folder created; audit log shows `allow` | | |
| 5.3 | Deny a Gatekeeper prompt | Action blocked; audit log shows `deny` | | |
| 5.4 | Let consent timeout | Action queued (does not execute); UI shows timeout | | |
| 5.5 | Trigger a DESTRUCTIVE action (e.g., delete a file) | Native modal / consent required; cannot proceed without human | | |
| 5.6 | Verify twin-snapshot behavior | Rapid UI change before action causes abort/no-op | | |

---

## 6. Delegation

| # | Step | Expected Result | Status | Notes |
|---|------|-----------------|--------|-------|
| 6.1 | Ask agent to run a coding task | Agent delegates to installed CLI (Claude Code / Codex / etc.) | | |
| 6.2 | Approve delegation spawn | Sub-agent runs; output streamed back | | |
| 6.3 | Deny delegation spawn | Sub-agent does not run; audit log shows `deny` | | |
| 6.4 | Verify concurrency limits | Cannot exceed configured max parallel sub-agents | | |

---

## 7. Safety

| # | Step | Expected Result | Status | Notes |
|---|------|-----------------|--------|-------|
| 7.1 | Press kill-switch hotkey | Daemon halts; tray shows halted state | | |
| 7.2 | Resume from halt | Daemon resumes; chat works again | | |
| 7.3 | Open sensitive site (banking/health) in browser | Gatekeeper escalates to `RequirePresenceAndConsent` | | |
| 7.4 | Attempt rapid repeated action | Anomaly detector pauses/halt agent | | |
| 7.5 | Review audit log | All actions logged with actor, action, result, HMAC chain valid | | |
| 7.6 | Run `replay.verify_integrity` | Returns `valid: true` | | |

---

## 8. Voice

| # | Step | Expected Result | Status | Notes |
|---|------|-----------------|--------|-------|
| 8.1 | On macOS: enable wake word | "hey synaptic" wakes overlay | | |
| 8.2 | On macOS: press-and-hold hotkey | Voice orb appears; release submits transcript | | |
| 8.3 | On Windows/Linux: attempt voice | Meaningful error: "audio capture not available on this platform..." | | |
| 8.4 | Configure cloud transcription | OpenAI Whisper fallback works with API key | | |
| 8.5 | TTS response | Speaker reads response aloud (macOS native; cloud on other OSes) | | |

---

## 9. Memory, Skills, Sync, Account

| # | Step | Expected Result | Status | Notes |
|---|------|-----------------|--------|-------|
| 9.1 | Ask agent something about prior conversation | Agent recalls context from memory | | |
| 9.2 | Install a skill from Hub | Skill appears in local list; can be invoked | | |
| 9.3 | Pair a second device | QR + PIN flow works; sync completes | | |
| 9.4 | Sign in with magic link / OAuth | Account menu shows signed-in state | | |
| 9.5 | Sign out | Account menu returns to signed-out state | | |

---

## 10. Backup / Restore / Uninstall

| # | Step | Expected Result | Status | Notes |
|---|------|-----------------|--------|-------|
| 10.1 | Trigger manual backup | Archive created in `<data-dir>/backups/`; manifest present | | |
| 10.2 | Verify archive encryption | SQLite files are not plaintext inside the zip | | |
| 10.3 | Restore from backup | Data restored; daemon reloads DB; API keys visible again | | |
| 10.4 | Trigger auto-backup | Scheduler creates archive within configured interval | | |
| 10.5 | Preview uninstall | Lists files to be removed; backup offered | | |
| 10.6 | Execute uninstall | Files removed; backup created first | | |
| 10.7 | Re-install and restore | Previous data restored from backup | | |

---

## 11. Auto-Update

| # | Step | Expected Result | Status | Notes |
|---|------|-----------------|--------|-------|
| 11.1 | Check for updates | Returns current version; no false positives | | |
| 11.2 | Verify manifest signature | Ed25519 signature valid; bad signature rejected | | |
| 11.3 | Simulated update (test channel) | Update applies and restarts correctly | | |

---

## 12. Performance Budgets

| # | Metric | Target | Measured | Pass/Fail |
|---|--------|--------|----------|-----------|
| 12.1 | Cold start to overlay-ready | < 500ms | | |
| 12.2 | Hotkey → overlay visible | < 100ms | | |
| 12.3 | First token from LLM | < 1.5s (streaming) | | |
| 12.4 | IPC round-trip (local) | < 5ms | | |
| 12.5 | Memory footprint (idle) | < 150MB | | |
| 12.6 | Binary size | < 20MB | | |

---

## Sign-off

| Role | Name | Date | Verdict |
|------|------|------|---------|
| QA Lead | | | |
| Security Reviewer | | | |
| Release Engineer | | | |
| Product Lead | | | |

**Ship-ready only if all rows are PASS and no P0/P1 issues remain open.**
