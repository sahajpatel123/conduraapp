import type { Metadata } from "next";
import PageChrome from "@/components/shell/PageChrome";
import BouncyAccordion from "@/components/motion/BouncyAccordion";
import { INVARIANTS, SITE } from "@/lib/site";

export const metadata: Metadata = {
  title: `Manifesto · ${SITE.name}`,
  description:
    "Why Condura exists: a privacy-first, local-first AI agent that treats your computer as yours.",
};

export default function ManifestoPage() {
  const items = INVARIANTS.map((inv) => ({
    id: inv.numeral,
    title: `${inv.numeral}. ${inv.title}`,
    body: inv.body,
  }));

  return (
    <PageChrome
      eyebrow="Manifesto"
      title="Your computer should work for you alone."
      description="Artificial intelligence is becoming how we use our machines. That shift is too important to hand to systems that watch everything you do."
    >
      <div className="space-y-8 text-[17px] leading-relaxed text-white/50">
        <p>
          {SITE.name} was built on a different premise: the most capable AI should also
          be the most respectful of you. It runs on your hardware, routes work through
          models you choose, and keeps your data where it belongs — on your machine.
        </p>
        <p>
          Local-first is not a feature. It is the foundation. Memory, API keys, skills,
          configuration, and audit logs live on your disk, encrypted at rest. When you
          sync devices, that data is end-to-end encrypted — never readable by us.
        </p>
        <p>
          Capability without a blank check: models hallucinate, prompts get injected,
          and some actions cannot be undone. Condura draws a hard line between thinking
          and acting.
        </p>
      </div>

      <div className="mt-12">
        <h2 className="mb-4 text-lg font-semibold text-white">The seven invariants</h2>
        <BouncyAccordion items={items} defaultOpenId="I" />
      </div>

      <p className="mt-12 text-sm text-white/40">
        Free, auditable, and built to earn trust instead of demanding it. Welcome to{" "}
        {SITE.name}.
      </p>
    </PageChrome>
  );
}
