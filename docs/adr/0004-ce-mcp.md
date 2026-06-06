# ADR-0004: Code-Execution MCP for Delegation

- **Status**: Accepted
- **Date**: 2026-06-06
- **Deciders**: Synaptic core team
- **Supersedes**: —
- **Superseded by**: —

---

## Context

Synaptic's sub-agents (Claude Code, Codex, Hermes, etc.) need to **use tools**: read files, search the web, query a database, run a command, call an API.

There are 3 patterns for giving an LLM tool access:

1. **Function calling (tool use)**: the LLM is given a list of tools with JSON schemas. It returns a tool call; we execute it. Repeat.
2. **Retrieval / RAG**: the LLM has access to a search system. It asks queries, gets back chunks.
3. **Code execution**: the LLM writes code (e.g., Python) that calls APIs. We run the code in a sandbox.

The traditional pattern is **#1 (function calling)**. It's simple, well-supported, and works.

But there's a pattern gaining traction: **#3 (code execution MCP)**. Anthropic and others have argued it's better for complex agents.

## Decision

**We support both patterns, but **prefer code-execution MCP for complex, multi-step tasks**, and **function calling for simple, single-step tasks**.**

The Delegation Bus picks the right pattern per sub-task.

## Rationale

### Why code-execution MCP

Anthropic's ["Code Execution with MCP"](https://www.anthropic.com/news/code-execution-with-mcp) makes the case:

1. **Context efficiency**: instead of stuffing tool definitions and tool results into the context, the LLM writes code that orchestrates tools, and only the final result goes into context.
2. **Deterministic control flow**: loops, conditionals, error handling — all in code, not in the LLM's probabilistic output.
3. **Privacy**: intermediate results stay in the sandbox, not in the LLM's context.
4. **Composability**: the LLM can build higher-level tools from primitives.
5. **Token savings**: ~70-90% reduction in tokens for multi-step workflows.

Example:

**Function-calling version** (high token cost, brittle):

```
LLM: "I'll search for flights from NYC to Tokyo next month."
  → tool: search_flights(origin="NYC", destination="Tokyo", date="2026-07-10")
  ← result: [{...}, {...}, ...]
LLM: "I see 3 options. Let me filter by price."
  → tool: filter(flights, max_price=1500)
  ← result: [{...}]
LLM: "Now let me check the user's calendar for conflicts."
  → tool: get_calendar(start="2026-07-10", end="2026-07-20")
  ← result: [...]
... (many more turns, each with tool definitions + results in context)
```

**Code-execution MCP version** (low token cost, deterministic):

```python
# LLM writes this code:
flights = search_flights(origin="NYC", destination="Tokyo", date="2026-07-10")
affordable = [f for f in flights if f.price <= 1500]
conflicts = check_calendar(start="2026-07-10", end="2026-07-20")
good = [f for f in affordable if not conflicts(f.date)]
return good[:3]
```

The LLM's context is just the final result. The intermediate steps never left the sandbox.

### Why we still support function calling

- **Simpler**: many tasks don't benefit from code execution.
- **Better for one-off tool calls**: "what's the weather?" doesn't need code.
- **Some models don't support code execution**: smaller local models may not.
- **Easier to debug**: tool-call traces are easier than code-execution traces.

### When each is used

The Delegation Bus (or the Strategist) picks:

- **Single tool call, simple result** → function calling.
- **Multi-step workflow, intermediate data** → code-execution MCP.
- **The user explicitly says** "use code" or "use tools" → respect that.

### How we ship this

We provide:

1. **A code-execution sandbox** (Docker container, gVisor for extra safety).
2. **Pre-installed tool libraries** (file system, web search, DB, calendar, etc.).
3. **A code-execution MCP server** that the LLM talks to.
4. **A function-calling adapter** for legacy tools.
5. **A router** that picks which pattern per task.

The sandbox is **per-session** and **destroyed** at session end. It has no network access to the public internet (only to the user's configured services). It has read-only access to `~/` (or whatever the user configures) and full access to `/tmp/synaptic/`.

---

## Consequences

### Positive

- Major token savings for complex tasks.
- Deterministic control flow.
- Privacy for intermediate data.
- Composability.

### Negative

- **More complex**: we have to maintain a sandbox, a code-execution MCP server, and a sandbox-aware tool library.
- **Debugging is harder**: a Python error in a sandbox is harder to surface than a tool-call error.
- **Some users may not trust it**: "the LLM is running code on my machine?" We'll need clear UI for what's running.

### Neutral

- We commit to **Docker for the sandbox** (or containerd).
- We commit to **gVisor** for additional safety (Linux only; macOS uses `sandbox-exec`, Windows uses Job Objects + AppContainer).

---

## The Sandbox

### macOS: `sandbox-exec`

We use Apple's `sandbox-exec` with a custom profile that:

- Allows read access to `~/` (configurable).
- Allows write access to `/tmp/synaptic/`.
- Allows network access only to the user's configured services.
- Disallows fork, exec of arbitrary binaries.
- Disallows kernel calls.

### Linux: gVisor + Docker

We use Docker with `runtime: runsc` (gVisor) for additional isolation. The container:

- Read-only root filesystem (except for `/tmp/synaptic/`).
- No capabilities.
- No network namespaces (only allowlisted services).
- Memory limit: 2GB.
- CPU limit: 2 cores.
- Time limit: 5 min per code execution.

### Windows: Job Objects + AppContainer

We use Windows Job Objects for resource limits and AppContainer for capability isolation. The sandbox:

- Has access only to `~/` and `/tmp/synaptic/`.
- Has no network (except allowlisted).
- Memory: 2GB.
- CPU: 2 cores.
- Time: 5 min per execution.

---

## The Tool Library

The sandbox has pre-installed:

- `fs` (read, write, list, search files)
- `web` (search, fetch, extract)
- `db` (SQLite, Postgres, MySQL connectors)
- `calendar` (Google Calendar, iCloud, local)
- `mail` (IMAP, SMTP, Gmail API, etc.)
- `shell` (sandboxed shell — see below)
- `image` (read, write, transform, OCR)
- `pdf` (read, write, extract)
- `code` (run a sub-agent in a sub-sandbox)
- `synaptic` (call back into the Synaptic daemon for higher-level ops)

Each tool is a Python module that the sandbox can `import`. The LLM's generated code uses these.

---

## The Sandboxed Shell

The `shell` tool is special: it allows the LLM to run **shell commands**, but in a restricted environment:

- No `sudo`, no `su`.
- No `rm -rf /` (path is checked against an allowlist).
- No network commands except allowlisted.
- No subprocess of arbitrary binaries.
- Output is captured and returned.

This is the **riskiest** tool. It's enabled by default but the user can disable it in Settings.

---

## The Code-Execution MCP Server

The LLM talks to a local MCP server that:

- Accepts code (Python 3.11+).
- Runs it in the sandbox.
- Returns the result (stdout, stderr, return value, files written).
- Streams events (progress, errors).

The server is implemented in **Python** (so it can be co-located with the sandbox) but speaks **MCP** (Model Context Protocol) over **stdio**.

---

## The Function-Calling Adapter

For legacy tools and simpler tasks, we provide a function-calling adapter:

- The Strategist exposes a list of tools with JSON schemas.
- The LLM returns tool calls.
- The adapter executes them (in-process for safe tools, in the sandbox for risky ones).
- Results are returned to the LLM.

This adapter is **a wrapper around the sandbox's tool library** so we don't duplicate code.

---

## The Router

The Delegation Bus picks the pattern:

```python
def pick_pattern(task, sub_task, model):
    if task.complexity == "simple" or len(task.steps) <= 2:
        return "function_calling"
    if task.has_intermediate_data and model == "claude-opus-4-7":
        return "code_execution"
    if user_prefers("function_calling"):
        return "function_calling"
    if user_prefers("code_execution"):
        return "code_execution"
    return "code_execution" if model.supports_code_execution() else "function_calling"
```

The user can override the default in Settings.

---

## Related Docs

- [00-overview.md](../architecture/00-overview.md) — The conductor pattern
- [06-delegation.md](../architecture/06-delegation.md) — Delegation Bus
- [Anthropic's Code Execution with MCP](https://www.anthropic.com/news/code-execution-with-mcp)
- [CLAUDE.md Section 7](../CLAUDE.md) — Sub-agents section
