# condura-brand

> **Visual identity & shared marketing assets.** The single source of truth
> for tokens, fonts, logos, palettes, and motion. Both `condura-gui/` (the
> desktop UI) and `condura-ui/` (the marketing site) import from here.

## Layout

```
condura-brand/
├── tokens/      # JSON/TS/CSS design tokens (colors, spacing, type scale)
├── fonts/       # Font binaries (Inter, IBM Plex Mono, …)
├── logos/       # SVG + PNG brand marks
├── palette/     # Color palette references (light/dark)
├── motion/      # Easing curves, durations, motion language
└── assets/      # Hero images, demo film stills, renders from condura-studio/
```

## Why a separate topic?

Design tokens drift. Without a single source, the desktop app's accent color
ends up a different hex than the marketing site's, and the spacing rhythm
gets inconsistent. By making `condura-brand/` the canonical owner:

- **One token file** flows into both frontends via `make brand`
- **One hero image** is committed once and consumed by both UIs
- **Font binaries** live in one place; both frontends reference them

## Workflow: change a token

1. Edit the token in `condura-brand/tokens/condura.css` (or `.json` / `.ts`)
2. Run `make brand` — this regenerates synced copies at:
   - `condura-gui/frontend/src/lib/tokens/`
   - `condura-ui/app/(appropriate import)`
3. Commit both the source change and the synced copies
4. The CI lint step (when added) catches drift between source and synced

## Workflow: add a font

1. Drop the binary into `condura-brand/fonts/<family>/`
2. Add the `@font-face` rule to `condura-brand/tokens/condura.css`
3. Run `make brand` — both frontends pick it up

## When this folder grows

If `condura-brand/` accumulates >100 assets, consider splitting it into:

```
condura-brand/
├── tokens/              # small, fast — kept here
├── fonts/               # small — kept here
└── archive/             # older logos, deprecated palettes, etc.
```

Don't split until it's actually a maintenance burden.