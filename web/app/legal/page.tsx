import type { Metadata } from "next";
import PageChrome from "@/components/shell/PageChrome";
import { readRepoMarkdown } from "@/lib/markdown";
import { SITE } from "@/lib/site";

export const metadata: Metadata = {
  title: `Legal · ${SITE.name}`,
  description: "The Condura End-User License Agreement.",
};

export default async function LegalPage() {
  const html = await readRepoMarkdown("EULA.md");

  return (
    <PageChrome
      eyebrow="Legal"
      title="End-User License Agreement"
      description="The terms governing use of the Condura application."
    >
      {html ? (
        <article className="prose-md" dangerouslySetInnerHTML={{ __html: html }} />
      ) : (
        <p className="text-white/45">
          The license agreement is not available right now. Contact{" "}
          <a className="underline hover:text-white" href="mailto:legal@condura.app">
            legal@condura.app
          </a>
          .
        </p>
      )}
    </PageChrome>
  );
}
