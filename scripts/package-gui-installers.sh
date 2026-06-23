#!/usr/bin/env bash
# package-gui-installers.sh — DMG (macOS) and NSIS setup (Windows) for the GUI.
set -euo pipefail

ROOT="$(cd "$(dirname "$0")/.." && pwd)"
GOOS="${GOOS:-$(go env GOOS)}"
GOARCH="${GOARCH:-$(go env GOARCH)}"
OUT_DIR="${ROOT}/dist/prebuilt"

package_dmg() {
  local app_zip="$1"
  local dmg_out="${OUT_DIR}/condura-gui-${GOOS}-${GOARCH}.dmg"
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
  hdiutil create -volname "Condura" -srcfolder "$vol" -ov -format UDZO "$dmg_out" >/dev/null
  echo "DMG: $dmg_out"
}

package_nsis() {
  local exe="$1"
  local setup_out="${OUT_DIR}/condura-gui-${GOOS}-${GOARCH}-setup.exe"
  local makensis_bin
  makensis_bin="$(find_makensis || true)"
  if [ -z "$makensis_bin" ]; then
    echo "makensis not found — skipping NSIS installer (install NSIS on Windows CI)" >&2
    return 0
  fi
  # NSIS File() is picky about paths on Windows CI — stage a short local copy.
  local staging="${OUT_DIR}/nsis-staging"
  mkdir -p "$staging"
  local payload="${staging}/condura.exe"
  cp "$exe" "$payload"
  local out_arg="$setup_out"
  local exe_arg="$payload"
  if command -v cygpath >/dev/null 2>&1; then
    out_arg="$(cygpath -w "$setup_out")"
    exe_arg="$(cygpath -w "$payload")"
  fi
  "$makensis_bin" -NOCD \
    -DOUTFILE="$out_arg" \
    -DEXE="$exe_arg" \
    "${ROOT}/scripts/condura-gui.nsi"
  rm -rf "$staging"
  echo "NSIS: $setup_out"
}

find_makensis() {
  if command -v makensis >/dev/null 2>&1; then
    command -v makensis
    return 0
  fi
  local p
  for p in \
    "/c/Program Files (x86)/NSIS/makensis.exe" \
    "/c/Program Files/NSIS/makensis.exe" \
    "/c/Program Files (x86)/NSIS/Bin/makensis.exe"; do
    if [ -f "$p" ]; then
      echo "$p"
      return 0
    fi
  done
  return 1
}

case "$GOOS" in
  darwin)
    zip="${OUT_DIR}/condura-gui-${GOOS}-${GOARCH}.zip"
    if [ -f "$zip" ]; then
      package_dmg "$zip"
    else
      echo "package-gui-installers: missing $zip" >&2
      exit 1
    fi
    ;;
  windows)
    exe="${OUT_DIR}/condura-gui-${GOOS}-${GOARCH}.exe"
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
