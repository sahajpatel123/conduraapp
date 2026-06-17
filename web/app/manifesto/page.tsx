import type { Metadata } from "next";
import { INVARIANTS, SITE } from "@/lib/site";

export const metadata: Metadata = {
  title: "Manifesto · Condura",
  description:
    "Why Condura exists: a privacy-first, local-first AI agent that treats your computer as yours.",
};

export default function ManifestoPage() {
  return (
    <main className="mx-auto max-w-2xl px-6 py-24 pt-[88px]">
      <p className="text-[13px] font-medium uppercase tracking-widest text-white/30">
        Manifesto
      </p>
      <h1 className="mt-3 text-[32px] font-semibold tracking-tighter text-white sm:text-[40px]">
        Your computer should work for you alone.
      </h1>

      <div className="mt-10 space-y-6 text-[17px] leading-relaxed text-white/50">
        <p>
          Artificial intelligence is becoming the way we use our machines. That
          shift is too important to hand to systems that watch everything you
          do, ship your keystrokes to a data center, and ask you to trust a
          black box with the most personal computer you own.
        </p>
        <p>
          Condura was built on a different premise: that the most capable AI
          should also be the most respectful of you. It runs on your hardware,
          routes work through models you choose and control, and keeps your
          data where it belongs — on your machine.
        </p>

        <h2 className="pt-6 text-[24px] font-semibold tracking-tighter text-white">
          Local-first is not a feature. It is the foundation.
        </h2>
        <p>
          {SITE.name} is local-first by default. Your memory, your API keys,
          your skills, your configuration, and your audit logs live on your
          disk, encrypted at rest. There is no account to create, no profile to
          build, and no telemetry collected unless you explicitly opt in.
        </p>
        <p>
          When you do choose to sync across your devices, that data is
          end-to-end encrypted, so that only your devices — never us, never a
          server — can read it. Privacy is not a setting you have to find. It is
          the state you start in.
        </p>

        <h2 className="pt-6 text-[24px] font-semibold tracking-tighter text-white">
          Capability without a blank check.
        </h2>
        <p>
          An agent that can act on your computer is powerful, and power without
          limits is dangerous. Models hallucinate. They can be prompt-injected.
          Some actions — sending an email, moving money, deleting a file — cannot
          be undone. We refuse to pretend otherwise.
        </p>
        <p>
          So Condura draws a hard line between thinking and acting. The part
          that reasons is a model. The part that decides whether an action is
          allowed is plain, deterministic code. They are never the same system,
          and no model output reaches a click, a keystroke, or a shell command
          without passing the gate.
        </p>

        <h2 className="pt-6 text-[24px] font-semibold tracking-tighter text-white">
          The invariants we will not break.
        </h2>
        <p>
          These are the commitments that hold even when they are inconvenient —
          the promises that make Condura safe to invite into your work.
        </p>
      </div>

      <ol className="mt-8 space-y-4">
        {INVARIANTS.map((inv) => (
          <li
            key={inv.numeral}
            className="rounded-2xl border border-white/[0.06] bg-white/[0.02] p-5"
          >
            <div className="flex items-baseline gap-3">
              <span className="font-mono text-sm text-white/30">
                {inv.numeral}
              </span>
              <h3 className="text-base font-semibold text-white">{inv.title}</h3>
            </div>
            <p className="mt-2 pl-7 text-sm leading-relaxed text-white/40">
              {inv.body}
            </p>
          </li>
        ))}
      </ol>

      <div className="mt-12 space-y-6 text-[17px] leading-relaxed text-white/50">
        <h2 className="pt-2 text-[24px] font-semibold tracking-tighter text-white">
          Free, and open to scrutiny.
        </h2>
        <p>
          Condura is free. Not free as in a trial, or free until we change our
          minds — free because a tool this close to your private life should not
          come with a meter running. Every action it takes is written to a
          tamper-resistant, append-only log you can inspect.
        </p>
        <p>
          This is the agent we wanted to exist: one that earns trust instead of
          demanding it. Welcome to {SITE.name}.
        </p>
      </div>
    </main>
  );
}
