# Onboarding Verification (Phase 14A)

Manual checklist for the converged 4-screen, value-first onboarding flow:
**EULA → Permissions → Hotkey → Ready**. Run this on a clean machine
before tagging a release. The goal: a brand-new user reaches a working
agent in **≤9 in-app clicks**, with no account and no API key required.

> Maintainer task. End users never run this — it verifies the build for
> the people shipping it.

## Prerequisites

- A clean environment (fresh VM/user account, or remove prior state):
  ```bash
  rm -rf ~/.synaptic
  ```
- A built GUI app for the target OS (see `scripts/package-gui-installers.sh`).
- Optional, to exercise the happy path: [Ollama](https://ollama.com/download)
  installed and running (`ollama serve`) with at least one model pulled
  (`ollama pull llama3.2`).

## macOS clean-install checklist (9 steps)

| # | Action | Expected result |
|---|--------|-----------------|
| 1 | Launch Condura for the first time | The **EULA screen** appears immediately. No login screen, no API-key prompt. |
| 2 | Try clicking **I Accept** without scrolling | Button is disabled. The checkbox is disabled until you scroll to the bottom. |
| 3 | Scroll the EULA to the bottom, tick the checkbox, click **I Accept** | Advances to the **Permissions** screen. Acceptance persists (see "Resumability" below). |
| 4 | Observe the Permissions screen | Exactly **two** rows: Accessibility and Screen Recording, each with a live status badge. No Microphone/Automation/Notifications here. |
| 5 | Click **Open System Settings** on Accessibility | macOS System Settings opens to **Privacy & Security → Accessibility** (deep link). Grant it; within ~2s the badge flips to `granted`. |
| 6 | Click **Skip for now** (or grant Screen Recording, then **Continue**) | Advances to the **Hotkey** screen. (Both the skip and grant paths must advance.) |
| 7 | Click the recorder, press a combo (e.g. `Cmd+Shift+Space`), click **Continue** | The combo is captured and shown; Continue was disabled until a valid combo was recorded. Advances to **Ready**. |
| 8 | Observe the Ready screen | With Ollama running: "Local model ready" + detected model(s). Without Ollama: "No local model detected" + an Ollama install link. Optional cards for API key / messaging are present but not required. |
| 9 | Click **Start using Condura** | Wizard dismisses, main chat UI mounts. With Ollama, sending a message gets a local response — **no API key was entered**. |

**Click budget:** Accept (1) + checkbox (1) + permissions Continue/Skip (1)
+ recorder click (1) + key press (counts as the combo) + Hotkey Continue (1)
+ Start using Condura (1) ≈ **6–9 clicks** depending on permission grants.
Confirm it never exceeds 9 in-app clicks (OS permission dialogs are out of
our control and not counted).

## Post-setup checks

- [ ] Relaunch the app — the wizard does **not** reappear (first-run marker
      + onboarding state both complete).
- [ ] Settings → **OS permissions** shows all five kinds (the wizard only
      asked for two; the rest are here).
- [ ] Settings → **Legal** → "View EULA" shows the full license text and
      the accepted version.
- [ ] Settings → **Setup** → "Re-run setup" relaunches the wizard at the
      EULA step without deleting data.
- [ ] Config persisted: `~/.synaptic/config.yaml` has the chosen
      `hotkey.overlay`, and `llm.providers.ollama.enabled: true` when
      Ollama was reachable at finish.

## Resumability / upgrade

- [ ] Quit the app mid-wizard (e.g. on the Hotkey screen) and relaunch —
      it resumes on the same step (state is in `onboarding_state`).
- [ ] Upgrade path: a user who finished onboarding in a pre-14A build
      (first-run marker set, legacy 8-step state) is **not** re-wizarded.
- [ ] EULA bump: increment the EULA version, relaunch — the wizard forces
      a re-accept of just the EULA step.

## Edge cases

- [ ] **Ollama not installed:** Ready shows the fallback card and the
      install link; finishing still works (chat shows a "configure a
      provider" empty state until a key/model is added).
- [ ] **Ollama running, no models:** Ready suggests `ollama pull llama3.2`.
- [ ] **Permissions denied + skipped:** the app still works for chat;
      computer-use features prompt to grant access when invoked.
- [ ] **Permission status `unknown`** (CI / unsupported platform): badge
      reads `unknown` and Continue is not blocked.
- [ ] **Invalid/empty hotkey:** Continue stays disabled; the daemon
      rejects an empty hotkey at finish.

## Cross-platform smoke (abbreviated)

Repeat steps 1–9 on Windows and Linux. The flow is identical; only the
permission deep links differ:

- **Windows:** `ms-settings:` URIs (e.g. `ms-settings:privacy-accessibility`).
- **Linux:** guides are textual (AT-SPI / portal instructions); deep links
  may be empty — verify the steps render and Continue still works.
