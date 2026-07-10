#!/usr/bin/env bash
# release-dry-run-local.sh — packaging dry-run that needs NO GitHub secrets.
#
# Proves GoReleaser config + archive layout locally.
# Does NOT prove Apple notarization or real UPDATE_SIGNING_KEY signing.
#
# Full secret dry-run: see condura-mind/docs/release-runbook.md §2c (v0.0.0-test tag).
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
cd "$ROOT"

echo "==> go test (short, race off for speed)"
go test -count=1 -short ./condura-app/...

echo "==> make release-snapshot"
make release-snapshot

if [[ -f dist/checksums.txt ]] || [[ -f dist/SHA256SUMS ]]; then
  echo "==> checksums present under dist/"
  ls -la dist/checksums.txt dist/SHA256SUMS 2>/dev/null || true
else
  echo "WARN: no checksums.txt/SHA256SUMS found (GoReleaser version may name them differently)"
  ls dist/ | head -40
fi

if command -v make >/dev/null && grep -q '^gen-manifest:' Makefile 2>/dev/null; then
  echo "==> make gen-manifest (unsigned)"
  make gen-manifest || echo "WARN: gen-manifest failed (non-fatal for packaging dry-run)"
fi

echo ""
echo "Local packaging dry-run OK."
echo "Next (needs secrets): configure UPDATE_SIGNING_KEY + 6 Apple secrets,"
echo "then push tag v0.0.0-test per docs/release-runbook.md §2c."
