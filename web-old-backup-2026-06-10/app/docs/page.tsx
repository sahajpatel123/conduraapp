import type { Metadata } from "next";
import { ArrowRight } from "lucide-react";
import { LinkButton } from "@/components/button";
import { Section, SectionHeader } from "@/components/section";
import { docsCards } from "@/lib/site-data";

export const metadata: Metadata = {
  title: "Docs",
  description: "Synaptic setup, permissions, model, and dashboard documentation shell.",
};

export default function DocsPage() {
  return (
    <main id="main-content">
      <Section>
        <SectionHeader
          label="Docs"
          title="Setup guides will match the real release, not a hypothetical product."
          lead="This documentation shell establishes the information architecture for install, permissions, models, local use, and optional account services."
        />
        <div className="mt-10 grid gap-4 md:grid-cols-2">
          {docsCards.map((card) => {
            const Icon = card.icon;
            return (
              <article key={card.title} className="panel p-5">
                <Icon className="text-signal-cyan" aria-hidden="true" size={22} />
                <h2 className="mt-5 text-xl font-semibold text-white">{card.title}</h2>
                <p className="mt-3 text-sm leading-6 text-white/60">{card.body}</p>
              </article>
            );
          })}
        </div>
        <div className="mt-8">
          <LinkButton href="/download" variant="secondary">
            See release readiness
            <ArrowRight aria-hidden="true" size={18} />
          </LinkButton>
        </div>
      </Section>
    </main>
  );
}
