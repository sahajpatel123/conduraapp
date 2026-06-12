import type { Metadata } from "next";
import { PlatformDownloadGrid, ReleaseReadiness } from "@/components/download-panel";
import { Section, SectionHeader } from "@/components/section";

export const metadata: Metadata = {
  title: "Download",
  description: "Download Synaptic installers when release builds are verified. Buttons are currently visible but not wired.",
};

export default function DownloadPage() {
  return (
    <main id="main-content">
      <Section className="border-b border-white/10">
        <SectionHeader
          label="Download"
          title="Installers will be connected after desktop verification."
          lead="Download buttons are visible now to shape the public product flow. They intentionally do not trigger downloads until signed release artifacts, checksums, and release notes exist."
        />
        <div className="mt-10">
          <PlatformDownloadGrid />
        </div>
        <div className="mt-5">
          <ReleaseReadiness />
        </div>
      </Section>
    </main>
  );
}
