import type { Metadata } from "next";
import LegalPageClient from "./LegalPageClient";
import { SITE } from "@/lib/site";

export const metadata: Metadata = {
  title: `Legal · ${SITE.name}`,
  description: `End-User License Agreement (EULA) for ${SITE.name}. Read the terms you agree to when installing the software.`,
};

export default function LegalPage() {
  return <LegalPageClient />;
}
