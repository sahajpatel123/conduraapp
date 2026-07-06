# CI Workflows Live at `condura-ops/ci/workflows/`

> GitHub's Actions tab is configured to discover workflows under
> `.github/workflows/` by default, but per the **2026-07-04 topic-sliced
> layout** (see `condura-mind/CLAUDE.md` §29.5), this project's CI lives
> in `condura-ops/ci/workflows/`. The files in this folder exist ONLY to
> point readers at the real workflows.

## Active workflows

| File | Purpose | Trigger |
|---|---|---|
| `condura-ops/ci/workflows/ci.yml` | go build + vet + race-tests + lint on every push and PR | every push, every PR |
| `condura-ops/ci/workflows/codeql.yml` | static security analysis | weekly, on push to `main` |
| `condura-ops/ci/workflows/release.yml` | tag-driven release: daemon + GUI installers + GoReleaser + GitHub release | push of `v*` tag |
| `condura-ops/ci/workflows/release-verify.yml` | verifies a signed update manifest roundtrips end-to-end | after `release.yml` |
| `condura-ops/ci/workflows/release-gui-patch.yml` | GUI-only patch releases (delta installers) | manual dispatch |

## Editing a workflow

Always edit the file under `condura-ops/ci/workflows/`. **Do not**
add a `.yml` file here (this README is the only file in this
directory). The GitHub Actions runner reads the canonical path.

This README is purely a discovery aid — its only job is to make
the Actions tab redirect a browsing contributor.

## Why not symlink

`.github/workflows/*.yml` can be real files but GitHub does not
follow symlinks from this directory by default in all plans; we
chose a README redirect over a copy/symlink because
copy = drift and symlink = platform-specific behavior. The
canonical source-of-truth stays under `condura-ops/`.
