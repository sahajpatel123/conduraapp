# condura-ui

> **The marketing website.** Next.js 14 on Vercel. Landing page, manifesto,
> changelog, legal pages, download links.

Lives at `synaptic.app` (in production) / `localhost:3000` (in dev).

## Layout

```
condura-ui/
├── app/                  # Next.js App Router (pages)
├── components/           # Shared React components
├── lib/                  # Site config + utilities
├── public/               # Static assets (served as-is)
├── hooks/                # React hooks
├── types/                # TypeScript types
├── context/              # React context providers
├── tests/                # Playwright / Vitest (when added)
├── _experiments/         # Archived marketing film HTMLs (kebab-case)
│                         #   - condura-demo-film.html
│                         #   - condura-launch-film.html
│                         #   - condura-masterpiece-film.html
│                         #   - condura-pre-launch.html
│                         #   - condura-ad-video.html
│                         #   - condura-gui-preview.html
│                         #   - condura-ritual-mockup.html
├── package.json
├── next.config.ts
├── tsconfig.json
└── eslint.config.mjs
```

## Develop

```bash
cd condura-ui
npm install
npm run dev          # http://localhost:3000
npm run build        # production build
npm run lint
```

## Deploy

Configured via Vercel. See `condura-ops/deploy/` for the Vercel project
config and DNS records (when added in v0.2.0).

## Changelog source

The **Changelog** page renders from `condura-mind/CHANGELOG.md` — keep that
file the canonical source.

## Marketing experiments

The `_experiments/` folder holds the standalone HTML landing pages we built
during the brand exploration phase. They are tracked in git as design history.
New experiments should go here, never at the repo root.

## See also

`condura-ui/AGENTS.md` and `condura-ui/CLAUDE.md` warn that this is a
non-standard Next.js build — read them before writing any Next.js code here.