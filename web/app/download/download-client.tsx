"use client";

/*
  The box office. The binaries are not on stage yet, and the page says so
  plainly — no fake buttons, no dark patterns. Your platform is detected
  and seated first.
*/
import Link from "next/link";
import { useSyncExternalStore } from "react";
import { Cascade, Item, Lines, Rise } from "@/components/motion/reveal";
import { PLATFORMS, type PlatformKey } from "@/lib/site";

function detectPlatform(): PlatformKey {
  const ua = navigator.userAgent.toLowerCase();
  if (ua.includes("win")) return "windows";
  if (ua.includes("linux") && !ua.includes("android")) return "linux";
  return "mac";
}

const noopSubscribe = () => () => {};

const PROMISES = [
  { k: "price", v: "free forever — no tiers, no trials" },
  { k: "telemetry", v: "none. there is nothing to opt out of" },
  { k: "data", v: "memory, skills and logs stay on your disk, encrypted" },
  { k: "keys", v: "your API keys go to your provider and nowhere else" },
  { k: "binaries", v: "signed and notarized, every release" },
];

export function DownloadClient() {
  // Server renders without a detected platform; the client snapshot
  // takes over after hydration without a mismatch.
  const platform = useSyncExternalStore<PlatformKey | null>(
    noopSubscribe,
    detectPlatform,
    () => null,
  );

  // Order stays stable across hydration; the detected card is
  // highlighted in place rather than reshuffled to the front.

  return (
    <>
      <section className="staff">
        <div className="mx-auto max-w-6xl px-5 pt-40 pb-20 md:px-8 md:pt-48 md:pb-28">
          <Rise>
            <p className="annotation">The box office</p>
          </Rise>
          <Lines
            as="h1"
            delay={0.15}
            className="display mt-6 text-[clamp(2.6rem,6.5vw,5rem)]"
            lines={[
              "The orchestra is",
              <span key="i" className="display-italic">
                still in rehearsal.
              </span>,
            ]}
          />
          <Rise delay={0.45}>
            <p className="mt-8 max-w-xl leading-relaxed text-ivory-dim">
              Synaptic is being built in public view of its own audit log.
              There is no binary to download yet, and we would rather tell
              you that plainly than hand you a button that lies. Opening
              night: when it is safe.
            </p>
          </Rise>
          <Rise delay={0.6}>
            <div className="mt-8 flex flex-wrap items-center gap-6">
              <Link href="/changelog" className="trace cta">
                Follow the rehearsals
              </Link>
              <span className="annotation">phase 6 of 8 complete</span>
            </div>
          </Rise>
        </div>
      </section>

      <section className="border-t border-line" aria-label="Platforms">
        <div className="mx-auto max-w-6xl px-5 py-20 md:px-8 md:py-28">
          <Rise>
            <h2 className="annotation">Reserved seating — ready on opening night</h2>
          </Rise>
          <Cascade delay={0.15} step={0.08} className="mt-8 grid gap-px border border-line bg-line md:grid-cols-3">
            {PLATFORMS.map((p) => {
              const detected = p.key === platform;
              return (
                <Item key={p.key} className={`relative px-6 py-8 ${detected ? "bg-ink-3" : "bg-ink"}`}>
                  {detected && (
                    <span className="annotation absolute top-4 right-5 !text-brass">
                      your machine
                    </span>
                  )}
                  <h3 className="font-display text-2xl font-normal">{p.name}</h3>
                  <p className="mt-2 text-sm text-ivory-dim">{p.requirement}</p>
                  <p className="mt-6 font-mono text-xs text-ivory-faint">
                    {p.artifact}
                    <span className="mx-2">·</span>
                    sha256 published with the release
                  </p>
                </Item>
              );
            })}
          </Cascade>
        </div>
      </section>

      <section className="border-t border-line bg-ink-2/40" aria-label="The promises">
        <div className="mx-auto max-w-6xl px-5 py-20 md:px-8 md:py-28">
          <Rise>
            <h2 className="annotation">Printed on every ticket</h2>
          </Rise>
          <Cascade delay={0.15} className="mt-8">
            {PROMISES.map((row) => (
              <Item
                key={row.k}
                className="grid gap-2 border-b border-line py-5 md:grid-cols-[10rem_1fr] md:gap-10"
              >
                <span className="annotation !text-brass">{row.k}</span>
                <span className="font-mono text-sm text-ivory-dim">{row.v}</span>
              </Item>
            ))}
          </Cascade>
          <Rise delay={0.3}>
            <p className="mt-10 max-w-xl text-sm leading-relaxed text-ivory-dim">
              Proprietary source, free binary. The repository is private; the
              promises above are enforced by architecture, not by policy — read{" "}
              <Link href="/manifesto" className="prose-link">
                the seven invariants
              </Link>{" "}
              to see how.
            </p>
          </Rise>
        </div>
      </section>
    </>
  );
}
