import type { Metadata } from "next";
import PageHeader from "@/components/shell/PageHeader";
import { readRepoMarkdown } from "@/lib/markdown";
import { SITE } from "@/lib/site";

export const metadata: Metadata = {
  title: `Changelog · ${SITE.name}`,
  description: "Notable changes to Condura, release by release.",
};

export default async function ChangelogPage() {
  const html = await readRepoMarkdown("CHANGELOG.md");

  return (
    <PageHeader
      eyebrow="Changelog"
      title="Constant"
      titleAccent="velocity."
      description="We ship improvements to Condura every week. From core engine performance to new native integrations. Here is the history of what we've built."
    >
      <div className="relative mx-auto mt-8 max-w-3xl">
        {/* timeline rail */}
        <div className="absolute bottom-0 left-[7px] top-2 w-px bg-gradient-to-b from-[rgba(20,17,11,0.25)] via-[rgba(20,17,11,0.12)] to-transparent md:left-[15px]" />
        {html ? (
          <article
            className="prose-condura relative pl-8 md:pl-12"
            dangerouslySetInnerHTML={{ __html: html }}
          />
        ) : (
          <div className="surface-card mt-10 p-12 text-center">
            <div className="mx-auto mb-6 grid h-16 w-16 place-items-center rounded-full border border-[rgba(20,17,11,0.14)] bg-[var(--color-paper)]">
              <svg width="24" height="24" viewBox="0 0 24 24" fill="none" stroke="var(--color-ink-mute)" strokeWidth="1.6" strokeLinecap="round" strokeLinejoin="round" aria-hidden>
                <path d="M12 2v20M17 5H9.5a3.5 3.5 0 0 0 0 7h5a3.5 3.5 0 0 1 0 7H6" />
              </svg>
            </div>
            <h3 className="font-display text-[24px] text-[var(--color-ink)]">No local changelog found</h3>
            <p className="mt-2 text-body text-[var(--color-ink-mute)] max-w-md mx-auto">
              We couldn&apos;t locate the CHANGELOG.md file in the current build environment.
            </p>
            <a
              className="btn btn-ghost mt-7 inline-flex"
              href="https://github.com/sahajpatel123/conduraapp/releases"
              target="_blank"
              rel="noreferrer"
            >
              View on GitHub <span aria-hidden>→</span>
            </a>
          </div>
        )}
      </div>
    </PageHeader>
  );
}
