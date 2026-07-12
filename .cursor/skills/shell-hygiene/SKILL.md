---
name: shell-hygiene
description: >-
  How shell/subagents must start, reuse, and never wait on forever-servers
  (Vite, condurad, docker). Use when launching Task shell agents, starting
  dev servers, or cleaning hung terminals.
---

# Shell hygiene for agents

## Quick protocol

1. **Probe** — `lsof -nP -iTCP:5173 -sTCP:LISTEN` and `:7666` (or needed ports).
2. **Reuse** — if listening and healthy, skip start.
3. **Start once** — `block_until_ms: 0`; watch only for a ready log line (e.g. `Local:` / `starting condurad`).
4. **Work** — run tests, edits, browser checks. Do **not** await the server PID.
5. **Leave** intentional Vite/daemon running for the session; **kill** hung installers and duplicate listeners.

## Forbidden patterns

```bash
# Bad — waits forever
npm run dev
await until process exits

# Bad — duplicate stack
# (Vite already on 5173) npm run dev -- --port 5173

# Bad — orphan installer
npx playwright install   # then leave it for 40+ minutes
```

## Good patterns

```bash
# Good — one-shot probe
lsof -nP -iTCP:5173 -sTCP:LISTEN

# Good — background start, verify once
# Shell: block_until_ms 0 → then AwaitShell pattern "Local:" once → continue

# Good — verify HTTP instead of waiting on PID
curl -s -o /dev/null -w "%{http_code}" http://127.0.0.1:5173/
```

## Subagent prompt snippet

Paste into Task `shell` prompts:

> Do not start Vite or condurad if ports 5173/7666 are already listening. If you must start one, use a background shell, confirm ready once, then proceed. Never wait for the server process to exit. Kill hung npx/playwright installs; prefer browser MCP for screenshots.
