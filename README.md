# Condura

> A free, OS-native AI agent that lives on your computer, orchestrates every AI tool you have, and costs you nothing.

Condura is a desktop app (Mac, Windows, Linux) that lets you summon an AI agent with a custom hotkey to control your computer — clicking, scrolling, typing, running sub-agents across Claude Code, Codex, Antigravity, OpenCode, Kilo, Hermes, and any ChatGPT Plus / Claude Pro / Gemini AI Pro / SuperGrok subscription you already have.

**Free forever. No lock-in. No tracking. No compromise on speed or safety.**

---

## Project Status (2026-06-15)

**Phase 13 (release/distribution) is complete.** [v0.1.0](https://github.com/sahajpatel123/conduraapp/releases/tag/v0.1.0) is published with signed auto-update (`manifest.json`), GoReleaser packages, and GUI installers (DMG / portable exe / Linux binary). `release-verify` CI runs on every `main` push. On-device verification (`docs/on-device-verification.md`) is the separate public-launch gate.

| Layer            | Status |
|------------------|--------|
| Foundation docs  | ✅ done |
| Core daemon + IPC | ✅ done |
| Wails GUI shell  | ✅ done |
| Trust & Recovery (Phase 11) | ✅ backend + GUI wiring |
| Reach & Ecosystem (Phase 12) | ✅ TUI, i18n, hub, sync RPCs |
| Release / auto-update (Phase 13) | ✅ **complete** — [v0.1.0](https://github.com/sahajpatel123/conduraapp/releases/tag/v0.1.0) live |
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

## Quickstart (v0.1.0)

1. **Download** from [GitHub Releases v0.1.0](https://github.com/sahajpatel123/conduraapp/releases/tag/v0.1.0) or [condura.app/download](https://condura.app/download):
   - **macOS:** `condura-gui-darwin-arm64.dmg`
   - **Windows:** `condura-gui-windows-amd64.exe` (or `-setup.exe` when present)
   - **Linux:** `condura-gui-linux-amd64` or `condurad_*_linux_amd64.deb`
2. **Install** — drag to Applications (mac), run installer/portable exe (win), or `chmod +x` the binary (linux).
3. **Launch** Condura and complete the first-run onboarding:
   - Accept the EULA
   - Connect a subscription (ChatGPT Plus, Claude Pro, Gemini AI Pro, SuperGrok) **or** paste an API key **or** use a local model (Ollama)
   - Grant OS permissions (Accessibility, Microphone, Screen Recording)
   - Pick a hotkey
4. **Tap your hotkey** → overlay appears. Type or speak a task. Done.

---

## Features

- **Custom global hotkey** — summon the agent from anywhere
- **Voice input** with local Whisper STT and the wake word "hey synaptic"
- **3-tier computer use** — Accessibility API first, background-first automation, vision CUA as last resort
- **Smart router** — picks the best model for each task across 12 LLM providers and 8 sub-agent CLIs
- **User-Adaptive Engine** — learns your style, preferences, and habits over time
- **P2P encrypted sync** — your memory and skills sync across your devices, no central server
- **Action replay** — scrubbable 24h timeline of everything the agent did
- **Public Skills Hub** — share and discover workflows at [hub.condura.app](https://hub.condura.app)
- **6 languages** at launch — English, Spanish, French, German, Japanese, Mandarin
- **Safety first** — deterministic Gatekeeper, twin-snapshot verification, HMAC-chained audit log, behavioral anomaly detection, hardware kill switch

---

## Supported Subscriptions and Models

Condura works with **whatever you already have**:

- **ChatGPT Plus / Pro** — connect via OAuth, no separate API cost
- **Claude Pro / Max** — connect via OAuth
- **Gemini AI Pro / Ultra** — connect via Antigravity
- **SuperGrok** — connect via xAI OAuth
- **Any OpenAI-compatible API** — OpenAI, Anthropic, Mistral, DeepSeek, OpenRouter, Together, Groq, Fireworks, custom URL
- **Local models** — Ollama, LM Studio, vLLM, llama.cpp

Condura also delegates to these AI coding CLIs when installed:

Claude Code · Codex · Antigravity · OpenCode · Kilo Code · Hermes Agent · Gemini CLI

---

## Documentation

- [CLAUDE.md](CLAUDE.md) — full project spec for AI agents and contributors
- [LOGBOOK.md](LOGBOOK.md) — append-only session log
- [EULA.md](EULA.md) — license terms
- [PRIVACY.md](PRIVACY.md) — privacy policy
- [docs/architecture/](docs/architecture/) — architecture deep-dives
- [docs/adr/](docs/adr/) — architecture decision records
- [docs/guides/](docs/guides/) — how-to guides

---

## Support

- **Discord**: [discord.gg/synaptic](https://discord.gg/synaptic)
- **GitHub Issues**: [github.com/sahajpatel123/conduraapp/issues](https://github.com/sahajpatel123/conduraapp/issues)
- **Email**: support@condura.app
- **Docs**: [condura.app/docs](https://condura.app/docs)

---

## Donations

Condura is free forever. If it makes your life better, consider donating:

- [GitHub Sponsors](https://github.com/sponsors/conduraapp)
- [Open Collective](https://opencollective.com/synaptic)
- [Stripe one-time](https://condura.app/donate)

---

## License

The Condura **binary** is free for personal and commercial use under the [Condura Freeware EULA v1](EULA.md).

The Condura **source code** is proprietary. All rights reserved.

© 2026 Condura.
