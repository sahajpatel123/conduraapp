import type { Metadata } from "next";
import { readRepoMarkdown } from "@/lib/markdown";
import { SITE } from "@/lib/site";
import LegalView from "@/components/legal/LegalView";

export const metadata: Metadata = {
  title: `Legal · ${SITE.name}`,
  description: "The Condura End-User License Agreement.",
};

export default async function LegalPage() {
  const html = await readRepoMarkdown("EULA.md");
  return <LegalView html={html} />;
}
