import Link from "next/link";
import { Cascade, Item, Lines, Rise } from "@/components/motion/reveal";
import { Magnetic } from "@/components/motion/magnetic";
import { Tilt } from "@/components/motion/tilt";
import { Terminal } from "@/components/pieces/terminal";
import { Breaker } from "@/components/pieces/breaker";
import { Counter } from "@/components/pieces/counter";
import { ToolMarquee } from "@/components/pieces/marquee";
import { Illumination } from "@/components/pieces/illumination";
import { ORCHESTRA } from "@/lib/site";

/*
  Act I: the dark, the bulb, the touch (inside <Illumination/>).
  Act II: the lit room — what the conductor does with the light on.
  A live wire runs down the left edge of Act II, from the bulb to "fin."
*/

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

const KILLS = [
  { name: "Hard hotkey", body: "One chord, anywhere, and the agent stops mid-gesture." },
  { name: "Watchdog timer", body: "If the agent stalls or runs long, it is cut off automatically." },
  { name: "Network isolation", body: "Sever every outbound call with a single switch." },
  { name: "Menu-bar kill", body: "The red item at the top of your screen. Always there." },
];

function WireNode({ label }: { label: string }) {
  return (
    <p className="annotation flex items-center gap-3">
      <span aria-hidden className="size-2 rotate-45 border border-line-strong bg-glow/40" />
      {label}
    </p>
  );
}

export default function Home() {
  return (
    <>
      {/* ───────────── Act I — the dark, the bulb, the touch ───────────── */}
      <Illumination />

      {/* ─────────────────────── Act II — lights on ─────────────────────── */}
      <div className="relative">
        {/* the live wire from the bulb, running the length of the act */}
        <svg
          aria-hidden
          className="absolute top-0 bottom-0 left-7 hidden w-px xl:block"
          preserveAspectRatio="none"
        >
          <line x1="0.5" y1="0" x2="0.5" y2="100%" stroke="var(--t-line-strong)" strokeWidth="1" />
          <line x1="0.5" y1="0" x2="0.5" y2="100%" stroke="var(--t-glow)" strokeWidth="1.5" className="current" opacity="0.6" />
        </svg>

        {/* the roster, in motion */}
        <section aria-label="The orchestra">
          <ToolMarquee />
        </section>

        {/* meet the conductor */}
        <section className="staff relative">
          <div className="mx-auto grid max-w-6xl gap-14 px-5 py-28 md:grid-cols-[1.1fr_1fr] md:items-center md:gap-16 md:px-8 md:py-36">
            <div>
              <Rise>
                <WireNode label="Scene 1 — the summon" />
              </Rise>
              <Lines
                delay={0.1}
                className="display mt-6 text-[clamp(2.2rem,5vw,3.9rem)]"
                lines={[
                  "Summoned,",
                  <span key="i" className="display-italic font-normal">
                    not launched.
                  </span>,
                ]}
              />
              <Rise delay={0.35}>
                <p className="mt-8 max-w-md leading-relaxed text-ivory-dim">
                  No dock icon to find, no window to wait for. Double-tap the
                  hotkey — or say the wake word — and Synaptic is at your
                  cursor before your hand leaves the keyboard. The bulb does
                  not warm up. It is simply on.
                </p>
              </Rise>
              <Cascade delay={0.45} className="mt-12 grid grid-cols-1 gap-6 sm:grid-cols-3">
                <Item>
                  <Counter to={500} unit="ms" />
                  <p className="annotation mt-2">cold start</p>
                </Item>
                <Item>
                  <Counter to={100} unit="ms" />
                  <p className="annotation mt-2">hotkey → overlay</p>
                </Item>
                <Item>
                  <Counter to={1.5} unit="s" decimals={1} />
                  <p className="annotation mt-2">first token</p>
                </Item>
              </Cascade>
            </div>
            <Rise delay={0.3}>
              <Terminal />
            </Rise>
          </div>
        </section>

        {/* the orchestra, seated */}
        <section className="border-t border-line bg-ink-2/60">
          <div className="mx-auto max-w-6xl px-5 py-28 md:px-8 md:py-36">
            <Rise>
              <WireNode label="Scene 2 — the orchestra" />
            </Rise>
            <Lines
              delay={0.1}
              className="display mt-6 max-w-3xl text-[clamp(2.2rem,5vw,3.9rem)]"
              lines={[
                "It doesn’t replace your AI.",
                <span key="i" className="display-italic font-normal text-brass">
                  It conducts it.
                </span>,
              ]}
            />
            <Rise delay={0.3}>
              <p className="mt-8 max-w-xl leading-relaxed text-ivory-dim">
                Claude Code, Codex, Ollama, the subscriptions you already pay
                for — today they sit in separate rooms, unaware of one another.
                Synaptic seats them as one orchestra: a router picks the right
                player for each passage, a delegation bus hands off the work,
                and your task comes back as a single finished piece.
              </p>
            </Rise>
            <Cascade delay={0.2} step={0.06} className="mt-14 grid grid-cols-2 gap-4 lg:grid-cols-4">
              {ORCHESTRA.map((tool) => (
                <Item key={tool}>
                  <Tilt className="group h-full border border-line bg-ink px-5 py-6 transition-colors duration-300 hover:border-line-strong">
                    <p className="font-mono text-sm">{tool}</p>
                    <p className="annotation mt-2 !tracking-[0.12em] transition-colors duration-300 group-hover:!text-brass">
                      {SECTIONS[tool]}
                    </p>
                    <p className="mt-5 font-mono text-[10px] text-ivory-faint">
                      ● seated when found on your machine
                    </p>
                  </Tilt>
                </Item>
              ))}
            </Cascade>
          </div>
        </section>

        {/* the gatekeeper */}
        <section className="relative border-t border-line">
          <div aria-hidden className="absolute inset-0 overflow-hidden">
            <div className="shaft top-[-10%] right-[12%] h-[120%] w-52" />
          </div>
          <div className="relative mx-auto max-w-6xl px-5 py-28 md:px-8 md:py-36">
            <Rise>
              <WireNode label="Scene 3 — the breaker on the wire" />
            </Rise>
            <Lines
              delay={0.1}
              className="display mt-6 max-w-3xl text-[clamp(2.2rem,5vw,3.9rem)]"
              lines={[
                "Models propose.",
                <span key="i" className="display-italic font-normal">
                  The Gatekeeper disposes.
                </span>,
              ]}
            />
            <Rise delay={0.3}>
              <p className="mt-8 max-w-xl leading-relaxed text-ivory-dim">
                An agent that clicks and types on your computer is a power
                tool, and power tools need breakers, not vibes. The Gatekeeper
                is deterministic code — not a model — wired between every
                model’s intention and your screen. Every action it passes is
                written to an HMAC-chained, append-only audit log.
              </p>
            </Rise>
            <Rise delay={0.2} className="mt-14">
              <Breaker />
            </Rise>
            <Rise delay={0.1} className="mt-20">
              <WireNode label="Four ways to cut the power" />
            </Rise>
            <Cascade delay={0.2} step={0.07} className="mt-6 grid grid-cols-1 gap-4 sm:grid-cols-2 lg:grid-cols-4">
              {KILLS.map((k, i) => (
                <Item key={k.name}>
                  <Tilt className="h-full border border-line bg-ink px-5 py-6">
                    <p className="font-mono text-xs text-halt">0{i + 1}</p>
                    <h3 className="mt-3 font-mono text-sm font-normal">{k.name}</h3>
                    <p className="mt-2 text-sm leading-relaxed text-ivory-dim">{k.body}</p>
                  </Tilt>
                </Item>
              ))}
            </Cascade>
            <p className="annotation mt-6">
              four independent mechanisms · the agent can disable none of them
            </p>
          </div>
        </section>

        {/* interlude */}
        <section className="border-t border-line bg-ink-2/60">
          <div className="mx-auto max-w-6xl px-5 py-28 text-center md:px-8 md:py-32">
            <Lines
              as="p"
              className="display-italic mx-auto max-w-4xl text-[clamp(2rem,4.6vw,3.6rem)] leading-tight"
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

        {/* coda */}
        <section className="staff relative border-t border-line">
          <div className="mx-auto max-w-6xl px-5 py-28 md:px-8 md:py-36">
            <Rise>
              <WireNode label="Curtain" />
            </Rise>
            <Lines
              delay={0.1}
              className="display mt-6 text-[clamp(2.6rem,6.5vw,4.8rem)]"
              lines={[
                "Leave the light on.",
              ]}
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
      </div>
    </>
  );
}
