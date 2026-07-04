# condura-mind

> **The project's constitution.** Every AI agent and human contributor starts
> here. Read `CLAUDE.md` end-to-end before any work.

This folder holds everything that *describes* the project — the rules, the
history, the architecture docs, the legal text, the AI agent's working memory.
It does **not** contain product code.

## Contents

| Path | Purpose |
|---|---|
| `CLAUDE.md` | The single source of truth. Project mission, locked decisions, non-negotiables, build order, repo structure, AI workflow rules. **Read this first.** |
| `LOGBOOK.md` | Append-only log of every AI session. Newest entries at the top. Add an entry when you finish a session. |
| `MISSION.md` | The one-paragraph mission statement and the "why this exists" paragraph. |
| `EULA.md`, `LICENSE` | The Synaptic Freeware EULA v1 and the source license. |
| `AGENTS.md` | Reference for the five specialized AI coding agents configured for this project. |
| `SECURITY.md`, `PRIVACY.md` | Public-facing security & privacy posture. |
| `STYLE.md`, `FOOTHPATH.md` | Visual + writing style guides. |
| `CHANGELOG.md` | Public changelog (rendered on the marketing site). |
| `CONTRIBUTING.md` | How to contribute. |
| `docs/` | Architecture, ADRs, guides, recipes, audits, threat models. |
| `synapse/` | AI understanding docs (`understanding.md` — the deep project map). |
| `legal/` | Reserved for clean legal-text home. |

## For AI agents

If you're an AI reading this: **read `CLAUDE.md` end-to-end before doing
anything**, then read `LOGBOOK.md` for the latest session state. Then either
follow the workflow in `CLAUDE.md §30` directly, or invoke the relevant skill.

When you finish a session, append a `LOGBOOK.md` entry per `CLAUDE.md §30.3`.

## For humans

If you're a human reading this: **`CLAUDE.md` is the real README**. Everything
else in this folder is reference material for that document.