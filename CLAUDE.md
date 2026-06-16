# Synaptic — CLAUDE.md

> **The single source of truth for the Synaptic project.**
> Every AI model and every human working on this project MUST read this file
> end-to-end before starting any work, and must read `LOGBOOK.md` for the latest
> session state. After every working session, append an entry to `LOGBOOK.md`.
>
> This document is append-only in spirit. Existing content may be corrected or
> expanded, but never silently deleted. If a decision is reversed, record both
> the old and new with rationale.

---

## Table of Contents

1. [Mission](#1-mission)
2. [The Survival Rule (Non-Negotiable Invariants)](#2-the-survival-rule)
3. [What Synaptic Is (and Is Not)](#3-what-synaptic-is-and-is-not)
4. [The 26 Locked Decisions](#4-the-26-locked-decisions)
5. [The 7 Technical Non-Negotiables](#5-the-7-technical-non-negotiables)
6. [Selective Perception (Battery + Safety)](#6-selective-perception)
7. [Architecture Overview](#7-architecture-overview)
8. [Tech Stack (Locked)](#8-tech-stack-locked)
9. [The User-Adaptive Engine](#9-the-user-adaptive-engine)
10. [The Safety Layer (5 Modules)](#10-the-safety-layer)
11. [The Computer-Use System](#11-the-computer-use-system)
12. [The Router (Hybrid with Memory)](#12-the-router)
13. [The Delegation Bus](#13-the-delegation-bus)
14. [Memory System (3 Layers)](#14-memory-system)
15. [Skills System](#15-skills-system)
16. [MCP Gateway](#16-mcp-gateway)
17. [P2P Sync](#17-p2p-sync)
18. [Action Replay](#18-action-replay)
19. [Global Hotkey + Overlay + Voice](#19-global-hotkey--overlay--voice)
20. [Onboarding Flow](#20-onboarding-flow)
21. [Interfaces (TUI + Wails + Web)](#21-interfaces)
22. [Distribution & Updates](#22-distribution--updates)
23. [i18n (6 Languages)](#23-i18n)
24. [Cloud Backups + Uninstall](#24-cloud-backups--uninstall)
25. [Spend Monitor + Failover](#25-spend-monitor--failover)
26. [Support + Donations + Marketing](#26-support--donations--marketing)
27. [The Autonomy Matrix](#27-the-autonomy-matrix)
28. [Build Order (37 Steps)](#28-build-order)
29. [Repository Structure (Target)](#29-repository-structure)
30. [The AI Workflow (Read This If You Are an AI)](#30-the-ai-workflow)
31. [Partner Commitment](#31-partner-commitment)
32. [Glossary](#32-glossary)

---

## 1. Mission

**Build a free, downloadable, OS-native AI agent that lives on a user's computer and acts as the conductor of every other AI tool installed there. It opens with a custom global hotkey, listens for the wake word, clicks and scrolls through any app, and runs sub-agents across Claude Code, Codex, Antigravity, OpenCode, Kilo, Hermes, Ollama, and any ChatGPT Plus / Claude Pro / Gemini AI Pro / SuperGrok subscription the user already has — all while costing the user nothing.**

Mission statement (one line):

> Make AI useful to every ordinary person, on every computer, for free. No lock-in. No tracking. No compromise on speed or safety.

**Why this exists:** Hermes Agent, OpenClaw, Antigravity, Claude Code, Codex — all amazing, all locked behind either subscriptions, cloud platforms, or single-vendor stacks. None of them talk to each other. None of them give the user a single hotkey that does anything on the computer. Synaptic is the missing conductor. Free, fast, theirs.

---

## 2. The Survival Rule

Synaptic performs physical, often irreversible actions on the user's operating system. A fallible multi-model system. Async-supervised. Operating with stale screen state. **This is not an optimization problem. It is a survival problem.**

If a feature conflicts with the invariants below, **the feature is wrong. Remove the feature.**

### 2.1 The Seven Non-Negotiable Invariants

1. **The Strategist and the Gatekeeper are separate systems.** The Strategist is a model. The Gatekeeper is deterministic code. They are never the same.

2. **The Gatekeeper is the only path to physical action.** No model output flows to a click, type, or shell exec without passing the Gatekeeper.

3. **Destructive actions require a real human at the keyboard.** Native modal dialog. Blocks until clicked. No exceptions. No "trust me, the model said it's safe."

4. **The user can always stop the agent.** Hard hotkey, watchdog timer, network isolation, menu bar kill. Four independent mechanisms. The agent cannot disable any of them.

5. **Every action is auditable, in a tamper-resistant log.** HMAC-chained, append-only, never deleted. If something goes wrong, we can prove exactly what happened.

6. **The agent is a guest, not an owner.** It requests permission to enter rooms (apps, files, URLs). The user grants or denies. We never escalate, never bypass, never pretend.

7. **OS permissions are granted by the user, on their machine.** We don't have access. We ask, they grant. The onboarding flow makes this easy and clear.

### 2.2 Hard Constraints (Never Break)

1. **The user's API key is sacred.** Never log it, never send it anywhere except the configured LLM provider, never include it in telemetry.
2. **The user is always in control.** No action without either explicit consent or a pre-approved rule. Any new app or sensitive action prompts first.
3. **Speed is the product.** Cold start < 500ms. Hotkey → overlay < 100ms. First token < 1.5s. No exceptions.
4. **Local-first.** Memory, skills, audit log, embeddings — all on disk, encrypted. The only network calls are to the LLM provider(s) the user configured.
5. **Free forever.** No feature gates. No premium tier. No nags. A donate button in the menu bar, that's it.
6. **Proprietary source, free binary.** Repo is private. Binaries are signed, notarized, and downloadable from synaptic.app.

---

## 3. What Synaptic Is (and Is Not)

### Is
- A **free** desktop application (Mac, Windows, Linux).
- A **persistent** agent (24/7, lives in the menu bar / system tray).
- A **conductor** that orchestrates other AI tools installed on the user's machine.
- A **learner** that adapts to the user's behavior over time.
- A **guest** on the user's computer, always requiring consent for actions.
- **Source-available on request** (proprietary license, but free binary).

### Is Not
- A cloud service (the model runs via the user's own API key or local Ollama).
- An open-source project.
- A single-vendor tool (works with 12+ LLM providers and 8+ CLI tools).
- An autonomous weapon (every destructive action needs explicit consent).
- A subscription product.

---

## 4. The 26 Locked Decisions

Every decision made during planning. Nothing is open. Implementation may surface new questions, but the foundation is locked.

| # | Decision | Value | Source |
|---|---|---|---|
| 1 | Project name | **Synaptic** | User |
| 2 | License (binary) | **Synaptic Freeware EULA v1** (free personal + commercial, no redistribution, revocable for abuse) | User |
| 3 | License (source) | **Proprietary** (private repo, source available on request) | User |
| 4 | Foundation approach | **From scratch in Go + TypeScript** (no Hermes fork) | User |
| 5 | Computer use backends | **All 3 + vision CUA, with 4-tier router** (ORAX Eye → mac-cua → macOS-MCP → vision CUA) | User |
| 6 | Routing strategy | **Hybrid with memory** (cost-first cascade, bias toward what worked) | User |
| 7 | Plan depth | **Exhaustive detail** (every small detail, since AI is building it) | User |
| 8 | Hotkey | **User must set on first run** (no default; suggestions: Option+Option, Cmd+Shift+Space, Ctrl+Space) | User |
| 9 | Web app stack | **Next.js 14 on Vercel** at `synaptic.app` | User |
| 10 | Donation platform | **All three** — GitHub Sponsors + Open Collective + Stripe | User |
| 11 | Languages at v0.1.0 | **English + Spanish + French + German + Japanese + Mandarin** (i18n from day 1) | User |
| 12 | Visual brand | **Decide later** (placeholder palette, iterate after first UI mockup) | User |
| 13 | Launch strategy | **Public v0.1.0, all in** — Product Hunt + Hacker News + Reddit (r/singularity, r/LocalLLaMA, r/AI_Agents) on same day | User |
| 14 | Support channels | **All** — Discord + GitHub Issues + support@synaptic.app | User |
| 15 | Provider down behavior | **Auto-failover** — Ollama local first, then any configured backup key | User |
| 16 | Multi-machine sync | **P2P encrypted sync** (device-to-device, E2E encrypted, no central server) | User |
| 17 | Uninstall behavior | **Auto-backup before uninstall** to `~/Documents/synaptic-backups/` | User |
| 18 | Skill sharing | **Public Skills Hub** at `hub.synaptic.app` (curated, safety-scanned, versioned) | User |
| 19 | Feedback UX | **Thumbs up/down** on every response (optional text; feeds adaptive engine) | User |
| 20 | Persona | **Adaptive, mirrors user** (no fixed character; learns communication style) | User |
| 21 | Sensitive data handling | **Warn + ask each time** (banking/health portals: native dialog before any action) | User |
| 22 | Compromised key | **All three** — auto-detect spend spikes + manual reporting + configurable spend limits | User |
| 23 | Concurrency | **Conservative** — default 2 parallel sub-agents, max 5, user-configurable | User |
| 24 | Autonomy | **Default cautious** — warn before any action, user opts into autonomous per-field | User |
| 25 | Uncertainty handling | **Ask user immediately** — overlay shows "I'm 60% sure you want X. Proceed?" | User |
| 26 | Energy budget | **Refuse, force user decision** — when budget hit and vision needed, pause and ask | User |
| 27 | Daemon autostart | **Auto-start on login** (LaunchAgent / Run key / systemd user) | User |
| 28 | Backup destination | `~/Documents/synaptic-backups/synaptic-backup-<date>.zip` | User |
| 29 | Cloud sync infra | **P2P sync** (Syncthing-style, no central server, E2E encrypted) | User |
| 30 | User account | **Email + magic link** (for hub, donations, support; sync is P2P, no account needed) | User |
| 31 | Action replay | **Yes, built-in** — last 24h scrubbable with screenshots + decisions | User |
| 32 | Versioning | **v0.1.0** (SemVer) | User |
| 33 | Web dashboard auth | Same magic-link as desktop app | User |
| 34 | Multi-install | **Block second install** (one stable instance per machine) | User |
| 35 | Wake word | **"hey synaptic"** (custom, local, openWakeWord) | User |
| 36 | EULA clauses | **Freeware EULA** — free personal + commercial, no redistribution, revocable for abuse | User |

> Note: The original count was "26 decisions" but additional small decisions were made during finalization. All are listed here for completeness.

---

## 5. The 7 Technical Non-Negotiables

These are the survival requirements that came out of the security analysis. **Every one is implemented. No exceptions.**

### 5.1 Action Classification by Blast Radius
Every action is classified before execution:
- **READ** — screenshot, copy text, inspect file. Low risk.
- **WRITE** — edit file, type text, paste content. Medium risk; verify target first.
- **NETWORK** — click link, submit form, send message. High risk; require approval.
- **DESTRUCTIVE** — delete, format, transfer, purchase, authorize. CRITICAL; human gate mandatory.

Critical actions: not a Telegram reply — a **native macOS dialog** that halts execution until the human physically clicks "Allow" on their actual machine. If agent is running remotely and user is away, those actions **queue, do not execute**.

### 5.2 Mandatory Pre-Action Verification (Twin Snapshots)
For every WRITE or NETWORK action:
1. Capture the Accessibility tree immediately before acting.
2. The agent must articulate exactly what it thinks it is about to click: "I see a button labeled 'Send Email' in window 'Gmail'."
3. Compare against a second snapshot taken milliseconds later.
4. If the tree changed between planning and execution, **abort**.

This is the anti-staleness mechanism. Without it, the agent plays darts with the OS.

### 5.3 The Kill Switch (3 Layers)
Independent of the agent process:
- **Layer 1**: Hard hotkey (Cmd+Shift+Escape on mac, Ctrl+Alt+Del on win, configurable). Kills the process instantly.
- **Layer 2**: Watchdog timer. If N seconds without verification, auto-pause.
- **Layer 3**: Network isolation toggle. A separate OS process owns a `pf` (mac) / `netsh` (win) rule blocking all egress from the user's UID except the LLM provider IPs. The agent process **cannot** stop it.

### 5.4 Audit Log of Everything
HMAC-chained, append-only, never deleted. Every screenshot, every AX tree dump, every model decision, every API call, every click coordinate. Not for debugging — for forensics. When the agent sends an email the user didn't authorize, the user can prove exactly which model and which prompt caused it.

### 5.5 Model Isolation, Not Just Switching
The Manager routes between models. But if Claude generates a Python script and Ollama executes it, Ollama must not have implicit context about what Claude intended. Each handoff must be explicit and sanitized. If a model outputs a shell command, it does not run. It is parsed, validated against an allowlist, and only then passed to a sandboxed executor.

**Never let Model A's output flow directly into Model B's execution context without a deterministic validation layer in between.**

### 5.6 The "Agent Went Insane" Detector (Behavioral Anomaly)
A cost guardrail is not enough. The behavioral anomaly detector fires when:
- Agent takes actions faster than a human could.
- Agent clicks the same coordinates repeatedly (stuck loop).
- Agent sends to network endpoints it has never used before.
- Agent runs >30 minutes on a single task without a verification checkpoint.

If any trigger fires, the agent **hard pauses** and pings the user. **Auto-recovery is the enemy.** A stuck agent is annoying. An un-stuck agent running unsupervised is dangerous.

### 5.7 The Strategist vs Gatekeeper Separation
**Never let an AI model decide, in a single turn, both WHAT to do and WHETHER it is safe to do unsupervised.**

Those must be separate systems:
- **The Strategist** (any model): "We need to click 'Submit' to send the email."
- **The Gatekeeper** (a small, local, **deterministic** rules engine): "This is a NETWORK action on a messaging app. The user is not present. Policy says: queue for approval."

If the Strategist and the Gatekeeper are the same model, or if the Gatekeeper is probabilistic, **there is no security. There is a suggestion box.**

---

## 6. Selective Perception

A unified system that delivers **safety + battery efficiency + performance + reliability** simultaneously. The system has one name and one mental model.

> **Selective Perception** — the agent perceives only what it must, when it must, and acts only after verifying what it perceived. Every perception has a purpose, a TTL, and a verification step. The cost of perception is amortized across decisions, and no perception is wasted on a decision that gets aborted.

### 6.1 The Problem It Solves
Naive screen recording is a battery vampire:
- **Idle screen**: ~0.3W
- **Single screenshot**: 0.5–0.8W spike
- **Continuous recording at 1fps**: 2–4W sustained → kills battery in 2-3h
- **30fps recording**: 5–8W → kills battery in 90 minutes
- **Vision API call per frame**: +$0.01-0.03 per call

An agent that screenshots-on-every-action **gets uninstalled** because the laptop dies. This is a make-or-break product issue.

### 6.2 The 5-Stage Pipeline
1. **Classify** (blast radius) — cost: ~0.1ms, 0 energy
2. **Capture** (battery-aware strategy) — cost varies from 0 (cache hit) to high (vision CUA)
3. **Verify** (twin snapshot) — anti-staleness
4. **Gatekeeper** (deterministic rules engine) — must allow
5. **Anomaly check** (post-action) — feeds back into the loop

### 6.3 Battery Magic (SmartCapturer strategies, cheapest first)
| Strategy | When | Battery | Latency |
|---|---|---|---|
| **None** | READ action, no state needed | 0 | 0 |
| **AX-only** (tree, no pixels) | Element is named | **~10× cheaper** | 50ms |
| **Window-rect** (one window/rect) | Rect is known | **~5× cheaper** | 80ms |
| **Differential** (only changed rect) | Dirty flag set | **~3× cheaper** | 100ms |
| **Full screen** | Last resort before vision | 1× baseline | 200ms |
| **Vision CUA** | Vision model | **~50× cost** + LLM | 2-3s |

### 6.4 Energy Budget
- `Low` (battery, no charger): 20% of session energy, AX-only.
- `Balanced` (default): 50% of session, allow window-rect.
- `High` (plugged in): 100%, allow vision CUA.
- `Auto`: detect power state, adjust.

When budget is hit, **pause and ask** (see Decision 26: refuse, force user decision).

### 6.5 Dirty Tracking (event-driven, no polling)
- macOS: `CGEventTap` + `NSWindowDidUpdateNotification` + AX notifications
- Windows: `EVENT_OBJECT_LOCATIONCHANGE` + `EVENT_OBJECT_NAMECHANGE`
- Linux: AT-SPI's `object:state-changed:defunct` signals

When the user is typing, **the agent is asleep**. When the user clicks, the agent wakes for one snapshot, verifies, and goes back to sleep. **Battery is only consumed when there's actual change to perceive.**

---

## 7. Architecture Overview

```
┌──────────────────────────────────────────────────────────────────┐
│                    USER INTERFACES                               │
│  Overlay (Wails) · TUI (Ink) · Web Dashboard · Voice · Menu Bar  │
└──────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│              ADAPTIVE ENGINE (the learner)                      │
│  User Model · Dialectic · Predictor · Feedback · Visibility      │
└──────────────────────────────────────────────────────────────────┘
                              │
                              ▼
┌──────────────────────────────────────────────────────────────────┐
│              CORE ORCHESTRATOR (the conductor)                   │
│  Planner · Router (hybrid w/ memory) · Agent Loop               │
└──────────────────────────────────────────────────────────────────┘
                              │
        ┌─────────────────────┼─────────────────────┐
        ▼                     ▼                     ▼
┌──────────────┐   ┌──────────────────┐   ┌────────────────────┐
│  SELECTIVE   │   │  SAFETY LAYER    │   │   DELEGATION BUS   │
│  PERCEPTION  │   │  (deterministic) │   │   (sub-agents)     │
│              │   │                  │   │                    │
│  Smart cap.  │   │  Blast radius    │   │  8 CLIs + Ollama   │
│  Twin snap.  │   │  Gatekeeper      │   │  Model isolation   │
│  Energy bud. │   │  Kill switch     │   │  Wave scheduling   │
│  Dirty track │   │  Anomaly det.    │   │  File coord        │
│              │   │  Audit (HMAC)    │   │                    │
│  Computer    │   │  Sensitive det.  │   │                    │
│  Use Router  │   │  Spend monitor   │   │                    │
└──────────────┘   └──────────────────┘   └────────────────────┘
        │                     │                     │
        ▼                     ▼                     ▼
┌──────────────────────────────────────────────────────────────────┐
│                    EXECUTION LAYER                               │
│  mac-cua · macOS-MCP · ORAX Eye · vision CUA                    │
│  Claude Code · Codex · Antigravity · OpenCode · Kilo · Hermes   │
│  Ollama · Local shells (sandboxed)                              │
└──────────────────────────────────────────────────────────────────┘
```

---

## 8. Tech Stack (Locked)

| Concern | Choice | Why |
|---|---|---|
| Core daemon | **Go 1.22+** | Single binary, fast startup, CGO for macOS |
| Desktop shell | **Wails v2** (Go + web) | Reuses Go daemon, ~10MB vs Electron's 100MB+ |
| UI framework | **React 18 + Vite** inside Wails, **plus** Ink TUI | Web for overlay (rich, animated), TUI for SSH/power users |
| IPC | **Unix socket + HTTP + Wails runtime bridge** | Wails binds Go methods directly to JS |
| TUI | **TypeScript + Ink (React)** | Hermes-proven, rich streaming UX |
| Web app | **Next.js 14 + Tailwind + Vercel** | Fast, free tier, perfect for marketing |
| Computer use backends | **Existing Python libs (subprocess)** | mac-cua, macOS-MCP, ORAX Eye — don't reinvent |
| Storage | **SQLite + FTS5 + sqlite-vec** | Local-first, single file, no deps |
| Embeddings | **all-MiniLM-L6-v2** via local ONNX or Ollama | 384-dim, fast, ~30MB |
| LLM SDK | **TypeScript** primary (Anthropic, OpenAI, Google) + Go (for fast paths) | TS SDKs are most mature |
| IPC protocol | **JSON-RPC 2.0 over Unix socket** | Standard, debuggable |
| macOS native | **CGO + ApplicationServices framework** | For things Python libs don't cover |
| Global hotkey | **`github.com/atotto/carbon`** (mac) + **`golang.design/x/hotkey`** (cross-platform) | Carbon is most reliable on mac |
| Overlay window | Wails frameless, always-on-top, transparent | Native feel |
| Voice STT | **whisper.cpp** (local) + OpenAI Whisper (cloud fallback) | Local = no cloud cost, fast, private |
| Voice TTS | OpenAI + ElevenLabs + native `say` (mac) | Multiple providers |
| Wake word | **openWakeWord** (local, custom phrase "hey synaptic") | Open source, runs offline |
| Auto-update | **`go-update` + Sparkle-like delta** (mac) / Squirrel (win) | Standard pattern, signed |
| Code signing | Apple Developer ID + `codesign --deep --strict --options=runtime`; Microsoft Authenticode | Required for distribution |
| Notarization | `notarytool` + altool for macOS | Required for Gatekeeper |
| Installers | **GoReleaser** (macOS dmg/pkg, Windows msi/exe, Linux deb/AppImage) | One config, all OSes |
| P2P sync | **Custom Kademlia DHT + Noise XX handshake** OR Syncthing-fork | E2E encrypted, no server |
| License (binary) | **Synaptic Freeware EULA v1** | Custom, free, revocable |
| License (source) | **Proprietary** | Private repo |
| Build | **GoReleaser** (Go) + **tsup** (TS) + **npm** + **Homebrew tap** | Standard |
| Package | **GitHub Releases** + **Homebrew** + direct download | Standard |
| Repo | **github.com/sahajpatel123/synapticapp** (private) | Standard |

**What we DON'T use:** Python in our main code. Python is *only* a substrate for the 3 computer-use MCPs we don't rewrite.

---

## 9. The User-Adaptive Engine

The crown jewel. The thing that makes users come back every day and feel "this is *my* agent". A named, first-class module, not a side-effect of memory.

### 9.1 Closed Learning Loop
1. **Observer** watches every interaction (locally, no telemetry)
2. **Dialectic** argues about what user actually meant (proposer + critic + adjudicator)
3. **Predictor** uses the model to suggest next steps
4. **Visibility** lets user audit and edit
5. **Forget** lets user remove anything
6. **Strength** slider (off / cautious / balanced / aggressive)

### 9.2 The User Model (Honcho-style structured)
```go
type UserModel struct {
    Identity       Identity
    Preferences    Preferences
    Style          Style
    Workflows      []WorkflowPattern
    Expertise      map[string]int    // topic → confidence (0-1)
    PetPeeves      []string
    TimePatterns   TimePatterns
    ToolsHabits    map[string]int
    ModelPrefs     map[string]string
    RiskTolerance  string             // cautious | balanced | aggressive
    Communication  Communication
    LastUpdated    time.Time
    Version        int
}
```

### 9.3 The Dialectic
- **Proposer** (primary LLM): "What does this tell us about the user?"
- **Critic** (often a cheaper model): "Is this over-fitting? Is the evidence strong?"
- **Adjudicator** (deterministic merge): "Apply if confidence > 0.6, else discard."

### 9.4 The Predictor
- Time-of-day patterns
- Sequence patterns (after Y, user usually does Z)
- Tool preferences per task type
- Style mirror (apply at prompt-building time)

### 9.5 Visibility (the user can SEE and EDIT)
- "What Synaptic has learned about you" view in Settings
- Every inferred preference with evidence
- One-click delete for any item
- Dialectic log: recent arguments the model had about you
- Strength slider: Off / Cautious / Balanced / Aggressive
- Export everything / Delete everything and start fresh

### 9.6 Adaptive Strength Setting
```yaml
adaptive:
  enabled: true
  strength: balanced
  schedule_review_reminders: weekly
  auto_apply_only:
    - verbosity
    - response_length
    - default_model
    - time_patterns
  require_confirm_for:
    - adding_new_skill
    - changing_default_backend
    - any_communication_style_change
  forget_after_days: 365
  dialectic:
    primary_model: primary
    critic_model: routing
    min_confidence_to_apply: 0.6
```

---

## 10. The Safety Layer

> ## ⚠️ BUILD STATUS & RELEASE GATE — read this first
>
> **The Safety Layer is the spine of the Survival Rule (§2).** It was *specified*
> first, but in practice it was **built only partially** while agent capability
> raced ahead. Closing that gap is now the **highest-priority debt in the
> project.** Per §2, a feature that reaches a user without the armor under it is
> *the wrong feature.*
>
> **Hard gate: no public `v0.1.0` binary ships until every module below is built,
> tested with `-race`, and lint-clean.** Treat the missing modules as a dedicated
> milestone — **"The Armor"** — and finish it before distribution.
>
> | # | Module | Status | Package |
> |---|---|---|---|
> | 10.1 | Blast Radius Classifier | ✅ built | `internal/blastradius` |
> | 10.2 | Gatekeeper **policy engine** (rules · consent · native modal · queue) | ❌ **v0 STUB** — `DenyBeyondRead` denies *every* non-READ action at runtime | `internal/gatekeeper` |
> | 10.3 | Kill Switch (3 layers) | ✅ built | `internal/halt` + hotkey |
> | 10.4 | **Behavioral Anomaly Detector** | ❌ **MISSING** | — |
> | 10.5 | Audit Log (HMAC-chained) | ✅ built | `internal/audit` |
> | 10.6 | **Model Isolation / Sanitizers** | ❌ **MISSING** | — |
> | 10.7 | **Sensitive Site Detector** | ❌ **MISSING** | — |
> | 10.8 | Spend Monitor | ✅ built | `internal/failover` |
> | 10.9 | **Autonomy Matrix** | ◐ config-only, **no engine** | `internal/config` |
>
> **Owed: the real 10.2 engine, plus 10.4, 10.6, 10.7, and the 10.9 engine.**
> Until 10.2 is real, **every non-READ action is denied at runtime** — the
> computer-use and MCP systems already "built" are gated to READ-only, and a
> future sub-agent spawn (a non-READ action) cannot execute. So the Armor is not
> just hardening; it is the **functional unblock for the entire agent**, and it
> must be built **before** the Conductor (delegation). These are not polish —
> they are the difference between "a powerful agent" and "an agent safe enough to
> hand to an ordinary person." Every one of them gates the first release.

5 modules, all critical. **Specified before any agent logic; must be _completed_
before any public binary ships (see the Build Status gate above).**

### 10.1 Blast Radius Classifier
Classifies every action: READ / WRITE / NETWORK / DESTRUCTIVE.

### 10.2 The Gatekeeper (deterministic rules engine, NOT an LLM)
Pure-rules, no-neural-net. Cannot be prompt-injected, cannot hallucinate.

User-editable policy in `~/.synaptic/policy.yaml`:
```yaml
rules:
  - match: { class: READ }
    decide: allow
  - match: { class: WRITE, target_app: ["Code", "VS Code", "Cursor", "Terminal"] }
    decide: allow
  - match: { class: NETWORK }
    decide: require_consent
    consent:
      type: native_dialog
      timeout_seconds: 300
      on_timeout: queue
  - match: { class: DESTRUCTIVE }
    decide: require_presence_and_consent
    consent:
      type: native_dialog
      require_user_active: true
      on_user_absent: queue
  - match: { target_app: ["1Password", "Keychain Access"] }
    decide: deny
```

**Default-deny**: if no rule matches, ask.

### 10.3 Kill Switch (3 layers)
- **Layer 1**: Hard hotkey (Cmd+Shift+Escape default)
- **Layer 2**: Watchdog timer
- **Layer 3**: Network isolation (separate process agent can't control)

### 10.4 Behavioral Anomaly Detector
- Speed: >20 actions/minute → pause
- Loop: same coordinates 3+ times → halt
- Duration: >30 min on one task → pause
- Failures: 5+ consecutive errors → pause
- New endpoint: never-used network target → warn

### 10.5 Audit Log (HMAC-chained)
- Append-only, never deleted
- HMAC-SHA256 chain (each entry includes hash of previous)
- 90-day retention (configurable)
- Secret redaction
- Forensically sound

### 10.6 Model Isolation (sanitizers)
- Shell command sanitizer: allowlist of binaries, arg pattern checks
- Python script sanitizer: AST parse, banned imports check
- File path sanitizer: no `..`, no system paths
- URL sanitizer: SSRF blocklist
- Message body sanitizer: PII detection

### 10.7 Sensitive Site Detector
- Domain allowlist (banking, health)
- Heuristic detection (form labels like "credit card", "SSN")
- User overrides

### 10.8 Spend Monitor
- Periodic check of LLM provider dashboards
- Alert on unusual spend
- Configurable per-provider hard limits

### 10.9 Autonomy Matrix
Per-task-type + per-app autonomy level. Default: warn. User can dial up to autonomous or down to block.

---

## 11. The Computer-Use System

### 11.1 Backends (4 tiers, cheapest first)
1. **ORAX Eye** — structured AX tree, free, fast (~50ms). MIT.
2. **mac-cua** — background-first, `CGEventPostToPid`, agent works without taking focus. Apache 2.0.
3. **macOS-MCP** — comprehensive, foreground interaction. MIT.
4. **Vision CUA** — Anthropic Computer Use or OpenAI CUA, ~$0.02-0.05/action. Last resort.

### 11.2 The 4-Tier Router
1. Try ORAX Eye first
2. Fall back to mac-cua
3. Fall back to macOS-MCP
4. Last resort: vision CUA

### 11.3 Smart Capturer (battery-aware)
See Section 6.

### 11.4 Twin-Snapshot Verification
Pre-action, two snapshots, abort on diff.

### 11.5 User Interruption Detection
CGEventTap watches for user keyboard/mouse activity. If user starts interacting with the app the agent is working in, agent yields.

### 11.6 Platform Support
- **macOS**: Full support (primary)
- **Windows**: xa11y-based, via UI Automation
- **Linux**: AT-SPI2, with Wayland portal support

### 11.7 TCC Permission Tiers
- **read_only** — attribute reads only
- **standard** — most actions, blocked from sensitive apps
- **elevated** — full access, user must explicitly enable per session

---

## 12. The Router (Hybrid with Memory)

### 12.1 Strategies
- **cascade** — try cheap first, escalate on failure (default initial)
- **pareto** — cheapest model above a quality threshold
- **hybrid** — cascade + memory bias (default)
- **user** — user picks per task

### 12.2 Hybrid Algorithm
```
score = w.cascade * cascadePosition
      + w.quality * qualityEstimate
      + w.cost    * (1 - normalizedCost)
      + w.memory  * memoryBias        // bias toward what worked
      + w.latency * latencyEstimate
```

### 12.3 Memory Bias
After N samples per task type per backend, bias toward the success rate. Requires `min_samples_for_bias` to activate.

### 12.4 Decision Logging
Every routing decision written to `routing_decisions` table with candidates, chosen, reason, cost, latency, success. This is what makes the memory bias actually work.

### 12.5 User Priority Override
Users can set their own priority per task type, e.g.:
```yaml
router:
  priority:
    coding:
      - backend: claude_code
        model: claude-sonnet-4-5
        reason: "Best for my coding tasks"
      - backend: opencode
        model: anthropic/claude-sonnet-4-5
        reason: "Backup"
    chat:
      - backend: codex        # uses ChatGPT Plus via OAuth
        model: gpt-5.5
```

User priority is the strongest signal. It overrides memory unless backend is down or offline.

---

## 13. The Delegation Bus

### 13.1 LLM Providers (12, with auth options)

| Provider | Auth | Models |
|---|---|---|
| Anthropic | API key OR Claude Pro OAuth | Claude Opus 4.7, Sonnet 4.5, Haiku 4.5 |
| OpenAI | API key OR ChatGPT Plus/Pro OAuth | GPT-5.5, GPT-5.5-codex, o3, o4-mini, gpt-image-2 |
| Google | API key OR Google AI Pro/Ultra OAuth | Gemini 3.5 Flash, 3.1 Pro, 2.5 Pro |
| xAI | API key OR SuperGrok OAuth | Grok-4.3, Grok-4.3-fast |
| Mistral | API key | Mistral Large 3, Codestral, Pixtral Large |
| DeepSeek | API key | DeepSeek-V4, R1 |
| OpenRouter | API key | 300+ models (Pareto router) |
| Together | API key | Llama, Qwen, DeepSeek, Mixtral |
| Groq | API key | Llama 4, Mixtral, Whisper |
| Fireworks | API key | Llama, Qwen, DeepSeek |
| Custom OpenAI-compatible | API key + base URL | Any user-specified |
| **Local (Ollama / LM Studio / vLLM / llama.cpp)** | None | User's local models |

### 13.2 Sub-Agent CLIs (8)
| CLI | Spawn command |
|---|---|
| Claude Code | `claude --print --output-format stream-json --model <m>` |
| Codex | `codex --json --model <m>` |
| Antigravity | `agy --output-format json --model <m>` |
| OpenCode | `opencode --format json` |
| Kilo Code | `kilo --json` |
| Hermes Agent | `hermes --format json` |
| Gemini CLI (legacy) | `gemini --output-format json` |
| Ollama | direct HTTP, no subprocess |

All auto-discovered in `$PATH`. Friendly "Install X? [Docs]" prompt if missing.

### 13.3 Code-Execution Delegation (CE-MCP)
Instead of round-tripping tool calls through LLM context, generate a single script that orchestrates multiple CLIs, computer-use actions, and APIs in one shot. **~70% token reduction** per MCP-Bench research.

### 13.4 Per-Task Concurrency
Default 2 parallel sub-agents, max 5, per-backend semaphores. User-configurable.

### 13.5 File Coordination
Parallel siblings use SQLite as the lock (not filesystem — race conditions). Shared scratch space in `~/.synaptic/scratch/`.

### 13.6 Execution Waves
Decompose task into DAG. Tasks within a wave run in parallel. Later waves wait for earlier.

### 13.7 Heartbeat + Supervisor
Watchdog detects stalled tasks. Auto-retry with fingerprinting to avoid infinite loops.

---

## 14. Memory System (3 Layers)

### 14.1 Episodic
Past sessions, indexed by FTS5 + vector embeddings. Recall = top-k by combined score. Summarized via LLM.

### 14.2 Semantic
Facts about the user (preferences, identity, expertise). Extracted from interactions via dialectic. Confidence-scored.

### 14.3 Procedural
Skills (separate package). Auto-created, self-improved.

### 14.4 Storage
- SQLite + FTS5 for full-text
- sqlite-vec for vector similarity
- All on disk, encrypted at rest

---

## 15. Skills System

### 15.1 Format
`agentskills.io` compatible — portable, shareable, community-contributed.

### 15.2 Lifecycle
1. **Created** after a complex task is solved (auto)
2. **Improved** after N uses (auto, with user consent)
3. **Bundled** (e.g., `/writing-day` loads 4 skills at once)
4. **Shared** via Skills Hub
5. **Scanned** for promptware on import

### 15.3 Skills Hub (`hub.synaptic.app`)
- Public, curated, safety-scanned
- Versioned (semver)
- User can subscribe to updates
- Trust levels: official, community, experimental

---

## 16. MCP Gateway

### 16.1 Architecture
- 10,000+ existing MCP servers, consumable via stdio / HTTP / SSE
- Prefix routing: `mcp__<server>__<tool>` to avoid collisions
- Tool-search: lazy-load definitions to keep context small
- Per-server OAuth, credentials in server env, never in prompt

### 16.2 Custom Servers
Users can author their own. We ship a Go SDK + TS SDK.

### 16.3 Curated Catalog
"Approved by Synaptic" catalog with one-click install. Mirrors the optional-skills pattern from Hermes.

---

## 17. P2P Sync

### 17.1 Why P2P
No central server. No account required for sync. Maximum privacy. Works on LAN and over the internet (via NAT traversal + relay fallback).

### 17.2 Architecture
- **Device identity**: Ed25519 keypair generated on first run
- **Discovery**: mDNS / Bonjour on LAN
- **Transport**: Custom Kademlia DHT or Syncthing-fork
- **Protocol**: Noise XX handshake, encrypted streams
- **Conflict resolution**: CRDT (Yjs-style) for memory + skills
- **Pairing**: QR code or 12-char code
- **What's synced**: memory + skills + config
- **What's NOT synced**: logs, audit, screenshots, API keys (always local)

### 17.3 Pairing Flow
Settings → "Pair a new device" → QR code with public key + LAN IP + pairing token → other device scans → both confirm → sync starts.

### 17.4 Revocation
Any paired device can revoke any other. Revoked device forgets the encryption keys.

### 17.5 Synaptic Account (separate from P2P)
For Skills Hub, donations, support. Optional. P2P sync works without it.

---

## 18. Action Replay

### 18.1 What
Last 24h of agent actions, scrubbable timeline with screenshots and decisions.

### 18.2 Why
Transparency. User can see exactly what the agent did and why. Forensics if something goes wrong.

### 18.3 Storage
~50MB per day. H.264 delta compression for screenshots. Auto-prune at 24h.

### 18.4 UI
- Timeline at bottom
- Click any moment → see screenshot + decision + model output
- Export to .mp4 for sharing

---

## 19. Global Hotkey + Overlay + Voice

### 19.1 Global Hotkey
- User picks on first run (no default; suggestions: Option+Option, Cmd+Shift+Space, Ctrl+Space, Ctrl+Ctrl)
- Multi-OS implementation
- Recurring tap detection (Option+Option = 2 taps in 300ms)

### 19.2 The Overlay
- Floating window, always-on-top, transparent
- Vibrancy (macOS) / Acrylic (Windows) backdrop
- Appears at cursor position or center
- 200ms slide-up + fade entrance
- Auto-dismiss after 5s inactivity, or pin
- `Esc` to dismiss, `Cmd+Enter` to submit, `Cmd+K` for command palette

### 19.3 Voice
- **STT**: whisper.cpp local (default), OpenAI Whisper (cloud fallback)
- **TTS**: OpenAI, ElevenLabs, native `say` (mac)
- **Wake word**: "hey synaptic" (custom, local, openWakeWord)
- **Push-to-talk** OR **continuous** (configurable)
- **Live transcription** while speaking
- **Submit on silence** (1.5s default)

### 19.4 Voice Orb UI
Animated waveform when listening. Pulsing dots. Color reflects confidence.

---

## 20. Onboarding Flow

> **Phase 14A update (converged "value-first" flow).** The original
> 7-screen plan (below, "Historical spec") put a login/power-source
> wall and a voice test in front of the user before they ever saw the
> agent work. That maximizes drop-off. The shipped flow is **4 screens,
> ≤9 clicks**, legal-first, value-first. The deferred screens move to
> Settings (progressive disclosure), not the critical path.

**Shipped flow (4 screens):**

1. **EULA** — Welcome copy + scrollable license; "I Accept" is disabled
   until the user scrolls to the bottom **and** ticks the checkbox.
   Legal consent happens **before** any system access. Acceptance is
   recorded with the EULA version so a future bump forces re-accept.
2. **Permissions** — only the two grants computer use actually needs:
   **Accessibility** + **Screen Recording**. Live status badges poll
   `permissions.status` every 2s; "Open System Settings" uses the
   per-platform deep link from `permissions.request_guide`. A **"Skip
   for now"** footer lets the user proceed; Continue is always enabled.
   Microphone / Automation / Notifications are requested lazily from
   Settings when the user enables those features.
3. **Hotkey** — the user records a combo (no silent default, per locked
   decision #8). Continue is enabled only once a valid combo is captured.
4. **Ready** — `onboarding.probe_power` auto-detects local Ollama (and
   installed CLIs) so the agent works immediately with **no account and
   no API key**. Optional cards ("Add an API key", "Connect messaging")
   deep-link into Settings. "Start using Synaptic" calls
   `onboarding.finish`, which persists the hotkey + EULA version, enables
   Ollama when reachable, writes the first-run marker, and dismisses the
   wizard.

**Architecture:** the daemon's `internal/onboarding` state machine is the
single source of truth (`eula → permissions → hotkey → complete`); the
Svelte wizard (`OnboardingWizard.svelte` + `lib/components/onboarding/*`)
renders the current step over the `onboarding.*` RPCs and never keeps a
parallel step list. Legacy 8-step persisted state is migrated forward on
load. **No account is required to use the local agent** (locked decision
#30: account is for Hub/dashboard/support only; sync is P2P). Settings
exposes a **Legal** section (view EULA) and **Re-run setup**.

After onboarding: menu bar icon (mac) / system tray (win) / status icon (linux) shows status.

<details>
<summary><strong>Historical spec (original 7-screen plan, superseded by the 4-screen flow above)</strong></summary>

1. **Welcome** — what Synaptic is, mission
2. **EULA acceptance** — must accept
3. **Power source** — choose: connect subscription (ChatGPT Plus, Claude Pro, Gemini AI Pro, SuperGrok) OR paste API key OR use local model OR mix
4. **Permission grants** — Accessibility, Screen Recording, Microphone, Notifications (macOS); equivalents on Windows/Linux
5. **Backend detection** — auto-detect installed CLIs, ask which to enable
6. **Hotkey configuration** — record the key combo user wants
7. **Voice test** — "Say something"

Rationale for change: power source, backend detection, and voice test are
all discoverable in Settings after first value. Forcing them up front (and
especially any login) costs users before they see the agent do anything.
</details>

---

## 21. Interfaces

### 21.1 TUI (Ink/React)
- Multiline composer
- Slash command autocomplete
- Conversation history
- Interrupt-and-redirect (Ctrl+C)
- Streaming tool output
- Status bar (per-turn stopwatch, git branch, token usage, cost)

### 21.2 Wails Desktop App
- Main window: full chat + sidebar with sessions, memory, skills, tasks, settings, audit log
- Overlay window: floating chat/voice
- Menu bar app: status, quick actions
- System tray on Windows
- Status icon on Linux

### 21.3 Web Dashboard
- At `synaptic.app/dashboard` (or local `localhost:7475`)
- Same auth as desktop (magic link)
- Mobile-responsive
- Real-time session tracking
- i18n (6 languages)

---

## 22. Distribution & Updates

### 22.1 Installers (per OS)
- **macOS**: `.dmg` (drag-to-Applications) and `.pkg`
- **Windows**: `.exe` (NSIS) and `.msi` (WiX)
- **Linux**: `.deb`, `.rpm`, `.AppImage`

### 22.2 Code Signing
- **macOS**: Apple Developer ID, hardened runtime, notarized via `notarytool`, stapled
- **Windows**: Authenticode (EV cert), SHA256 verified
- **Linux**: GPG-signed

### 22.3 Auto-Update
- Channels: stable / beta / dev
- `go-update` with delta patches
- Ed25519 signature verification
- Atomic rollback on failure

### 22.4 Update Server
GitHub Releases. Synaptic checks every 6h and on launch.

### 22.5 Block Second Install
One stable instance per machine. Second install blocked with friendly message.

---

## 23. i18n (6 Languages at v0.1.0)

- **English** (en) — default
- **Spanish** (es)
- **French** (fr)
- **German** (de)
- **Japanese** (ja)
- **Mandarin** (zh)

### 23.1 Implementation
- **Go**: `golang.org/x/text/message` for backend messages
- **TypeScript**: `i18next` + `react-i18next` for UI
- **JSON catalogs** in `app/frontend/src/locales/<lang>.json`
- **Crowdin** for community translations
- **LLM** responds in user's language regardless of UI language

---

## 24. Cloud Backups + Uninstall

### 24.1 Auto-Backup on Uninstall
- Triggered by OS uninstaller
- Zips: memory + skills + config + sanitized audit
- Saves to `~/Documents/synaptic-backups/synaptic-backup-<ISO-date>.zip`
- Then proceeds with uninstall

### 24.2 Manual Backup
Settings → "Export everything" → choose destination.

### 24.3 P2P Backup Alternative
For users who don't want local files: a paired device can be the backup target.

---

## 25. Spend Monitor + Failover

### 25.1 Spend Monitor
- Periodic check of LLM provider dashboards (OAuth)
- Alert on unusual spend via native notification + email
- Configurable per-provider hard limits (e.g., "stop at $50/day on OpenAI")

### 25.2 Failover
When primary provider is down or key is rejected:
1. Try Ollama local first
2. Then any configured backup key
3. Notify user
4. Queue tasks

---

## 26. Support + Donations + Marketing

### 26.1 Support
- **Discord**: `discord.gg/synaptic` (community + dev)
- **GitHub Issues**: bug reports + feature requests
- **Email**: `support@synaptic.app`
- **Docs site**: `synaptic.app/docs`

### 26.2 Donations
- **GitHub Sponsors**: recurring
- **Open Collective**: recurring + transparent ledger
- **Stripe one-time**: via `synaptic.app/donate`

### 26.3 Marketing
- `synaptic.app` — landing, demo GIF/video, download, donate, changelog
- `hub.synaptic.app` — Skills Hub
- Launch on Product Hunt + Hacker News + Reddit (r/singularity, r/LocalLLaMA, r/AI_Agents)
- Demo video: 60s showing overlay → voice → task done
- Screenshots of overlay, main window, settings, on each OS

---

## 27. The Autonomy Matrix

The user-defining setting. Default: **all warn (yellow)**. User dials each cell to **autonomous (green)** or **block (red)**.

```yaml
# ~/.synaptic/autonomy.yaml
autonomy:
  task_types:
    coding:           warn
    file_operations:  warn
    web_browsing:     warn
    email:            warn
    calendar:         warn
    messaging:        warn
    shell_commands:   warn
    computer_use:     warn
    research:         autonomous
    image_generation: autonomous
    code_review:      autonomous
  apps:
    com.apple.Mail:       warn
    com.tinyspeck.chatly: warn
    com.google.Chrome:    autonomous
    com.apple.finder:     autonomous
    com.microsoft.VSCode: autonomous
    com.banking.app:      block
  global_default: warn
  show_warnings_for_read: false
  max_consecutive_warns_before_asking_anyway: 5
```

Gatekeeper reads this matrix on every action.

---

## 28. Build Order (37 Steps, ~91 Working Days)

| # | Step | Days |
|---|---|---|
| 1 | Bootstrap (Makefile, go.work, CI, lint, .goreleaser) | 1 |
| 2 | Storage (SQLite + FTS5 + vec, migrations, crypto) | 2 |
| 3 | Config (YAML loader, schema, validation) | 1 |
| 4 | IPC (JSON-RPC 2.0 + WebSocket + auth) | 2 |
| 5 | LLM clients (12 providers + streaming + caching) | 2 |
| 6 | API key manager (encrypted, validated, OAuth) | 1 |
| 7 | Provider failover + spend monitor | 1 |
| 8 | Safety: blast_radius + Gatekeeper | 2 |
| 9 | Safety: model_isolation + sanitizers | 2 |
| 10 | Safety: kill_switch (3 layers) | 1 |
| 11 | Safety: anomaly_detector | 2 |
| 12 | Safety: audit (HMAC-chained) + sensitive_detector | 2 |
| 12.5 | **Selective Perception** (unified battery + safety) | 4 |
| 13 | Permissions module (mac/win/linux) + onboarding | 2 |
| 14 | Agent loop + planner | 3 |
| 15 | Router (cascade → hybrid) + autonomy matrix | 2 |
| 16 | Delegation (8 CLIs + Ollama + sanitize) | 4 |
| 17 | TUI (Ink) | 3 |
| 18 | Wails app shell + main window | 3 |
| 19 | Overlay window + global hotkey (multi-OS) | 2 |
| 20 | Voice (whisper.cpp local + 3 TTS + wake word) | 3 |
| 21 | Onboarding (API key, permissions, backends, hotkey) | 2 |
| 22 | Computer use (3 backends + 4-tier router + verify) | 4 |
| 23 | Memory (3 layers + recall) | 3 |
| 24 | Skills (agentskills.io + auto-create + self-improve + hub client) | 2 |
| 25 | **User-Adaptive Engine** (model + dialectic + predictor + feedback) | 5 |
| 26 | MCP gateway | 3 |
| 26.5 | **P2P Sync** (E2E encrypted, no server) | 6 |
| 26.6 | **Action Replay** (24h scrubbable) | 3 |
| 27 | Auto-backup + uninstall flow | 1 |
| 28 | Web app (Next.js on Vercel, 6 pages, i18n) | 3 |
| 29 | Public Skills Hub (hub.synaptic.app) | 3 |
| 30 | i18n (6 languages, TS + Go) | 3 |
| 31 | Marketing assets (PH page, screenshots, demo video) | 2 |
| 32 | Build, sign, notarize, distribute v0.1.0 | 3 |
| 33 | Polish, docs, EULA.md, CLAUDE.md final | 3 |

**Total: 37 steps, ~91 working days (~18 weeks for solo + AI pair).**

---

## 29. Repository Structure (Target)

```
synaptic/
├── CLAUDE.md                          # This file
├── LOGBOOK.md                         # Append-only AI session log
├── README.md
├── EULA.md                            # Synaptic Freeware License v1
├── LICENSE                            # Proprietary
├── CONTRIBUTING.md
├── SECURITY.md
├── PRIVACY.md
├── Makefile
├── go.work, go.mod
├── .goreleaser.yml
├── .github/workflows/                 # ci, release, web, codeql
├── .golangci.yml
├── cmd/
│   └── synapticd/                     # Daemon entry
├── internal/
│   ├── config/
│   ├── storage/                       # SQLite + FTS5 + vec
│   ├── memory/                        # 3 layers
│   ├── skills/                        # agentskills.io
│   ├── router/                        # Hybrid with memory
│   ├── delegation/                    # CLI spawning
│   ├── subscription/                  # OAuth proxies
│   ├── computeruse/                   # 4-tier router
│   ├── mcp/                           # MCP client/server
│   ├── ipc/                           # JSON-RPC 2.0
│   ├── safety/                        # 5 safety modules
│   ├── perception/                    # Selective Perception
│   ├── adaptive/                      # User-Adaptive Engine
│   ├── agent/                         # Main loop + planner
│   ├── llm/                           # LLM SDK wrappers
│   ├── hotkey/                        # Global hotkey
│   ├── overlay/                       # Window control
│   ├── voice/                         # STT + TTS + wake word
│   ├── update/                        # Auto-update
│   ├── license/                       # EULA + install tracking
│   ├── api_key/                       # User API key manager
│   ├── backup/                        # Auto-backup
│   ├── sync/                          # P2P sync
│   ├── replay/                        # Action replay
│   ├── permissions/                   # TCC / UI Automation
│   ├── presence/                      # User activity heartbeat
│   ├── notify/                        # Native notifications
│   ├── autonomy/                      # Autonomy matrix
│   └── onboarding/                    # First-run flow
├── pkg/polymath/                      # Public Go SDK
├── api/                               # Generated types
├── bridge/                            # Python helpers (computer use)
├── ts/                                # TypeScript workspace
│   └── packages/
│       ├── protocol/
│       ├── tui/
│       ├── dashboard/
│       ├── mcp-client/
│       ├── llm/
│       └── cli/
├── app/                               # Wails app
│   ├── main.go, app.go
│   └── frontend/                      # React UI
├── web/                               # Next.js marketing site
├── hub/                               # Next.js Skills Hub
├── marketing/                         # Launch assets
├── configs/                           # default.yaml, schemas
├── migrations/                        # Symlink to internal/storage
├── scripts/                           # install, dev, bootstrap_bridge
├── docs/                              # Architecture, ADRs, guides
└── test/                              # integration, fixtures, mocks
```

---

## 30. The AI Workflow

**This section is critical for any AI model picking up this project.**

### 30.1 Before You Start
1. **Read `CLAUDE.md` end-to-end.** This is the single source of truth.
2. **Read `LOGBOOK.md`** to see the most recent session state.
3. **Read the relevant `docs/architecture/`** file(s) for the area you're working in.
4. **Read the relevant `docs/adr/`** (Architecture Decision Records) for context.
5. **Read `CONTRIBUTING.md`** for code style and PR conventions.

### 30.2 While You Work
- Follow the tech stack in Section 8 strictly.
- Follow the file structure in Section 29.
- Follow the survival invariants in Section 2.
- Never bypass the Gatekeeper, the audit log, or the safety layer.
- Never hardcode API keys, never log secrets, never commit `.env`.
- Use the local-first principle: data stays on disk.

### 30.3 When You Finish
**Append a session entry to `LOGBOOK.md`.** The format is:

```markdown
## [YYYY-MM-DD HH:MM] AI Model: <name>
**Session ID:** <unique id, ULID>
**Task:** <what you were asked to do>
**Files created:**
- `path/to/file.go` — <purpose>
- `path/to/file.ts` — <purpose>
**Files modified:**
- `path/to/existing.go` — <what changed>
**Decisions made:** <any new decisions, with rationale>
**Bugs/issues encountered:** <anything that blocked you>
**Open questions for next session:** <anything unresolved>
**Next steps:** <what should be done next>
---
```

### 30.4 Tone
- Be direct. No fluff.
- Be honest about mistakes. If something is broken, say so in the log.
- Be careful. This is an OS-level product. Sloppiness = real-world damage.
- Be efficient. The user has time pressure.
- Ask questions in the logbook if you get stuck. The next session will pick them up.

### 30.5 Hard Rules for AI Agents
1. **Never** delete or rewrite `CLAUDE.md` content silently. Append or annotate.
2. **Never** commit secrets, API keys, OAuth tokens, or `.env` files.
3. **Never** bypass the safety layer to "make something work faster".
4. **Never** introduce a new dependency without documenting it in `CLAUDE.md`.
5. **Never** skip tests for the safety or perception modules.
6. **Always** update the LOGBOOK before you finish.
7. **Always** read the latest LOGBOOK before you start.

---

## 31. Partner Commitment

This project is being built by a human + AI partnership. The human is the architect and product lead. The AI is the implementer and reviewer. We move fast. We do not ship broken code. We do not cut corners on security. We ship the best version of what we imagined, then we ship a better one. **Laziness is not in our vocabulary. Everything is possible.**

---

## 32. Glossary

| Term | Meaning |
|---|---|
| **Synaptic** | The product name. |
| **The Conductor** | The main agent loop + router. |
| **The Strategist** | Any LLM that decides WHAT to do. |
| **The Gatekeeper** | Deterministic rules engine that decides IF it's safe. |
| **Selective Perception** | Unified system for battery + safety + performance. |
| **Blast Radius** | READ / WRITE / NETWORK / DESTRUCTIVE classification. |
| **Smart Capturer** | Battery-aware screen access strategy selector. |
| **Twin Snapshot** | Pre-action verification using two AX tree snapshots. |
| **Anomaly Detector** | Behavioral watchdog for stuck loops, speed, duration. |
| **Audit Chain** | HMAC-chained append-only log. |
| **Adaptive Engine** | User-Adaptive Engine: closed learning loop. |
| **Dialectic** | Proposer + critic + adjudicator pattern for learning. |
| **P2P Sync** | Device-to-device encrypted sync, no central server. |
| **Hub** | Public Skills Hub at hub.synaptic.app. |
| **Overlay** | Floating chat/voice box. |
| **TUI** | Terminal UI (Ink). |
| **Wails** | Desktop shell framework (Go + web). |
| **CE-MCP** | Code-Execution MCP, context-decoupled delegation. |
| **CRDT** | Conflict-free Replicated Data Type (for sync). |
| **TCC** | Transparency, Consent, Control (macOS permission system). |
| **AT-SPI** | Assistive Technology Service Provider Interface (Linux). |
| **CGEventTap** | macOS low-level event hook. |

---

**This document is the foundation. Read it. Trust it. Extend it carefully. Never lose it.**
