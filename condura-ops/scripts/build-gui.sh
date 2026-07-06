#!/usr/bin/env bash
# build-gui.sh — build the Condura Wails desktop app for the current OS/arch.
# Output: dist/prebuilt/condura-gui-<goos>-<goarch>[.exe]
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/../.." && pwd)"
GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"
VERSION="${VERSION:-$(git describe --tags --always --dirty 2>/dev/null || echo v0.0.0-dev)}"
COMMIT="${COMMIT:-$(git rev-parse HEAD 2>/dev/null || echo none)}"
BUILD_DATE="${BUILD_DATE:-$(date -u +"%Y-%m-%dT%H:%M:%SZ")}"

LDFLAGS="-s -w \
  -X github.com/sahajpatel123/conduraapp/condura-app/internal/version.Version=${VERSION} \
  -X github.com/sahajpatel123/conduraapp/condura-app/internal/version.Commit=${COMMIT} \
  -X github.com/sahajpatel123/conduraapp/condura-app/internal/version.BuildDate=${BUILD_DATE}"

OUT_DIR="${ROOT}/dist/prebuilt"
EXT=""
if [ "$GOOS" = "windows" ]; then
  EXT=".exe"
fi
DEST="${OUT_DIR}/condura-gui-${GOOS}-${GOARCH}${EXT}"
mkdir -p "$OUT_DIR"

cd "${ROOT}/condura-app/cmd/condura-gui"

# The Svelte frontend lives in condura-gui/frontend/, not as a sibling of the
# Wails shell. Tell wails explicitly where to find it.
FRONTEND_DIR="${ROOT}/condura-gui/frontend"

if ! command -v wails >/dev/null 2>&1; then
  echo "installing wails CLI..."
  go install github.com/wailsapp/wails/v2/cmd/wails@v2.12.0
  export PATH="$(go env GOPATH)/bin:${PATH}"
fi

echo "Building frontend..."
(cd "$FRONTEND_DIR" && npm ci && npm run build)

# Vite writes directly to condura-gui/frontend/assets/dist (see
# vite.config.ts outDir). That is exactly where //go:embed all:dist
# in assets/assets.go looks. Do NOT copy from frontend/dist/ — that
# path is a stale legacy artifact and was the root cause of the GUI
# shipping old sidebar code while src/ had the new NavOrbit.
ASSETS_DIR="${ROOT}/condura-gui/frontend/assets"
if [ ! -f "${ASSETS_DIR}/dist/index.html" ]; then
  echo "frontend build did not produce ${ASSETS_DIR}/dist/index.html" >&2
  exit 1
fi
if ! grep -q 'lp-nav-row' "${ASSETS_DIR}/dist/assets/"*.js 2>/dev/null; then
  echo "embedded bundle missing lp-nav-row — frontend build is stale" >&2
  exit 1
fi
echo "Frontend bundle OK ($(wc -c < "${ASSETS_DIR}/dist/assets/"*.js | awk '{print $1}') bytes JS)"

echo "Building Wails app for ${GOOS}/${GOARCH}..."
# `-s` skips wails's internal frontend step; we already built above.
if [ "$GOOS" = "linux" ]; then
  wails build -clean -trimpath -s -platform "${GOOS}/${GOARCH}" -ldflags "${LDFLAGS}" -tags webkit2_41
else
  wails build -clean -trimpath -s -platform "${GOOS}/${GOARCH}" -ldflags "${LDFLAGS}"
fi

# Wails outputfilename is "web" — normalize to condura for releases.
case "$GOOS" in
  darwin)
    APP="${ROOT}/condura-app/cmd/condura-gui/build/bin/condura.app"
    if [ ! -d "$APP" ]; then
      APP="${ROOT}/condura-app/cmd/condura-gui/build/bin/web.app"
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
    BIN="${ROOT}/condura-app/cmd/condura-gui/build/bin/web.exe"
    if [ ! -f "$BIN" ]; then
      BIN="${ROOT}/condura-app/cmd/condura-gui/build/bin/condura.exe"
    fi
    cp "$BIN" "$DEST"
  ;;
  linux)
    BIN="${ROOT}/condura-app/cmd/condura-gui/build/bin/web"
    if [ ! -f "$BIN" ]; then
      BIN="${ROOT}/condura-app/cmd/condura-gui/build/bin/condura"
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
chmod +x "${ROOT}/condura-ops/scripts/package-gui-installers.sh"
"${ROOT}/condura-ops/scripts/package-gui-installers.sh"
