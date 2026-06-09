import Link from "next/link";
import { Cascade, Item, Lines, Rise } from "@/components/motion/reveal";
import { Magnetic } from "@/components/motion/magnetic";
import { Terminal } from "@/components/pieces/terminal";
import { Gatekeeper } from "@/components/pieces/gatekeeper";
import { SummonKeys } from "@/components/pieces/summon-keys";
import { NumeralSection } from "@/components/pieces/numeral";
import { ORCHESTRA } from "@/lib/site";

/* The score: Overture → Mvt. I Summon → Mvt. II Orchestrate →
   Mvt. III The Gatekeeper → Interlude → Coda. */

const SECTIONS: Record<(typeof ORCHESTRA)[number], string> = {
  "Claude Code": "first violin",
  Codex: "second violin",
  Antigravity: "viola",
  OpenCode: "cello",
  Hermes: "double bass",
  Kilo: "woodwinds",
  Ollama: "percussion",
  Gemini: "horns",
};

const TEMPI = [
  { mark: "presto", value: "< 500 ms", label: "cold start" },
  { mark: "prestissimo", value: "< 100 ms", label: "hotkey to overlay" },
  { mark: "vivace", value: "< 1.5 s", label: "first token" },
];

const KILLS = [
  {
    name: "Hard hotkey",
    body: "One chord, anywhere, and the agent stops mid-gesture.",
  },
  {
    name: "Watchdog timer",
    body: "If the agent stalls or runs long, it is cut off automatically.",
  },
  {
    name: "Network isolation",
    body: "Sever every outbound call with a single switch.",
  },
  {
    name: "Menu-bar kill",
    body: "The red item at the top of your screen. Always there.",
  },
];

export default function Home() {
  return (
    <>
      {/* ───────────────────────── Overture ───────────────────────── */}
      <section className="staff relative">
        <div className="mx-auto grid max-w-6xl gap-14 px-5 pt-40 pb-28 md:grid-cols-[1.2fr_1fr] md:items-center md:gap-10 md:px-8 md:pt-48 md:pb-36">
          <div>
            <Rise>
              <p className="annotation">Op. 1 — for every computer, free</p>
            </Rise>
            <Lines
              as="h1"
              delay={0.15}
              className="display mt-6 text-[clamp(2.8rem,7.5vw,5.6rem)]"
              lines={[
                "Every AI on your",
                "machine. One",
                <span key="c" className="display-italic text-brass">
                  conductor.
                </span>,
              ]}
            />
            <Rise delay={0.5}>
              <p className="mt-8 max-w-md text-base leading-relaxed text-ivory-dim">
                Synaptic lives in your menu bar and conducts the AI tools you
                already own — summoned by hotkey, governed by a deterministic
                Gatekeeper, local-first, and free forever.
              </p>
            </Rise>
            <Rise delay={0.65}>
              <div className="mt-10 flex flex-wrap items-center gap-6">
                <Magnetic>
                  <Link href="/download" className="trace cta">
                    Get Synaptic
                  </Link>
                </Magnetic>
                <Link href="/manifesto" className="prose-link text-sm">
                  Read the manifesto
                </Link>
              </div>
              <p className="annotation mt-5">
                no subscription · no telemetry · no lock-in
              </p>
            </Rise>
          </div>
          <Rise delay={0.4}>
            <Terminal />
          </Rise>
        </div>
      </section>

      {/* ─────────────────────── The roster ──────────────────────── */}
      <section aria-label="The orchestra" className="border-y border-line">
        <Cascade
          step={0.05}
          className="mx-auto flex max-w-6xl flex-wrap items-baseline gap-x-8 gap-y-3 px-5 py-8 md:px-8"
        >
          <Item>
            <span className="annotation">The orchestra —</span>
          </Item>
          {ORCHESTRA.map((tool) => (
            <Item key={tool}>
              <span className="font-mono text-sm text-ivory-dim transition-colors duration-200 hover:text-brass">
                {tool}
              </span>
            </Item>
          ))}
          <Item>
            <span className="annotation">+ any key you already hold</span>
          </Item>
        </Cascade>
      </section>

      {/* ──────────────────── Mvt. I — Summon ─────────────────────── */}
      <NumeralSection numeral="I">
        <div className="mx-auto grid max-w-6xl gap-14 px-5 py-28 md:grid-cols-2 md:items-center md:gap-20 md:px-8 md:py-40">
          <div>
            <Rise>
              <p className="annotation">Mvt. I — allegro, 87 ms</p>
            </Rise>
            <Lines
              delay={0.1}
              className="display mt-6 text-[clamp(2.2rem,5vw,4rem)]"
              lines={["Summoned,", "not launched."]}
            />
            <Rise delay={0.35}>
              <p className="mt-8 max-w-md leading-relaxed text-ivory-dim">
                No dock icon to find, no window to wait for. Double-tap the
                hotkey — or say the wake word — and Synaptic is standing at the
                podium before your hand leaves the keyboard.
              </p>
            </Rise>
            <Cascade delay={0.5} className="mt-12 space-y-4">
              {TEMPI.map((t) => (
                <Item
                  key={t.label}
                  className="flex items-baseline gap-4 border-b border-line pb-4"
                >
                  <span className="annotation w-24 shrink-0 !text-brass">{t.mark}</span>
                  <span className="font-display text-2xl">{t.value}</span>
                  <span className="text-sm text-ivory-dim">{t.label}</span>
                </Item>
              ))}
            </Cascade>
          </div>
          <Rise delay={0.3}>
            <SummonKeys />
          </Rise>
        </div>
      </NumeralSection>

      {/* ────────────────── Mvt. II — Orchestrate ─────────────────── */}
      <NumeralSection numeral="II" className="border-t border-line bg-ink-2/40">
        <div className="mx-auto max-w-6xl px-5 py-28 md:px-8 md:py-40">
          <Rise>
            <p className="annotation">Mvt. II — tutti</p>
          </Rise>
          <Lines
            delay={0.1}
            className="display mt-6 max-w-3xl text-[clamp(2.2rem,5vw,4rem)]"
            lines={[
              "It doesn’t replace",
              "your AI tools.",
              <span key="i" className="display-italic">
                It conducts them.
              </span>,
            ]}
          />
          <Rise delay={0.35}>
            <p className="mt-8 max-w-xl leading-relaxed text-ivory-dim">
              Claude Code, Codex, Ollama, the subscriptions you already pay for —
              today they sit in separate rooms, unaware of one another. Synaptic
              seats them as an orchestra: a router picks the right player for
              each passage, a delegation bus hands off the work, and your task
              comes back as one finished piece.
            </p>
          </Rise>
          <Cascade
            delay={0.2}
            step={0.06}
            className="mt-16 grid grid-cols-2 gap-px border border-line bg-line lg:grid-cols-4"
          >
            {ORCHESTRA.map((tool) => (
              <Item
                key={tool}
                className="group bg-ink px-5 py-6 transition-colors duration-300 hover:bg-ink-3"
              >
                <p className="font-mono text-sm text-ivory">{tool}</p>
                <p className="annotation mt-2 !tracking-[0.12em] transition-colors duration-300 group-hover:!text-brass">
                  {SECTIONS[tool]}
                </p>
              </Item>
            ))}
          </Cascade>
          <p className="annotation mt-6">
            seated automatically when found on your machine — never installed for you
          </p>
        </div>
      </NumeralSection>

      {/* ───────────────── Mvt. III — The Gatekeeper ──────────────── */}
      <NumeralSection numeral="III" className="border-t border-line">
        <div className="mx-auto max-w-6xl px-5 py-28 md:px-8 md:py-40">
          <Rise>
            <p className="annotation">Mvt. III — grave, deterministic</p>
          </Rise>
          <Lines
            delay={0.1}
            className="display mt-6 max-w-3xl text-[clamp(2.2rem,5vw,4rem)]"
            lines={[
              "Models propose.",
              <span key="g" className="display-italic">
                The Gatekeeper disposes.
              </span>,
            ]}
          />
          <Rise delay={0.35}>
            <p className="mt-8 max-w-xl leading-relaxed text-ivory-dim">
              An agent that clicks and types on your computer is a power tool,
              and power tools need guards, not vibes. The Gatekeeper is
              deterministic code — not a model — and it is the only path from a
              model’s intention to your screen. Every action it allows is
              written to an HMAC-chained, append-only audit log.
            </p>
          </Rise>

          <Rise delay={0.2} className="mt-16">
            <Gatekeeper />
          </Rise>

          <Rise delay={0.1} className="mt-20">
            <p className="annotation">Four ways to drop the baton</p>
          </Rise>
          <Cascade
            delay={0.2}
            step={0.07}
            className="mt-6 grid gap-px border border-line bg-line sm:grid-cols-2 lg:grid-cols-4"
          >
            {KILLS.map((k, i) => (
              <Item key={k.name} className="bg-ink px-5 py-6">
                <p className="font-mono text-xs text-halt">0{i + 1}</p>
                <h3 className="mt-3 font-mono text-sm font-normal text-ivory">{k.name}</h3>
                <p className="mt-2 text-sm leading-relaxed text-ivory-dim">{k.body}</p>
              </Item>
            ))}
          </Cascade>
          <p className="annotation mt-6">
            four independent mechanisms · the agent can disable none of them
          </p>
        </div>
      </NumeralSection>

      {/* ──────────────────────── Interlude ───────────────────────── */}
      <section className="border-t border-line bg-ink-2/40">
        <div className="mx-auto max-w-6xl px-5 py-28 text-center md:px-8 md:py-36">
          <Lines
            as="p"
            className="display-italic mx-auto max-w-4xl text-[clamp(1.8rem,4.2vw,3.4rem)] leading-tight"
            lines={["“The agent is a guest,", "not an owner.”"]}
          />
          <Rise delay={0.4}>
            <p className="annotation mt-8">Invariant VI of VII</p>
            <Link href="/manifesto" className="prose-link mt-4 inline-block text-sm">
              Read all seven invariants
            </Link>
          </Rise>
        </div>
      </section>

      {/* ─────────────────────────── Coda ─────────────────────────── */}
      <section className="staff border-t border-line">
        <div className="mx-auto max-w-6xl px-5 py-28 md:px-8 md:py-40">
          <Rise>
            <p className="annotation">Coda</p>
          </Rise>
          <Lines
            delay={0.1}
            className="display mt-6 text-[clamp(2.6rem,6.5vw,5rem)]"
            lines={["Take your seat."]}
          />
          <Rise delay={0.3}>
            <p className="mt-8 max-w-md leading-relaxed text-ivory-dim">
              Free forever. No feature gates, no premium tier, no nags. A
              signed binary on your machine and a donate button in the menu
              bar — that is the whole business model.
            </p>
          </Rise>
          <Rise delay={0.45}>
            <div className="mt-10">
              <Magnetic>
                <Link href="/download" className="trace cta">
                  Get Synaptic — macOS · Windows · Linux
                </Link>
              </Magnetic>
            </div>
          </Rise>
        </div>
      </section>
    </>
  );
}
