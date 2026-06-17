import type { Metadata } from "next";
import { readRepoMarkdown } from "@/lib/markdown";
import { SITE } from "@/lib/site";
import ChangelogView from "@/components/changelog/ChangelogView";

export const metadata: Metadata = {
  title: `Changelog · ${SITE.name}`,
  description: "Notable changes to Condura, release by release.",
};

export default async function ChangelogPage() {
  const html = await readRepoMarkdown("CHANGELOG.md");
  return <ChangelogView html={html} />;
}
