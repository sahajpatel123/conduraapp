# Architecture 05 — The User-Adaptive Engine

> Synaptic learns how you work. You can always see, edit, and delete what it learns.

---

## The Goal

A new user installs Synaptic. It works.

A user uses Synaptic for 3 months. It **anticipates**.

- It knows which app you reach for in the morning.
- It knows your writing style and uses it.
- It knows which email domains you trust and which to be cautious of.
- It knows when you're focused vs. scattered.
- It knows "the way you do X" and does X that way.

This is **personalization without surveillance**. All learning happens **on-device**. The user can **always** see, edit, and delete anything the engine has learned.

---

## The Closed Learning Loop

```
   ┌──────────┐
   │  User    │
   └────┬─────┘
        │ actions (keystrokes, app switches, queries, edits)
        ▼
   ┌──────────────┐
   │  OBSERVER    │  local-only, no telemetry
   └────┬─────────┘
        │ observations (raw, time-stamped, encrypted)
        ▼
   ┌──────────────┐
   │  DIALECTIC   │  proposer + critic + adjudicator
   └────┬─────────┘
        │ user model updates
        ▼
   ┌──────────────┐
   │  USER MODEL  │  structured, versioned, exportable
   └────┬─────────┘
        │ predictions (preferences, style, expertise, habits)
        ▼
   ┌──────────────┐
   │  PREDICTOR   │  next-action suggestions, smart defaults
   └────┬─────────┘
        │ suggestions
        ▼
   ┌──────────┐
   │  User    │  can accept, modify, or reject
   └──────────┘
```

---

## The Observer

**What it watches**:
- Which apps the user opens, when, and for how long.
- What they type in those apps (with PII redaction; see below).
- What the user asks the agent.
- What the user accepts, rejects, or edits in agent responses.
- Calendar events, file activity, browser tabs.
- (Optional) screen context via the **lowest-energy** perception strategy (see Selective Perception).

**What it does NOT watch**:
- 1Password, banking apps, password managers — hardcoded blocklist.
- Encrypted files (file system reports them as "encrypted" without opening).
- Incognito / private browsing mode.
- Apps the user has marked as "do not observe."

**PII redaction**: text observations are passed through the **same PII redactor** as perception. Passwords, credit card numbers, SSNs are redacted before they ever reach the user model. Names, emails, phones are configurable.

**The Observer runs in a low-priority goroutine**. It's never on the hot path. It batches observations and writes to the user model at most every 5 minutes.

---

## The User Model

A **structured, versioned, exportable** representation of the user.

```yaml
user_model:
  version: 47
  last_updated: 2026-06-06T14:32:00Z
  
  identity:
    name: "Sahaj Patel"
    role: "Software engineer"
    timezone: "America/Los_Angeles"
    languages: [en-US, hi-IN]
  
  preferences:
    writing_style: "concise, technical, friendly"
    verbosity: "low"  # for the agent's responses
    code_style: "Go idiomatic, no comments unless asked"
    tools_preferred: [VSCode, terminal, vim, fzf]
    apps_frequent: [Safari, Mail, Slack, Linear, Terminal]
    apps_avoided: [Twitter, TikTok]
  
  expertise:
    domains:
      programming: expert
      system_design: expert
      product: intermediate
      finance: novice
      cooking: novice
  
  habits:
    morning_routine: [email, calendar, code, news]
    focus_blocks: ["09:00-12:00", "14:00-17:00"]
    energy_peak: morning
    energy_low: after_lunch
  
  social:
    trust_domains: [github.com, wellfound.com, stripe.com]
    distrust_domains: [random-email-claim.com]
    contact_priority: ["email", "slack", "sms"]
  
  consent:
    observation_level: "moderate"  # none | minimal | moderate | full
    personalization: true
    predictive_suggestions: true
    
  beliefs:
    - id: 1
      statement: "User prefers terminal-first workflows"
      confidence: 0.92
      source: "inferred"
      created: 2026-05-12
```

The model is **versioned**. Every change is a new version. The user can diff versions, roll back, and see history.

---

## The Dialectic (Honcho-Style)

Inspired by [Honcho](https://github.com/plastic-labs/honcho), the engine uses a **dialectic** approach:

- **Proposer** (LLM, e.g., local small model): generates a candidate user-model update from observations.
- **Critic** (LLM, possibly different model): challenges the proposer's update. "This conclusion is not supported by the observations. Counter-evidence: ..."
- **Adjudicator** (deterministic rules + small model): makes the final call. If the critic's challenge is strong, the update is rejected. If the proposer is convincing and the critic has weak counter-evidence, the update is accepted.

The dialectic prevents the engine from forming **hallucinated beliefs** about the user. If the evidence is weak, the engine doesn't learn the wrong thing.

**Example**:
- Proposer: "User prefers dark mode."
- Critic: "They've been using light mode in Safari for the last 30 days. Evidence is 12 light-mode sessions vs. 2 dark-mode sessions. Reject."
- Adjudicator: Reject. Confidence threshold not met.

**The dialectic runs locally, on a small model or even rule-based.** No cloud calls for this.

---

## The Predictor

Given the user model, the predictor makes **suggestions**:

- "You're opening Linear. Want me to fetch your open issues?"
- "You're writing an email to <name>. Want me to draft a response in your style?"
- "You usually do <X> at this time. Want me to start?"
- "This email is from an untrusted domain. Want me to be cautious about any links?"

**Three modes**:

1. **Suggest (default)**: show suggestions in the overlay, user clicks to accept.
2. **Anticipate**: agent pre-loads data (e.g., fetches emails) so the response is instant when the user asks.
3. **Auto**: agent takes low-risk actions without asking (e.g., summarizes email subject lines).

The user picks the strength. By default, **Suggest only** is on.

---

## The 4 Strength Levels

```yaml
adaptive:
  level: 1   # 0=off, 1=suggest, 2=anticipate, 3=auto
  visibility: "always"  # always | when-suggesting | when-acting | never
```

| Level | Behavior | Risk |
|---|---|---|
| **0 — Off** | Engine does not learn. Pure stateless agent. | None |
| **1 — Suggest** | Engine learns. Surfaces suggestions in overlay. User accepts/rejects. | Low |
| **2 — Anticipate** | Engine learns + pre-loads. User can still veto. | Medium |
| **3 — Auto** | Engine learns + acts on low-risk predictions. User can audit. | High (user must explicitly enable) |

**Most users will use Level 1 or 2.**

---

## Visibility — What the User Sees

When `visibility: always` (default for Level 1+):

- A small **"🧠 Engine"** indicator in the overlay shows what it's thinking.
- Tapping opens a panel: "I noticed you opened Linear 3 times this morning. Should I remember that as a habit?"
- A daily digest: "Here's what I learned about you this week. [Review] [Edit] [Delete]."

When `visibility: when-suggesting`:
- Engine is silent until it has a suggestion. Then it shows the suggestion with its reasoning.

When `visibility: when-acting`:
- Engine is silent for Level 1 and 2. For Level 3, it announces actions.

When `visibility: never`:
- Engine runs but the user is blind to it. **Strongly discouraged.**

---

## On-Device Only, Period

The user model **never** leaves the device. It is **never** synced to our servers, **never** used to train a model, **never** sold. It is the user's.

**The P2P sync feature** is **opt-in** and the model is end-to-end encrypted between the user's own devices. Even we cannot read it.

**The Skills Hub** does **not** see the user model. Skills that import settings get only the user-approved subset, with consent at install time.

---

## Export, Edit, Delete

The user can:

- **Export** the entire user model as JSON: `Settings → Adaptive Engine → Export`.
- **Edit** any field directly: `Settings → Adaptive Engine → Edit`.
- **Delete** any belief: `Settings → Adaptive Engine → Beliefs → Delete`.
- **Pause** observation: `Settings → Privacy → Pause observation`.
- **Reset** the entire model: `Settings → Adaptive Engine → Reset to defaults`.

Resetting does not delete the audit log (that's separate, for safety reasons).

---

## Open Questions (Deferred)

- **Multi-user devices**: if the device is shared (family computer), how do we partition the user model? (Decision deferred to v0.2 — v0.1 assumes single primary user.)
- **Cross-device model merging**: when P2P sync is enabled, how do we merge two models? (Decision deferred to v0.2.)
- **Time-decay**: should old beliefs decay in confidence over time? (Likely yes; not yet specified.)

---

## Related Docs

- [00-overview.md](00-overview.md) — The closed learning loop
- [01-router.md](01-router.md) — How the router uses user-model preferences
- [08-sync.md](08-sync.md) — How the user model syncs across devices (E2E encrypted)
- [PRIVACY.md](../PRIVACY.md) — Privacy guarantees for the user model
