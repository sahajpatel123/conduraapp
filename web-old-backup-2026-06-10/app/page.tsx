import { ArrowRight, Download, ExternalLink, LockKeyhole, MousePointer2, ShieldCheck } from "lucide-react";
import { AgentLoop } from "@/components/agent-loop";
import { ControlTheater } from "@/components/control-theater";
import { LinkButton } from "@/components/button";
import { HeroCommandSurface } from "@/components/hero-command-surface";
import { PlatformDownloadGrid, ReleaseReadiness } from "@/components/download-panel";
import { Reveal } from "@/components/reveal";
import { Section, SectionHeader } from "@/components/section";
import { TrustGrid } from "@/components/trust-grid";
import { accountPrinciples, docsCards, trustPillars } from "@/lib/site-data";

export default function HomePage() {
  return (
    <main id="main-content">
      <section className="os-hero">
        <img
          src="/media/synaptic-blackbox-layer.png"
          alt=""
          aria-hidden="true"
          className="os-hero-media"
        />
        <div className="os-hero-shade" aria-hidden="true" />
        <div className="os-hero-grid container-wide">
          <Reveal>
            <div className="hero-briefing">
              <p className="section-label">SYNAPTIC / LOCAL COMMAND LAYER</p>
              <h1 className="hero-title">
                The agent that waits at the edge of your OS.
              </h1>
              <p className="hero-deck">
                Free, local-first desktop intelligence that appears on command, routes through your own models, and stops at deterministic approval boundaries before touching the machine.
              </p>
              <div className="hero-actions">
                <LinkButton href="/download">
                  <Download aria-hidden="true" size={18} />
                  Download preview
                </LinkButton>
                <LinkButton href="/safety" variant="secondary">
                  Inspect Gatekeeper
                  <ArrowRight aria-hidden="true" size={18} />
                </LinkButton>
              </div>

              <div className="system-ledger" aria-label="Synaptic trust principles">
                <div>
                  <span>Local</span>
                  <strong>No account required for desktop use.</strong>
                </div>
                <div>
                  <span>Owned</span>
                  <strong>Models and keys are configured by the user.</strong>
                </div>
                <div>
                  <span>Bounded</span>
                  <strong>Risky actions stop at deterministic policy.</strong>
                </div>
                <div>
                  <span>Proven</span>
                  <strong>Important actions leave an audit trail.</strong>
                </div>
              </div>
            </div>
          </Reveal>
          <Reveal delay={0.12}>
            <div className="hero-state-board" aria-label="Synaptic launch sequence">
              <div className="state-board-head">
                <span>LIVE SEQUENCE</span>
                <strong>summon / speak / plan / gate / audit</strong>
              </div>
              <HeroCommandSurface />
            </div>
          </Reveal>
        </div>
      </section>

      <Section id="product" className="blackbox-section">
        <div className="blackbox-grid">
          <div>
            <p className="section-label">Product position</p>
            <h2 className="section-title">A command layer, not another AI tab.</h2>
          </div>
          <div className="blackbox-copy">
            <p>
              Synaptic belongs at the operating-system layer: summoned by hotkey, spoken to when useful, silent when not, and constrained by deterministic safety code before touching the real machine.
            </p>
            <p>
              The product story is a state machine: observe only what is needed, form a visible plan, stop at the Gatekeeper, wait for the human, then record what happened.
            </p>
          </div>
        </div>
      </Section>

      <Section className="command-film-section">
        <div className="film-layout">
          <div>
            <SectionHeader
              label="Approval path"
              title="Power is introduced through the boundary."
              lead="The first-class interaction is not a prompt box. It is a visible sequence from intent to permission to audit."
            />
            <div className="protocol-note">
              <ShieldCheck aria-hidden="true" size={20} />
              <span>Network, write, and destructive actions are designed to stop at policy before execution.</span>
            </div>
          </div>
          <AgentLoop />
        </div>
      </Section>

      <Section className="protocol-section">
        <div className="section-ruler">
          <SectionHeader
            label="Operating covenant"
            title="Every capability is paired with a boundary."
            lead="The premium feeling should come from trust, speed, and control, not from pretending the agent is magic."
          />
        </div>
        <TrustGrid items={trustPillars} />
      </Section>

      <Section className="perception-section">
        <div className="grid gap-10 lg:grid-cols-[0.82fr_1.18fr] lg:items-center">
          <div>
            <SectionHeader
              label="Selective perception"
              title="Synaptic observes only what the task requires."
              lead="The control stack makes the invisible architecture feel concrete: cheaper perception first, verification before action, and human approval at risk boundaries."
            />
            <p className="mt-6 max-w-xl text-sm leading-6 text-white/56">
              The motion is restrained on purpose. It maps to product states instead of decorative movement: perceive, verify, gate, execute, and record.
            </p>
          </div>
          <ControlTheater />
        </div>
      </Section>

      <Section id="download" className="download-section">
        <div className="section-ruler">
          <SectionHeader
            label="Release artifacts"
            title="The release bay is visible. The installers are not faked."
            lead="These controls intentionally do not start downloads yet. Installers will be connected only after the desktop product is verified, signed, and checksummed."
          />
        </div>
        <PlatformDownloadGrid />
        <div className="mt-5">
          <ReleaseReadiness />
        </div>
      </Section>

      <Section className="identity-section">
        <div className="grid gap-8 lg:grid-cols-[0.9fr_1.1fr] lg:items-center">
          <Reveal>
            <div>
              <SectionHeader
                label="Optional account"
                title="Download first. Sign in only when it creates value."
                lead="Synaptic's core local use should not require an account. Authentication belongs in the browser for dashboard and device services later."
              />
              <div className="mt-8 flex flex-col gap-3 sm:flex-row">
                <LinkButton href="/dashboard" variant="secondary">
                  Dashboard placeholder
                  <ExternalLink aria-hidden="true" size={18} />
                </LinkButton>
                <LinkButton href="/privacy" variant="ghost">
                  Privacy details
                </LinkButton>
              </div>
            </div>
          </Reveal>
          <Reveal delay={0.08}>
            <div className="panel p-5">
              <div className="flex items-center gap-3 border-b border-white/10 pb-4">
                <span className="grid h-10 w-10 place-items-center rounded-md border border-white/15 bg-white/5 text-white">
                  <LockKeyhole aria-hidden="true" size={19} />
                </span>
                <div>
                  <h3 className="text-lg font-semibold text-white">Browser-based sign-in later</h3>
                  <p className="text-sm text-white/50">No embedded password form inside the desktop app.</p>
                </div>
              </div>
              <ul className="mt-5 grid gap-3">
                {accountPrinciples.map((item) => (
                  <li key={item} className="flex gap-3 rounded-md border border-white/10 bg-white/[0.035] p-3 text-sm text-white/62">
                    <MousePointer2 className="mt-0.5 shrink-0 text-white" aria-hidden="true" size={17} />
                    {item}
                  </li>
                ))}
              </ul>
            </div>
          </Reveal>
        </div>
      </Section>

      <Section>
        <div className="section-ruler">
          <SectionHeader
            label="Docs and trust"
            title="The important information is visible before launch."
            lead="Privacy, terms, setup, changelog, and support surfaces are part of the foundation now, so the public site can grow without changing its trust model."
          />
        </div>
        <div className="grid gap-4 md:grid-cols-2 lg:grid-cols-4">
          {docsCards.map((card, index) => {
            const Icon = card.icon;
            return (
              <Reveal key={card.title} delay={index * 0.05}>
                <article className="panel h-full p-5">
                  <Icon className="text-slate-950" aria-hidden="true" size={22} />
                  <h3 className="mt-5 text-lg font-semibold text-slate-950">{card.title}</h3>
                  <p className="mt-3 text-sm leading-6 text-slate-600">{card.body}</p>
                </article>
              </Reveal>
            );
          })}
        </div>
      </Section>
    </main>
  );
}
