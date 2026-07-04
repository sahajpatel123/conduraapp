# ADR-0003: Python Subprocess Bridges for Computer-Use Libraries

- **Status**: Accepted
- **Date**: 2026-06-06
- **Deciders**: Condura core team
- **Supersedes**: —
- **Superseded by**: —

---

## Context

Condura needs to control the computer:

- **macOS**: AX APIs, AppleScript, native CLI tools.
- **Windows**: UI Automation, PowerShell, native CLI tools.
- **Linux**: AT-SPI, xdotool, native CLI tools.

The best libraries for cross-platform computer use are **Python**:

- `pyautogui` — cross-platform mouse/keyboard.
- `pyobjc` / `Quartz` — macOS native.
- `uiautomation` / `pywinauto` — Windows native.
- `python-xlib` / `python-atspi` — Linux native.

**None of these have mature Go equivalents.** The Go community has some libraries (e.g., `go-rod` for browser automation), but for OS-level computer use, Python is the de facto standard.

We have to use Python for this. The question is **how**.

## Decision

**The Go daemon spawns Python subprocesses as bridges.**

Each computer-use library is wrapped in a small Python script that:

1. Exposes a JSON-RPC 2.0 interface over **stdio**.
2. Translates Go's requests (in our protocol) into the library's API.
3. Returns results in JSON.

The Go daemon manages the subprocess lifecycle: spawn, health-check, restart on crash, graceful shutdown.

The 3 bridges are:

1. **`bridge-orax`** (macOS only): wraps `pyobjc`, `Quartz`, AX APIs.
2. **`bridge-pyautogui`** (cross-platform): wraps `pyautogui` and per-OS native APIs.
3. **`bridge-mcp`** (cross-platform): wraps MCP computer-use servers (Playwright, file system, etc.).

## Rationale

### Why subprocess, not in-process Python (cgo, etc.)

- **Isolation**: a crash in the bridge doesn't take down the daemon.
- **Process management**: easy to restart, easy to monitor CPU/memory.
- **Versioning**: Python version is per-bridge, not per-daemon.
- **Distribution**: we ship the Python interpreter with the app (via PyOxidizer or a venv), so the user doesn't need to install Python.
- **No cgo**: avoids the complexity of mixing Go and Python in one process.

### Why JSON-RPC over stdio

- **Simple**: just write JSON to stdin, read JSON from stdout.
- **Streaming**: easy to stream events (screenshot updates, action progress).
- **Cross-language**: Python and Go both have first-class JSON support.
- **Bidirectional**: the bridge can push events back to the daemon.

### Why we don't use gRPC

- **Heavier**: requires protobuf, codegen, etc.
- **Overkill**: stdio JSON-RPC is fine for our throughput.
- **Distribution**: protobuf descriptors are a hassle.

### Why we don't use HTTP

- **Slower**: socket setup, parsing overhead.
- **More attack surface**: even if localhost-only, it's a bigger surface than stdio.
- **No benefit**: stdio is faster.

### Why not rewrite the libraries in Go

- **Time**: would take 6-12 months to rewrite `pyautogui` and friends.
- **Maintenance**: the Python libraries are actively maintained; a Go port would be a perpetual porting effort.
- **Coverage**: Python has 100% of what we need; Go has ~30%.

## Consequences

### Positive

- We get the best computer-use libraries in the world.
- Isolation = resilience.
- Easy to add new bridges.
- The bridges are reusable by other tools.

### Negative

- **Distribution complexity**: we have to ship a Python runtime. We use **PyOxidizer** to bundle Python + the bridges into a single executable, then we can re-exec it as a subprocess.
- **Cold start**: the bridge process has to start. We keep it warm (long-running) so this is ~50ms.
- **Two languages**: contributors need to know both Go and Python. The bridges are small (~500-1000 LOC each), so this is manageable.
- **Process management**: we have to monitor and restart the bridges. This is in `internal/bridge/supervisor.go`.

### Neutral

- We commit to **PyOxidizer** for the Python distribution.
- We commit to **JSON-RPC 2.0 over stdio** for the bridge protocol.

---

## The Bridge Architecture

```
┌────────────────────────────────────────────────────┐
│  Go Daemon (synapticd)                             │
│                                                    │
│  ┌──────────────┐                                  │
│  │  Computer    │                                  │
│  │  Use Module  │                                  │
│  └──────┬───────┘                                  │
│         │                                          │
│         │ JSON-RPC over stdio                      │
│         │                                          │
│  ┌──────▼───────┐  ┌───────────────┐               │
│  │  Bridge      │  │  Bridge       │               │
│  │  Supervisor  │  │  Pool         │               │
│  └──────┬───────┘  └───────┬───────┘               │
│         │                  │                       │
└─────────┼──────────────────┼───────────────────────┘
          │                  │
          │ spawn            │ spawn
          ▼                  ▼
┌──────────────────┐  ┌──────────────────┐
│  bridge-orax     │  │  bridge-pyautogui│
│  (Python)        │  │  (Python)        │
│                  │  │                  │
│  - macOS AX      │  │  - pyautogui     │
│  - AppleScript   │  │  - nut.js        │
│  - pyobjc        │  │  - xdotool       │
└──────────────────┘  └──────────────────┘
```

---

## The Bridge Protocol

The bridge exposes a JSON-RPC 2.0 interface. Example:

**Request** (Go → Python):
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "method": "ax.tree",
  "params": {
    "app": "Safari",
    "max_depth": 5
  }
}
```

**Response** (Python → Go):
```json
{
  "jsonrpc": "2.0",
  "id": 1,
  "result": {
    "app": "Safari",
    "elements": [
      { "role": "button", "name": "Back", "rect": [10, 80, 30, 30] },
      { "role": "textfield", "name": "Search", "rect": [50, 80, 400, 30], "value": "" }
    ]
  }
}
```

For streaming, the bridge sends notifications (no `id`) for events:

```json
{
  "jsonrpc": "2.0",
  "method": "_event",
  "params": {
    "event": "screenshot.taken",
    "data": { "path": "/tmp/synaptic-12345.png" }
  }
}
```

---

## The Bridge Methods

| Method | Description |
|---|---|
| `ax.tree` | Read the AX tree of an app |
| `ax.find` | Find an element by name, role, value |
| `ax.click` | Click on an AX element |
| `ax.type` | Type into an AX element |
| `screenshot.region` | Take a screenshot of a region |
| `screenshot.full` | Take a full-screen screenshot |
| `mouse.move` | Move the mouse |
| `mouse.click` | Click at coordinates |
| `mouse.scroll` | Scroll |
| `keyboard.type` | Type a string |
| `keyboard.hotkey` | Press a hotkey combo |
| `osascript.run` | Run AppleScript (macOS only) |
| `applescript.run` | Same as above (alias) |
| `clipboard.read` | Read the clipboard |
| `clipboard.write` | Write to the clipboard |
| `app.launch` | Launch an app by name/path |
| `app.quit` | Quit an app |
| `app.list` | List running apps |
| `window.list` | List windows of an app |
| `window.focus` | Focus a window |
| `health.ping` | Health check |

The full schema is in `configs/schemas/bridge/`.

---

## The Bridge Lifecycle

1. **Spawn** (on first use, or at daemon startup): supervisor starts the bridge process.
2. **Health-check**: every 5s, supervisor pings the bridge (`health.ping`).
3. **Restart**: if the bridge dies or fails health check 3 times, supervisor restarts it.
4. **Graceful shutdown**: on daemon SIGTERM, supervisor sends SIGTERM to the bridge, waits 5s, then SIGKILL.
5. **Resource limits**: cgroup (Linux) / job object (Windows) / sandbox (macOS) limits to 500MB RAM, 50% CPU.

---

## Security

The bridges are **powerful** — they can move the mouse, type, take screenshots. We isolate them:

- **Process isolation**: the bridge is a separate process. A crash doesn't kill the daemon.
- **Capability tokens**: each bridge has a token that proves the daemon authorized it. The bridge checks the token on every method call.
- **No network**: the bridges are **not allowed** to make network calls. Enforced by firewall rules (macOS: socket filter; Linux: nftables; Windows: WFP).
- **No filesystem access** outside of the Condura data directory and the screenshot temp dir.
- **Audit log**: every bridge call is logged.

---

## Related Docs

- [ADR-0001](0001-go-over-python.md) — Why Go for the daemon
- [02-computer-use.md](../architecture/02-computer-use.md) — 4-tier computer use
- [CLAUDE.md Section 6](../CLAUDE.md) — Stack details
