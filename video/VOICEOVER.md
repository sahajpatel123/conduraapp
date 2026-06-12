# Synaptic — Demo Voiceover Script

Narration for the ~61s demo (`src/DemoVideo.tsx`). Timed to the scene cues so
a separately-produced VO track drops straight in. Record at 48 kHz, mix to
roughly **−12 LUFS**; the on-screen captions mirror these lines.

To enable audio in the render, drop the files in `public/`:

- `public/voiceover.mp3` — the narration below
- `public/music.mp3` — the score (cinematic synth pad → arpeggio → climax)

`src/DemoVideo.tsx` mounts them automatically when present (VO at full volume,
music at ~0.32).

| Time | Scene | Line |
|------|-------|------|
| 0.0–3.2s | Ignition | "Imagine every AI you own — working together. On your computer. For free." |
| 3.2–7.6s | Problem | "Today, AI power is locked behind subscriptions, separate apps, and endless context-switching." |
| 7.6–12.6s | Touch | "Synaptic changes that. One hotkey. One agent. It unlocks every AI you already have." |
| 12.6–18.4s | Hotkey | "Press your hotkey, and the overlay appears instantly — ready to listen." |
| 18.4–24.4s | Voice | "Speak naturally. Synaptic uses local Whisper speech-to-text — no cloud cost, zero latency." |
| 24.4–30.6s | Perception | "It perceives only what's needed, verifies the screen hasn't changed, checks the Gatekeeper, then routes the task to the best AI tool." |
| 30.6–38.4s | Montage | "From summarizing documents to drafting emails and scheduling meetings, Synaptic orchestrates every step — keeping you in the flow." |
| 38.4–44.4s | Audit | "Every action is recorded in a tamper-proof audit log. And you keep full control — hard hotkey, watchdog, network isolation, menu-bar kill." |
| 44.4–50.8s | Features | "Plus peer-to-peer encrypted sync, action replay, a public skills hub, and an adaptive engine that learns your style — all local, all free." |
| 50.8–57.8s | CTA | "Ready to conduct your own AI orchestra? Join the wait-list, chat with the community, and star the repo. Synaptic — free AI, yours to command." |
| 57.8–61.0s | Outro | *(silence, then a single soft chime as the filament dims.)* |

## Audio spec (from the brief)

- **Voice:** professional, neutral, ~150 wpm, clear diction.
- **Music:** original electronic-cinematic, ~120 BPM. Ambient intro → tension →
  build → uplifting climax → resolution. Royalty-free or composed.
- **SFX (low in the mix):** soft click on hotkey press; whoosh on the light
  sweep; waveform blip on voice input; check-tick on completed actions;
  low pulse on audit-log entries; gentle chime at the end.
- **Levels:** narration ~−12 LUFS, music ~−18 LUFS, true-peak ≤ −1 dBTP.
