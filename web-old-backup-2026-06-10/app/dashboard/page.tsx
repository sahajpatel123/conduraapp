import type { Metadata } from "next";
import { ExternalLink } from "lucide-react";
import { LinkButton } from "@/components/button";
import { Section, SectionHeader } from "@/components/section";
import { dashboardFeatures } from "@/lib/site-data";

export const metadata: Metadata = {
  title: "Dashboard",
  description: "Placeholder for optional browser-based Synaptic dashboard sign-in.",
};

export default function DashboardPage() {
  return (
    <main id="main-content">
      <Section>
        <SectionHeader
          label="Optional browser account"
          title="Dashboard sign-in will live in the browser, not inside the desktop app."
          lead="This placeholder keeps the navigation honest while the technical side continues. Download and local desktop use remain public and account-free."
        />
        <div className="mt-10 grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {dashboardFeatures.map((feature) => {
            const Icon = feature.icon;
            return (
              <article key={feature.title} className="panel p-5">
                <Icon className="text-signal-green" aria-hidden="true" size={22} />
                <h2 className="mt-5 text-lg font-semibold text-white">{feature.title}</h2>
                <p className="mt-3 text-sm leading-6 text-white/60">{feature.body}</p>
              </article>
            );
          })}
        </div>
        <div className="mt-8 flex flex-col gap-3 sm:flex-row">
          <LinkButton href="/download">
            Download without account
          </LinkButton>
          <LinkButton href="/privacy" variant="secondary">
            Why browser sign-in
            <ExternalLink aria-hidden="true" size={18} />
          </LinkButton>
        </div>
      </Section>
    </main>
  );
}
