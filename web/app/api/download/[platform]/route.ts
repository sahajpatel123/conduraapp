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

// Map of platform keys to GitHub artifact filenames
const ARTIFACTS: Record<string, string> = {
  // macOS GUI installers
  mac: "condura-gui-darwin-arm64.dmg",
  "mac-intel": "condura-gui-darwin-amd64.dmg",

  // Windows GUI installers
  windows: "condura-gui-windows-amd64-setup.exe",
  "windows-portable": "condura-gui-windows-amd64.exe",

  // Linux packages
  linux: "condurad_0.1.0_linux_amd64.deb",
  "linux-rpm": "condura-0.1.0-1.x86_64.rpm",
  "linux-appimage": "condura-gui-linux-amd64",

  // Daemon-only builds (no GUI)
  "daemon-mac": "condurad-darwin-arm64.tar.gz",
  "daemon-mac-intel": "condurad-darwin-amd64.tar.gz",
  "daemon-windows": "condurad-windows-amd64.tar.gz",
  "daemon-linux": "condurad_0.1.0_linux_amd64.tar.gz",

  // CLI-only builds
  "cli-mac": "condura-cli-darwin-arm64.tar.gz",
  "cli-mac-intel": "condura-cli-darwin-amd64.tar.gz",
  "cli-windows": "condura-cli-windows-amd64.tar.gz",
  "cli-linux": "condura-cli-linux-amd64.tar.gz",
};

// Human-readable filenames for Content-Disposition
const FILENAMES: Record<string, string> = {
  mac: "condura-installer-mac.dmg",
  "mac-intel": "condura-installer-mac-intel.dmg",
  windows: "condura-setup.exe",
  "windows-portable": "condura-portable.exe",
  linux: "condura.deb",
  "linux-rpm": "condura.rpm",
  "linux-appimage": "condura.AppImage",
  "daemon-mac": "condura-daemon-mac.tar.gz",
  "daemon-mac-intel": "condura-daemon-mac-intel.tar.gz",
  "daemon-windows": "condura-daemon-windows.tar.gz",
  "daemon-linux": "condura-daemon-linux.tar.gz",
  "cli-mac": "condura-cli-mac.tar.gz",
  "cli-mac-intel": "condura-cli-mac-intel.tar.gz",
  "cli-windows": "condura-cli-windows.tar.gz",
  "cli-linux": "condura-cli-linux.tar.gz",
};

export const runtime = "nodejs"; // Need Node.js runtime for streaming

export async function GET(
  _req: NextRequest,
  { params }: { params: Promise<{ platform: string }> }
) {
  const { platform: platformParam } = await params;
  const platform = platformParam.toLowerCase();

  // Look up the artifact filename
  const artifact = ARTIFACTS[platform];
  if (!artifact) {
    return NextResponse.json(
      {
        error: "Unknown platform",
        message: `Platform "${platform}" is not supported.`,
        availablePlatforms: Object.keys(ARTIFACTS),
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
      return NextResponse.json(
        {
          error: "Download failed",
          message: `Could not fetch the ${platform} installer from GitHub.`,
          status: response.status,
        },
        { status: response.status === 404 ? 404 : 502 }
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
