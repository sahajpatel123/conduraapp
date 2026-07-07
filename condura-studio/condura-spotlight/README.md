# condura-spotlight

> **AI-authored Remotion films for Condura.** The flagship project for
> promotional video output. Independent Node project; render artifacts land in
> `out/` and are then promoted to `condura-brand/assets/`.

---

## Flagship film — ConduraConcerto

*5-beat, 20-second editorial. Paper, ink, and a single warm accent.*

A baton line draws horizontally across a warm-paper canvas. Scattered AI-tool
dots (Claude, GPT, Gemini, Ollama, Codex, Antigravity, Cursor, OpenCode, Kilo)
fade in, then snap into formation as the baton passes — order from chaos. A
single downbeat ripple. The brand resolves with quiet confidence.

**Visual language**: warm paper (`#f5efe3`), ink (`#1b1a17`), one pollen accent
(`#c18a4a`). Georgia serif, monospace labels, generous letter-spacing.
Generous negative space. No neon, no glitch, no HUD. Slow Bézier easing,
spring-driven micro-motions. 1.5px ink line, 11px pollen baton tip, 13px tool
dots. Dark to warm accent on alternating dots. Paper grain texture (multiply
blend, 16% opacity) adds tactile depth.

| Composition | 1920×1080 (16:9) | 20s @ 30fps | MP4 (1.3 MB) |
|---|---|---|---|

```sh
open /Users/sahajpatel/synaptic/condura-brand/assets/condura-concerto.mp4
```

### Story beats

1. **Silence** (0–2.5s): Empty paper. A faint 1px ink hairline fades in at
center. A pollen baton-tip sits quietly at left.
2. **Scatter** (2.5–5s): In scattered positions, the tool dots fade in — subtle,
gentle drift. Labels appear only as each dot aligns.
3. **The Baton** (5–12s): The tip travels left→right, drawing the line. As it
passes each dot, the dot snaps to the line. Order from chaos.
4. **Concert** (12–16.5s): All nine dots in formation. A subtle downbeat
ripple (center-out scale pulse). "Your tools, *in concert.*" appears in large
serif, *in concert* in pollen italic.
5. **Resolve** (16.5–20s): Line and dots contract. "Condura" resolves in a
large serif lock-up. Tagline, then `condura.app`.

---

## Earlier experiments (archived)

| Composition | Format | Duration | Output |
|---|---|---|---|
| `ConduraHypeReel` | 1080×1920 (9:16) | 22s | `condura-brand/assets/condura-hype-reel.mp4` |
| `ConduraMasterpiece` | 1920×1080 (16:9) | 33s | `condura-brand/assets/condura-masterpiece.mp4` |

These were neon/maximalist direction experiments. Code is kept for reference but
the Concerto above is the current launch film.

---

## Render commands

```sh
npm install

# Start Remotion Studio (edit any composition interactively)
npm run studio

# Render the flagship
npm run render:concerto       # → out/condura-concerto.mp4
npm run still:concerto          # → out/condura-concerto-cover.png

# Render older experiments
npm run render:hype             # → out/condura-hype-reel.mp4
npm run still:hype
npm run render:masterpiece      # → out/condura-masterpiece.mp4
npm run still:masterpiece

# Typecheck only
npm run typecheck
```

## Adding a film here

1. Create `src/<YourFilm>.tsx` and a matching `src/<your-film>.css`.
2. Register a `<Composition>` in `src/Root.tsx`.
3. Add `render:<id>` and `still:<id>` scripts to `package.json`.
4. Document the storyt beats above.
5. When output is final, render and move the MP4 / PNG to
   `condura-brand/assets/` so the marketing site can pick it up.

## Output policy

Final outputs belong in `condura-brand/assets/`. Intermediate renders,
`node_modules/`, and `out/` are gitignored.