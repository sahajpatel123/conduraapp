import type { PlatformKey } from "@/lib/site";

/**
 * Download URLs
 *
 * All downloads are proxied through our Next.js API routes to provide:
 * 1. Direct downloads (no GitHub redirect to release page)
 * 2. Clean URLs on our domain
 * 3. Optional analytics and rate limiting
 *
 * The API route at /api/download/[platform] handles the actual
 * download from GitHub Releases and streams it back to the user.
 */
const DOWNLOAD_BASE = "/api/download";

export const DOWNLOADS = {
  mac: {
    primary: { label: ".dmg installer", href: `${DOWNLOAD_BASE}/mac` },
    secondary: { label: "Daemon only", href: `${DOWNLOAD_BASE}/daemon-mac` },
  },
  windows: {
    primary: { label: "Setup .exe", href: `${DOWNLOAD_BASE}/windows` },
    secondary: { label: "Portable .exe", href: `${DOWNLOAD_BASE}/windows-portable` },
  },
  linux: {
    primary: { label: ".deb package", href: `${DOWNLOAD_BASE}/linux` },
    secondary: { label: "CLI tarball", href: `${DOWNLOAD_BASE}/cli-linux` },
  },
} as const satisfies Record<
  PlatformKey,
  { primary: { label: string; href: string }; secondary: { label: string; href: string } }
>;

export const RELEASE_TAG =
  "https://github.com/sahajpatel123/conduraapp/releases/tag/v0.1.0";
