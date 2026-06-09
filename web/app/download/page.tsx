import type { Metadata } from "next";
import { DownloadClient } from "./download-client";

export const metadata: Metadata = {
  title: "Download",
  description:
    "Synaptic for macOS, Windows and Linux. Free forever, signed binaries, no telemetry. Currently in rehearsal — follow the build.",
};

export default function DownloadPage() {
  return <DownloadClient />;
}
