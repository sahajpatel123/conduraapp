# condura-ops

> **Everything that operates the Condura factory.** CI, scripts, release
> tooling, deployment configs. None of this ships to users; it makes the
> product, signs it, tests it, and ships it.

## Layout

```
condura-ops/
├── ci/
│   ├── workflows/      # GitHub Actions (ci, release, release-verify, …)
│   ├── CODEOWNERS
│   └── dependabot.yml
│
├── scripts/
│   ├── install.sh            # Cross-platform install
│   ├── build-gui.sh          # `make build-gui` runner (Wails + embeds)
│   ├── package-gui-installers.sh   # DMG / NSIS packaging
│   └── verify-release-artifacts.sh # Post-publish artifact verification
│
├── release/
│   ├── goreleaser.yml        # Single source for cross-platform binaries
│   ├── condura-gui.nsi       # NSIS installer config (Windows)
│   └── homebrew/             # Homebrew formula + cask
│
├── deploy/             # Per-target deployment configs
│   ├── web/            # (v0.2.0) Vercel/KV setup for condura-ui
│   └── hub/            # (v0.2.0) Vercel/KV setup for condura-hub
│
└── monitoring/         # Future: observability configs
```

## Key workflows

| When | Run |
|---|---|
| Local full check | `make verify` (at repo root) |
| Local GUI build | `make build-gui` |
| Local release dry run | `make release-snapshot` |
| Tag a release | `git tag vX.Y.Z && git push --tags` — CI runs release.yml |
| Verify a release tag | `make verify-release TAG=vX.Y.Z` |
| Install on a target machine | `condura-ops/scripts/install.sh` |

## CI

Workflows are YAML files in `condura-ops/ci/workflows/`. They reference
paths in their respective topics (`condura-app/`, `condura-gui/`,
`condura-ui/`, `condura-ops/`). When you move code between topics, update
the `paths:` filters on the affected workflows.

## Adding a deploy target

1. `mkdir condura-ops/deploy/<target>`
2. Add the deploy config (vercel.json, Terraform, k8s manifest, etc.)
3. Add a workflow under `condura-ops/ci/workflows/` that triggers on tag
4. Document the runbook in `condura-ops/deploy/<target>/README.md`