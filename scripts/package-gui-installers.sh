#!/usr/bin/env bash
# package-gui-installers.sh — DMG (macOS) and NSIS setup (Windows) for the GUI.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"
OUT_DIR="${ROOT}/dist/prebuilt"

package_dmg() {
  local app_zip="$1"
  local dmg_out="${OUT_DIR}/synaptic-gui-${GOOS}-${GOARCH}.dmg"
  local staging
  staging="$(mktemp -d)"
  trap 'rm -rf "$staging"' RETURN

  unzip -q "$app_zip" -d "$staging"
  local app
  app="$(find "$staging" -maxdepth 2 -name '*.app' | head -1)"
  if [ -z "$app" ]; then
    echo "package-gui-installers: no .app in $app_zip" >&2
    exit 1
  fi

  local vol="${staging}/vol"
  mkdir -p "$vol"
  cp -R "$app" "$vol/"
  ln -s /Applications "$vol/Applications"

  rm -f "$dmg_out"
  hdiutil create -volname "Synaptic" -srcfolder "$vol" -ov -format UDZO "$dmg_out" >/dev/null
  echo "DMG: $dmg_out"
}

package_nsis() {
  local exe="$1"
  local setup_out="${OUT_DIR}/synaptic-gui-${GOOS}-${GOARCH}-setup.exe"
  if ! command -v makensis >/dev/null 2>&1; then
    echo "makensis not found — skipping NSIS installer (install NSIS on Windows CI)" >&2
    return 0
  fi
  makensis -NOCD \
    -DOUTFILE="$setup_out" \
    -DEXE="$exe" \
    "${ROOT}/scripts/synaptic-gui.nsi"
  echo "NSIS: $setup_out"
}

case "$GOOS" in
  darwin)
    zip="${OUT_DIR}/synaptic-gui-${GOOS}-${GOARCH}.zip"
    if [ -f "$zip" ]; then
      package_dmg "$zip"
    else
      echo "package-gui-installers: missing $zip" >&2
      exit 1
    fi
    ;;
  windows)
    exe="${OUT_DIR}/synaptic-gui-${GOOS}-${GOARCH}.exe"
    if [ -f "$exe" ]; then
      package_nsis "$exe"
    else
      echo "package-gui-installers: missing $exe" >&2
      exit 1
    fi
    ;;
  linux)
    echo "Linux GUI uses tarball/AppImage path — no extra installer step"
    ;;
  *)
    echo "unsupported GOOS=$GOOS" >&2
    exit 1
    ;;
esac
