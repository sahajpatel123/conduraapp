# Architecture 09 — JSON-RPC 2.0 IPC

> The wire protocol between the Go daemon and every client.

---

## The Goal

Condura has multiple clients (overlay, TUI, web dashboard, voice daemon) that all need to talk to the same Go daemon. The protocol must be:

1. **Simple** — easy to implement on any client (TypeScript, Rust, Go, Python).
2. **Bidirectional** — server can push events to clients.
3. **Streaming** — supports streaming responses (LLM tokens, action progress).
4. **Strongly typed** — schema-validated.
5. **Cancellable** — long-running calls can be cancelled.
6. **Multi-transport** — works over Unix socket, named pipe, TCP, WebSocket.

We use **JSON-RPC 2.0** (the standard) with a small set of conventions.

---

## The Transports

| Client | Transport | Why |
|---|---|---|
| **Overlay (Wails)** | Unix socket (macOS/Linux) / named pipe (Windows) | Local-only, no network surface |
| **TUI (Ink/TypeScript)** | Unix socket / named pipe | Local-only |
| **Web dashboard** | HTTPS over local TCP (port 7666) | Browser needs HTTP |
| **Voice daemon** | Unix socket | Local-only |
| **Remote (iPad)** | HTTPS + auth | Optional, opt-in |

**The local sockets are the primary path.** Remote access is opt-in and behind authentication.

---

## The Wire Format

Standard JSON-RPC 2.0 with one extension for streaming.

### Request

```json
{
  "jsonrpc": "2.0",
  "id": "req-1234",
  "method": "agent.run",
  "params": {
    "goal": "Book a flight to Tokyo",
    "context": { ... },
    "options": { ... }
  }
}
```

### Response (Success)

```json
{
  "jsonrpc": "2.0",
  "id": "req-1234",
  "result": {
    "run_id": "run-5678",
    "status": "started"
  }
}
```

### Response (Error)

```json
{
  "jsonrpc": "2.0",
  "id": "req-1234",
  "error": {
    "code": -32601,
    "message": "Method not found",
    "data": { "method": "agent.run2" }
  }
}
```

### Streaming (Extension)

For streaming responses (e.g., LLM tokens, action progress), we use **JSON-RPC notifications** (no `id`) on the same channel. The stream is correlated with the request by `run_id` or `request_id`.

```json
{
  "jsonrpc": "2.0",
  "method": "_stream",
  "params": {
    "request_id": "req-1234",
    "run_id": "run-5678",
    "event": "token",
    "data": { "delta": "Hello, " }
  }
}
```

The final notification has `event: "done"` (or `"error"`):

```json
{
  "jsonrpc": "2.0",
  "method": "_stream",
  "params": {
    "request_id": "req-1234",
    "run_id": "run-5678",
    "event": "done",
    "data": { "result": { ... } }
  }
}
```

---

## The Method Namespace

Methods are namespaced by subsystem:

| Namespace | Example methods |
|---|---|
| `agent.*` | `agent.run`, `agent.cancel`, `agent.pause`, `agent.resume` |
| `router.*` | `router.plan`, `router.delegates.list`, `router.delegates.health` |
| `perception.*` | `perception.frame`, `perception.strategy.set`, `perception.profiles.list` |
| `memory.*` | `memory.recall`, `memory.forget`, `memory.search`, `memory.export` |
| `skills.*` | `skills.list`, `skills.install`, `skills.uninstall`, `skills.run` |
| `safety.*` | `safety.policy.get`, `safety.policy.set`, `safety.audit.list` |
| `sync.*` | `sync.devices.list`, `sync.pair`, `sync.revoke`, `sync.now` |
| `adaptive.*` | `adaptive.model.get`, `adaptive.model.edit`, `adaptive.beliefs.list` |
| `settings.*` | `settings.get`, `settings.set` |
| `presence.*` | `presence.get`, `presence.set` |
| `system.*` | `system.status`, `system.version`, `system.shutdown` |

The full method reference is in `api/methods.md` (to be written in Phase 4).

---

## The Event Namespace

Server-pushed events:

| Event | When | Payload |
|---|---|---|
| `agent.run.started` | A run begins | `run_id`, `task` |
| `agent.run.progress` | Run emits progress | `run_id`, `step`, `action`, `status` |
| `agent.run.completed` | Run finishes | `run_id`, `result` |
| `agent.run.failed` | Run errors out | `run_id`, `error` |
| `agent.run.cancelled` | Run was cancelled | `run_id` |
| `safety.consent.request` | Gatekeeper needs user consent | `action_id`, `reason`, `options` |
| `safety.anomaly.alert` | Anomaly detector tripped | `severity`, `description` |
| `perception.frame.captured` | A new frame is captured (debug) | `strategy`, `app`, `redacted` |
| `perception.app.changed` | Foreground app changed | `from`, `to` |
| `sync.device.paired` | A new device paired | `device` |
| `sync.device.revoked` | A device was revoked | `device` |
| `sync.conflict.detected` | Sync conflict | `record_id`, `versions` |
| `router.delegate.error` | A delegate failed | `delegate`, `error` |
| `system.update.available` | An update is available | `version`, `severity` |
| `system.battery.critical` | Battery hit critical | `percent` |
| `adaptive.belief.formed` | A new belief was learned | `belief` |

---

## The Schema

All params and results are JSON Schema-validated. The schemas are in `configs/schemas/`. They are auto-generated from Go types via `jsonschema-gen`.

Example (simplified):

```yaml
# configs/schemas/agent.run.params.schema.yaml
$schema: http://json-schema.org/draft-07/schema#
$ref: "#/definitions/AgentRunParams"
definitions:
  AgentRunParams:
    type: object
    required: [goal]
    properties:
      goal:
        type: string
        minLength: 1
        maxLength: 100000
      context:
        type: object
      options:
        type: object
        properties:
          model:
            type: string
          stream:
            type: boolean
            default: true
          autonomy:
            type: string
            enum: [supervised, warn, autonomous]
```

The client SDKs (TypeScript, Go, Python) are generated from these schemas.

---

## Authentication

### Local (Unix socket / named pipe / localhost TCP)

**Trusted.** The OS already authenticated the local user. No additional auth needed.

If the local user wants to lock the daemon, they can set a **daemon password** in Settings. Clients must supply it on every connection (HMAC challenge-response).

### Remote (WAN / web)

- **HTTPS only.** TLS 1.3 minimum.
- **Device-pairing token**: a long random token, generated when the user opts in to remote access.
- **Ed25519 signatures**: each request is signed with the device's private key.
- **Rate-limited**: 100 req/min by default.
- **Auto-revoke**: 5 failed auth attempts → daemon refuses further connections for 10 min.

---

## The Daemon's RPC Server

The Go daemon uses a JSON-RPC 2.0 server library (`net/rpc/jsonrpc` or a third-party like `lthibault/jsonrpc2`). All methods are registered at startup.

```go
package daemon

func (d *Daemon) RegisterRPC(server *rpc.Server) {
    rpc.RegisterDefaultNamespaces = false
    server.Register(&AgentNamespace{daemon: d})
    server.Register(&RouterNamespace{daemon: d})
    server.Register(&PerceptionNamespace{daemon: d})
    // ...
}
```

Each namespace is a struct with methods that match the `Namespace.method` naming. Each method takes `(ctx context.Context, params <Type>) (result <Type>, err error)`.

---

## Cancellation

A long-running call can be cancelled by sending a cancel notification:

```json
{
  "jsonrpc": "2.0",
  "method": "_cancel",
  "params": { "request_id": "req-1234" }
}
```

The server uses Go's `context.Context` to propagate cancellation. The handler's context is cancelled, and any goroutine using `ctx.Done()` exits cleanly.

For HTTP/WebSocket transports, the client can also just close the connection.

---

## Backpressure & Buffering

Streaming events are buffered per-client. The buffer is 1024 events. If the client is too slow, the oldest events are dropped, and a `_overflow` event is sent.

The client can also enable **slow mode** (throttle to N events/sec) for UI rendering.

---

## Latency Targets

| Operation | Target |
|---|---|
| Method call (no streaming) | < 5ms |
| Streaming event delivery | < 5ms (excluding actual LLM time) |
| First byte of a streamed run | < 50ms |
| Cancel propagation | < 10ms |

---

## Versioning

The protocol has a version number. The client sends a `system.handshake` on connect:

```json
{
  "jsonrpc": "2.0",
  "id": "hs-1",
  "method": "system.handshake",
  "params": {
    "client": "synaptic-overlay",
    "client_version": "0.1.0",
    "protocol_version": "1.0"
  }
}
```

The server responds with its own version. If the protocol versions are incompatible, the client refuses to use the connection.

**Protocol v1.0 is locked for v0.1.0.** Breaking changes require v2.0.

---

## SDKs

Auto-generated SDKs in:

- **TypeScript** (for the overlay, TUI, web dashboard).
- **Go** (for internal use and CLI tools).
- **Python** (for the computer-use bridges).
- **Swift** (for any iOS/iPadOS app, future).
- **Rust** (for performance-critical clients, future).

SDKs are in `ts/packages/sdk`, `internal/sdk`, etc.

---

## Related Docs

- [00-overview.md](00-overview.md) — The conductor pattern
- [06-delegation.md](06-delegation.md) — How the Delegation Bus uses IPC
- [CLAUDE.md Section 9](../CLAUDE.md) — IPC section
