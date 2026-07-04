# ADR-0001: Go for the Core Daemon

- **Status**: Accepted
- **Date**: 2026-06-06
- **Deciders**: Condura core team
- **Supersedes**: —
- **Superseded by**: —

---

## Context

Condura needs a **persistent, on-device daemon** that:

1. Runs 24/7 in the background with low resource use.
2. Speaks to 12 LLM providers and 8 sub-agent CLIs.
3. Performs fast, deterministic safety checks on every action.
4. Handles IPC from multiple local clients.
5. Cross-compiles cleanly to macOS, Windows, Linux.
6. Distributes as a single static binary.

We considered Python, Rust, and Go.

## Decision

**We use Go 1.22+ for the core daemon.**

## Rationale

### Go wins on

- **Single static binary**: trivial to distribute. No runtime to install.
- **Cross-compilation**: `GOOS=windows GOARCH=amd64 go build` just works.
- **Concurrency primitives**: goroutines and channels are exactly what we need for a multi-delegate, multi-client, streaming server.
- **Standard library**: `net/rpc`, `os/exec`, `crypto/*`, `net/http` are all first-class.
- **Performance**: 3-5x faster than Python for our hot paths. Memory: ~50-100MB idle.
- **Type safety**: catches bugs at compile time, which is critical for a safety-critical system.
- **Ecosystem**: `gorilla/websocket`, `libp2p`, `mattn/go-sqlite3`, etc., are all mature.
- **Team velocity**: 1.5-2x faster development than C++/Rust for the kind of code we write (network, IPC, business logic).

### Why not Python

- **GIL**: limits true parallelism for our multi-delegate architecture.
- **Distribution**: requires a Python runtime, which is fragile to install.
- **Performance**: 5-10x slower on hot paths.
- **Type safety**: dynamic typing is a bad fit for a safety-critical system.
- **Memory**: harder to reason about (no zero-cost abstractions).

That said, **we use Python for the 3 computer-use bridges** (see ADR-0003). The bridges are subprocesses that wrap `pyautogui`, `nut.js`, etc. — libraries that don't exist in Go.

### Why not Rust

- **Slower development velocity**: borrow checker and lifetime annotations are not free.
- **Smaller ecosystem** for our specific needs (LLM SDKs, MCP, libp2p).
- **Harder to hire for**.
- **Distribution**: also a single static binary, so this is a wash.

Rust would be a defensible alternative. We chose Go for **velocity and ecosystem** at v0.1.0. We may revisit for v1.0 if we hit performance walls (we don't expect to).

### Why not Node.js / TypeScript

- **TypeScript** is used for the **client-side** (overlay, TUI). It's the right tool there.
- **Node.js** is **not** a good fit for the daemon: harder to distribute, slower cold start, no real concurrency model.

## Consequences

### Positive

- Fast cross-platform builds.
- Easy hiring (Go is widely known).
- Robust concurrency for multi-delegate architecture.
- Strong typing in the safety-critical code.

### Negative

- Generics are still relatively new (Go 1.18+). We'll have to be careful in some places.
- The error-handling ceremony (`if err != nil`) is verbose. We mitigate with `errors.Wrap` and structured errors.
- Some LLM SDKs are Python-first (e.g., LangChain). We'll write our own thin Go clients.

### Neutral

- We commit to Go for the daemon. If we add a second daemon (e.g., a low-level perception daemon), it should also be Go unless there's a strong reason otherwise.

---

## Alternatives Considered

| Alternative | Pros | Cons | Why rejected |
|---|---|---|---|
| **Python (asyncio)** | Fast to write, LLM SDKs | GIL, slow, distribution | Slower, harder to distribute |
| **Rust** | Performance, safety | Velocity, ecosystem | Too slow to build v0.1.0 |
| **Node.js** | JS everywhere | Distribution, no real concurrency | Bad fit for the daemon |
| **C++** | Performance, control | Slow dev, complex build | Way too slow to build v0.1.0 |
| **Java/Kotlin** | Robust, JVM | Heavyweight, slow cold start | Bad fit for an on-device daemon |

---

## References

- [CLAUDE.md Section 5](../CLAUDE.md) — Stack decisions
- [CLAUDE.md Section 6](../CLAUDE.md) — Why Go + Python + TS
- [Go's concurrency is not parallelism](https://go.dev/blog/waza-talk)
- [libp2p in Go](https://github.com/libp2p/go-libp2p)
