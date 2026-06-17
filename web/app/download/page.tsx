import type { Metadata } from "next";
import PageChrome from "@/components/shell/PageChrome";
import DownloadExperience from "@/components/download/DownloadExperience";
import { SITE } from "@/lib/site";

export const metadata: Metadata = {
  title: `Download · ${SITE.name}`,
  description: `Download ${SITE.name} for macOS, Windows, and Linux.`,
};

export default function DownloadPage() {
  return (
    <PageChrome
      eyebrow="Download"
      title="Install Condura on your machine"
      description={SITE.description}
      badge="v0.1.0"
    >
      <DownloadExperience />
    </PageChrome>
  );
}
