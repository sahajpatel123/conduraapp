#!/bin/bash
# Condura v0.1.0 — one-line macOS installer
# Usage: curl -fsSL https://condura.app/install.sh | bash

set -e

REPO="sahajpatel123/conduraapp"
DMG="condura-gui-darwin-arm64.dmg"
URL="https://github.com/${REPO}/releases/latest/download/${DMG}"

echo "==> Downloading Condura..."
curl -fsSL -o /tmp/condura.dmg "$URL"

echo "==> Mounting disk image..."
VOLUME=$(hdiutil attach /tmp/condura.dmg -nobrowse -readonly | grep '/Volumes/' | awk '{print $NF}')
trap 'hdiutil detach "$VOLUME" 2>/dev/null; rm -f /tmp/condura.dmg' EXIT

echo "==> Installing to /Applications..."
cp -R "$VOLUME/Condura.app" /Applications/

echo ""
echo "✅ Condura installed to /Applications/Condura.app"
echo "   Open it and press your hotkey to summon the agent."
echo ""
echo "   Or install via Homebrew: brew install --cask condura"
