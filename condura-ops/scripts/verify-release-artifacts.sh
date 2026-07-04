#!/usr/bin/env bash
# verify-release-artifacts.sh — download a GitHub release and verify checksums + manifest signature.
# Usage: ./scripts/verify-release-artifacts.sh [v0.1.0]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
TAG="${1:-v0.1.0}"
REPO="${GITHUB_REPOSITORY:-sahajpatel123/conduraapp}"
WORKDIR="${ROOT}/dist/verify-${TAG}"

rm -rf "$WORKDIR"
mkdir -p "$WORKDIR"

echo "Downloading release ${TAG} from ${REPO}..."
gh release download "$TAG" -R "$REPO" --dir "$WORKDIR" \
  --pattern 'checksums.txt' \
  --pattern 'manifest.json' \
  --pattern 'condurad-*' \
  --pattern 'condura-cli-*' \
  --pattern 'condura-gui-*' || true

if [ ! -f "$WORKDIR/checksums.txt" ]; then
  echo "checksums.txt missing from release ${TAG}" >&2
  exit 1
fi

if [ ! -f "$WORKDIR/manifest.json" ]; then
  echo "manifest.json missing from release ${TAG}" >&2
  exit 1
fi

echo "Verifying manifest Ed25519 signature..."
(cd "$ROOT" && go run ./cmd/gen-update-manifest verify "$WORKDIR/manifest.json")

echo "Verifying archive checksums..."
while read -r hash file; do
  [ -z "$hash" ] && continue
  path="$WORKDIR/$file"
  if [ ! -f "$path" ]; then
    echo "skip (not downloaded): $file"
    continue
  fi
  actual=$(shasum -a 256 "$path" | awk '{print $1}')
  if [ "$actual" != "$hash" ]; then
    echo "CHECKSUM MISMATCH: $file" >&2
    echo "  expected: $hash" >&2
    echo "  actual:   $actual" >&2
    exit 1
  fi
  echo "ok: $file"
done < "$WORKDIR/checksums.txt"

echo "Release ${TAG} artifacts verified."
