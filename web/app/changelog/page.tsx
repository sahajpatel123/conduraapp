import type { Metadata } from "next";
import { readRepoMarkdown } from "@/lib/markdown";

export const metadata: Metadata = {
  title: "Changelog · Condura",
  description: "Notable changes to Condura, release by release.",
};

export default async function ChangelogPage() {
  const html = await readRepoMarkdown("CHANGELOG.md");

  return (
    <main className="mx-auto max-w-2xl px-6 py-20">
      <p className="text-sm font-medium uppercase tracking-widest text-neutral-500">
        Changelog
      </p>

      {html ? (
        <article
          className="prose-md mt-6"
          dangerouslySetInnerHTML={{ __html: html }}
        />
      ) : (
        <div className="mt-6">
          <h1 className="text-4xl font-semibold tracking-tight text-white">
            Changelog
          </h1>
          <p className="mt-4 text-neutral-400">
            The changelog is not available right now. Check the{" "}
            <a
              className="underline hover:text-white"
              href="https://github.com/sahajpatel123/conduraapp/releases"
            >
              GitHub releases
            </a>{" "}
            for the latest changes.
          </p>
        </div>
      )}
    </main>
  );
}
