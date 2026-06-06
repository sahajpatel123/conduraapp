# Architecture 02 — Computer Use (4-Tier)

> How Synaptic actually controls the computer — and why we use a tiered system.

---

## The Problem

To do physical things in apps (click a button, type a message, drag a file), the agent needs to control the computer. There are several ways to do this, each with tradeoffs:

| Tier | Method | Speed | Reliability | Battery | Privacy | Cross-platform |
|---|---|---|---|---|---|---|
| 1 | **OS CLI / AppleScript** | Fastest | Brittle (depends on app) | Best | Best | Poor (Apple only) |
| 2 | **Accessibility API (AX/UI Automation/AT-SPI)** | Fast | Good (modern apps) | Good | Good | Yes (with quirks) |
| 3 | **Computer-Use MCP servers** (PyAutoGUI, nut.js, etc.) | Medium | Medium | Medium | OK | Yes |
| 4 | **Vision CUA** (Claude/GPT-4o computer use) | Slowest | Variable | Worst | Worst (sends screenshots to LLM) | Yes |

**Tier 4 is the only tier that works in 100% of cases**, but it's slow, expensive, battery-hungry, and **sends screenshots of your screen to an LLM** — a privacy disaster.

**The right answer is: use the lowest tier that works for the current task, escalate only when needed.**

---

## The 4-Tier System

```
            ┌──────────────────────────────────┐
Tier 4 ──►  │   Vision CUA (last resort)      │ ◄── "the screenshot loop"
            │   (Anthropic/OpenAI CUA)        │
            └──────────────┬───────────────────┘
                           │ escalate (with reason)
            ┌──────────────▼───────────────────┐
Tier 3 ──►  │   Cross-platform MCP            │ ◄── PyAutoGUI / nut.js
            │   (computer-use MCP server)     │
            └──────────────┬───────────────────┘
                           │ escalate (AX insufficient)
            ┌──────────────▼───────────────────┐
Tier 2 ──►  │   Accessibility API             │ ◄── ORAX Eye / Windows UIA / AT-SPI
            │   (deterministic, fast)         │
            └──────────────┬───────────────────┘
                           │ escalate (Tier 1 script missing)
            ┌──────────────▼───────────────────┐
Tier 1 ──►  │   OS CLI / AppleScript          │ ◄── Fastest, but brittle
            │   (where available)             │
            └──────────────────────────────────┘
```

### Tier 1: OS Native (AppleScript / Windows PowerShell / Linux xdotool)

For native apps with a CLI bridge:

- **macOS**: AppleScript, JXA, `osascript`, app-specific CLI tools (`shortcuts`, `automator`).
- **Windows**: PowerShell, AutoHotkey, app-specific CLIs.
- **Linux**: xdotool, wmctrl, app-specific CLIs.

Pros: fast, deterministic, low battery.
Cons: requires per-app knowledge, brittle to UI changes.

The agent has a "Tier 1 script library" — a growing collection of known-good scripts for common apps. When a new app is encountered, the agent tries to find or generate a script. If that fails, escalate.

### Tier 2: Accessibility API

The OS exposes the UI tree (every visible element with a name, role, value, position). The agent reads this tree and reasons about it.

- **macOS**: `XCTest` + `AXUIElement` API (via Swift bridge), or the ORAAX Eye bridge we'll build.
- **Windows**: `UI Automation` API (via PowerShell or C# bridge).
- **Linux**: `AT-SPI` via D-Bus.

Pros: fast, deterministic, doesn't require sending screenshots anywhere.
Cons: doesn't work for canvas/3D/image content, doesn't work for apps that don't expose their UI (some games, some Electron apps).

**This is the workhorse tier.** ~70% of all computer use actions go through Tier 2.

### Tier 3: Cross-Platform MCP

When AX is insufficient, fall back to a computer-use MCP server (PyAutoGUI on macOS/Windows, nut.js or `xdotool` on Linux). This tier:

- Takes control of mouse/keyboard via OS events.
- Takes screenshots for verification.
- Uses element detection (e.g., find the button labeled "Submit" by image match).

Pros: works in 100% of cases that vision works in.
Cons: medium speed, medium battery, screenshots are still taken (but only briefly).

### Tier 4: Vision CUA (Last Resort)

When nothing else works (canvas-rendered apps, games, image-only PDFs), the agent falls back to vision-based computer use (Anthropic's CUA, OpenAI's Operator, Gemini's CUA, etc.). The LLM:

1. Looks at a screenshot.
2. Decides: "click at (x, y)" or "type 'hello'".
3. Action is executed.
4. New screenshot is taken.
5. Repeat.

Pros: works everywhere.
Cons: slow (1-5s per step), expensive (a screenshot is hundreds of KB, sent to LLM), battery-hungry (continuous screenshots), privacy-poor (screenshots of your screen are sent to the LLM).

**This tier is the only one that always works. It's also the only one the user can disable.** A user with privacy concerns can disable Tier 4 entirely — Synaptic will refuse tasks that require it.

---

## When Each Tier is Picked

The "Tier Picker" is a small, deterministic classifier:

```
For an action (e.g., "click the Submit button"):
  Tier 1: Is there a known CLI for this app that does "submit"?
          If yes → use Tier 1.
  Tier 2: Can AX read the element (label="Submit", role=button, visible)?
          If yes → use Tier 2.
  Tier 3: Can MCP find the element (image match, OCR)?
          If yes → use Tier 3.
  Tier 4: Take screenshot, ask vision CUA.
          If user has disabled Tier 4 → refuse.
```

The decision is **logged** with the reasoning. The user can see: "Clicked Submit via Tier 2 (AX) — element found by name."

---

## The Computer-Use Cycle (One Action)

```
1. Strategist proposes action: "click Submit on this form"
2. Selective Perception verifies the element is still there
   (twin-snapshot: pre-action screen matches what Strategist saw)
3. Tier Picker picks the tier (usually 1 or 2)
4. Gatekeeper checks: is this destructive? Does it need consent?
5. If consent needed → native dialog, wait
6. Execute via picked tier
7. Verify the action took effect (post-action snapshot, AX read)
8. If verification fails → roll back if possible, or report
9. Audit log
```

This entire cycle, for Tier 1 or 2, takes 50-300ms.

---

## The Three Backends (Pinned)

The CLAUDE.md pins these three computer-use libraries:

1. **macOS**: **ORAAX** (the only macOS-specific CUA we ship; or use `pyautogui` as fallback).
2. **Windows / Linux / fallback**: **PyAutoGUI** (battle-tested, cross-platform, in Python).
3. **Linux (preferred over PyAutoGUI)**: **`nut.js`** (Node.js) or **`xdotool`** (C, fastest).

These are wrapped as **Python subprocesses** that the Go daemon spawns over the **JSON-RPC bridge**. See ADR-0003.

---

## Selective Perception + Computer Use

Computer use is **only** triggered when:

1. The user explicitly asks for an action ("click the button"), OR
2. The Strategist decides the action is necessary, AND
3. The Gatekeeper approves (consent for destructive), AND
4. The user's policy allows it for this app/action type, AND
5. Energy budget is sufficient (if battery is critical, refuse or warn).

The **Selective Perception** module is what tells the Strategist what's on screen in the first place. It uses the **lowest-capture-strategy** that gives enough info:

- **None**: nothing captured (e.g., user is offline/away).
- **AX-only**: AX tree only, no screenshots.
- **Window-rect**: screenshot of the active window only.
- **Differential**: screenshots of dirty regions only.
- **Full**: full screen capture.
- **Vision CUA**: full + vision model.

For Tier 1/2 actions, **AX-only is enough**. The agent reads the AX tree, finds the element by name/role, sends the action.

For Tier 3/4 actions, **a screenshot is required** (to verify or to feed the CUA). The screenshot is encrypted at rest in the replay log.

See [03-perception.md](03-perception.md) for full details.

---

## Failure Modes & Recovery

| Failure | Recovery |
|---|---|
| Tier 2 AX tree empty (app not exposing UI) | Escalate to Tier 3 |
| Tier 3 can't find element by image match | Escalate to Tier 4 |
| Tier 4 CUA loops (clicks same button twice) | Anomaly detector pauses, user notified |
| Action verification fails (post-snapshot doesn't match expectation) | Roll back if possible, or report to user |
| Screenshot sensitive (banking app, password field) | Hardcoded blocklist, refuse, ask user |
| Tier 4 disabled by user | Refuse task, suggest alternative |
| App crashes during action | Catch, report, pause, let user fix |

---

## Privacy Hardpoints

- **Tier 1/2/3 actions never send a screenshot to an LLM.** Only the AX tree and the action are sent.
- **Tier 4 screenshots go to the CUA provider.** This is a hard fact — the user must consent before Tier 4 is used.
- **Screenshots taken for Tier 3 verification** are stored locally in `~/.synaptic/replay/` and **encrypted at rest**.
- **The replay log is per-user-only.** No one else can read it. **The user can delete it at any time.**
- **The replay log auto-deletes after 30 days** by default (user-configurable).
- **Banking/PII hardcoded blocklist**: never take a screenshot of a bank app, never send a credit card number to an LLM. (See [04-safety.md](04-safety.md).)

---

## Action Replay

Every action is recorded with:

- The pre-action state (screenshot, AX tree, app context).
- The action (tier, what was done, target).
- The post-action state.
- The verification result.
- A "diff" view: "What changed on screen because of this action."

The user can scrub through past sessions like a video. This is the **first-class observability** feature.

**Encrypted at rest. Auto-deleted after 30 days by default.**

---

## Related Docs

- [00-overview.md](00-overview.md) — The conductor pattern
- [01-router.md](01-router.md) — How the router picks Tier-4 vision models
- [03-perception.md](03-perception.md) — Selective Perception
- [04-safety.md](04-safety.md) — Hardcoded blocklist
- [07-memory.md](07-memory.md) — How memory of "apps the user uses" informs tier choice
