# Changelog

All notable changes to Condura are documented in this file.

The format is based on [Keep a Changelog](https://keepachangelog.com/en/1.1.0/),
and this project adheres to [Semantic Versioning](https://semver.org/spec/v2.0.0.html).

## [Unreleased]

### Added
- Phase 0: project foundation (CLAUDE.md, LOGBOOK.md, EULA.md, LICENSE, README.md, CONTRIBUTING.md, SECURITY.md, PRIVACY.md).
- Phase 0: architecture documentation (00-overview through 09-ipc).
- Phase 0: 5 architecture decision records (ADR-0001 through ADR-0005).
- Phase 0: AI onboarding guide and code style guide.
- **Phase 1: Repo Skeleton + Core Daemon.**
  - Bootstrap: Makefile, `.golangci.yml` (v2 schema), `.goreleaser.yml`, GitHub Actions CI (lint, test matrix on mac/win/linux × amd64/arm64, build matrix, integration, govulncheck).
  - `internal/version` — build metadata (ldflags).
  - `internal/logger` — slog wrapper with key+value redaction.
  - `internal/config` — YAML loader, env-override (`SYNAPTIC_<SEC>__<FIELD>`), `Validate()`.
  - `internal/secrets` — OS keyring (`zalando/go-keyring`) with file fallback; injectable backend.
  - `internal/storage` — `modernc.org/sqlite` (pure Go) + AES-256-GCM column-level encryption; schema v1.
  - `internal/api_key` — manager over storage + secrets; OAuth interface; Google PKCE.
  - `internal/llm` — `Provider` interface; OpenAICompat for 9 providers; dedicated Anthropic + Google impls; pricing registry + `EstimateCost`.
  - `internal/failover` — per-provider circuit breaker, breaker registry, daily spend monitor, chain runner, failover orchestrator.
  - `internal/health` — concurrent check aggregation.
  - `internal/ipc` — JSON-RPC 2.0 server, batch + notifications, HTTP + WebSocket transport, bearer-token auth, plus a JSON-RPC HTTP `Client`.
  - `cmd/synapticd` — daemon entry wiring config → logger → secrets → storage → api_key → LLM → failover → health → IPC; SIGINT/SIGTERM handling; sidecar `<data_dir>/synapticd.addr`; Unix socket on non-Windows.
  - `cmd/synaptic` — CLI client with `ping`, `version`, `status`, `config`, `llm chat|providers`, `apikeys list|set|delete`.
- Test coverage on every internal package exceeds 80% (10/10 packages).

### Changed
- `.gitignore` now ignores `synapticd` and `synaptic` binaries in the repo root (use `bin/` for builds).

### Known issues
- 416 pre-existing golangci-lint v2 issues remain (mostly `errcheck` on `defer x.Close()` patterns, `goconst`, `mnd` in non-test code). Tracked for a follow-up "lint hygiene" pass before v0.1.0.
- `cmd/synaptic` `--stream` is a no-op in Phase 1; streaming will be added in Phase 2 (per-Provider `Stream()` is already implemented in the LLM package).
- `Makefile` `daemon-init` / `daemon-stop` targets call into CLI subcommands that don't exist yet; they will be added with the install / LaunchAgent work in Phase 5.

[Unreleased]: https://github.com/sahajpatel123/conduraapp/compare/main...HEAD
