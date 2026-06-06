# Architecture 00 — Overview

> The conductor pattern: how Synaptic orchestrates every AI tool on the user's computer.

---

## The One-Sentence Architecture

Synaptic is a **persistent, on-device AI agent** that perceives the user's screen and actions through a battery-aware vision system (**Selective Perception**), decides what to do via a **hybrid-with-memory router** that picks the best model for each task, governs every action through a **deterministic Gatekeeper** that requires human consent for anything dangerous, and learns the user's behavior over time through a **closed learning loop** (the **Adaptive Engine**).

---

## The Three-Layer Mental Model

```
┌─────────────────────────────────────────────────────────────────┐
│  Layer 1: INTERFACES                                            │
│  Overlay (Wails) · TUI (Ink) · Web Dashboard · Voice · Menu Bar │
└─────────────────────────────────────────────────────────────────┘
                              ↕
┌─────────────────────────────────────────────────────────────────┐
│  Layer 2: ORCHESTRATION (The Conductor)                         │
│                                                                 │
│  ┌──────────────────┐  ┌──────────────────┐  ┌──────────────┐  │
│  │ Adaptive Engine  │  │  Agent Loop      │  │  Router      │  │
│  │  (learns user)   │  │  (planner+exec)  │  │  (hybrid)    │  │
│  └──────────────────┘  └──────────────────┘  └──────────────┘  │
└─────────────────────────────────────────────────────────────────┘
                              ↕
┌─────────────────────────────────────────────────────────────────┐
│  Layer 3: EXECUTION                                             │
│                                                                 │
│  ┌──────────────┐  ┌──────────────┐  ┌────────────────────┐    │
│  │  Computer    │  │  Delegation  │  │   MCP / Tools      │    │
│  │  Use         │  │  Bus         │  │                    │    │
│  │  (3 backends)│  │ (8 CLIs+12   │  │                    │    │
│  │              │  │  providers)  │  │                    │    │
│  └──────────────┘  └──────────────┘  └────────────────────┘    │
└─────────────────────────────────────────────────────────────────┘
```

### Layer 1: Interfaces
The user-facing surface. The **overlay** (floating chat/voice box) is the primary surface — summoned by hotkey. The **TUI** is for power users, SSH, and servers. The **web dashboard** is for management on the go. **Voice** is first-class.

### Layer 2: Orchestration (The Conductor)
The brain. The **Adaptive Engine** provides the user model (preferences, style, expertise). The **Agent Loop** decomposes tasks and executes them. The **Router** picks which model or CLI to use for each sub-task.

### Layer 3: Execution
The hands. **Computer Use** for physical GUI actions. **Delegation Bus** for sub-agents. **MCP / Tools** for filesystem, DB, web, etc.

---

## The Conductor Pattern

Synaptic is **not** a single model that does everything. It's a **conductor** that orchestrates many specialists. The metaphor:

- **The user** is the composer.
- **Synaptic** is the conductor.
- **The models, CLIs, and computer-use systems** are the orchestra.

The conductor doesn't play any instrument. It knows when to cue the strings, when to bring in the brass, when to let the woodwinds carry the melody. Each specialist does what it does best.

### Concretely

When the user says "Find the cheapest flight from NYC to Tokyo next month and book it":

1. **Adaptive Engine** checks the user model. The user prefers Kayak, not Expedia. They want economy. They have a $1500 budget.

2. **Router** picks the best model for "research" tasks. Today, that's `claude_code` (user priority). Falls back to `codex` if that fails.

3. **Agent Loop** decomposes into:
   - "Search Kayak for NYC → Tokyo, next month"
   - "Filter by economy, under $1500"
   - "Show top 3 to user, ask which to book"
   - "Click 'Select' on chosen flight, fill passenger form, click 'Book'"
   - "Confirm with user before payment"

4. **Computer Use** opens Safari (or the user's default browser), navigates to Kayak, types the search.

5. **Selective Perception** uses Accessibility API (ORAX Eye) to read the page. Only falls back to screenshots if AX is insufficient. Only falls back to vision CUA if both fail.

6. **Gatekeeper** says: "This is a NETWORK action that may involve a purchase. DESTRUCTIVE-adjacent. Need user consent." A native dialog pops up.

7. **User clicks "Allow"** in the dialog.

8. **Sub-agent** (a spawned Claude Code or Hermes) fills the form.

9. **Payment screen** comes up. **Gatekeeper** says: "DESTRUCTIVE. Real money. Need explicit human-in-the-loop." Another native dialog.

10. **User clicks "Allow"** with 2FA on their phone.

11. **Done.** Agent pauses, notifies user via overlay + native notification.

12. **Adaptive Engine** observes: "User chose Kayak, economy, $1500 budget. Update user model."

---

## The Closed Learning Loop

The Adaptive Engine is what makes Synaptic feel like *yours* over time.

```
   ┌──────────┐
   │  User    │
   └────┬─────┘
        │ actions
        ▼
   ┌──────────────┐
   │   Observer   │  (local, no telemetry)
   └────┬─────────┘
        │ observations
        ▼
   ┌──────────────┐
   │  Dialectic   │  (proposer + critic + adjudicator)
   └────┬─────────┘
        │ updates
        ▼
   ┌──────────────┐
   │  User Model  │  (Honcho-style structured)
   └────┬─────────┘
        │ predictions
        ▼
   ┌──────────────┐
   │  Predictor   │  (next-action suggestions)
   └────┬─────────┘
        │ suggestions
        ▼
   ┌──────────┐
   │  User    │
   └──────────┘
```

The user can **always** see, edit, and delete anything the engine has learned. This is non-negotiable.

---

## What's Different From Existing Agents

| Tool | Architecture | What it does well | What it lacks |
|---|---|---|---|
| **Hermes Agent** | Persistent + skills + memory | Self-improving skills, multi-platform | Weak computer use (Linux only) |
| **Antigravity CLI** | Single-vendor Google stack | Multi-agent orchestration in IDE | Locked to Google, no cross-CLI |
| **OpenClaw** | Multi-agent platform | Multi-channel, cron, sub-agents | No native computer use |
| **Claude Code** | Sub-agent CLI | Deep codebase reasoning | Single-vendor |
| **mac-cua / macOS-MCP** | Computer-use MCPs | Control macOS apps | Single-purpose, no orchestration |
| **Synaptic** | **Conductor + Adaptive + Selective Perception** | **All of the above, plus learning, plus safety, plus free** | _(this is what we're building)_ |

Synaptic is the **only** tool that combines:
- Persistent 24/7 runtime
- Native computer use (background-first, battery-aware)
- Cross-vendor LLM and CLI orchestration
- Closed-loop user learning
- Public Skills Hub
- P2P encrypted sync
- 5-language UI
- Free forever
- Deterministic safety layer

---

## Performance Targets

| Metric | Target |
|---|---|
| Cold start to overlay-ready | < 500ms |
| Hotkey → overlay visible | < 100ms |
| First token from LLM | < 1.5s |
| AX-only computer use | < 200ms |
| Vision computer use | < 3s |
| IPC round-trip | < 5ms |
| Memory (idle) | < 150MB |
| Binary size | < 20MB |

These are non-negotiable. See `CLAUDE.md` Section 28 (build order) for the plan to achieve them.

---

## The Survival Invariants (Recap)

See `CLAUDE.md` Section 2 for full text. The seven non-negotiables:

1. Strategist and Gatekeeper are separate systems.
2. Gatekeeper is the only path to physical action.
3. Destructive actions require a real human at the keyboard.
4. User can always stop the agent.
5. Every action is auditable.
6. Agent is a guest, not an owner.
7. OS permissions are granted by the user.

---

## Next Architecture Docs

- [01-router.md](01-router.md) — The hybrid-with-memory router
- [02-computer-use.md](02-computer-use.md) — 4-tier computer use system
- [03-perception.md](03-perception.md) — Selective Perception
- [04-safety.md](04-safety.md) — The safety layer
- [05-adaptive.md](05-adaptive.md) — The User-Adaptive Engine
- [06-delegation.md](06-delegation.md) — Delegation bus and sub-agents
- [07-memory.md](07-memory.md) — 3-layer memory system
- [08-sync.md](08-sync.md) — P2P sync protocol
- [09-ipc.md](09-ipc.md) — JSON-RPC 2.0 IPC
