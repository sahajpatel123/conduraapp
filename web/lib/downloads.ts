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
    primary: { label: "CLI + TUI .zip", href: `${DOWNLOAD_BASE}/windows` },
    secondary: { label: "Daemon .zip", href: `${DOWNLOAD_BASE}/daemon-windows` },
  },
  linux: {
    primary: { label: ".deb (daemon only)", href: `${DOWNLOAD_BASE}/linux` },
    secondary: { label: "GUI binary", href: `${DOWNLOAD_BASE}/linux-appimage` },
  },
} as const satisfies Record<
  PlatformKey,
  { primary: { label: string; href: string }; secondary: { label: string; href: string } }
>;

// NOTE: This release tag is manually pinned and must be bumped for each
// release. The download API routes derive the artifact names from it.
export const RELEASE_TAG =
  "https://github.com/sahajpatel123/conduraapp/releases/tag/v0.1.1";
