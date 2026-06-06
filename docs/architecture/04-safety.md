# Architecture 04 — The Safety Layer

> The 5 modules and 7 non-negotiables that keep Synaptic safe.

---

## The Philosophy

Synaptic performs physical, often irreversible actions on the user's computer. A single bug, a single prompt injection, a single moment of inattention could result in data loss, financial damage, or worse.

Therefore: **safety is not a feature. It is the foundation.** Every other feature is built on top of safety. If safety and convenience conflict, safety wins, every time, no exceptions.

The 7 non-negotiables (from `CLAUDE.md` Section 2) are the contract. If a feature requires violating one, the feature is not built.

---

## The 5 Modules

The safety layer has 5 modules that work together:

```
┌────────────────────────────────────────────────────────────┐
│                      USER REQUEST                          │
└──────────────────────────┬─────────────────────────────────┘
                           ▼
              ┌─────────────────────────┐
              │  1. STRATEGIST (LLM)    │  proposes action
              └────────────┬────────────┘
                           ▼
              ┌─────────────────────────┐
              │  2. GATEKEEPER          │  approves/denies
              │     (deterministic)     │
              └────────────┬────────────┘
                           ▼
              ┌─────────────────────────┐
              │  3. BLAST-RADIUS        │  classifies impact
              │     CLASSIFIER          │
              └────────────┬────────────┘
                           ▼
              ┌─────────────────────────┐
              │  4. ANOMALY DETECTOR    │  pauses on weirdness
              │     (behavioral)        │
              └────────────┬────────────┘
                           ▼
              ┌─────────────────────────┐
              │  5. AUDIT LOG           │  records everything
              │     (HMAC-chained)      │
              └─────────────────────────┘
```

---

## Module 1: The Strategist (LLM)

**What it is**: The LLM that plans and proposes. It is **not** the LLM that executes.

**Properties**:
- Can be ANY LLM (Claude, GPT, Gemini, local Ollama, custom).
- Has no authority to take physical action.
- Speaks in plans and proposals, not commands.
- **Can be prompt-injected.** We assume it can be. That's why the Gatekeeper exists.

**Output**: A `Plan` containing ordered `Action` objects with:
- Goal
- Target (app, file, URL, command)
- Pre-conditions
- Expected post-conditions
- Confidence score

The Strategist **never sees the final "Approve" button**. The Gatekeeper is the only one who can press it.

---

## Module 2: The Gatekeeper (Deterministic)

**What it is**: A rule-based system (in Go) that has the **exclusive authority** to execute physical actions. Cannot be prompt-injected because it **does not process LLM output** in a way that affects its decision logic. It only sees the structured `Action` proposal.

**The flow**:
```
For each proposed Action:
  1. Validate structure (target exists? pre-conditions met?)
  2. Look up the action's blast radius (via Module 3)
  3. Look up the user's policy for this app + action type
  4. Check user presence + consciousness (see Presence Tracker)
  5. Decide: ALLOW | ALLOW_WITH_CONSENT | DENY
  6. If ALLOW_WITH_CONSENT: pop native dialog, wait
  7. If ALLOW: hand to executor
  8. If DENY: surface reason, do not execute
  9. Log everything
```

**The Gatekeeper is the ONLY path to physical action.** This is non-negotiable. There is no "fast path" or "trusted mode" that bypasses it.

**Properties**:
- Written in Go, not in LLM code.
- Pure rule-based. No ML.
- Coverage: 100% of actions.
- Latency: <1ms per check.

---

## Module 3: Blast-Radius Classifier

**What it is**: A static, deterministic classifier that assigns one of 4 levels to every action.

| Level | Description | Examples | Default policy |
|---|---|---|---|
| **READ** | Reading data | Read file, read email, take screenshot | ALLOW (audit only) |
| **LOCAL** | Local, reversible | Create a note, save a draft, rename file | ALLOW (audit) |
| **NETWORK** | Outbound, partially reversible | Send email, post to social, open URL | ALLOW_WITH_CONSENT (per session) |
| **DESTRUCTIVE** | Cannot easily undo | Delete file, send message, transfer money, run shell command | ALLOW_WITH_CONSENT (per action) + 2FA if financial |

**The classifier is a rule table**, not ML. New actions are added to the table when encountered. The user can override the classification for any action type.

**The classifier is independent of the LLM.** It is a Go function: `func classify(action Action) BlastRadius`.

---

## Module 4: Anomaly Detector

**What it is**: A behavioral watchdog that pauses the agent on suspicious patterns.

**What it watches for**:
- Action rate spikes (e.g., 50 actions in 10 seconds).
- Repeated identical actions (the agent is stuck in a loop).
- Cost anomalies (e.g., burning $20 in 5 minutes).
- Tool-call patterns that match known exploits.
- Mismatch between user context and action context.
- Reaching a sensitive page (banking, password manager).
- Model output containing exfiltration patterns.
- Network calls to non-allowlisted domains.
- Process spawning (suspicious subprocesses).

**On detection**:
1. **Pause** all agent activity immediately.
2. **Lock** the agent: require user to dismiss the alert before continuing.
3. **Surface** a native dialog: "I noticed [pattern]. I paused myself. Are you sure you want me to continue?"
4. **Log** the incident in the audit log with severity.

The Anomaly Detector is the "is this okay?" voice. It can be **inconvenient** (false positives) but it is **never** disabled. It can be set to "less sensitive" but never "off".

---

## Module 5: Audit Log (HMAC-Chained)

**What it is**: An append-only log of every action the agent takes (or attempts to take).

**What is logged**:
- Timestamp (UTC, monotonic)
- Session ID
- Action type, target, params
- Blast radius classification
- Gatekeeper decision (allow/deny/consent)
- User response (if consent was needed)
- Anomaly score
- Result (success/fail)
- LLM/model used
- Tokens used, $ spent
- Screenshot hash (not the image itself, for privacy)
- Action replay ID (so user can scrub to that moment)

**Why HMAC-chained**:
- Each entry includes a HMAC of `(prev_hash || this_entry)`.
- Tampering with any entry breaks the chain from that point on.
- The HMAC key is derived from a per-install secret, never sent anywhere.
- The user can verify the chain at any time: `synaptic audit verify`.

**Retention**:
- Default: 90 days.
- User-configurable: 30 days to forever.
- Critical events (destructive actions, anomalies) are kept indefinitely unless the user purges.

**Where it's stored**:
- `~/.synaptic/audit.log` (encrypted at rest via the store).
- Encrypted with the same key as the rest of the local store.
- Can be exported to JSON for the user's own analysis.

---

## The Presence Tracker

A safety subsystem: is the user actually present and aware?

| Signal | What it means |
|---|---|
| **Active input** (keyboard/mouse) in last 60s | Likely present |
| **Screen locked** | Definitely not present |
| **Lid closed** (laptop) | Definitely not present |
| **User away >5 min** (configurable) | Not present |
| **User logged out** | Not present |
| **Active audio** (mic input) | Possibly present |
| **Camera input** (face detection) | Possibly present |

**Behavior when not present**:
- READ actions: allowed (the agent can do background research, e.g., "summarize my unread emails while I'm out").
- LOCAL actions: queue, ask for consent on return.
- NETWORK actions: queue + require consent on return + wait 1 hour for user to confirm.
- DESTRUCTIVE actions: queued + cannot run without the user unlocking the device.

**"Lock-and-leave" mode**: user can set "I'll be back at 7pm." Agent will queue all actions and only resume at 7pm (or when user returns, whichever is first).

---

## The Kill Switch (3 Layers)

**Layer 1: Hotkey** (hard, fast)
- Default: `Ctrl+Alt+\` on all platforms (configurable).
- Press: agent immediately stops whatever it's doing.
- All in-flight actions are aborted.
- An overlay appears: "I stopped. Resume? Review? Forget it?"

**Layer 2: Watchdog**
- A separate process watches the agent.
- If the agent process is unresponsive for >5s, watchdog kills it.
- If the agent is using too much CPU/RAM/network, watchdog kills it.
- If the agent has been running >2 hours without user interaction, watchdog pauses it.

**Layer 3: Network Isolation**
- The agent can be put in "offline mode" by a single click in the overlay.
- All outbound network is blocked. The agent can only do local actions.
- The router refuses to send to any LLM.
- This is the "I'm freaking out, give me air" button.

All three layers are **independent of the agent**. The agent cannot disable its own kill switch.

---

## The 7 Non-Negotiables (Recap)

From `CLAUDE.md` Section 2:

1. **Strategist and Gatekeeper are separate systems.**
2. **Gatekeeper is the only path to physical action.**
3. **Destructive actions require a real human at the keyboard.**
4. **User can always stop the agent.**
5. **Every action is auditable.**
6. **Agent is a guest, not an owner.**
7. **OS permissions are granted by the user.**

These are inviolable. If a feature requires violating one, the feature is not built.

---

## Threat-Model-Specific Defenses

### Defending against Prompt Injection

The Strategist is assumed to be prompt-injectable. The defenses:

- **Threat pattern scanner**: known injection patterns in tool output, web pages, files are flagged.
- **Delimiter markers**: tool output is wrapped in `<TOOL_OUTPUT>...</TOOL_OUTPUT>` and the Strategist is told to treat content within as untrusted.
- **Confirmation step**: for any action involving a tool result (e.g., "open the link in this email"), the user is asked: "This came from an email. Is the source trusted?"
- **Skill sandboxing**: skills are run with limited tool access.
- **No tool calls in tool output**: an LLM never sees a tool result and gets to call a tool in the same turn. The result is parsed, redacted, and the next turn is a fresh prompt.

### Defending against the LLM Going Rogue

- **Action replay**: every action is replayable and undoable (when possible).
- **Anomaly detector**: behavior that deviates from the user's pattern is paused.
- **Cost cap**: per-session and per-day spend limits. The agent cannot exceed them.
- **Domain allowlist**: network calls are restricted to allowlisted domains. The agent cannot call home to a non-allowlisted server.
- **HMAC chain**: tampering is detected.
- **No self-modification**: the agent cannot modify its own binary, code, or policy. Period.

### Defending against the User Making a Mistake

- **Confirmation dialogs** for any DESTRUCTIVE action.
- **Preview** of what will happen: "I will delete 47 files in `~/Downloads/old/`. Continue?"
- **Undo window** for LOCAL actions: the user can press Cmd+Z within 30s to undo a series of actions.
- **Backup before destructive**: if the action involves deleting or modifying files, a backup is taken first.
- **"What did you just do?" button**: in the overlay, shows the last 5 actions with an undo button.

---

## Related Docs

- [00-overview.md](00-overview.md) — The conductor pattern, the survival invariants
- [02-computer-use.md](02-computer-use.md) — How computer use interacts with the safety layer
- [03-perception.md](03-perception.md) — The PII redaction and pause-on-privacy
- [CLAUDE.md Section 2](../CLAUDE.md) — The 7 non-negotiables
