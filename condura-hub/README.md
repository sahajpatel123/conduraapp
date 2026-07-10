# condura-hub

> **Reserved.** This folder is a placeholder for the public **Skills Hub** —
> the curated, safety-scanned, versioned marketplace for community-contributed
> skills. See `condura-mind/CLAUDE.md` §15 (Skills System) and §18 (Skills Hub)
> for the full design.

## Status

Not yet implemented. The hub is targeted for **v0.2.0** per
`condura-mind/docs/roadmap-v0.2.0.md`.

When work begins, this folder will hold the **Next.js** hub application
(`hub.condura.app`), with:

- Skill upload + scan pipeline (promptware detection, secret scanning)
- Versioned skill registry with semver + signed releases
- Trust levels (official / community / experimental)
- User subscriptions + update flow
- Moderation queue

## How to start work here

Until v0.2.0 begins, treat this folder as **read-only scaffolding**. Any
experimentation should live in `condura-ui/_experiments/` or a local branch.

When the hub is built, the connections will be:

```
condura-hub/                (Next.js app, public)
        ↑
        │  shares brand, design tokens, copy patterns
        │
condura-brand/              ← condura-hub imports from here
condura-mind/docs/          ← condura-hub links to docs
```

## Related folders

- `condura-ui/` — the marketing website (sister Next.js app)
- `condura-brand/` — shared visual identity (tokens, logos, fonts)
- `condura-ops/deploy/` — future hub deployment configs will live there
- `condura-mind/docs/roadmap-v0.2.0.md` — full v0.2.0 plan