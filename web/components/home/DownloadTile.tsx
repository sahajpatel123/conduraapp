"use client";

import { useState } from "react";
import { DOWNLOADS, RELEASE_TAG } from "@/lib/downloads";
import { PLATFORMS, type PlatformKey } from "@/lib/site";
import MorphingModal from "@/components/motion/MorphingModal";
import { usePlatform } from "@/hooks/usePlatform";

export default function DownloadTile() {
  const detected = usePlatform();
  const [selectedPlatform, setSelectedPlatform] = useState<PlatformKey | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const platform = selectedPlatform ?? detected;

  const handleDownload = (href: string) => {
    const link = document.createElement("a");
    link.href = href;
    link.setAttribute("download", "");
    document.body.appendChild(link);
    link.click();
    document.body.removeChild(link);
    setModalOpen(true);
  };

  const activePlatformInfo = PLATFORMS.find((p) => p.key === platform)!;
  const downloadUrls = DOWNLOADS[platform];

  return (
    <section
      id="download-tile"
      className="relative w-full bg-[#000000] py-[140px] px-6 flex flex-col items-center overflow-hidden border-t border-white/[0.08]"
    >
      <div className="mx-auto w-full max-w-4xl text-center">
        <h2 className="font-hero-display text-[#ffffff]">
          Download the bundle.
        </h2>
        <p className="mt-6 font-lead-airy text-[#a1a1aa] max-w-2xl mx-auto">
          Condura runs fully sandbox-contained on your device. Complete visual verification is performed on-disk. Check releases or fetch packages directly.
        </p>

        {/* Minimal Dark Segmented Control */}
        <div className="relative z-10 mt-14 inline-flex select-none items-center rounded-2xl border border-white/[0.08] bg-white/[0.035] p-1">
          {PLATFORMS.map((p) => {
            const isActive = platform === p.key;
            return (
              <button
                key={p.key}
                onClick={() => setSelectedPlatform(p.key)}
                className={`rounded-xl px-6 py-2 font-body-mature text-[14px] font-medium transition-colors duration-150 ${
                  isActive ? "bg-white/[0.10] text-[#ffffff]" : "text-[#a1a1aa] hover:text-[#ffffff]"
                }`}
              >
                {p.name}
              </button>
            );
          })}
        </div>

        {/* Download Box */}
        <div className="mature-panel relative z-10 mx-auto mt-12 max-w-md rounded-3xl p-10 text-left">
          <div className="mb-8 text-center">
            <h3 className="font-body-mature text-[20px] font-medium text-[#ffffff]">
              {activePlatformInfo.name} Installer
            </h3>
            <p className="text-[14px] text-[#a1a1aa] mt-2 font-body-mature">
              Requires {activePlatformInfo.requirement}
            </p>
          </div>

          <button
            onClick={() => handleDownload(downloadUrls.primary.href)}
            className="mature-button block w-full py-4 text-center"
          >
            Download for {activePlatformInfo.name}
          </button>

          <div className="mt-8 pt-6 border-t border-white/10 flex flex-col gap-4 text-[13px] font-body-mature">
            <div className="flex items-center justify-between">
              <span className="text-[#a1a1aa]">Primary Archive:</span>
              <a href={downloadUrls.primary.href} className="text-[#ffffff] hover:underline">
                {downloadUrls.primary.label}
              </a>
            </div>
            <div className="flex items-center justify-between">
              <span className="text-[#a1a1aa]">Secondary:</span>
              <a href={downloadUrls.secondary.href} className="text-[#ffffff] hover:underline">
                {downloadUrls.secondary.label}
              </a>
            </div>
            <div className="flex items-center justify-between pt-4 border-t border-white/10 text-[12px]">
              <span className="text-[#a1a1aa]">Verify release</span>
              <a href={RELEASE_TAG} target="_blank" rel="noopener noreferrer" className="text-[#ffffff] hover:underline">
                Release v0.1.0 on GitHub
              </a>
            </div>
          </div>
        </div>
      </div>

      <MorphingModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        title={`${activePlatformInfo.name} Guide`}
        footer={
          <button onClick={() => setModalOpen(false)} className="mature-button px-5 py-3">
            Got it, continue
          </button>
        }
      >
        <div className="mt-4 space-y-5 rounded-2xl border border-white/[0.08] bg-white/[0.035] p-6 text-left font-body-mature text-[#ffffff]">
          <p className="text-[15px] text-[#a1a1aa]">
            Your download has started. Follow these steps:
          </p>
          {platform === "mac" && (
            <ol className="list-decimal pl-5 space-y-3.5 text-[14px] text-[#a1a1aa]">
              <li><strong className="text-white">Install Package</strong>: Open the <code>.dmg</code> and drag <code>Condura.app</code> to Applications.</li>
              <li><strong className="text-white">TCC Permissions</strong>: Grant Accessibility and Screen Recording in Settings.</li>
              <li><strong className="text-white">Summon Agent</strong>: Double-tap your hotkey to call the Conductor.</li>
            </ol>
          )}
          {platform === "windows" && (
            <ol className="list-decimal pl-5 space-y-3.5 text-[14px] text-[#a1a1aa]">
              <li><strong className="text-white">Run Setup</strong>: Double click the downloaded <code>.exe</code>.</li>
              <li><strong className="text-white">Bypass SmartScreen</strong>: Click <em>More Info</em> &rarr; <em>Run Anyway</em>.</li>
              <li><strong className="text-white">Launch</strong>: Set your hotkey in the dashboard.</li>
            </ol>
          )}
          {platform === "linux" && (
            <ol className="list-decimal pl-5 space-y-3.5 text-[14px] text-[#a1a1aa]">
              <li><strong className="text-white">Extract</strong>: Install the DEB package via `dpkg`.</li>
              <li><strong className="text-white">Service</strong>: Ensure the systemd service starts.</li>
              <li><strong className="text-white">TUI Access</strong>: Open <code>condura-tui</code>.</li>
            </ol>
          )}
        </div>
      </MorphingModal>
    </section>
  );
}
