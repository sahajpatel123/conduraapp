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

# Wails v2.12.0 removed the -frontend flag AND removed the implicit
# discovery of the frontend outside the Wails project root. The
# custom asset handler is driven by
# `condura-gui/frontend/assets/assets.go`, whose `//go:embed all:dist`
# pattern picks up files from a sibling `dist/` next to the assets
# package. The actual `dist/` lives one directory up
# (`../dist` from `assets/`).
#
# We cannot use a symlink here: Go's `//go:embed` considers a
# symlink to be "irregular" if it points outside the package's
# directory tree, and rejects the embed with:
#   pattern all:dist: cannot embed irregular file dist
# So we copy the dist tree INTO the assets package directory at
# build time, where `//go:embed` happily finds regular files.
# The copy is gitignored via .gitignore (add the entry if not
# present) so the working tree stays clean.
ASSETS_DIR="${ROOT}/condura-gui/frontend/assets"
rm -rf "${ASSETS_DIR}/dist"
mkdir -p "${ASSETS_DIR}/dist"
cp -R "${ROOT}/condura-gui/frontend/dist/." "${ASSETS_DIR}/dist/"
cleanup() { rm -rf "${ASSETS_DIR}/dist"; }
trap cleanup EXIT

echo "Building Wails app for ${GOOS}/${GOARCH}..."
# `-skipfrontend` because the build-gui.sh script already ran
# `(cd "$FRONTEND_DIR" && npm ci && npm run build)` above using
# $FRONTEND_DIR (an absolute path). Wails v2's internal
# `frontend:install` / `frontend:build` (driven by wails.json's
# relative paths) resolve against a different CWD and break with
# the topic-sliced layout. The wails.json skipfrontend field is
# NOT supported by v2.12.0; only the CLI flag works.
if [ "$GOOS" = "linux" ]; then
  # Ubuntu 24.04+ ships webkit2gtk-4.1 only (see wails.io docs).
  wails build -clean -trimpath -skipfrontend -platform "${GOOS}/${GOARCH}" -ldflags "${LDFLAGS}" -tags webkit2_41
else
  wails build -clean -trimpath -skipfrontend -platform "${GOOS}/${GOARCH}" -ldflags "${LDFLAGS}"
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
