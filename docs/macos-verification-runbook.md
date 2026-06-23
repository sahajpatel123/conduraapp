# macOS Verification Runbook

> Step-by-step guide for executing Phase 15 on-device verification
> on a clean macOS machine. Follow each step in order.
> Record PASS/FAIL for each step and capture evidence.

## Prerequisites

- Clean macOS machine (arm64 preferred, or Intel Mac)
- No prior Condura/Synaptic install
- No `~/.condura/` or `~/.synaptic/` directory
- No Ollama running, no API keys pre-configured
- A browser (Safari or Chrome)
- Terminal access
- ~2-3 hours

## Evidence Folder

Create before starting:

```bash
mkdir -p ~/Desktop/condura-v0.1.0-verify-macos-$(date +%Y-%m-%d)
```

Save all screenshots and log excerpts here.

---

## Step 1: Download

1. Open browser, go to `https://condura.app/download`
2. Page loads with platform auto-detected as macOS
3. Click the macOS download button
4. Download completes — file is `condura-gui-darwin-arm64.dmg`
5. Verify file size matches (should be ~30-50MB)

**Evidence:** Screenshot of download page + completed download in Finder.

**Status:** [ ] PASS / [ ] FAIL

---

## Step 2: Verify Checksum

Open Terminal:

```bash
cd ~/Downloads
shasum -a 256 condura-gui-darwin-arm64.dmg
```

Compare against the checksum in `checksums.txt` from the release page.

**Evidence:** Paste terminal output.

**Status:** [ ] PASS / [ ] FAIL

---

## Step 3: Verify Code Signature

```bash
codesign -dv --verbose=4 /Applications/Condura.app 2>&1 | head -20
```

Or before moving to Applications:

```bash
codesign -dv --verbose=4 ~/Downloads/Condura.app 2>&1 | head -20
```

Look for:
- `Signature size` should be non-zero
- `Authority=Developer ID Application: ...`
- `Signed Time` should be recent

**Evidence:** Paste terminal output.

**Status:** [ ] PASS / [ ] FAIL

---

## Step 4: Install

1. Double-click `condura-gui-darwin-arm64.dmg`
2. Drag `Condura.app` to `Applications`
3. Eject the disk image
4. Open `Applications`, find `Condura.app`
5. Double-click to launch
6. macOS may show Gatekeeper dialog — click "Open"
7. Menu bar icon appears (brain/network icon)

**Evidence:** Screenshot of DMG, Applications folder, first launch.

**Status:** [ ] PASS / [ ] FAIL

---

## Step 5: Onboarding — EULA

1. Onboarding wizard appears with EULA screen
2. Scroll to bottom of the EULA text
3. Checkbox becomes enabled after scrolling
4. Tick the checkbox
5. Click "I Accept"

**Evidence:** Screenshot of EULA screen.

**Status:** [ ] PASS / [ ] FAIL

---

## Step 6: Onboarding — Permissions

1. Permissions screen shows Accessibility + Screen Recording
2. Click "Open System Settings" for Accessibility
3. In System Settings → Privacy & Security → Accessibility, find Condura
4. Toggle it ON
5. Return to Condura — status shows "granted"
6. Click "Open System Settings" for Screen Recording
7. In System Settings → Privacy & Security → Screen Recording, find Condura
8. Toggle it ON
9. Return to Condura — status shows "granted" (or "unknown" with guide)
10. Click Continue (or "Skip for now" if you want to defer)

**Evidence:** Screenshot of permissions screen + System Settings.

**Status:** [ ] PASS / [ ] FAIL

---

## Step 7: Onboarding — Hotkey

1. Hotkey recording screen appears
2. Click "Record Hotkey"
3. Press a key combination (e.g., Option+Option or Ctrl+Space)
4. Combo is displayed
5. Click Continue

**Evidence:** Screenshot of hotkey screen.

**Status:** [ ] PASS / [ ] FAIL

---

## Step 8: Onboarding — Ready

1. Ready screen shows detected backends (Ollama if present, or API key option)
2. Click "Start using Condura"
3. Wizard dismisses
4. Main chat UI appears

**Evidence:** Screenshot of ready screen + main UI.

**Status:** [ ] PASS / [ ] FAIL

---

## Step 9: Configure API Key

1. Open Settings (gear icon or menu bar → Settings)
2. Find "API Keys" section
3. Add an OpenAI or Anthropic API key
4. Save

**Evidence:** Screenshot of settings (redact the key).

**Status:** [ ] PASS / [ ] FAIL

---

## Step 10: Chat

1. Type "Hello, what is 2+2?" in the chat input
2. Press Enter or click Send
3. Streaming response begins (tokens appear one by one)
4. Response completes
5. Click "New Conversation" — new thread created
6. Previous conversation appears in sidebar
7. Click on previous conversation — messages load correctly

**Evidence:** Screenshot of streaming response + conversation list.

**Status:** [ ] PASS / [ ] FAIL

---

## Step 11: Hotkey Overlay

1. Press the hotkey combo you recorded
2. Overlay window appears at cursor position (or center)
3. Type a message in the overlay
4. Submit — response streams in overlay
5. Press Esc to dismiss overlay

**Evidence:** Screenshot of overlay.

**Status:** [ ] PASS / [ ] FAIL

---

## Step 12: Computer Use — Gatekeeper

1. Ask agent: "Open Finder and create a new folder on the Desktop"
2. Gatekeeper consent modal appears (WRITE action)
3. Click "Approve"
4. Folder is created on Desktop
5. Audit log shows the action

**Evidence:** Screenshot of consent modal + created folder.

**Status:** [ ] PASS / [ ] FAIL

---

## Step 13: Computer Use — Deny

1. Ask agent: "Delete the folder I just created"
2. Gatekeeper consent modal appears (DESTRUCTIVE action)
3. Click "Deny"
4. Action is blocked
5. Audit log shows "deny"

**Evidence:** Screenshot of denied action + audit log.

**Status:** [ ] PASS / [ ] FAIL

---

## Step 14: Safety — Kill Switch

1. Press Cmd+Shift+Escape (or your configured kill hotkey)
2. Agent halts — menu bar shows halted state
3. Try sending a message — agent does not respond
4. Press kill hotkey again (or resume from menu) to resume
5. Agent resumes — chat works again

**Evidence:** Screenshot of halted state.

**Status:** [ ] PASS / [ ] FAIL

---

## Step 15: Audit Log

1. Open Settings → Audit Log (or use CLI: `condura audit list`)
2. Log shows entries with timestamp, actor, action, result
3. HMAC chain is valid (log entry hashes link correctly)

**Evidence:** Screenshot of audit log.

**Status:** [ ] PASS / [ ] FAIL

---

## Step 16: Backup

1. Open Settings → Backup
2. Click "Create Backup"
3. Backup file appears in list
4. Verify file exists at `~/Documents/condura-backups/`

**Evidence:** Screenshot of backup list + Finder showing backup file.

**Status:** [ ] PASS / [ ] FAIL

---

## Step 17: Voice (macOS only)

1. Open Settings → Voice
2. Enable wake word
3. Say "hey condura" (or your configured wake word)
4. Voice orb appears
5. Speak a question
6. Transcription appears
7. Response is read aloud via TTS

**Evidence:** Short screen recording (5-10 seconds).

**Status:** [ ] PASS / [ ] FAIL

---

## Step 18: Uninstall

1. Quit Condura (Cmd+Q or menu bar → Quit)
2. Drag `Condura.app` from Applications to Trash
3. Delete `~/.condura/` directory:
   ```bash
   rm -rf ~/.condura
   ```
4. Optionally delete backups: `rm -rf ~/Documents/condura-backups/`

**Evidence:** Screenshot of empty Applications + removed data dir.

**Status:** [ ] PASS / [ ] FAIL

---

## Step 19: Re-install Clean

1. Re-download `condura-gui-darwin-arm64.dmg`
2. Install fresh
3. Onboarding wizard appears (first-run state)
4. No previous data or config persists

**Evidence:** Screenshot of fresh onboarding.

**Status:** [ ] PASS / [ ] FAIL

---

## Performance Measurements

| Metric | Target | Measured | Pass/Fail |
|--------|--------|----------|-----------|
| Cold start to overlay-ready | < 500ms | | |
| Hotkey → overlay visible | < 100ms | | |
| First token from LLM | < 1.5s | | |
| IPC round-trip (local) | < 5ms | | |
| Memory footprint (idle) | < 150MB | | |
| Binary size | < 20MB | | |

---

## Sign-off

| Role | Name | Date | Verdict |
|------|------|------|---------|
| QA Lead | | | |
| Security Reviewer | | | |
| Release Engineer | | | |
| Product Lead | | | |

**Ship-ready only if all rows are PASS and no P0/P1 issues remain open.**
