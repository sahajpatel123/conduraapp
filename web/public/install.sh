#!/bin/bash
# Synaptic v0.1.0 — one-line macOS installer
# Usage: curl -fsSL https://synaptic.app/install.sh | bash

set -e

REPO="sahajpatel123/synapticapp"
DMG="synaptic-gui-darwin-arm64.dmg"
URL="https://github.com/${REPO}/releases/latest/download/${DMG}"

echo "==> Downloading Synaptic..."
curl -fsSL -o /tmp/synaptic.dmg "$URL"

echo "==> Mounting disk image..."
VOLUME=$(hdiutil attach /tmp/synaptic.dmg -nobrowse -readonly | grep '/Volumes/' | awk '{print $NF}')
trap 'hdiutil detach "$VOLUME" 2>/dev/null; rm -f /tmp/synaptic.dmg' EXIT

echo "==> Installing to /Applications..."
cp -R "$VOLUME/Synaptic.app" /Applications/

echo ""
echo "✅ Synaptic installed to /Applications/Synaptic.app"
echo "   Open it and press your hotkey to summon the agent."
echo ""
echo "   Or install via Homebrew: brew install --cask synaptic"
