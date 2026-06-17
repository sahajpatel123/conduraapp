import { PLATFORMS, SITE } from "@/lib/site";

const RELEASE_BASE =
  "https://github.com/sahajpatel123/conduraapp/releases/latest/download";

const DOWNLOADS = {
  mac: {
    dmg: `${RELEASE_BASE}/condura-gui-darwin-arm64.dmg`,
    zip: `${RELEASE_BASE}/condura-gui-darwin-arm64.zip`,
    daemon: `${RELEASE_BASE}/condurad-darwin-arm64.tar.gz`,
  },
  windows: {
    setup: `${RELEASE_BASE}/condura-gui-windows-amd64-setup.exe`,
    exe: `${RELEASE_BASE}/condura-gui-windows-amd64.exe`,
    daemon: `${RELEASE_BASE}/condurad-windows-amd64.zip`,
  },
  linux: {
    deb: `${RELEASE_BASE}/condurad_0.1.0_linux_amd64.deb`,
    cli: `${RELEASE_BASE}/condura-cli-linux-amd64.tar.gz`,
    gui: `${RELEASE_BASE}/condura-gui-linux-amd64`,
  },
} as const;

const RELEASE_TAG = "https://github.com/sahajpatel123/conduraapp/releases/tag/v0.1.0";

export default function DownloadPage() {
  return (
    <main className="mx-auto max-w-3xl px-6 py-16 text-neutral-100">
      <h1 className="text-3xl font-semibold tracking-tight">Download Condura</h1>
      <p className="mt-4 text-neutral-400">
        {SITE.description} Installers are signed, checksummed, and published from{" "}
        <a
          className="underline hover:text-white"
          href="https://github.com/sahajpatel123/conduraapp/releases"
        >
          GitHub Releases
        </a>
        .
      </p>

      <ul className="mt-10 space-y-8">
        {PLATFORMS.map((p) => (
          <li key={p.key} className="rounded-lg border border-neutral-800 p-6">
            <h2 className="text-xl font-medium">{p.name}</h2>
            <p className="mt-1 text-sm text-neutral-500">{p.requirement}</p>
            <div className="mt-4 flex flex-wrap gap-3">
              {p.key === "mac" && (
                <>
                  <a className="rounded bg-white px-4 py-2 text-sm font-medium text-black" href={DOWNLOADS.mac.dmg}>
                    Download .dmg
                  </a>
                  <a className="rounded border border-neutral-600 px-4 py-2 text-sm" href={DOWNLOADS.mac.daemon}>
                    Daemon only
                  </a>
                </>
              )}
              {p.key === "windows" && (
                <>
                  <a className="rounded bg-white px-4 py-2 text-sm font-medium text-black" href={DOWNLOADS.windows.setup}>
                    Download installer
                  </a>
                  <a className="rounded border border-neutral-600 px-4 py-2 text-sm" href={DOWNLOADS.windows.exe}>
                    Portable .exe
                  </a>
                </>
              )}
              {p.key === "linux" && (
                <>
                  <a className="rounded bg-white px-4 py-2 text-sm font-medium text-black" href={DOWNLOADS.linux.deb}>
                    Download .deb (daemon)
                  </a>
                  <a className="rounded border border-neutral-600 px-4 py-2 text-sm" href={DOWNLOADS.linux.cli}>
                    CLI tarball
                  </a>
                </>
              )}
            </div>
          </li>
        ))}
      </ul>

      <p className="mt-12 text-sm text-neutral-500">
        Current release:{" "}
        <a className="underline hover:text-white" href={RELEASE_TAG}>
          v0.1.0 on GitHub
        </a>
        . Auto-update manifest:{" "}
        <a className="underline" href={`${RELEASE_BASE}/manifest.json`}>
          manifest.json
        </a>
        . Verify with{" "}
        <code className="text-neutral-300">go run ./cmd/gen-update-manifest verify manifest.json</code>.
      </p>
    </main>
  );
}
