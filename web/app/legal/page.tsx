import type { Metadata } from "next";
import { readRepoMarkdown } from "@/lib/markdown";

export const metadata: Metadata = {
  title: "Legal · Condura",
  description: "The Condura End-User License Agreement.",
};

export default async function LegalPage() {
  const html = await readRepoMarkdown("EULA.md");

  return (
    <main className="mx-auto max-w-2xl px-6 py-24 pt-[88px]">
      <p className="text-[13px] font-medium uppercase tracking-widest text-white/30">
        Legal
      </p>

      {html ? (
        <article className="prose-md mt-6" dangerouslySetInnerHTML={{ __html: html }} />
      ) : (
        <div className="mt-6">
          <h1 className="text-[32px] font-semibold tracking-tighter text-white sm:text-[40px]">
            End-User License Agreement
          </h1>
          <p className="mt-4 text-white/40">
            The license agreement is not available right now. Please check back
            shortly, or reach out at{" "}
            <a className="underline transition-colors hover:text-white" href="mailto:legal@condura.app">
              legal@condura.app
            </a>
            .
          </p>
        </div>
      )}
    </main>
  );
}
