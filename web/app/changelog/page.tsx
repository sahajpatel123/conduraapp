import type { Metadata } from "next";
import PageChrome from "@/components/shell/PageChrome";
import { readRepoMarkdown } from "@/lib/markdown";
import { SITE } from "@/lib/site";

export const metadata: Metadata = {
  title: `Changelog · ${SITE.name}`,
  description: "Notable changes to Condura, release by release.",
};

import FadeInStagger from "@/components/motion/FadeInStagger";

export default async function ChangelogPage() {
  const html = await readRepoMarkdown("CHANGELOG.md");

  return (
    <PageChrome
      eyebrow="Changelog"
      title="What shipped"
      description="Release notes pulled from the repository changelog when available."
    >
      {html ? (
        <FadeInStagger>
          <article className="prose-md" dangerouslySetInnerHTML={{ __html: html }} />
        </FadeInStagger>
      ) : (
        <p className="text-white/45">
          The changelog is not available right now. Check{" "}
          <a className="underline hover:text-white" href="https://github.com/sahajpatel123/conduraapp/releases">
            GitHub releases
          </a>{" "}
          for the latest changes.
        </p>
      )}
    </PageChrome>
  );
}
