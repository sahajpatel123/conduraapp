import type { Metadata } from "next";
import { FileClock } from "lucide-react";
import { LinkButton } from "@/components/button";
import { Section, SectionHeader } from "@/components/section";
import { emptyChangelog } from "@/lib/site-data";

export const metadata: Metadata = {
  title: "Changelog",
  description: "Release notes and public build history for Synaptic.",
};

export default function ChangelogPage() {
  return (
    <main id="main-content">
      <Section>
        <SectionHeader
          label="Changelog"
          title="Release history starts with the first signed public build."
          lead="The public changelog is ready, but it should not invent releases before installers exist."
        />
        <div className="panel mt-10 p-8 text-center">
          <FileClock className="mx-auto text-signal-cyan" aria-hidden="true" size={34} />
          <h2 className="mt-5 text-2xl font-semibold text-white">{emptyChangelog.title}</h2>
          <p className="mx-auto mt-3 max-w-2xl text-sm leading-6 text-white/62">{emptyChangelog.body}</p>
          <div className="mt-7 flex justify-center">
            <LinkButton href="/download" variant="secondary">
              View download status
            </LinkButton>
          </div>
        </div>
      </Section>
    </main>
  );
}
