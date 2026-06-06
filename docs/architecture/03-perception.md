# Architecture 03 — Selective Perception

> One unified system for **battery awareness** AND **safety**. The agent sees only as much as it needs — and the system is safer for it.

---

## The Insight

"Perception" in an AI agent means: what does the system **see** about the user's world?

Two problems:

1. **Battery**: continuous screen capture drains battery. 30fps full-screen capture is roughly 15-25% CPU. This is unacceptable on laptops.
2. **Safety**: the more the agent sees, the more it can do, and the more it can do wrong. Sending a screenshot of your banking app to an LLM is a privacy disaster.

These are actually **the same problem**: the agent should only see as much as it needs to do the current task. **Less perception = less battery = less privacy exposure = less that can go wrong.**

Selective Perception is one unified system that solves both.

---

## The 6 Capture Strategies

The system picks a strategy per "perception frame" (whenever the agent needs to know what's on screen):

| Strategy | What is captured | Battery cost | Privacy cost | When used |
|---|---|---|---|---|
| **None** | Nothing | Zero | Zero | Agent is idle / user is away / battery critical |
| **AX-only** | Accessibility tree (no pixels) | Negligible | Minimal (no screenshots) | 70% of cases — workhorse |
| **Window-rect** | Screenshot of the active window only | Low | Low | When AX is insufficient for the active app |
| **Differential** | Only "dirty" regions of the screen | Low-medium | Low | Long-running tasks, watching for change |
| **Full** | Full screen capture | Medium-high | Medium | Rarely, when full context is needed |
| **Vision CUA** | Full + sent to vision LLM | High | High | Last resort, user opt-in per session |

---

## Dirty Tracking

The key to "differential" is knowing what changed. We implement dirty tracking natively:

- **macOS**: `CGEventTap` to listen for mouse/keyboard events, plus `CGWindowListCopyWindowInfo` for window changes. Combine with periodic full-frame hashing (every 5s) to catch content changes (e.g., animation).
- **Windows**: `SetWindowsHookEx` for input, `EnumWindows` for window changes, frame differencing for content.
- **Linux**: `XRecord` extension or evdev, plus `xdotool getactivewindow` for window changes.

When the strategy is "Differential", we capture **only the bounding rects that changed** since the last frame. This can be 90%+ savings on a static screen.

---

## Energy Budget

The user sets an energy mode (or "Auto"):

```yaml
perception:
  energy_mode: auto   # low | balanced | high | auto
  low_battery_threshold: 20  # percent
  critical_battery_threshold: 10
  high_battery_charging: true
```

| Mode | Behavior |
|---|---|
| **Low** | Only AX-only, Window-rect for active app, No differential. Max 2 fps if differential. |
| **Balanced** (default) | AX-only preferred, Differential up to 5fps, Full only on demand. |
| **High** | Differential up to 10fps, Full on demand, Vision CUA enabled. |
| **Auto** | Switches based on battery %: <20% → Low, <10% → Critical (no perception, refuse tasks), charging → High. |

### The Critical Refuse

When battery is **critical** (<10%, or user-set threshold) AND the task requires vision, the system **refuses the task and asks the user to plug in**. This is non-negotiable.

### When Vision is Needed + Budget Hit

If the user asks for something that requires vision CUA (Tier 4) but energy is Low, the system surfaces a native dialog: "This task requires vision. Switch to Balanced/High energy mode for this session, or skip the vision part?" The user decides.

---

## Per-App Profiles

The user (and the Adaptive Engine) can set per-app capture strategy overrides:

```yaml
perception:
  app_profiles:
    safari:     { default: ax-only, max: differential }
    finder:     { default: ax-only, max: ax-only }
    figma:      { default: full, max: vision-cua }  # canvas-rendered, no AX
    photoshop:  { default: window-rect, max: vision-cua }
    banking_*:
      default: none    # never capture
      max: none
      deny: true       # never perceive, refuse tasks
    email_*:
      default: ax-only
      redact_pii: true  # see PII redaction
```

The agent can request a higher strategy than the default, but the Gatekeeper enforces the **max** ceiling.

---

## PII Redaction

When AX-only is used, the AX tree may contain PII (text in form fields, document content, etc.). Before sending the AX tree to an LLM, we run a **PII redaction pass**:

- **Always redact**: passwords, credit card numbers, SSNs, API keys (regex + entropy check).
- **Configurable**: name, email, phone, address (user can disable redaction in Settings).
- **Smart redaction**: replace with `[REDACTED:EMAIL]` or with a hash for context-preserving redaction.

The redacted AX tree is what goes to the LLM. The original is **never** sent.

This applies to:
- AX-only captures
- Differential captures (only redacted regions)
- Full captures (redact before send)
- Vision CUA (we add a system prompt: "If you see a credit card number, do not repeat it. Refer to it as [REDACTED:CC].")

---

## Pause on Privacy

For sensitive app categories (banking, password managers, 1Password, Authy, signal), the system can be configured to **never perceive**:

```yaml
perception:
  pause_apps:
    - com.1password.*
    - com.agilebits.*
    - com.bankofamerica.*
    - com.chase.*
    - com.paypal.*
    - com.coinbase.*
    - com.robinhood.*
    - com.venmo.*
    - "*Authy*"
    - "*Signal*"
    - "*WhatsApp*"
  pause_default: pause   # default to paused for unknown apps
  resume_action: tap-to-resume  # user must tap to resume
```

When a "pause app" is foregrounded, **all perception is suspended**. The agent will not know what's on screen. If a task requires perception, the system says: "This requires me to see what's in 1Password. Please unlock it manually, then tap to resume."

**Hardcoded list of "never perceive" apps** is shipped (banking, password managers, 2FA apps). User can add to it.

---

## The Perception Pipeline (One Frame)

```
1. Trigger: Strategist needs to know "what's on screen"
2. Dirty tracker: what's changed since last frame?
3. Energy check: can we capture given battery?
4. App profile: what's the max for this app?
5. Pick strategy:  None | AX-only | Window-rect | Differential | Full | Vision CUA
6. Capture: do it
7. PII redact: scrub sensitive content
8. Hand to Strategist LLM
9. Audit: "Perception frame N: strategy=X, app=Y, redacted=Z, sentTo=LLM"
```

**All of this is logged.** The user can see: "Frame 234: Window-rect, Safari, redacted 3 emails, sent to claude_code."

---

## Why "Unified" Matters

If perception and safety were separate, the user would have to configure two things:

- "Max battery: 15% CPU on screen capture"
- "Never send screenshots of banking app"

These are the same kind of decision: "see less." Selective Perception is **one knob** (energy mode + per-app profiles) that does both.

The Adaptive Engine can also **learn** the user's preferences over time:

- "User opens 1Password often — add to pause list."
- "User has Outlook open 8 hours a day — switch to AX-only for Outlook (it's stable)."
- "User is at 8% battery — refuse all vision tasks."

This is the **closed learning loop** with safety.

---

## What the User Sees

A small persistent indicator in the overlay:

```
┌─ Perception ─────────────────────┐
│  ⚡ Energy: Balanced (78%)        │
│  👁️  AX-only (Safari)             │
│  🔒 PII: 3 emails redacted       │
│  🛑 Banking apps: paused         │
│  Last frame: 120ms ago           │
└──────────────────────────────────┘
```

Tapping expands to a per-frame log:

```
Frame 1234 · 14:32:01 · Safari · AX-only · 89ms
Frame 1233 · 14:32:00 · Mail ·   AX-only · 102ms · 1 phone redacted
Frame 1232 · 14:31:55 · Safari · AX-only · 76ms
Frame 1231 · 14:31:50 · 1Password · NONE (pause rule)
Frame 1230 · 14:31:48 · Safari · AX-only · 92ms
```

This is **full transparency**. The user always knows what the agent is seeing.

---

## Related Docs

- [00-overview.md](00-overview.md) — The conductor pattern
- [02-computer-use.md](02-computer-use.md) — How computer use consumes perception
- [04-safety.md](04-safety.md) — The hardcoded blocklist and twin-snapshot verification
- [05-adaptive.md](05-adaptive.md) — How the Adaptive Engine informs perception
