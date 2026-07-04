# condura-studio

> **Video / film creation.** Remotion-based projects that produce hero
> videos, demo films, and brand artifacts for the marketing site.

Each subfolder is one Remotion project. They share no code with the rest of
the Condura repo; they are independent Node projects that produce MP4 / GIF
output that lands in `condura-brand/assets/` for consumption by
`condura-ui/`.

## Current projects

| Path | Status | Output |
|---|---|---|
| `condura-demo/` | shipping | 60s product demo film |
| `my-video/` | exploration | earlier brand experiments |

## Develop

```bash
cd condura-studio/condura-demo
npm install
npm run dev          # Remotion Studio (browser-based editor)
npm run build        # render to .mp4
```

## Adding a new project

1. `mkdir condura-studio/<project-name> && cd $_`
2. `npm init -y && npm install remotion @remotion/cli`
3. Create `src/` with your `Composition` components
4. Add a `README.md` to the project describing what it produces
5. When the output is final, copy the rendered video to
   `condura-brand/assets/` so the marketing site can pick it up

## Output policy

Final outputs belong in `condura-brand/assets/`. Intermediate renders,
`node_modules/`, and `build/` directories are gitignored.