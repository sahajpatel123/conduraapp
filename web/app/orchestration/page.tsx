import type { Metadata } from "next";
import OrchestrationPageClient from "./OrchestrationPageClient";
import { SITE } from "@/lib/site";

export const metadata: Metadata = {
  title: `How it works · ${SITE.name}`,
  description: `Sub-agent delegation with a safety gate. ${SITE.name} spawns AI CLIs and gates each action through a deterministic gatekeeper.`,
};

export default function OrchestrationPage() {
  return <OrchestrationPageClient />;
}
