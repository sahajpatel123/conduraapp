# Condura

> A free, OS-native AI agent that lives on your computer, orchestrates your AI tools, and costs nothing.

Condura is a desktop agent you summon with a single hotkey. It runs sub-agents across the AI coding CLIs you already have installed, talks to 15 LLM backends via your own API keys or a local Ollama, and never costs you a subscription.

**Free forever. No lock-in. No tracking. No compromise on speed or safety.**

---

## Platform Support & v0.1.1 Status

Condura is honest about what runs today.

| Platform | What works in v0.1.1 |
|---|---|
| **macOS** | Full feature set — GUI overlay, voice + wake word, computer-use (Accessibility / vision), and delegation. Requires macOS 13+, Apple silicon (Intel via Rosetta). |
| **Windows** | Daemon + CLI + TUI. **GUI overlay is v0.2.0.** Windows 10+, x64. |
| **Linux** | Daemon + CLI + TUI. **GUI overlay is v0.2.0.** glibc 2.31+, x64. |

### What works in v0.1.1

- One global hotkey summons the daemon from anywhere on macOS
- 15 LLM backends — 11 cloud APIs (Anthropic, OpenAI, Google, xAI, Mistral, DeepSeek, OpenRouter, Groq, Together, Fireworks, plus a Custom OpenAI-compatible slot) plus 4 local runtimes (Ollama, LM Studio, vLLM, LocalAI)
- 8 sub-agent CLIs with auto-detection from your `$PATH` — Claude Code, Codex, Antigravity, OpenCode, Kilo, Hermes, Gemini CLI, and Ollama
- **Deterministic Gatekeeper** — code, not a model; every computer-use action and shell command passes through it
- HMAC-chained, append-only audit log
- AES-256-GCM encrypted secrets at rest
- Shell command sanitization before anything is run
- Twin-snapshot computer-use verification (macOS)
- **Talk → act tool dispatch** (N2 Path A, just shipped in v0.1.1) — spoken/typed tasks can drive computer-use directly from chat
- Voice input with local Whisper STT and the wake word **"hey condura"**
- Auto-backup and Ed25519-signed auto-update
- Telegram channels
- 6 languages at launch — English, Spanish, French, German, Japanese, Mandarin

### Deferred to v0.2.0 (not in v0.1.1)

These are designed and specced, but **not shipped yet**. They are tracked in [`docs/roadmap-v0.2.0.md`](docs/roadmap-v0.2.0.md):

- Hybrid LLM router (TaskSpec, cost-first cascade, memory bias)
- Subscription OAuth for ChatGPT Plus, Claude Pro, SuperGrok
- Parallel wave / DAG orchestration across sub-agents (v0.1.1 spawns individual sub-agents only)
- Vector embeddings / semantic recall
- Public Skills Hub at `hub.condura.app`
- WhatsApp / Signal channels (Telegram works today)
- Hard Layer 3 network guard as a separate OS process (`pf` / `netsh`); v0.1.1 ships an in-process guard instead
- GUI overlays for Windows and Linux

---

## Project Status (2026-06-27)

Phase 13 (release/distribution) shipped with signed auto-update (`manifest.json`), GoReleaser packages, and macOS installers (DMG). [v0.1.0](https://github.com/sahajpatel123/conduraapp/releases/tag/v0.1.0) is published; [v0.1.1](https://github.com/sahajpatel123/conduraapp/releases/tag/v0.1.1) adds fail-closed destructive consent, watchdog-by-default, talk → act dispatch, and the honest marketing/README pass. `release-verify` CI runs on every `main` push. On-device verification (`docs/on-device-verification.md`) is the separate public-launch gate.

| Layer            | Status |
|------------------|--------|
| Foundation docs  | ✅ done |
| Core daemon + IPC | ✅ done |
| Wails GUI shell (mac) | ✅ done |
| Trust & Recovery (Phase 11) | ✅ backend + GUI wiring |
| Reach & Ecosystem (Phase 12) | ✅ TUI, i18n, sync RPCs |
| Release / auto-update (Phase 13) | ✅ complete — v0.1.0 live |
| N2 Path A: talk → act dispatch | ✅ shipped in v0.1.1 |
| Public launch sign-off | ⏳ on-device verification checklist (manual) |

### Try it locally
```bash
git clone https://github.com/sahajpatel123/conduraapp
cd synaptic
make build
./bin/synapticd --data-dir ./build/data &
./bin/synaptic --data-dir ./build/data ping
```

---

## Quickstart (v0.1.1)

1. **Download** from [GitHub Releases](https://github.com/sahajpatel123/conduraapp/releases) or [condura.app/download](https://condura.app/download):
   - **macOS:** `condura-gui-darwin-arm64.dmg` (full GUI overlay)
   - **Windows:** `condura-cli-windows-amd64.zip` (CLI + TUI today; GUI is v0.2.0)
   - **Linux:** `condura-cli-linux-amd64` / `condurad_*_linux_amd64.deb` (CLI + TUI today; GUI is v0.2.0)
2. **Install** — drag to Applications (mac), unzip and add to `PATH` (win/linux).
3. **Launch** Condura (mac) or run the daemon (win/linux) and complete first-run onboarding:
   - Accept the EULA
   - Grant OS permissions (macOS only — Accessibility + Screen Recording). Microphone is requested later from Settings when you enable voice.
   - Pick a global hotkey (no default per locked decision #8)
   - Optionally connect a provider in Settings, or use the local Ollama instance if it's already running
4. **Tap your hotkey** → overlay appears. Type or speak a task. Done.

> **Providers in v0.1.1:** Condura accepts API keys for 15 LLM backends
> (Anthropic, OpenAI, Google, xAI, Mistral, DeepSeek, OpenRouter, Groq,
> Together, Fireworks, plus a Custom OpenAI-compatible slot — and local
> Ollama, LM Studio, vLLM, LocalAI). You
> bring the keys, or point it at a running Ollama. Subscription OAuth
> (ChatGPT Plus, Claude Pro, SuperGrok) is **v0.2.0** — see
> `docs/roadmap-v0.2.0.md`. v0.1.1 uses the single provider configured in
> Settings; the hybrid router that picks the best model per turn is v0.2.0.

---

## Features

- **Custom global hotkey** — summon the agent from anywhere (macOS GUI; win/linux TUI)
- **Voice input** with local Whisper STT and the wake word **"hey condura"**
- **3-tier computer use (macOS)** — Accessibility API first, background-first automation, vision CUA as last resort. Windows/Linux get computer-use when their GUI overlay ships in v0.2.0.
- **Talk → act tool dispatch** — spoken/typed tasks drive computer-use directly from chat (N2 Path A, shipped in v0.1.1)
- **Sub-agent delegation** — spawn Claude Code, Codex, Antigravity, OpenCode, Kilo, Hermes, Gemini CLI, or Ollama. Parallel wave/DAG orchestration is v0.2.0.
- **Deterministic Gatekeeper** — code, not a model; the only path to a click, keystroke, or shell command
- **User-Adaptive Engine** — learns from your interactions, fully editable in Settings
- **Action replay** — scrubbable timeline of everything the agent did
- **6 languages** at launch — English, Spanish, French, German, Japanese, Mandarin
- **Safety first** — deterministic Gatekeeper, twin-snapshot verification (mac), HMAC-chained audit log, AES-256-GCM encrypted secrets, shell sanitization, in-process network guard (hard OS-process guard in v0.2.0)

---

## Supported Models

Condura works with **whatever you already have**:

- **Any OpenAI-compatible API key** — Anthropic, OpenAI, Google, xAI, Mistral, DeepSeek, OpenRouter, Groq, Together, Fireworks, or a custom URL
- **Local models** — Ollama, LM Studio, vLLM (and LocalAI)
- **Subscription billing** (ChatGPT Plus, Claude Pro, SuperGrok) — *v0.2.0; v0.1.1 uses API keys only*

Condura also delegates to these AI coding CLIs when installed (auto-detected from your `$PATH`):

Claude Code · Codex · Antigravity · OpenCode · Kilo · Hermes Agent · Gemini CLI · Ollama

---

## Documentation

- [CLAUDE.md](CLAUDE.md) — full project spec for AI agents and contributors
- [LOGBOOK.md](LOGBOOK.md) — append-only session log
- [docs/roadmap-v0.2.0.md](docs/roadmap-v0.2.0.md) — what's deferred from v0.1.1 to v0.2.0 and why
- [EULA.md](EULA.md) — license terms
- [PRIVACY.md](PRIVACY.md) — privacy policy
- [docs/architecture/](docs/architecture/) — architecture deep-dives
- [docs/adr/](docs/adr/) — architecture decision records
- [docs/guides/](docs/guides/) — how-to guides

---

## Support

- **Discord**: [discord.gg/condura](https://discord.gg/condura)
- **GitHub Issues**: [github.com/sahajpatel123/conduraapp/issues](https://github.com/sahajpatel123/conduraapp/issues)
- **Email**: support@condura.app
- **Docs**: [condura.app/docs](https://condura.app/docs)

---

## Donations

Condura is free forever. If it makes your life better, consider donating:

- [GitHub Sponsors](https://github.com/sponsors/conduraapp)
- [Open Collective](https://opencollective.com/condura)
- [Stripe one-time](https://condura.app/donate)

---

## License

The Condura **binary** is free for personal and commercial use under the [Condura Freeware EULA v1](EULA.md).

The Condura **source code** is proprietary. All rights reserved.

© 2026 Condura.