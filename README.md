# Condura

> **Free, local-first, OS-native AI agent.** Conductor of every other AI tool on
> your computer. Talks to Claude, GPT, Gemini, Ollama, Claude Code, Codex,
> Antigravity, and any ChatGPT Plus / Claude Pro / Gemini AI Pro subscription.
> Costs nothing.

A topic-sliced monorepo. Each top-level folder is one product surface; you can
navigate the codebase by topic without ever touching code that isn't relevant.

---

## The map

```
condura-app/        Backend daemon + Go libraries (the conductor)
condura-gui/        All in-app user-facing surfaces (desktop app, TUI)
condura-ui/         Marketing website (Next.js)
condura-studio/     Video / film creation (Remotion)
condura-brand/      Visual identity & marketing assets
condura-ops/        CI, scripts, release tooling, deployment configs
condura-mind/       Project constitution + agent context (CLAUDE.md, docs)

condura-hub/        Reserved — public Skills Hub (v0.2.0)
condura-sdk/        Reserved — public Go SDK (v0.2.0)

bin/                Build artifacts (gitignored)
```

---

## Quick start

```bash
# Build everything Go
make build

# Run the daemon (in foreground)
make run-daemon

# Build the desktop GUI (requires Wails CLI + Node)
make build-gui

# Run tests + lint + vet
make verify
```

## Where to look

| You want to… | Go to |
|---|---|
| Understand the project | `condura-mind/README.md`, `condura-mind/CLAUDE.md` |
| Change the daemon / backend | `condura-app/cmd/condurad/main.go`, `condura-app/internal/` |
| Change the desktop GUI shell | `condura-app/cmd/condura-gui/main.go` |
| Change the Svelte UI | `condura-gui/frontend/src/` |
| Change the TUI | `condura-app/cmd/condura-tui/main.go` |
| Change the marketing site | `condura-ui/app/` |
| Change design tokens / brand mark | `condura-brand/` (starter kit — see its README) |
| Update CI | `condura-ops/ci/` |
| Build a release | `condura-ops/release/goreleaser.yml` |
| Read the most recent session log | `condura-mind/LOGBOOK.md` |
| Find an architectural decision | `condura-mind/docs/adr/` |

## Why this shape?

Topic-slicing trades Go's idiomatic layer-slicing (`cmd/`, `internal/`, `pkg/`)
for **product-slicing**: every folder maps to a thing the user sees or the
operator does. The trade-off is that the Go `internal/` rule forces the shell
and TUI binaries (which depend on `condura-app/internal/...`) to live under
`condura-app/cmd/` — see `condura-gui/README.md` for the full story.

## License

Source-available on request. Binary distributed under the Synaptic Freeware
EULA v1 — see `condura-mind/EULA.md`.