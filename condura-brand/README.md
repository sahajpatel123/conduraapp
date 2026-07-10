# condura-brand

> **Visual identity & shared marketing assets** for Condura.
> Honest scope: this is a **starter brand kit**, not a full agency package.

## What ships today

| Path | Status | Notes |
|---|---|---|
| `tokens/` | **Canonical** | Primitives, semantic, motion CSS/TS — seeded from the desktop GUI tokens |
| `palette/` | **Reference** | Named brand colors (paper / ink / plum) for designers |
| `motion/` | **Reference** | Motion grammar summary for film + marketing |
| `logos/` | **Minimal** | `condura-mark.svg` monogram only — not a full logo suite |
| `fonts/` | **Policy only** | No binary fonts; system stacks + app-side loading (see `fonts/README.md`) |
| `assets/` | **Media** | Hero stills + rendered studio films (MP4s gitignored; see `.gitignore`) |

## What this is NOT

- Not a complete logo system (no wordmark lockups, no app-icon set, no favicon pack).
- Not a vendored font library.
- Not auto-synced into frontends yet (`make brand` is aspirational — see below).

Desktop tokens that the Svelte app actually loads still live under
`condura-gui/frontend/src/lib/tokens/`. The files in `condura-brand/tokens/`
are the **shared source of truth** we keep in lockstep by hand until a
sync target lands.

## Layout

```
condura-brand/
├── tokens/      # CSS/TS design tokens (colors, spacing, motion)
├── fonts/       # Policy README — no binaries today
├── logos/       # Minimal SVG mark
├── palette/     # Color palette reference
├── motion/      # Motion language reference
└── assets/      # Hero images, demo film stills / renders
```

## Workflow: change a token

1. Edit the matching file in **both** `condura-brand/tokens/` and
   `condura-gui/frontend/src/lib/tokens/` (keep them identical until sync exists).
2. If marketing CSS needs the same value, update `condura-ui` accordingly.
3. Commit both sides in one change.

## When `make brand` lands

A future Makefile target will copy `condura-brand/tokens/*` into the GUI
and marketing import paths and fail CI on drift. Until then, dual-edit.

## Adding a real logo suite later

1. Place SVGs under `logos/` with a short `logos/README.md` describing usage.
2. Export PNGs only if a surface cannot consume SVG.
3. Update this README's status table — do not claim "full brand kit" until
   wordmark + app icon + favicon + mark all exist.
