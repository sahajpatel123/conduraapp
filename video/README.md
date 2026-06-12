# Synaptic — Demo Video

A ~61-second cinematic demo for **Synaptic — The Free AI Conductor**, built as
code with [Remotion](https://remotion.dev) (React → real MP4). It renders to the
brief's delivery spec: **MP4 / H.264 / 1080p / 30 fps**, `yuv420p`, CRF 18.

The film shares the website's design system ("The Touch"): the Ink / Ivory /
Brass palette, Archivo + Instrument Serif + Geist, the hanging-bulb motif, staff
lines, film grain, dust motes, and the single ease curve
`cubic-bezier(0.16, 1, 0.3, 1)` — all lifted from `web/`.

## Quick start

```bash
cd video
npm install
node scripts/fetch-fonts.mjs   # self-host the fonts (one-time; needs network)
npm run dev                    # open Remotion Studio to preview/scrub
npm run render                 # → out/synaptic-demo.mp4
npm run still                  # → out/poster.png (a poster frame)
```

> `scripts/fetch-fonts.mjs` downloads the four font families into
> `public/fonts/` and writes `public/fonts.css`, so the render needs **no
> network** (and works behind a TLS-intercepting proxy that headless Chrome
> won't trust).

## Storyboard → scenes

Each second-by-second beat from the brief is one component in `src/scenes/`,
sequenced on the master timeline in `src/timeline.ts`.

| Scene | Time | What it shows |
|-------|------|---------------|
| `Ignition` | 0.0–3.2s | A single filament catches; the title settles. |
| `Problem` | 3.2–7.6s | Scattered AI tools, a frantic cursor, the pains. |
| `Touch` | 7.6–12.6s | A hand ignites the bulb; the room lights up. "One hotkey. Unlimited AI." |
| `Hotkey` | 12.6–18.4s | The overlay + a `ping → pong`; cold-start / hotkey latencies count up. |
| `Voice` | 18.4–24.4s | A spoken command, live waveform, local-STT privacy tags. |
| `Perception` | 24.4–30.6s | The pipeline lights stage by stage → delegation fan-out. |
| `Montage` | 30.6–38.4s | PDF summary → email draft → calendar, each stamped with a check. |
| `Audit` | 38.4–44.4s | HMAC-chained log scrolls; the four kill-switches arm. |
| `Features` | 44.4–50.8s | P2P sync, action replay, skills hub, adaptive engine. |
| `Cta` | 50.8–57.8s | The mark, the orchestra question, three doors. |
| `Outro` | 57.8–61.0s | The filament dims; `synaptic.app`. |

## Audio

The visuals carry the narrative through kinetic captions, so the MP4 stands
alone. To add the voiceover + score, see **`VOICEOVER.md`** — drop
`public/voiceover.mp3` and `public/music.mp3` and they're picked up
automatically (no code change). The exact narration, timing table, and the
audio mix spec live there.

## Layout

```
video/
  src/
    Root.tsx           Composition registration (1920×1080, 30fps, 1830 frames)
    DemoVideo.tsx      Master timeline, grain/vignette/baton, optional audio
    timeline.ts        Scene [from, duration] map
    theme.ts           Palette + ease + font tokens (mirrors web/globals.css)
    fonts.ts           Self-hosted @font-face loader
    components/        Bulb, Hand, Background, Chrome (windows/overlay), Primitives
    scenes/            One file per storyboard beat
  scripts/fetch-fonts.mjs   Font self-hosting
  public/fonts/             Downloaded woff2 (+ fonts.css)
```
