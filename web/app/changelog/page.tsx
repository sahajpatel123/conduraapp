import type { Metadata } from "next";
import Link from "next/link";
import { Cascade, Item, Lines, Rise } from "@/components/motion/reveal";

export const metadata: Metadata = {
  title: "Rehearsals",
  description:
    "The rehearsal log — what has been built, phase by phase, on the way to Synaptic's opening night.",
};

/* Real progress, phase by phase, drawn from the project logbook. */
const REHEARSALS = [
  {
    date: "June 2026",
    numeral: "VI",
    title: "Living presence",
    status: "complete",
    notes: [
      "One status enum drives tray, overlay, voice and session state.",
      "Every agent action now routes through the gated executor — the Gatekeeper bridge is structural, not optional.",
      "Double-tap hotkey summon wired to the presence orchestrator.",
      "Voice pipeline with SHA-256-pinned whisper binary and model.",
      "End-to-end loop: voice → transcript → model stream → speech.",
    ],
  },
  {
    date: "June 2026",
    numeral: "—",
    title: "Hardening pass",
    status: "complete",
    notes: [
      "Comprehensive security and correctness fixes across the daemon.",
      "Race-condition and cross-platform test failures eliminated; the suite runs clean under the race detector.",
    ],
  },
  {
    date: "June 2026",
    numeral: "V",
    title: "Hands and memory",
    status: "complete",
    notes: [
      "Computer-use interfaces with an expanded action classifier.",
      "macOS Accessibility bridge — the agent can see what it touches.",
      "Twin-snapshot verification: every action checked against before/after reality.",
      "Three-layer memory: episodic, semantic, procedural.",
      "Planner-driven multi-step agent loop.",
    ],
  },
  {
    date: "June 2026",
    numeral: "IV",
    title: "A voice and a face",
    status: "complete",
    notes: [
      "Recorder, transcriber and speaker implementations.",
      "Push-to-talk hotkey and the presence orchestrator.",
      "Overlay controller and tray state indicator.",
      "Thin agent loop with Gatekeeper integration.",
    ],
  },
  {
    date: "Spring 2026",
    numeral: "I–III",
    title: "The pit is built",
    status: "complete",
    notes: [
      "The daemon: IPC transport, JSON-RPC server, config, secrets.",
      "The router, failover breaker and spend monitor.",
      "HMAC-chained append-only audit log.",
      "Streaming pipeline across Anthropic, Google and OpenAI-compatible providers.",
    ],
  },
] as const;

const UPCOMING = [
  {
    numeral: "VII",
    title: "The full ensemble",
    notes: "Delegation bus across CLIs, skills system, MCP gateway.",
  },
  {
    numeral: "VIII",
    title: "Opening night",
    notes: "Signed, notarized binaries for macOS, Windows and Linux.",
  },
];

export default function Changelog() {
  return (
    <>
      <section className="staff">
        <div className="mx-auto max-w-6xl px-5 pt-40 pb-20 md:px-8 md:pt-48 md:pb-28">
          <Rise>
            <p className="annotation">The rehearsal log</p>
          </Rise>
          <Lines
            as="h1"
            delay={0.15}
            className="display mt-6 text-[clamp(2.6rem,6.5vw,5rem)]"
            lines={[
              "Built in public view",
              <span key="i" className="display-italic">
                of its own audit log.
              </span>,
            ]}
          />
          <Rise delay={0.45}>
            <p className="mt-8 max-w-xl leading-relaxed text-ivory-dim">
              Every phase below shipped with its tests passing under the race
              detector and its lint clean — the same discipline the agent will
              be held to on your machine. Six of eight rehearsals are done.
            </p>
          </Rise>
        </div>
      </section>

      <section className="border-t border-line" aria-label="Completed phases">
        <div className="mx-auto max-w-6xl px-5 py-20 md:px-8 md:py-28">
          {REHEARSALS.map((r) => (
            <Cascade
              key={r.title}
              className="grid gap-6 border-b border-line py-12 md:grid-cols-[11rem_8rem_1fr] md:gap-10"
            >
              <Item>
                <p className="annotation">{r.date}</p>
                <p className="annotation mt-2 !text-brass">{r.status}</p>
              </Item>
              <Item>
                <span className="numeral-outline text-5xl md:text-6xl">{r.numeral}</span>
              </Item>
              <Item>
                <h2 className="font-display text-2xl md:text-3xl">{r.title}</h2>
                <ul className="mt-5 space-y-2.5">
                  {r.notes.map((note) => (
                    <li key={note} className="flex gap-3 text-sm leading-relaxed text-ivory-dim">
                      <span aria-hidden className="mt-2.5 h-px w-3 shrink-0 bg-brass/60" />
                      {note}
                    </li>
                  ))}
                </ul>
              </Item>
            </Cascade>
          ))}
        </div>
      </section>

      <section className="border-t border-line bg-ink-2/40" aria-label="Upcoming phases">
        <div className="mx-auto max-w-6xl px-5 py-20 md:px-8 md:py-28">
          <Rise>
            <h2 className="annotation">Still to rehearse</h2>
          </Rise>
          <Cascade delay={0.15} className="mt-8 grid gap-px border border-line bg-line md:grid-cols-2">
            {UPCOMING.map((u) => (
              <Item key={u.title} className="bg-ink px-6 py-8">
                <span className="numeral-outline text-4xl">{u.numeral}</span>
                <h3 className="mt-4 font-display text-xl font-normal">{u.title}</h3>
                <p className="mt-2 text-sm leading-relaxed text-ivory-dim">{u.notes}</p>
              </Item>
            ))}
          </Cascade>
          <Rise delay={0.3}>
            <p className="mt-10 text-sm text-ivory-dim">
              When the last bar is rehearsed, the binaries appear at{" "}
              <Link href="/download" className="prose-link">
                the box office
              </Link>
              .
            </p>
          </Rise>
        </div>
      </section>
    </>
  );
}
