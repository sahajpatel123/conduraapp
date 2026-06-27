import type { Metadata } from "next";
import PrivacyPageClient from "./PrivacyPageClient";
import { SITE } from "@/lib/site";

export const metadata: Metadata = {
  title: `Privacy · ${SITE.name}`,
  description: `${SITE.name} collects no telemetry, no analytics, and no personal data. Your data stays on your machine.`,
};

export default function PrivacyPage() {
  return <PrivacyPageClient />;
}
