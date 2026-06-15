# Synaptic

> A free, OS-native AI agent that lives on your computer, orchestrates every AI tool you have, and costs you nothing.

Synaptic is a desktop app (Mac, Windows, Linux) that lets you summon an AI agent with a custom hotkey to control your computer — clicking, scrolling, typing, running sub-agents across Claude Code, Codex, Antigravity, OpenCode, Kilo, Hermes, and any ChatGPT Plus / Claude Pro / Gemini AI Pro / SuperGrok subscription you already have.

**Free forever. No lock-in. No tracking. No compromise on speed or safety.**

---

## Project Status (2026-06-15)

**Phases 1–13 implemented on `main`.** Phase 13 (release/distribution) includes
signed auto-update, GoReleaser packaging, DMG/NSIS GUI installers,
`release-verify` CI, and the `v0.1.0` release pipeline. On-device
verification (see `docs/on-device-verification.md`) is the final gate
before calling the public launch done.

| Layer            | Status |
|------------------|--------|
| Foundation docs  | ✅ done |
| Core daemon + IPC | ✅ done |
| Wails GUI shell  | ✅ done |
| Trust & Recovery (Phase 11) | ✅ backend + GUI wiring |
| Reach & Ecosystem (Phase 12) | ✅ TUI, i18n, hub, sync RPCs |
| Release / auto-update (Phase 13) | ✅ signed updates, GoReleaser, DMG/NSIS, `release-verify`, `v0.1.0` tag |
| Public launch sign-off | ⏳ on-device verification checklist |

### Try it locally
```bash
git clone https://github.com/sahajpatel123/synapticapp
cd synaptic
make build
./bin/synapticd --data-dir ./build/data &
./bin/synaptic --data-dir ./build/data ping
```

---

## Quickstart (planned v0.1.0)

1. **Download** the installer for your OS from [synaptic.app/download](https://synaptic.app/download).
2. **Install** — drag to Applications (mac), run installer (win), or `chmod +x` the AppImage (linux).
3. **Launch** Synaptic and complete the first-run onboarding:
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
- **Public Skills Hub** — share and discover workflows at [hub.synaptic.app](https://hub.synaptic.app)
- **6 languages** at launch — English, Spanish, French, German, Japanese, Mandarin
- **Safety first** — deterministic Gatekeeper, twin-snapshot verification, HMAC-chained audit log, behavioral anomaly detection, hardware kill switch

---

## Supported Subscriptions and Models

Synaptic works with **whatever you already have**:

- **ChatGPT Plus / Pro** — connect via OAuth, no separate API cost
- **Claude Pro / Max** — connect via OAuth
- **Gemini AI Pro / Ultra** — connect via Antigravity
- **SuperGrok** — connect via xAI OAuth
- **Any OpenAI-compatible API** — OpenAI, Anthropic, Mistral, DeepSeek, OpenRouter, Together, Groq, Fireworks, custom URL
- **Local models** — Ollama, LM Studio, vLLM, llama.cpp

Synaptic also delegates to these AI coding CLIs when installed:

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
- **GitHub Issues**: [github.com/sahajpatel123/synapticapp/issues](https://github.com/sahajpatel123/synapticapp/issues)
- **Email**: support@synaptic.app
- **Docs**: [synaptic.app/docs](https://synaptic.app/docs)

---

## Donations

Synaptic is free forever. If it makes your life better, consider donating:

- [GitHub Sponsors](https://github.com/sponsors/synapticapp)
- [Open Collective](https://opencollective.com/synaptic)
- [Stripe one-time](https://synaptic.app/donate)

---

## License

The Synaptic **binary** is free for personal and commercial use under the [Synaptic Freeware EULA v1](EULA.md).

The Synaptic **source code** is proprietary. All rights reserved.

© 2026 Synaptic.
