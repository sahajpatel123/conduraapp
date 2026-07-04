import type { Metadata } from "next";
import DownloadPageView from "@/components/download/DownloadPageView";
import { SITE } from "@/lib/site";

export const metadata: Metadata = {
  title: `Download · ${SITE.name}`,
  description: `Download ${SITE.name} for macOS, Windows, and Linux. Free forever. Local-first. No account required.`,
};

export default function DownloadPage() {
  return <DownloadPageView />;
}
