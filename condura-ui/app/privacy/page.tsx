import type { Metadata } from "next";
import PageHeader from "@/components/shell/PageHeader";
import { readRepoMarkdown } from "@/lib/markdown";
import { SITE } from "@/lib/site";

export const metadata: Metadata = {
  title: `Privacy · ${SITE.name}`,
  description: `${SITE.name} collects no telemetry, no analytics, and no personal data. Your data stays on your machine.`,
};

// Audit 2026-06-29 fix: hard-coded SECTIONS array previously drifted
// from the canonical PRIVACY.md. Now read PRIVACY.md at build time
// via readRepoMarkdown (same pattern as changelog + legal pages).
export default async function PrivacyPage() {
  const html = await readRepoMarkdown("PRIVACY.md");

  return (
    <PageHeader
      eyebrow="Privacy"
      title="What we"
      titleAccent="do not collect."
      description="Condura is local-first. Your conversations, your API keys, your files, your audit log — they live on your machine. Here is the canonical privacy policy."
    >
      <div className="relative mx-auto mt-8 max-w-3xl">
        {html ? (
          <article
            className="prose-condura"
            dangerouslySetInnerHTML={{ __html: html }}
          />
        ) : (
          <div className="surface-card mt-10 p-12 text-center">
            <h2 className="font-display text-2xl text-[var(--color-ink)]">
              Privacy policy temporarily unavailable
            </h2>
            <p className="mt-3 text-[var(--color-ink-soft)]">
              The repository PRIVACY.md could not be read at build time. The
              canonical text lives in the repo; for now, email{" "}
              <a href="mailto:privacy@condura.app" className="underline">
                privacy@{SITE.name.toLowerCase()}.app
              </a>{" "}
              for the latest version.
            </p>
          </div>
        )}
      </div>
    </PageHeader>
  );
}
