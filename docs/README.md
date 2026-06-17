# Condura Documentation

> Comprehensive documentation for the Condura project.

---

## Start Here

- **[CLAUDE.md](../CLAUDE.md)** — The single source of truth. Every AI agent and human contributor must read this first.
- **[LOGBOOK.md](../LOGBOOK.md)** — Append-only session log. Read this to see the current state of the project.
- **[README.md](../README.md)** — Public-facing overview.

---

## Architecture Deep-Dives

Detailed architecture docs for each subsystem.

| # | Document | Description |
|---|---|---|
| 00 | [overview.md](architecture/00-overview.md) | High-level architecture, the conductor pattern |
| 01 | [router.md](architecture/01-router.md) | The hybrid-with-memory router |
| 02 | [computer-use.md](architecture/02-computer-use.md) | 4-tier computer use system |
| 03 | [perception.md](architecture/03-perception.md) | Selective Perception (battery + safety) |
| 04 | [safety.md](architecture/04-safety.md) | The safety layer (5 modules + invariants) |
| 05 | [adaptive.md](architecture/05-adaptive.md) | The User-Adaptive Engine |
| 06 | [delegation.md](architecture/06-delegation.md) | Delegation bus and sub-agents |
| 07 | [memory.md](architecture/07-memory.md) | 3-layer memory system |
| 08 | [sync.md](architecture/08-sync.md) | P2P sync protocol |
| 09 | [ipc.md](architecture/09-ipc.md) | JSON-RPC 2.0 IPC |

---

## Architecture Decision Records (ADRs)

Why we made the choices we made.

- [ADR-0001](adr/0001-go-over-python.md) — Go for the core daemon
- [ADR-0002](adr/0002-typescript-for-ui.md) — TypeScript + React for the UI
- [ADR-0003](adr/0003-bridge-pattern.md) — Python subprocesses for computer-use libs
- [ADR-0004](adr/0004-ce-mcp.md) — Code-Execution MCP for delegation
- [ADR-0005](adr/0005-p2p-sync.md) — P2P sync over a central server

---

## Guides

How-to guides for contributors.

- [AI Onboarding](guides/ai-onboarding.md) — Step-by-step guide for AI agents picking up this project
- [Code Style](guides/code-style.md) — Go and TypeScript conventions
- _(More to come as the project evolves)_

---

## User Guide

End-user documentation. _(To be written in later phases.)_

- [Installation](user-guide/installation.md)
- [Quickstart](user-guide/quickstart.md)
- [Configuration](user-guide/configuration.md)
- [Backends](user-guide/backends.md)
- [Skills](user-guide/skills.md)
- [Memory](user-guide/memory.md)
- [Safety](user-guide/safety.md)
- [Troubleshooting](user-guide/troubleshooting.md)

---

## Recipes

Worked examples. _(To be written in later phases.)_

- [Research Agent](recipes/research-agent.md)
- [Dev Helper](recipes/dev-helper.md)
- [Safari Automation](recipes/safari-automation.md)
- [Scheduled Tasks](recipes/scheduled-tasks.md)

---

## API Reference

JSON-RPC method/event reference. _(To be written when the protocol is locked.)_

- [Methods](api/methods.md)
- [Events](api/events.md)
- [Types](api/types.md)

---

## Contributing

- [CONTRIBUTING.md](../CONTRIBUTING.md) — Code style, PR process, conventions
- [SECURITY.md](../SECURITY.md) — Vulnerability disclosure
- [LOGBOOK.md](../LOGBOOK.md) — Session log (every AI agent must read and append)
