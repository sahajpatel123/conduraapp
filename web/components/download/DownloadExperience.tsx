"use client";

import { useState } from "react";
import SharedLayoutTabs from "@/components/motion/SharedLayoutTabs";
import MagneticButton from "@/components/motion/MagneticButton";
import StatefulButton from "@/components/motion/StatefulButton";
import MorphingModal from "@/components/motion/MorphingModal";
import BottomSheet from "@/components/motion/BottomSheet";
import ActionSwap from "@/components/motion/ActionSwap";
import { useToast } from "@/context/ToastContext";
import { useIsland } from "@/context/IslandContext";
import { usePlatform } from "@/hooks/usePlatform";
import { DOWNLOADS, RELEASE_TAG } from "@/lib/downloads";
import { PLATFORMS, SITE, type PlatformKey } from "@/lib/site";

export default function DownloadExperience() {
  const detected = usePlatform();
  const [selectedPlatform, setSelectedPlatform] = useState<PlatformKey | null>(null);
  const [modalOpen, setModalOpen] = useState(false);
  const [sheetOpen, setSheetOpen] = useState(false);
  const [copied, setCopied] = useState(false);
  const { push } = useToast();
  const { pulseDownload } = useIsland();

  const platform = selectedPlatform ?? detected;
  const current = DOWNLOADS[platform];
  const platformMeta = PLATFORMS.find((p) => p.key === platform)!;

  const startDownload = async () => {
    pulseDownload(platformMeta.name);
    push({
      title: "Download started",
      description: `${current.primary.label} for ${platformMeta.name}.`,
      tone: "success",
    });
    window.location.href = current.primary.href;
    return true;
  };

  const copyLink = async () => {
    try {
      await navigator.clipboard.writeText(current.primary.href);
      setCopied(true);
      push({ title: "Link copied", description: "Direct download URL ready to share." });
      window.setTimeout(() => setCopied(false), 1600);
      return true;
    } catch {
      push({
        title: "Copy failed",
        description: "Your browser blocked clipboard access.",
        tone: "error",
      });
      return false;
    }
  };

  const platformTabs = PLATFORMS.map((p) => ({
    id: p.key,
    label: p.name,
    content: (
      <div className="mature-panel rounded-3xl p-6">
        <p className="text-sm text-white/45">{p.requirement}</p>
        <div className="mt-6 flex flex-wrap gap-3">
          <MagneticButton
            href={DOWNLOADS[p.key].primary.href}
            className="mature-button rounded-full px-6 py-3 text-sm font-semibold"
          >
            {DOWNLOADS[p.key].primary.label}
          </MagneticButton>
          <MagneticButton
            href={DOWNLOADS[p.key].secondary.href}
            className="mature-button-secondary rounded-full px-5 py-3 text-sm"
          >
            {DOWNLOADS[p.key].secondary.label}
          </MagneticButton>
          <button
            type="button"
            onClick={() => setModalOpen(true)}
            className="mature-button-secondary rounded-full px-4 py-3 text-sm text-white/60 hover:text-white"
          >
            Verify checksum
          </button>
        </div>
      </div>
    ),
  }));

  return (
    <>
      <div className="hidden sm:block">
        <SharedLayoutTabs
          layoutId="download-platforms"
          value={platform}
          onChange={(id) => setSelectedPlatform(id as PlatformKey)}
          items={platformTabs}
        />
      </div>

      <div className="sm:hidden">
        <p className="text-sm text-white/45">
          Detected {platformMeta.name}. Open the sheet for all artifacts.
        </p>
        <button
          type="button"
          onClick={() => setSheetOpen(true)}
          className="mature-button-secondary mt-4 w-full rounded-2xl px-4 py-3 text-sm font-medium text-white"
        >
          Choose download
        </button>
      </div>

      <div className="mt-10 flex flex-wrap gap-3">
        <StatefulButton
          className="mature-button"
          idleLabel={`Download ${platformMeta.name} build`}
          loadingLabel="Starting…"
          successLabel="Started"
          onAction={startDownload}
        />
        <MagneticButton
          onClick={copyLink}
          className="mature-button-secondary rounded-full px-5 py-3.5 text-sm"
        >
          <ActionSwap primary="Copy direct link" secondary="Copied" active={copied} />
        </MagneticButton>
      </div>

      <p className="mt-8 text-sm text-white/35">
        Release notes on{" "}
        <a className="underline hover:text-white/60" href={RELEASE_TAG}>
          GitHub v0.1.0
        </a>
        . {SITE.name} binaries are signed and published from GitHub Releases.
      </p>

      <MorphingModal
        open={modalOpen}
        onClose={() => setModalOpen(false)}
        title="Verify before you install"
        footer={
          <MagneticButton
            onClick={() => setModalOpen(false)}
            className="mature-button-secondary rounded-full px-4 py-2 text-sm"
          >
            Close
          </MagneticButton>
        }
      >
        <p>
          Compare the SHA-256 on the GitHub release page with your download. Run{" "}
          <code className="rounded bg-white/[0.06] px-1.5 py-0.5">shasum -a 256</code> on
          macOS/Linux or{" "}
          <code className="rounded bg-white/[0.06] px-1.5 py-0.5">Get-FileHash</code> on
          Windows. Never install if the hash mismatches.
        </p>
      </MorphingModal>

      <BottomSheet open={sheetOpen} onClose={() => setSheetOpen(false)} title="Downloads">
        <ul className="space-y-2">
          {PLATFORMS.map((p) => (
            <li key={p.key}>
              <a
                href={DOWNLOADS[p.key].primary.href}
                className="flex items-center justify-between rounded-xl border border-white/[0.08] bg-white/[0.03] px-4 py-3 text-sm text-white/80 transition-colors hover:bg-white/[0.05]"
                onClick={() => pulseDownload(p.name)}
              >
                <span>{p.name}</span>
                <span className="text-white/35">{DOWNLOADS[p.key].primary.label}</span>
              </a>
            </li>
          ))}
        </ul>
      </BottomSheet>
    </>
  );
}
