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

const RELEASE_TAG =
  "https://github.com/sahajpatel123/conduraapp/releases/tag/v0.1.0";

export default function DownloadPage() {
  return (
    <main className="mx-auto max-w-3xl px-6 py-24 pt-[88px] text-neutral-100">
      <h1
        className="text-[32px] font-semibold tracking-tighter text-white sm:text-[40px]"
        style={{ fontFamily: "var(--font-display)" }}
      >
        Download Condura
      </h1>
      <p className="mt-4 text-[17px] leading-relaxed text-neutral-400">
        {SITE.description} Installers are signed, checksummed, and published from{" "}
        <a
          className="underline transition-colors hover:text-white"
          href="https://github.com/sahajpatel123/conduraapp/releases"
        >
          GitHub Releases
        </a>
        .
      </p>

      <ul className="mt-12 space-y-6">
        {PLATFORMS.map((p) => (
          <li
            key={p.key}
            className="rounded-2xl border border-white/[0.06] bg-white/[0.03] p-6 transition-colors hover:border-white/[0.12] hover:bg-white/[0.05]"
          >
            <div className="flex items-baseline justify-between">
              <h2 className="text-xl font-semibold text-white">{p.name}</h2>
              <p className="text-sm text-neutral-500">{p.requirement}</p>
            </div>
            <div className="mt-4 flex flex-wrap gap-3">
              {p.key === "mac" && (
                <>
                  <a
                    className="inline-flex items-center rounded-full bg-white px-5 py-2.5 text-sm font-medium text-black shadow-product transition-colors hover:bg-neutral-200"
                    href={DOWNLOADS.mac.dmg}
                  >
                    Download .dmg
                  </a>
                  <a
                    className="inline-flex items-center rounded-full border border-neutral-700 px-5 py-2.5 text-sm text-neutral-300 transition-colors hover:border-neutral-500 hover:text-white"
                    href={DOWNLOADS.mac.daemon}
                  >
                    Daemon only
                  </a>
                </>
              )}
              {p.key === "windows" && (
                <>
                  <a
                    className="inline-flex items-center rounded-full bg-white px-5 py-2.5 text-sm font-medium text-black shadow-product transition-colors hover:bg-neutral-200"
                    href={DOWNLOADS.windows.setup}
                  >
                    Download installer
                  </a>
                  <a
                    className="inline-flex items-center rounded-full border border-neutral-700 px-5 py-2.5 text-sm text-neutral-300 transition-colors hover:border-neutral-500 hover:text-white"
                    href={DOWNLOADS.windows.exe}
                  >
                    Portable .exe
                  </a>
                </>
              )}
              {p.key === "linux" && (
                <>
                  <a
                    className="inline-flex items-center rounded-full bg-white px-5 py-2.5 text-sm font-medium text-black shadow-product transition-colors hover:bg-neutral-200"
                    href={DOWNLOADS.linux.deb}
                  >
                    Download .deb (daemon)
                  </a>
                  <a
                    className="inline-flex items-center rounded-full border border-neutral-700 px-5 py-2.5 text-sm text-neutral-300 transition-colors hover:border-neutral-500 hover:text-white"
                    href={DOWNLOADS.linux.cli}
                  >
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
        <code className="text-neutral-300">
          go run ./cmd/gen-update-manifest verify manifest.json
        </code>
        .
      </p>
    </main>
  );
}
