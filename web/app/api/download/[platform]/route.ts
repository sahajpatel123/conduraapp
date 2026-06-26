import { NextRequest, NextResponse } from "next/server";

/**
 * Download Proxy API Route
 *
 * Proxies download requests to GitHub Releases to provide:
 * 1. Direct downloads without GitHub redirect to release page
 * 2. Clean URLs (e.g., /api/download/mac instead of GitHub URLs)
 * 3. Optional GitHub token authentication to avoid rate limits
 * 4. Proper Content-Disposition headers to force download
 *
 * Usage:
 *   /api/download/mac          -> macOS DMG installer
 *   /api/download/mac-intel    -> macOS Intel DMG installer
 *   /api/download/windows      -> Windows setup EXE
 *   /api/download/windows-portable -> Windows portable EXE
 *   /api/download/linux        -> Linux .deb package
 *   /api/download/linux-rpm    -> Linux .rpm package
 *   /api/download/linux-appimage -> Linux AppImage
 *   /api/download/daemon-mac   -> macOS daemon only
 *   /api/download/daemon-linux -> Linux daemon only
 *   /api/download/daemon-windows -> Windows daemon only
 */

const GITHUB_REPO = "sahajpatel123/conduraapp";
const RELEASE_BASE = `https://github.com/${GITHUB_REPO}/releases/latest/download`;
const RELEASE_PAGE = `https://github.com/${GITHUB_REPO}/releases/latest`;
const API_BASE = `https://api.github.com/repos/${GITHUB_REPO}/releases/latest`;

// Map of platform keys to artifact filename patterns.
// For GUI artifacts: exact filenames (no version in name).
// For daemon/CLI archives: GoReleaser uses versioned names
// (e.g., condurad-v0.1.0-darwin-arm64.tar.gz), so we use a
// prefix-based lookup via the GitHub API.
const ARTIFACTS: Record<string, string> = {
  // macOS GUI installer (built by build-gui.sh, Apple silicon only, no version in name)
  mac: "condura-gui-darwin-arm64.dmg",

  // Windows GUI artifact (Wails produces a zip, not a signed installer)
  windows: "condura-gui-windows-amd64.zip",

  // Linux packages (GoReleaser nfpms uses underscores, no v prefix)
  linux: "condurad_0.1.1_linux_amd64.deb",
  "linux-appimage": "condura-gui-linux-amd64",
};

// Daemon/CLI archives use GoReleaser versioned names.
// We store a prefix and match dynamically from release assets.
// Note: GoReleaser names use "0.1.1" not "v0.1.1".
const VERSIONED_PREFIXES: Record<string, string> = {
  "daemon-mac": "condurad-",
  "daemon-mac-intel": "condurad-",
  "daemon-windows": "condurad-",
  "daemon-linux": "condurad-",
  "cli-mac": "condura-cli-",
  "cli-mac-intel": "condura-cli-",
  "cli-windows": "condura-cli-",
  "cli-linux": "condura-cli-",
};

// Suffix patterns to match the correct OS/arch from versioned names
const VERSIONED_SUFFIXES: Record<string, string> = {
  "daemon-mac": "-darwin-arm64.tar.gz",
  "daemon-mac-intel": "-darwin-amd64.tar.gz",
  "daemon-windows": "-windows-amd64.tar.gz",
  "daemon-linux": "-linux-amd64.tar.gz",
  "cli-mac": "-darwin-arm64.tar.gz",
  "cli-mac-intel": "-darwin-amd64.tar.gz",
  "cli-windows": "-windows-amd64.tar.gz",
  "cli-linux": "-linux-amd64.tar.gz",
};

// Human-readable filenames for Content-Disposition
const FILENAMES: Record<string, string> = {
  mac: "condura-installer-mac.dmg",
  windows: "condura-windows.zip",
  linux: "condura.deb",
  "linux-appimage": "condura-gui-linux-amd64",
  "daemon-mac": "condura-daemon-mac.tar.gz",
  "daemon-mac-intel": "condura-daemon-mac-intel.tar.gz",
  "daemon-windows": "condura-daemon-windows.zip",
  "daemon-linux": "condura-daemon-linux.tar.gz",
  "cli-mac": "condura-cli-mac.tar.gz",
  "cli-mac-intel": "condura-cli-mac-intel.tar.gz",
  "cli-windows": "condura-cli-windows.zip",
  "cli-linux": "condura-cli-linux.tar.gz",
};

export const runtime = "nodejs"; // Need Node.js runtime for streaming

async function findVersionedArtifact(
  prefix: string,
  suffix: string
): Promise<string | null> {
  try {
    const headers: HeadersInit = { Accept: "application/vnd.github+json" };
    if (process.env.GITHUB_TOKEN) {
      headers.Authorization = `Bearer ${process.env.GITHUB_TOKEN}`;
    }
    const res = await fetch(API_BASE, { headers });
    if (!res.ok) return null;
    const release = await res.json();
    const assets: Array<{ name: string }> = release.assets || [];
    const match = assets.find(
      (a) => a.name.startsWith(prefix) && a.name.endsWith(suffix)
    );
    return match?.name || null;
  } catch {
    return null;
  }
}

export async function GET(
  _req: NextRequest,
  { params }: { params: Promise<{ platform: string }> }
) {
  const { platform: platformParam } = await params;
  const platform = platformParam.toLowerCase();

  // Look up the artifact filename
  let artifact: string | undefined = ARTIFACTS[platform];
  if (!artifact && VERSIONED_PREFIXES[platform]) {
    // Resolve versioned artifact dynamically from GitHub API
    artifact =
      (await findVersionedArtifact(
        VERSIONED_PREFIXES[platform],
        VERSIONED_SUFFIXES[platform]
      )) || undefined;
  }
  if (!artifact) {
    return NextResponse.json(
      {
        error: "Unknown platform",
        message: `Platform "${platform}" is not supported.`,
        availablePlatforms: [
          ...Object.keys(ARTIFACTS),
          ...Object.keys(VERSIONED_PREFIXES),
        ],
      },
      { status: 404 }
    );
  }

  const filename = FILENAMES[platform] || artifact;
  const githubUrl = `${RELEASE_BASE}/${artifact}`;

  try {
    // Fetch from GitHub Releases
    // If GITHUB_TOKEN is set, use it to avoid rate limits
    const headers: HeadersInit = {
      Accept: "application/octet-stream",
      "User-Agent": "condura-website-download-proxy",
    };

    if (process.env.GITHUB_TOKEN) {
      headers.Authorization = `Bearer ${process.env.GITHUB_TOKEN}`;
    }

    const response = await fetch(githubUrl, { headers });

    if (!response.ok) {
      console.error(
        `GitHub fetch failed for ${platform}: ${response.status} ${response.statusText}`
      );
      // If artifact not found, redirect to the GitHub release page
      // so the user can still download from GitHub directly
      if (response.status === 404) {
        return NextResponse.redirect(RELEASE_PAGE, { status: 302 });
      }
      return NextResponse.json(
        {
          error: "Download failed",
          message: `Could not fetch the ${platform} installer from GitHub.`,
          status: response.status,
        },
        { status: 502 }
      );
    }

    // Get the response body as a stream
    const body = response.body;
    if (!body) {
      return NextResponse.json(
        { error: "No content", message: "GitHub returned an empty response." },
        { status: 502 }
      );
    }

    // Stream the response back to the client with proper download headers
    return new NextResponse(body, {
      status: 200,
      headers: {
        "Content-Type": response.headers.get("Content-Type") || "application/octet-stream",
        "Content-Disposition": `attachment; filename="${filename}"`,
        "Content-Length": response.headers.get("Content-Length") || "",
        "Cache-Control": "public, max-age=3600, s-maxage=3600",
        "X-Content-Type-Options": "nosniff",
      },
    });
  } catch (error) {
    console.error(`Download proxy error for ${platform}:`, error);
    return NextResponse.json(
      {
        error: "Proxy error",
        message: "An unexpected error occurred while proxying the download.",
      },
      { status: 500 }
    );
  }
}
