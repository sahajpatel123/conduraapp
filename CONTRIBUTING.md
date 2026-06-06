# Contributing

> Synaptic is a proprietary, source-closed project. This document describes the conventions for **AI agents and a small set of named human contributors** working on the codebase under NDA.

If you are a user and have a feature request, please open a GitHub Issue or post in Discord (see README.md). This document is for the implementation team only.

---

## AI Agent Workflow (Mandatory)

1. **Read** `CLAUDE.md` end-to-end. This is the source of truth.
2. **Read** `LOGBOOK.md` to see the most recent session state.
3. **Read** the relevant architecture docs and ADRs.
4. **Read** `docs/guides/code-style.md` before writing any code.
5. **Do** your work.
6. **Append** a session entry to `LOGBOOK.md` before you finish.
7. **Stop** and wait for the next session if you encounter hard blockers.

If you skip step 1, you will write code that violates the survival invariants. Don't skip it.

---

## Code Style

### Go (`.go`)

- **Go 1.22+** required.
- **Format**: `gofmt` (mandatory), `goimports` (mandatory).
- **Linter**: `golangci-lint` with the config in `.golangci.yml` (mandatory).
- **Errors**: wrap with `fmt.Errorf("...: %w", err)`. Never use `panic` for recoverable errors.
- **Logging**: use `slog` (structured logging). Never `fmt.Println` in production code.
- **Naming**: package names are short, lowercase, no underscores. Interface names end in `-er` when possible.
- **Comments**: every exported symbol MUST have a doc comment. No `// TODO` without a corresponding GitHub issue.
- **Tests**: every package has a `*_test.go` file. Critical packages (safety, perception, agent, llm, ipc) MUST have >80% test coverage.
- **No CGO** unless absolutely necessary. Where used (macOS native), isolate to dedicated files.

### TypeScript (`.ts`, `.tsx`)

- **TypeScript 5.4+** required.
- **Format**: `prettier` (mandatory).
- **Linter**: `eslint` with `@typescript-eslint/recommended-type-checked`.
- **Naming**: `camelCase` for variables/functions, `PascalCase` for components/types, `UPPER_SNAKE_CASE` for constants.
- **Types**: prefer `type` over `interface` for object literals. Use `interface` for extensible contracts.
- **Async**: prefer `async/await` over `.then()`. Always handle promise rejection.
- **React**: functional components, hooks, no class components. Use Ink for TUI, React for Wails web.
- **Tests**: Vitest for unit tests, Playwright for E2E.

### File Headers

Every Go file starts with:

```go
// Package <name> provides <one-line description>.
// <Longer description if needed.>
package <name>
```

Every TypeScript file starts with:

```typescript
/**
 * <One-line description>
 *
 * <Longer description if needed.>
 */
```

---

## Commit Message Format

```
<type>(<scope>): <subject>

<body>

<footer>
```

Types: `feat`, `fix`, `docs`, `style`, `refactor`, `test`, `chore`, `perf`, `revert`.

Examples:
- `feat(safety): add blast-radius classifier`
- `fix(perception): cache AX tree for 500ms`
- `docs(adr): add ADR-0006 for P2P sync choice`
- `test(gatekeeper): add tests for destructive class`

---

## Pull Request Process

1. **Branch** from `main`: `git checkout -b feat/<name>` or `fix/<name>`.
2. **Commit** with the format above.
3. **Push** and open a PR.
4. **CI** must pass (lint, test, build on all 3 OSes).
5. **Review** by at least one other named contributor.
6. **Squash and merge** with a conventional commit message.

---

## The Survival Invariants (Reminder)

When in doubt, re-read Section 2 of `CLAUDE.md`. The seven non-negotiables are not negotiable.

If a feature conflicts with an invariant, **remove the feature**.

---

## Adding a New Dependency

1. Justify it in the PR description.
2. Update `CLAUDE.md` Section 8 (Tech Stack) with the new dependency and rationale.
3. Note the license compatibility — Synaptic ships with various open-source components and the combined license footprint must remain compatible with the proprietary distribution.
4. Pin the version.

---

## Performance Budgets

These are non-negotiable for v0.1.0:

| Metric | Target |
|---|---|
| Cold start to overlay-ready | < 500ms |
| Hotkey → overlay visible | < 100ms |
| First token from LLM | < 1.5s (streaming) |
| Computer-use action (AX-only) | < 200ms |
| Computer-use action (vision) | < 3s |
| IPC round-trip (local) | < 5ms |
| Memory footprint (idle) | < 150MB |
| Binary size (per OS) | < 20MB |

Any PR that regresses a budget must include a benchmark showing the regression and a justification, or be blocked.

---

## Test Strategy

- **Unit tests**: every package, alongside the code.
- **Integration tests**: `test/integration/`, run via `make test-integration`.
- **E2E tests**: Playwright for the web dashboard, manual scripts for desktop.
- **Safety tests**: dedicated `test/safety/` directory, runs on every PR.
- **Performance tests**: benchmarks for cold start, hotkey, IPC. Fails the build if budget violated.

Run all tests: `make test`.

---

## What NOT to Do

- ❌ Commit API keys, OAuth tokens, or `.env` files.
- ❌ Bypass the safety layer for any reason.
- ❌ Modify `CLAUDE.md` content silently.
- ❌ Skip the LOGBOOK entry.
- ❌ Add a new dependency without documenting it.
- ❌ Ship code without tests for the safety, perception, agent, llm, or ipc packages.
- ❌ Use `panic` for recoverable errors.
- ❌ Use `fmt.Println` in production Go code.
- ❌ Use `console.log` in production TypeScript code (use the logger).
- ❌ Hard-code paths, ports, or URLs.
- ❌ Use `any` in TypeScript without justification.
- ❌ Use `interface{}` in Go (use `any`).

---

**When in doubt, ask in the LOGBOOK. The next session will pick it up.**
