# condura-app

> **The backend.** The daemon that owns the on-disk database, the OS keyring,
> the LLM registry, the safety layer, the agent loop, and the JSON-RPC IPC
> server that every client (desktop GUI, CLI, TUI, future Skills Hub) talks to.

This folder holds **all Go code** in the project. The desktop GUI's Wails shell
and the TUI binary are Go, so they live here too — under `cmd/` — even though
they are "user-facing." The Go `internal/` access rule forced this layout:
those binaries depend on `internal/*` packages, and only siblings of `internal/`
can import them.

## Layout

```
condura-app/
├── cmd/
│   ├── condurad/         # The daemon (background process)
│   ├── condura/          # The CLI client (JSON-RPC over HTTP/Unix socket)
│   ├── condura-tui/      # The terminal UI (Bubble Tea)
│   ├── condura-gui/      # The Wails desktop shell + wails.json
│   ├── gen-update-manifest/  # Release manifest generator
│   └── build_all_test.go # Meta-test that all four user-facing binaries compile
│
├── internal/             # The Go library surface (60+ packages)
│   ├── config/           # YAML loader + schema
│   ├── storage/          # SQLite + FTS5 + sqlite-vec
│   ├── ipc/              # JSON-RPC 2.0 + WebSocket
│   ├── llm/              # 12-provider LLM clients
│   ├── safety/           # Blast radius + Gatekeeper + audit
│   ├── autonomy/         # Autonomy matrix
│   ├── agent/            # Main loop + planner
│   ├── perception/       # Selective Perception
│   ├── computeruse/      # 4-tier computer-use router
│   ├── delegation/       # CLI spawning + sanitizers
│   ├── voice/            # STT + TTS + wake word
│   ├── memory/           # 3-layer memory
│   ├── skills/           # agentskills.io
│   ├── adaptive/         # User-Adaptive Engine
│   ├── sync/             # P2P sync
│   ├── reach/            # External channels (Telegram, etc.)
│   ├── hub/              # Hub client (v0.2.0)
│   ├── onboarding/       # First-run flow state machine
│   └── ... (45+ more)
│
├── configs/              # Default YAML, embedded via //go:embed
└── test/                 # Test fixtures (artifacts.json, checksums.txt, prebuilt/)
```

## Build & test

```bash
# Build all four user-facing binaries
make build                  # condurad, condura, condura-tui (into ./bin)
go build -o bin/condura-gui ./condura-app/cmd/condura-gui

# Run all tests with race detection
go test -race -count=1 ./...

# Verify (CI-equivalent gate)
make verify
```

## Where to look

| You want to… | Look in |
|---|---|
| Change daemon startup or config loading | `cmd/condurad/main.go`, `internal/config/` |
| Change IPC | `internal/ipc/` |
| Add an LLM provider | `internal/llm/` (extend `Provider` interface) |
| Change the safety layer | `internal/{blastradius,gatekeeper,halt,anomaly,audit,sanitize,sensitive,autonomy}/` |
| Change the agent loop | `internal/agent/` |
| Add a computer-use backend | `internal/computeruse/backends/` |
| Change the Wails desktop shell | `cmd/condura-gui/` |
| Change the TUI | `cmd/condura-tui/main.go` + `internal/tui/` |

## Module

This folder is the root of the Go module `github.com/sahajpatel123/conduraapp`
(`go.mod` lives at the repo root, but conceptually all Go in this tree is
"the app"). See `/` README for why we kept one Go module instead of one per
topic.