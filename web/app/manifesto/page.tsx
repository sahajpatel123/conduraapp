import type { Metadata } from "next";
import Link from "next/link";
import { Cascade, Item, Lines, Rise } from "@/components/motion/reveal";
import { INVARIANTS } from "@/lib/site";

export const metadata: Metadata = {
  title: "Manifesto",
  description:
    "The mission and the seven non-negotiable invariants that govern Synaptic — the free, local-first conductor for every AI on your computer.",
};

const IS = [
  "A free desktop application — Mac, Windows, Linux.",
  "A persistent agent that lives in your menu bar, around the clock.",
  "A conductor for the AI tools already installed on your machine.",
  "A learner that adapts to how you work, on disk, encrypted.",
  "A guest that asks before it acts. Every time it matters.",
];

const IS_NOT = [
  "Not a cloud service — your keys, your machine, your model calls.",
  "Not a subscription. Not freemium. Not a trial.",
  "Not a single-vendor tool — twelve providers, eight CLIs, and counting.",
  "Not an autonomous weapon — destructive actions need a human click.",
  "Not interested in your data. There is no telemetry to opt out of.",
];

export default function Manifesto() {
  return (
    <>
      <section className="staff">
        <div className="mx-auto max-w-6xl px-5 pt-40 pb-24 md:px-8 md:pt-48 md:pb-32">
          <Rise>
            <p className="annotation">The manifesto</p>
          </Rise>
          <Lines
            as="h1"
            delay={0.15}
            className="display mt-6 max-w-4xl text-[clamp(2.4rem,6vw,4.8rem)]"
            lines={[
              "Make AI useful to every",
              "ordinary person, on every",
              <span key="i">
                computer, <span className="display-italic text-brass">for free.</span>
              </span>,
            ]}
          />
          <Rise delay={0.5}>
            <p className="mt-10 max-w-xl leading-relaxed text-ivory-dim">
              No lock-in. No tracking. No compromise on speed or safety.
              Synaptic exists because the best agents are locked behind
              subscriptions, clouds, and single-vendor stacks — and none of
              them give you one hotkey that can do anything on your computer.
              The missing piece was never another model. It was a conductor.
            </p>
          </Rise>
        </div>
      </section>

      <section className="border-t border-line" aria-labelledby="invariants">
        <div className="mx-auto max-w-6xl px-5 py-24 md:px-8 md:py-32">
          <Rise>
            <p className="annotation" id="invariants">
              The seven invariants — non-negotiable, in any order of features
            </p>
          </Rise>
          <Rise delay={0.2}>
            <p className="mt-6 max-w-xl leading-relaxed text-ivory-dim">
              Synaptic performs physical, often irreversible actions on a real
              operating system. That is not an optimization problem; it is a
              survival problem. If a feature conflicts with an invariant,
              the feature is wrong — the feature is removed.
            </p>
          </Rise>
          <div className="mt-16 space-y-0">
            {INVARIANTS.map((inv, i) => (
              <Cascade key={inv.numeral} className="grid gap-4 border-t border-line py-10 md:grid-cols-[7rem_1fr] md:gap-10">
                <Item>
                  <span className="numeral-outline text-6xl md:text-7xl">{inv.numeral}</span>
                </Item>
                <Item>
                  <h2 className="font-display text-2xl leading-snug md:text-3xl">
                    {inv.title}
                  </h2>
                  <p className="mt-4 max-w-xl leading-relaxed text-ivory-dim">{inv.body}</p>
                  {i === 3 && (
                    <p className="annotation mt-4 !text-halt">
                      hard hotkey · watchdog · network cut · menu-bar kill
                    </p>
                  )}
                </Item>
              </Cascade>
            ))}
          </div>
        </div>
      </section>

      <section className="border-t border-line bg-ink-2/40">
        <div className="mx-auto grid max-w-6xl gap-16 px-5 py-24 md:grid-cols-2 md:px-8 md:py-32">
          <div>
            <Rise>
              <h2 className="display text-3xl md:text-4xl">What Synaptic is</h2>
            </Rise>
            <Cascade delay={0.15} className="mt-8 space-y-5">
              {IS.map((line) => (
                <Item key={line} className="flex gap-4 border-b border-line pb-5">
                  <span aria-hidden className="mt-2 size-1.5 shrink-0 rotate-45 bg-brass" />
                  <p className="text-sm leading-relaxed text-ivory-dim">{line}</p>
                </Item>
              ))}
            </Cascade>
          </div>
          <div>
            <Rise>
              <h2 className="display-italic text-3xl md:text-4xl">What it is not</h2>
            </Rise>
            <Cascade delay={0.25} className="mt-8 space-y-5">
              {IS_NOT.map((line) => (
                <Item key={line} className="flex gap-4 border-b border-line pb-5">
                  <span aria-hidden className="mt-2 h-px w-3 shrink-0 bg-ivory-faint" />
                  <p className="text-sm leading-relaxed text-ivory-dim">{line}</p>
                </Item>
              ))}
            </Cascade>
          </div>
        </div>
      </section>

      <section className="border-t border-line">
        <div className="mx-auto max-w-6xl px-5 py-24 text-center md:px-8 md:py-32">
          <Lines
            as="p"
            className="display-italic mx-auto text-[clamp(1.6rem,3.6vw,2.8rem)]"
            lines={["Free, fast, and yours.", "That is the whole promise."]}
          />
          <Rise delay={0.4}>
            <p className="annotation mt-8">— The Synaptic Project</p>
            <Link href="/download" className="prose-link mt-6 inline-block text-sm">
              Take your seat
            </Link>
          </Rise>
        </div>
      </section>
    </>
  );
}
