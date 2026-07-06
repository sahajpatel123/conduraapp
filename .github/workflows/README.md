# CI Workflows — at `.github/workflows/`

> **GitHub Actions reads workflows from this directory by convention.**
> That includes the actual yml files that run on every push and tag.
>
> There is also a `condura-ops/ci/workflows/` directory that mirrors
> the same yml files for the **topic-sliced layout** described in
> `condura-mind/CLAUDE.md.legacy` §29.5 (the 2026-07-04 reorg).
> **GitHub Actions does not read that directory** — only this one.
>
> The two directories can drift. If you edit a workflow here,
> consider mirroring the change to `condura-ops/ci/workflows/` for
> consistency, and vice versa. Tracking this as open work.

## Active workflows

| File | Purpose | Trigger |
|---|---|---|
| `ci.yml` | go build + vet + race-tests + lint + GUI smoke + integration | every push, every PR |
| `codeql.yml` | static security analysis | weekly + push to main |
| `release.yml` | tag-driven release: daemon + GUI installers + GoReleaser + GitHub release | push of `v*` tag |
| `release-verify.yml` | verifies a signed update manifest roundtrips end-to-end | after `release.yml` |
| `release-gui-patch.yml` | GUI-only patch releases (delta installers) | manual dispatch |

## Known drift (Phase 14 workspace cleanup)

- `.github/workflows/ci.yml` line 229 carries `continue-on-error: true`
  on the `gui-build` (GUI Build smoke) job. This is the working copy.
- `condura-ops/ci/workflows/ci.yml` is the SAME workflow FILE copied
  to the topic-sliced location, but **does NOT have the
  `continue-on-error: true` annotation** (it was added by commit
  `3535692 fix(ci): mark GUI Build smoke check as continue-on-error to
  unblock CI` which modified this directory's path; the reorg moved
  the file before the annotation landed on both copies).
- GitHub Actions reads THIS directory's `ci.yml`. CI passes because
  THIS file's `continue-on-error` is intact.
- The orphan in `condura-ops/ci/workflows/` would, if GitHub ever
  started reading it, fail CI because of the missing annotation. It
  is currently documentation-only drift; it is NOT a CI incident.

## Open work

- Decide whether workflows live canonically at `.github/workflows/`
  (current GH-Actions reality) or at `condura-ops/ci/workflows/`
  (topic-sliced target). Until then, both copies must stay in sync.
  Possible paths: (a) wire a `.github/workflows/ci.yml` that delegates
  via `workflow_call` to `condura-ops/ci/workflows/ci.yml@main` as a
  reusable workflow; (b) accept the duplication and write a CI lint
  check that fails if the two copies diverge; (c) revert the reorg
  for CI files only.
