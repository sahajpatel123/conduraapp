import type { Metadata } from "next";
import ManifestoPageClient from "./ManifestoPageClient";
import { SITE } from "@/lib/site";

export const metadata: Metadata = {
  title: `Manifesto · ${SITE.name}`,
  description: `Why ${SITE.name} exists. A local-first, privacy-first, free AI conductor for your machine.`,
};

export default function ManifestoPage() {
  return <ManifestoPageClient />;
}
