# condura-gui

> **All user-facing surfaces, except the marketing website.** The Svelte desktop
> UI and the TUI (which is also a Go binary, so it lives in `condura-app/cmd/`
> for `internal/` reasons — see below).

## Layout

```
condura-gui/
├── frontend/                  # The Svelte desktop UI
│   ├── package.json
│   ├── vite.config.ts
│   ├── src/
│   │   ├── lib/
│   │   │   ├── condura/       # Current-direction design system
│   │   │   ├── v2/            # Redesigned surfaces (the v2 ships)
│   │   │   ├── components/    # Legacy but still in use
│   │   │   └── routes/        # Page-level routes
│   │   ├── routes/            # Top-level routing
│   │   ├── stores/            # Svelte stores (state)
│   │   ├── tokens/            # (synced from condura-brand/tokens/)
│   │   └── utils/
│   ├── static/                # Locales, fonts, images
│   ├── wailsjs/               # Generated Wails bindings (regen on `wails build`)
│   └── assets/                # Embedded assets package for the Wails shell
│       ├── assets.go          #   `//go:embed all:dist`
│       └── dist/              #   Built frontend bundle (gitignored)
```

## Where the Go shell actually lives

The Wails shell's Go code (`main.go`, `app.go`, tray wiring, overlay
controller, quick prompt) lives at **`condura-app/cmd/condura-gui/`**, not in
this folder. The reason is Go's `internal/` access rule: the shell imports
`condura-app/internal/{conductor,daemon,config,...}`, and only siblings of
`internal/` (i.e., code under `condura-app/`) can import those packages.

The shell's `wails.json` and the embed-built `dist/` directory live with the
shell at `condura-app/cmd/condura-gui/`. The build script
(`condura-ops/scripts/build-gui.sh`) ties them together by passing
`-frontend ../../../condura-gui/frontend` to `wails build`.

If you came here looking for the Wails shell, go to
`condura-app/cmd/condura-gui/`.

## Build & run

**Package manager:** **npm only** for `frontend/` (and the rest of the monorepo’s
Node packages). Do not commit `pnpm-lock.yaml` here — dual lockfiles cause
supply-chain noise. `package.json` sets `"packageManager": "npm@10"`.

```bash
# Build the desktop app for the current OS
make build-gui

# Develop the frontend in isolation
cd condura-gui/frontend
npm install
npm run dev          # standalone Vite dev server
npm test             # vitest
```

## Where the TUI lives

The TUI's Go code lives at `condura-app/cmd/condura-tui/`, for the same
`internal/` reason. Conceptually it's "the terminal UI" and belongs to this
topic; structurally it has to live under `condura-app/`.

## Design tokens

Tokens (colors, spacing, typography) live in **`condura-brand/tokens/`** as
the source of truth. `make brand` regenerates the synced copy under
`condura-gui/frontend/src/lib/tokens/`. Edit `condura-brand/tokens/`, not
the local copy.