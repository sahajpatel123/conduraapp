import type { Metadata } from "next";
import EcosystemPageClient from "./EcosystemPageClient";
import { SITE } from "@/lib/site";

export const metadata: Metadata = {
  title: `Integrations · ${SITE.name}`,
  description: `${SITE.name} connects API-key LLM providers and agent CLIs you already use. One provider at a time in v0.1.x; multi-provider routing is v0.2.0.`,
};

export default function EcosystemPage() {
  return <EcosystemPageClient />;
}
