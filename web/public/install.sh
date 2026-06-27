#!/bin/bash
# Condura — one-line macOS installer with supply-chain verification.
# Usage: curl -fsSL https://condura.app/install.sh | bash
#        ./install.sh [version]        (e.g. ./install.sh 0.1.1 or ./install.sh v0.1.1)
#
# Before copying anything to /Applications, this installer verifies, in order:
#   1. SHA-256 of the downloaded DMG against the release's checksums.txt
#      (defends against transport / CDN / MITM tampering).
#   2. codesign --verify --deep --strict  (Apple Developer ID signature).
#   3. spctl --assess --type execute       (Apple notarization ticket — the
#      real trust anchor against a compromised GitHub account: a swapped
#      artifact cannot carry a valid notarization ticket without the
#      developer's Apple credentials).
# Any failure stops the install. There is no bypass flag.

set -euo pipefail

REPO="sahajpatel123/conduraapp"
DMG="condura-gui-darwin-arm64.dmg"
APP_NAME="Condura.app"

# Version resolution: arg > $CONDURA_VERSION > 0.1.1 (pinned default).
# Accept either "0.1.1" or "v0.1.1"; the URL always uses the v-prefixed tag.
# Never use releases/latest/download/ — pin to a specific tag.
VERSION="${1:-${CONDURA_VERSION:-0.1.1}}"
case "$VERSION" in
  v*) TAG="$VERSION" ;;
  *)  TAG="v${VERSION}" ;;
esac

BASE_URL="https://github.com/${REPO}/releases/download/${TAG}"

# Temp workspace + cleanup trap. VOLUME is guarded so the trap is a no-op
# until the DMG is actually mounted.
WORKDIR=""
VOLUME=""
cleanup() {
  if [ -n "$VOLUME" ]; then
    hdiutil detach "$VOLUME" 2>/dev/null || true
  fi
  if [ -n "$WORKDIR" ] && [ -d "$WORKDIR" ]; then
    rm -rf "$WORKDIR"
  fi
}
trap cleanup EXIT ERR

WORKDIR="$(mktemp -d)"
DMG_PATH="${WORKDIR}/${DMG}"
CHECKSUMS_PATH="${WORKDIR}/checksums.txt"

echo "==> Downloading Condura ${VERSION} (${TAG})..."
if ! curl -fsSL -o "$CHECKSUMS_PATH" "${BASE_URL}/checksums.txt"; then
  echo "ERROR: failed to download checksums.txt for tag ${TAG}." >&2
  echo "       Check that the version exists: https://github.com/${REPO}/releases/tag/${TAG}" >&2
  exit 1
fi
if ! curl -fsSL -o "$DMG_PATH" "${BASE_URL}/${DMG}"; then
  echo "ERROR: failed to download ${DMG} for tag ${TAG}." >&2
  exit 1
fi

echo "==> Verifying SHA-256 against checksums.txt..."
# GoReleaser checksums.txt format: "<sha256>  <filename>" per line.
# Match a filename field that equals $DMG or ends with "/$DMG".
expected=""
while read -r hash file; do
  [ -z "${hash:-}" ] && continue
  case "$file" in
    "$DMG" | */"$DMG")
      expected="$hash"
      break
      ;;
  esac
done < "$CHECKSUMS_PATH"

if [ -z "$expected" ]; then
  echo "ERROR: ${DMG} not found in checksums.txt for tag ${TAG}." >&2
  echo "       The release may be missing GUI artifacts (release pipeline issue)." >&2
  exit 1
fi

actual="$(shasum -a 256 "$DMG_PATH" | awk '{print $1}')"
if [ "$actual" != "$expected" ]; then
  echo "ERROR: SHA-256 mismatch for ${DMG}." >&2
  echo "  expected: $expected" >&2
  echo "  actual:   $actual" >&2
  exit 1
fi
echo "    sha256 ok ($actual)"

echo "==> Mounting disk image..."
# Capture the last /Volumes/... token from hdiutil output. Robust to spaces
# in volume names (e.g. "/Volumes/Condura 1" when an earlier mount exists).
VOLUME="$(hdiutil attach "$DMG_PATH" -nobrowse -readonly 2>/dev/null | grep -o '/Volumes/.*' | tail -1)"

if [ -z "$VOLUME" ] || [ ! -d "$VOLUME" ]; then
  echo "ERROR: could not locate mounted volume for ${DMG}." >&2
  exit 1
fi

APP="${VOLUME}/${APP_NAME}"
if [ ! -d "$APP" ]; then
  echo "ERROR: ${APP_NAME} not found in mounted DMG." >&2
  exit 1
fi

echo "==> Verifying Apple Developer ID signature..."
if ! codesign --verify --deep --strict --verbose=2 "$APP" 2>&1; then
  echo "ERROR: codesign verification failed for ${APP_NAME}." >&2
  echo "       The DMG is not signed with a valid Apple Developer ID." >&2
  exit 1
fi
echo "    signature ok"

echo "==> Verifying Apple notarization..."
if ! spctl --assess --type execute --verbose "$APP" 2>&1; then
  echo "ERROR: notarization assessment failed for ${APP_NAME}." >&2
  echo "       spctl rejected the app — it is not notarized or the ticket is invalid." >&2
  echo "       Refusing to install. See https://condura.app for notarized builds." >&2
  exit 1
fi
echo "    notarization ok"

echo "==> Installing to /Applications..."
cp -R "$APP" /Applications/

echo ""
echo "✅ Condura ${VERSION} installed to /Applications/${APP_NAME}"
echo "   Open it, and on first launch grant Accessibility + Screen Recording"
echo "   in System Settings → Privacy & Security."
echo ""
echo "   Or install via Homebrew: brew install --cask condura"
