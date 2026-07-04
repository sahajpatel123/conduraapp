import type { Metadata } from "next";
import EcosystemPageClient from "./EcosystemPageClient";
import { SITE } from "@/lib/site";

export const metadata: Metadata = {
  title: `Integrations · ${SITE.name}`,
  description: `${SITE.name} works with 12+ LLM providers and 8+ agent CLIs. Connect the AI tools you already use.`,
};

export default function EcosystemPage() {
  return <EcosystemPageClient />;
}
