import type { Metadata } from "next";
import SecurityPageClient from "./SecurityPageClient";
import { SITE } from "@/lib/site";

export const metadata: Metadata = {
  title: `Security · ${SITE.name}`,
  description: `How ${SITE.name} keeps your data safe: deterministic gatekeeper, air-gapped memory, immutable audit trail.`,
};

export default function SecurityPage() {
  return <SecurityPageClient />;
}
