# AI Onboarding Guide

> A step-by-step guide for AI agents (or humans) picking up the Synaptic project.

---

## Welcome

You're about to work on **Synaptic**, a free, on-device, persistent AI agent. The project is in **Phase 0 (Foundation)** as of 2026-06-06. Before you write a single line of code, read this guide.

---

## Step 1: Read the Source of Truth (in order)

1. **[CLAUDE.md](../../CLAUDE.md)** — the master thinking document. It contains:
   - 36 locked decisions
   - 7 non-negotiables
   - All module designs (router, computer use, perception, safety, adaptive, delegation, memory, sync, IPC)
   - The 37-step build order
   - The AI Workflow rules

   **You must read this.** It is non-negotiable. If a decision is in CLAUDE.md, it is locked. You do not re-litigate it.

2. **[LOGBOOK.md](../../LOGBOOK.md)** — the append-only AI session log. It contains:
   - What the previous AI did.
   - Decisions made in flight.
   - Open questions.
   - Next steps.

   **Read the latest entry** to know where we are. **Append a new entry** when you finish your work.

3. **[README.md](../../README.md)** — the public-facing overview. Understand what we're building and why.

4. **[EULA.md](../../EULA.md)** and **[LICENSE](../../LICENSE)** — the license. Source is proprietary, binary is free.

5. **[SECURITY.md](../../SECURITY.md)** and **[PRIVACY.md](../../PRIVACY.md)** — security and privacy commitments.

6. **[CONTRIBUTING.md](../../CONTRIBUTING.md)** — code style, PR process, conventions.

---

## Step 2: Read the Architecture Docs (in order)

These are the deep-dives referenced from CLAUDE.md. Read them in this order:

1. [00-overview.md](../architecture/00-overview.md) — The conductor pattern.
2. [01-router.md](../architecture/01-router.md) — The hybrid-with-memory router.
3. [02-computer-use.md](../architecture/02-computer-use.md) — 4-tier computer use.
4. [03-perception.md](../architecture/03-perception.md) — Selective Perception.
5. [04-safety.md](../architecture/04-safety.md) — The safety layer.
6. [05-adaptive.md](../architecture/05-adaptive.md) — The User-Adaptive Engine.
7. [06-delegation.md](../architecture/06-delegation.md) — Delegation Bus.
8. [07-memory.md](../architecture/07-memory.md) — 3-layer memory.
9. [08-sync.md](../architecture/08-sync.md) — P2P sync.
10. [09-ipc.md](../architecture/09-ipc.md) — JSON-RPC IPC.

Then read the ADRs:

- [ADR-0001: Go over Python](../adr/0001-go-over-python.md)
- [ADR-0002: TypeScript for UI](../adr/0002-typescript-for-ui.md)
- [ADR-0003: Bridge pattern](../adr/0003-bridge-pattern.md)
- [ADR-0004: Code-Execution MCP](../adr/0004-ce-mcp.md)
- [ADR-0005: P2P sync](../adr/0005-p2p-sync.md)

---

## Step 3: Read the Code Style Guides

- [code-style.md](code-style.md) — Go and TypeScript conventions.

---

## Step 4: Know Where We Are

The project is in **Phase 0 (Foundation)**. The current state:

- ✅ `CLAUDE.md` exists (~1,800 lines).
- ✅ `LOGBOOK.md` exists with the first session entry.
- ✅ `EULA.md`, `LICENSE`, `README.md`, `CONTRIBUTING.md` exist.
- ✅ `SECURITY.md`, `PRIVACY.md` exist.
- ✅ `docs/` is structured (architecture, adr, guides, user-guide, recipes, api).
- ✅ `docs/architecture/00-overview.md` through `09-ipc.md` exist.
- ✅ `docs/adr/0001` through `0005` exist.
- ✅ `docs/guides/ai-onboarding.md` (this doc) and `code-style.md` exist.
- ⏳ `docs/user-guide/` and `docs/recipes/` are empty placeholders.
- ⏳ `docs/api/` is empty.
- ❌ No code yet. Not a single line of Go, TypeScript, or Python.

**The next phase is Phase 1: Repo Skeleton** (see CLAUDE.md Section 28 for the 37-step build order). It will:

1. Initialize the Go module.
2. Set up the Wails app.
3. Set up the React/TypeScript overlay.
4. Set up the Ink TUI.
5. Set up the Python bridge skeletons.
6. Set up CI/CD (GitHub Actions).
7. Set up GoReleaser.

---

## Step 5: Understand the AI Workflow Rules

These rules are **non-negotiable**. They are in `CLAUDE.md` Section 30, but here they are again for emphasis:

### Before Each Session

1. Read `CLAUDE.md` and `LOGBOOK.md`.
2. Read the latest `LOGBOOK.md` entry to know the state.
3. Read the architecture doc(s) for the area you're working on.
4. Read the relevant ADR(s).

### During the Session

1. **Stay within the locked decisions.** If a decision is in CLAUDE.md, do not re-litigate it. If you think it's wrong, append to LOGBOOK.md with a clear rationale and ask the user.
2. **Optimize for the long term.** Don't hack. Don't take shortcuts. Don't add "TODO" comments instead of doing the work.
3. **Write tests for safety-critical code.** Coverage must be >80% for: `internal/safety`, `internal/perception`, `internal/agent`, `internal/llm`, `internal/ipc`.
4. **No silent failures.** If something fails, surface it. Log it. Report it.
5. **Performance budgets are non-negotiable.** See CLAUDE.md for the targets.
6. **Document as you go.** Public functions get doc comments. Files get header doc comments. New concepts get a doc in `docs/`.

### After Each Session

1. **Append a new entry to LOGBOOK.md.** Format:

   ```markdown
   ## [<ISO timestamp UTC>] AI Model: <name>
   
   **Session ID**: <unique id>
   **Branch**: <git branch>
   **Task**: <one-line summary>
   
   ### Files created/modified
   - ...
   
   ### Decisions
   - ...
   
   ### Bugs
   - ...
   
   ### Open questions
   - ...
   
   ### Next steps
   - ...
   ```

2. **Update CLAUDE.md** if a new decision was made (with explicit user approval).
3. **Commit your work** — but only if the user asks. (You don't commit unless explicitly told.)
4. **Tell the user what you did** — concise, no fluff.

---

## Step 6: The Hard Rules

These are the **inviolable rules**. Violating any of them is a **build failure**.

1. **The Strategist and Gatekeeper are separate systems.** The LLM never executes physical action. The Gatekeeper (deterministic, in Go) is the only path.
2. **The Gatekeeper is the only path to physical action.** There is no fast path, no trusted mode, no override.
3. **Destructive actions require a real human at the keyboard.** Always.
4. **The user can always stop the agent.** 3 layers of kill switch, all independent of the agent.
5. **Every action is auditable.** HMAC-chained, append-only log.
6. **The agent is a guest, not an owner.** It has no special privileges beyond what the user grants.
7. **OS permissions are granted by the user.** We request them. We do not assume them.

If a feature requires violating one of these, **the feature is not built**.

---

## Step 7: The Performance Budgets (Non-Negotiable for v0.1.0)

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
| Test coverage (safety/perception/agent/llm/ipc) | > 80% |

---

## Step 8: The Style Rules (Quick Reference)

Full rules in [code-style.md](code-style.md). Quick rules:

### Go

- All public functions get doc comments.
- All files start with `// Package <name>...` or `// File: <name>.go ...` header.
- `gofmt` + `goimports` enforced.
- `golangci-lint` with default rules + `revive` + `gocritic` + `gosec`.
- Errors are wrapped: `fmt.Errorf("doing X: %w", err)`.
- No global state. Use dependency injection.
- Context is the first arg: `func Foo(ctx context.Context, ...)`.
- Tests for every public function. Table-driven where applicable.

### TypeScript / React

- All public functions get JSDoc.
- All files start with `/** ... */` JSDoc header.
- `eslint` + `prettier` enforced.
- Functional components only (no class components).
- Hooks: follow the rules of hooks.
- State: Zustand for global, `useState` for local, TanStack Query for server.
- Styling: Tailwind CSS.
- Tests: Vitest + React Testing Library.

### Python (Bridges only)

- All public functions get docstrings.
- Type hints everywhere.
- `ruff` + `black` enforced.
- No global state. Use dependency injection.
- Tests: `pytest`.

---

## Step 9: The 37-Step Build Order

See `CLAUDE.md` Section 28 for the full list. The 37 steps are grouped into 7 phases:

- **Phase 0 — Foundation (DONE)**: docs, decisions, conventions.
- **Phase 1 — Repo Skeleton**: Go module, Wails app, React overlay, Ink TUI, Python bridges, CI/CD.
- **Phase 2 — Core Daemon**: IPC server, config, store, secrets, logger.
- **Phase 3 — Safety & Perception**: Gatekeeper, blast-radius classifier, anomaly detector, audit log, Selective Perception, AX bridge.
- **Phase 4 — LLM & Router**: 12 provider clients, hybrid router, streaming, cancel.
- **Phase 5 — Computer Use & Memory**: 4-tier computer use, 3-layer memory, Adaptive Engine.
- **Phase 6 — Sub-Agents & Skills**: 8 CLI delegates, Skills Hub, P2P sync, voice.
- **Phase 7 — Polish & Launch**: i18n, accessibility, performance, packaging, marketing.

**Total estimated time: ~91 working days.**

---

## Step 10: Get to Work

When you're ready to start, follow this checklist:

1. ✅ Read CLAUDE.md (entirely).
2. ✅ Read LOGBOOK.md (latest entry).
3. ✅ Read the architecture doc(s) for your area.
4. ✅ Read the relevant ADR(s).
5. ✅ Read code-style.md.
6. ✅ Read CONTRIBUTING.md.
7. ⏳ Verify the current phase. If it's not your turn, wait or hand off.
8. ⏳ Make a plan. Get the user's approval.
9. ⏳ Implement. Test. Document.
10. ⏳ Update LOGBOOK.md.
11. ⏳ Commit (only if asked).
12. ⏳ Report to the user.

---

## A Note on "Laziness Is Not in Our Vocabulary"

The user has stated this as a guiding principle. It means:

- **No half-measures.** Do the work properly.
- **No "we'll do it later" placeholders.** If it's in scope, do it.
- **No silent failures.** If something is hard, surface it.
- **No skipping tests.** Test coverage is non-negotiable.
- **No "TODO" comments.** If you can't finish it, don't start it.

The user is the architect. The AI is the implementer. The user reviews. If you find a problem, you tell the user. If you have a better idea, you propose it. The user decides.

---

## Welcome to the Project

Synaptic is an ambitious project. It's a free, on-device AI agent that orchestrates every AI tool a user has. It has a safety-first design, a battery-aware perception system, a closed-loop learning engine, and a public Skills Hub.

If you follow the rules, respect the decisions, and do the work, you'll do well. If you cut corners, you'll be told.

Let's build it.
