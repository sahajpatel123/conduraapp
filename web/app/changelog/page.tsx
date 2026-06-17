import type { Metadata } from "next";
import { readRepoMarkdown } from "@/lib/markdown";

export const metadata: Metadata = {
  title: "Changelog · Condura",
  description: "Notable changes to Condura, release by release.",
};

export default async function ChangelogPage() {
  const html = await readRepoMarkdown("CHANGELOG.md");

  return (
    <main className="mx-auto max-w-2xl px-6 py-24 pt-[88px]">
      <p className="text-[13px] font-medium uppercase tracking-widest text-white/30">
        Changelog
      </p>

      {html ? (
        <article className="prose-md mt-6" dangerouslySetInnerHTML={{ __html: html }} />
      ) : (
        <div className="mt-6">
          <h1 className="text-[32px] font-semibold tracking-tighter text-white sm:text-[40px]">
            Changelog
          </h1>
          <p className="mt-4 text-white/40">
            The changelog is not available right now. Check the{" "}
            <a className="underline transition-colors hover:text-white" href="https://github.com/sahajpatel123/conduraapp/releases">
              GitHub releases
            </a>{" "}
            for the latest changes.
          </p>
        </div>
      )}
    </main>
  );
}
