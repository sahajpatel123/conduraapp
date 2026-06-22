#!/usr/bin/env bash
# build-gui.sh — build the Condura Wails desktop app for the current OS/arch.
# Output: dist/prebuilt/synaptic-gui-<goos>-<goarch>[.exe]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo v0.0.0-dev)}"
COMMIT="${COMMIT:-$(git rev-parse HEAD 2>/dev/null || echo none)}"
BUILD_DATE="${BUILD_DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"

LDFLAGS="-s -w \
  -X github.com/sahajpatel123/synapticapp/internal/version.Version=${VERSION} \
  -X github.com/sahajpatel123/synapticapp/internal/version.Commit=${COMMIT} \
  -X github.com/sahajpatel123/synapticapp/internal/version.BuildDate=${BUILD_DATE}"

OUT_DIR="${ROOT}/dist/prebuilt"
EXT=""
if [ "$GOOS" = "windows" ]; then
  EXT=".exe"
fi
DEST="${OUT_DIR}/synaptic-gui-${GOOS}-${GOARCH}${EXT}"
mkdir -p "$OUT_DIR"

cd "${ROOT}/app/web"

if ! command -v wails >/dev/null 2>&1; then
  echo "installing wails CLI..."
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
  export PATH="$(go env GOPATH)/bin:${PATH}"
fi

echo "Building frontend..."
(cd frontend && npm ci && npm run build)

echo "Building Wails app for ${GOOS}/${GOARCH}..."
if [ "$GOOS" = "linux" ]; then
  # Ubuntu 24.04+ ships webkit2gtk-4.1 only (see wails.io docs).
  wails build -clean -trimpath -platform "${GOOS}/${GOARCH}" -ldflags "${LDFLAGS}" -tags webkit2_41
else
  wails build -clean -trimpath -platform "${GOOS}/${GOARCH}" -ldflags "${LDFLAGS}"
fi

# Wails outputfilename is "web" — normalize to synaptic for releases.
case "$GOOS" in
  darwin)
    APP="${ROOT}/app/web/build/bin/condura.app"
    if [ ! -d "$APP" ]; then
      APP="${ROOT}/app/web/build/bin/web.app"
    fi
    if [ ! -d "$APP" ]; then
      echo "wails build did not produce .app bundle under build/bin/" >&2
      exit 1
    fi
    rm -f "$DEST"
    ditto -c -k --keepParent "$APP" "${DEST}.zip"
    DEST="${DEST}.zip"
  ;;
  windows)
    BIN="${ROOT}/app/web/build/bin/web.exe"
    if [ ! -f "$BIN" ]; then
      BIN="${ROOT}/app/web/build/bin/synaptic.exe"
    fi
    cp "$BIN" "$DEST"
  ;;
  linux)
    BIN="${ROOT}/app/web/build/bin/web"
    if [ ! -f "$BIN" ]; then
      BIN="${ROOT}/app/web/build/bin/synaptic"
    fi
    cp "$BIN" "$DEST"
    chmod 755 "$DEST"
  ;;
  *)
    echo "unsupported GOOS=$GOOS" >&2
    exit 1
  ;;
esac

echo "GUI artifact: $DEST"
ls -la "$DEST"

# Platform installers (DMG / NSIS) for end-user distribution.
chmod +x "${ROOT}/scripts/package-gui-installers.sh"
"${ROOT}/scripts/package-gui-installers.sh"
