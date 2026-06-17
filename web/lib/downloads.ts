import type { PlatformKey } from "@/lib/site";

export const RELEASE_BASE =
  "https://github.com/sahajpatel123/conduraapp/releases/latest/download";

export const DOWNLOADS = {
  mac: {
    primary: { label: ".dmg installer", href: `${RELEASE_BASE}/condura-gui-darwin-arm64.dmg` },
    secondary: { label: "Daemon only", href: `${RELEASE_BASE}/condurad-darwin-arm64.tar.gz` },
  },
  windows: {
    primary: { label: "Setup .exe", href: `${RELEASE_BASE}/condura-gui-windows-amd64-setup.exe` },
    secondary: { label: "Portable .exe", href: `${RELEASE_BASE}/condura-gui-windows-amd64.exe` },
  },
  linux: {
    primary: { label: ".deb package", href: `${RELEASE_BASE}/condurad_0.1.0_linux_amd64.deb` },
    secondary: { label: "CLI tarball", href: `${RELEASE_BASE}/condura-cli-linux-amd64.tar.gz` },
  },
} as const satisfies Record<
  PlatformKey,
  { primary: { label: string; href: string }; secondary: { label: string; href: string } }
>;

export const RELEASE_TAG =
  "https://github.com/sahajpatel123/conduraapp/releases/tag/v0.1.0";
