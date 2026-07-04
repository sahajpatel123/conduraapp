# condura-sdk

> **Reserved.** This folder is a placeholder for the public **Condura Go SDK**
> — the small, stable, well-documented API surface that lets third parties
> embed Condura in their own Go programs (custom MCP servers, IDE plugins,
> alternate front-ends, automation tools). See `condura-mind/CLAUDE.md` §29
> (Repository Structure → `pkg/polymath/`) for the original target.

## Status

Not yet implemented. The SDK is targeted for **v0.2.0** per
`condura-mind/docs/roadmap-v0.2.0.md`.

When work begins, this folder will hold:

- `pkg/` — public Go packages (renamed from `internal/` patterns that are safe
  to expose)
- `go.mod` — separate module so external consumers can `go get` it without
  pulling in the full daemon
- `examples/` — runnable examples for each SDK API
- `docs/` — generated godoc + human-readable guides
- `CHANGELOG.md` — public API stability log

## API design principles (planned)

- **No internal types leak.** Anything exported must be safe for third-party
  consumption — no `internal/` imports, no daemon-only state.
- **One package per concern.** E.g. `pkg/llm/` (LLM client factory),
  `pkg/ipc/` (JSON-RPC client), `pkg/skills/` (skill authoring SDK).
- **Backward compatible.** Breaking changes bump the major version, period.
- **Plays well with `context`.** Every entry point takes a `context.Context`.

## How to start work here

Until v0.2.0 begins, treat this folder as **read-only scaffolding**. Don't
move `internal/` packages into `pkg/` yet — the boundary between "internal"
and "public SDK" is a design decision that should be made deliberately when
the SDK work begins.

## Related folders

- `condura-app/internal/` — the candidate surface area (most SDK packages
  will be carved out of here, with renaming and dep cleanup)
- `condura-app/cmd/condura/` — the legacy CLI client that already exercises
  much of the IPC contract; useful as an SDK usage reference
- `condura-mind/docs/architecture/` — has the IPC protocol spec that the SDK
  client must implement