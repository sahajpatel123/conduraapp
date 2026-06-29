import type { Metadata } from "next";
import PageHeader from "@/components/shell/PageHeader";
import { readRepoMarkdown } from "@/lib/markdown";
import { SITE } from "@/lib/site";

export const metadata: Metadata = {
  title: `Legal · ${SITE.name}`,
  description: `End-User License Agreement (EULA) for ${SITE.name}. Read the terms you agree to when installing the software.`,
};

// Audit 2026-06-29 fix: the previous version of this page
// hard-coded a SECTIONS array that drifted from the canonical
// EULA.md. Now we read EULA.md at build time via readRepoMarkdown
// so the marketing page and the EULA the user accepts in
// onboarding stay in sync. The CHANGELOG page already uses this
// pattern (see changelog/page.tsx).
export default async function LegalPage() {
  const html = await readRepoMarkdown("EULA.md");

  return (
    <PageHeader
      eyebrow="Legal"
      title="The"
      titleAccent="contract."
      description="We believe legal documents shouldn't be hidden in tiny text. Here is exactly what you agree to when you install Condura on your machine."
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
              EULA temporarily unavailable
            </h2>
            <p className="mt-3 text-[var(--color-ink-soft)]">
              The repository EULA.md could not be read at build time. The
              canonical text lives in the repo; for now, email{" "}
              <a href="mailto:legal@condura.app" className="underline">
                legal@{SITE.name.toLowerCase()}.app
              </a>{" "}
              for the latest version.
            </p>
          </div>
        )}
      </div>
    </PageHeader>
  );
}
